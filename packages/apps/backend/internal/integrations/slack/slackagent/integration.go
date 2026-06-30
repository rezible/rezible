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

func MakeIntegration(app *App, msgs rez.MessageService, events rez.ProviderEventPipelineService) (*Integration, error) {
	svc, svcErr := slackintegration.NewAppService(app, msgs, events)
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

func (i *Integration) ValidateConfig(m map[string]any) (externalRef string, validationErr error) {
	return "", nil
}

func (i *Integration) ValidateUserSettings(m map[string]any) error {
	return nil
}

func (i *Integration) GetInstalledIntegration(intg *ent.Integration) rez.InstalledIntegration {
	return i.makeInstalledIntegration(intg)
}

func (i *Integration) makeInstalledIntegration(intg *ent.Integration) *InstalledIntegration {
	return &InstalledIntegration{intg: intg}
}

type InstalledIntegration struct {
	intg *ent.Integration
}

func (ii *InstalledIntegration) Integration() *ent.Integration {
	return ii.intg
}

func (ii *InstalledIntegration) DisplayName() string {
	return "Slack Agent"
}

func (ii *InstalledIntegration) ProviderName() string {
	return slackintegration.ProviderName
}

func (ii *InstalledIntegration) config() (*slackintegration.InstallationConfig, error) {
	return slackintegration.DecodeInstallationConfig(ii.intg)
}

func (ii *InstalledIntegration) GetSanitizedConfig() map[string]any {
	return ii.intg.InstallationConfig
}

func (ii *InstalledIntegration) GetCapabilities() map[string]bool {
	return map[string]bool{
		"chat":  true,
		"users": true,
	}
}
