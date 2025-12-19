package oidc

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gorilla/sessions"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type identityProvider interface {
	GetAuthCodeOptions(r *http.Request) []oauth2.AuthCodeOption
	ExtractTokenSession(token *oidc.IDToken) (*rez.AuthProviderSession, error)
}

type AuthSessionProvider struct {
	callbackPath string
	providerId   string
	displayName  string
	sessionStore sessions.Store
	oauth2Config oauth2.Config
	verifier     *oidc.IDTokenVerifier
	idp          identityProvider
}

func (s *AuthSessionProvider) DisplayName() string {
	return s.displayName
}

func (s *AuthSessionProvider) Id() string {
	return s.providerId
}

var userMapping = &ent.User{
	Name:  "y",
	Email: "y",
}

func (s *AuthSessionProvider) UserMapping() *ent.User {
	return userMapping
}

func (s *AuthSessionProvider) setSession(w http.ResponseWriter, r *http.Request, sess *session) error {
	return setSession(w, r, s.sessionStore, s.providerId, sess)
}

func (s *AuthSessionProvider) getSession(r *http.Request) (*session, error) {
	return getSession(r, s.sessionStore, s.providerId)
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

func (s *AuthSessionProvider) HandleStartAuthFlow(w http.ResponseWriter, r *http.Request) {
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

func (s *AuthSessionProvider) HandleAuthFlowRequest(w http.ResponseWriter, r *http.Request, onCreated func(rez.AuthProviderSession)) bool {
	if r.URL.Path == s.callbackPath {
		if cbErr := s.handleFlowCallback(r, onCreated); cbErr != nil {
			log.Debug().Err(cbErr).Msg("oidc callback error")
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		}
		return true
	}
	return false
}

func (s *AuthSessionProvider) handleFlowCallback(r *http.Request, onCreated func(rez.AuthProviderSession)) error {
	sess, sessErr := s.getSession(r)
	if sessErr != nil || sess == nil || sess.State == "" {
		return fmt.Errorf("get session: %w", sessErr)
	}

	if rs := getRequestState(r); rs != sess.State {
		log.Warn().Msg("invalid request session state")
		if !rez.Config.DebugMode() {
			return errors.New("state mismatch")
		}
	}

	token, tokenErr := s.exchangeAndVerifyIdToken(r)
	if tokenErr != nil {
		return fmt.Errorf("failed to get id token: %w", tokenErr)
	}

	provSess, psErr := s.idp.ExtractTokenSession(token)
	if psErr != nil {
		return fmt.Errorf("failed to extract token session: %w", psErr)
	}

	onCreated(*provSess)

	return nil
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
