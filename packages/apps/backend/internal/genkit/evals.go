package genkit

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/execution"
)

type EvaluationReport struct {
	RunID                uuid.UUID                    `json:"runId"`
	Scenario             rezai.EvalScenarioDefinition `json:"scenario"`
	StartedAt            time.Time                    `json:"startedAt"`
	DurationMilliseconds int64                        `json:"durationMilliseconds"`
	Complete             bool                         `json:"complete"`
	Passed               bool                         `json:"passed"`
	FailureStage         string                       `json:"failureStage,omitempty"`
	Error                string                       `json:"error,omitempty"`
	Result               *EvaluationResult            `json:"result,omitempty"`
	Scores               []ai.Score                   `json:"scores,omitempty"`
}

type EvaluationResult struct {
	DurationMilliseconds int64                 `json:"durationMilliseconds"`
	FinishReason         aix.AgentFinishReason `json:"finishReason,omitempty"`
	Response             *ai.Message           `json:"response,omitempty"`
	Messages             []*ai.Message         `json:"messages,omitempty"`
	Artifacts            []*aix.Artifact       `json:"artifacts,omitempty"`
	Error                string                `json:"error,omitempty"`
}

type EvaluationService struct {
	db rez.Database
	ai rez.AiService
}

type EvaluationRunner struct {
	db       rez.Database
	ai       rez.AiService
	scenario rezai.EvalScenario

	report     *EvaluationReport
	agent      *rez.AiAgentConfig
	session    *ent.AgentSession
	turn       *ent.AgentTurn
	invocation *rez.AiAgentInvocationResult
}

func NewEvaluationService(db rez.Database, ai rez.AiService) *EvaluationService {
	return &EvaluationService{db: db, ai: ai}
}

func (s *EvaluationService) MakeRunner(scenario rezai.EvalScenario) (*EvaluationRunner, error) {
	if scenario == nil {
		return nil, fmt.Errorf("evaluation scenario is required")
	}

	scd := scenario.Definition()
	runner := &EvaluationRunner{
		db:       s.db,
		ai:       s.ai,
		scenario: scenario,
		report: &EvaluationReport{
			RunID:     uuid.New(),
			StartedAt: time.Now().UTC(),
			Scenario:  scd,
		},
	}
	for _, agent := range s.ai.GetAgents() {
		if agent.Name == scd.AgentName {
			runner.agent = &agent
			break
		}
	}
	if runner.agent == nil {
		return nil, fmt.Errorf("agent %s not found", scd.AgentName)
	}
	return runner, nil
}

func (r *EvaluationRunner) RunEvaluation(ctx context.Context) EvaluationReport {
	tenantCtx, seedErr := r.seedEvaluationSession(ctx)
	if seedErr != nil {
		return r.report.fail("database", seedErr)
	}
	ctx = tenantCtx

	input, inputErr := r.ai.MakeInitialAgentTurnInput(ctx, r.session)
	if inputErr != nil {
		return r.report.fail("prepare", inputErr)
	}

	if invokeErr := r.invokeAgent(ctx, input); invokeErr != nil {
		return r.report.fail("invoke", invokeErr)
	}

	r.report.addExecutionScore(r.invocation)
	if judgeErr := r.judgeScenario(ctx); judgeErr != nil {
		return r.report.fail("judge", judgeErr)
	}

	return r.report.complete()
}

