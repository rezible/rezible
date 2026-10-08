package webhook

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

const (
	ProviderName    = "webhook"
	integrationName = "webhook"
)

var capabilities = []string{"deployments"}

type Integration struct {
	webhookHandler http.Handler
}

func MakeIntegration(ts rez.TelemetryService, clock rez.Clock, events rez.ProviderEventPipelineService, installations rez.IntegrationInstallationLookup) (*Integration, error) {
	logger := ts.NewLogger(rez.NewLoggerOptions{Name: "webhook_deliveries"})
	i := &Integration{
		webhookHandler: newWebhookHandler(logger, clock, events, installations),
	}
	return i, nil
}

func (i *Integration) Name() string {
	return integrationName
}

func (i *Integration) DisplayName() string {
	return "Webhook"
}

func (i *Integration) Description() string {
	return "Receive deployment reports from your deploy pipelines"
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

// ValidateInstallationConfig accepts only a known preset. Installation config is written only at install, so
// an installation's preset never changes; a different preset is a new installation.
func (i *Integration) ValidateInstallationConfig(raw []byte) (rez.IntegrationInstallationConfig, error) {
	config, decodeErr := decodeInstallationConfig(raw)
	if decodeErr != nil {
		return nil, decodeErr
	}
	config.installationRef = uuid.NewString()
	return config, nil
}

// ValidateUserSettings rejects any setting: the webhook integration has none.
func (i *Integration) ValidateUserSettings(settings map[string]any) error {
	if len(settings) > 0 {
		return fmt.Errorf("%w: the webhook integration has no settings", rez.ErrInvalidInput)
	}
	return nil
}

func (i *Integration) GetInstalledIntegration(intg *ent.Integration) (rez.InstalledIntegration, error) {
	config, decodeErr := decodeInstallationConfig(intg.InstallationConfig)
	if decodeErr != nil {
		return nil, decodeErr
	}
	config.installationRef = intg.ProviderInstallationRef
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

// Metadata tells the settings page which preset the installation accepts.
func (ii *InstalledIntegration) Metadata() map[string]string {
	return map[string]string{"preset": ii.config.Preset}
}

// InstallationConfig names the installation's preset. As for Alertmanager, there is no provider-side identity,
// so each installation gets a random reference to satisfy the unique provider installation reference. It is
// never encoded; the installation's identity is its integration ID.
type InstallationConfig struct {
	Preset          string `json:"preset"`
	installationRef string
}

func decodeInstallationConfig(raw []byte) (*InstallationConfig, error) {
	var config InstallationConfig
	if len(raw) > 0 {
		if decodeErr := json.Unmarshal(raw, &config); decodeErr != nil {
			return nil, fmt.Errorf("%w: invalid installation config: %w", rez.ErrInvalidInput, decodeErr)
		}
	}
	if _, known := presets[config.Preset]; !known {
		return nil, fmt.Errorf("%w: unknown preset %q", rez.ErrInvalidInput, config.Preset)
	}
	return &config, nil
}

func (c *InstallationConfig) Encode() ([]byte, error) {
	return json.Marshal(c)
}

func (c *InstallationConfig) InstallationTargetRef() rez.ProviderResourceRef {
	return rez.ProviderResourceRef{
		Provider:          ProviderName,
		ProviderNamespace: integrationName,
		ResourceRef:       c.installationRef,
	}
}
