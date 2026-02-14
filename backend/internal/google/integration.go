package google

import (
	"context"
	"encoding/json"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"google.golang.org/api/option"
)

const integrationName = "google"

var supportedDataKinds = []string{"video_conference"}

type integration struct {
	services *rez.Services
}

func SetupIntegration(ctx context.Context, svcs *rez.Services) (rez.IntegrationPackage, error) {
	i := &integration{services: svcs}

	if msgsErr := i.registerMessageHandlers(); msgsErr != nil {
		return nil, fmt.Errorf("registering message handlers: %w", msgsErr)
	}

	return i, nil
}

func (i *integration) Name() string {
	return integrationName
}

func (i *integration) IsAvailable() (bool, error) {
	// TODO: check config
	return true, nil
}

func (i *integration) SupportedDataKinds() []string {
	return supportedDataKinds
}

func (i *integration) OAuthConfigRequired() bool {
	return false
}

func (i *integration) lookupConfiguredIntegration(ctx context.Context) (*ConfiguredIntegration, error) {
	intg, intgErr := i.services.Integrations.Get(ctx, integrationName)
	if intgErr != nil {
		if ent.IsNotFound(intgErr) {
			return nil, nil
		}
		return nil, fmt.Errorf("error looking up integration: %w", intgErr)
	}
	return i.newConfiguredIntegration(intg), nil
}

func (i *integration) GetConfiguredIntegration(intg *ent.Integration) rez.ConfiguredIntegration {
	return i.newConfiguredIntegration(intg)
}

func (i *integration) newConfiguredIntegration(intg *ent.Integration) *ConfiguredIntegration {
	return &ConfiguredIntegration{svcs: i.services, intg: intg}
}

type ConfiguredIntegration struct {
	svcs *rez.Services
	intg *ent.Integration
}

type integrationConfig struct {
	UserConfig integrationUserConfig
}

type integrationUserConfig struct {
	ServiceAccountCredentials json.RawMessage
}

func (ci *ConfiguredIntegration) isServiceAccountConfigured() bool {
	var cfg integrationConfig
	if jsonErr := json.Unmarshal(ci.intg.Config, &cfg); jsonErr != nil {
		return false
	}
	return cfg.UserConfig.ServiceAccountCredentials != nil
}

func (ci *ConfiguredIntegration) getUserPreference(key string, defaultVal any) any {
	if pref, ok := ci.intg.UserPreferences[key]; ok {
		return pref
	}
	return defaultVal
}

func (ci *ConfiguredIntegration) getBoolUserPreference(key string, defaultVal bool) bool {
	pref := ci.getUserPreference(key, defaultVal)
	switch v := pref.(type) {
	case bool:
		return v
	case string:
		return v != "false"
	default:
		return defaultVal
	}
}

func (ci *ConfiguredIntegration) isVideoConferenceEnabled() bool {
	if !ci.isServiceAccountConfigured() {
		return false
	}
	return ci.getBoolUserPreference(PreferenceEnableIncidentVideoConferences, true)
}

func (ci *ConfiguredIntegration) Name() string {
	return integrationName
}

func (ci *ConfiguredIntegration) EnabledDataKinds() []string {
	return supportedDataKinds
}

func (ci *ConfiguredIntegration) RawConfig() json.RawMessage {
	return ci.intg.Config
}

func (ci *ConfiguredIntegration) UserPreferences() map[string]any {
	return ci.intg.UserPreferences
}

func (ci *ConfiguredIntegration) unmarshalConfig() (*integrationConfig, error) {
	var cfg integrationConfig
	if err := json.Unmarshal(ci.RawConfig(), &cfg); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}
	return &cfg, nil
}

func (ci *ConfiguredIntegration) GetSanitizedConfig() (json.RawMessage, error) {
	cfg, cfgErr := ci.unmarshalConfig()
	if cfgErr != nil {
		return nil, cfgErr
	}
	cfg.UserConfig.ServiceAccountCredentials = nil
	return json.Marshal(cfg)
}

func (ci *ConfiguredIntegration) getServiceAccountAuthCredentials() (option.ClientOption, error) {
	cfg, cfgErr := ci.unmarshalConfig()
	if cfgErr != nil {
		return nil, cfgErr
	}
	if cfg.UserConfig.ServiceAccountCredentials == nil {
		return nil, fmt.Errorf("missing service account credentials")
	}
	authOpt := option.WithAuthCredentialsJSON(option.ServiceAccount, cfg.UserConfig.ServiceAccountCredentials)
	return authOpt, nil
}

func (ci *ConfiguredIntegration) VideoConferenceIntegration(ctx context.Context) (rez.VideoConferenceIntegration, error) {
	svcAuthOpt, authErr := ci.getServiceAccountAuthCredentials()
	if authErr != nil {
		return nil, authErr
	}
	return newMeetService(ctx, ci.svcs.Incidents, svcAuthOpt)
}
