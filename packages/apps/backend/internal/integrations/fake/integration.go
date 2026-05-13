package fakeprovider

import (
	"context"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

const integrationName = "fake"

var supportedDataKinds = []string{"topology_entities"}

type integration struct{}

func SetupIntegration(ctx context.Context, svcs *rez.Services) (rez.IntegrationPackage, error) {
	intg := &integration{}
	return intg, nil
}

func (i *integration) Name() string {
	return integrationName
}

func (i *integration) IsAvailable() (bool, error) {
	return rez.Config.DebugMode(), nil
}

func (i *integration) SupportedDataKinds() []string {
	return supportedDataKinds
}

func (i *integration) OAuthConfigRequired() bool {
	return false
}

func (i *integration) ValidateConfig(cfg map[string]any) error {
	return nil
}

func (i *integration) ValidateUserPreferences(prefs map[string]any) error {
	return nil
}

func (i *integration) GetConfiguredIntegration(intg *ent.Integration) rez.ConfiguredIntegration {
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

func (ci *ConfiguredIntegration) GetDataKinds() map[string]bool {
	return map[string]bool{}
}

type IntegrationConfig struct{}
