package integrations

import (
	"fmt"
	"log/slog"
	"net/http"

	"golang.org/x/oauth2"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type PackageRegistry struct {
	logger            *slog.Logger
	nameMap           map[string]rez.IntegrationPackage
	availablePackages []rez.IntegrationPackage
}

func (r *PackageRegistry) RegisterPackage(pkg rez.IntegrationPackage) error {
	available, configErr := pkg.IsAvailable()
	if !available {
		nameAttr := slog.Any("name", pkg.Name())
		if configErr != nil {
			errAttr := slog.Any("config_error", configErr.Error())
			slog.Warn("integration config error", nameAttr, errAttr)
		} else {
			slog.Info("integration not available", nameAttr)
		}
		return configErr
	}
	r.logger.Debug("loaded integration", "name", pkg.Name())
	r.availablePackages = append(r.availablePackages, pkg)
	r.nameMap[pkg.Name()] = pkg

	return nil
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
