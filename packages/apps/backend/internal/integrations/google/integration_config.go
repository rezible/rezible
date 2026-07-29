package google

import (
	"encoding/json"
)

type InstallationConfig struct {
	ServiceAccountCredentials json.RawMessage
}

func (c *InstallationConfig) Encode() ([]byte, error) {
	return json.Marshal(c)
}

func (c *InstallationConfig) ExternalRef() string {
	return "todo"
}

type UserSettings struct {
	EnableVideoConference bool
}
