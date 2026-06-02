package slackincidents

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-viper/mapstructure/v2"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/internal/integrations/slack"
	"golang.org/x/oauth2"
)

const integrationName = "slack_incidents"

func MakeIntegration(
	appCfg rez.Config,
	db rez.Database,
	intgs rez.IntegrationService,
	incidents rez.IncidentService,
	users rez.UserService,
	eventAnnos rez.EventAnnotationsService,
	messages rez.MessageService,
	provEvents rez.ProviderEventService,
) (*Integration, error) {
	cfg := appCfg.Integrations.Slack.Incidents

	incApp, appErr := makeApp(appCfg, db, messages, incidents)
	if appErr != nil {
		return nil, fmt.Errorf("make incident app: %w", appErr)
	}

	svcParams := slackintegration.NewServiceParams{
		AppConfig:                   cfg,
		IntegrationName:             integrationName,
		MessageService:              messages,
		ProviderEventService:        provEvents,
		OAuthScopes:                 oAuthScopes,
		EventsApiHandler:            incApp.handleEventsApiEvent,
		SlashCommandHandlers:        incApp.slashCommandHandlers(),
		InteractionCallbackHandlers: incApp.interactionCallbackHandlers(),
	}
	svc, svcErr := slackintegration.NewService(svcParams)
	if svcErr != nil {
		return nil, fmt.Errorf("failed to initialize slack service: %w", svcErr)
	}

	intg := &Integration{
		cfg:          cfg,
		service:      svc,
		users:        users,
		integrations: intgs,
		incidents:    incidents,
		eventAnnos:   eventAnnos,
	}

	return intg, nil
}

type Integration struct {
	cfg rez.IntegrationsConfigSlackApp

	service *slackintegration.Service

	db           rez.Database
	users        rez.UserService
	integrations rez.IntegrationService
	incidents    rez.IncidentService
	eventAnnos   rez.EventAnnotationsService
}

func (i *Integration) Name() string {
	return integrationName
}

func (i *Integration) Provider() string {
	return slackintegration.ProviderName
}

func (i *Integration) IsAvailable() (bool, error) {
	return i.cfg.Enabled, nil
}

func (i *Integration) OAuthInstallRequired() bool {
	return true
}

func (i *Integration) OAuth2Config() *oauth2.Config {
	return i.service.OAuth2Config()
}

func (i *Integration) RetrieveInstallationTargetOptions(ctx context.Context, t *oauth2.Token) ([]rez.IntegrationInstallationTarget, error) {
	return i.service.RetrieveInstallationTargetOptions(ctx, t)
}

var oAuthScopes = []string{
	"app_mentions:read",
	"assistant:write",
	"channels:history",
	"channels:join",
	"channels:read",
	"chat:write",
	"chat:write.customize",
	"chat:write.public",
	"commands",
	"groups:history",
	"groups:read",
	"im:history",
	"im:read",
	"im:write",
	"im:write.topic",
	"incoming-webhook",
	"metadata.message:read",
	"mpim:history",
	"pins:read",
	"reactions:read",
	"usergroups:read",
	"users.profile:read",
	"users:read",
	"users:read.email",
	"channels:write.topic",
	"channels:manage",
	"channels:write.invites",
}

func (i *Integration) MaxInstalls() *int {
	return nil
}

func (i *Integration) WebhookHandler() http.Handler {
	return i.service.WebhookHandler()
}

func (i *Integration) Start(ctx context.Context) error {
	return i.service.Start(ctx)
}

func (i *Integration) Shutdown(ctx context.Context) error {
	return i.service.Shutdown(ctx)
}

func (i *Integration) SupportedDataKinds() []string {
	return nil
}

func (i *Integration) GetConfiguredIntegration(intg *ent.Integration) rez.InstalledIntegration {
	return i.makeInstalledIntegration(intg)
}

func (i *Integration) ValidateInstallationConfig(m map[string]any) (externalRef string, validationErr error) {
	return "", nil
}

func (i *Integration) ValidateUserSettings(m map[string]any) error {
	return nil
}

func (i *Integration) GetInstalledIntegration(intg *ent.Integration) rez.InstalledIntegration {
	return i.makeInstalledIntegration(intg)
}

type InstalledIntegration struct {
	intg *ent.Integration

	db           rez.Database
	users        rez.UserService
	integrations rez.IntegrationService
	incidents    rez.IncidentService
	eventAnnos   rez.EventAnnotationsService
}

func (i *Integration) makeInstalledIntegration(intg *ent.Integration) *InstalledIntegration {
	return &InstalledIntegration{
		intg:         intg,
		db:           i.db,
		users:        i.users,
		integrations: i.integrations,
		incidents:    i.incidents,
		eventAnnos:   i.eventAnnos,
	}
}

func (ii *InstalledIntegration) Integration() *ent.Integration {
	return ii.intg
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
