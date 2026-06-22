package slackincidents

import (
	"context"
	"net/http"

	"github.com/go-viper/mapstructure/v2"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/internal/integrations/slack"
	"golang.org/x/oauth2"
)

const integrationName = "slack_incidents"

func MakeIntegration(svc *slackintegration.AppService[*App]) (*Integration, error) {
	return &Integration{appSvc: svc}, nil
}

type Integration struct {
	appSvc *slackintegration.AppService[*App]
}

func (i *Integration) Name() string {
	return integrationName
}

func (i *Integration) DisplayName() string {
	return "Slack Incident Management"
}

func (i *Integration) Description() string {
	return "Manage Rezible Incidents in Slack"
}

func (i *Integration) Provider() string {
	return slackintegration.ProviderName
}

func (i *Integration) IsAvailable() (bool, error) {
	return i.appSvc.App().AppConfig().Enabled, nil
}

func (i *Integration) OAuthInstallRequired() bool {
	return true
}

func (i *Integration) OAuth2Config() *oauth2.Config {
	return i.appSvc.OAuth2Config()
}

func (i *Integration) RetrieveInstallationTargetOptions(ctx context.Context, t *oauth2.Token) ([]rez.IntegrationInstallationTarget, error) {
	return i.appSvc.RetrieveInstallationTargetOptions(ctx, t)
}

func (i *Integration) MaxInstalls() *int {
	return nil
}

func (i *Integration) WebhookHandler() http.Handler {
	return i.appSvc.WebhookHandler()
}

func (i *Integration) SupportedCapabilities() []string {
	return nil
}

func (i *Integration) ValidateInstallationConfig(m map[string]any) (externalRef string, validationErr error) {
	return "", nil
}

func (i *Integration) ValidateUserSettings(m map[string]any) error {
	return nil
}

func (i *Integration) GetInstalledIntegration(intg *ent.Integration) rez.InstalledIntegration {
	return &InstalledIntegration{intg: intg}
}

type InstalledIntegration struct {
	intg *ent.Integration
}

func (ii *InstalledIntegration) Integration() *ent.Integration {
	return ii.intg
}

func (ii *InstalledIntegration) DisplayName() string {
	return "Slack Incident Management"
}

func (ii *InstalledIntegration) ProviderName() string {
	return slackintegration.ProviderName
}

func (ii *InstalledIntegration) config() (*slackintegration.InstallationConfig, error) {
	return slackintegration.DecodeInstallationConfig(ii.intg)
}

type UserSettings struct {
	Incidents UserSettingsIncidents `mapstructure:"incidents"`
}

type UserSettingsIncidents struct {
	AnnouncementChannelID     string `mapstructure:"announcement_channel_id"`
	ChannelNamePattern        string `mapstructure:"channel_name_pattern"`
	AutoCreateVideoConference bool   `mapstructure:"create_video_conference"`
	InviteMode                string `mapstructure:"invite_mode"`
}

func (ii *InstalledIntegration) userPreferences() (*UserSettings, error) {
	var settings UserSettings
	if decErr := mapstructure.Decode(ii.intg.UserSettings, &settings); decErr != nil {
		return nil, decErr
	}
	return &settings, nil
}

func (ii *InstalledIntegration) SanitizedInstallationConfig() map[string]any {
	return ii.intg.InstallationConfig
}

var defaultIncidentPreferences = UserSettingsIncidents{
	AnnouncementChannelID:     "",
	ChannelNamePattern:        "incident-{slug}",
	AutoCreateVideoConference: false,
	InviteMode:                "assigned_users",
}

func (ii *InstalledIntegration) GetCapabilities() map[string]bool {
	return map[string]bool{
		"chat":  true,
		"users": true,
	}
}
