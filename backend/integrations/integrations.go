package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"reflect"
	"runtime"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/internal/fake"
	"github.com/rezible/rezible/internal/google"
	"github.com/rezible/rezible/internal/slack"
)

var (
	packageNameMap     = map[string]rez.IntegrationPackage{}
	availablePackages  []rez.IntegrationPackage
	supportedDataKinds []string
	packageSetupFuncs  = []rez.SetupPackageFunc{
		fakeprovider.SetupIntegration,
		slack.SetupIntegration,
		google.SetupIntegration,
	}
)

func Setup(ctx context.Context, svcs *rez.Services) error {
	availablePackages = make([]rez.IntegrationPackage, 0, len(availablePackages))
	packageNameMap = make(map[string]rez.IntegrationPackage)
	enabledSupportedDataKinds := mapset.NewSet[string]()
	for _, setupFn := range packageSetupFuncs {
		pkg, pkgErr := setupFn(ctx, svcs)
		if pkgErr != nil {
			funcName := runtime.FuncForPC(reflect.ValueOf(setupFn).Pointer()).Name()
			return fmt.Errorf("%s: %w", funcName, pkgErr)
		}
		enabled, configErr := pkg.IsAvailable()
		if !enabled {
			continue
		}
		if configErr != nil {
			log.Error().
				Err(configErr).
				Str("integration", pkg.Name()).
				Msg("integration setup error")
			continue
		}
		availablePackages = append(availablePackages, pkg)
		packageNameMap[pkg.Name()] = pkg
		enabledSupportedDataKinds.Append(pkg.SupportedDataKinds()...)
	}

	if enabledSupportedDataKinds.Cardinality() == 0 {
		return fmt.Errorf("no supported data kinds found")
	}
	supportedDataKinds = enabledSupportedDataKinds.ToSlice()
	// TODO: check if required data kinds are supported

	return nil
}

func GetPackage(name string) (rez.IntegrationPackage, error) {
	p, valid := packageNameMap[name]
	if !valid {
		return nil, fmt.Errorf("unknown integration: %s", name)
	}
	return p, nil
}

func GetAvailable() []rez.IntegrationPackage {
	return availablePackages
}

type IntegrationWithEventListeners interface {
	EventListeners() map[string]rez.EventListener
}

func GetEventListeners() map[string]rez.EventListener {
	els := make(map[string]rez.EventListener)
	for _, p := range availablePackages {
		if elIntegration, ok := p.(IntegrationWithEventListeners); ok {
			for name, l := range elIntegration.EventListeners() {
				els[name] = l
			}
		}
	}
	return els
}

type IntegrationWithWebhookHandlers interface {
	WebhookHandlers() map[string]http.Handler
}

func GetWebhookHandlers() map[string]http.Handler {
	whs := make(map[string]http.Handler)
	for _, p := range availablePackages {
		if elIntegration, ok := p.(IntegrationWithWebhookHandlers); ok {
			for prefix, h := range elIntegration.WebhookHandlers() {
				whs[path.Join(p.Name(), prefix)] = h
			}
		}
	}
	return whs
}

type IntegrationWithOAuth2SetupFlow interface {
	OAuth2Config() *oauth2.Config
	ExtractIntegrationConfigFromToken(*oauth2.Token) (json.RawMessage, error)
}

func GetOAuthIntegration(name string) (IntegrationWithOAuth2SetupFlow, error) {
	ip, ipErr := GetPackage(name)
	if ipErr != nil {
		return nil, fmt.Errorf("invalid integration %s: %w", name, ipErr)
	}
	oauth2Intg, ok := ip.(IntegrationWithOAuth2SetupFlow)
	if !ok {
		return nil, fmt.Errorf("oauth2 flow not supported for integration %s", name)
	}
	if oauth2Intg.OAuth2Config() == nil {
		return nil, fmt.Errorf("nil integration oauth2 configuration")
	}
	return oauth2Intg, nil
}

type IntegrationWithOIDCAuthSessionIdentityProvider interface {
	GetOIDCAuthSessionIdentityProvider() (bool, rez.OIDCAuthSessionIdentityProvider, error)
}

func GetOIDCAuthSessionIdentityProviders() ([]rez.OIDCAuthSessionIdentityProvider, error) {
	var providers []rez.OIDCAuthSessionIdentityProvider
	for _, p := range availablePackages {
		if provider, ok := p.(IntegrationWithOIDCAuthSessionIdentityProvider); ok {
			enabled, prov, cfgErr := provider.GetOIDCAuthSessionIdentityProvider()
			if !enabled {
				continue
			}
			if cfgErr != nil {
				return nil, fmt.Errorf("oidc auth session provider config error: %w", cfgErr)
			}
			providers = append(providers, prov)
		}
	}
	return providers, nil
}

