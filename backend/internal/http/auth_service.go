package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	oapi "github.com/rezible/rezible/openapi"
)

var (
	authSessionContextKey = "auth_session"
	authSessionCookieName = "rez_session"
)

type AuthService struct {
	users         rez.UserService
	sessProvider  rez.AuthSessionProvider
	sessionSecret []byte
}

func NewAuthService(ctx context.Context, users rez.UserService, pl rez.ProviderLoader, sessionSecretKey string) (*AuthService, error) {
	sessProv, sessErr := pl.LoadAuthSessionProvider(ctx)
	if sessErr != nil {
		return nil, sessErr
	}

	return &AuthService{
		users:         users,
		sessProvider:  sessProv,
		sessionSecret: []byte(sessionSecretKey),
	}, nil
}

func (s *AuthService) storeSession(ctx context.Context, sess *rez.AuthSession) context.Context {
	return context.WithValue(ctx, authSessionContextKey, sess)
}

func (s *AuthService) GetSession(ctx context.Context) (*rez.AuthSession, error) {
	sess, ok := ctx.Value(authSessionContextKey).(*rez.AuthSession)
	if !ok || sess == nil {
		return nil, rez.ErrNoAuthSession
	}
	return sess, nil
}

func isRedirectableError(err error) bool {
	return errors.Is(err, rez.ErrAuthSessionExpired) || errors.Is(err, rez.ErrNoAuthSession)
}

func writeAuthError(w http.ResponseWriter, err error) {
	var resp oapi.StatusError
	if errors.Is(err, rez.ErrNoAuthSession) {
		resp = oapi.ErrorUnauthorized("no_session")
	} else if errors.Is(err, rez.ErrAuthSessionExpired) {
		resp = oapi.ErrorUnauthorized("session_expired")
	} else if errors.Is(err, rez.ErrAuthSessionUserMissing) {
		resp = oapi.ErrorUnauthorized("missing_user")
	} else {
		resp = oapi.ErrorUnauthorized("unknown")
	}
	respBody, jsonErr := json.Marshal(resp)
	if jsonErr != nil {
		http.Error(w, jsonErr.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(resp.GetStatus())
	_, writeErr := w.Write(respBody)
	if writeErr != nil {
		log.Warn().Err(writeErr).Msg("failed to write auth error response")
	}
}

func (s *AuthService) MakeRequireAuthMiddleware(redirectStartFlow bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sess, sessErr := s.getRequestAuthSession(r)
			if sessErr == nil {
				next.ServeHTTP(w, r.WithContext(s.storeSession(r.Context(), sess)))
				return
			}

			if redirectStartFlow && isRedirectableError(sessErr) {
				s.sessProvider.StartAuthFlow(w, r)
				return
			}

			writeAuthError(w, sessErr)
		})
	}
}

func (s *AuthService) MakeAuthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		providerHandled := s.providerAuthFlow(w, r)
		if providerHandled {
			return
		}

		if r.URL.Path == "/auth/logout" {
			// TODO: do logout
			http.Redirect(w, r, rez.FrontendUrl, http.StatusFound)
			return
		}

		_, sessErr := s.getRequestAuthSession(r)
		if sessErr != nil && isRedirectableError(sessErr) {
			s.sessProvider.StartAuthFlow(w, r)
		} else {
			http.Redirect(w, r, rez.FrontendUrl, http.StatusFound)
		}
	})
}

func (s *AuthService) providerAuthFlow(w http.ResponseWriter, r *http.Request) bool {
	var redirectUrl string
	var createSessionErr error

	onUserSessionCreated := func(provUser *ent.User, expiresAt time.Time, redirect string) {
		redirectUrl = redirect
		expiry := expiresAt
		if expiresAt.IsZero() {
			expiry = time.Now().Add(time.Hour)
		}

		id, lookupErr := s.lookupProviderUserId(r.Context(), provUser)
		if lookupErr != nil {
			createSessionErr = fmt.Errorf("failed to lookup provider user: %w", lookupErr)
		} else {
			createSessionErr = s.storeAuthSession(w, r, &rez.AuthSession{ExpiresAt: expiry, UserId: id})
		}
	}

	providerHandled := s.sessProvider.HandleAuthFlowRequest(w, r, onUserSessionCreated)
	if !providerHandled {
		return false
	}

	if createSessionErr != nil {
		writeAuthError(w, createSessionErr)
	} else if redirectUrl != "" {
		http.Redirect(w, r, redirectUrl, http.StatusFound)
	}

	return true
}

