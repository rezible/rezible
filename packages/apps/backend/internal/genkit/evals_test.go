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

type evaluationScenario struct {
	agentName string
	status    string
	judgeErr  error
}

func (s evaluationScenario) Definition() rezai.EvalScenarioDefinition {
	return rezai.EvalScenarioDefinition{Name: "test/scenario", AgentName: s.agentName}
}

func (evaluationScenario) Seed(context.Context, *ent.Client) (rezai.EvalScenarioSeed, error) {
	return rezai.EvalScenarioSeed{Input: testAgentInput{}}, nil
}

func (s evaluationScenario) Judge(context.Context, *ent.Client, *rez.AiAgentInvocationResult) ([]ai.Score, error) {
	if s.judgeErr != nil {
		return nil, s.judgeErr
	}
	return []ai.Score{{Id: "scenario_check", Score: s.status == ai.ScoreStatusPass.String(), Status: s.status}}, nil
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
	response := &ai.ModelResponse{Message: ai.NewModelTextMessage("hello")}
	agent := makeTestAgent[testAgentState](ai.NewUserTextMessage("say hello"))
	agent.def.Model = "test/model"
	aiService := NewAiService(s.Config())
	s.Require().NoError(aiService.Init(ctx, withTestModel(response), WithAgent(agent)))

	service := NewEvaluationService(database, aiService)
	runner, runnerErr := service.MakeRunner(evaluationScenario{
		agentName: agent.def.Name,
		status:    ai.ScoreStatusPass.String(),
	})
	s.Require().NoError(runnerErr)

	report := runner.RunEvaluation(ctx)
	s.True(report.Complete)
	s.True(report.Passed)
	s.Require().NotNil(report.Result)
	s.Equal("hello", report.Result.Response.Text())
	s.Len(report.Scores, 2)
}

func (s *EvaluationServiceSuite) TestReportsJudgeErrorsAtJudgeStage() {
	ctx := s.SeedTenantContext()
	database := s.CreateTestDatabase()
	response := &ai.ModelResponse{Message: ai.NewModelTextMessage("hello")}
	agent := makeTestAgent[testAgentState](ai.NewUserTextMessage("say hello"))
	agent.def.Model = "test/model"
	aiService := NewAiService(s.Config())
	s.Require().NoError(aiService.Init(ctx, withTestModel(response), WithAgent(agent)))
	expectedErr := errors.New("judge unavailable")

	service := NewEvaluationService(database, aiService)
	runner, runnerErr := service.MakeRunner(evaluationScenario{
		agentName: agent.def.Name,
		judgeErr:  expectedErr,
	})
	s.Require().NoError(runnerErr)

	report := runner.RunEvaluation(ctx)
	s.False(report.Complete)
	s.Equal("judge", report.FailureStage)
}
