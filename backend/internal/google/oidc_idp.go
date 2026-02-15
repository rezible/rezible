package google

import (
	"context"
	"fmt"
	"net/http"

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

type oidcIdp struct {
	provider     *oidc.Provider
	clientId     string
	clientSecret string
}

func (i *integration) GetOIDCAuthSessionIdentityProvider() (bool, rez.OIDCAuthSessionIdentityProvider, error) {
	if !rez.Config.GetBool("google.oidc.enabled") {
		return false, nil, nil
	}

	op := &oidcIdp{
		clientId:     rez.Config.GetString("google.oidc.client_id"),
		clientSecret: rez.Config.GetString("google.oidc.client_secret"),
	}

	return true, op, nil
}

func (op *oidcIdp) Id() string {
	return integrationName
}

func (op *oidcIdp) DisplayName() string {
	return "Google"
}

func (op *oidcIdp) LoadConfig(ctx context.Context, redirectUrl string) (*oidc.IDTokenVerifier, *oauth2.Config, error) {
	prov, oidcErr := oidc.NewProvider(ctx, googleIssuerUrl)
	if oidcErr != nil {
		return nil, nil, fmt.Errorf("oidc.NewProvider: %w", oidcErr)
	}
	op.provider = prov

	oauth2Config := oauth2.Config{
		ClientID:     op.clientId,
		ClientSecret: op.clientSecret,
		RedirectURL:  redirectUrl,
		Endpoint:     prov.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	verifier := prov.Verifier(&oidc.Config{ClientID: op.clientId})

	return verifier, &oauth2Config, nil
}

func (op *oidcIdp) GetAuthCodeOptions(r *http.Request) []oauth2.AuthCodeOption {
	return nil
}

func (op *oidcIdp) ExtractTokenSession(token *oidc.IDToken) (*rez.AuthProviderSession, error) {
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
			AuthProviderID: token.Subject,
			Email:          claims.Email,
			Confirmed:      claims.Verified,
			Name:           claims.Name,
			Timezone:       claims.Locale,
		},
		Organization: ent.Organization{
			ExternalID: claims.OrgId,
			Name:       claims.OrgId, // TODO: use domain?
		},
		ExpiresAt:   token.Expiry,
		RedirectUrl: rez.Config.AppUrl(),
	}

	return &ps, nil
}
