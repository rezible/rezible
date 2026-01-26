package fakeprovider

import (
	"context"
	"encoding/json"

	rez "github.com/rezible/rezible"
)

const integrationName = "fake"

type integration struct{}

func SetupIntegration(ctx context.Context, svcs *rez.Services) (rez.IntegrationPackage, error) {
	intg := &integration{}
	return intg, nil
}

func (d *integration) Name() string {
	return integrationName
}

func (d *integration) Enabled() bool {
	return rez.Config.DebugMode()
}

func (d *integration) SupportedDataKinds() []string {
	return []string{}
}

func (d *integration) OAuthConfigRequired() bool {
	return false
}

func (d *integration) ValidateConfig(cfg json.RawMessage) (bool, error) {

	return true, nil
}

func (d *integration) GetUserConfig(rawCfg json.RawMessage) (json.RawMessage, error) {
	return rawCfg, nil
}

func (d *integration) MergeUserConfig(rawCfg json.RawMessage, rawUserCfg json.RawMessage) (json.RawMessage, error) {
	return rawCfg, nil
}
