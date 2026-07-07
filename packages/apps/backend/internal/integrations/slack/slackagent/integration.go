package slackagent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/internal/integrations/slack"
	"golang.org/x/oauth2"
)

const integrationName = "slack_agent"

func MakeIntegration(app *App, msgs rez.MessageService, intgs rez.IntegrationService, users rez.UserService, events rez.ProviderEventPipelineService) (*Integration, error) {
	svc, svcErr := slackintegration.NewAppService(app, msgs, intgs, users, events)
	if svcErr != nil {
		return nil, fmt.Errorf("making slackintegration: %w", svcErr)
	}
	return &Integration{appSvc: svc}, nil
}

type Integration struct {
	appSvc *slackintegration.AppService[*App]
}

func (i *Integration) Start(ctx context.Context) error {
	return i.appSvc.Start(ctx)
}

func (i *Integration) Shutdown(ctx context.Context) error {
	return i.appSvc.Shutdown(ctx)
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

func (i *Integration) MaxInstalls() *int {
	return nil
}

func (i *Integration) IsAvailable() (bool, error) {
	return i.appSvc.App().Config().Enabled, nil
}

func (i *Integration) OAuthInstallRequired() bool {
	return true
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

func (i *Integration) ValidateInstallationConfig(cfg json.RawMessage) (rez.IntegrationInstallationConfig, error) {
	return slackintegration.GetValidatedConfig(cfg)
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
	return &InstalledIntegration{intg: intg, config: cfg}, nil
}

type InstalledIntegration struct {
	intg   *ent.Integration
	config *slackintegration.InstallationConfig
}

func (ii *InstalledIntegration) Integration() *ent.Integration {
	return ii.intg
}

func (ii *InstalledIntegration) ProviderName() string {
	return slackintegration.ProviderName
}

func (ii *InstalledIntegration) Config() rez.IntegrationInstallationConfig {
	return ii.config
}

func (ii *InstalledIntegration) GetCapabilities() map[string]bool {
	return map[string]bool{
		"chat":  true,
		"users": true,
	}
}
