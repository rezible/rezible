package http

import (
	"context"
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
)

const (
	defaultSessionDuration = time.Hour
)

type AuthSessionService struct {
	users         rez.UserService
	sessProvider  rez.AuthSessionProvider
	sessionSecret []byte
}

func NewAuthSessionService(ctx context.Context, users rez.UserService, sessProv rez.AuthSessionProvider, sessionSecretKey string) (*AuthSessionService, error) {
	return &AuthSessionService{
		users:         users,
		sessProvider:  sessProv,
		sessionSecret: []byte(sessionSecretKey),
	}, nil
}

func (s *AuthSessionService) ProviderName() string {
	return s.sessProvider.Name()
}

func (s *AuthSessionService) CreateSessionContext(ctx context.Context, sess *rez.AuthSession) context.Context {
	return context.WithValue(ctx, authSessionContextKey, sess)
}

func (s *AuthSessionService) GetSession(ctx context.Context) (*rez.AuthSession, error) {
	sess, ok := ctx.Value(authSessionContextKey).(*rez.AuthSession)
	if !ok || sess == nil {
		return nil, rez.ErrNoAuthSession
	}
	return sess, nil
}

func isRedirectableError(err error) bool {
	return errors.Is(err, rez.ErrAuthSessionExpired) || errors.Is(err, rez.ErrNoAuthSession)
}

func (s *AuthSessionService) MakeFrontendAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/favicon.ico" {
				next.ServeHTTP(w, r)
				return
			}

			sess, sessErr := s.verifyRequestAuthSessionTokenCookie(r)
			if sessErr == nil {
				next.ServeHTTP(w, r.WithContext(s.CreateSessionContext(r.Context(), sess)))
				return
			}

			if isRedirectableError(sessErr) {
				s.sessProvider.StartAuthFlow(w, r)
				return
			}
			http.Error(w, sessErr.Error(), http.StatusInternalServerError)
		})
	}
}

func (s *AuthSessionService) MakeMCPAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sess, sessErr := s.verifyRequestAuthSessionBearerToken(r)
			if sessErr == nil {
				next.ServeHTTP(w, r.WithContext(s.CreateSessionContext(r.Context(), sess)))
				return
			}

			if isRedirectableError(sessErr) {
				log.Error().Msg("TODO: MCP OAuth flow")
			}
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		})
	}
}

func (s *AuthSessionService) MakeUserAuthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if providerHandled := s.providerAuthFlow(w, r); providerHandled {
			return
		}

		if r.URL.Path == "/auth/logout" {
			s.clearAuthSession(w, r)
			http.Redirect(w, r, rez.FrontendUrl, http.StatusFound)
			return
		}

		_, sessErr := s.verifyRequestAuthSessionTokenCookie(r)
		if sessErr != nil && isRedirectableError(sessErr) {
			s.sessProvider.StartAuthFlow(w, r)
		} else {
			http.Redirect(w, r, rez.FrontendUrl, http.StatusFound)
		}
	})
}

func (s *AuthSessionService) providerAuthFlow(w http.ResponseWriter, r *http.Request) bool {
	var redirectUrl string
	var createSessionErr error

	onUserSessionCreated := func(provUser *ent.User, expiresAt time.Time, redirect string) {
		redirectUrl = redirect
		expiry := expiresAt
		if expiresAt.IsZero() {
			expiry = time.Now().Add(defaultSessionDuration)
		}

		userId, matchIdErr := s.matchUserIdFromProvider(r.Context(), provUser)
		if matchIdErr != nil {
			createSessionErr = fmt.Errorf("failed to match user id from provider details: %w", matchIdErr)
		} else {
			if userId == uuid.Nil {
				log.Debug().Msg("no internal user exists for auth provider supplied details")
			}
			createSessionErr = s.storeAuthSession(w, r, &rez.AuthSession{ExpiresAt: expiry, UserId: userId})
		}
	}

	providerHandled := s.sessProvider.HandleAuthFlowRequest(w, r, onUserSessionCreated)
	if !providerHandled {
		return false
	}

	if createSessionErr != nil {
		log.Error().Err(createSessionErr).Msg("failed to create session")
		http.Error(w, createSessionErr.Error(), http.StatusInternalServerError)
		//if writeErr := writeAuthError(w, createSessionErr); writeErr != nil {
		//	log.Warn().Err(writeErr).Msg("failed to write auth create session error response")
		//}
	} else if redirectUrl != "" {
		http.Redirect(w, r, redirectUrl, http.StatusFound)
	}

	return true
}

