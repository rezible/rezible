package oidc

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"golang.org/x/oauth2"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type AppCookieWriter interface {
	Set(w http.ResponseWriter, sess *ent.UserAuthSession)
	Get(r *http.Request) *http.Cookie
	Clear(w http.ResponseWriter)
}

type UserAuthService struct {
	authSess  rez.AuthSessionService
	appCookie AppCookieWriter

	oauth *oauthHandler
}

func NewUserAuthHandler(cfg rez.Config, authSess rez.AuthSessionService, appCookie AppCookieWriter) (http.Handler, error) {
	oauthRedirectUrl := cfg.HttpServer.Auth.Oidc.RedirectUrl
	if oauthRedirectUrl == "" {
		feRedirectUrl, urlErr := cfg.App.GetFrontendUrl(cfg.App.FrontendApiPath, "/auth/callback")
		if urlErr != nil {
			return nil, fmt.Errorf("oauth redirect url: %w", urlErr)
		}
		oauthRedirectUrl = feRedirectUrl.String()
	}

	codec, codecErr := newCookieCodec(cfg.HttpServer.Auth.SessionSecret)
	if codecErr != nil {
		return nil, fmt.Errorf("cookie codec: %w", codecErr)
	}

	apiAudience := cfg.App.ApiDomain
	if apiAudience == "" {
		return nil, fmt.Errorf("no api url configured, can't verify token audience")
	}
	oauth := &oauthHandler{
		cfg:            cfg.HttpServer.Auth.Oidc,
		redirectUrl:    oauthRedirectUrl,
		codec:          codec,
		apiAudience:    apiAudience,
		resourceOption: oauth2.SetAuthURLParam("resource", apiAudience),
	}
	if cfg.App.SingleTenant.Enabled {
		oauth.singleTenantOrg = &ent.Organization{
			AuthProviderID: "default",
			Name:           cfg.App.SingleTenant.OrgName,
		}
	}

	s := &UserAuthService{
		authSess:  authSess,
		appCookie: appCookie,
		oauth:     oauth,
	}

	return s.Handler(), nil
}

func (s *UserAuthService) Handler() http.Handler {
	r := chi.NewRouter()
	r.Get("/login", handleAndRedirect(s.handleLogin))
	r.Get("/callback", handleAndRedirect(s.handleCallback))
	r.Get("/logout", handleAndRedirect(s.handleLogout))
	r.NotFound(http.RedirectHandler("/", http.StatusFound).ServeHTTP)
	return r
}

var (
	errRedirect         = fmt.Errorf("create_redirect")
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

func (s *UserAuthService) handleLogin(w http.ResponseWriter, r *http.Request) (string, error) {
	authUrl, authErr := s.oauth.createAuthRedirect(w, r)
	if authErr != nil {
		slog.Debug("Failed to create auth redirect", "error", authErr)
		return "", errRedirect
	}
	return authUrl, nil
}

func (s *UserAuthService) handleCallback(w http.ResponseWriter, r *http.Request) (string, error) {
	info, returnTo, callbackErr := s.oauth.doCallbackExchange(w, r)
	if callbackErr != nil {
		slog.Debug("callback exchange", "error", callbackErr)
		return "", errCallbackExchange
	}

	sess, sessErr := s.authSess.SyncFromAuthProvider(r.Context(), info.user, info.org)
	if sessErr != nil {
		slog.Debug("user session create", "error", sessErr)
		return "", errIdentitySync
	}
	s.appCookie.Set(w, sess)

	return returnTo, nil
}

func (s *UserAuthService) handleLogout(w http.ResponseWriter, r *http.Request) (string, error) {
	s.appCookie.Clear(w)
	fmt.Printf("clear app cookie\n")
	return "/login", nil
}
