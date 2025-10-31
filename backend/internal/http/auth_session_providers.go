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
	key := "auth." + strings.ToUpper(name) + ".enabled"
	return strings.ToLower(rez.Config.GetString(key)) == "true"
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

	if isAuthProviderEnabled("google") {
		clientID := rez.Config.GetString("auth.google.client_id")
		clientSecret := rez.Config.GetString("auth.google.client_secret")
		if clientID == "" || clientSecret == "" {
			return nil, fmt.Errorf("google client id/secret not set")
		}

		googleCfg := oidc.ProviderConfig{
			SessionSecret: secretKey,
			ClientID:      clientID,
			ClientSecret:  clientSecret,
		}
		googleProv, googleErr := oidc.NewGoogleAuthSessionProvider(ctx, googleCfg)
		if googleErr != nil {
			return nil, fmt.Errorf("oidc.NewGoogleProvider: %w", googleErr)
		}

		provs = append(provs, googleProv)
	}

	return provs, nil
}
