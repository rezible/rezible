package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"golang.org/x/oauth2"
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

func (gpc *GenericProviderConfig) GetOidcProvider(ctx context.Context) (*oidc.Provider, error) {
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

func NewGenericAuthSessionProvider(ctx context.Context, cfg GenericProviderConfig) (*AuthSessionProvider, error) {
	callbackPath, cbPathErr := url.JoinPath(rez.Config.ApiRouteBase(), rez.Config.AuthRouteBase(), cfg.ProviderID, "callback")
	if cbPathErr != nil {
		return nil, fmt.Errorf("callback path: %w", cbPathErr)
	}
	redirectUrl, urlErr := url.JoinPath(rez.Config.AppUrl(), callbackPath)
	if urlErr != nil {
		return nil, fmt.Errorf("creating redirect url: %w", urlErr)
	}

	prov, provErr := cfg.GetOidcProvider(ctx)
	if provErr != nil {
		return nil, fmt.Errorf("fetch provider config: %w", provErr)
	}

	oauth2Config := oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  redirectUrl,
		Endpoint:     prov.Endpoint(),
		Scopes:       cfg.Scopes,
	}

	return &AuthSessionProvider{
		providerId:   cfg.ProviderID,
		callbackPath: callbackPath,
		displayName:  cfg.DisplayName,
		oauth2Config: oauth2Config,
		verifier:     prov.Verifier(&oidc.Config{ClientID: cfg.ClientID}),
		sessionStore: configureSessionStore(),
		idp:          &genericIdentity{},
	}, nil
}

type genericIdentity struct{}

func (p *genericIdentity) GetAuthCodeOptions(r *http.Request) []oauth2.AuthCodeOption {
	return nil
}

func (p *genericIdentity) ExtractTokenSession(token *oidc.IDToken) (*rez.AuthProviderSession, error) {
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
