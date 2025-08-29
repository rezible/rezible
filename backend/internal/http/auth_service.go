package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
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

func (s *AuthService) Providers() []rez.AuthSessionProvider {
	return s.providers
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
	bearerToken := oapi.GetRequestBearerToken(r)
	if bearerToken == "" {
		return nil, rez.ErrNoAuthSession
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

		http.Redirect(w, r, rez.FrontendUrl, http.StatusFound)
	})
}

func (s *AuthService) makeUserSessionCreatedCallback(w http.ResponseWriter, r *http.Request, provFlowRoute string) func(ps rez.AuthProviderSession) {
	ctx := r.Context()

	return func(ps rez.AuthProviderSession) {
		redirect := ps.RedirectUrl
		if redirect == "" || redirect == provFlowRoute {
			redirect = rez.FrontendUrl
		}

		expiry := ps.ExpiresAt
		if expiry.IsZero() {
			expiry = time.Now().Add(defaultSessionDuration)
		}

		usr, usrErr := s.users.FindOrCreateAuthProviderUser(ctx, &ps.User, &ps.Tenant)
		if usrErr != nil {
			log.Error().Err(usrErr).Msg("FindOrCreateAuthProviderUser")
			http.Error(w, "session error", http.StatusInternalServerError)
			return
		}

		token, tokenErr := s.IssueAuthSessionToken(newUserAuthSession(usr.ID, ps.ExpiresAt), nil)
		if tokenErr != nil {
			log.Error().Err(tokenErr).Msg("failed to issue session token")
			http.Error(w, "session error", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, s.makeSessionCookie(r, token, ps.ExpiresAt, 0))
		http.Redirect(w, r, redirect, http.StatusFound)
	}
}

func (s *AuthService) delegateAuthFlowToProvider(w http.ResponseWriter, r *http.Request) bool {
	for _, prov := range s.providers {
		provFlowRoute := "/auth/" + strings.ToLower(prov.Name())
		if r.URL.Path == provFlowRoute {
			prov.StartAuthFlow(w, r)
			return true
		}

		onSessionCreatedCallback := s.makeUserSessionCreatedCallback(w, r, provFlowRoute)
		if prov.HandleAuthFlowRequest(w, r, onSessionCreatedCallback) {
			return true
		}
	}
	return false
}

func (s *AuthService) makeSessionCookie(r *http.Request, token string, expires time.Time, maxAge int) *http.Cookie {
	cookie := &http.Cookie{
		Name:     oapi.SessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expires,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   maxAge,
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
