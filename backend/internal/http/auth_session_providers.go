package http

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/failsafe-go/failsafe-go"
	"github.com/failsafe-go/failsafe-go/retrypolicy"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/internal/http/oidc"
	"github.com/rezible/rezible/internal/http/saml"
)

func isAuthProviderEnabled(name string) bool {
	return rez.Config.GetBool("auth." + strings.ToUpper(name) + ".enabled")
}

func getAuthSessionProviders(ctx context.Context, secretKey string) ([]rez.AuthSessionProvider, error) {
	var provs []rez.AuthSessionProvider

	// TODO: these should be loaded from integrations anyway
	if rez.Config.DataSyncMode() {
		return provs, nil
	}

	defaultRetryPolicy := retrypolicy.NewBuilder[rez.AuthSessionProvider]().
		//HandleErrors(ErrConnecting).
		WithDelay(time.Second).
		WithMaxRetries(3)

	if isAuthProviderEnabled("saml") {
		loadSamlProvFn := func() (rez.AuthSessionProvider, error) {
			return saml.NewAuthSessionProvider(ctx)
		}
		samlProv, provErr := failsafe.With(defaultRetryPolicy.Build()).Get(loadSamlProvFn)
		if provErr != nil {
			return nil, fmt.Errorf("saml.NewAuthSessionProvider: %w", provErr)
		}
		provs = append(provs, samlProv)
	}

	oidc.SessionSecretKey = []byte(secretKey)

	if isAuthProviderEnabled("google") {
		googleCfg := oidc.GoogleProviderConfig{
			ClientID:     rez.Config.GetString("auth.google.client_id"),
			ClientSecret: rez.Config.GetString("auth.google.client_secret"),
		}
		loadGoogleProvFn := func() (rez.AuthSessionProvider, error) {
			return oidc.NewGoogleAuthSessionProvider(ctx, googleCfg)
		}
		googleProv, googleErr := failsafe.With(defaultRetryPolicy.Build()).Get(loadGoogleProvFn)
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
		loadGenericProvFn := func() (rez.AuthSessionProvider, error) {
			return oidc.NewGenericAuthSessionProvider(ctx, cfg)
		}
		prov, provErr := failsafe.With(defaultRetryPolicy.Build()).Get(loadGenericProvFn)
		if provErr != nil {
			return nil, fmt.Errorf("oidc.NewGenericAuthSessionProvider: %w", provErr)
		}
		provs = append(provs, prov)
	}

	return provs, nil
}
