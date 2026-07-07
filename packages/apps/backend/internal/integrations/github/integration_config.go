package github

import (
	"encoding/json"
	"fmt"
	"strconv"

	gh "github.com/google/go-github/v84/github"
)

func (i *Integration) MakeInstallationConfigFromGithub(inst *gh.Installation) (*InstallationConfig, error) {
	if inst.GetAppID() != i.cfg.App.AppID {
		return nil, fmt.Errorf("invalid app id %d", inst.GetAppID())
	}
	cfg := &InstallationConfig{
		Org:            inst.GetAccount().GetLogin(),
		InstallationID: inst.GetID(),
	}
	return cfg, nil
}

type InstallationConfig struct {
	Org            string
	InstallationID int64
}

func (ic *InstallationConfig) Encode() ([]byte, error) {
	return json.Marshal(&ic)
}

func (ic *InstallationConfig) ExternalRef() string {
	return strconv.FormatInt(ic.InstallationID, 10)
}
