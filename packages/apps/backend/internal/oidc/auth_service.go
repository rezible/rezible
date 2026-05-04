package oidc

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/execution"
	oapiv1 "github.com/rezible/rezible/openapi/v1"
)

type Config struct {
	SessionSecret       []byte     `koanf:"session_secret"`
	Oidc                oidcConfig `koanf:"oidc"`
	SingleTenantOrgName string     `koanf:"single_tenant_org_name"`
}

type oidcConfig struct {
	Issuer       string `koanf:"issuer"`
	ClientID     string `koanf:"client_id"`
	ClientSecret string `koanf:"client_secret"`
	RedirectUrl  string `koanf:"redirect_url"`
}

type AuthSessionService struct {
	orgs  rez.OrganizationService
	users rez.UserService

	cfg        Config
	cookiePath string
	codec      *cookieCodec
	oauth      *oauthHandler
}

var _ rez.AuthSessionService = (*AuthSessionService)(nil)

func NewAuthSessionService(ctx context.Context, orgs rez.OrganizationService, users rez.UserService) (*AuthSessionService, error) {
	oauthRedirectUrl, redirectErr := url.JoinPath(rez.Config.AppUrl(), "/api/auth/callback")
	if redirectErr != nil {
		return nil, fmt.Errorf("redirect url: %w", redirectErr)
	}

	cfg := Config{
		SingleTenantOrgName: "Default",
		Oidc: oidcConfig{
			RedirectUrl: oauthRedirectUrl,
		},
	}
	if cfgErr := rez.Config.Unmarshal("auth", &cfg); cfgErr != nil {
		return nil, fmt.Errorf("config: %w", cfgErr)
	}

	codec, codecErr := newCookieCodec(cfg.SessionSecret)
	if codecErr != nil {
		return nil, fmt.Errorf("cookie codec: %w", codecErr)
	}

	oauth, oauthErr := makeOAuthHandler(ctx, cfg, codec)
	if oauthErr != nil {
		return nil, fmt.Errorf("oauth handler: %w", oauthErr)
	}

	s := &AuthSessionService{
		orgs:       orgs,
		users:      users,
		cfg:        cfg,
		cookiePath: "/api",
		codec:      codec,
		oauth:      oauth,
	}

	return s, nil
}

func (s *AuthSessionService) AuthHandler() http.Handler {
	r := chi.NewRouter()
	r.Get("/login", handleAndRedirect(s.handleLogin))
	r.Get("/callback", handleAndRedirect(s.handleCallback))
	r.Get("/logout", handleAndRedirect(s.handleLogout))
	r.NotFound(http.RedirectHandler("/", http.StatusFound).ServeHTTP)
	return r
}

const (
	authStateCookieName = "rez_auth_state"
)

var (
	errRedirect         = fmt.Errorf("create_redirect")
	errWriteAuthSession = fmt.Errorf("write_auth_session")
	errWriteAuthState   = fmt.Errorf("write_auth_state")
	errReadAuthState    = fmt.Errorf("read_auth_state")
	errCallbackExchange = fmt.Errorf("callback_exchange")
	errIdentitySync     = fmt.Errorf("identity_sync")
)

func handleAndRedirect(handler func(http.ResponseWriter, *http.Request) (string, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		redirectUrl, err := handler(w, r)
		if err != nil {
			redirectUrl = fmt.Sprintf("/login?error=%s", url.QueryEscape(err.Error()))
		}
		http.Redirect(w, r, redirectUrl, http.StatusFound)
	}
}

func (s *AuthSessionService) handleLogin(w http.ResponseWriter, r *http.Request) (string, error) {
	authUrl, vs, authErr := s.oauth.createAuthRedirect(r)
	if authErr != nil {
		slog.Debug("Failed to create auth redirect", "error", authErr)
		return "", errRedirect
	}
	if cookieErr := s.writeCookie(w, authStateCookieName, vs, 10*time.Minute); cookieErr != nil {
		slog.Debug("Failed to write auth state cookie", "error", cookieErr)
		return "", errWriteAuthState
	}
	return authUrl, nil
}

func (s *AuthSessionService) handleCallback(w http.ResponseWriter, r *http.Request) (string, error) {
	var as AuthFlowState
	stateCookie, readCookieErr := r.Cookie(authStateCookieName)
	if stateCookie == nil || readCookieErr != nil {
		return "", errReadAuthState
	}
	if decodeErr := s.codec.decode(stateCookie.Value, &as); decodeErr != nil {
		return "", errReadAuthState
	}
	s.clearCookie(w, authStateCookieName)

	info, callbackErr := s.oauth.doCallbackExchange(r, as)
	if callbackErr != nil {
		slog.Debug("callback exchange", "error", callbackErr)
		return "", errCallbackExchange
	}

	usr, usrErr := s.users.SyncFromAuthProvider(r.Context(), info.org, info.user)
	if usrErr != nil {
		slog.Debug("user auth sync", "error", usrErr)
		return "", errIdentitySync
	}

	sess := rez.AuthSession{
		UserId:    usr.ID,
		ExpiresAt: info.expiresAt,
		Scopes:    []string{},
	}

	if cookieErr := s.writeCookie(w, oapiv1.AppCookieName, sess, time.Until(sess.ExpiresAt)); cookieErr != nil {
		return "", errWriteAuthSession
	}

	return as.ReturnTo, nil
}

func (s *AuthSessionService) handleLogout(w http.ResponseWriter, r *http.Request) (string, error) {
	s.clearCookie(w, oapiv1.AppCookieName)
	return "/login", nil
}

func (s *AuthSessionService) writeCookie(w http.ResponseWriter, name string, value any, maxAge time.Duration) error {
	encoded, encErr := s.codec.encode(value)
	if encErr != nil {
		return fmt.Errorf("encode value: %w", encErr)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    encoded,
		Path:     s.cookiePath,
		MaxAge:   int(maxAge.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func (s *AuthSessionService) clearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     s.cookiePath,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *AuthSessionService) Authenticate(ctx context.Context, token string) (context.Context, error) {
	if token == "" {
		return nil, rez.ErrAuthSessionMissing
	}

	var sess rez.AuthSession
	if decodeErr := s.codec.decode(token, &sess); decodeErr != nil {
		slog.Debug("decoding auth session cookie token", "error", decodeErr)
		return nil, rez.ErrAuthSessionInvalid
	}

	if sess.ExpiresAt.Before(time.Now()) {
		return nil, rez.ErrAuthSessionExpired
	}

	usr, lookupErr := s.users.Get(execution.SystemContext(ctx), user.ID(sess.UserId))
	if lookupErr != nil {
		slog.Debug("get user", "error", lookupErr, "sess", sess)
		return nil, rez.ErrAuthSessionInvalid
	}

	return execution.UserContext(ctx, *usr, &sess), nil
}
