package oidc

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type authSessionCookieWriter interface {
	Set(w http.ResponseWriter, sess *ent.UserAuthSession)
	Clear(w http.ResponseWriter)
}

type UserAuthService struct {
	auth            rez.AuthSessionService
	scw             authSessionCookieWriter
	oidc            *oidcHandler
	singleTenantOrg *ent.Organization
}

func NewUserAuthHandler(cfg rez.Config, authSess rez.AuthSessionService, appCookie authSessionCookieWriter) (http.Handler, error) {
	oh, ohErr := newOidcHandler(cfg)
	if ohErr != nil {
		return nil, fmt.Errorf("oidc: %w", ohErr)
	}

	var singleTenantOrg *ent.Organization
	if cfg.App.SingleTenant.Enabled {
		singleTenantOrg = &ent.Organization{
			AuthProviderID: "default",
			Name:           cfg.App.SingleTenant.OrgName,
		}
	}

	s := &UserAuthService{
		auth:            authSess,
		scw:             appCookie,
		oidc:            oh,
		singleTenantOrg: singleTenantOrg,
	}

	return s.Handler(), nil
}

func (s *UserAuthService) Handler() http.Handler {
	r := chi.NewRouter()
	r.Get("/login", s.handleAndRedirect(s.handleLogin))
	r.Get("/callback", s.handleAndRedirect(s.handleCallback))
	r.Get("/logout", s.handleAndRedirect(s.handleLogout))
	r.NotFound(http.RedirectHandler("/", http.StatusFound).ServeHTTP)
	return r
}

type redirectingHandlerFn = func(w http.ResponseWriter, r *http.Request) (redirectTo string, err error)

func (s *UserAuthService) handleAndRedirect(handlerFn redirectingHandlerFn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		redirectUrl, err := handlerFn(w, r)
		if err != nil {
			redirectUrl = fmt.Sprintf("/login?error=%s", url.QueryEscape(err.Error()))
		}
		http.Redirect(w, r, redirectUrl, http.StatusFound)
	}
}

var (
	errCreateRedirect    = fmt.Errorf("create_redirect")
	errWriteAuthState    = fmt.Errorf("write_auth_state")
	errReadAuthState     = fmt.Errorf("read_auth_state")
	errCallbackExchange  = fmt.Errorf("callback_exchange")
	errCreateAuthSession = fmt.Errorf("identity_sync")
)

func (s *UserAuthService) handleLogin(w http.ResponseWriter, r *http.Request) (string, error) {
	authUrl, authErr := s.oidc.createAuthRedirect(w, r)
	if authErr != nil {
		slog.Debug("Failed to create auth redirect", "error", authErr)
		return "", errCreateRedirect
	}
	return authUrl, nil
}

func (s *UserAuthService) handleCallback(w http.ResponseWriter, r *http.Request) (string, error) {
	res, callbackErr := s.oidc.doCallbackExchange(w, r)
	if callbackErr != nil {
		slog.Debug("callback exchange", "error", callbackErr)
		return "", errCallbackExchange
	}

	if s.singleTenantOrg != nil {
		slog.Debug("using single tenant organization")
		res.Session.Org = *s.singleTenantOrg
	}

	sess, sessErr := s.auth.CreateFromUserAuthResponse(r.Context(), res.Session)
	if sessErr != nil {
		slog.Debug("user session create", "error", sessErr)
		return "", errCreateAuthSession
	}
	s.scw.Set(w, sess)

	return res.ReturnTo, nil
}

func (s *UserAuthService) handleLogout(w http.ResponseWriter, r *http.Request) (string, error) {
	ctx := r.Context()
	if clearErr := s.oidc.doLogout(ctx); clearErr != nil {
		slog.ErrorContext(ctx, "oidc logout", "error", clearErr)
	}
	s.scw.Clear(w)
	return "/login", nil
}
