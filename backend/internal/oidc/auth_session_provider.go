package oidc

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gorilla/sessions"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type Config struct {
	SessionSecret string
	ProviderName  string `json:"provider_name"`
	ClientID      string `json:"client_id"`
	ClientSecret  string `json:"client_secret"`
	IssuerUrl     string `json:"issuer_url"`
}

var userMapping = &ent.User{
	Name:  "y",
	Email: "y",
}

type AuthSessionProvider struct {
	callbackPath string
	providerId   string
	displayName  string
	sessionStore sessions.Store
	oauth2Config oauth2.Config
	verifier     *oidc.IDTokenVerifier
}

func NewAuthSessionProvider(ctx context.Context, cfg Config) (*AuthSessionProvider, error) {
	providerId := strings.ToLower(cfg.ProviderName)
	callbackPath := fmt.Sprintf("/auth/%s/%s", providerId, "callback")
	redirectUrl, urlErr := url.JoinPath(rez.BackendUrl, callbackPath)
	if urlErr != nil {
		return nil, fmt.Errorf("creating redirect url: %w", urlErr)
	}

	oidcProvider, oidcErr := oidc.NewProvider(ctx, cfg.IssuerUrl)
	if oidcErr != nil {
		return nil, fmt.Errorf("oidc.NewProvider: %w", oidcErr)
	}

	oauth2Config := oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  redirectUrl,
		Endpoint:     oidcProvider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return &AuthSessionProvider{
		providerId:   providerId,
		callbackPath: callbackPath,
		displayName:  cfg.ProviderName,
		oauth2Config: oauth2Config,
		verifier:     oidcProvider.Verifier(&oidc.Config{ClientID: cfg.ClientID}),
		sessionStore: configureSessionStore(cfg.SessionSecret),
	}, nil
}

func (s *AuthSessionProvider) DisplayName() string {
	return s.displayName
}

func (s *AuthSessionProvider) Id() string {
	return s.providerId
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

func (s *AuthSessionProvider) StartAuthFlow(w http.ResponseWriter, r *http.Request) {
	state, stateErr := getOrGenerateRequestState(r)
	if stateErr != nil {
		log.Error().Err(stateErr).Msg("could not create redirect state")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// TODO: PKCE / nonce gen?
	authUrl := s.oauth2Config.AuthCodeURL(state)
	log.Debug().Str("authUrl", authUrl).Msg("starting auth flow")

	sess := &session{
		State: state,
	}

	if setStateErr := s.setSession(w, r, sess); setStateErr != nil {
		log.Error().Err(setStateErr).Msg("could not set session redirect state")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, authUrl, http.StatusFound)
}

func getOrGenerateRequestState(r *http.Request) (string, error) {
	state := r.URL.Query().Get("state")
	if state != "" {
		return state, nil
	}

	nonceBytes := make([]byte, 64)
	_, err := io.ReadFull(rand.Reader, nonceBytes)
	if err != nil {
		return "", fmt.Errorf("generating nonce: %w", err)
	}
	return base64.URLEncoding.EncodeToString(nonceBytes), nil
}

func (s *AuthSessionProvider) HandleAuthFlowRequest(w http.ResponseWriter, r *http.Request, onCreated func(rez.AuthProviderSession)) bool {
	if r.URL.Path == s.callbackPath {
		if cbErr := s.handleFlowCallback(r, onCreated); cbErr != nil {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		}
		return true
	}
	return false
}

func (s *AuthSessionProvider) handleFlowCallback(r *http.Request, onCreated func(rez.AuthProviderSession)) error {
	if validateErr := s.validateRequestSessionState(r); validateErr != nil {
		log.Warn().Msg("invalid request session state")
		if !rez.DebugMode {
			return validateErr
		}
	}

	token, tokenErr := s.getVerifiedIdToken(r)
	if tokenErr != nil {
		return fmt.Errorf("failed to get id token: %w", tokenErr)
	}

	ps, psErr := s.extractTokenSession(token)
	if psErr != nil {
		return fmt.Errorf("failed to extract token session: %w", psErr)
	}

	onCreated(*ps)

	return nil
}

func (s *AuthSessionProvider) validateRequestSessionState(r *http.Request) error {
	sess, sessErr := s.getSession(r)
	if sessErr != nil || sess == nil || sess.State == "" {
		return fmt.Errorf("get session: %w", sessErr)
	}

	params := r.URL.Query()
	reqState := params.Get("state")
	if reqState == "" && r.Method == http.MethodPost {
		reqState = r.FormValue("state")
	}

	if reqState != sess.State {
		return errors.New("state mismatch")
	}
	return nil
}

func (s *AuthSessionProvider) getVerifiedIdToken(r *http.Request) (*oidc.IDToken, error) {
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

func (s *AuthSessionProvider) extractTokenSession(token *oidc.IDToken) (*rez.AuthProviderSession, error) {
	// TODO: use different claims depending on issuer

	var claims struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Verified bool   `json:"email_verified"`
		Locale   string `json:"locale"`
		Nonce    string `json:"nonce"`
		TenantId string `json:"hd"`
	}
	if claimsErr := token.Claims(&claims); claimsErr != nil {
		return nil, fmt.Errorf("failed to parse id token claims: %w", claimsErr)
	}

	log.Warn().Str("nonce", claims.Nonce).Msg("TODO: verify nonce")

	ps := rez.AuthProviderSession{
		User: ent.User{
			ProviderID: token.Subject,
			Email:      claims.Email,
			Confirmed:  claims.Verified,
			Name:       claims.Name,
			Timezone:   claims.Locale,
		},
		Tenant: ent.Tenant{
			Name:       claims.TenantId,
			ProviderID: claims.TenantId,
		},
		ExpiresAt:   token.Expiry,
		RedirectUrl: rez.FrontendUrl,
	}

	return &ps, nil
}
