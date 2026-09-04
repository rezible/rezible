package slackintegration

import (
	"encoding/json"
	"fmt"
	"strings"

	rez "github.com/rezible/rezible"
)

type InstallationConfig struct {
	AccessToken         string
	BotUserID           string
	WebhookChannelId    string
	IsEnterpriseInstall bool
	Team                *TeamInfo
	Enterprise          *TeamInfo
}

type TeamInfo struct {
	Id   string
	Name string
}

func MakeInstallationConfig(integrationName string, credentials *InstallationConfig) rez.IntegrationInstallationConfig {
	return appInstallationConfig{
		providerNamespace: integrationName,
		credentials:       credentials,
	}
}

type appInstallationConfig struct {
	providerNamespace string
	credentials       *InstallationConfig
}

func (c appInstallationConfig) Encode() ([]byte, error) {
	return c.credentials.Encode()
}

func (c appInstallationConfig) InstallationTargetRef() rez.ProviderResourceRef {
	ids := InstallationIds{}
	if c.credentials.Team != nil {
		ids.TeamId = c.credentials.Team.Id
	}
	if c.credentials.Enterprise != nil {
		ids.EnterpriseId = c.credentials.Enterprise.Id
	}
	return rez.ProviderResourceRef{
		Provider:          ProviderName,
		ProviderNamespace: c.providerNamespace,
		ResourceRef:       ids.InstallationTargetResourceRef(),
	}
}

func GetValidatedConfig(c []byte) (*InstallationConfig, error) {
	var cfg InstallationConfig
	if decErr := json.Unmarshal(c, &cfg); decErr != nil {
		return nil, fmt.Errorf("decode Slack installation config: %w", decErr)
	}
	if validateErr := cfg.Validate(); validateErr != nil {
		return nil, validateErr
	}
	return &cfg, nil
}

func (c *InstallationConfig) Validate() error {
	if c == nil {
		return fmt.Errorf("installation config is required")
	}
	if strings.TrimSpace(c.AccessToken) == "" {
		return fmt.Errorf("access token is required")
	}
	if strings.TrimSpace(c.BotUserID) == "" {
		return fmt.Errorf("bot user ID is required")
	}
	if c.Team != nil && strings.TrimSpace(c.Team.Id) == "" {
		return fmt.Errorf("team ID is required")
	}
	if c.Enterprise != nil && strings.TrimSpace(c.Enterprise.Id) == "" {
		return fmt.Errorf("enterprise ID is required")
	}
	if c.Team == nil && c.Enterprise == nil {
		return fmt.Errorf("team or enterprise installation target is required")
	}
	return nil
}

func (c *InstallationConfig) Encode() ([]byte, error) {
	return json.Marshal(c)
}

func (c *InstallationConfig) DisplayName() string {
	if c == nil || c.Team == nil && c.Enterprise == nil {
		return ""
	}
	if c.Enterprise == nil {
		return c.Team.Name
	}
	if c.Team == nil {
		return c.Enterprise.Name + " (Enterprise)"
	}
	return fmt.Sprintf("%s (%s)", c.Team.Name, c.Enterprise.Name)
}
