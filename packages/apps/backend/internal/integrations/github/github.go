package github

import (
	"net/http"

	"github.com/go-viper/mapstructure/v2"
	"golang.org/x/oauth2"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

const (
	integrationName = "github"
	providerName    = "github"
)

var supportedCapabilities = []string{"repositories", "change_events"}

type Integration struct {
	cfg            rez.IntegrationsConfigGithub
	oauth2Config   *oauth2.Config
	webhookHandler http.Handler
}

func MakeIntegration(cfg rez.Config, provEvents rez.ProviderEventPipelineService) (*Integration, error) {
	i := &Integration{
		cfg:            cfg.Integrations.Github,
		webhookHandler: http.NotFoundHandler(),
	}

	i.oauth2Config = i.loadOAuthConfig()

	if i.cfg.Enabled {
		i.webhookHandler = newWebhookHandler(i.cfg.WebhookSecret, provEvents)
	}

	return i, nil
}

func (i *Integration) Name() string {
	return integrationName
}

func (i *Integration) DisplayName() string {
	return "Github"
}

func (i *Integration) Description() string {
	return "Watch for change events & extract repository information"
}

func (i *Integration) Provider() string {
	return providerName
}

func (i *Integration) MaxInstalls() *int {
	return nil
}

func (i *Integration) IsAvailable() (bool, error) {
	return i.cfg.Enabled, nil
}

func (i *Integration) WebhookHandler() http.Handler {
	return i.webhookHandler
}

func (i *Integration) SupportedCapabilities() []string {
	return supportedCapabilities
}

type installationConfig struct {
	Org            string
	InstallationID int64
}

func (i *Integration) ValidateInstallationConfig(m map[string]any) (externalRef string, validationErr error) {
	return "", nil
}

func (i *Integration) ValidateUserSettings(settings map[string]any) error {
	return nil
}

func (i *Integration) GetInstalledIntegration(intg *ent.Integration) rez.InstalledIntegration {
	return i.newInstalledIntegration(intg)
}

// InstalledIntegration wraps an *ent.Integration for a specific tenant installation.
type InstalledIntegration struct {
	intg *ent.Integration
}

func (i *Integration) newInstalledIntegration(intg *ent.Integration) *InstalledIntegration {
	return &InstalledIntegration{
		intg: intg,
	}
}

func (ii *InstalledIntegration) Integration() *ent.Integration {
	return ii.intg
}

func (ii *InstalledIntegration) ProviderName() string {
	return providerName
}

func (ii *InstalledIntegration) DisplayName() string {
	return "Github"
}

const (
	configOrg            = "org"
	configInstallationID = "installation_id"
)

func (ii *InstalledIntegration) config() (*installationConfig, error) {
	var cfg installationConfig
	if decErr := mapstructure.Decode(ii.intg.InstallationConfig, &cfg); decErr != nil {
		return nil, decErr
	}
	return &cfg, nil
}

func (ii *InstalledIntegration) SanitizedInstallationConfig() map[string]any {
	return ii.intg.InstallationConfig
}

func (ii *InstalledIntegration) GetCapabilities() map[string]bool {
	return map[string]bool{
		"repositories":  true,
		"change_events": true,
	}
}
