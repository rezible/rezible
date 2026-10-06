package genkit

import (
	"context"
	"testing"

	gkai "github.com/firebase/genkit/go/ai"
	"github.com/rezible/rezible/test/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/test"
)

type AiRuntimeSuite struct {
	test.Suite
}

func TestAiRuntimeSuite(t *testing.T) {
	suite.Run(t, &AiRuntimeSuite{Suite: test.NewSuite()})
}

func (s *AiRuntimeSuite) makeRuntime(ctx context.Context, opts ...AiRuntimeOption) *AiRuntime {
	svc := NewAiRuntime(s.Config())
	var anyModels bool
	for _, opt := range opts {
		if opt.kind == AiRuntimeOptionKindModel {
			anyModels = true
			break
		}
	}
	if !anyModels {
		s.T().Log("adding default test model to runtime init opts as none were defined")
		response := &gkai.ModelResponse{Message: gkai.NewModelTextMessage("foo")}
		opts = append(opts, WithDefinedModel(makeTestOutputModel(response)))
	}
	s.Require().NoError(svc.Init(ctx, opts...))
	return svc
}

func (s *AiRuntimeSuite) TestSituationJudgeRequiresADefaultModel() {
	enabled := rez.AiConfig{SituationJudge: rez.AiSituationJudgeConfig{Enabled: true}}
	s.Error((&AiRuntime{cfg: enabled}).validateConfig(), "enabling the judge without a default model fails")
	s.NoError((&AiRuntime{cfg: enabled, defaultModel: "test/output"}).validateConfig())
	s.NoError((&AiRuntime{}).validateConfig(), "the judge is disabled by default")
}

func makeTestOutputModel(response *gkai.ModelResponse) ModelDefinition[any] {
	return ModelDefinition[any]{
		Name:      "test/output",
		IsDefault: true,
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
