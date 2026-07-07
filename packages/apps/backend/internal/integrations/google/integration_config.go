package google

import (
	"encoding/json"
)

type InstallationConfig struct {
	ServiceAccountCredentials []byte
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