func (r *EvaluationRunner) seedEvaluationSession(ctx context.Context) (context.Context, error) {
	client := r.db.Client(ctx)
	tenant, tenantErr := client.Tenant.Create().Save(execution.NewSystemContext(ctx))
	if tenantErr != nil {
		return nil, fmt.Errorf("create evaluation tenant: %w", tenantErr)
	}
	ctx = execution.NewTenantContext(ctx, tenant.ID)

	seed, seedErr := r.scenario.Seed(ctx, client)
	if seedErr != nil {
		return nil, seedErr
	}
	if seed.Input == nil {
		return nil, fmt.Errorf("scenario returned nil agent input")
	}

	inputJSON, inputJSONErr := json.Marshal(seed.Input)
	if inputJSONErr != nil {
		return nil, fmt.Errorf("marshal agent input: %w", inputJSONErr)
	}
	if _, validationErr := r.ai.ValidateAgentSessionInput(r.agent.Name, inputJSON); validationErr != nil {
		return nil, fmt.Errorf("validate agent input: %w", validationErr)
	}

	r.session = &ent.AgentSession{
		ID:               uuid.New(),
		TenantID:         tenant.ID,
		AgentName:        r.agent.Name,
		Input:            inputJSON,
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

func (r *EvaluationRunner) invokeAgent(ctx context.Context, input *rez.AiAgentTurnInput) error {
	startedAt := time.Now()
	invocation, invocationErr := r.ai.InvokeAgentTurn(ctx, rez.InvokeAgentTurnParams{
		Session: r.session,
		Turn:    r.turn,
		Input:   input,
	})
	r.invocation = invocation

	result := &EvaluationResult{
		DurationMilliseconds: time.Since(startedAt).Milliseconds(),
	}
	if invocation == nil {
		if invocationErr == nil {
			invocationErr = fmt.Errorf("agent returned no invocation result")
		}
	} else {
		result.FinishReason = invocation.FinishReason
		result.Response = invocation.Response
		result.Messages = invocation.State.Messages
		result.Artifacts = invocation.State.Artifacts
		if invocation.Error != nil {
			result.Error = invocation.Error.Error()
		}
	}
	r.report.Result = result

	return invocationErr
}

func (r *EvaluationReport) addExecutionScore(invocation *rez.AiAgentInvocationResult) {
	executionPassed := invocation.Error == nil && invocation.FinishReason != aix.AgentFinishReasonFailed
	executionStatus := ai.ScoreStatusFail.String()
	executionDetail := "agent execution completed"
	if executionPassed {
		executionStatus = ai.ScoreStatusPass.String()
	} else if invocation.Error != nil {
		executionDetail = invocation.Error.Error()
	} else {
		executionDetail = "agent reported a failed finish reason"
	}

	r.Scores = append(r.Scores, ai.Score{
		Id:      "agent_execution",
		Score:   executionPassed,
		Status:  executionStatus,
		Details: map[string]any{"detail": executionDetail},
	})
}

func (r *EvaluationRunner) judgeScenario(ctx context.Context) error {
	scores, judgeErr := r.scenario.Judge(ctx, r.db.Client(ctx), r.invocation)
	if judgeErr != nil {
		return judgeErr
	}
	if len(scores) == 0 {
		return fmt.Errorf("scenario returned no scores")
	}

	seen := mapset.NewSet("agent_execution")
	for _, score := range scores {
		if strings.TrimSpace(score.Id) == "" {
			return fmt.Errorf("scenario returned a score without an ID")
		}
		if seen.Contains(score.Id) {
			return fmt.Errorf("scenario returned duplicate score ID %q", score.Id)
		}
		if score.Status != ai.ScoreStatusPass.String() && score.Status != ai.ScoreStatusFail.String() {
			return fmt.Errorf("scenario score %q has invalid status %q", score.Id, score.Status)
		}
		seen.Add(score.Id)
	}
	r.report.Scores = append(r.report.Scores, scores...)
	return nil
}

func (r *EvaluationReport) complete() EvaluationReport {
	r.Complete = true
	r.Passed = true
	for _, score := range r.Scores {
		if score.Status != ai.ScoreStatusPass.String() {
			r.Passed = false
			break
		}
	}
	r.DurationMilliseconds = time.Since(r.StartedAt).Milliseconds()
	return *r
}

func (r *EvaluationReport) Write(reportWriter io.Writer) error {
	if reportWriter == nil {
		reportWriter = io.Discard
	}
	encoder := json.NewEncoder(reportWriter)
	encoder.SetIndent("", "  ")
	if reportErr := encoder.Encode(r); reportErr != nil {
		return fmt.Errorf("encode evaluation report: %w", reportErr)
	}
	return nil
}

func (r *EvaluationReport) fail(stage string, runErr error) EvaluationReport {
	r.DurationMilliseconds = time.Since(r.StartedAt).Milliseconds()
	r.FailureStage = stage
	r.Error = runErr.Error()
	return *r
}
