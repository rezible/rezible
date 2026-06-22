package demoprovider

import (
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

const (
	integrationName = "demo"
	providerName    = "rezible"
)

type Integration struct {
	available bool
}

func MakeIntegration(cfg rez.Config) *Integration {
	return &Integration{
		available: cfg.App.DebugMode,
	}
}

func (i *Integration) Name() string {
	return integrationName
}

func (i *Integration) DisplayName() string {
	return "Demo Data Provider"
}

func (i *Integration) Description() string {
	return "Provides demo data for testing purposes"
}

func (i *Integration) Provider() string {
	return providerName
}

func (i *Integration) MaxInstalls() *int {
	return new(1)
}

func (i *Integration) IsAvailable() (bool, error) {
	return i.available, nil
}

func (i *Integration) OAuthInstallRequired() bool {
	return false
}

func (i *Integration) ValidateInstallationConfig(m map[string]any) (externalRef string, validationErr error) {
	return "", nil
}

func (i *Integration) ValidateUserSettings(settings map[string]any) error {
	return nil
}

func (i *Integration) GetInstalledIntegration(intg *ent.Integration) rez.InstalledIntegration {
	return &InstalledIntegration{intg: intg}
}

type InstalledIntegration struct {
	intg *ent.Integration
}

func (ii *InstalledIntegration) Integration() *ent.Integration {
	return ii.intg
}

func (ii *InstalledIntegration) ProviderName() string {
	return providerName
}

func (ii *InstalledIntegration) SanitizedInstallationConfig() map[string]any {
	return ii.intg.InstallationConfig
}

type IntegrationConfig struct{}
