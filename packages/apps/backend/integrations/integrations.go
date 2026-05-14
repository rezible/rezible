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
	"github.com/rezible/rezible/telemetry"
	"golang.org/x/oauth2"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/internal/integrations/fake"
	"github.com/rezible/rezible/internal/integrations/github"
	"github.com/rezible/rezible/internal/integrations/google"
	"github.com/rezible/rezible/internal/integrations/slack"
)

var (
	packageNameMap     = map[string]rez.IntegrationPackage{}
	availablePackages  []rez.IntegrationPackage
	supportedDataKinds []string
	setupFuncs         = []rez.PackageSetupFunc{
		fakeprovider.SetupIntegration,
		slack.SetupIntegration,
		google.SetupIntegration,
		github.SetupIntegration,
	}
)

func Setup(ctx context.Context, svcs *rez.Services) error {
	availablePackages = make([]rez.IntegrationPackage, 0, len(availablePackages))
	packageNameMap = make(map[string]rez.IntegrationPackage)
	enabledSupportedDataKinds := mapset.NewSet[string]()
	logger := telemetry.NewLogger(ctx, telemetry.WithLogPackage("integrations"))
	for _, setupFn := range setupFuncs {
		pkg, pkgErr := setupFn(ctx, svcs)
		if pkgErr != nil {
			funcName := runtime.FuncForPC(reflect.ValueOf(setupFn).Pointer()).Name()
			return fmt.Errorf("%s: %w", funcName, pkgErr)
		}
		available, configErr := pkg.IsAvailable()
		if !available || configErr != nil {
			lvl := slog.LevelInfo
			l := logger.With("name", pkg.Name())
			if configErr != nil {
				lvl = slog.LevelWarn
				l = l.With("config_error", configErr.Error())
			}
			l.Log(ctx, lvl, "integration not available")
			continue
		}
		logger.DebugContext(ctx, "loaded integration", "name", pkg.Name())
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
	ExtractIntegrationOptionsFromToken(*oauth2.Token) ([]rez.ExternalIntegrationOption, error)
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

type IntegrationPackageWithProviderEventProcessor interface {
	MakeProviderEventProcessor() rez.ProviderEventProcessor
}

func GetProviderEventProcessors() map[string]rez.ProviderEventProcessor {
	els := make(map[string]rez.ProviderEventProcessor)
	for _, pkg := range availablePackages {
		procPkg, ok := pkg.(IntegrationPackageWithProviderEventProcessor)
		if ok {
			els[pkg.Name()] = procPkg.MakeProviderEventProcessor()
		}
	}
	return els
}

type IntegrationPackageWithProviderEventQuerier interface {
	MakeProviderEventQuerier(*ent.Integration) (rez.ProviderEventQuerier, error)
}

func GetProviderEventQuerier(intg *ent.Integration) (rez.ProviderEventQuerier, error) {
	pkg, valid := packageNameMap[intg.Provider]
	if !valid {
		return nil, fmt.Errorf("unknown integration package: %s", intg.Provider)
	}
	querierPkg, ok := pkg.(IntegrationPackageWithProviderEventQuerier)
	if !ok {
		return nil, nil
	}
	return querierPkg.MakeProviderEventQuerier(intg)
}
