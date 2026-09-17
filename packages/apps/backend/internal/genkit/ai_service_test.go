package genkit

import (
	"context"
	"os"
	"testing"

	"github.com/firebase/genkit/go/ai"
	gk "github.com/firebase/genkit/go/genkit"
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

func (s *AiRuntimeSuite) checkSkip(name string) {
	if os.Getenv("AI_TESTS_ALL") != "true" && os.Getenv("AI_TESTS_"+name) != "true" {
		s.T().Skipf("Skipping live AI test '%s'", name)
	}
}

func withTestModel(response *ai.ModelResponse) AiRuntimeOption {
	return AiRuntimeOption{
		kind: AiRuntimeOptionKindModel,
		optFn: func(s *AiRuntime) error {
			gk.DefineModel(s.gk, "test/model", &ai.ModelOptions{
				Supports: &ai.ModelSupports{
					Constrained: ai.ConstrainedSupportAll,
					Multiturn:   true,
					SystemRole:  true,
				},
			}, func(ctx context.Context, req *ai.ModelRequest, cb ai.ModelStreamCallback) (*ai.ModelResponse, error) {
				return response, nil
			})
			return nil
		},
	}
}

func (s *AiRuntimeSuite) makeService(ctx context.Context, opts ...AiRuntimeOption) *AiRuntime {
	svc := NewAiRuntime(s.Config())
	s.Require().NoError(svc.Init(ctx, opts...))
	return svc
}

func (s *AiRuntimeSuite) TestIntegrationToolsMiddlewareLoadsToolsPerTurn() {
	agentName := "test-agent"

	tool := ai.NewTool[any, map[string]any](
		"test_integration_lookup",
		"test integration lookup",
		func(ctx *ai.ToolContext, input any) (map[string]any, error) {
			return map[string]any{"ok": true}, nil
		},
	)

	intgs := mocks.NewMockIntegrationService(s.T())
	intgs.EXPECT().
		GetAvailableAgentTools(mock.Anything, rez.GetAvailableAiAgentToolsParams{AgentName: agentName}).
		Return([]ai.Tool{tool}, nil).
		Twice()

	ctx := s.T().Context()

	mw := newIntegrationToolsMiddleware(agentName, intgs)
	firstHooks, firstErr := mw.New(ctx)
	s.Require().NoError(firstErr, "first middleware init")
	s.Require().NotEmpty(firstHooks.Tools, "no tools supplied")
	s.Require().Equal(tool.Name(), firstHooks.Tools[0].Name(), "unexpected tool name")

	secondHooks, secondErr := mw.New(ctx)
	s.Require().NoError(secondErr, "second middleware init")
	s.Require().NotEmpty(secondHooks.Tools, "no tools supplied")
	s.Require().Equal(tool.Name(), secondHooks.Tools[0].Name(), "unexpected tool name")
}
