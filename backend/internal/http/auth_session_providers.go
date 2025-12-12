package http

import (
	"context"
	"fmt"
	"strings"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/internal/http/oidc"
	"github.com/rezible/rezible/internal/http/saml"
)

func isAuthProviderEnabled(name string) bool {
	return rez.Config.GetBool("auth." + strings.ToUpper(name) + ".enabled")
}

func getAuthSessionProviders(ctx context.Context, secretKey string) ([]rez.AuthSessionProvider, error) {
	var provs []rez.AuthSessionProvider

	if isAuthProviderEnabled("saml") {
		samlProv, spErr := saml.NewAuthSessionProvider(ctx)
		if spErr != nil {
			return nil, fmt.Errorf("saml.NewAuthSessionProvider: %w", spErr)
		}
		provs = append(provs, samlProv)
	}

	oidc.SessionSecretKey = []byte(secretKey)

	if isAuthProviderEnabled("google") {
		googleCfg := oidc.GoogleProviderConfig{
			ClientID:     rez.Config.GetString("auth.google.client_id"),
			ClientSecret: rez.Config.GetString("auth.google.client_secret"),
		}
		googleProv, googleErr := oidc.NewGoogleAuthSessionProvider(ctx, googleCfg)
		if googleErr != nil {
			return nil, fmt.Errorf("oidc.NewGoogleProvider: %w", googleErr)
		}

		provs = append(provs, googleProv)
	}

	if isAuthProviderEnabled("oidc") {
		cfg := oidc.GenericProviderConfig{
			ProviderID:   rez.Config.GetString("auth.oidc.provider_id"),
			DisplayName:  rez.Config.GetString("auth.oidc.display_name"),
			ClientID:     rez.Config.GetString("auth.oidc.client_id"),
			ClientSecret: rez.Config.GetString("auth.oidc.client_secret"),
			Scopes:       rez.Config.GetStrings("auth.oidc.scopes"),
			IssuerUrl:    rez.Config.GetString("auth.oidc.issuer_url"),
			DiscoveryUrl: rez.Config.GetString("auth.oidc.discovery_url"),
		}
		prov, provErr := oidc.NewGenericAuthSessionProvider(ctx, cfg)
		if provErr != nil {
			return nil, fmt.Errorf("oidc.NewGenericAuthSessionProvider: %w", provErr)
		}

		provs = append(provs, prov)
	}

	return provs, nil
}
