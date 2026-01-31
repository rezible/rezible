package google

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"google.golang.org/api/option"
)

const integrationName = "google"

var supportedDataKinds = []string{"video_conferencing"}

type integration struct {
	services *rez.Services
}

func SetupIntegration(ctx context.Context, svcs *rez.Services) (rez.IntegrationPackage, error) {
	intg := &integration{services: svcs}
	return intg, nil
}

func (i *integration) Name() string {
	return integrationName
}

func (i *integration) IsAvailable() (bool, error) {
	// TODO: check config
	return true, nil
}

func (i *integration) EventListeners() map[string]rez.EventListener {
	return nil
}

func (i *integration) WebhookHandlers() map[string]http.Handler {
	return nil
}

func (i *integration) SupportedDataKinds() []string {
	return supportedDataKinds
}

func (i *integration) OAuthConfigRequired() bool {
	return false
}

func (i *integration) GetConfiguredIntegration(intg *ent.Integration) rez.ConfiguredIntegration {
	return &ConfiguredIntegration{intg: intg, svcs: i.services}
}

type ConfiguredIntegration struct {
	intg *ent.Integration
	svcs *rez.Services
}

type integrationConfig struct {
	UserConfig integrationUserConfig
}

type integrationUserConfig struct {
	ServiceAccountCredentials json.RawMessage
}

func (ci *ConfiguredIntegration) Name() string {
	return integrationName
}

func (ci *ConfiguredIntegration) RawConfig() json.RawMessage {
	return ci.intg.Config
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

func (ci *ConfiguredIntegration) UserPreferences() map[string]any {
	return ci.intg.UserPreferences
}

func (ci *ConfiguredIntegration) EnabledDataKinds() []string {
	return supportedDataKinds
}

func (ci *ConfiguredIntegration) getServiceAccountAuthCredentials() (option.ClientOption, error) {
	cfg, cfgErr := ci.unmarshalConfig()
	if cfgErr != nil {
		return nil, cfgErr
	}
	authOpt := option.WithAuthCredentialsJSON(option.ServiceAccount, cfg.UserConfig.ServiceAccountCredentials)
	return authOpt, nil
}

func (ci *ConfiguredIntegration) VideoConferenceIntegration(ctx context.Context) (rez.VideoConferenceIntegration, error) {
	svcAuthOpt, authErr := ci.getServiceAccountAuthCredentials()
	if authErr != nil {
		return nil, authErr
	}
	return newMeetService(ctx, ci.svcs.Messages, svcAuthOpt)
}
