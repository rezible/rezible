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
	SessionSecret []byte     `koanf:"session_secret"`
	Oidc          oidcConfig `koanf:"oidc"`
}

type oidcConfig struct {
	Issuer       string `koanf:"issuer"`
	ClientID     string `koanf:"client_id"`
	ClientSecret string `koanf:"client_secret"`
}

type AuthService struct {
	orgs  rez.OrganizationService
	users rez.UserService

	cfg     Config
	cookies *cookieWriter
	oauth   *oauthHandler
}

var _ rez.AuthService = (*AuthService)(nil)

func NewAuthService(ctx context.Context, orgs rez.OrganizationService, users rez.UserService) (*AuthService, error) {
	cfg := Config{
		Oidc: oidcConfig{},
	}
	if cfgErr := rez.Config.Unmarshal("auth", &cfg); cfgErr != nil {
		return nil, cfgErr
	}

	cookies, cookieErr := newCookieWriter(cfg.SessionSecret)
	if cookieErr != nil {
		return nil, fmt.Errorf("cookie codec error: %w", cookieErr)
	}

	s := &AuthService{
		orgs:    orgs,
		users:   users,
		cfg:     cfg,
		cookies: cookies,
	}

	return s, nil
}

func (s *AuthService) getOAuthHandler(ctx context.Context) (*oauthHandler, error) {
	if s.oauth == nil {
		o, oauthErr := makeOAuthHandler(ctx, s.cfg.Oidc)
		if oauthErr != nil {
			return nil, fmt.Errorf("make oauth handler: %w", oauthErr)
		}
		s.oauth = o
	}
	return s.oauth, nil
}

func (s *AuthService) Handler() http.Handler {
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

func (s *AuthService) handleLogin(w http.ResponseWriter, r *http.Request) (string, error) {
	oauth, oauthErr := s.getOAuthHandler(r.Context())
	if oauthErr != nil {
		return "", fmt.Errorf("oauth_handler")
	}
	authUrl, vs, authErr := oauth.createAuthRedirect()
	if authErr != nil {
		return "", fmt.Errorf("create_redirect")
	}
	if cookieErr := s.cookies.write(w, authVerificationCookieName, vs, 10*time.Minute); cookieErr != nil {
		return "", fmt.Errorf("write_cookie")
	}
	return authUrl, nil
}

func (s *AuthService) handleCallback(w http.ResponseWriter, r *http.Request) (string, error) {
	var as AuthVerifyState
	if cookieErr := s.cookies.read(r, authVerificationCookieName, &as); cookieErr != nil {
		return "", fmt.Errorf("read_auth_state")
	}
	s.cookies.clear(w, authVerificationCookieName)

	sess, sessErr := s.getTokenAuthSession(r, as)
	if sessErr != nil {
		log.Debug().Err(sessErr).Msg("get token auth session")
		return "", fmt.Errorf("create_session")
	}

	if cookieErr := s.cookies.write(w, oapiv1.AppCookieName, sess, time.Until(sess.ExpiresAt)); cookieErr != nil {
		return "", fmt.Errorf("set_cookie")
	}

	return as.RedirectURL, nil
}

func (s *AuthService) handleLogout(w http.ResponseWriter, r *http.Request) (string, error) {
	s.cookies.clear(w, oapiv1.AppCookieName)
	return "/login", nil
}

func (s *AuthService) getTokenAuthSession(r *http.Request, as AuthVerifyState) (*rez.AuthSession, error) {
	ctx := r.Context()
	oauth, oauthErr := s.getOAuthHandler(ctx)
	if oauthErr != nil {
		return nil, oauthErr
	}

	token, tokenErr := oauth.doCallbackTokenExchange(r, as.State, as.CodeVerifier)
	if tokenErr != nil {
		return nil, fmt.Errorf("exchange token: %w", tokenErr)
	}

	idTokenStr, idOk := token.Extra("id_token").(string)
	if !idOk {
		return nil, fmt.Errorf("no id_token")
	}

	id, idErr := oauth.verifyIdToken(ctx, idTokenStr, as.Nonce)
	if idErr != nil {
		return nil, fmt.Errorf("verify id token: %w", idErr)
	}

	usr, usrErr := s.users.SyncFromAuthProvider(ctx, id.getOrganization(), id.getUser())
	if usrErr != nil {
		return nil, fmt.Errorf("sync user: %w", usrErr)
	}

	sess := rez.AuthSession{
		UserId:    usr.ID,
		Scopes:    []string{},
		ExpiresAt: token.Expiry,
	}
	return &sess, nil
}

type authUserSessionContextKey struct{}

func (s *AuthService) SetAuthSessionContextFromAppCookie(ctx context.Context, cookieStr string) (context.Context, error) {
	if cookieStr == "" {
		return nil, rez.ErrAuthSessionMissing
	}
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

func (s *AuthService) SetAuthSessionContextFromApiToken(ctx context.Context, tokenStr string) (context.Context, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *AuthService) makeAuthSessionContext(ctx context.Context, u *ent.User, sess rez.AuthSession) context.Context {
	return context.WithValue(access.WithUser(ctx, u), authUserSessionContextKey{}, sess)
}

func (s *AuthService) GetAuthSession(ctx context.Context) rez.AuthSession {
	if sess, ok := ctx.Value(authUserSessionContextKey{}).(rez.AuthSession); ok {
		return sess
	}
	return rez.AuthSession{}
}
