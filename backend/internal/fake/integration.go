package fakeprovider

import (
	"context"
	"encoding/json"
	"fmt"

	rez "github.com/rezible/rezible"
)

const integrationName = "fake"

type integration struct{}

func SetupIntegration(ctx context.Context, svcs *rez.Services) (rez.IntegrationPackage, error) {
	intg := &integration{}
	return intg, nil
}

func (d *integration) Name() string {
	return integrationName
}

func (d *integration) Enabled() bool {
	return rez.Config.DebugMode()
}

func (d *integration) SupportedDataKinds() []string {
	return []string{}
}

func (d *integration) OAuthConfigRequired() bool {
	return false
}

func (d *integration) ValidateConfig(raw json.RawMessage) (bool, error) {
	return true, nil
}

func (d *integration) MergeUserConfig(full json.RawMessage, userCfg json.RawMessage) (json.RawMessage, error) {
	var cfg IntegrationConfig
	if cfgErr := json.Unmarshal(full, &cfg); cfgErr != nil {
		return nil, fmt.Errorf("failed to decode integration config: %w", cfgErr)
	}
	if userCfgErr := json.Unmarshal(userCfg, &cfg.UserConfig); userCfgErr != nil {
		return nil, fmt.Errorf("failed to decode user config: %w", userCfgErr)
	}
	return json.Marshal(cfg)
}

func (d *integration) GetSanitizedConfig(rawCfg json.RawMessage) (json.RawMessage, error) {
	var cfg IntegrationConfig
	if rawErr := json.Unmarshal(rawCfg, &cfg); rawErr != nil {
		return nil, fmt.Errorf("failed to decode integration config: %w", rawErr)
	}
	return json.Marshal(cfg)
}

type IntegrationConfig struct {
	UserConfig struct{}
}

func (c *IntegrationConfig) GetSanitized() (json.RawMessage, error) {
	return []byte("{}"), nil
}

func (c *IntegrationConfig) MergeUserConfig(rawUserCfg json.RawMessage) (json.RawMessage, error) {
	return []byte("{}"), nil
}
