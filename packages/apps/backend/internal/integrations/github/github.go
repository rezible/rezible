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

type Integration struct {
	cfg                   rez.IntegrationsConfigGithub
	oauth2Config          *oauth2.Config
	listUserInstallations func(context.Context, string) ([]*gh.Installation, error)
	webhookHandler        http.Handler
}

func MakeIntegration(cfg rez.Config, provEvents rez.ProviderEventService) (*Integration, error) {
	i := &Integration{
		cfg:                   cfg.Integrations.Github,
		listUserInstallations: listUserInstallations,
		webhookHandler:        http.NotFoundHandler(),
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

func (i *Integration) IsAvailable() (bool, error) {
	if !i.cfg.Enabled {
		return false, nil
	}
	return true, nil
}

func (i *Integration) WebhookHandler() http.Handler {
	return i.webhookHandler
}

func (i *Integration) SupportedDataKinds() []string {
	return supportedDataKinds
}

func (i *Integration) OAuthConfigRequired() bool {
	return true
}

func (i *Integration) OAuth2Config() *oauth2.Config {
	return i.oauth2Config
}

func (i *Integration) loadOAuthConfig() *oauth2.Config {
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

func (i *Integration) ExtractIntegrationOptionsFromToken(t *oauth2.Token) ([]rez.ExternalIntegrationOption, error) {
	if t == nil || t.AccessToken == "" {
		return nil, fmt.Errorf("missing access token")
	}
	if i.listUserInstallations == nil {
		i.listUserInstallations = listUserInstallations
	}

	// The shared OAuth Integration contract does not carry a request context into this hook.
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

type userConfig struct {
	data objx.Map
}

func (i *Integration) ValidateUserConfig(cfg map[string]any) error {
	//uc := userConfig{data: objx.New(cfg)}
	/*
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
	*/

	return nil
}

func (i *Integration) ValidateUserPreferences(prefs map[string]any) error {
	return nil
}

func (i *Integration) GetConfiguredIntegration(intg *ent.Integration) rez.ConfiguredIntegration {
	return i.newConfiguredIntegration(intg)
}

// ConfiguredIntegration wraps an *ent.Integration for a specific tenant installation.
type ConfiguredIntegration struct {
	intg *ent.Integration
	cfg  userConfig
}

func (i *Integration) newConfiguredIntegration(intg *ent.Integration) *ConfiguredIntegration {
	return &ConfiguredIntegration{
		intg: intg,
		cfg:  userConfig{data: objx.New(intg.Config)},
	}
}

func (ci *ConfiguredIntegration) Integration() *ent.Integration {
	return ci.intg
}

func (ci *ConfiguredIntegration) tenantContext(ctx context.Context) context.Context {
	return execution.NewTenantContext(ctx, ci.intg.TenantID)
}

func (ci *ConfiguredIntegration) config() userConfig {
	return userConfig{data: objx.New(ci.intg.Config)}
}

const (
	configOrg            = "org"
	configInstallationID = "installation_id"
)

func (ci *ConfiguredIntegration) orgName() string {
	return ci.cfg.data.Get(configOrg).String()
}

func (ci *ConfiguredIntegration) installationID() int64 {
	v := ci.cfg.data.Get(configInstallationID)
	if v.IsFloat64() {
		return int64(v.Float64())
	}
	return 0
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
	return ci.cfg.data.Exclude([]string{})
}

func (ci *ConfiguredIntegration) GetUserPreferences() map[string]any {
	return ci.intg.UserPreferences
}

func (ci *ConfiguredIntegration) GetAvailableDataKinds() map[string]bool {
	return map[string]bool{
		"repositories":  true,
		"change_events": true,
	}
}
