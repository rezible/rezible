package genkit

import (
	"context"
	"testing"

	gkai "github.com/firebase/genkit/go/ai"
	"github.com/rezible/rezible/test/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/test"
)

type AiRuntimeSuite struct {
	test.Suite
}

func TestAiRuntimeSuite(t *testing.T) {
	suite.Run(t, &AiRuntimeSuite{Suite: test.NewSuite()})
}

// makeRuntime initializes a runtime serving models, with testModelName as the default model.
func (s *AiRuntimeSuite) makeRuntime(ctx context.Context, models ...testModel) *AiRuntime {
	if len(models) == 0 {
		s.T().Log("adding default test model to runtime as none were given")
		models = append(models, makeTestOutputModel(&gkai.ModelResponse{Message: gkai.NewModelTextMessage("foo")}))
	}
	cfg := s.Config()
	cfg.AI.Models.Default = testModelName
	svc := newTestRuntime(cfg, models...)
	s.Require().NoError(svc.Init(ctx))
	return svc
}

func newTestRuntime(cfg rez.Config, models ...testModel) *AiRuntime {
	actions := make([]*gkai.ModelAction, len(models))
	for i, model := range models {
		actions[i] = gkai.NewModelAction(model.Name, model.opts, model.fn)
	}
	return NewAiRuntime(cfg, actions...)
}

func (s *AiRuntimeSuite) TestInitRejectsModelsWithoutAProvider() {
	ctx := s.T().Context()
	response := &gkai.ModelResponse{Message: gkai.NewModelTextMessage("foo")}

	cfg := s.Config()
	cfg.AI.Models.Default = testModelName
	cfg.AI.Models.Investigation = "unconfigured/model"
	initErr := newTestRuntime(cfg, makeTestOutputModel(response)).Init(ctx)
	s.ErrorContains(initErr, `"unconfigured/model"`)
}

func (s *AiRuntimeSuite) TestInitDefaultsToTheOnlyEnabledProvidersModel() {
	cfg := s.Config()
	cfg.AI.Providers.Anthropic = rez.AiProviderConfigAnthropic{Enabled: true, APIKey: "test-key"}
	svc := NewAiRuntime(cfg)
	s.Require().NoError(svc.Init(s.T().Context()))
	s.Equal("anthropic/claude-haiku-5-5", svc.modelFor(rezai.ChatAgent.Name))

	cfg.AI.Providers.GoogleAI = rez.AiProviderConfigGoogleAI{Enabled: true, APIKey: "test-key"}
	s.ErrorContains(NewAiRuntime(cfg).Init(s.T().Context()), "ai.models.default is required")
}

// testModelName is the model makeRuntime configures as the default.
const testModelName = "test/output"

// testModel is a stand-in model, which a test can adjust before a runtime serves it.
type testModel struct {
	Name string
	opts *gkai.ModelOptions
	fn   gkai.ModelActionFunc[any]
}

func makeTestOutputModel(response *gkai.ModelResponse) testModel {
	return testModel{
		Name: testModelName,
		opts: &gkai.ModelOptions{
			Supports: &gkai.ModelSupports{
				Constrained: gkai.ConstrainedSupportAll,
				Multiturn:   true,
				SystemRole:  true,
			},
		},
		fn: func(ctx context.Context, req *gkai.ModelRequest, cfg any, cb gkai.ModelStreamCallback) (*gkai.ModelResponse, error) {
			return response, nil
		},
	}
}

func (s *AiRuntimeSuite) TestIntegrationToolsMiddlewareLoadsToolsPerTurn() {
	agentName := "test-agent"

	tool := gkai.NewTool[any, map[string]any](
		"test_integration_lookup",
		"test integration lookup",
		func(ctx *gkai.ToolContext, input any) (map[string]any, error) {
			return map[string]any{"ok": true}, nil
		},
	)

	nextTool := gkai.NewTool[any, map[string]any](
		"test_next_lookup",
		"next tool",
		func(ctx *gkai.ToolContext, input any) (map[string]any, error) {
			return map[string]any{"next": true}, nil
		},
	)
	intgs := mocks.NewMockIntegrationService(s.T())
	intgs.EXPECT().
		GetAvailableAgentTools(mock.Anything, rez.GetAvailableAiAgentToolsParams{AgentName: agentName}).
		Return([]gkai.Tool{tool}, nil).
		Once()
	intgs.EXPECT().
		GetAvailableAgentTools(mock.Anything, rez.GetAvailableAiAgentToolsParams{AgentName: agentName}).
		Return([]gkai.Tool{nextTool}, nil).
		Once()

	ctx := s.T().Context()

	mw := newIntegrationToolsMiddleware(agentName, intgs)
	firstHooks, firstErr := mw.New(ctx)
	s.Require().NoError(firstErr, "first middleware init")
	s.Require().NotEmpty(firstHooks.Tools, "no tools supplied")
	s.Require().Equal(tool.Name(), firstHooks.Tools[0].Name(), "unexpected tool name")

	secondHooks, secondErr := mw.New(ctx)
	s.Require().NoError(secondErr, "second middleware init")
	s.Require().Len(secondHooks.Tools, 1)
	s.Require().Equal(nextTool.Name(), secondHooks.Tools[0].Name(), "unexpected tool name")
}
