package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/execution"
	oapiv1 "github.com/rezible/rezible/openapi/v1"
)

type Config struct {
	SessionSecret       []byte     `cfg:"session_secret"`
	Oidc                oidcConfig `cfg:"oidc"`
	SingleTenantOrgName string     `cfg:"single_tenant_org_name"`
}

type oidcConfig struct {
	Issuer       string `cfg:"issuer"`
	ClientID     string `cfg:"client_id"`
	ClientSecret string `cfg:"client_secret"`
	RedirectUrl  string `cfg:"redirect_url"`
}

type AuthSessionService struct {
	orgs  rez.OrganizationService
	users rez.UserService

	cfg        Config
	cookiePath string
	codec      *cookieCodec
	oauth      *oauthHandler
}

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

func (s *AuthSessionService) Handler() http.Handler {
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

type AppAuthCookie struct {
	TenantId  int       `json:"tenant_id"`
	UserId    uuid.UUID `json:"uid"`
	ExpiresAt time.Time `json:"exp"`
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

	sess := AppAuthCookie{
		UserId:    usr.ID,
		ExpiresAt: info.expiresAt,
	}

	if cookieErr := s.writeCookie(w, oapiv1.AppCookieName, sess, time.Until(sess.ExpiresAt)); cookieErr != nil {
		return "", errWriteAuthSession
	}

	return as.ReturnTo, nil
}

func (s *AuthSessionService) ExecutionContextMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authedReq, authErr := s.setRequestExecutionContext(r)
			if authErr == nil {
				next.ServeHTTP(w, authedReq)
				return
			}
			apiErr := oapiv1.ConvertAuthStatusError(authErr)
			w.WriteHeader(apiErr.GetStatus())
			if respErr := json.NewEncoder(w).Encode(apiErr); respErr != nil {
				slog.Warn("failed to write api error response", "error", respErr)
			}
		})
	}
}

func (s *AuthSessionService) setRequestExecutionContext(r *http.Request) (*http.Request, error) {
	if apiToken := oapiv1.GetRequestApiTokenValue(r); apiToken != "" {
		return nil, rez.ErrAuthSessionInvalid
	}
	if appCookie := oapiv1.GetRequestAppCookieValue(r); appCookie != "" {
		execCtx, authErr := s.authenticateAppCookie(r.Context(), appCookie)
		if authErr != nil {
			return nil, authErr
		}
		return r.WithContext(execCtx), nil
	}
	return nil, rez.ErrAuthSessionMissing
}

func (s *AuthSessionService) authenticateAppCookie(ctx context.Context, appCookie string) (context.Context, error) {
	var sess AppAuthCookie
	if decodeErr := s.codec.decode(appCookie, &sess); decodeErr != nil {
		slog.Debug("decoding auth session cookie token", "error", decodeErr)
		return nil, rez.ErrAuthSessionInvalid
	}

	if sess.ExpiresAt.Before(time.Now()) {
		return nil, rez.ErrAuthSessionExpired
	}

	lookupCtx := execution.NewTenantContext(ctx, sess.TenantId)
	usr, lookupErr := s.users.Get(lookupCtx, user.ID(sess.UserId))
	if lookupErr != nil {
		slog.Debug("get user", "error", lookupErr, "sess", sess)
		return nil, rez.ErrAuthSessionInvalid
	}

	return execution.NewUserAuthContext(ctx, *usr, sess.ExpiresAt), nil
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
