package slackagent

import (
	"context"
	"fmt"
	"net/http"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/internal/integrations/slack"
	"golang.org/x/oauth2"
)

const integrationName = "slack_agent"

func MakeIntegration(
	appCfg rez.Config,
	db rez.Database,
	intgs rez.IntegrationService,
	incs rez.IncidentService,
	users rez.UserService,
	eventAnnos rez.EventAnnotationsService,
	messages rez.MessageService,
	provEvents rez.ProviderEventService,
) (*Integration, error) {
	cfg := appCfg.Integrations.Slack.Agent

	agentApp, appErr := makeApp(appCfg, db, messages, eventAnnos)
	if appErr != nil {
		return nil, fmt.Errorf("make incident app: %w", appErr)
	}

	svcParams := slackintegration.NewServiceParams{
		AppConfig:                   cfg,
		IntegrationName:             integrationName,
		MessageService:              messages,
		ProviderEventService:        provEvents,
		OAuthScopes:                 oAuthScopes,
		EventsApiHandler:            agentApp.handleEventsApiEvent,
		SlashCommandHandlers:        agentApp.slashCommandHandlers(),
		InteractionCallbackHandlers: agentApp.interactionCallbackHandlers(),
	}
	svc, svcErr := slackintegration.NewService(svcParams)
	if svcErr != nil {
		return nil, fmt.Errorf("failed to initialize slack service: %w", svcErr)
	}

	return &Integration{
		cfg:          cfg,
		service:      svc,
		users:        users,
		integrations: intgs,
		incidents:    incs,
		eventAnnos:   eventAnnos,
	}, nil
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

func (i *Integration) MaxInstalls() *int {
	return nil
}

func (i *Integration) IsAvailable() (bool, error) {
	return i.cfg.Enabled, nil
}

func (i *Integration) OAuthInstallRequired() bool {
	return true
}

func (i *Integration) WebhookHandler() http.Handler {
	return i.service.WebhookHandler()
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

func (i *Integration) Start(ctx context.Context) error {
	return i.service.Start(ctx)
}

func (i *Integration) Shutdown(ctx context.Context) error {
	return i.service.Shutdown(ctx)
}

var supportedDataKinds = []string{"chat", "users"}

func (i *Integration) SupportedDataKinds() []string {
	return supportedDataKinds
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

type InstalledIntegration struct {
	intg *ent.Integration

	db           rez.Database
	users        rez.UserService
	integrations rez.IntegrationService
	incidents    rez.IncidentService
	eventAnnos   rez.EventAnnotationsService
}

func (ii *InstalledIntegration) Integration() *ent.Integration {
	return ii.intg
}

func (ii *InstalledIntegration) config() (*slackintegration.InstallationConfig, error) {
	return slackintegration.DecodeInstallationConfig(ii.intg)
}

func (ii *InstalledIntegration) SanitizedInstallationConfig() map[string]any {
	return ii.intg.InstallationConfig
}

func (ii *InstalledIntegration) GetCapabilities() map[string]bool {
	return map[string]bool{
		"chat":  true,
		"users": true,
	}
}
