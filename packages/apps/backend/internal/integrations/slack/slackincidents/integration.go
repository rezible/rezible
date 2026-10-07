package slackincidents

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/go-viper/mapstructure/v2"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/internal/integrations/slack"
	"github.com/rezible/rezible/pkg/integrations"
	"golang.org/x/oauth2"
)

const (
	integrationName              = "slack_incidents"
	incidentManagementCapability = "incident_management"
)

func MakeIntegration(appSvc *slackintegration.AppService[*App]) *Integration {
	return &Integration{appSvc: appSvc}
}

type Integration struct {
	appSvc *slackintegration.AppService[*App]
}

func (i *Integration) Name() string {
	return integrationName
}

func (i *Integration) Provider() string {
	return slackintegration.ProviderName
}

func (i *Integration) DisplayName() string {
	return "Slack Incident Management"
}

func (i *Integration) Description() string {
	return "Manage Rezible Incidents in Slack"
}

func (i *Integration) LifecycleService() rez.LifecycleService {
	if i.appSvc.HasLifecycle() {
		return i.appSvc
	}
	return nil
}

func (i *Integration) MessageHandlers() []rez.MessageEventHandler {
	return i.appSvc.MakeMessageHandlers()
}

func (i *Integration) Capabilities() []string {
	return []string{incidentManagementCapability}
}

func (i *Integration) IsAvailable() (bool, error) {
	return i.appSvc.Config().Enabled, nil
}

func (i *Integration) OAuthInstallRequired() bool {
	return true
}

func (i *Integration) InstallationLinks() []rez.IntegrationInstallationLink {
	return nil
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

// CheckInstallRequirements only allows Slack incident management when Rezible incident management is enabled
// and no other installed integration already manages incidents.
func (i *Integration) CheckInstallRequirements(state *integrations.IntegrationInstallState) error {
	if state.Preferences == nil || !state.Preferences.EnableIncidentManagement {
		return fmt.Errorf("incident management is not enabled")
	}
	for _, ii := range state.Installed {
		if slices.Contains(ii.Capabilities(), incidentManagementCapability) {
			return fmt.Errorf("an incident management integration is already installed")
		}
	}
	return nil
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
	return []string{incidentManagementCapability}
}

var _ integrations.ChatChannelQuerier = (*InstalledIntegration)(nil)

func (ii *InstalledIntegration) ListChatChannels(ctx context.Context, params integrations.ListChatChannelsParams) (*integrations.ChatChannelPage, error) {
	cw, clientErr := slackintegration.NewClientWrapper(ii.intg)
	if clientErr != nil {
		return nil, fmt.Errorf("create slack client: %w", clientErr)
	}
	return slackintegration.ListChatChannels(ctx, cw.Client(), params)
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
