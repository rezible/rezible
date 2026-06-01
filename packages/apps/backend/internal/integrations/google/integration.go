package google

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/stretchr/objx"
	"google.golang.org/api/option"
)

const integrationName = "google"

var supportedDataKinds = []string{"video_conference"}

type Integration struct {
	users        rez.UserService
	integrations rez.IntegrationService
	messages     rez.MessageService
	incidents    rez.IncidentService
	eventAnnos   rez.EventAnnotationsService
}

func MakeIntegration(
	cfg rez.Config,
	users rez.UserService,
	integrations rez.IntegrationService,
	messages rez.MessageService,
	incidents rez.IncidentService,
	eventAnnos rez.EventAnnotationsService,
) (*Integration, error) {
	i := &Integration{
		users:        users,
		integrations: integrations,
		messages:     messages,
		incidents:    incidents,
		eventAnnos:   eventAnnos,
	}

	if msgsErr := i.registerMessageHandlers(); msgsErr != nil {
		return nil, fmt.Errorf("registering message handlers: %w", msgsErr)
	}

	return i, nil
}

func (i *Integration) Name() string {
	return integrationName
}

func (i *Integration) IsAvailable() (bool, error) {
	// TODO: check config
	return true, nil
}

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
	return i.newConfiguredIntegration(intg)
}

func (i *Integration) newConfiguredIntegration(intg *ent.Integration) *ConfiguredIntegration {
	return &ConfiguredIntegration{intg: intg, incidents: i.incidents}
}

type ConfiguredIntegration struct {
	intg      *ent.Integration
	incidents rez.IncidentService
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
	return ci.config().Exclude([]string{})
}

func (ci *ConfiguredIntegration) GetUserPreferences() map[string]any {
	return ci.userPreferences()
}

func (ci *ConfiguredIntegration) GetAvailableDataKinds() map[string]bool {
	return map[string]bool{
		"video_conference": ci.isVideoConferenceEnabled(),
	}
}

const (
	configServiceAccountCredentials = "service_account_credentials"
)

func (ci *ConfiguredIntegration) config() objx.Map {
	return objx.New(ci.intg.Config)
}

const (
	userPreferenceEnableVideoConferencing = "video_conferencing"
)

func (ci *ConfiguredIntegration) userPreferences() objx.Map {
	return objx.New(ci.intg.UserPreferences)
}

func (ci *ConfiguredIntegration) getServiceAccountCredentials() []byte {
	cfg := objx.New(ci.intg.Config)
	if v := cfg.Get(configServiceAccountCredentials); !v.IsNil() {
		if data, ok := v.Data().([]byte); ok {
			return data
		}
	}
	return nil
}

func (ci *ConfiguredIntegration) isVideoConferenceEnabled() bool {
	if creds := ci.getServiceAccountCredentials(); creds == nil {
		return false
	}
	return ci.userPreferences().Get(userPreferenceEnableVideoConferencing).Bool()
}

func (ci *ConfiguredIntegration) getAuthCredentials() (option.ClientOption, error) {
	creds := ci.getServiceAccountCredentials()
	if creds == nil {
		return nil, fmt.Errorf("missing service account credentials")
	}
	return option.WithAuthCredentialsJSON(option.ServiceAccount, creds), nil
}

func (ci *ConfiguredIntegration) MakeVideoConferenceService(ctx context.Context) (rez.VideoConferenceService, error) {
	return newMeetService(ci), nil
}
