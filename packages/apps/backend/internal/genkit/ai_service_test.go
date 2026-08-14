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

type AiServiceSuite struct {
	test.Suite
}

func TestAiServiceSuite(t *testing.T) {
	suite.Run(t, &AiServiceSuite{Suite: test.NewSuite()})
}

func (s *AiServiceSuite) checkSkip(name string) {
	if os.Getenv("AI_TESTS_ALL") != "true" && os.Getenv("AI_TESTS_"+name) != "true" {
		s.T().Skipf("Skipping live AI test '%s'", name)
	}
}

func withTestModel(response *ai.ModelResponse) AiServiceOption {
	return AiServiceOption{
		kind: AiServiceOptionKindModel,
		optFn: func(s *AiService) error {
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

func (s *AiServiceSuite) makeService(ctx context.Context, opts ...AiServiceOption) *AiService {
	svc := NewAiService(s.Config())
	s.Require().NoError(svc.Init(ctx, opts...))
	return svc
}

func (s *AiServiceSuite) TestIntegrationToolsMiddlewareLoadsToolsPerTurn() {
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
		GetAvailableAgentTools(mock.Anything, rez.GetAvailableAgentToolsParams{AgentName: agentName}).
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
