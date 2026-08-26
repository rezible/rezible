package genkit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/core/api"
	"github.com/firebase/genkit/go/genkit"
	"github.com/google/uuid"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/firebase/genkit/go/core/tracing"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/ai/evals"
	"github.com/rezible/rezible/pkg/execution"
)

type EvaluationService struct {
	db rez.Database
	ai *AiService

	evaluateScenarioFlow *core.Flow[EvaluationFlowInput, rezai.EvalScenarioRunResult, struct{}]
	evaluatorAction      *ai.EvaluatorAction
}

var scenarioChecksActionName = api.NewName("rezible", "scenario_checks")

func NewEvaluationService(db rez.Database, aiSvc *AiService) *EvaluationService {
	s := &EvaluationService{db: db, ai: aiSvc}

	s.evaluateScenarioFlow = genkit.DefineFlow(aiSvc.gk, "evaluate_agent_scenario", s.runEvalScenarioFlow)
	options := &ai.EvaluatorOptions{
		DisplayName: "Scenario Check",
		Definition:  "Presents scenario grading checks as Genkit metrics.",
		IsBilled:    false,
	}
	s.evaluatorAction = genkit.DefineEvaluatorAction(aiSvc.gk, scenarioChecksActionName, options, evaluateScenarioChecks)

	return s
}

func (s *EvaluationService) RunScenarioEvalFlow(ctx context.Context, name string) (*rezai.EvalScenarioRunResult, error) {
	input := EvaluationFlowInput{ScenarioName: name}
	fmt.Printf("running scenario eval %s\n", name)
	output, flowErr := s.evaluateScenarioFlow.Run(ctx, input)
	if flowErr != nil {
		return nil, fmt.Errorf("run flow: %w", flowErr)
	}
	fmt.Printf("output %+v\n", output)
	scenarioExample := &ai.Example{
		TestCaseId: uuid.NewString(),
		Input:      input,
		Output:     output,
	}
	evalReq := &ai.EvaluatorRequest{
		Dataset:      []*ai.Example{scenarioExample},
		EvaluationId: uuid.New().String(),
		Options:      struct{}{},
	}
	evalResp, evalErr := s.evaluatorAction.Evaluate(ctx, evalReq)
	if evalErr != nil {
		return nil, fmt.Errorf("evaluator action: %w", evalErr)
	}
	fmt.Printf("evaluator resp: %+v\n", evalResp)
	return &output, nil
}

func (s *EvaluationService) RunNamedScenario(ctx context.Context, name string) (*rezai.EvalScenarioRunResult, error) {
	scenario, lookupErr := evals.Lookup(strings.TrimSpace(name))
	if lookupErr != nil {
		return nil, lookupErr
	}
	result := s.RunScenario(ctx, scenario)
	return &result, nil
}

func (s *EvaluationService) RunScenario(ctx context.Context, scenario rezai.EvalScenario) rezai.EvalScenarioRunResult {
	return s.executeEvaluationRun(ctx, &evaluationRun{service: s, scenario: scenario})
}

type EvaluationFlowInput struct {
	ScenarioName string `json:"scenario_name" jsonschema:"description=Registered evaluation scenario name,minLength=1"`
}

func (s *EvaluationService) runEvalScenarioFlow(ctx context.Context, input EvaluationFlowInput) (rezai.EvalScenarioRunResult, error) {
	res, resErr := s.RunNamedScenario(ctx, input.ScenarioName)
	if resErr != nil {
		return rezai.EvalScenarioRunResult{}, resErr
	}
	return *res, nil
}

func evaluateScenarioChecks(ctx context.Context, req *ai.EvaluatorCallbackRequest, cfg struct{}) (*ai.EvaluatorCallbackResponse, error) {
	if req == nil || req.Input.Output == nil {
		return nil, fmt.Errorf("evaluation output is required")
	}
	e := &scenarioChecksEvaluator{}
	return e.evaluate(req.Input)
}

type evaluationRun struct {
	service    *EvaluationService
	scenario   rezai.EvalScenario
	result     rezai.EvalScenarioRunResult
	session    *ent.AgentSession
	turn       *ent.AgentTurn
	invocation *rez.AiAgentInvocationResult
}

