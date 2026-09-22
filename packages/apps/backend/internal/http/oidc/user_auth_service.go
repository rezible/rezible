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

type UserAuthProvider struct {
	sess            rez.AuthSessionService
	cookie          rez.AppAuthSessionCookie
	oidc            *oidcHandler
	singleTenantOrg *ent.Organization
}

func NewUserAuthProvider(cfg rez.Config, sess rez.AuthSessionService, cookie rez.AppAuthSessionCookie) (*UserAuthProvider, error) {
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

	s := &UserAuthProvider{
		sess:            sess,
		cookie:          cookie,
		oidc:            oh,
		singleTenantOrg: singleTenantOrg,
	}

	return s, nil
}

func (s *UserAuthProvider) MakeAuthHandler() http.Handler {
	r := chi.NewRouter()
	r.Get("/login", s.handleAndRedirect(s.handleLogin))
	r.Get("/callback", s.handleAndRedirect(s.handleCallback))
	r.Get("/logout", s.handleAndRedirect(s.handleLogout))
	r.NotFound(http.RedirectHandler("/", http.StatusFound).ServeHTTP)
	return r
}

type redirectingHandlerFn = func(w http.ResponseWriter, r *http.Request) (redirectTo string, err error)

func (s *UserAuthProvider) handleAndRedirect(handlerFn redirectingHandlerFn) http.HandlerFunc {
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

func (s *UserAuthProvider) handleLogin(w http.ResponseWriter, r *http.Request) (string, error) {
	authUrl, authErr := s.oidc.createAuthRedirect(w, r)
	if authErr != nil {
		slog.Debug("Failed to create auth redirect", "error", authErr)
		return "", errCreateRedirect
	}
	return authUrl, nil
}

func (s *UserAuthProvider) handleCallback(w http.ResponseWriter, r *http.Request) (string, error) {
	res, callbackErr := s.oidc.doCallbackExchange(w, r)
	if callbackErr != nil {
		slog.Debug("callback exchange", "error", callbackErr)
		return "", errCallbackExchange
	}

	// TODO: move this out of here
	if s.singleTenantOrg != nil {
		slog.Debug("using single tenant organization")
		res.Session.Org = *s.singleTenantOrg
	}

	sess, sessErr := s.sess.CreateFromUserAuthResponse(r.Context(), res.Session)
	if sessErr != nil {
		slog.Debug("user session create", "error", sessErr)
		return "", errCreateAuthSession
	}
	s.cookie.Set(w, sess)

	return res.ReturnTo, nil
}

func (s *UserAuthProvider) handleLogout(w http.ResponseWriter, r *http.Request) (string, error) {
	ctx := r.Context()
	if clearErr := s.oidc.doLogout(ctx); clearErr != nil {
		slog.ErrorContext(ctx, "oidc logout", "error", clearErr)
	}
	s.cookie.Clear(w)
	return "/login", nil
}
