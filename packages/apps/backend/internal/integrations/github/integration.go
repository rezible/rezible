package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

const (
	ProviderName       = "github"
	integrationName    = "github"
	sourcePushEvent    = "push"
	sourcePullEvent    = "pull_request"
	sourceRepositories = "repositories"
)

type Integration struct {
	cfg            rez.IntegrationsConfigGithub
	webhookHandler http.Handler
}

func MakeIntegration(cfg rez.Config, events rez.ProviderEventPipelineService) (*Integration, error) {
	i := &Integration{
		cfg:            cfg.Integrations.Github,
		webhookHandler: http.NotFoundHandler(),
	}

	if i.cfg.Enabled {
		i.webhookHandler = newWebhookHandler(i.cfg.WebhookSecret, events)
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
	return ProviderName
}

func (i *Integration) Capabilities() []string {
	return []string{"event_sync", "code_changes", "repositories"}
}

func (i *Integration) MaxInstalls() *int {
	return nil
}

func (i *Integration) InstallationLinks() []rez.IntegrationInstallationLink {
	if i.cfg.App.Slug == "" {
		return nil
	}
	installUrl := url.URL{
		Scheme: "https",
		Host:   "github.com",
		Path:   path.Join("/apps", i.cfg.App.Slug, "installations/new"),
	}
	installLink := rez.IntegrationInstallationLink{
		Kind:  "install_app",
		Label: "Install the Rezible GitHub App",
		URL:   installUrl.String(),
	}
	return []rez.IntegrationInstallationLink{installLink}
}

func (i *Integration) IsAvailable() (bool, error) {
	return i.cfg.Enabled, nil
}

func (i *Integration) WebhookHandler() http.Handler {
	return i.webhookHandler
}

func (i *Integration) ValidateInstallationConfig(raw []byte) (rez.IntegrationInstallationConfig, error) {
	return i.decodeValidateInstallationConfig(raw)
}

func (i *Integration) decodeValidateInstallationConfig(raw []byte) (*InstallationConfig, error) {
	var cfg InstallationConfig
	if encErr := json.Unmarshal(raw, &cfg); encErr != nil {
		return nil, encErr
	}
	// TODO: validate fields
	return &cfg, nil
}

func (i *Integration) ValidateUserSettings(settings map[string]any) error {
	return nil
}

func (i *Integration) GetInstalledIntegration(intg *ent.Integration) (rez.InstalledIntegration, error) {
	return i.newInstalledIntegration(intg)
}

// InstalledIntegration wraps an *ent.Integration for a specific tenant installation.
type InstalledIntegration struct {
	intg   *ent.Integration
	config *InstallationConfig
}

func (i *Integration) newInstalledIntegration(intg *ent.Integration) (*InstalledIntegration, error) {
	cfg, cfgErr := i.decodeValidateInstallationConfig(intg.InstallationConfig)
	if cfgErr != nil {
		return nil, fmt.Errorf("validate config: %w", cfgErr)
	}
	return &InstalledIntegration{intg: intg, config: cfg}, nil
}

func (ii *InstalledIntegration) Integration() *ent.Integration {
	return ii.intg
}

func (ii *InstalledIntegration) Config() rez.IntegrationInstallationConfig {
	return ii.config
}

func (ii *InstalledIntegration) Capabilities() []string {
	return []string{"event_sync", "code_changes", "repositories"}
}
