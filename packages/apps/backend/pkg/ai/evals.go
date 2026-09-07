package ai

import (
	"context"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type (
	EvalScenario interface {
		Definition() EvalScenarioDefinition
		Seed(context.Context, *ent.Client) (EvalScenarioSeed, error)
		Grade(context.Context, *ent.Client, *rez.AiAgentInvocationResult) (EvalScenarioGrade, error)
	}

	EvalScenarioRunner interface {
		RunScenario(context.Context, EvalScenario) EvalScenarioRunResult
		RunNamedScenario(context.Context, string) (*EvalScenarioRunResult, error)
	}
)

type (
	EvalScenarioDefinition struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		AgentName   string `json:"agentName"`
	}

	EvalScenarioSeed struct {
		Session *ent.AgentSession
		Turn    *ent.AgentTurn
	}

	EvalScenarioGrade struct {
		Output any         `json:"output,omitempty"`
		Checks []EvalCheck `json:"checks"`
	}

	EvalCheck struct {
		ID       string `json:"id"`
		Passed   bool   `json:"passed"`
		Summary  string `json:"summary"`
		Expected any    `json:"expected,omitempty"`
		Observed any    `json:"observed,omitempty"`
	}

	EvalScenarioRunResult struct {
		Scenario  EvalScenarioDefinition `json:"scenario"`
		Agent     rez.AiAgentConfig      `json:"agent"`
		Status    EvalRunStatus          `json:"status"`
		Execution *EvalAgentExecution    `json:"execution,omitempty"`
		Output    any                    `json:"output,omitempty"`
		Checks    []EvalCheck            `json:"checks,omitempty"`
		Error     *EvalRunError          `json:"error,omitempty"`
	}

	EvalRunStatus string

	EvalAgentExecution struct {
		Succeeded            bool   `json:"succeeded"`
		FinishReason         string `json:"finishReason,omitempty"`
		DurationMilliseconds int64  `json:"durationMilliseconds"`
		Error                string `json:"error,omitempty"`
	}

	EvalRunError struct {
		Stage   EvalRunStage `json:"stage"`
		Message string       `json:"message"`
	}

	EvalRunStage string
)

const (
	EvalRunStatusPassed EvalRunStatus = "passed"
	EvalRunStatusFailed EvalRunStatus = "failed"
	EvalRunStatusError  EvalRunStatus = "error"

	EvalRunStageSetup   EvalRunStage = "setup"
	EvalRunStageSeed    EvalRunStage = "seed"
	EvalRunStagePrepare EvalRunStage = "prepare"
	EvalRunStageInvoke  EvalRunStage = "invoke"
	EvalRunStageGrade   EvalRunStage = "grade"
)
