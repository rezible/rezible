package oidc

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type AuthSessionProvider struct {
	sessionStore sessions.Store
	oauth2Config oauth2.Config
	verifier     *oidc.IDTokenVerifier
	idp          rez.OIDCAuthSessionIdentityProvider
	flowPath     string
	flowHandler  http.Handler
}

func LoadAuthSessionProvider(ctx context.Context, idp rez.OIDCAuthSessionIdentityProvider) (*AuthSessionProvider, error) {
	sp := &AuthSessionProvider{
		idp:          idp,
		flowPath:     "/oidc/" + idp.Id(),
		sessionStore: configureSessionStore(),
	}

	redirectUrl, urlErr := rez.Config.GetMountedAppRoute(rez.Config.AuthPath(), "flow", sp.FlowPath(), "callback")
	if urlErr != nil {
		return nil, fmt.Errorf("creating redirect url: %w", urlErr)
	}

	v, ocfg, loadErr := idp.LoadConfig(ctx, redirectUrl)
	if loadErr != nil {
		return nil, fmt.Errorf("failed to load oidc idp config: %w", loadErr)
	}
	sp.oauth2Config = *ocfg
	sp.verifier = v

	return sp, nil
}

func (s *AuthSessionProvider) Id() string { return s.idp.Id() }

func (s *AuthSessionProvider) DisplayName() string {
	return s.idp.DisplayName()
}

var userMapping = &ent.User{
	Name:  "y",
	Email: "y",
}

func (s *AuthSessionProvider) UserMapping() *ent.User {
	return userMapping
}

func (s *AuthSessionProvider) setSession(w http.ResponseWriter, r *http.Request, sess *session) error {
	return setSession(w, r, s.sessionStore, s.idp.Id(), sess)
}

func (s *AuthSessionProvider) getSession(r *http.Request) (*session, error) {
	return getSession(r, s.sessionStore, s.idp.Id())
}

func (s *AuthSessionProvider) clearSession(w http.ResponseWriter, r *http.Request) error {
	return clearSession(w, r, s.sessionStore)
}

func (s *AuthSessionProvider) SessionExists(r *http.Request) bool {
	sess, sessErr := s.getSession(r)
	return sess == nil || sessErr != nil
}

func (s *AuthSessionProvider) ClearSession(w http.ResponseWriter, r *http.Request) error {
	return s.clearSession(w, r)
}

func getRequestState(r *http.Request) string {
	reqState := r.URL.Query().Get("state")
	if reqState == "" && r.Method == http.MethodPost {
		reqState = r.FormValue("state")
	}
	return reqState
}

func generateRequestState() (string, error) {
	nonceBytes := make([]byte, 64)
	_, err := io.ReadFull(rand.Reader, nonceBytes)
	if err != nil {
		return "", fmt.Errorf("generating nonce: %w", err)
	}
	return base64.URLEncoding.EncodeToString(nonceBytes), nil
}

func (s *AuthSessionProvider) FlowPath() string {
	return s.flowPath
}

func (s *AuthSessionProvider) MakeFlowPathHandler(onCreated rez.AuthSessionCreatedCallback) http.Handler {
	r := chi.NewRouter()
	r.Get("/", s.handleStartAuthFlow)
	r.Get("/callback", func(w http.ResponseWriter, r *http.Request) {
		provSess, cbErr := s.handleFlowCallback(r)
		if cbErr != nil {
			log.Debug().Err(cbErr).Msg("oidc callback error")
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		} else {
			onCreated(w, r, provSess)
		}
	})
	return r
}

func (s *AuthSessionProvider) handleStartAuthFlow(w http.ResponseWriter, r *http.Request) {
	state := getRequestState(r)
	if state == "" {
		var stateErr error
		state, stateErr = generateRequestState()
		if stateErr != nil {
			log.Error().Err(stateErr).Msg("could not create redirect state")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	authUrl := s.oauth2Config.AuthCodeURL(state, s.idp.GetAuthCodeOptions(r)...)
	sess := &session{State: state}

	if setStateErr := s.setSession(w, r, sess); setStateErr != nil {
		log.Error().Err(setStateErr).Msg("could not set session redirect state")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, authUrl, http.StatusFound)
}

func (s *AuthSessionProvider) handleFlowCallback(r *http.Request) (*rez.AuthProviderSession, error) {
	sess, sessErr := s.getSession(r)
	if sessErr != nil || sess == nil || sess.State == "" {
		return nil, fmt.Errorf("get session: %w", sessErr)
	}

	if rs := getRequestState(r); rs != sess.State {
		log.Warn().Msg("invalid request session state")
		if !rez.Config.DebugMode() {
			return nil, errors.New("state mismatch")
		}
	}

	token, tokenErr := s.exchangeAndVerifyIdToken(r)
	if tokenErr != nil {
		return nil, fmt.Errorf("failed to get id token: %w", tokenErr)
	}

	provSess, psErr := s.idp.ExtractTokenSession(token)
	if psErr != nil {
		return nil, fmt.Errorf("failed to extract token session: %w", psErr)
	}

	return provSess, nil
}

func (s *AuthSessionProvider) exchangeAndVerifyIdToken(r *http.Request) (*oidc.IDToken, error) {
	ctx := r.Context()
	authCode := r.URL.Query().Get("code")

	token, exchangeErr := s.oauth2Config.Exchange(ctx, authCode)
	if exchangeErr != nil {
		return nil, fmt.Errorf("failed to exchange authorization code for access token: %w", exchangeErr)
	}

	rawIDToken, idTokenOk := token.Extra("id_token").(string)
	if !idTokenOk {
		return nil, fmt.Errorf("no id_token field in oauth2 token")
	}

	idToken, verifyErr := s.verifier.Verify(ctx, rawIDToken)
	if verifyErr != nil {
		return nil, fmt.Errorf("failed to verify id token: %w", verifyErr)
	}

	return idToken, nil
}
