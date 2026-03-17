package fakeprovider

import (
	"context"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

const integrationName = "fake"

var supportedDataKinds = []string{}

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

func (ci *ConfiguredIntegration) Name() string {
	return integrationName
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
