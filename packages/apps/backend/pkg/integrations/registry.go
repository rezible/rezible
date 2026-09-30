package integrations

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"

	"github.com/firebase/genkit/go/ai"

	rez "github.com/rezible/rezible"
)

type Registry struct {
	available   []rez.IntegrationDefinition
	definitions map[string]rez.IntegrationDefinition
}

func NewRegistry(defs ...rez.IntegrationDefinition) (*Registry, error) {
	r := &Registry{
		available:   make([]rez.IntegrationDefinition, 0, len(defs)),
		definitions: make(map[string]rez.IntegrationDefinition, len(defs)),
	}
	for _, def := range defs {
		available, availableErr := def.IsAvailable()
		if availableErr != nil {
			return nil, fmt.Errorf("check availability of integration %s: %w", def.Name(), availableErr)
		}
		slog.Debug("register integration",
			"integration", def.Name(),
			"available", available,
		)
		if !available {
			continue
		}
		if _, exists := r.definitions[def.Name()]; exists {
			return nil, fmt.Errorf("register integration %s: name already registered", def.Name())
		}
		r.available = append(r.available, def)
		r.definitions[def.Name()] = def
	}
	return r, nil
}

// GetAvailable returns the available definitions in registration order.
func (r *Registry) GetAvailable() []rez.IntegrationDefinition {
	return r.available
}

// Get returns the available definition with this name, or ErrUnknownIntegration.
func (r *Registry) Get(name string) (rez.IntegrationDefinition, error) {
	def, exists := r.definitions[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrUnknownIntegration, name)
	}
	return def, nil
}

// Lookup returns the named definition as capability T.
// It returns ErrUnknownIntegration or ErrCapabilityNotSupported.
func (r *Registry) Lookup[T any](name string) (T, error) {
	var capability T
	def, getErr := r.Get(name)
	if getErr != nil {
		return capability, getErr
	}
	capability, supported := def.(T)
	if !supported {
		return capability, fmt.Errorf("%w: %s does not implement %s", ErrCapabilityNotSupported, name, reflect.TypeFor[T]())
	}
	return capability, nil
}

// All returns every available definition that implements capability T, in registration order.
func (r *Registry) All[T any]() []T {
	capabilities := make([]T, 0, len(r.available))
	for _, def := range r.available {
		if capability, supported := def.(T); supported {
			capabilities = append(capabilities, capability)
		}
	}
	return capabilities
}

// GetOAuth2FlowIntegration also rejects a definition with a nil OAuth2Config.
func (r *Registry) GetOAuth2FlowIntegration(name string) (rez.OAuth2FlowIntegration, error) {
	oauthIntg, lookupErr := r.Lookup[rez.OAuth2FlowIntegration](name)
	if lookupErr != nil {
		return nil, lookupErr
	}
	if oauthIntg.OAuth2Config() == nil {
		return nil, fmt.Errorf("%w: %s has no oauth2 configuration", ErrCapabilityNotSupported, name)
	}
	return oauthIntg, nil
}

// GetProviderEventQuerier builds the querier for an installation.
func (r *Registry) GetProviderEventQuerier(ii rez.InstalledIntegration) (rez.ProviderEventQuerier, error) {
	intg := ii.Integration()
	querierIntg, lookupErr := r.Lookup[IntegrationWithProviderEventQuerier](intg.Name)
	if lookupErr != nil {
		return nil, lookupErr
	}
	return querierIntg.MakeProviderEventQuerier(intg)
}

// GetAvailableWebhookHandlers maps each available definition name to its webhook handler.
func (r *Registry) GetAvailableWebhookHandlers() map[string]http.Handler {
	handlers := make(map[string]http.Handler)
	for _, def := range r.available {
		if webhookIntg, hasWebhook := def.(IntegrationWithWebhookHandler); hasWebhook {
			handlers[def.Name()] = webhookIntg.WebhookHandler()
		}
	}
	return handlers
}

// GetAvailableAgentTools groups installations by definition and collects each definition's agent tools.
func (r *Registry) GetAvailableAgentTools(ctx context.Context, intgs []rez.InstalledIntegration, params rez.GetAvailableAiAgentToolsParams) (map[rez.IntegrationDefinition][]ai.Tool, error) {
	installationsByName := make(map[string][]rez.InstalledIntegration)
	for _, ii := range intgs {
		name := ii.Integration().Name
		installationsByName[name] = append(installationsByName[name], ii)
	}

	toolsByDefinition := make(map[rez.IntegrationDefinition][]ai.Tool)
	for name, installations := range installationsByName {
		def, getErr := r.Get(name)
		if getErr != nil {
			slog.ErrorContext(ctx, "failed to get integration",
				"integration", name,
				"error", getErr)
			continue
		}
		toolsIntg, providesTools := def.(IntegrationWithAgentToolProvider)
		if !providesTools {
			continue
		}
		tools, toolsErr := toolsIntg.GetAvailableAgentTools(ctx, installations, params)
		if toolsErr != nil {
			return nil, fmt.Errorf("get agent tools for integration %s: %w", name, toolsErr)
		}
		toolsByDefinition[def] = tools
	}
	return toolsByDefinition, nil
}
