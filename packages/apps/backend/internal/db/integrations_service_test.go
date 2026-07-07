package db

import (
	"encoding/json"
	"fmt"
	"testing"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/pkg/integrations"
	"github.com/rezible/rezible/test"
	"github.com/rezible/rezible/test/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type IntegrationsServiceSuite struct {
	test.Suite
}

func TestIntegrationsServiceSuite(t *testing.T) {
	suite.Run(t, &IntegrationsServiceSuite{Suite: test.NewSuite()})
}

func (s *IntegrationsServiceSuite) newService(reg *integrations.PackageRegistry) *IntegrationsService {
	jobs := mocks.NewMockJobService(s.T())
	jobs.EXPECT().Insert(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)

	svc, err := NewIntegrationsService(s.Config().App, s.Database(), jobs, reg, nil)
	s.Require().NoError(err)
	return svc
}

func (s *IntegrationsServiceSuite) TestInstallIntegration() {
	ctx := s.SeedTenantContext()

	i := &testIntegration{
		available:   true,
		maxInstalls: new(1),
	}
	reg := integrations.NewPackageRegistry()
	s.Require().NoError(reg.RegisterPackage(i))

	svc := s.newService(reg)

	cfg := testInstalledIntegrationConfig{
		TestRef: "foobar",
	}
	rawCfg, jsonErr := json.Marshal(cfg)
	s.Require().NoError(jsonErr)

	ic, cfgErr := i.ValidateInstallationConfig(rawCfg)
	s.Require().NoError(cfgErr)

	target := rez.IntegrationInstallationTarget{
		IntegrationName: i.Name(),
		DisplayName:     i.DisplayName(),
		Config:          ic,
	}
	ii, installErr := svc.InstallFromTarget(ctx, target)
	s.Require().NoError(installErr)

	s.Require().Equal(cfg.TestRef, ii.Config().ExternalRef())
}

type testIntegration struct {
	available   bool
	maxInstalls *int
}

func (p *testIntegration) Name() string {
	return "test-package"
}

func (p *testIntegration) DisplayName() string {
	return "Test Package"
}

func (p *testIntegration) Description() string {
	return ""
}

func (p *testIntegration) Provider() string {
	return "testing"
}

func (p *testIntegration) IsAvailable() (bool, error) {
	return p.available, nil
}

func (p *testIntegration) MaxInstalls() *int {
	return p.maxInstalls
}

func (p *testIntegration) OAuthInstallRequired() bool {
	return false
}

func (p *testIntegration) ValidateInstallationConfig(raw []byte) (rez.IntegrationInstallationConfig, error) {
	var cfg testInstalledIntegrationConfig
	if jsonErr := json.Unmarshal(raw, &cfg); jsonErr != nil {
		return nil, fmt.Errorf("unmarshal test config: %w", jsonErr)
	}
	return &cfg, nil
}

func (p *testIntegration) ValidateUserSettings(m map[string]any) error {
	return nil
}

func (p *testIntegration) GetInstalledIntegration(intg *ent.Integration) (rez.InstalledIntegration, error) {
	ii := &testInstalledIntegration{intg: intg}
	if jsonErr := json.Unmarshal(intg.InstallationConfig, &ii.cfg); jsonErr != nil {
		return nil, fmt.Errorf("unmarshal test installed config: %w", jsonErr)
	}
	return ii, nil
}

type testInstalledIntegration struct {
	intg *ent.Integration
	cfg  *testInstalledIntegrationConfig
}

func (i *testInstalledIntegration) Integration() *ent.Integration {
	return i.intg
}

func (i *testInstalledIntegration) Config() rez.IntegrationInstallationConfig {
	return i.cfg
}

type testInstalledIntegrationConfig struct {
	TestRef string
}

func (c *testInstalledIntegrationConfig) Encode() ([]byte, error) {
	return json.Marshal(c)
}

func (c *testInstalledIntegrationConfig) ExternalRef() string {
	return c.TestRef
}
