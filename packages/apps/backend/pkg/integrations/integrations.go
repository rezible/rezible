package integrations

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/firebase/genkit/go/ai"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type Registry struct {
	pkgsMu            sync.RWMutex
	nameMap           map[string]rez.IntegrationDefinition
	availablePackages []rez.IntegrationDefinition
}

func NewRegistry() *Registry {
	return &Registry{
		nameMap: make(map[string]rez.IntegrationDefinition),
	}
}

func (r *Registry) Register(pkg rez.IntegrationDefinition) error {
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

func (r *Registry) GetAvailable() []rez.IntegrationDefinition {
	return r.availablePackages
}

func (r *Registry) Get(name string) (rez.IntegrationDefinition, error) {
	p, valid := r.nameMap[name]
	if !valid {
		return nil, fmt.Errorf("unknown integration: %s", name)
	}
	return p, nil
}

type IntegrationWithWebhookHandler interface {
	WebhookHandler() http.Handler
}

func (r *Registry) GetAvailableWebhookHandlers() map[string]http.Handler {
	whs := make(map[string]http.Handler)
	for _, pkg := range r.availablePackages {
		if whPkg, hasWebhook := pkg.(IntegrationWithWebhookHandler); hasWebhook {
			whs[pkg.Name()] = whPkg.WebhookHandler()
		}
	}
	return whs
}

type IntegrationWithProviderEventQuerier interface {
	MakeProviderEventQuerier(*ent.Integration) (rez.ProviderEventQuerier, error)
}

func (r *Registry) GetProviderEventQuerier(ii rez.InstalledIntegration) (rez.ProviderEventQuerier, error) {
	intg := ii.Integration()
	pkg, valid := r.nameMap[intg.Name]
	if !valid {
		return nil, fmt.Errorf("unknown integration package: %s", intg.Name)
	}
	if querierPkg, ok := pkg.(IntegrationWithProviderEventQuerier); ok {
		return querierPkg.MakeProviderEventQuerier(intg)
	}
	return nil, fmt.Errorf("integration does not provide an event querier")
}

func (r *Registry) GetOAuth2FlowIntegration(name string) (rez.OAuth2FlowIntegration, error) {
	ip, ipErr := r.Get(name)
	if ipErr != nil {
		return nil, fmt.Errorf("invalid integration %s: %w", name, ipErr)
	}
	oauth2Intg, ok := ip.(rez.OAuth2FlowIntegration)
	if !ok {
		return nil, fmt.Errorf("oauth2 flow not supported for integration %s", name)
	}
	if oauth2Intg.OAuth2Config() == nil {
		return nil, fmt.Errorf("empty integration oauth2 configuration")
	}
	return oauth2Intg, nil
}

type IntegrationWithAgentToolProvider interface {
	GetAvailableAgentTools(context.Context, []rez.InstalledIntegration, rez.GetAvailableAgentToolsParams) ([]ai.Tool, error)
}

func (r *Registry) GetAvailableAgentTools(ctx context.Context, intgs []rez.InstalledIntegration, params rez.GetAvailableAgentToolsParams) (map[rez.IntegrationDefinition][]ai.Tool, error) {
	packageMap := make(map[string][]rez.InstalledIntegration)
	for _, ii := range intgs {
		pkgName := ii.Integration().Name
		packageMap[pkgName] = append(packageMap[pkgName], ii)
	}

	pkgToolsMap := make(map[rez.IntegrationDefinition][]ai.Tool)
	for pkgName, installations := range packageMap {
		pkg, pkgErr := r.Get(pkgName)
		if pkgErr != nil {
			slog.ErrorContext(ctx, "failed to get integration package",
				"integration", pkgName,
				"error", pkgErr)
			continue
		}
		if toolsPkg, providesTools := pkg.(IntegrationWithAgentToolProvider); providesTools {
			pkgTools, toolsErr := toolsPkg.GetAvailableAgentTools(ctx, installations, params)
			if toolsErr != nil {
				return nil, fmt.Errorf("get agent tools for integration %s: %w", pkg.Name(), toolsErr)
			}
			pkgToolsMap[pkg] = pkgTools
		}
	}
	return pkgToolsMap, nil
}
