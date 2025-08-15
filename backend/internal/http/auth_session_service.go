package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	oapi "github.com/rezible/rezible/openapi"
)

const (
	defaultSessionDuration = time.Hour
)

type AuthSessionService struct {
	users         rez.UserService
	prov          rez.AuthSessionProvider
	sessionSecret []byte
}

var _ rez.AuthSessionService = (*AuthSessionService)(nil)

func NewAuthSessionService(users rez.UserService, prov rez.AuthSessionProvider) (*AuthSessionService, error) {
	secretKey := os.Getenv("AUTH_SESSION_SECRET_KEY")
	if secretKey == "" {
		return nil, errors.New("AUTH_SESSION_SECRET_KEY must be set")
	}
	return &AuthSessionService{
		users:         users,
		prov:          prov,
		sessionSecret: []byte(secretKey),
	}, nil
}

func (s *AuthSessionService) Provider() rez.AuthSessionProvider {
	return s.prov
}

type authUserSessionContextKey struct{}

func newUserAuthSession(userId uuid.UUID, expiresAt time.Time) *rez.AuthSession {
	return &rez.AuthSession{UserId: userId, ExpiresAt: expiresAt}
}

func (s *AuthSessionService) SetAuthSessionContext(ctx context.Context, sess *rez.AuthSession) context.Context {
	return context.WithValue(ctx, authUserSessionContextKey{}, sess)
}

func (s *AuthSessionService) GetAuthSession(ctx context.Context) (*rez.AuthSession, error) {
	sess, ok := ctx.Value(authUserSessionContextKey{}).(*rez.AuthSession)
	if !ok || sess == nil {
		return nil, rez.ErrNoAuthSession
	}
	return sess, nil
}

func isRedirectableError(err error) bool {
	return errors.Is(err, rez.ErrAuthSessionExpired) || errors.Is(err, rez.ErrNoAuthSession)
}

func (s *AuthSessionService) MCPServerMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sess, authErr := s.getMCPUserSession(r)
			if authErr != nil {
				// w.Header().Add("WWW-Authenticate", `Bearer resource_metadata="/.well-known/oauth-protected-resource"`)
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r.WithContext(s.SetAuthSessionContext(r.Context(), sess)))
		})
	}
}

func (s *AuthSessionService) getMCPUserSession(r *http.Request) (*rez.AuthSession, error) {
	bearerToken, tokenErr := oapi.GetRequestApiBearerToken(r)
	if tokenErr != nil {
		return nil, tokenErr
	}
	// TODO: a lot
	// return s.VerifyUserSessionToken(bearerToken)
	log.Debug().Str("bearer", bearerToken).Msg("skipping mcp auth verification")
	return newUserAuthSession(uuid.New(), time.Now().Add(time.Hour)), nil
}

func (s *AuthSessionService) AuthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth/logout" {
			s.prov.ClearSession(w, r)
			http.SetCookie(w, oapi.MakeSessionCookie(r, "", time.Now(), -1))
			http.Redirect(w, r, rez.FrontendUrl, http.StatusFound)
			return
		}

		if s.delegateAuthFlowToProvider(w, r) {
			return
		}

		token, cookieErr := oapi.GetRequestSessionCookieToken(r)
		if cookieErr != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		} else {
			_, sessErr := s.VerifyAuthSessionToken(token, nil)
			if sessErr != nil && isRedirectableError(sessErr) {
				s.prov.StartAuthFlow(w, r)
				return
			}
		}
		http.Redirect(w, r, rez.FrontendUrl, http.StatusFound)
	})
}

func (s *AuthSessionService) delegateAuthFlowToProvider(w http.ResponseWriter, r *http.Request) bool {
	return s.prov.HandleAuthFlowRequest(w, r, func(prov *ent.User, expiresAt time.Time, redirect string) {
		ctx := r.Context()

		if expiresAt.IsZero() {
			expiresAt = time.Now().Add(defaultSessionDuration)
		}

		var sessErr error
		usr, provUserErr := s.users.LookupProviderUser(ctx, prov)
		if provUserErr != nil && !ent.IsNotFound(provUserErr) {
			sessErr = fmt.Errorf("failed to match user from provider details: %w", provUserErr)
		} else {
			// we create a session with nil user id to indicate a mismatch between provider users and db users
			userId := uuid.Nil
			if usr != nil {
				userId = usr.ID
			}
			token, tokenErr := s.IssueAuthSessionToken(newUserAuthSession(userId, expiresAt), nil)
			if tokenErr != nil {
				sessErr = fmt.Errorf("failed to issue user session token: %w", tokenErr)
			} else {
				http.SetCookie(w, oapi.MakeSessionCookie(r, token, expiresAt, 0))
			}
		}

		if sessErr != nil {
			log.Error().Err(sessErr).Msg("failed to create session")
			http.Error(w, sessErr.Error(), http.StatusInternalServerError)
		} else if redirect != "" {
			http.Redirect(w, r, redirect, http.StatusFound)
		}
	})
}

type authSessionTokenClaims struct {
	jwt.RegisteredClaims
	Scope  map[string]string `json:"scope"`
	UserId uuid.UUID         `json:"userId"`
}

func (s *AuthSessionService) IssueAuthSessionToken(sess *rez.AuthSession, scope map[string]string) (string, error) {
	claims := jwt.MapClaims{
		"userId": sess.UserId,
		"scope":  scope,
		"exp":    jwt.NewNumericDate(sess.ExpiresAt),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.sessionSecret)
}

func (s *AuthSessionService) VerifyAuthSessionToken(token string, scope map[string]string) (*rez.AuthSession, error) {
	if token == "" {
		return nil, rez.ErrNoAuthSession
	}

	keyFunc := func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.sessionSecret, nil
	}

	parsed, parseErr := jwt.ParseWithClaims(token, &authSessionTokenClaims{}, keyFunc)
	if parseErr != nil {
		return nil, fmt.Errorf("parse: %w", parseErr)
	}

	claims, claimsOk := parsed.Claims.(*authSessionTokenClaims)
	if !claimsOk {
		return nil, fmt.Errorf("invalid claims")
	}

	if claims.UserId == uuid.Nil {
		return nil, rez.ErrAuthSessionUserMissing
	}

	for name, v := range claims.Scope {
		cv, ok := scope[name]
		if !ok || cv != v {
			return nil, rez.ErrAuthSessionInvalidScope
		}
	}

	exp, expErr := claims.GetExpirationTime()
	if expErr != nil || exp.Before(time.Now()) {
		return nil, rez.ErrAuthSessionExpired
	}

	return newUserAuthSession(claims.UserId, exp.Time), nil
}