func (s *EvaluationService) executeEvaluationRun(ctx context.Context, r *evaluationRun) rezai.EvalScenarioRunResult {
	r.result = rezai.EvalScenarioRunResult{
		Status: rezai.EvalRunStatusError,
	}
	if r.scenario == nil {
		return r.fail(rezai.EvalRunStageSetup, "evaluation scenario is required")
	}
	r.result.Scenario = r.scenario.Definition()
	d := r.result.Scenario
	if strings.TrimSpace(d.Name) == "" || strings.TrimSpace(d.Description) == "" || strings.TrimSpace(d.AgentName) == "" {
		return r.fail(rezai.EvalRunStageSetup, "scenario definition requires name, description, and agent name")
	}
	found := false
	for _, agent := range s.ai.GetAgents() {
		if agent.Name == d.AgentName {
			r.result.Agent = agent
			found = true
			break
		}
	}
	if !found {
		return r.fail(rezai.EvalRunStageSetup, fmt.Sprintf("agent %q not found", d.AgentName))
	}

	seedCtx, seedErr := traceStep(ctx, "seed", func(ctx context.Context) (context.Context, error) {
		return r.seed(ctx)
	})
	if seedErr != nil {
		return r.fail(rezai.EvalRunStageSeed, seedErr.Error())
	}
	ctx = seedCtx

	input, prepareErr := traceStep(ctx, "prepare", func(ctx context.Context) (*rez.AiAgentTurnInput, error) {
		return s.ai.MakeInitialAgentTurnInput(ctx, r.session)
	})
	if prepareErr != nil || input == nil {
		if prepareErr == nil {
			prepareErr = fmt.Errorf("agent returned no initial turn input")
		}
		return r.fail(rezai.EvalRunStagePrepare, prepareErr.Error())
	}

	started := time.Now()
	invocation, invokeErr := traceStep(ctx, "invoke", func(ctx context.Context) (*rez.AiAgentInvocationResult, error) {
		return s.ai.InvokeAgentTurn(ctx, rez.InvokeAgentTurnParams{Session: r.session, Turn: r.turn, Input: input})
	})
	r.invocation = invocation
	r.result.Execution = r.summarizeExecution(invocation, invokeErr, time.Since(started).Milliseconds())
	if invokeErr != nil {
		return r.fail(rezai.EvalRunStageInvoke, invokeErr.Error())
	}
	if invocation == nil {
		return r.fail(rezai.EvalRunStageInvoke, "agent returned no invocation result")
	}
	if !r.result.Execution.Succeeded {
		r.result.Status = rezai.EvalRunStatusFailed
		return r.result
	}

	grade, gradeErr := traceStep(ctx, "grade", func(ctx context.Context) (rezai.EvalScenarioGrade, error) {
		return r.scenario.Grade(ctx, s.db.Client(ctx), invocation)
	})
	if gradeErr != nil {
		return r.fail(rezai.EvalRunStageGrade, gradeErr.Error())
	}
	if gradeErr = r.validateGrade(grade); gradeErr != nil {
		return r.fail(rezai.EvalRunStageGrade, gradeErr.Error())
	}
	r.result.Output, r.result.Checks = grade.Output, grade.Checks
	r.result.Status = rezai.EvalRunStatusPassed
	for _, check := range grade.Checks {
		if !check.Passed {
			r.result.Status = rezai.EvalRunStatusFailed
			break
		}
	}
	return r.result
}

