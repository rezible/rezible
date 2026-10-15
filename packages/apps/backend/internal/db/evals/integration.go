package evals

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

// EvaluationProviderName names the evaluation integration and the provider of every event a scenario ingests.
const EvaluationProviderName = "evaluation"

// Integration is installed in each agent evaluation's tenant. Its event processor passes a scenario's events
// into the pipeline. Only agent evaluation runs register it.
type Integration struct{}

func NewIntegration() *Integration {
	return &Integration{}
}

func (i *Integration) Name() string {
	return EvaluationProviderName
}

func (i *Integration) DisplayName() string {
	return "Evaluation"
}

func (i *Integration) Description() string {
	return "Sets up an agent evaluation scenario's environment."
}

func (i *Integration) Provider() string {
	return EvaluationProviderName
}

func (i *Integration) Capabilities() []string {
	return nil
}

func (i *Integration) MaxInstalls() *int {
	return new(1)
}

func (i *Integration) IsAvailable() (bool, error) {
	return true, nil
}

func (i *Integration) OAuthInstallRequired() bool {
	return false
}

func (i *Integration) InstallationLinks() []rez.IntegrationInstallationLink {
	return nil
}

// ValidateInstallationConfig gives a new installation its own ref.
func (i *Integration) ValidateInstallationConfig(raw []byte) (rez.IntegrationInstallationConfig, error) {
	var cfg integrationConfig
	if decodeErr := json.Unmarshal(raw, &cfg); decodeErr != nil {
		return nil, fmt.Errorf("decode installation config: %w", decodeErr)
	}
	if cfg.InstallationRef == "" {
		cfg.InstallationRef = uuid.NewString()
	}
	return &cfg, nil
}

func (i *Integration) ValidateUserSettings(map[string]any) error {
	return nil
}

func (i *Integration) GetInstalledIntegration(intg *ent.Integration) (rez.InstalledIntegration, error) {
	cfg, cfgErr := i.ValidateInstallationConfig(intg.InstallationConfig)
	if cfgErr != nil {
		return nil, cfgErr
	}
	return &installedIntegration{intg: intg, config: cfg}, nil
}

type installedIntegration struct {
	intg   *ent.Integration
	config rez.IntegrationInstallationConfig
}

func (ii *installedIntegration) Integration() *ent.Integration {
	return ii.intg
}

func (ii *installedIntegration) Config() rez.IntegrationInstallationConfig {
	return ii.config
}

func (ii *installedIntegration) Capabilities() []string {
	return nil
}

type integrationConfig struct {
	InstallationRef string `json:"installation_ref"`
}

func (c *integrationConfig) Encode() ([]byte, error) {
	return json.Marshal(c)
}

func (c *integrationConfig) InstallationTargetRef() rez.ProviderResourceRef {
	return rez.ProviderResourceRef{
		Provider:          EvaluationProviderName,
		ProviderNamespace: EvaluationProviderName,
		ResourceRef:       c.InstallationRef,
	}
}

// EventProcessor passes a scenario event through: the provider event's attributes are a ScenarioEvent, already
// in Rezible's shape. It has no dependencies.
type EventProcessor struct{}

func (p EventProcessor) ProcessProviderEvent(_ context.Context, prov rez.ProviderEvent) (ent.NormalizedEvents, error) {
	var event ScenarioEvent
	if decodeErr := json.Unmarshal(prov.Attributes, &event); decodeErr != nil {
		return nil, fmt.Errorf("decode scenario event: %w", decodeErr)
	}
	if event.OccurredAt == nil {
		return nil, fmt.Errorf("scenario event %s has no occurred_at", prov.ProviderEventRef)
	}
	attributes, completeErr := completeEntityRefs(event.Attributes, prov.ProviderNamespace)
	if completeErr != nil {
		return nil, fmt.Errorf("scenario event %s: %w", prov.ProviderEventRef, completeErr)
	}
	normalized := &ent.NormalizedEvent{
		Provider:            EvaluationProviderName,
		ProviderNamespace:   prov.ProviderNamespace,
		ProviderResourceRef: event.ResourceRef,
		ProviderEventSource: prov.ProviderEventSource,
		ProviderEventRef:    prov.ProviderEventRef,
		Kind:                event.Kind,
		OccurredAt:          event.OccurredAt.UTC(),
		ReceivedAt:          prov.ReceivedAt,
		Attributes:          attributes,
	}
	return ent.NormalizedEvents{normalized}, nil
}

// completeEntityRefs gives every object in the attributes that has a resource_ref and no provider this
// integration's provider and the namespace, so a scenario writes an entity reference by its resource ref alone.
func completeEntityRefs(raw json.RawMessage, namespace string) ([]byte, error) {
	var attributes any
	if decodeErr := json.Unmarshal(raw, &attributes); decodeErr != nil {
		return nil, fmt.Errorf("decode attributes: %w", decodeErr)
	}
	var complete func(value any)
	complete = func(value any) {
		switch node := value.(type) {
		case map[string]any:
			if _, isRef := node["resource_ref"]; isRef {
				if provider, _ := node["provider"].(string); provider == "" {
					node["provider"] = EvaluationProviderName
					node["provider_namespace"] = namespace
				}
			}
			for _, child := range node {
				complete(child)
			}
		case []any:
			for _, child := range node {
				complete(child)
			}
		}
	}
	complete(attributes)
	return json.Marshal(attributes)
}
