package slackagent

import (
	"context"
	"net/http"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/internal/integrations/slack"
	"github.com/rezible/rezible/pkg/jobs"
	"golang.org/x/oauth2"
)

const integrationName = "slack_agent"

func MakeIntegration(svc *slackintegration.AppService[*App]) *Integration {
	return &Integration{appSvc: svc}
}

type Integration struct {
	appSvc *slackintegration.AppService[*App]
}

func (i *Integration) RegisterJobs(registry *jobs.Registry) error {
	return i.appSvc.App().RegisterJobs(registry)
}

func (i *Integration) Lifecycle() *rez.ServiceLifecycle {
	return i.appSvc.Lifecycle()
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

func (i *Integration) Capabilities() []string {
	return []string{"chat_context"}
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

func (i *Integration) ValidateInstallationConfig(cfg []byte) (rez.IntegrationInstallationConfig, error) {
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

func (ii *InstalledIntegration) Capabilities() []string {
	return []string{"chat_context"}
}
