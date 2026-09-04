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

func MakeIntegration(appSvc *slackintegration.AppService[*App]) *Integration {
	return &Integration{appSvc: appSvc}
}

type Integration struct {
	appSvc *slackintegration.AppService[*App]
}

func (i *Integration) Lifecycle() *rez.ServiceLifecycle {
	return i.appSvc.Lifecycle()
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

func (i *Integration) Capabilities() []string {
	return []string{"incident_management"}
}

func (i *Integration) IsAvailable() (bool, error) {
	return i.appSvc.App().Config().Enabled, nil
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

func (i *Integration) ValidateInstallationConfig(m []byte) (rez.IntegrationInstallationConfig, error) {
	return i.appSvc.ValidateInstallationConfig(m)
}

func (i *Integration) ValidateUserSettings(m map[string]any) error {
	return nil
}

func (i *Integration) GetInstalledIntegration(intg *ent.Integration) (rez.InstalledIntegration, error) {
	return i.makeInstalledIntegration(intg)
}

func (i *Integration) makeInstalledIntegration(intg *ent.Integration) (*InstalledIntegration, error) {
	cfg, cfgErr := slackintegration.GetValidatedConfig(intg.InstallationConfig)
	if cfgErr != nil {
		return nil, cfgErr
	}
	ii := &InstalledIntegration{intg: intg, config: cfg}
	if decErr := mapstructure.Decode(intg.UserSettings, &ii.settings); decErr != nil {
		return nil, decErr
	}
	return ii, nil
}

type InstalledIntegration struct {
	intg     *ent.Integration
	config   *slackintegration.InstallationConfig
	settings *UserSettings
}

func (ii *InstalledIntegration) Integration() *ent.Integration {
	return ii.intg
}

func (ii *InstalledIntegration) Config() rez.IntegrationInstallationConfig {
	return slackintegration.MakeInstallationConfig(integrationName, ii.config)
}

func (ii *InstalledIntegration) Capabilities() []string {
	return []string{"incident_management"}
}

type UserSettings struct {
	Incidents UserSettingsIncidents
}

type UserSettingsIncidents struct {
	AnnouncementChannelID     string
	ChannelNamePattern        string
	AutoCreateVideoConference bool
	InviteMode                string
}

var defaultIncidentPreferences = UserSettingsIncidents{
	AnnouncementChannelID:     "",
	ChannelNamePattern:        "incident-{slug}",
	AutoCreateVideoConference: false,
	InviteMode:                "assigned_users",
}