func getDataProviders[DP any, I any](intgs ent.Integrations, fn func(I, *ent.Integration) (DP, error)) ([]DP, error) {
	provs := make([]DP, 0, len(availablePackages))
	for _, intg := range intgs {
		if p, valid := packageNameMap[intg.Name]; valid {
			dpProv, hasSupport := p.(I)
			log.Debug().
				Str("integration", p.Name()).
				Bool("supports", hasSupport).
				Msg("getDataProviders")
			if hasSupport {
				prov, provErr := fn(dpProv, intg)
				if provErr != nil {
					return nil, fmt.Errorf("loading data provider: %w", provErr)
				}
				provs = append(provs, prov)
			}
		}
	}
	return provs, nil
}

type IntegrationWithUserDataProvider interface {
	MakeUserDataProvider(context.Context, *ent.Integration) (rez.UserDataProvider, error)
}

func GetUserDataProviders(ctx context.Context, intgs ent.Integrations) ([]rez.UserDataProvider, error) {
	return getDataProviders(intgs, func(dpi IntegrationWithUserDataProvider, i *ent.Integration) (rez.UserDataProvider, error) {
		return dpi.MakeUserDataProvider(ctx, i)
	})
}

type IntegrationWithTeamDataProvider interface {
	MakeTeamDataProvider(context.Context, *ent.Integration) (rez.TeamDataProvider, error)
}

func GetTeamDataProviders(ctx context.Context, intgs ent.Integrations) ([]rez.TeamDataProvider, error) {
	return getDataProviders(intgs, func(dpi IntegrationWithTeamDataProvider, i *ent.Integration) (rez.TeamDataProvider, error) {
		return dpi.MakeTeamDataProvider(ctx, i)
	})
}

type IntegrationWithOncallDataProvider interface {
	MakeOncallDataProvider(context.Context, *ent.Integration) (rez.OncallDataProvider, error)
}

func GetOncallDataProviders(ctx context.Context, intgs ent.Integrations) ([]rez.OncallDataProvider, error) {
	return getDataProviders(intgs, func(dpi IntegrationWithOncallDataProvider, i *ent.Integration) (rez.OncallDataProvider, error) {
		return dpi.MakeOncallDataProvider(ctx, i)
	})
}

type IntegrationWithAlertDataProvider interface {
	MakeAlertDataProvider(context.Context, *ent.Integration) (rez.AlertDataProvider, error)
}

func GetAlertDataProviders(ctx context.Context, intgs ent.Integrations) ([]rez.AlertDataProvider, error) {
	return getDataProviders(intgs, func(dpi IntegrationWithAlertDataProvider, i *ent.Integration) (rez.AlertDataProvider, error) {
		return dpi.MakeAlertDataProvider(ctx, i)
	})
}

type IntegrationWithIncidentDataProvider interface {
	MakeIncidentDataProvider(context.Context, *ent.Integration) (rez.IncidentDataProvider, error)
}

func GetIncidentDataProviders(ctx context.Context, intgs ent.Integrations) ([]rez.IncidentDataProvider, error) {
	return getDataProviders(intgs, func(dpi IntegrationWithIncidentDataProvider, i *ent.Integration) (rez.IncidentDataProvider, error) {
		return dpi.MakeIncidentDataProvider(ctx, i)
	})
}

type IntegrationWithSystemComponentsDataProvider interface {
	MakeSystemComponentsDataProvider(context.Context, *ent.Integration) (rez.SystemComponentsDataProvider, error)
}

func GetSystemComponentsDataProviders(ctx context.Context, intgs ent.Integrations) ([]rez.SystemComponentsDataProvider, error) {
	return getDataProviders(intgs, func(dpi IntegrationWithSystemComponentsDataProvider, i *ent.Integration) (rez.SystemComponentsDataProvider, error) {
		return dpi.MakeSystemComponentsDataProvider(ctx, i)
	})
}

type IntegrationWithTicketDataProvider interface {
	MakeTicketDataProvider(context.Context, *ent.Integration) (rez.TicketDataProvider, error)
}

func GetTicketDataProviders(ctx context.Context, intgs ent.Integrations) ([]rez.TicketDataProvider, error) {
	return getDataProviders(intgs, func(dpi IntegrationWithTicketDataProvider, i *ent.Integration) (rez.TicketDataProvider, error) {
		return dpi.MakeTicketDataProvider(ctx, i)
	})
}

type IntegrationWithPlaybookDataProvider interface {
	MakePlaybookDataProvider(context.Context, *ent.Integration) (rez.PlaybookDataProvider, error)
}

func GetPlaybookDataProviders(ctx context.Context, intgs ent.Integrations) ([]rez.PlaybookDataProvider, error) {
	return getDataProviders(intgs, func(dpi IntegrationWithPlaybookDataProvider, i *ent.Integration) (rez.PlaybookDataProvider, error) {
		return dpi.MakePlaybookDataProvider(ctx, i)
	})
}
