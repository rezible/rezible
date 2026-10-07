package alertmanager

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

const (
	ProviderName    = "alertmanager"
	integrationName = "alertmanager"
	sourceAlerts    = "alerts"
)

var capabilities = []string{"alerts"}

type Integration struct {
	webhookHandler http.Handler
}

func MakeIntegration(ts rez.TelemetryService, clock rez.Clock, events rez.ProviderEventPipelineService, installations rez.IntegrationInstallationLookup) (*Integration, error) {
	logger := ts.NewLogger(rez.NewLoggerOptions{Name: "alertmanager_webhooks"})
	i := &Integration{
		webhookHandler: newWebhookHandler(logger, clock, events, installations),
	}
	return i, nil
}

func (i *Integration) Name() string {
	return integrationName
}

func (i *Integration) DisplayName() string {
	return "Alertmanager"
}

func (i *Integration) Description() string {
	return "Receive Prometheus alerts from an Alertmanager webhook receiver"
}

func (i *Integration) Provider() string {
	return ProviderName
}

func (i *Integration) Capabilities() []string {
	return capabilities
}

func (i *Integration) MaxInstalls() *int {
	return nil
}

func (i *Integration) IsAvailable() (bool, error) {
	return true, nil
}

func (i *Integration) OAuthInstallRequired() bool {
	return false
}

func (i *Integration) InstallationLinks() []rez.IntegrationInstallationLink {
	return nil
}

func (i *Integration) WebhookHandler() http.Handler {
	return i.webhookHandler
}

// ValidateInstallationConfig ignores client input: installing asks for nothing.
func (i *Integration) ValidateInstallationConfig([]byte) (rez.IntegrationInstallationConfig, error) {
	return &InstallationConfig{installationRef: uuid.NewString()}, nil
}

func (i *Integration) ValidateUserSettings(settings map[string]any) error {
	_, parseErr := parseInstallationSettings(settings)
	return parseErr
}

func (i *Integration) GetInstalledIntegration(intg *ent.Integration) (rez.InstalledIntegration, error) {
	config := &InstallationConfig{installationRef: intg.ProviderInstallationRef}
	return &InstalledIntegration{intg: intg, config: config}, nil
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
	return capabilities
}

// InstallationConfig is empty. Alertmanager has no provider-side identity, so each installation gets a random
// reference to satisfy the unique provider installation reference. It is never encoded and nothing reads it
// after install; the installation's identity is its integration ID.
type InstallationConfig struct {
	installationRef string
}

func (c *InstallationConfig) Encode() ([]byte, error) {
	return json.Marshal(struct{}{})
}

func (c *InstallationConfig) InstallationTargetRef() rez.ProviderResourceRef {
	return rez.ProviderResourceRef{
		Provider:          ProviderName,
		ProviderNamespace: integrationName,
		ResourceRef:       c.installationRef,
	}
}