func (r *evaluationRun) seed(ctx context.Context) (context.Context, error) {
	client := r.service.db.Client(ctx)
	tenant, tenantErr := client.Tenant.Create().Save(execution.NewSystemContext(ctx))
	if tenantErr != nil {
		return nil, fmt.Errorf("create evaluation tenant: %w", tenantErr)
	}
	ctx = execution.NewTenantContext(ctx, tenant.ID)

	seed, seedErr := r.scenario.Seed(ctx, client)
	if seedErr != nil {
		return nil, seedErr
	} else if seed.Input == nil {
		return nil, fmt.Errorf("scenario returned nil agent input")
	}

	raw, inputErr := json.Marshal(seed.Input)
	if inputErr != nil {
		return nil, fmt.Errorf("marshal agent input: %w", inputErr)
	}
	if _, inputErr = r.service.ai.ValidateAgentSessionInput(r.result.Agent.Name, raw); inputErr != nil {
		return nil, fmt.Errorf("validate agent input: %w", inputErr)
	}

	r.session = &ent.AgentSession{
		ID:               uuid.New(),
		TenantID:         tenant.ID,
		AgentName:        r.result.Agent.Name,
		Input:            raw,
		SystemAnalysisID: seed.SystemAnalysisID,
	}
	r.turn = &ent.AgentTurn{
		ID:             uuid.New(),
		TenantID:       tenant.ID,
		AgentSessionID: r.session.ID,
		Sequence:       1,
	}

	return ctx, nil
}

func (r *evaluationRun) fail(stage rezai.EvalRunStage, message string) rezai.EvalScenarioRunResult {
	r.result.Status = rezai.EvalRunStatusError
	r.result.Output = nil
	r.result.Checks = nil
	r.result.Error = &rezai.EvalRunError{Stage: stage, Message: message}
	return r.result
}

func (r *evaluationRun) summarizeExecution(inv *rez.AiAgentInvocationResult, callErr error, duration int64) *rezai.EvalAgentExecution {
	e := &rezai.EvalAgentExecution{
		DurationMilliseconds: duration,
	}
	if inv != nil {
		e.FinishReason = string(inv.FinishReason)
	}
	switch {
	case callErr != nil:
		e.Error = callErr.Error()
	case inv == nil:
		e.Error = "agent returned no invocation result"
	case inv.Error != nil:
		e.Error = inv.Error.Error()
	case inv.FinishReason != aix.AgentFinishReasonStop:
		e.Error = fmt.Sprintf("agent finished with reason %q", string(inv.FinishReason))
	default:
		e.Succeeded = true
	}
	return e
}

func (r *evaluationRun) validateGrade(g rezai.EvalScenarioGrade) error {
	if len(g.Checks) == 0 {
		return fmt.Errorf("scenario returned no checks")
	}
	if _, jsonErr := json.Marshal(g.Output); jsonErr != nil {
		return fmt.Errorf("serialize grade output: %w", jsonErr)
	}
	seenChecks := mapset.NewSet("scenario_completed", "agent_execution")
	for _, c := range g.Checks {
		id := strings.TrimSpace(c.ID)
		if id == "" {
			return fmt.Errorf("scenario returned a check without an ID")
		}
		if !seenChecks.Add(id) {
			return fmt.Errorf("scenario returned duplicate or reserved check ID %q", id)
		}
		if strings.TrimSpace(c.Summary) == "" {
			return fmt.Errorf("scenario check %q has a blank summary", id)
		}
		if _, expectedErr := json.Marshal(c.Expected); expectedErr != nil {
			return fmt.Errorf("serialize check %q expected value: %w", id, expectedErr)
		}
		if _, observedErr := json.Marshal(c.Observed); observedErr != nil {
			return fmt.Errorf("serialize check %q observed value: %w", id, observedErr)
		}
	}
	return nil
}

func traceStep[I any](ctx context.Context, name string, fn func(context.Context) (I, error)) (I, error) {
	metadata := &tracing.SpanMetadata{
		Name:    name,
		Type:    "action",
		Subtype: "util",
	}
	return tracing.RunInNewSpan(ctx, metadata, struct{}{}, func(ctx context.Context, _ struct{}) (I, error) {
		return fn(ctx)
	})
}

type scenarioChecksEvaluator struct{}

