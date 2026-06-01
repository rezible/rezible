package slackagent

import (
	"context"
	"fmt"
	"net/http"

	slackgo "github.com/slack-go/slack"

	"github.com/google/uuid"
	"github.com/stretchr/objx"
	"golang.org/x/oauth2"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/execution"
	"github.com/rezible/rezible/internal/integrations/slack"
)

const integrationName = "slackagent"

func MakeIntegration(
	cfg rez.Config,
	intgs rez.IntegrationService,
	incs rez.IncidentService,
	users rez.UserService,
	eventAnnos rez.EventAnnotationsService,
	messages rez.MessageService,
	provEvents rez.ProviderEventService,
) (*Integration, error) {
	svc, svcErr := slack.NewService(integrationName, messages)
	if svcErr != nil {
		return nil, fmt.Errorf("failed to initialize slack service: %w", svcErr)
	}

	slackCfg := cfg.Integrations.Slack
	intg := &Integration{
		service:      svc,
		cfg:          slackCfg,
		users:        users,
		integrations: intgs,
		incidents:    incs,
		eventAnnos:   eventAnnos,
		oauth2Config: makeOAuth2Config(slackCfg),
	}

	if slackCfg.EnableSocketMode {
		smClient := slackgo.New(slackCfg.BotToken, slackgo.OptionAppLevelToken(slackCfg.AppToken))
		if smErr := svc.SetupSocketMode(smClient); smErr != nil {
			return nil, fmt.Errorf("failed to setup socket mode: %w", smErr)
		}
	} else {
		if whErr := svc.SetupWebhooks(slackCfg.WebhookSigningSecret); whErr != nil {
			return nil, fmt.Errorf("failed to setup webhooks: %w", whErr)
		}
	}

	return intg, nil
}

type Integration struct {
	cfg          rez.IntegrationsConfigSlack
	oauth2Config *oauth2.Config

	service *slack.Service

	db           rez.Database
	users        rez.UserService
	integrations rez.IntegrationService
	incidents    rez.IncidentService
	eventAnnos   rez.EventAnnotationsService
}

func (i *Integration) Name() string {
	return integrationName
}

func (i *Integration) IsAvailable() (bool, error) {
	return i.cfg.Enabled, nil
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

var supportedDataKinds = []string{"chat", "users"}

func (i *Integration) SupportedDataKinds() []string {
	return supportedDataKinds
}

func (i *Integration) GetConfiguredIntegration(intg *ent.Integration) rez.ConfiguredIntegration {
	return i.makeConfiguredIntegration(intg)
}

func (i *Integration) ValidateUserConfig(cfg map[string]any) error {
	return nil
}

func (i *Integration) ValidateUserPreferences(prefs map[string]any) error {
	return nil
}

type ConfiguredIntegration struct {
	intg *ent.Integration

	db           rez.Database
	users        rez.UserService
	integrations rez.IntegrationService
	incidents    rez.IncidentService
	eventAnnos   rez.EventAnnotationsService
}

func (ci *ConfiguredIntegration) Integration() *ent.Integration {
	return ci.intg
}

func (i *Integration) makeConfiguredIntegration(intg *ent.Integration) *ConfiguredIntegration {
	return &ConfiguredIntegration{
		intg:         intg,
		db:           i.db,
		users:        i.users,
		integrations: i.integrations,
		incidents:    i.incidents,
		eventAnnos:   i.eventAnnos,
	}
}

func (ci *ConfiguredIntegration) tenantContext(ctx context.Context) context.Context {
	return execution.NewTenantContext(ctx, ci.intg.TenantID)
}

func (ci *ConfiguredIntegration) config() objx.Map {
	return objx.New(ci.intg.Config)
}

const (
	configAccessToken      = "access_token"
	configBotUserID        = "bot_user_id"
	configWebhookChannelId = "webhook_channel_id"
	configTeam             = "team"
	configEnterprise       = "enterprise"
)

func (ci *ConfiguredIntegration) accessToken() string {
	return ci.config().Get(configAccessToken).String()
}

func (ci *ConfiguredIntegration) teamId() string {
	return ci.config().Get(configTeam + ".id").String()
}

func (ci *ConfiguredIntegration) botUserID() string {
	return ci.config().Get(configBotUserID).String()
}

func (ci *ConfiguredIntegration) userPreferences() objx.Map {
	return objx.New(ci.intg.UserPreferences)
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
	return ci.intg.UserPreferences
}

func (ci *ConfiguredIntegration) GetAvailableDataKinds() map[string]bool {
	return map[string]bool{
		"chat":  true,
		"users": true,
	}
}
