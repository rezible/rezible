package http

import (
	"context"
	"fmt"
	"os"
	"strings"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/internal/http/oidc"
	"github.com/rezible/rezible/internal/http/saml"
)

func isAuthProviderEnabled(name string) bool {
	return strings.ToLower(os.Getenv("AUTH_ENABLE_"+strings.ToUpper(name))) == "true"
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

	if isAuthProviderEnabled("google_oidc") {
		clientID := os.Getenv("GOOGLE_OIDC_CLIENT_ID")
		clientSecret := os.Getenv("GOOGLE_OIDC_CLIENT_SECRET")
		if clientID == "" || clientSecret == "" {
			return nil, fmt.Errorf("client id/secret env vars not set")
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
