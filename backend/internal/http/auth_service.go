package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
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

type AuthService struct {
	users         rez.UserService
	providers     []rez.AuthSessionProvider
	sessionSecret []byte
}

var _ rez.AuthService = (*AuthService)(nil)

func NewAuthService(secretKey string, users rez.UserService, providers []rez.AuthSessionProvider) (*AuthService, error) {
	return &AuthService{
		users:         users,
		providers:     providers,
		sessionSecret: []byte(secretKey),
	}, nil
}

func (s *AuthService) Provider() rez.AuthSessionProvider {
	return s.providers[0]
}

type authUserSessionContextKey struct{}

func newUserAuthSession(userId uuid.UUID, expiresAt time.Time) *rez.AuthSession {
	return &rez.AuthSession{UserId: userId, ExpiresAt: expiresAt}
}

func (s *AuthService) SetAuthSession(ctx context.Context, sess *rez.AuthSession) context.Context {
	return context.WithValue(ctx, authUserSessionContextKey{}, sess)
}

func (s *AuthService) GetAuthSession(ctx context.Context) (*rez.AuthSession, error) {
	sess, ok := ctx.Value(authUserSessionContextKey{}).(*rez.AuthSession)
	if !ok || sess == nil {
		return nil, rez.ErrNoAuthSession
	}
	return sess, nil
}

func (s *AuthService) MCPServerMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sess, authErr := s.getMCPUserSession(r)
			if authErr != nil {
				// w.Header().Add("WWW-Authenticate", `Bearer resource_metadata="/.well-known/oauth-protected-resource"`)
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(s.SetAuthSession(r.Context(), sess)))
		})
	}
}

func (s *AuthService) getMCPUserSession(r *http.Request) (*rez.AuthSession, error) {
	bearerToken, tokenErr := oapi.GetRequestBearerToken(r)
	if tokenErr != nil {
		return nil, tokenErr
	}

	// TODO: actually verify stuff
	log.Debug().Str("bearer", bearerToken).Msg("skipping mcp auth verification")
	fakeSess := newUserAuthSession(uuid.New(), time.Now().Add(time.Hour))

	return fakeSess, nil
}

func (s *AuthService) UserAuthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth/logout" {
			for _, prov := range s.providers {
				prov.ClearSession(w, r)
			}
			http.SetCookie(w, s.makeSessionCookie(r, "", time.Now(), -1))
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
		}
		_, sessErr := s.VerifyAuthSessionToken(token, nil)
		if sessErr != nil {
			http.Error(w, "failed to verify session", http.StatusUnauthorized)
			return
			//isRedirectable := errors.Is(sessErr, rez.ErrAuthSessionExpired) || errors.Is(sessErr, rez.ErrNoAuthSession)
			//if isRedirectable {
			//	// TODO: dont redirect or check which provider used
			//	s.providers[0].StartAuthFlow(w, r)
			//	return
			//}
		}
		http.Redirect(w, r, rez.FrontendUrl, http.StatusFound)
	})
}

func (s *AuthService) delegateAuthFlowToProvider(w http.ResponseWriter, r *http.Request) bool {
	for _, prov := range s.providers {
		if r.URL.Path == "/auth/"+strings.ToLower(prov.Name()) {
			prov.StartAuthFlow(w, r)
			return true
		}
	}

	ctx := r.Context()

	onSessionCreatedCallback := func(provUser *ent.User, expiresAt time.Time, redirect string) {
		if expiresAt.IsZero() {
			expiresAt = time.Now().Add(defaultSessionDuration)
		}

		dbUser, provUserErr := s.users.LookupProviderUser(ctx, provUser)
		if provUserErr != nil && !ent.IsNotFound(provUserErr) {
			log.Error().Err(provUserErr).Msg("failed to match user from provider details")
			http.Error(w, "failed to match user", http.StatusInternalServerError)
			return
		}

		// nil user id indicates provider user not found in db
		userId := uuid.Nil
		if dbUser != nil {
			userId = dbUser.ID
		}

		token, tokenErr := s.IssueAuthSessionToken(newUserAuthSession(userId, expiresAt), nil)
		if tokenErr != nil {
			log.Error().Err(tokenErr).Msg("failed to issue user session token")
			http.Error(w, "failed to issue session token", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, s.makeSessionCookie(r, token, expiresAt, 0))

		if redirect != "" {
			http.Redirect(w, r, redirect, http.StatusFound)
		}
	}

	for _, prov := range s.providers {
		if prov.HandleAuthFlowRequest(w, r, onSessionCreatedCallback) {
			return true
		}
	}
	return false
}

func (s *AuthService) makeSessionCookie(r *http.Request, value string, expires time.Time, maxAge int) *http.Cookie {
	domain := r.Host
	if host, _, splitErr := net.SplitHostPort(r.Host); splitErr == nil {
		domain = host
	}
	cookie := &http.Cookie{
		Name:     oapi.SessionCookieName,
		Value:    value,
		Domain:   domain,
		Path:     "/",
		Expires:  expires,
		Secure:   !rez.DebugMode, // r.URL.Scheme == "https",
		HttpOnly: true,
		MaxAge:   maxAge,
		// SameSite: http.SameSiteLaxMode,
	}
	return cookie
}

func (s *AuthService) CreateVerifiedApiAuthContext(ctx context.Context, token string, requiredScopes []string) (context.Context, error) {
	sess, tokenErr := s.VerifyAuthSessionToken(token, nil)
	if tokenErr != nil {
		return nil, tokenErr
	}

	for _, scope := range requiredScopes {
		log.Debug().Str("scope", scope).Msg("check scope")
	}

	userCtx, userErr := s.users.CreateUserContext(ctx, sess.UserId)
	if userErr != nil {
		log.Debug().Err(userErr).Msg("failed to create user auth context")
		return nil, userErr
	}
	return s.SetAuthSession(userCtx, sess), nil
}

type authSessionTokenClaims struct {
	jwt.RegisteredClaims
	Scope  map[string]string `json:"scope"`
	UserId uuid.UUID         `json:"userId"`
}

func (s *AuthService) IssueAuthSessionToken(sess *rez.AuthSession, scope map[string]string) (string, error) {
	claims := jwt.MapClaims{
		"userId": sess.UserId,
		"scope":  scope,
		"exp":    jwt.NewNumericDate(sess.ExpiresAt),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.sessionSecret)
}

func (s *AuthService) VerifyAuthSessionToken(token string, scope map[string]string) (*rez.AuthSession, error) {
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
