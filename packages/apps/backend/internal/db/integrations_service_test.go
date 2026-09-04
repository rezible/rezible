package db

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/firebase/genkit/go/ai"
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

func (s *IntegrationsServiceSuite) newRegistry(pkgs ...rez.IntegrationDefinition) rez.IntegrationRegistry {
	reg := integrations.NewRegistry()
	for _, pkg := range pkgs {
		s.Require().NoError(reg.Register(pkg))
	}
	return reg
}

func (s *IntegrationsServiceSuite) newService(tdb rez.Database, reg rez.IntegrationRegistry) *IntegrationsService {
	jobs := mocks.NewMockJobService(s.T())
	jobs.EXPECT().Insert(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, nil)

	svc, err := NewIntegrationsService(s.Config(), tdb, jobs, reg)
	s.Require().NoError(err)

	return svc
}

func (s *IntegrationsServiceSuite) installTestIntegration(ctx context.Context, svc *IntegrationsService, i rez.IntegrationDefinition, ref string) rez.InstalledIntegration {
	config := &testInstalledIntegrationConfig{
		TestRef:           ref,
		ProviderNamespace: i.Name(),
	}
	target := rez.IntegrationInstallationTarget{
		DisplayName: i.DisplayName(),
		Config:      config,
	}
	s.T().Logf("installing integration %+v", target)
	ii, installErr := svc.InstallFromTarget(ctx, target)
	s.Require().NoError(installErr)
	return ii
}

func (s *IntegrationsServiceSuite) TestInstallIntegration() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()

	i := &testIntegration{
		maxInstalls: new(1),
	}
	reg := s.newRegistry(i)

	svc := s.newService(tdb, reg)

	cfg := testInstalledIntegrationConfig{TestRef: "foobar"}
	rawCfg, jsonErr := json.Marshal(cfg)
	s.Require().NoError(jsonErr)

	ic, cfgErr := i.ValidateInstallationConfig(rawCfg)
	s.Require().NoError(cfgErr)

	target := rez.IntegrationInstallationTarget{
		DisplayName: i.DisplayName(),
		Config:      ic,
	}
	ii, installErr := svc.InstallFromTarget(ctx, target)
	s.Require().NoError(installErr)

	s.Require().Equal(cfg.TestRef, ii.Config().InstallationTargetRef().ResourceRef)
}

func (s *IntegrationsServiceSuite) TestInstallSameTargetUpdatesExistingInstallation() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()

	i := &testIntegration{}
	svc := s.newService(tdb, s.newRegistry(i))

	first := s.installTestIntegration(ctx, svc, i, "same-target")
	second := s.installTestIntegration(ctx, svc, i, "same-target")

	s.Equal(first.Integration().ID, second.Integration().ID)
	query := tdb.Client(ctx).Integration.Query()
	count, countErr := query.Count(ctx)
	s.Require().NoError(countErr)
	s.Equal(1, count)
}

func (s *IntegrationsServiceSuite) TestGetAvailableAgentToolsSkipsIntegrationsWithoutTools() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	client := tdb.Client(ctx)

	intgs := client.Integration.Query().AllX(ctx)
	for _, intg := range intgs {
		fmt.Printf("\nintegration %+v\n", intg)
	}

	i := &testIntegration{}
	svc := s.newService(tdb, s.newRegistry(i))
	s.installTestIntegration(ctx, svc, i, "target-a")

	tools, toolsErr := svc.GetAvailableAgentTools(ctx, rez.GetAvailableAgentToolsParams{})
	s.Require().NoError(toolsErr)
	s.Empty(tools)
}

func (s *IntegrationsServiceSuite) TestGetAvailableAgentToolsRejectsDuplicateToolNames() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()

	tools := []ai.Tool{newTestAgentTool("duplicate_tool")}
	i1 := &testIntegration{name: "pkg-a", tools: tools}
	i2 := &testIntegration{name: "pkg-b", tools: tools}

	svc := s.newService(tdb, s.newRegistry(i1, i2))
	s.installTestIntegration(ctx, svc, i1, "target-a")
	s.installTestIntegration(ctx, svc, i2, "target-b")

	_, toolsErr := svc.GetAvailableAgentTools(ctx, rez.GetAvailableAgentToolsParams{})
	s.Require().Error(toolsErr)
	s.ErrorContains(toolsErr, "duplicate agent tool")
}

type testIntegration struct {
	name        string
	unavailable bool
	maxInstalls *int
	tools       []ai.Tool
}

func (p *testIntegration) Name() string {
	if p.name != "" {
		return p.name
	}
	return "test-package"
}

func (p *testIntegration) DisplayName() string {
	return "Test Package"
}

func (p *testIntegration) Description() string {
	return ""
}

func (p *testIntegration) Capabilities() []string {
	return nil
}

func (p *testIntegration) Provider() string {
	return "testing"
}

func (p *testIntegration) IsAvailable() (bool, error) {
	return !p.unavailable, nil
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
	cfg.ProviderNamespace = p.Name()
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

func (p *testIntegration) GetAvailableAgentTools(ctx context.Context, installations []rez.InstalledIntegration, params rez.GetAvailableAgentToolsParams) ([]ai.Tool, error) {
	return p.tools, nil
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

func (i *testInstalledIntegration) Capabilities() []string {
	return nil
}

type testInstalledIntegrationConfig struct {
	TestRef           string
	ProviderNamespace string
}

func (c *testInstalledIntegrationConfig) Encode() ([]byte, error) {
	return json.Marshal(c)
}

func (c *testInstalledIntegrationConfig) InstallationTargetRef() rez.ProviderResourceRef {
	return rez.ProviderResourceRef{
		Provider:          "testing",
		ProviderNamespace: c.ProviderNamespace,
		ResourceRef:       c.TestRef,
	}
}

func newTestAgentTool(name string) ai.Tool {
	return ai.NewTool[map[string]any, map[string]any](
		name,
		"test tool",
		func(ctx *ai.ToolContext, input map[string]any) (map[string]any, error) {
			return map[string]any{"ok": true}, nil
		},
	)
}
