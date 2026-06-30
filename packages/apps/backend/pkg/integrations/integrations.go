package integrations

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/go-viper/mapstructure/v2"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"golang.org/x/oauth2"
)

func GetSourceQueryCursor(cursors map[string]string, source string) (string, bool) {
	sc, ok := cursors[source]
	return sc, ok || len(cursors) == 0
}

type PackageRegistry struct {
	pkgsMu            sync.RWMutex
	nameMap           map[string]rez.IntegrationPackage
	availablePackages []rez.IntegrationPackage
}

func NewPackageRegistry() *PackageRegistry {
	return &PackageRegistry{
		nameMap: make(map[string]rez.IntegrationPackage),
	}
}

func (r *PackageRegistry) RegisterPackage(pkg rez.IntegrationPackage) error {
	r.pkgsMu.Lock()
	defer r.pkgsMu.Unlock()

	available, configErr := pkg.IsAvailable()
	slog.Debug("register integration package",
		"name", pkg.Name(),
		"available", available,
		"error", configErr,
	)
	if !available {
		return configErr
	}
	r.availablePackages = append(r.availablePackages, pkg)
	r.nameMap[pkg.Name()] = pkg

	return nil
}

func (r *PackageRegistry) GetAvailable() []rez.IntegrationPackage {
	return r.availablePackages
}

func (r *PackageRegistry) GetPackage(name string) (rez.IntegrationPackage, error) {
	p, valid := r.nameMap[name]
	if !valid {
		return nil, fmt.Errorf("unknown integration: %s", name)
	}
	return p, nil
}

type IntegrationWithWebhookHandler interface {
	WebhookHandler() http.Handler
}

func (r *PackageRegistry) GetWebhookHandlers() map[string]http.Handler {
	whs := make(map[string]http.Handler)
	for _, pkg := range r.availablePackages {
		if whPkg, hasWebhook := pkg.(IntegrationWithWebhookHandler); hasWebhook {
			whs[pkg.Name()] = whPkg.WebhookHandler()
		}
	}
	return whs
}

func (r *PackageRegistry) GetProviderEventQuerier(intg *ent.Integration) (rez.ProviderEventQuerier, error) {
	type IntegrationWithProviderEventQuerier interface {
		MakeProviderEventQuerier(*ent.Integration) (rez.ProviderEventQuerier, error)
	}
	pkg, valid := r.nameMap[intg.IntegrationName]
	if !valid {
		return nil, fmt.Errorf("unknown integration package: %s", intg.IntegrationName)
	}
	if querierPkg, ok := pkg.(IntegrationWithProviderEventQuerier); ok {
		return querierPkg.MakeProviderEventQuerier(intg)
	}
	return nil, fmt.Errorf("integration does not provide an event querier")
}

type IntegrationWithOAuth2Flow interface {
	OAuth2Config() *oauth2.Config
	RetrieveInstallationTargetOptions(context.Context, *oauth2.Token) ([]rez.IntegrationInstallationTarget, error)
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
		return nil, fmt.Errorf("empty integration oauth2 configuration")
	}
	return oauth2Intg, nil
}

func EncodeInstallationTargetOptions(options []rez.IntegrationInstallationTarget) ([]map[string]any, error) {
	result := make([]map[string]any, 0, len(options))
	for _, opt := range options {
		var optMap map[string]any
		if encErr := mapstructure.Decode(opt, &optMap); encErr != nil {
			return nil, fmt.Errorf("failed to encode to map: %w", encErr)
		}
		result = append(result, optMap)
	}
	return result, nil
}

func DecodeInstallationTargetOptions(opts []map[string]any) ([]rez.IntegrationInstallationTarget, error) {
	result := make([]rez.IntegrationInstallationTarget, 0, len(opts))
	for _, opt := range opts {
		var decoded rez.IntegrationInstallationTarget
		if decErr := mapstructure.Decode(opt, &decoded); decErr != nil {
			return nil, fmt.Errorf("failed to decode option from map: %w", decErr)
		}
		result = append(result, decoded)
	}
	return result, nil
}
