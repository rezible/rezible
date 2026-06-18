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
	authSess   rez.AuthSessionService
	authCookie authSessionCookieWriter

	oauth *oauthHandler
}

func NewUserAuthHandler(cfg rez.Config, authSess rez.AuthSessionService, appCookie authSessionCookieWriter) (http.Handler, error) {
	oauth, oauthErr := newOAuthHandler(cfg)
	if oauthErr != nil {
		return nil, oauthErr
	}

	s := &UserAuthService{
		authSess:   authSess,
		authCookie: appCookie,
		oauth:      oauth,
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

func (s *UserAuthService) handleAndRedirect(handler func(http.ResponseWriter, *http.Request) (string, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		redirectUrl, err := handler(w, r)
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
	authUrl, authErr := s.oauth.createAuthRedirect(w, r)
	if authErr != nil {
		slog.Debug("Failed to create auth redirect", "error", authErr)
		return "", errCreateRedirect
	}
	return authUrl, nil
}

func (s *UserAuthService) handleCallback(w http.ResponseWriter, r *http.Request) (string, error) {
	ps, returnTo, callbackErr := s.oauth.doCallbackExchange(w, r)
	if callbackErr != nil {
		slog.Debug("callback exchange", "error", callbackErr)
		return "", errCallbackExchange
	}
	if ps == nil {
		slog.Warn("no auth provider session returned, no error?")
		return "", errCallbackExchange
	}

	sess, sessErr := s.authSess.CreateFromUserAuth(r.Context(), ps)
	if sessErr != nil {
		slog.Debug("user session create", "error", sessErr)
		return "", errCreateAuthSession
	}
	s.authCookie.Set(w, sess)

	return returnTo, nil
}

func (s *UserAuthService) handleLogout(w http.ResponseWriter, r *http.Request) (string, error) {
	s.authCookie.Clear(w)
	return "/login", nil
}
