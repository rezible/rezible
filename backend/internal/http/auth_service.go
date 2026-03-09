package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/failsafe-go/failsafe-go"
	"github.com/failsafe-go/failsafe-go/retrypolicy"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/integrations"
	"github.com/rezible/rezible/internal/http/oidc"
	"github.com/rezible/rezible/internal/http/saml"
)

const (
	defaultSessionDuration = time.Hour
)

type AuthService struct {
	orgs          rez.OrganizationService
	users         rez.UserService
	providers     []rez.AuthSessionProvider
	sessionSecret []byte
}

var _ rez.AuthService = (*AuthService)(nil)

func NewAuthSessionService(ctx context.Context, orgs rez.OrganizationService, users rez.UserService) (*AuthService, error) {
	secretKey := []byte(rez.Config.GetString("auth.session_secret_key"))
	if len(secretKey) == 0 {
		return nil, errors.New("auth session secret key must be set")
	}

	oidc.SessionSecretKey = secretKey

	return &AuthService{
		orgs:          orgs,
		users:         users,
		sessionSecret: secretKey,
	}, nil
}

func (s *AuthService) LoadSessionProviders(ctx context.Context) error {
	loadFuncs := make(map[string]func() (rez.AuthSessionProvider, error))

	if saml.ProviderEnabled() {
		loadFuncs["saml"] = func() (rez.AuthSessionProvider, error) {
			return saml.NewAuthSessionProvider(ctx)
		}
	}

	oidcIdps, oidcIdpsErr := integrations.GetOIDCAuthSessionIdentityProviders()
	if oidcIdpsErr != nil {
		return fmt.Errorf("integrations.GetOIDCAuthSessionIdentityProviders: %w", oidcIdpsErr)
	}
	if enabled, genericOidcIdp, cfgErr := oidc.GetGenericOIDCAuthSessionProvider(); enabled {
		if cfgErr != nil {
			return fmt.Errorf("oidc GenericOIDCAuthSessionProvider: %w", cfgErr)
		}
		oidcIdps = append(oidcIdps, genericOidcIdp)
	}
	for _, idp := range oidcIdps {
		loadFuncs["oidc."+idp.Id()] = func() (rez.AuthSessionProvider, error) {
			return oidc.LoadAuthSessionProvider(ctx, idp)
		}
	}

	defaultRetryPolicy := retrypolicy.NewBuilder[rez.AuthSessionProvider]().
		//HandleErrors(ErrConnecting).
		WithDelay(time.Second).
		WithMaxRetries(1)

	for name, loadFn := range loadFuncs {
		prov, loadErr := failsafe.With(defaultRetryPolicy.Build()).Get(loadFn)
		if loadErr != nil {
			return fmt.Errorf("loading auth session provider '%s': %w", name, loadErr)
		}
		s.providers = append(s.providers, prov)
	}

	return nil
}

func (s *AuthService) Providers() []rez.AuthSessionProvider {
	return s.providers
}

type authUserSessionContextKey struct{}

func newUserAuthSession(userId uuid.UUID, expiresAt time.Time) *rez.AuthSession {
	return &rez.AuthSession{UserId: userId, ExpiresAt: expiresAt}
}

func makeSessionCookie(token string, expires time.Time, maxAge int) *http.Cookie {
	cookie := &http.Cookie{
		Name:     rez.Config.AuthSessionCookieName(),
		Value:    token,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Expires:  expires,
		MaxAge:   maxAge,
	}
	return cookie
}

func makeRemoveSessionCookie() *http.Cookie {
	return makeSessionCookie("", time.Now(), -1)
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
	bearerToken := r.Header.Get("Authorization") //openapi.GetRequestBearerToken(r)
	if bearerToken == "" {
		return nil, rez.ErrNoAuthSession
	}

	// TODO: actually verify stuff
	log.Debug().Str("bearer", bearerToken).Msg("skipping mcp auth verification")
	fakeSess := newUserAuthSession(uuid.New(), time.Now().Add(time.Hour))

	return fakeSess, nil
}

func (s *AuthService) AuthRouteHandler() http.Handler {
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Debug().Str("path", r.URL.Path).Msg("auth route")
			next.ServeHTTP(w, r)
		})
	})
	r.Post("/logout", s.handleLogout)
	r.Route("/flow", func(fr chi.Router) {
		for _, p := range s.providers {
			fr.Mount(p.FlowPath(), p.MakeFlowPathHandler(s.authSessionCreatedCallback))
		}
	})
	r.NotFound(http.RedirectHandler(rez.Config.AppUrl(), http.StatusFound).ServeHTTP)
	return r
}

func (s *AuthService) GetProviderStartFlowPath(p rez.AuthSessionProvider) (string, error) {
	return url.JoinPath(rez.Config.AuthPath(), "flow", p.FlowPath())
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
	http.SetCookie(w, makeRemoveSessionCookie())
	http.Redirect(w, r, rez.Config.AppUrl(), http.StatusFound)
}

func (s *AuthService) authSessionCreatedCallback(w http.ResponseWriter, r *http.Request, ps *rez.AuthProviderSession) {
	ctx := r.Context()

	redirect := ps.RedirectUrl
	if redirect == "" || redirect == r.URL.Path {
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

	sess := newUserAuthSession(usr.ID, ps.ExpiresAt)
	token, tokenErr := s.IssueAuthSessionToken(sess, nil)
	if tokenErr != nil {
		log.Error().Err(tokenErr).Msg("failed to issue session token")
		http.Error(w, "session error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, makeSessionCookie(token, ps.ExpiresAt, 0))
	http.Redirect(w, r, redirect, http.StatusFound)
}

type authSessionTokenClaims struct {
	jwt.RegisteredClaims
	UserId uuid.UUID `json:"userId"`
	Scopes []string  `json:"scopes"`
}

func (s *AuthService) IssueAuthSessionToken(sess *rez.AuthSession, scopes []string) (string, error) {
	claims := jwt.MapClaims{
		"userId": sess.UserId,
		"scopes": scopes,
		"exp":    jwt.NewNumericDate(sess.ExpiresAt),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.sessionSecret)
}

func (s *AuthService) VerifyAuthSessionToken(token string, requiredScopes []string) (*rez.AuthSession, error) {
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

	//claimScopes := mapset.NewSet[string](claims.Scopes...)
	//for _, scope := range requiredScopes {
	//	if !claimScopes.Contains(scope) {
	//		return nil, rez.ErrAuthSessionMissingScope
	//	}
	//}

	return newUserAuthSession(claims.UserId, exp.Time), nil
}
