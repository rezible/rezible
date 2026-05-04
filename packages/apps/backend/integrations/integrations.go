package integrations

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"path"
	"reflect"
	"runtime"

	mapset "github.com/deckarep/golang-set/v2"
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
	setupFuncs         = []rez.PackageSetupFunc{
		fakeprovider.SetupIntegration,
		slack.SetupIntegration,
		google.SetupIntegration,
	}
)

func Setup(ctx context.Context, svcs *rez.Services) error {
	availablePackages = make([]rez.IntegrationPackage, 0, len(availablePackages))
	packageNameMap = make(map[string]rez.IntegrationPackage)
	enabledSupportedDataKinds := mapset.NewSet[string]()
	for _, setupFn := range setupFuncs {
		pkg, pkgErr := setupFn(ctx, svcs)
		if pkgErr != nil {
			funcName := runtime.FuncForPC(reflect.ValueOf(setupFn).Pointer()).Name()
			return fmt.Errorf("%s: %w", funcName, pkgErr)
		}
		available, configErr := pkg.IsAvailable()
		slog.Debug("integration package",
			"name", pkg.Name(),
			"configErr", configErr,
			"available", available,
		)
		if !available {
			continue
		}
		if configErr != nil {
			slog.Error("integration setup error",
				"error", configErr,
				"integration", pkg.Name(),
			)
			continue
		}
		availablePackages = append(availablePackages, pkg)
		packageNameMap[pkg.Name()] = pkg
		enabledSupportedDataKinds.Append(pkg.SupportedDataKinds()...)
	}
	supportedDataKinds = enabledSupportedDataKinds.ToSlice()

	// TODO: check if required data kinds are supported
	//if enabledSupportedDataKinds.Cardinality() == 0 {
	//	return nil, fmt.Errorf("no supported data kinds found")
	//}

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

func GetEventListeners() map[string]rez.EventListener {
	els := make(map[string]rez.EventListener)
	for _, pkg := range availablePackages {
		elPkg, ok := pkg.(interface {
			EventListeners() map[string]rez.EventListener
		})
		if ok {
			for name, listener := range elPkg.EventListeners() {
				els[name] = listener
			}
		}
	}
	return els
}

func GetWebhookHandlers() map[string]http.Handler {
	whs := make(map[string]http.Handler)
	for _, pkg := range availablePackages {
		whPkg, ok := pkg.(interface {
			WebhookHandlers() map[string]http.Handler
		})
		if ok {
			for prefix, hook := range whPkg.WebhookHandlers() {
				whs[path.Join(pkg.Name(), prefix)] = hook
			}
		}
	}
	return whs
}

type IntegrationWithOAuth2Flow interface {
	OAuth2Config() *oauth2.Config
	ExtractIntegrationConfigFromToken(*oauth2.Token) (map[string]any, error)
}

func GetOAuthIntegration(name string) (IntegrationWithOAuth2Flow, error) {
	ip, ipErr := GetPackage(name)
	if ipErr != nil {
		return nil, fmt.Errorf("invalid integration %s: %w", name, ipErr)
	}
	oauth2Intg, ok := ip.(IntegrationWithOAuth2Flow)
	if !ok {
		return nil, fmt.Errorf("oauth2 flow not supported for integration %s", name)
	}
	if oauth2Intg.OAuth2Config() == nil {
		return nil, fmt.Errorf("nil integration oauth2 configuration")
	}
	return oauth2Intg, nil
}

func getDataProviders[DP any, I any](intgs ent.Integrations, fn func(I, *ent.Integration) (DP, error)) (map[string]DP, error) {
	provs := make(map[string]DP)
	for _, intg := range intgs {
		if p, valid := packageNameMap[intg.Name]; valid {
			dpProv, hasSupport := p.(I)
			if hasSupport {
				prov, provErr := fn(dpProv, intg)
				if provErr != nil {
					return nil, fmt.Errorf("loading data provider: %w", provErr)
				}
				provs[intg.Name] = prov
			}
		}
	}
	return provs, nil
}

func GetUserDataProviders(ctx context.Context, intgs ent.Integrations) (map[string]rez.UserDataProvider, error) {
	type integrationWithUserDataProvider interface {
		MakeUserDataProvider(context.Context, *ent.Integration) (rez.UserDataProvider, error)
	}
	return getDataProviders(intgs, func(dpi integrationWithUserDataProvider, i *ent.Integration) (rez.UserDataProvider, error) {
		return dpi.MakeUserDataProvider(ctx, i)
	})
}

func GetTeamDataProviders(ctx context.Context, intgs ent.Integrations) (map[string]rez.TeamDataProvider, error) {
	type integrationWithTeamDataProvider interface {
		MakeTeamDataProvider(context.Context, *ent.Integration) (rez.TeamDataProvider, error)
	}
	return getDataProviders(intgs, func(dpi integrationWithTeamDataProvider, i *ent.Integration) (rez.TeamDataProvider, error) {
		return dpi.MakeTeamDataProvider(ctx, i)
	})
}

func GetOncallDataProviders(ctx context.Context, intgs ent.Integrations) (map[string]rez.OncallDataProvider, error) {
	type integrationWithOncallDataProvider interface {
		MakeOncallDataProvider(context.Context, *ent.Integration) (rez.OncallDataProvider, error)
	}
	return getDataProviders(intgs, func(dpi integrationWithOncallDataProvider, i *ent.Integration) (rez.OncallDataProvider, error) {
		return dpi.MakeOncallDataProvider(ctx, i)
	})
}

func GetAlertDataProviders(ctx context.Context, intgs ent.Integrations) (map[string]rez.AlertDataProvider, error) {
	type integrationWithAlertDataProvider interface {
		MakeAlertDataProvider(context.Context, *ent.Integration) (rez.AlertDataProvider, error)
	}
	return getDataProviders(intgs, func(dpi integrationWithAlertDataProvider, i *ent.Integration) (rez.AlertDataProvider, error) {
		return dpi.MakeAlertDataProvider(ctx, i)
	})
}

func GetIncidentDataProviders(ctx context.Context, intgs ent.Integrations) (map[string]rez.IncidentDataProvider, error) {
	type integrationWithIncidentDataProvider interface {
		MakeIncidentDataProvider(context.Context, *ent.Integration) (rez.IncidentDataProvider, error)
	}
	return getDataProviders(intgs, func(dpi integrationWithIncidentDataProvider, i *ent.Integration) (rez.IncidentDataProvider, error) {
		return dpi.MakeIncidentDataProvider(ctx, i)
	})
}

func GetSystemComponentsDataProviders(ctx context.Context, intgs ent.Integrations) (map[string]rez.SystemComponentsDataProvider, error) {
	type integrationWithSystemComponentsDataProvider interface {
		MakeSystemComponentsDataProvider(context.Context, *ent.Integration) (rez.SystemComponentsDataProvider, error)
	}
	return getDataProviders(intgs, func(dpi integrationWithSystemComponentsDataProvider, i *ent.Integration) (rez.SystemComponentsDataProvider, error) {
		return dpi.MakeSystemComponentsDataProvider(ctx, i)
	})
}
