package oidc

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/coreos/go-oidc/v3/oidc"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

const (
	googleProviderId = "google"
	googleIssuerUrl  = "https://accounts.google.com"
)

func NewGoogleAuthSessionProvider(ctx context.Context, cfg ProviderConfig) (*AuthSessionProvider, error) {
	callbackPath, cbPathErr := url.JoinPath(rez.Config.ApiRouteBase(), rez.Config.AuthRouteBase(), googleProviderId, "callback")
	if cbPathErr != nil {
		return nil, fmt.Errorf("callback path: %w", cbPathErr)
	}
	redirectUrl, urlErr := url.JoinPath(rez.Config.BackendUrl(), callbackPath)
	if urlErr != nil {
		return nil, fmt.Errorf("creating redirect url: %w", urlErr)
	}

	prov, oidcErr := oidc.NewProvider(ctx, googleIssuerUrl)
	if oidcErr != nil {
		return nil, fmt.Errorf("oidc.NewProvider: %w", oidcErr)
	}

	oauth2Config := oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  redirectUrl,
		Endpoint:     prov.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return &AuthSessionProvider{
		providerId:   googleProviderId,
		callbackPath: callbackPath,
		displayName:  "Google",
		oauth2Config: oauth2Config,
		verifier:     prov.Verifier(&oidc.Config{ClientID: cfg.ClientID}),
		sessionStore: configureSessionStore(cfg.SessionSecret),
		idp:          &googleIdentity{},
	}, nil
}

type googleIdentity struct{}

func (p *googleIdentity) GetAuthCodeOptions(r *http.Request) []oauth2.AuthCodeOption {
	return nil
}

func (p *googleIdentity) ExtractTokenSession(token *oidc.IDToken) (*rez.AuthProviderSession, error) {
	var claims struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Verified bool   `json:"email_verified"`
		Locale   string `json:"locale"`
		Nonce    string `json:"nonce"`
		OrgId    string `json:"hd"`
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
		Organization: ent.Organization{
			ProviderID: claims.OrgId,
			Name:       claims.OrgId, // TODO: use domain?
		},
		ExpiresAt:   token.Expiry,
		RedirectUrl: rez.Config.AppUrl(),
	}

	return &ps, nil
}
