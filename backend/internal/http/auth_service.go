package http

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rs/zerolog/log"
	"net"
	"net/http"
	"strings"
	"time"
)

var (
	authSessionContextKey = "user_auth_session"
	authSessionCookieName = "rez_auth"
	authMountPath         = "/auth"

	errInvalidAuthToken = errors.New("invalid bearer token")
)

// AutoCreateUsers TODO: don't auto-create users
const AutoCreateUsers = true

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

func (s *AuthService) MakeRequireAuthMiddleware(redirect bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sess, sessErr := s.loadAuthSession(r)
			if sessErr == nil {
				authedReq := r.WithContext(context.WithValue(r.Context(), authSessionContextKey, sess))
				next.ServeHTTP(w, authedReq)
				return
			}

			log.Debug().Err(sessErr).Str("path", r.URL.Path).Msg("auth middleware")
			if errors.Is(sessErr, rez.ErrNoSessionUser) {
				log.Debug().Msg("session exists, no user")
				http.Error(w, "no_user", http.StatusForbidden)
				return
			}

			if !redirect {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if errors.Is(sessErr, rez.ErrNoAuthSession) {
				s.sessProvider.StartAuthFlow(w, r)
			} else {
				http.Error(w, "auth failed", http.StatusInternalServerError)
			}
		})
	}
}

func (s *AuthService) MakeAuthHandler() http.Handler {
	redirectAuth := http.RedirectHandler(rez.FrontendUrl, http.StatusFound)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handledByProvider := s.providerAuthFlow(w, r)
		if handledByProvider {
			return
		}

		_, sessErr := s.loadAuthSession(r)
		if sessErr != nil && !errors.Is(sessErr, rez.ErrNoSessionUser) {
			s.sessProvider.StartAuthFlow(w, r)
			return
		}
		log.Debug().Str("path", r.URL.Path).Msg("auth handler redirecting")
		redirectAuth.ServeHTTP(w, r)
	})
}

func (s *AuthService) providerAuthFlow(w http.ResponseWriter, r *http.Request) bool {
	var redirectUrl string
	var createSessionErr error
	onSessionCreated := func(sess *rez.AuthSession, ru string) {
		redirectUrl = ru
		user, userErr := s.lookupProviderUser(r.Context(), &sess.User)
		if userErr != nil {
			if ent.IsNotFound(userErr) {
				log.Debug().Str("redirect", redirectUrl).Msg("provider session created, user does not exist")
				user = &ent.User{}
			} else {
				createSessionErr = fmt.Errorf("failed to lookup provider user: %w", userErr)
			}
		}
		if user != nil {
			sess.User = *user
			createSessionErr = s.storeAuthSession(w, r, sess)
		}
	}

	handledByProvider := s.sessProvider.HandleAuthFlowRequest(w, r, onSessionCreated)
	if handledByProvider {
		if createSessionErr != nil {
			http.Error(w, createSessionErr.Error(), http.StatusBadRequest)
		} else if redirectUrl != "" {
			http.Redirect(w, r, redirectUrl, http.StatusFound)
		}
	}

	return handledByProvider
}

func (s *AuthService) lookupProviderUser(ctx context.Context, usr *ent.User) (*ent.User, error) {
	// TODO: use provider mapping to match user details
	return s.users.GetByEmail(ctx, usr.Email)
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

func (s *AuthService) loadAuthSession(r *http.Request) (*rez.AuthSession, error) {
	tokenValue, tokenErr := s.loadAuthSessionToken(r)
	if tokenErr != nil {
		return nil, tokenErr
	}

	sess, verifyErr := s.VerifySessionToken(tokenValue)
	if verifyErr != nil {
		return nil, fmt.Errorf("failed to verify session: %w", verifyErr)
	}

	if sess.User.ID == uuid.Nil {
		return nil, rez.ErrNoSessionUser
	}
	return sess, nil
}

func (s *AuthService) loadAuthSessionToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		split := strings.Split(authHeader, " ")
		if len(split) != 2 {
			return "", errInvalidAuthToken
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

func (s *AuthService) GetSession(ctx context.Context) (*rez.AuthSession, error) {
	sess, ok := ctx.Value(authSessionContextKey).(*rez.AuthSession)
	if !ok || sess == nil {
		return nil, rez.ErrNoAuthSession
	}
	return sess, nil
}

type authSessionTokenClaims struct {
	jwt.RegisteredClaims
	Session rez.AuthSession `json:"session"`
}

func (s *AuthService) IssueSessionToken(session *rez.AuthSession) (string, error) {
	if session == nil {
		return "", errors.New("nil session")
	}

	fmt.Printf("issuing: %+v\n", session)

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

func (s *AuthService) VerifySessionToken(tokenStr string) (*rez.AuthSession, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
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

	exp, expErr := claims.GetExpirationTime()
	if expErr != nil || exp.Before(time.Now()) {
		return nil, fmt.Errorf("claims expired: %s", exp)
	}

	return &claims.Session, nil
}