func (s *AuthService) lookupProviderUserId(ctx context.Context, usr *ent.User) (uuid.UUID, error) {
	email := usr.Email
	if rez.DebugMode && os.Getenv("REZ_DEBUG_DEFAULT_USER_EMAIL") != "" {
		log.Debug().Msg("using debug email")
		email = os.Getenv("REZ_DEBUG_DEFAULT_USER_EMAIL")
	}

	// TODO: use provider mapping to match user details
	user, lookupErr := s.users.GetByEmail(ctx, email)
	if lookupErr != nil {
		if ent.IsNotFound(lookupErr) {
			log.Debug().Str("email", email).Msg("failed to match provider user")
			return uuid.Nil, nil
		}
		return uuid.Nil, lookupErr
	}
	return user.ID, nil
}

func (s *AuthService) storeAuthSession(w http.ResponseWriter, r *http.Request, sess *rez.AuthSession) error {
	token, tokenErr := s.IssueSessionToken(sess)
	if tokenErr != nil {
		return tokenErr
	}

	cookie := &http.Cookie{
		Name:     authSessionCookieName,
		Value:    token,
		Domain:   r.Host,
		Path:     "/",
		Expires:  sess.ExpiresAt,
		Secure:   r.URL.Scheme == "https",
		HttpOnly: true,
		// SameSite: http.SameSiteLaxMode,
	}
	if domain, _, splitErr := net.SplitHostPort(r.Host); splitErr == nil {
		cookie.Domain = domain
	}

	http.SetCookie(w, cookie)
	return nil
}

func (s *AuthService) getRequestAuthSession(r *http.Request) (*rez.AuthSession, error) {
	tokenValue, tokenErr := s.getRequestAuthSessionToken(r)
	if tokenErr != nil {
		return nil, tokenErr
	}

	return s.VerifySessionToken(tokenValue)
}

func (s *AuthService) getRequestAuthSessionToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		split := strings.Split(authHeader, " ")
		if len(split) != 2 {
			return "", rez.ErrNoAuthSession
		}
		authType := split[0]
		token := split[1]
		if authType != "Bearer" {
			log.Debug().Str("authType", authType).Msg("invalid auth type")
			return "", fmt.Errorf("invalid auth type %s", authType)
		}
		return token, nil
	}

	authCookie, cookieErr := r.Cookie(authSessionCookieName)
	if cookieErr != nil {
		if errors.Is(cookieErr, http.ErrNoCookie) {
			return "", rez.ErrNoAuthSession
		}
		return "", cookieErr
	}
	return authCookie.Value, nil
}

type authSessionTokenClaims struct {
	jwt.RegisteredClaims
	Session rez.AuthSession `json:"session"`
}

func (s *AuthService) IssueSessionToken(session *rez.AuthSession) (string, error) {
	if session == nil {
		return "", errors.New("nil session")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"session": *session,
		"exp":     jwt.NewNumericDate(session.ExpiresAt),
	})

	signedToken, signErr := token.SignedString(s.sessionSecret)
	if signErr != nil {
		return "", fmt.Errorf("failed to sign token: %w", signErr)
	}

	return signedToken, nil
}

func (s *AuthService) parseSessionTokenClaims(tokenStr string) (*authSessionTokenClaims, error) {
	keyFunc := func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.sessionSecret, nil
	}

	token, parseErr := jwt.ParseWithClaims(tokenStr, &authSessionTokenClaims{}, keyFunc)
	if parseErr != nil {
		return nil, fmt.Errorf("parse: %w", parseErr)
	}

	claims, claimsOk := token.Claims.(*authSessionTokenClaims)
	if !claimsOk {
		return nil, fmt.Errorf("invalid claims")
	}
	return claims, nil
}

func (s *AuthService) VerifySessionToken(tokenStr string) (*rez.AuthSession, error) {
	claims, parseErr := s.parseSessionTokenClaims(tokenStr)
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse token: %w", parseErr)
	}

	if claims.Session.UserId == uuid.Nil {
		return nil, rez.ErrAuthSessionUserMissing
	}

	exp, expErr := claims.GetExpirationTime()
	if expErr != nil || exp.Before(time.Now()) {
		return nil, rez.ErrAuthSessionExpired
	}

	return &claims.Session, nil
}
