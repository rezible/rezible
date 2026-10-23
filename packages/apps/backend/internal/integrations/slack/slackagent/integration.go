package slackagent

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/go-viper/mapstructure/v2"
	"github.com/slack-go/slack"
	"golang.org/x/oauth2"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/internal/integrations/slack"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/integrations"
	"github.com/rezible/rezible/pkg/jobs"
)

const integrationName = "slack_agent"

func MakeIntegration(svc *slackintegration.AppService[*App]) *Integration {
	return &Integration{appSvc: svc}
}

type Integration struct {
	appSvc *slackintegration.AppService[*App]
}

func (i *Integration) Name() string {
	return integrationName
}

func (i *Integration) DisplayName() string {
	return "Slack Agent"
}

func (i *Integration) Description() string {
	return "Rezible Slack Agent"
}

func (i *Integration) Provider() string {
	return slackintegration.ProviderName
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
	return []string{"chat_context"}
}

func (i *Integration) MaxInstalls() *int {
	return nil
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

func (i *Integration) WebhookHandler() http.Handler {
	return i.appSvc.WebhookHandler()
}

func (i *Integration) OAuth2Config() *oauth2.Config {
	return i.appSvc.OAuth2Config()
}

func (i *Integration) RetrieveInstallationTargetOptions(ctx context.Context, t *oauth2.Token) ([]rez.IntegrationInstallationTarget, error) {
	return i.appSvc.RetrieveInstallationTargetOptions(ctx, t)
}

func (i *Integration) ValidateInstallationConfig(cfg []byte) (rez.IntegrationInstallationConfig, error) {
	return i.appSvc.ValidateInstallationConfig(cfg)
}

func (i *Integration) ValidateUserSettings(m map[string]any) error {
	settings, decodeErr := decodeUserSettings(m)
	if decodeErr != nil {
		return fmt.Errorf("%w: %w", errs.ErrInvalidInput, decodeErr)
	}
	if channelID := settings.SituationNoticeChannelID; channelID != "" && !channelIDPattern.MatchString(channelID) {
		return fmt.Errorf("%w: situation notice channel %q is not a Slack channel ID", errs.ErrInvalidInput, channelID)
	}
	return nil
}

var channelIDPattern = regexp.MustCompile(`^[CG][A-Z0-9]{2,}$`)

// UserSettings are an installation's settings, chosen by an admin.
type UserSettings struct {
	// SituationNoticeChannelID is where raised situations are announced; empty announces nothing.
	SituationNoticeChannelID string `mapstructure:"situation_notice_channel_id"`
}

func decodeUserSettings(m map[string]any) (*UserSettings, error) {
	var settings UserSettings
	if decodeErr := mapstructure.Decode(m, &settings); decodeErr != nil {
		return nil, fmt.Errorf("decode settings: %w", decodeErr)
	}
	return &settings, nil
}

func (i *Integration) GetInstalledIntegration(intg *ent.Integration) (rez.InstalledIntegration, error) {
	return i.makeInstalledIntegration(intg)
}

func (i *Integration) makeInstalledIntegration(intg *ent.Integration) (*InstalledIntegration, error) {
	cfg, cfgErr := slackintegration.GetValidatedConfig(intg.InstallationConfig)
	if cfgErr != nil {
		return nil, cfgErr
	}
	return &InstalledIntegration{intg: intg, config: cfg, clientOptions: i.appSvc.App().clientOptions()}, nil
}

type InstalledIntegration struct {
	intg          *ent.Integration
	config        *slackintegration.InstallationConfig
	clientOptions []slack.Option
}

func (ii *InstalledIntegration) Integration() *ent.Integration {
	return ii.intg
}

func (ii *InstalledIntegration) ProviderName() string {
	return slackintegration.ProviderName
}

func (ii *InstalledIntegration) Config() rez.IntegrationInstallationConfig {
	return slackintegration.MakeInstallationConfig(integrationName, ii.config)
}

func (ii *InstalledIntegration) Capabilities() []string {
	return []string{"chat_context"}
}

var _ integrations.ChatChannelQuerier = (*InstalledIntegration)(nil)

func (ii *InstalledIntegration) ListChatChannels(ctx context.Context, params integrations.ListChatChannelsParams) (*integrations.ChatChannelPage, error) {
	cw, clientErr := ii.clientWrapper()
	if clientErr != nil {
		return nil, clientErr
	}
	return slackintegration.ListChatChannels(ctx, cw.Client(), params)
}
