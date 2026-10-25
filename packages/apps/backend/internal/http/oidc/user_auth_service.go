package oidc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/oauth2"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/pkg/errs"
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
	r.Post("/refresh", s.handleRefresh)
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
		slog.WarnContext(r.Context(), "Sign-in refused", refusalAttrs(callbackErr)...)
		return "", errCallbackExchange
	}
	s.setSingleTenantOrg(res.Session)

	sess, sessErr := s.sess.CreateFromUserAuthResponse(r.Context(), res.Session)
	if sessErr != nil {
		slog.Debug("user session create", "error", sessErr)
		return "", errCreateAuthSession
	}
	s.cookie.Set(w, sess)

	return res.ReturnTo, nil
}

// setSingleTenantOrg places a single-tenant session in the configured organization.
func (s *UserAuthProvider) setSingleTenantOrg(ps *rez.UserAuthProviderSession) {
	if s.singleTenantOrg != nil {
		ps.Org = *s.singleTenantOrg
	}
}

func (s *UserAuthProvider) refreshExchange(ctx context.Context, refreshToken string) (*rez.UserAuthProviderSession, error) {
	ps, refreshErr := s.oidc.refreshSession(ctx, refreshToken)
	if refreshErr != nil {
		return nil, refreshErr
	}
	s.setSingleTenantOrg(ps)
	return ps, nil
}

const refreshTimeout = 10 * time.Second

// handleRefresh refreshes the session named by the session cookie. It answers 204 with the cookie set again, 401
// with the cookie cleared when the session is missing, expired or refused, and 503 when the refresh failed
// otherwise, such as when the provider could not be reached.
func (s *UserAuthProvider) handleRefresh(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sessionID, cookieErr := s.cookie.Get(r)
	if cookieErr != nil || sessionID == uuid.Nil {
		s.cookie.Clear(w)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// A hung provider would otherwise hold the session's row lock, and every tab waiting on it, open.
	refreshCtx, cancel := context.WithTimeout(ctx, refreshTimeout)
	defer cancel()
	sess, refreshErr := s.sess.Refresh(refreshCtx, sessionID, s.refreshExchange)
	if refreshErr != nil {
		if errors.Is(refreshErr, errs.ErrAuthSessionInvalid) || errors.Is(refreshErr, errs.ErrAuthSessionExpired) {
			slog.WarnContext(ctx, "Session refresh refused", refusalAttrs(refreshErr)...)
			s.cookie.Clear(w)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		slog.ErrorContext(ctx, "Session refresh failed", refusalAttrs(refreshErr)...)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	s.cookie.Set(w, sess)
	w.WriteHeader(http.StatusNoContent)
}

// refusalAttrs are the log attributes of a refused sign-in or refresh, with the provider's error when it sent one.
func refusalAttrs(refusal error) []any {
	attrs := []any{"error", refusal}
	if retrieveErr, isRetrieveErr := errors.AsType[*oauth2.RetrieveError](refusal); isRetrieveErr {
		attrs = append(attrs, "oauth_error", retrieveErr.ErrorCode, "oauth_error_description", retrieveErr.ErrorDescription)
	}
	return attrs
}

func (s *UserAuthProvider) handleLogout(w http.ResponseWriter, r *http.Request) (string, error) {
	ctx := r.Context()
	if clearErr := s.oidc.doLogout(ctx); clearErr != nil {
		slog.ErrorContext(ctx, "oidc logout", "error", clearErr)
	}
	s.cookie.Clear(w)
	return "/login", nil
}
