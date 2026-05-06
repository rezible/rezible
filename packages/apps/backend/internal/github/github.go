package github

import (
	"context"
	"fmt"
	"net/http"

	"github.com/stretchr/objx"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/execution"
)

const integrationName = "github"

var supportedDataKinds = []string{"repositories", "change_events"}

type integration struct {
	cfg             Config
	services        *rez.Services
	webhookHandlers map[string]http.Handler
}

func SetupIntegration(ctx context.Context, svcs *rez.Services) (rez.IntegrationPackage, error) {
	i := &integration{
		services:        svcs,
		webhookHandlers: make(map[string]http.Handler),
	}

	if cfgErr := rez.Config.Unmarshal("github", &i.cfg); cfgErr != nil {
		return nil, fmt.Errorf("config error: %w", cfgErr)
	}

	if i.cfg.Enabled {
		svcs.ProviderEvents.RegisterEventProcessor(integrationName, "push", &pushEventProcessor{services: svcs})
		svcs.ProviderEvents.RegisterEventProcessor(integrationName, "pull_request", &pullRequestEventProcessor{services: svcs})
		i.webhookHandlers["/"] = newWebhookHandler(i.cfg.WebhookSecret, svcs)
	}

	return i, nil
}

func (i *integration) Name() string {
	return integrationName
}

func (i *integration) IsAvailable() (bool, error) {
	if !i.cfg.Enabled {
		return false, nil
	}
	return true, i.cfg.validate()
}

func (i *integration) SupportedDataKinds() []string {
	return supportedDataKinds
}

func (i *integration) OAuthConfigRequired() bool {
	return true
}

func (i *integration) ValidateConfig(cfg map[string]any) error {
	// Extract app config fields from the map
	app, hasApp := cfg["app"].(map[string]any)
	if !hasApp {
		return fmt.Errorf("missing app configuration")
	}

	var c Config
	if appID, ok := app["app_id"].(float64); ok {
		c.App.AppID = int64(appID)
	}
	if clientID, ok := app["client_id"].(string); ok {
		c.App.ClientID = clientID
	}
	if clientSecret, ok := app["client_secret"].(string); ok {
		c.App.ClientSecret = clientSecret
	}
	if privateKeyPEM, ok := app["private_key_pem"].(string); ok {
		c.App.PrivateKeyPEM = privateKeyPEM
	}

	return c.validate()
}

func (i *integration) ValidateUserPreferences(_ map[string]any) error {
	return nil
}

func (i *integration) GetConfiguredIntegration(intg *ent.Integration) rez.ConfiguredIntegration {
	return newConfiguredIntegration(i.services, intg)
}

func (i *integration) WebhookHandlers() map[string]http.Handler {
	return i.webhookHandlers
}

func (i *integration) MakeSystemComponentsDataProvider(ctx context.Context, intg *ent.Integration) (rez.SystemComponentsDataProvider, error) {
	ci := newConfiguredIntegration(i.services, intg)
	client, err := newClient(ctx, ci)
	if err != nil {
		return nil, fmt.Errorf("create github client: %w", err)
	}
	return &systemComponentsProvider{client: client, services: i.services}, nil
}

// ConfiguredIntegration wraps an *ent.Integration for a specific tenant installation.
type ConfiguredIntegration struct {
	svcs *rez.Services
	intg *ent.Integration
}

func newConfiguredIntegration(svcs *rez.Services, intg *ent.Integration) *ConfiguredIntegration {
	return &ConfiguredIntegration{svcs: svcs, intg: intg}
}

func (ci *ConfiguredIntegration) tenantContext(ctx context.Context) context.Context {
	return execution.AnonymousTenantContext(ctx, ci.intg.TenantID)
}

func (ci *ConfiguredIntegration) config() objx.Map {
	return objx.New(ci.intg.Config)
}

const (
	configOrg            = "org"
	configInstallationID = "installation_id"
)

func (ci *ConfiguredIntegration) orgName() string {
	return ci.config().Get(configOrg).String()
}

func (ci *ConfiguredIntegration) installationID() int64 {
	v := ci.config().Get(configInstallationID)
	if v.IsFloat64() {
		return int64(v.Float64())
	}
	return 0
}

type appConfig struct {
	AppID         int64
	PrivateKeyPEM string
}

func (ci *ConfiguredIntegration) appConfig() appConfig {
	// App credentials come from the global config, not per-tenant config
	var cfg Config
	_ = rez.Config.Unmarshal("github", &cfg)
	return appConfig{
		AppID:         cfg.App.AppID,
		PrivateKeyPEM: cfg.App.PrivateKeyPEM,
	}
}

func (ci *ConfiguredIntegration) Name() string {
	return integrationName
}

func (ci *ConfiguredIntegration) GetSanitizedConfig() map[string]any {
	return ci.config().Exclude([]string{})
}

func (ci *ConfiguredIntegration) GetUserPreferences() map[string]any {
	return ci.intg.UserPreferences
}

func (ci *ConfiguredIntegration) GetDataKinds() map[string]bool {
	return map[string]bool{
		"repositories":  true,
		"change_events": true,
	}
}
