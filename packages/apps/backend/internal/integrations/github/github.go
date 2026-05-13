package github

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	gh "github.com/google/go-github/v84/github"
	"github.com/google/uuid"
	"github.com/stretchr/objx"
	"golang.org/x/oauth2"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/execution"
)

const integrationName = "github"

var supportedDataKinds = []string{"repositories", "change_events"}

type integration struct {
	cfg                   Config
	services              *rez.Services
	oauth2Config          *oauth2.Config
	listUserInstallations func(context.Context, string) ([]*gh.Installation, error)
	webhookHandlers       map[string]http.Handler
}

func SetupIntegration(ctx context.Context, svcs *rez.Services) (rez.IntegrationPackage, error) {
	i := &integration{
		services:              svcs,
		listUserInstallations: listUserInstallations,
		webhookHandlers:       make(map[string]http.Handler),
	}

	if cfgErr := rez.Config.Unmarshal("github", &i.cfg); cfgErr != nil {
		return nil, fmt.Errorf("config error: %w", cfgErr)
	}
	i.oauth2Config = i.loadOAuthConfig()

	if i.cfg.Enabled {
		svcs.ProviderEvents.RegisterEventProcessors(integrationName, map[string]rez.ProviderEventProcessor{
			"push":         &pushEventProcessor{services: svcs},
			"pull_request": &pullRequestEventProcessor{services: svcs},
			"repositories": &repositoryObservedProcessor{services: svcs},
		})
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

func (i *integration) OAuth2Config() *oauth2.Config {
	return i.oauth2Config
}

func (i *integration) loadOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     i.cfg.App.ClientID,
		ClientSecret: i.cfg.App.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://github.com/login/oauth/authorize",
			TokenURL:  "https://github.com/login/oauth/access_token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}
}

func listUserInstallations(ctx context.Context, accessToken string) ([]*gh.Installation, error) {
	client := gh.NewClient(nil).WithAuthToken(accessToken)
	var all []*gh.Installation
	opts := &gh.ListOptions{PerPage: 100}
	for page := 1; ; page++ {
		opts.Page = page
		installations, resp, err := client.Apps.ListUserInstallations(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("list user installations page %d: %w", page, err)
		}
		all = append(all, installations...)
		if resp.NextPage == 0 {
			break
		}
	}
	return all, nil
}

func (i *integration) ExtractIntegrationOptionsFromToken(t *oauth2.Token) ([]rez.ExternalIntegrationOption, error) {
	if t == nil || t.AccessToken == "" {
		return nil, fmt.Errorf("missing access token")
	}
	if i.listUserInstallations == nil {
		i.listUserInstallations = listUserInstallations
	}

	// The shared OAuth integration contract does not carry a request context into this hook.
	installations, err := i.listUserInstallations(context.TODO(), t.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("list github installations: %w", err)
	}

	matches := make([]*gh.Installation, 0, len(installations))
	for _, installation := range installations {
		if installation == nil || installation.GetID() == 0 || installation.GetAccount().GetLogin() == "" {
			continue
		}
		if i.cfg.App.AppID != 0 && installation.GetAppID() != 0 && installation.GetAppID() != i.cfg.App.AppID {
			continue
		}
		matches = append(matches, installation)
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no valid github app installations found for authenticated user")
	}

	options := make([]rez.ExternalIntegrationOption, 0, len(matches))
	for _, installation := range matches {
		options = append(options, rez.ExternalIntegrationOption{
			ExternalRef: strconv.FormatInt(installation.GetID(), 10),
			DisplayName: installation.GetAccount().GetLogin(),
			Config: map[string]any{
				configOrg:            installation.GetAccount().GetLogin(),
				configInstallationID: installation.GetID(),
			},
		})
	}
	return options, nil
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

func (i *integration) MakeProviderSourceEventQueriers(ctx context.Context, intg *ent.Integration) ([]rez.ProviderEventQuerier, error) {
	ci := newConfiguredIntegration(i.services, intg)
	client, clientErr := newClient(ctx, ci)
	if clientErr != nil {
		return nil, clientErr
	}
	return []rez.ProviderEventQuerier{
		&repositoryEventQuerier{ci: ci, client: client},
	}, nil
}

func (i *integration) WebhookHandlers() map[string]http.Handler {
	return i.webhookHandlers
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
	return execution.NewTenantContext(ctx, ci.intg.TenantID)
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

func (ci *ConfiguredIntegration) ID() uuid.UUID {
	return ci.intg.ID
}

func (ci *ConfiguredIntegration) Provider() string {
	return ci.intg.Provider
}

func (ci *ConfiguredIntegration) DisplayName() string {
	return ci.intg.DisplayName
}

func (ci *ConfiguredIntegration) ExternalRef() string {
	return ci.intg.ExternalRef
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
