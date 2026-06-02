package slackintegration

import (
	"fmt"

	"github.com/go-viper/mapstructure/v2"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type InstallationConfig struct {
	AccessToken         string    `mapstructure:"access_token"`
	BotUserID           string    `mapstructure:"bot_user_id"`
	WebhookChannelId    string    `mapstructure:"webhook_channel_id"`
	IsEnterpriseInstall bool      `mapstructure:"is_enterprise_install"`
	Team                *TeamInfo `mapstructure:"team"`
	Enterprise          *TeamInfo `mapstructure:"enterprise"`
}

type TeamInfo struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (c *InstallationConfig) ExternalRef() (string, error) {
	if c.Team == nil && c.Enterprise == nil {
		return "", fmt.Errorf("no team or enterprise configured")
	}
	ids := IntegrationInstallIds{
		TeamId:       c.Team.Id,
		EnterpriseId: c.Enterprise.Id,
	}
	return ids.asRef(), nil
}

func (c *InstallationConfig) DisplayName() (string, error) {
	if c.Team == nil && c.Enterprise == nil {
		return "", fmt.Errorf("no team or enterprise configured")
	}
	if c.Team == nil {
		return c.Enterprise.Name + " (Enterprise)", nil
	}
	return c.Team.Name, nil
}

func (c *InstallationConfig) EncodeConfig() (map[string]any, error) {
	var data map[string]any
	if err := mapstructure.Decode(c, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func DecodeInstallationConfig(intg *ent.Integration) (*InstallationConfig, error) {
	var cfg InstallationConfig
	if decErr := mapstructure.Decode(intg.InstallationConfig, &cfg); decErr != nil {
		return nil, decErr
	}
	return &cfg, nil
}

func MakeConfigInstallationTarget(c *InstallationConfig) (*rez.IntegrationInstallationTarget, error) {
	var target rez.IntegrationInstallationTarget
	var err error
	if target.ExternalRef, err = c.ExternalRef(); err != nil {
		return nil, fmt.Errorf("external ref: %w", err)
	}
	if target.DisplayName, err = c.DisplayName(); err != nil {
		return nil, fmt.Errorf("display name: %w", err)
	}
	if target.InstallationConfig, err = c.EncodeConfig(); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
	return &target, nil
}
