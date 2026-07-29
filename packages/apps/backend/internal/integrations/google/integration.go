package google

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-viper/mapstructure/v2"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"google.golang.org/api/option"
)

const (
	integrationName = "google"
	providerName    = "google"
)

type Integration struct {
	users        rez.UserService
	integrations rez.IntegrationService
	messages     rez.MessageService
	incidents    rez.IncidentService
	events       rez.EventsService
}

func MakeIntegration(
	cfg rez.Config,
	users rez.UserService,
	integrations rez.IntegrationService,
	messages rez.MessageService,
	incidents rez.IncidentService,
	events rez.EventsService,
) (*Integration, error) {
	i := &Integration{
		users:        users,
		integrations: integrations,
		messages:     messages,
		incidents:    incidents,
		events:       events,
	}

	if msgsErr := i.registerMessageHandlers(); msgsErr != nil {
		return nil, fmt.Errorf("registering message handlers: %w", msgsErr)
	}

	return i, nil
}

func (i *Integration) Name() string {
	return integrationName
}

func (i *Integration) DisplayName() string {
	return "Google Workspace"
}

func (i *Integration) Description() string {
	return "Interact with Google Workspace services"
}

func (i *Integration) Provider() string {
	return providerName
}

func (i *Integration) Capabilities() []string {
	return []string{"video_conferencing"}
}

func (i *Integration) MaxInstalls() *int {
	return new(1)
}

func (i *Integration) IsAvailable() (bool, error) {
	// TODO: check config
	return true, nil
}

func (i *Integration) OAuthInstallRequired() bool {
	return false
}

func (i *Integration) ValidateInstallationConfig(cfg []byte) (rez.IntegrationInstallationConfig, error) {
	return i.decodeValidateInstallationConfig(cfg)
}

func (i *Integration) decodeValidateInstallationConfig(m []byte) (*InstallationConfig, error) {
	var c InstallationConfig
	return &c, json.Unmarshal(m, &c)
}

func (i *Integration) ValidateUserSettings(m map[string]any) error {
	var settings UserSettings
	return mapstructure.Decode(m, &settings)
}

func (i *Integration) GetInstalledIntegration(intg *ent.Integration) (rez.InstalledIntegration, error) {
	return i.newInstalledIntegration(intg)
}

func (i *Integration) newInstalledIntegration(intg *ent.Integration) (*InstalledIntegration, error) {
	ii := &InstalledIntegration{
		intg:      intg,
		incidents: i.incidents,
	}
	cfg, cfgErr := i.decodeValidateInstallationConfig(intg.InstallationConfig)
	if cfgErr != nil {
		return nil, cfgErr
	}
	ii.config = cfg
	if decErr := mapstructure.Decode(intg.UserSettings, &ii.settings); decErr != nil {
		return nil, decErr
	}
	return ii, nil
}

type InstalledIntegration struct {
	intg      *ent.Integration
	incidents rez.IncidentService
	config    *InstallationConfig
	settings  *UserSettings
}

func (ii *InstalledIntegration) Integration() *ent.Integration {
	return ii.intg
}

func (ii *InstalledIntegration) Config() rez.IntegrationInstallationConfig {
	return ii.config
}

func (ii *InstalledIntegration) Capabilities() []string {
	return []string{"video_conferencing"}
}

func (ii *InstalledIntegration) isVideoConferenceEnabled() bool {
	if ii.config.ServiceAccountCredentials == nil {
		return false
	}
	if ii.settings == nil || !ii.settings.EnableVideoConference {
		return false
	}
	return true
}

func (ii *InstalledIntegration) getAuthCredentials() (option.ClientOption, error) {
	if ii.config.ServiceAccountCredentials == nil {
		return nil, fmt.Errorf("missing service account credentials")
	}
	return option.WithAuthCredentialsJSON(option.ServiceAccount, ii.config.ServiceAccountCredentials), nil
}

func (ii *InstalledIntegration) MakeVideoConferenceService(ctx context.Context) (rez.VideoConferenceService, error) {
	return newMeetService(ii), nil
}
