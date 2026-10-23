package slackincidents

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/internal/integrations/slack"
	"github.com/rezible/rezible/pkg/integrations"
	"github.com/slack-go/slack"
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

func (i *Integration) ValidateUserSettings(m []byte) error {
	_, decodeErr := integrations.DecodeUserSettings[UserSettings](m)
	return decodeErr
}

func (i *Integration) GetInstalledIntegration(intg *ent.Integration) (rez.InstalledIntegration, error) {
	return i.makeInstalledIntegration(intg)
}

func (i *Integration) makeInstalledIntegration(intg *ent.Integration) (*InstalledIntegration, error) {
	cfg, cfgErr := slackintegration.GetValidatedConfig(intg.InstallationConfig)
	if cfgErr != nil {
		return nil, cfgErr
	}
	settings, decErr := integrations.DecodeUserSettings[UserSettings](intg.UserSettings)
	if decErr != nil {
		return nil, decErr
	}
	return &InstalledIntegration{intg: intg, config: cfg, settings: settings, clientOptions: i.appSvc.App().clientOptions()}, nil
}

type InstalledIntegration struct {
	intg          *ent.Integration
	config        *slackintegration.InstallationConfig
	settings      *UserSettings
	clientOptions []slack.Option
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
	cw, clientErr := slackintegration.NewClientWrapper(ii.intg, ii.clientOptions...)
	if clientErr != nil {
		return nil, fmt.Errorf("create slack client: %w", clientErr)
	}
	return slackintegration.ListChatChannels(ctx, cw.Client(), params)
}

type UserSettings struct {
	Incidents UserSettingsIncidents `json:"incidents"`
}

type UserSettingsIncidents struct {
	AnnouncementChannelID     string `json:"announcement_channel_id"`
	ChannelNamePattern        string `json:"channel_name_pattern"`
	AutoCreateVideoConference bool   `json:"auto_create_video_conference"`
	InviteMode                string `json:"invite_mode"`
}

var defaultIncidentPreferences = UserSettingsIncidents{
	AnnouncementChannelID:     "",
	ChannelNamePattern:        "incident-{slug}",
	AutoCreateVideoConference: false,
	InviteMode:                "assigned_users",
}
