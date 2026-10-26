package slackincidents

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strings"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/internal/integrations/slack"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/integrations"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/slack-go/slack"
	"golang.org/x/oauth2"
)

const (
	integrationName              = "slack_incidents"
	incidentManagementCapability = "incident_management"
	// agentIntegrationName is the Slack agent, whose member sync maps Slack users to members.
	agentIntegrationName = "slack_agent"
)

var errAgentNotInstalled = errs.New(errs.CodeConflict, "Add the Rezible Slack agent to this workspace first.")

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

var _ jobs.WorkerProvider = (*Integration)(nil)

func (i *Integration) JobWorkers() []jobs.WorkerDefinition {
	return i.appSvc.App().JobWorkers()
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

// CheckInstallRequirements only allows Slack incident management when Rezible incident management is enabled,
// the Slack agent is installed in the same workspace, and no other installed integration already manages
// incidents. Before the workspace is known, any workspace with the agent is enough.
func (i *Integration) CheckInstallRequirements(state *integrations.IntegrationInstallState) error {
	if state.Preferences == nil || !state.Preferences.EnableIncidentManagement {
		return fmt.Errorf("incident management is not enabled")
	}
	hasAgent := slices.ContainsFunc(state.Installed, func(ii rez.InstalledIntegration) bool {
		intg := ii.Integration()
		return intg.Name == agentIntegrationName && (state.Target == nil || intg.ProviderInstallationRef == state.Target.ResourceRef)
	})
	if !hasAgent {
		return errAgentNotInstalled
	}
	for _, ii := range state.Installed {
		if slices.Contains(ii.Capabilities(), incidentManagementCapability) {
			return fmt.Errorf("an incident management integration is already installed")
		}
	}
	return nil
}

var channelIDPattern = regexp.MustCompile(`^[CG][A-Z0-9]{2,}$`)

func (i *Integration) ValidateUserSettings(m []byte) error {
	settings, decodeErr := integrations.DecodeUserSettings[UserSettings](m)
	if decodeErr != nil {
		return decodeErr
	}
	incidents := settings.Incidents
	if channelID := incidents.AnnouncementChannelID; channelID != "" && !channelIDPattern.MatchString(channelID) {
		return fmt.Errorf("%w: announcement channel %q is not a Slack channel ID", errs.ErrInvalidInput, channelID)
	}
	pattern := strings.TrimSpace(incidents.ChannelNamePattern)
	if pattern == "" {
		return fmt.Errorf("%w: a channel name pattern is required", errs.ErrInvalidInput)
	}
	if !strings.Contains(pattern, "{number}") && !strings.Contains(pattern, "{id}") {
		return fmt.Errorf("%w: the channel name pattern needs {number} or {id}, so each incident's channel has its own name", errs.ErrInvalidInput)
	}
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
	// AnnouncementChannelID is where new incidents are announced; empty uses the channel chosen at install.
	AnnouncementChannelID string `json:"announcement_channel_id"`
	ChannelNamePattern    string `json:"channel_name_pattern"`
	// AutoCreateVideoConference is accepted and ignored.
	AutoCreateVideoConference bool `json:"auto_create_video_conference"`
}

var defaultIncidentPreferences = UserSettingsIncidents{
	ChannelNamePattern: "inc-{number}",
}

// readIncidentSettings decodes the installation's settings, with the defaults for missing keys.
func readIncidentSettings(intg *ent.Integration) (*UserSettingsIncidents, error) {
	settings, decodeErr := integrations.DecodeUserSettings[UserSettings](intg.UserSettings)
	if decodeErr != nil {
		return nil, fmt.Errorf("decode slack incidents settings: %w", decodeErr)
	}
	incidents := settings.Incidents
	if strings.TrimSpace(incidents.ChannelNamePattern) == "" {
		incidents.ChannelNamePattern = defaultIncidentPreferences.ChannelNamePattern
	}
	return &incidents, nil
}

// announcementChannel is the channel set in the settings, or else the one chosen when the app was installed.
func (s *UserSettingsIncidents) announcementChannel(cfg *slackintegration.InstallationConfig) string {
	if s.AnnouncementChannelID != "" {
		return s.AnnouncementChannelID
	}
	return cfg.WebhookChannelId
}
