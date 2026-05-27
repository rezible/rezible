package integrations

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
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
	setupFuncs = []rez.PackageSetupFunc{
		fakeprovider.SetupIntegration,
		slack.SetupIntegration,
		google.SetupIntegration,
		github.SetupIntegration,
	}
)

type PackageRegistry struct {
	nameMap           map[string]rez.IntegrationPackage
	availablePackages []rez.IntegrationPackage
}

func NewPackageRegistry() *PackageRegistry {
	return &PackageRegistry{
		nameMap: make(map[string]rez.IntegrationPackage),
	}
}

type integrationPackageWithEventListeners interface {
	EventListeners() map[string]rez.EventListener
}

func (r *PackageRegistry) SetupPackages(ctx context.Context, svcs *rez.Services) (map[string]rez.EventListener, error) {
	enabledSupportedDataKinds := mapset.NewSet[string]()
	els := make(map[string]rez.EventListener)
	logger := telemetry.NewPackageLogger("integrations")
	for _, setupFn := range setupFuncs {
		pkg, pkgErr := setupFn(ctx, svcs)
		if pkgErr != nil {
			funcName := runtime.FuncForPC(reflect.ValueOf(setupFn).Pointer()).Name()
			return nil, fmt.Errorf("%s: %w", funcName, pkgErr)
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
		r.availablePackages = append(r.availablePackages, pkg)
		r.nameMap[pkg.Name()] = pkg
		enabledSupportedDataKinds.Append(pkg.SupportedDataKinds()...)

		if elPkg, ok := pkg.(integrationPackageWithEventListeners); ok {
			for name, listener := range elPkg.EventListeners() {
				els[name] = listener
			}
		}
	}

	return els, nil
}

func (r *PackageRegistry) GetPackage(name string) (rez.IntegrationPackage, error) {
	p, valid := r.nameMap[name]
	if !valid {
		return nil, fmt.Errorf("unknown integration: %s", name)
	}
	return p, nil
}

func (r *PackageRegistry) GetAvailable() []rez.IntegrationPackage {
	return r.availablePackages
}

func (r *PackageRegistry) GetWebhookHandlers() map[string]http.Handler {
	type IntegrationWithWebhookHandler interface {
		WebhookHandler() http.Handler
	}
	whs := make(map[string]http.Handler)
	for _, pkg := range r.availablePackages {
		if whPkg, hasWebhook := pkg.(IntegrationWithWebhookHandler); hasWebhook {
			whs[pkg.Name()] = whPkg.WebhookHandler()
		}
	}
	return whs
}

type IntegrationWithOAuth2Flow interface {
	OAuth2Config() *oauth2.Config
	ExtractIntegrationOptionsFromToken(*oauth2.Token) ([]rez.ExternalIntegrationOption, error)
}

func (r *PackageRegistry) GetOAuthIntegration(name string) (IntegrationWithOAuth2Flow, error) {
	ip, ipErr := r.GetPackage(name)
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

func (r *PackageRegistry) GetProviderEventProcessors() map[string]rez.ProviderEventProcessor {
	type IntegrationWithProviderEventProcessor interface {
		MakeProviderEventProcessor() rez.ProviderEventProcessor
	}
	els := make(map[string]rez.ProviderEventProcessor)
	for _, pkg := range r.availablePackages {
		if procPkg, ok := pkg.(IntegrationWithProviderEventProcessor); ok {
			els[pkg.Name()] = procPkg.MakeProviderEventProcessor()
		}
	}
	return els
}

func (r *PackageRegistry) GetProviderEventQuerier(intg *ent.Integration) (rez.ProviderEventQuerier, error) {
	type IntegrationWithProviderEventQuerier interface {
		MakeProviderEventQuerier(*ent.Integration) (rez.ProviderEventQuerier, error)
	}
	pkg, valid := r.nameMap[intg.Provider]
	if !valid {
		return nil, fmt.Errorf("unknown integration package: %s", intg.Provider)
	}
	if querierPkg, ok := pkg.(IntegrationWithProviderEventQuerier); ok {
		return querierPkg.MakeProviderEventQuerier(intg)
	}
	return nil, nil
}