func (s *AuthSessionService) matchUserIdFromProvider(ctx context.Context, provUser *ent.User) (uuid.UUID, error) {
	// TODO: use provider mapping to match user details, not just by email
	email := provUser.Email
	if rez.DebugMode && os.Getenv("REZ_DEBUG_DEFAULT_USER_EMAIL") != "" {
		email = os.Getenv("REZ_DEBUG_DEFAULT_USER_EMAIL")
		log.Debug().Str("email", email).Msg("using debug auth email")
	}

	user, lookupErr := s.users.GetByEmail(ctx, email)
	if lookupErr != nil {
		if ent.IsNotFound(lookupErr) {
			return uuid.Nil, nil
		}
		return uuid.Nil, lookupErr
	}
	return user.ID, nil
}

func (s *AuthSessionService) makeSessionCookie(r *http.Request, value string, expires time.Time) *http.Cookie {
	cookie := &http.Cookie{
		Name:     oapi.SessionCookieName,
		Value:    value,
		Domain:   r.Host,
		Path:     "/",
		Expires:  expires,
		Secure:   r.URL.Scheme == "https",
		HttpOnly: true,
		// SameSite: http.SameSiteLaxMode,
	}
	if domain, _, splitErr := net.SplitHostPort(r.Host); splitErr == nil {
		cookie.Domain = domain
	}
	return cookie
}

func (s *AuthSessionService) storeAuthSession(w http.ResponseWriter, r *http.Request, sess *rez.AuthSession) error {
	token, tokenErr := s.IssueSessionToken(sess)
	if tokenErr != nil {
		return tokenErr
	}

	cookie := s.makeSessionCookie(r, token, sess.ExpiresAt)
	http.SetCookie(w, cookie)

	return nil
}

func (s *AuthSessionService) clearAuthSession(w http.ResponseWriter, r *http.Request) {
	clearCookie := s.makeSessionCookie(r, "", time.Now())
	clearCookie.MaxAge = -1
	http.SetCookie(w, clearCookie)
	s.sessProvider.ClearSession(w, r)
}

func (s *AuthSessionService) verifyRequestAuthSessionTokenCookie(r *http.Request) (*rez.AuthSession, error) {
	cookieToken, cookieErr := getRequestAuthSessionTokenCookie(r)
	if cookieErr != nil {
		return nil, fmt.Errorf("error getting token from cookie: %w", cookieErr)
	} else if cookieToken == "" {
		return nil, rez.ErrNoAuthSession
	}
	return s.VerifySessionToken(cookieToken)
}

func (s *AuthSessionService) verifyRequestAuthSessionBearerToken(r *http.Request) (*rez.AuthSession, error) {
	bearerToken, bearerErr := getRequestAuthorizationBearerToken(r)
	if bearerErr != nil {
		return nil, fmt.Errorf("error getting bearer token from authorization header: %w", bearerErr)
	} else if bearerToken != "" {
		return nil, rez.ErrNoAuthSession
	}
	return s.VerifySessionToken(bearerToken)
}

func getRequestTokenValue(r *http.Request) (string, error) {
	cookieToken, cookieErr := getRequestAuthSessionTokenCookie(r)
	if cookieErr != nil {
		return "", fmt.Errorf("error getting token from cookie: %w", cookieErr)
	} else if cookieToken != "" {
		return cookieToken, nil
	}

	bearerToken, bearerErr := getRequestAuthorizationBearerToken(r)
	if bearerErr != nil {
		return "", fmt.Errorf("error getting bearer token from authorization header: %w", bearerErr)
	} else if bearerToken != "" {
		return bearerToken, nil
	}

	return "", rez.ErrNoAuthSession
}

func getRequestAuthSessionTokenCookie(r *http.Request) (string, error) {
	authCookie, cookieErr := r.Cookie(oapi.SessionCookieName)
	if cookieErr != nil {
		if errors.Is(cookieErr, http.ErrNoCookie) {
			return "", nil
		}
		return "", cookieErr
	}
	return authCookie.Value, nil
}

func getRequestAuthorizationBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", nil
	}
	split := strings.Split(authHeader, " ")
	if len(split) != 2 {
		return "", nil
	}
	authType := split[0]
	token := split[1]
	if authType != "Bearer" {
		return "", fmt.Errorf("invalid Authorization type %s", authType)
	}
	return token, nil
}

type authSessionTokenClaims struct {
	jwt.RegisteredClaims
	Session rez.AuthSession `json:"session"`
}

func (s *AuthSessionService) IssueSessionToken(session *rez.AuthSession) (string, error) {
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

func (s *AuthSessionService) VerifySessionToken(tokenStr string) (*rez.AuthSession, error) {
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

func (s *AuthSessionService) parseSessionTokenClaims(tokenStr string) (*authSessionTokenClaims, error) {
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
