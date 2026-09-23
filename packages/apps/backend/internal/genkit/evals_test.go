package genkit

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/firebase/genkit/go/ai"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/agentturn"
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

func (s testEvalScenario) Seed(ctx context.Context, client *ent.Client) (rezai.EvalScenarioSeed, error) {
	createSession := client.AgentSession.Create().
		SetAgentName(s.agentName).
		SetInput([]byte(`{}`))
	session, sessionErr := createSession.Save(ctx)
	if sessionErr != nil {
		return rezai.EvalScenarioSeed{}, fmt.Errorf("create session: %w", sessionErr)
	}
	createTurn := client.AgentTurn.Create().
		SetID(uuid.New()).
		SetAgentSessionID(session.ID).
		SetSequence(1).
		SetRiverJobID(0).
		SetStatus(agentturn.StatusRunning)
	turn, turnErr := createTurn.Save(ctx)
	if turnErr != nil {
		return rezai.EvalScenarioSeed{}, fmt.Errorf("create turn: %w", turnErr)
	}
	return rezai.EvalScenarioSeed{Session: session, Turn: turn}, nil
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

func (s *EvaluationServiceSuite) makeRuntime(ctx context.Context, opts ...AiRuntimeOption) *AiRuntime {
	cfg := s.Config()
	runtime := NewAiRuntime(cfg)
	baseOpts := []AiRuntimeOption{WithDevEvals()}
	s.Require().NoError(runtime.Init(ctx, append(baseOpts, opts...)...))
	return runtime
}

func (s *EvaluationServiceSuite) makeEvalService(tdb rez.Database, runtime *AiRuntime, scenarios ...rezai.EvalScenario) *EvaluationService {
	workflowBuilder := NewWorkflowBuilder(runtime, &testWorkflowRunner{})
	service, serviceErr := MakeEvaluationService(tdb, runtime, workflowBuilder)
	s.Require().NoError(serviceErr)
	for _, sc := range scenarios {
		s.Require().NoError(service.RegisterScenario(sc))
	}
	return service
}

func (s *EvaluationServiceSuite) runEval(ctx context.Context, tdb rez.Database, runtime *AiRuntime, scenario testEvalScenario) rezai.EvalScenarioRunResult {
	service := s.makeEvalService(tdb, runtime, scenario)

	result, runErr := service.RunScenario(ctx, scenario.Definition().Name)
	s.Require().NoError(runErr)
	return result
}

func (s *EvaluationServiceSuite) TestRunsAgentAndProducesPassingReport() {
	ctx := s.SeedTenantContext()
	response := &ai.ModelResponse{
		Message:      ai.NewModelTextMessage("hello"),
		FinishReason: ai.FinishReasonStop,
	}
	testModel := makeTestOutputModel(response)
	agent := makeTestAgent[testAgentState](ai.NewUserTextMessage("say hello"))
	agent.def.Model = testModel.Name
	runtime := s.makeRuntime(ctx, WithDefinedModel(testModel), WithAgent(agent))

	tdb := s.CreateTestDatabase()
	scenario := testEvalScenario{
		agentName: agent.def.Name,
		passed:    true,
	}
	result := s.runEval(ctx, tdb, runtime, scenario)

	s.Equal(rezai.EvalRunStatusPassed, result.Status)
	s.Require().NotNil(result.Execution)
	s.True(result.Execution.Succeeded)
	s.Equal(testModel.Name, result.Agent.Model)
	s.Len(result.Checks, 1)
}

func (s *EvaluationServiceSuite) TestReportsGradeErrorsAtGradeStage() {
	ctx := s.SeedTenantContext()
	response := &ai.ModelResponse{
		Message:      ai.NewModelTextMessage("hello"),
		FinishReason: ai.FinishReasonStop,
	}
	testModel := makeTestOutputModel(response)
	agent := makeTestAgent[testAgentState](ai.NewUserTextMessage("say hello"))
	agent.def.Model = testModel.Name
	runtime := s.makeRuntime(ctx, WithDefinedModel(testModel), WithAgent(agent))
	expectedErr := errors.New("grading unavailable")

	tdb := s.CreateTestDatabase()
	scenario := testEvalScenario{
		agentName: agent.def.Name,
		gradeErr:  expectedErr,
	}
	result := s.runEval(ctx, tdb, runtime, scenario)
	s.Equal(rezai.EvalRunStatusError, result.Status)
	s.Require().NotNil(result.Error)
	s.Equal(rezai.EvalRunStageGrade, result.Error.Stage)
	s.Require().NotNil(result.Execution)
	s.True(result.Execution.Succeeded)
}
