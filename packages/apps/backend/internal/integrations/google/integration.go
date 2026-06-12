package google

import (
	"context"
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

var supportedCapabilities = []string{"video_conference"}

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

func (i *Integration) MaxInstalls() *int {
	return nil
}

func (i *Integration) IsAvailable() (bool, error) {
	// TODO: check config
	return true, nil
}

func (i *Integration) SupportedCapabilities() []string {
	return supportedCapabilities
}

func (i *Integration) OAuthInstallRequired() bool {
	return false
}

func (i *Integration) ValidateInstallationConfig(m map[string]any) (externalRef string, validationErr error) {
	//TODO implement me
	panic("implement me")
}

func (i *Integration) ValidateUserSettings(m map[string]any) error {
	//TODO implement me
	panic("implement me")
}

func (i *Integration) GetInstalledIntegration(intg *ent.Integration) rez.InstalledIntegration {
	return i.newInstalledIntegration(intg)
}

func (i *Integration) newInstalledIntegration(intg *ent.Integration) *InstalledIntegration {
	return &InstalledIntegration{intg: intg, incidents: i.incidents}
}

type InstalledIntegration struct {
	intg      *ent.Integration
	incidents rez.IncidentService
}

func (ii *InstalledIntegration) Integration() *ent.Integration {
	return ii.intg
}

func (ii *InstalledIntegration) ProviderName() string {
	return providerName
}

func (ii *InstalledIntegration) SanitizedInstallationConfig() map[string]any {
	return ii.intg.InstallationConfig
}

func (ii *InstalledIntegration) GetCapabilities() map[string]bool {
	return map[string]bool{
		"video_conference": ii.isVideoConferenceEnabled(),
	}
}

type installationConfig struct {
	ServiceAccountCredentials []byte
}

type userSettings struct {
	EnableVideoConference bool
}

func (ii *InstalledIntegration) config() (*installationConfig, error) {
	var cfg installationConfig
	if decErr := mapstructure.Decode(ii.intg.InstallationConfig, &cfg); decErr != nil {
		return nil, decErr
	}
	return &cfg, nil
}

func (ii *InstalledIntegration) userSettings() (*userSettings, error) {
	var settings userSettings
	if decErr := mapstructure.Decode(ii.intg.UserSettings, &settings); decErr != nil {
		return nil, decErr
	}
	return &settings, nil
}

func (ii *InstalledIntegration) isVideoConferenceEnabled() bool {
	if cfg, cfgErr := ii.config(); cfgErr != nil || cfg.ServiceAccountCredentials == nil {
		return false
	}
	if settings, settingsErr := ii.userSettings(); settingsErr != nil || !settings.EnableVideoConference {
		return false
	}
	return true
}

func (ii *InstalledIntegration) getAuthCredentials() (option.ClientOption, error) {
	cfg, cfgErr := ii.config()
	if cfgErr != nil || cfg.ServiceAccountCredentials == nil {
		return nil, fmt.Errorf("missing service account credentials")
	}
	return option.WithAuthCredentialsJSON(option.ServiceAccount, cfg.ServiceAccountCredentials), nil
}

func (ii *InstalledIntegration) MakeVideoConferenceService(ctx context.Context) (rez.VideoConferenceService, error) {
	return newMeetService(ii), nil
}
