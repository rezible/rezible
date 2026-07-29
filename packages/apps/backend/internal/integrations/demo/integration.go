package demoprovider

import (
	"encoding/json"
	"net/http"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

const (
	integrationName = "demo"
	providerName    = "demo"
)

type Integration struct {
	available      bool
	webhookHandler http.Handler
}

func MakeIntegration(cfg rez.Config, provEvents rez.ProviderEventPipelineService) (*Integration, error) {
	i := &Integration{
		available:      cfg.App.DebugMode,
		webhookHandler: newWebhookHandler(provEvents),
	}

	return i, nil
}

func (i *Integration) Name() string {
	return integrationName
}

func (i *Integration) DisplayName() string {
	return "Demo Data Provider"
}

func (i *Integration) Description() string {
	return "Provides demo data for testing purposes"
}

func (i *Integration) Provider() string {
	return providerName
}

func (i *Integration) Capabilities() []string {
	return []string{"demo_data", "event_sync", "code_changes"}
}

func (i *Integration) MaxInstalls() *int {
	return new(1)
}

func (i *Integration) IsAvailable() (bool, error) {
	return i.available, nil
}

func (i *Integration) OAuthInstallRequired() bool {
	return false
}

func (i *Integration) WebhookHandler() http.Handler {
	return i.webhookHandler
}

func (i *Integration) ValidateInstallationConfig(c []byte) (rez.IntegrationInstallationConfig, error) {
	var cfg InstallationConfig
	return &cfg, json.Unmarshal(c, &cfg)
}

func (i *Integration) ValidateUserSettings(settings map[string]any) error {
	return nil
}

func (i *Integration) GetInstalledIntegration(intg *ent.Integration) (rez.InstalledIntegration, error) {
	return &InstalledIntegration{intg: intg}, nil
}

type InstalledIntegration struct {
	intg   *ent.Integration
	config *InstallationConfig
}

func (ii *InstalledIntegration) Integration() *ent.Integration {
	return ii.intg
}

func (ii *InstalledIntegration) Config() rez.IntegrationInstallationConfig {
	return ii.config
}

func (ii *InstalledIntegration) Capabilities() []string {
	return []string{"demo_data", "event_sync", "code_changes"}
}

type InstallationConfig struct{}

func (i *InstallationConfig) Encode() ([]byte, error) {
	return json.Marshal(i)
}

func (i *InstallationConfig) ExternalRef() string {
	return "demo"
}
