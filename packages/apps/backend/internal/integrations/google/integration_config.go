package google

import (
	"encoding/json"
	"fmt"

	rez "github.com/rezible/rezible"
)

type InstallationConfig struct {
	CustomerID                string
	ServiceAccountCredentials json.RawMessage
}

func (c *InstallationConfig) Encode() ([]byte, error) {
	return json.Marshal(c)
}

func (c *InstallationConfig) InstallationTargetRef() rez.ProviderResourceRef {
	return rez.ProviderResourceRef{
		Provider:          "google",
		ProviderNamespace: "google",
		ResourceRef:       c.CustomerID,
	}
}

func (c *InstallationConfig) Validate() error {
	if c.CustomerID == "" {
		return fmt.Errorf("customer ID is required")
	}
	if len(c.ServiceAccountCredentials) == 0 {
		return fmt.Errorf("service account credentials are required")
	}
	return nil
}

type UserSettings struct {
	EnableVideoConference bool
}
