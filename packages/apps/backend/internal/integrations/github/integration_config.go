package github

import (
	"encoding/json"
	"fmt"
	"strconv"

	gh "github.com/google/go-github/v84/github"
	rez "github.com/rezible/rezible"
)

func (i *Integration) MakeInstallationConfigFromGithub(inst *gh.Installation) (*InstallationConfig, error) {
	if inst.GetAppID() != i.cfg.App.AppID {
		return nil, fmt.Errorf("invalid app id %d", inst.GetAppID())
	}
	cfg := &InstallationConfig{
		Org:            inst.GetAccount().GetLogin(),
		AccountID:      inst.GetAccount().GetID(),
		InstallationID: inst.GetID(),
	}
	if cfg.AccountID == 0 {
		return nil, fmt.Errorf("github installation is missing account id")
	}
	return cfg, nil
}

type InstallationConfig struct {
	Org            string
	AccountID      int64
	InstallationID int64
}

func (ic *InstallationConfig) Encode() ([]byte, error) {
	return json.Marshal(&ic)
}

func (ic *InstallationConfig) InstallationTargetRef() rez.ProviderResourceRef {
	return rez.ProviderResourceRef{
		Provider:          providerName,
		ProviderNamespace: integrationName,
		ResourceRef:       strconv.FormatInt(ic.AccountID, 10),
	}
}
