package genkit

import (
	"context"
	"errors"
	"testing"

	"github.com/firebase/genkit/go/ai"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/test"
	"github.com/stretchr/testify/suite"
)

type testEvalScenario struct {
	agentName string
	passed    bool
	gradeErr  error
}

func (s testEvalScenario) Definition() rezai.EvalScenarioDefinition {
	return rezai.EvalScenarioDefinition{
		Name:        "test/scenario",
		Description: "test scenario",
		AgentName:   s.agentName,
	}
}

func (testEvalScenario) Seed(context.Context, *ent.Client) (rezai.EvalScenarioSeed, error) {
	return rezai.EvalScenarioSeed{Input: testAgentInput{}}, nil
}

func (s testEvalScenario) Grade(context.Context, *ent.Client, *rez.AiAgentInvocationResult) (rezai.EvalScenarioGrade, error) {
	if s.gradeErr != nil {
		return rezai.EvalScenarioGrade{}, s.gradeErr
	}
	checks := []rezai.EvalCheck{
		{ID: "scenario_check", Passed: s.passed, Summary: "Scenario check."},
	}
	return rezai.EvalScenarioGrade{Output: "hello", Checks: checks}, nil
}

type EvaluationServiceSuite struct {
	test.Suite
}

func TestEvaluationServiceSuite(t *testing.T) {
	suite.Run(t, &EvaluationServiceSuite{Suite: test.NewSuite()})
}

func (s *EvaluationServiceSuite) TestRunsAgentAndProducesPassingReport() {
	ctx := s.SeedTenantContext()
	database := s.CreateTestDatabase()
	response := &ai.ModelResponse{
		Message:      ai.NewModelTextMessage("hello"),
		FinishReason: ai.FinishReasonStop,
	}
	agent := makeTestAgent[testAgentState](ai.NewUserTextMessage("say hello"))
	agent.def.Model = "test/model"
	aiService := NewAiService(s.Config())
	s.Require().NoError(aiService.Init(ctx, withTestModel(response), WithAgent(agent)))

	service := NewEvaluationService(database, aiService)
	result := service.RunScenario(ctx, testEvalScenario{
		agentName: agent.def.Name,
		passed:    true,
	})
	s.Equal(rezai.EvalRunStatusPassed, result.Status)
	s.Require().NotNil(result.Execution)
	s.True(result.Execution.Succeeded)
	s.Equal("test/model", result.Agent.Model)
	s.Len(result.Checks, 1)
}

func (s *EvaluationServiceSuite) TestReportsGradeErrorsAtGradeStage() {
	ctx := s.SeedTenantContext()
	database := s.CreateTestDatabase()
	response := &ai.ModelResponse{
		Message:      ai.NewModelTextMessage("hello"),
		FinishReason: ai.FinishReasonStop,
	}
	agent := makeTestAgent[testAgentState](ai.NewUserTextMessage("say hello"))
	agent.def.Model = "test/model"
	aiService := NewAiService(s.Config())
	s.Require().NoError(aiService.Init(ctx, withTestModel(response), WithAgent(agent)))
	expectedErr := errors.New("grading unavailable")

	service := NewEvaluationService(database, aiService)
	result := service.RunScenario(ctx, testEvalScenario{
		agentName: agent.def.Name,
		gradeErr:  expectedErr,
	})
	s.Equal(rezai.EvalRunStatusError, result.Status)
	s.Require().NotNil(result.Error)
	s.Equal(rezai.EvalRunStageGrade, result.Error.Stage)
	s.Require().NotNil(result.Execution)
	s.True(result.Execution.Succeeded)
}
