package fakeprovider

import (
	"context"
	"encoding/json"

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

func (d *integration) Name() string {
	return integrationName
}

func (d *integration) IsAvailable() (bool, error) {
	return rez.Config.DebugMode(), nil
}

func (d *integration) SupportedDataKinds() []string {
	return supportedDataKinds
}

func (d *integration) OAuthConfigRequired() bool {
	return false
}

func (d *integration) GetConfiguredIntegration(i *ent.Integration) rez.ConfiguredIntegration {
	return &ConfiguredIntegration{intg: i}
}

type ConfiguredIntegration struct {
	intg *ent.Integration
}

func (ci *ConfiguredIntegration) Name() string {
	return integrationName
}

func (ci *ConfiguredIntegration) RawConfig() json.RawMessage {
	return ci.intg.Config
}

func (ci *ConfiguredIntegration) UserPreferences() map[string]any {
	return ci.intg.UserPreferences
}

func (ci *ConfiguredIntegration) EnabledDataKinds() []string {
	return supportedDataKinds
}

func (ci *ConfiguredIntegration) GetSanitizedConfig() (json.RawMessage, error) {
	return json.Marshal(ci.RawConfig())
}

type IntegrationConfig struct{}
