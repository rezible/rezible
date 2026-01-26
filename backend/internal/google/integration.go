package google

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	rez "github.com/rezible/rezible"
)

const integrationName = "google"

type integration struct {
	meetService *meetService
}

func SetupIntegration(ctx context.Context, svcs *rez.Services) (rez.IntegrationPackage, error) {
	ms, msErr := newMeetService(ctx, svcs)
	if msErr != nil {
		return nil, fmt.Errorf("unable to create MeetService: %v", msErr)
	}

	intg := &integration{
		meetService: ms,
	}
	return intg, nil
}

func (d *integration) Name() string {
	return integrationName
}

func (d *integration) Enabled() bool {
	// TODO: check config
	return true
}

func (d *integration) EventListeners() map[string]rez.EventListener {
	return nil
}

func (d *integration) WebhookHandlers() map[string]http.Handler {
	return nil
}

func (d *integration) GetVideoConferenceService() rez.VideoConferenceService {
	return d.meetService
}

func (d *integration) SupportedDataKinds() []string {
	return []string{"video_conferencing"}
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
	cfg.UserConfig.ServiceAccountCredentials = nil
	return json.Marshal(cfg)
}

type IntegrationConfig struct {
	UserConfig IntegrationUserConfig
}

type IntegrationUserConfig struct {
	ServiceAccountCredentials json.RawMessage
}

func lookupIntegrationConfig(ctx context.Context, integrations rez.IntegrationsService) (*IntegrationConfig, error) {
	intg, lookupErr := integrations.GetIntegration(ctx, integrationName)
	if lookupErr != nil {
		return nil, lookupErr
	}
	var cfg IntegrationConfig
	if jsonErr := json.Unmarshal(intg.Config, &cfg); jsonErr != nil {
		return nil, fmt.Errorf("unmarshal integration: %w", jsonErr)
	}
	return &cfg, nil
}
