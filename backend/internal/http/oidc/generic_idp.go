package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type GenericProviderConfig struct {
	ProviderID   string
	DisplayName  string
	ClientID     string
	ClientSecret string
	Scopes       []string
	IssuerUrl    string
	DiscoveryUrl string
}

func (gpc *GenericProviderConfig) LoadProvider(ctx context.Context) (*oidc.Provider, error) {
	if gpc.IssuerUrl != "" {
		return oidc.NewProvider(ctx, gpc.IssuerUrl)
	}

	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, gpc.DiscoveryUrl, nil)
	if reqErr != nil {
		return nil, fmt.Errorf("failed to create request: %w", reqErr)
	}
	res, doErr := http.DefaultClient.Do(req)
	if doErr != nil {
		return nil, fmt.Errorf("failed to do request: %w", doErr)
	}
	var provCfg oidc.ProviderConfig
	if jsonErr := json.NewDecoder(res.Body).Decode(&provCfg); jsonErr != nil {
		return nil, fmt.Errorf("failed to parse response body: %w", jsonErr)
	}
	return provCfg.NewProvider(ctx), nil
}

type genericProvider struct {
	provider *oidc.Provider
	config   GenericProviderConfig
}

func GetGenericOIDCAuthSessionProvider() (bool, rez.OIDCAuthSessionIdentityProvider, error) {
	if !rez.Config.GetBool("auth.oidc.enabled") {
		return false, nil, nil
	}

	cfg := GenericProviderConfig{
		ProviderID:   rez.Config.GetString("auth.oidc.provider_id"),
		DisplayName:  rez.Config.GetString("auth.oidc.display_name"),
		ClientID:     rez.Config.GetString("auth.oidc.client_id"),
		ClientSecret: rez.Config.GetString("auth.oidc.client_secret"),
		Scopes:       rez.Config.GetStrings("auth.oidc.scopes"),
		IssuerUrl:    rez.Config.GetString("auth.oidc.issuer_url"),
		DiscoveryUrl: rez.Config.GetString("auth.oidc.discovery_url"),
	}

	return true, &genericProvider{config: cfg}, nil
}

func (op *genericProvider) Id() string {
	return op.config.ProviderID
}

func (op *genericProvider) DisplayName() string {
	return op.config.DisplayName
}

func (op *genericProvider) LoadConfig(ctx context.Context, redirectUrl string) (*oidc.IDTokenVerifier, *oauth2.Config, error) {
	prov, provErr := op.config.LoadProvider(ctx)
	if provErr != nil {
		return nil, nil, fmt.Errorf("fetch provider config: %w", provErr)
	}
	op.provider = prov

	oauth2Config := oauth2.Config{
		ClientID:     op.config.ClientID,
		ClientSecret: op.config.ClientSecret,
		RedirectURL:  redirectUrl,
		Endpoint:     prov.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	verifier := prov.Verifier(&oidc.Config{ClientID: op.config.ClientID})

	return verifier, &oauth2Config, nil
}

func (op *genericProvider) GetAuthCodeOptions(r *http.Request) []oauth2.AuthCodeOption {
	return nil
}

func (op *genericProvider) ExtractTokenSession(token *oidc.IDToken) (*rez.AuthProviderSession, error) {
	var claims struct {
		Email      string `json:"email"`
		GivenName  string `json:"given_name"`
		FamilyName string `json:"family_name"`
	}
	if claimsErr := token.Claims(&claims); claimsErr != nil {
		return nil, fmt.Errorf("failed to parse id token claims: %w", claimsErr)
	}

	domain := strings.Split(claims.Email, "@")[1]

	org := ent.Organization{
		ExternalID: domain,
		Name:       domain,
	}

	usr := ent.User{
		AuthProviderID: token.Subject,
		Email:          claims.Email,
		Name:           claims.GivenName,
	}

	ps := rez.AuthProviderSession{
		User:         usr,
		Organization: org,
		ExpiresAt:    token.Expiry,
		RedirectUrl:  rez.Config.AppUrl(),
	}

	return &ps, nil
}

//func NewGenericAuthSessionProvider(ctx context.Context, cfg GenericProviderConfig) (*AuthSessionProvider, error) {
//	callbackPath, cbPathErr := url.JoinPath(rez.Config.ApiRouteBase(), rez.Config.AuthRouteBase(), cfg.ProviderID, "callback")
//	if cbPathErr != nil {
//		return nil, fmt.Errorf("callback path: %w", cbPathErr)
//	}
//	redirectUrl, urlErr := url.JoinPath(rez.Config.AppUrl(), callbackPath)
//	if urlErr != nil {
//		return nil, fmt.Errorf("creating redirect url: %w", urlErr)
//	}
//
//	prov, provErr := cfg.GetOidcProvider(ctx)
//	if provErr != nil {
//		return nil, fmt.Errorf("fetch provider config: %w", provErr)
//	}
//
//	oauth2Config := oauth2.Config{
//		ClientID:     cfg.ClientID,
//		ClientSecret: cfg.ClientSecret,
//		RedirectURL:  redirectUrl,
//		Endpoint:     prov.Endpoint(),
//		Scopes:       cfg.Scopes,
//	}
//
//	return &AuthSessionProvider{
//		providerId:   cfg.ProviderID,
//		callbackPath: callbackPath,
//		displayName:  cfg.DisplayName,
//		oauth2Config: oauth2Config,
//		verifier:     prov.Verifier(&oidc.Config{ClientID: cfg.ClientID}),
//		sessionStore: configureSessionStore(),
//		idp:          &genericIdentity{},
//	}, nil
//}
