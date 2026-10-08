package grafana

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/pkg/errs"
)

const (
	ProviderName    = "grafana"
	integrationName = "grafana"
)

var capabilities = []string{"logs", "metrics"}

type Integration struct {
	clock      rez.Clock
	httpClient *http.Client
}

func MakeIntegration(clock rez.Clock) (*Integration, error) {
	httpClient := &http.Client{
		// A redirect would send the read somewhere the admin did not configure; report it instead.
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	i := &Integration{
		clock:      clock,
		httpClient: httpClient,
	}
	return i, nil
}

func (i *Integration) Name() string {
	return integrationName
}

func (i *Integration) DisplayName() string {
	return "Grafana"
}

func (i *Integration) Description() string {
	return "Read logs and metrics from Loki and Prometheus through Grafana"
}

func (i *Integration) Provider() string {
	return ProviderName
}

func (i *Integration) Capabilities() []string {
	return capabilities
}

func (i *Integration) MaxInstalls() *int {
	return new(1)
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

func (i *Integration) ValidateInstallationConfig(raw []byte) (rez.IntegrationInstallationConfig, error) {
	return parseInstallationConfig(raw)
}

func (i *Integration) ValidateUserSettings(settings map[string]any) error {
	_, parseErr := parseInstallationSettings(settings)
	return parseErr
}

func (i *Integration) GetInstalledIntegration(intg *ent.Integration) (rez.InstalledIntegration, error) {
	config, configErr := parseInstallationConfig(intg.InstallationConfig)
	if configErr != nil {
		return nil, fmt.Errorf("installation config: %w", configErr)
	}
	settings, settingsErr := parseInstallationSettings(intg.UserSettings)
	if settingsErr != nil {
		return nil, fmt.Errorf("installation settings: %w", settingsErr)
	}
	ii := &InstalledIntegration{
		intg:     intg,
		config:   config,
		settings: settings,
		client:   &client{http: i.httpClient, baseURL: config.URL, token: config.Token},
		clock:    i.clock,
	}
	return ii, nil
}

type InstalledIntegration struct {
	intg     *ent.Integration
	config   *InstallationConfig
	settings *installationSettings
	client   *client
	clock    rez.Clock
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

// InstallationConfig is what an admin supplies when installing. The token is a secret: it is never returned by
// the API or logged.
type InstallationConfig struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

func parseInstallationConfig(raw []byte) (*InstallationConfig, error) {
	var config InstallationConfig
	if decodeErr := json.Unmarshal(raw, &config); decodeErr != nil {
		return nil, fmt.Errorf("%w: config must be an object with a url and a token", errs.ErrInvalidInput)
	}
	normalizedURL, urlErr := normalizeURL(config.URL)
	if urlErr != nil {
		return nil, urlErr
	}
	config.URL = normalizedURL
	config.Token = strings.TrimSpace(config.Token)
	if config.Token == "" {
		return nil, fmt.Errorf("%w: a service account token is required", errs.ErrInvalidInput)
	}
	return &config, nil
}

// normalizeURL returns Grafana's base URL without a trailing slash.
func normalizeURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("%w: a Grafana URL is required", errs.ErrInvalidInput)
	}
	parsed, parseErr := url.Parse(trimmed)
	if parseErr != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", fmt.Errorf("%w: the Grafana URL must be an http or https URL", errs.ErrInvalidInput)
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("%w: the Grafana URL must not contain credentials, a query or a fragment", errs.ErrInvalidInput)
	}
	base := parsed.Scheme + "://" + parsed.Host + parsed.EscapedPath()
	return strings.TrimRight(base, "/"), nil
}

func (c *InstallationConfig) Encode() ([]byte, error) {
	return json.Marshal(c)
}

// InstallationTargetRef identifies the installation by its URL, so installing again with the same URL replaces
// the token and keeps the settings.
func (c *InstallationConfig) InstallationTargetRef() rez.ProviderResourceRef {
	return rez.ProviderResourceRef{
		Provider:          ProviderName,
		ProviderNamespace: integrationName,
		ResourceRef:       c.URL,
	}
}
