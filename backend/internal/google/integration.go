package google

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/stretchr/objx"
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

func (i *integration) ValidateConfig(cfg map[string]any) error {
	return nil
}

func (i *integration) ValidateUserPreferences(prefs map[string]any) error {
	return nil
}

func (i *integration) GetConfiguredIntegration(intg *ent.Integration) rez.ConfiguredIntegration {
	return newConfiguredIntegration(intg, i.services)
}

func newConfiguredIntegration(intg *ent.Integration, svcs *rez.Services) *ConfiguredIntegration {
	return &ConfiguredIntegration{intg: intg, svcs: svcs}
}

type ConfiguredIntegration struct {
	svcs *rez.Services
	intg *ent.Integration
}

func (ci *ConfiguredIntegration) Name() string {
	return integrationName
}

func (ci *ConfiguredIntegration) GetSanitizedConfig() map[string]any {
	return ci.config().Exclude([]string{})
}

func (ci *ConfiguredIntegration) GetUserPreferences() map[string]any {
	return ci.userPreferences()
}

func (ci *ConfiguredIntegration) GetDataKinds() map[string]bool {
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
