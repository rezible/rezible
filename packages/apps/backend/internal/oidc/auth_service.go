package oidc

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/user"
	oapiv1 "github.com/rezible/rezible/openapi/v1"
	"github.com/rs/zerolog/log"
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

	cfg     Config
	cookies *cookieWriter
	oauth   *oauthHandler
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

	cookies, cookieErr := newCookieWriter(cfg.SessionSecret)
	if cookieErr != nil {
		return nil, fmt.Errorf("cookie writer: %w", cookieErr)
	}

	oauth, oauthErr := makeOAuthHandler(ctx, cfg)
	if oauthErr != nil {
		return nil, fmt.Errorf("oauth handler: %w", oauthErr)
	}

	s := &AuthSessionService{
		orgs:    orgs,
		users:   users,
		cfg:     cfg,
		cookies: cookies,
		oauth:   oauth,
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

const authVerificationCookieName = "rez_auth_verify"

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
		return "", fmt.Errorf("create_redirect")
	}
	if cookieErr := s.cookies.write(w, authVerificationCookieName, vs, 10*time.Minute); cookieErr != nil {
		return "", fmt.Errorf("write_cookie")
	}
	return authUrl, nil
}

func (s *AuthSessionService) handleCallback(w http.ResponseWriter, r *http.Request) (string, error) {
	var as AuthFlowState
	if cookieErr := s.cookies.read(r, authVerificationCookieName, &as); cookieErr != nil {
		return "", fmt.Errorf("read_auth_state")
	}
	s.cookies.clear(w, authVerificationCookieName)

	info, callbackErr := s.oauth.doCallbackExchange(r, as)
	if callbackErr != nil {
		log.Debug().Err(callbackErr).Msg("callback exchange")
		return "", fmt.Errorf("exchange")
	}

	usr, usrErr := s.users.SyncFromAuthProvider(r.Context(), info.org, info.user)
	if usrErr != nil {
		log.Debug().Err(usrErr).Msg("user auth sync")
		return "", fmt.Errorf("sync")
	}

	sess := rez.AuthSession{
		UserId:    usr.ID,
		ExpiresAt: info.expiresAt,
		Scopes:    []string{},
	}

	if cookieErr := s.cookies.write(w, oapiv1.AppCookieName, sess, time.Until(sess.ExpiresAt)); cookieErr != nil {
		return "", fmt.Errorf("set_cookie")
	}

	return as.ReturnTo, nil
}

func (s *AuthSessionService) handleLogout(w http.ResponseWriter, r *http.Request) (string, error) {
	s.cookies.clear(w, oapiv1.AppCookieName)
	return "/login", nil
}

func (s *AuthSessionService) SetAuthSessionContext(ctx context.Context, appCookie, apiToken string) (context.Context, error) {
	if appCookie != "" {
		return s.setContextFromAppCookie(ctx, appCookie)
	} else if apiToken != "" {
		return s.setContextFromApiToken(ctx, apiToken)
	}
	return nil, rez.ErrAuthSessionMissing
}

func (s *AuthSessionService) setContextFromAppCookie(ctx context.Context, cookieStr string) (context.Context, error) {
	var sess rez.AuthSession
	if decodeErr := s.cookies.decode(cookieStr, &sess); decodeErr != nil {
		log.Debug().Err(decodeErr).Msg("decoding auth session token")
		return nil, rez.ErrAuthSessionInvalid
	}
	usr, usrErr := s.users.Get(access.SystemContext(ctx), user.ID(sess.UserId))
	if usrErr != nil {
		log.Debug().Err(usrErr).Interface("sess", sess).Msg("get user")
		return nil, rez.ErrAuthSessionInvalid
	}
	return s.makeAuthSessionContext(ctx, usr, sess), nil
}

func (s *AuthSessionService) setContextFromApiToken(ctx context.Context, tokenStr string) (context.Context, error) {
	return nil, fmt.Errorf("not implemented")
}

type authUserSessionContextKey struct{}

func (s *AuthSessionService) makeAuthSessionContext(ctx context.Context, u *ent.User, sess rez.AuthSession) context.Context {
	return context.WithValue(access.WithUser(ctx, u), authUserSessionContextKey{}, sess)
}

func (s *AuthSessionService) GetAuthSession(ctx context.Context) rez.AuthSession {
	if sess, ok := ctx.Value(authUserSessionContextKey{}).(rez.AuthSession); ok {
		return sess
	}
	return rez.AuthSession{}
}
