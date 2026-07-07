package slackintegration

import (
	"encoding/json"
	"fmt"
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

func GetValidatedConfig(c []byte) (*InstallationConfig, error) {
	var cfg InstallationConfig
	if decErr := json.Unmarshal(c, &cfg); decErr != nil {
		return nil, decErr
	}
	// TODO: validate
	return &cfg, nil
}

func (c *InstallationConfig) Encode() ([]byte, error) {
	return json.Marshal(c)
}

func (c *InstallationConfig) ExternalRef() string {
	var teamId string
	if c.Team != nil {
		teamId = c.Team.Id
	}
	var enterpriseId string
	if c.Enterprise != nil {
		enterpriseId = c.Enterprise.Id
	}
	ids := InstallationIds{TeamId: teamId, EnterpriseId: enterpriseId}
	return ids.asRef()
}

func (c *InstallationConfig) DisplayName() string {
	if c.Enterprise == nil {
		return c.Team.Name
	}
	if c.Team == nil {
		return c.Enterprise.Name + " (Enterprise)"
	}
	return fmt.Sprintf("%s (%s)", c.Team.Name, c.Enterprise.Name)
}
