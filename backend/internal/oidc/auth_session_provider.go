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
	"os"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

//func ConfigureSessionStore(secretKey string) {
//	maxAge := 86400 * 30 // 30 days
//
//	store := sessions.NewCookieStore([]byte(secretKey))
//	store.MaxAge(maxAge)
//	store.Options.Path = "/"
//	store.Options.HttpOnly = true
//	store.Options.SameSite = http.SameSiteStrictMode
//	store.Options.Secure = true
//
//	gothic.Store = store
//}

type AuthSessionProvider struct {
	callbackPath string
	displayName  string
	oauth2Config oauth2.Config
	verifier     *oidc.IDTokenVerifier
}

type Config struct {
	ProviderName string `json:"provider_name"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	IssuerUrl    string `json:"issuer_url"`
	//DiscoveryUrl string `json:"oidc_discovery_url,omitempty"`
}

func NewGoogleProvider(ctx context.Context) (*AuthSessionProvider, error) {
	clientID := os.Getenv("GOOGLE_OIDC_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_OIDC_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("client id/secret env vars not set")
	}
	cfg := Config{
		ProviderName: "Google",
		ClientID:     clientID,
		ClientSecret: clientSecret,
		IssuerUrl:    "https://accounts.google.com",
		//DiscoveryUrl: "https://accounts.google.com/.well-known/openid-configuration",
	}

	return NewAuthSessionProvider(ctx, cfg)
}

func NewAuthSessionProvider(ctx context.Context, cfg Config) (*AuthSessionProvider, error) {
	callbackPath := fmt.Sprintf("/auth/%s/%s", strings.ToLower(cfg.ProviderName), "callback")
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
		callbackPath: callbackPath,
		displayName:  cfg.ProviderName,
		oauth2Config: oauth2Config,
		verifier:     oidcProvider.Verifier(&oidc.Config{ClientID: cfg.ClientID}),
	}, nil
}

func (s *AuthSessionProvider) Name() string {
	return s.displayName
}

func (s *AuthSessionProvider) ClearSession(w http.ResponseWriter, r *http.Request) {

}

var userMapping = &ent.User{
	Name:  "y",
	Email: "y",
}

func (s *AuthSessionProvider) GetUserMapping() *ent.User {
	return userMapping
}

func (s *AuthSessionProvider) StartAuthFlow(w http.ResponseWriter, r *http.Request) {
	redirect, redirectErr := s.createProviderSessionRedirect(w, r)
	if redirectErr != nil {
		log.Error().Err(redirectErr).Msg("could not create provider session redirect")
		http.Error(w, redirectErr.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, redirect, http.StatusFound)
}

func (s *AuthSessionProvider) createProviderSessionRedirect(w http.ResponseWriter, r *http.Request) (string, error) {
	state := r.URL.Query().Get("state")
	if state == "" {
		nonceBytes := make([]byte, 64)
		_, err := io.ReadFull(rand.Reader, nonceBytes)
		if err != nil {
			return "", fmt.Errorf("could not generate session state nonce: %w", err)
		}
		state = base64.URLEncoding.EncodeToString(nonceBytes)
	}

	// TODO: PKCE / nonce gen?
	authUrl := s.oauth2Config.AuthCodeURL(state)
	log.Debug().Str("authUrl", authUrl).Msg("redirect")

	return authUrl, nil
}

func (s *AuthSessionProvider) HandleAuthFlowRequest(w http.ResponseWriter, r *http.Request, onCreated func(session rez.AuthProviderSession)) bool {
	if r.URL.Path == s.callbackPath {
		if cbErr := s.handleFlowCallback(r, onCreated); cbErr != nil {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		}
		return true
	}
	return false
}

func (s *AuthSessionProvider) handleFlowCallback(r *http.Request, onCreated func(session rez.AuthProviderSession)) error {
	ctx := r.Context()

	state := ""
	if validateErr := s.validateRequestSessionState(r, state); validateErr != nil {
		log.Warn().Msg("invalid request session state")
		if !rez.DebugMode {
			return validateErr
		}
	}

	token, tokenErr := s.getVerifiedIdToken(ctx, r.URL.Query().Get("code"))
	if tokenErr != nil {
		return fmt.Errorf("failed to get id token: %w", tokenErr)
	}

	var claims struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Verified bool   `json:"email_verified"`
		Locale   string `json:"locale"`
		Nonce    string `json:"nonce"`
		TenantId string `json:"hd"`
	}
	if claimsErr := token.Claims(&claims); claimsErr != nil {
		return fmt.Errorf("failed to parse id token claims: %w", claimsErr)
	}

	log.Warn().Str("nonce", claims.Nonce).Msg("TODO: verify nonce")

	user := &ent.User{
		AuthProviderID: token.Subject,
		Email:          claims.Email,
		Confirmed:      claims.Verified,
		Name:           claims.Name,
		Timezone:       claims.Locale,
	}

	onCreated(rez.AuthProviderSession{
		User:        user,
		ExpiresAt:   token.Expiry,
		TenantId:    claims.TenantId,
		RedirectUrl: rez.FrontendUrl,
	})

	return nil
}

func (s *AuthSessionProvider) getVerifiedIdToken(ctx context.Context, authCode string) (*oidc.IDToken, error) {
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

func (s *AuthSessionProvider) validateRequestSessionState(r *http.Request, state string) error {
	params := r.URL.Query()
	reqState := params.Get("state")
	if reqState == "" && r.Method == http.MethodPost {
		reqState = r.FormValue("state")
	}

	if reqState != state {
		return errors.New("state mismatch")
	}
	return nil
}
