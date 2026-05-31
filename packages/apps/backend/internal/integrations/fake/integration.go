package fakeprovider

import (
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

const integrationName = "fake"

type Integration struct {
	available bool
}

func MakeIntegration(cfg rez.Config) *Integration {
	return &Integration{
		available: cfg.App.DebugMode,
	}
}

func (i *Integration) Name() string {
	return integrationName
}

func (i *Integration) IsAvailable() (bool, error) {
	return i.available, nil
}

var supportedDataKinds = []string{"alerts", "incidents", "system_topology"}

func (i *Integration) SupportedDataKinds() []string {
	return supportedDataKinds
}

func (i *Integration) OAuthConfigRequired() bool {
	return false
}

func (i *Integration) ValidateUserConfig(cfg map[string]any) error {
	return nil
}

func (i *Integration) ValidateUserPreferences(prefs map[string]any) error {
	return nil
}

func (i *Integration) GetConfiguredIntegration(intg *ent.Integration) rez.ConfiguredIntegration {
	return &ConfiguredIntegration{intg: intg}
}

type ConfiguredIntegration struct {
	intg *ent.Integration
}

func (ci *ConfiguredIntegration) ID() uuid.UUID {
	return ci.intg.ID
}

func (ci *ConfiguredIntegration) Provider() string {
	return ci.intg.Provider
}

func (ci *ConfiguredIntegration) DisplayName() string {
	return ci.intg.DisplayName
}

func (ci *ConfiguredIntegration) ExternalRef() string {
	return ci.intg.ExternalRef
}

func (ci *ConfiguredIntegration) GetSanitizedConfig() map[string]any {
	return ci.intg.Config
}

func (ci *ConfiguredIntegration) GetUserPreferences() map[string]any {
	return ci.intg.UserPreferences
}

func (ci *ConfiguredIntegration) GetAvailableDataKinds() map[string]bool {
	return map[string]bool{}
}

type IntegrationConfig struct{}
