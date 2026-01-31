package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rezible/rezible/access"
	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
	oapi "github.com/rezible/rezible/openapi/v1"
)

const (
	defaultSessionDuration = time.Hour
)

type AuthService struct {
	authRoute     string
	orgs          rez.OrganizationService
	users         rez.UserService
	providers     []rez.AuthSessionProvider
	sessionSecret []byte
}

var _ rez.AuthService = (*AuthService)(nil)

func NewAuthSessionService(ctx context.Context, orgs rez.OrganizationService, users rez.UserService) (*AuthService, error) {
	secretKey := rez.Config.GetString("auth.session_secret_key")
	if secretKey == "" {
		return nil, errors.New("auth session secret key must be set")
	}

	authRoute, routeErr := url.JoinPath(rez.Config.ApiRouteBase(), rez.Config.AuthRouteBase())
	if routeErr != nil {
		return nil, fmt.Errorf("loading auth route: %w", routeErr)
	}

	providers, provsErr := getAuthSessionProviders(ctx, secretKey)
	if provsErr != nil {
		return nil, fmt.Errorf("loading session providers: %w", provsErr)
	}

	return &AuthService{
		authRoute:     authRoute,
		orgs:          orgs,
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

func (s *AuthService) CreateAuthContext(ctx context.Context, sess *rez.AuthSession) (context.Context, error) {
	userCtx, userErr := s.users.CreateUserContext(ctx, sess.UserId)
	if userErr != nil {
		log.Debug().Err(userErr).Msg("failed to create user auth context")
		return nil, userErr
	}
	return context.WithValue(userCtx, authUserSessionContextKey{}, sess), nil
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
			ctx, ctxErr := s.CreateAuthContext(r.Context(), sess)
			if ctxErr != nil {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(ctx))
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

func (s *AuthService) AuthRouteHandler() http.Handler {
	logoutPath := s.authRoute + "/logout"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == logoutPath {
			s.handleLogout(w, r)
			return
		}

		if s.delegateAuthFlowToProvider(w, r) {
			return
		}

		http.Redirect(w, r, rez.Config.AppUrl(), http.StatusFound)
	})
}

func (s *AuthService) handleLogout(w http.ResponseWriter, r *http.Request) {
	for _, prov := range s.providers {
		if !prov.SessionExists(r) {
			continue
		}
		if sessErr := prov.ClearSession(w, r); sessErr != nil {
			log.Error().Err(sessErr).Msg("failed to clear session")
		}
	}
	http.SetCookie(w, s.makeSessionCookie(r, "", time.Now(), -1))
	http.Redirect(w, r, rez.Config.AppUrl(), http.StatusFound)
}

func (s *AuthService) makeUserSessionCreatedCallback(w http.ResponseWriter, r *http.Request, flowRoute string) func(ps rez.AuthProviderSession) {
	ctx := r.Context()
	return func(ps rez.AuthProviderSession) {
		redirect := ps.RedirectUrl
		if redirect == "" || redirect == flowRoute {
			redirect = rez.Config.AppUrl()
		}

		expiry := ps.ExpiresAt
		if expiry.IsZero() {
			expiry = time.Now().Add(defaultSessionDuration)
		}

		org, orgErr := s.orgs.FindOrCreateFromProvider(ctx, ps.Organization)
		if orgErr != nil {
			log.Error().Err(orgErr).Msg("FindOrCreateFromProvider")
			http.Error(w, "session error", http.StatusInternalServerError)
			return
		}

		findUserCtx := access.TenantContext(ctx, org.TenantID)
		usr, usrErr := s.users.FindOrCreateAuthProviderUser(findUserCtx, ps.User)
		if usrErr != nil {
			log.Error().Err(usrErr).Msg("FindOrCreateAuthProviderUser")
			http.Error(w, "session error", http.StatusInternalServerError)
			return
		}

		token, tokenErr := s.IssueAuthSessionToken(newUserAuthSession(usr.ID, ps.ExpiresAt))
		if tokenErr != nil {
			log.Error().Err(tokenErr).Msg("failed to issue session token")
			http.Error(w, "session error", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, s.makeSessionCookie(r, token, ps.ExpiresAt, 0))
		http.Redirect(w, r, redirect, http.StatusFound)
	}
}

func (s *AuthService) GetProviderStartFlowPath(prov rez.AuthSessionProvider) string {
	return s.authRoute + "/" + strings.ToLower(prov.Id())
}

func (s *AuthService) delegateAuthFlowToProvider(w http.ResponseWriter, r *http.Request) bool {
	for _, prov := range s.providers {
		provFlowRoute := s.GetProviderStartFlowPath(prov)
		if !strings.HasPrefix(r.URL.Path, provFlowRoute) {
			continue
		}

		if r.URL.Path == provFlowRoute {
			prov.HandleStartAuthFlow(w, r)
			return true
		}

		if prov.HandleAuthFlowRequest(w, r, s.makeUserSessionCreatedCallback(w, r, provFlowRoute)) {
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

type authSessionTokenClaims struct {
	jwt.RegisteredClaims
	Scopes rez.AuthSessionScopes `json:"scopes"`
	UserId uuid.UUID             `json:"userId"`
}

func (s *AuthService) IssueAuthSessionToken(sess *rez.AuthSession) (string, error) {
	claims := jwt.MapClaims{
		"userId": sess.UserId,
		"scopes": sess.Scopes,
		"exp":    jwt.NewNumericDate(sess.ExpiresAt),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.sessionSecret)
}

func (s *AuthService) VerifyAuthSessionToken(token string, scopes rez.AuthSessionScopes) (*rez.AuthSession, error) {
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
		return nil, rez.ErrInvalidUser
	}

	exp, expErr := claims.GetExpirationTime()
	if expErr != nil || exp.Before(time.Now()) {
		return nil, rez.ErrAuthSessionExpired
	}

	for name, v := range claims.Scopes {
		cv, ok := scopes[name]
		if !ok {
			return nil, rez.ErrAuthSessionInvalidScope
		}
		if !mapset.NewSet(v...).Equal(mapset.NewSet(cv...)) {
			return nil, rez.ErrAuthSessionInvalidScope
		}
	}

	return newUserAuthSession(claims.UserId, exp.Time), nil
}