func (e *scenarioChecksEvaluator) evaluate(example ai.Example) (*ai.EvaluatorCallbackResponse, error) {
	raw, marshalErr := json.Marshal(example.Output)
	if marshalErr != nil {
		return nil, fmt.Errorf("marshal evaluation output: %w", marshalErr)
	}

	var result rezai.EvalScenarioRunResult
	if decodeErr := json.Unmarshal(raw, &result); decodeErr != nil {
		return nil, fmt.Errorf("decode evaluation output: %w", decodeErr)
	}

	if validationErr := e.validateScenarioResult(result); validationErr != nil {
		return nil, validationErr
	}

	var evalScores []ai.Score
	addScore := func(id string, passed bool, summary string, expected, observed any) {
		status := ai.ScoreStatusFail.String()
		if passed {
			status = ai.ScoreStatusPass.String()
		}
		details := map[string]any{"summary": summary}
		if expected != nil {
			details["expected"] = expected
		}
		if observed != nil {
			details["observed"] = observed
		}
		evalScores = append(evalScores, ai.Score{Id: id, Score: passed, Status: status, Details: details})
	}

	addScore(
		"scenario_completed",
		result.Status != rezai.EvalRunStatusError,
		"Scenario run completed without a framework error.",
		result.Error,
		result.Status,
	)

	executionPassed := result.Execution != nil && result.Execution.Succeeded
	executionSummary := "Agent execution did not produce a result."
	if result.Execution != nil {
		executionSummary = result.Execution.Error
		if result.Execution.Succeeded {
			executionSummary = "Agent execution succeeded."
		}
	}

	addScore("agent_execution", executionPassed, executionSummary, nil, result.Execution)

	for _, check := range result.Checks {
		addScore(check.ID, check.Passed, check.Summary, check.Expected, check.Observed)
	}

	return &ai.EvaluatorCallbackResponse{
		TestCaseId: example.TestCaseId,
		Evaluation: evalScores,
	}, nil
}

func (e *scenarioChecksEvaluator) validateScenarioResult(result rezai.EvalScenarioRunResult) error {
	if result.Execution != nil {
		if result.Execution.DurationMilliseconds < 0 {
			return fmt.Errorf("execution duration must be non-negative")
		}
		if result.Execution.Succeeded && (result.Execution.FinishReason != "stop" || strings.TrimSpace(result.Execution.Error) != "") {
			return fmt.Errorf("successful execution has invalid finish reason or error")
		}
		if !result.Execution.Succeeded && strings.TrimSpace(result.Execution.Error) == "" {
			return fmt.Errorf("unsuccessful execution requires an error")
		}
	}

	switch result.Status {
	case rezai.EvalRunStatusPassed:
		if result.Error != nil || result.Execution == nil || !result.Execution.Succeeded || len(result.Checks) == 0 {
			return fmt.Errorf("invalid passed result shape")
		}
		for _, check := range result.Checks {
			if !check.Passed {
				return fmt.Errorf("passed result contains failed check %q", check.ID)
			}
		}
	case rezai.EvalRunStatusFailed:
		if result.Error != nil || result.Execution == nil {
			return fmt.Errorf("invalid failed result shape")
		}
		if !result.Execution.Succeeded {
			if len(result.Checks) != 0 {
				return fmt.Errorf("failed execution must not contain checks")
			}
			return nil
		}
		if len(result.Checks) == 0 {
			return fmt.Errorf("graded failure requires checks")
		}
		for _, check := range result.Checks {
			if !check.Passed {
				return nil
			}
		}
		return fmt.Errorf("failed result has no failed check")
	case rezai.EvalRunStatusError:
		if result.Error == nil || len(result.Checks) != 0 || result.Output != nil {
			return fmt.Errorf("invalid error result shape")
		}
		if strings.TrimSpace(result.Error.Message) == "" {
			return fmt.Errorf("error result requires a message")
		}
		switch result.Error.Stage {
		case rezai.EvalRunStageSetup, rezai.EvalRunStageSeed, rezai.EvalRunStagePrepare:
			if result.Execution != nil {
				return fmt.Errorf("%s error must not contain execution", result.Error.Stage)
			}
		case rezai.EvalRunStageInvoke:
			if result.Execution == nil || result.Execution.Succeeded {
				return fmt.Errorf("invoke error requires unsuccessful execution")
			}
		case rezai.EvalRunStageGrade:
			if result.Execution == nil || !result.Execution.Succeeded {
				return fmt.Errorf("grade error requires successful execution")
			}
		default:
			return fmt.Errorf("unknown error stage %q", result.Error.Stage)
		}
	default:
		return fmt.Errorf("unknown evaluation status %q", result.Status)
	}
	return nil
}
