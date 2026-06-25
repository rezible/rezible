package agents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/rezible/rezible/ent"
)

type (
	WorkflowRunnerFunc[Input any, Output any] func(context.Context, WorkflowRunContext, Input) (*WorkflowRunnerResult[Output], error)

	RunWorkflowResult struct {
		EncodedOutput []byte
		Error         error
		Metadata      map[string]string
	}

	RunWorkflowFunc func(context.Context, *ent.AgentTask, *ent.AgentRun) RunWorkflowResult
)

type WorkflowRegistry struct {
	runners   map[string]RunWorkflowFunc
	runnersMu sync.RWMutex
}

func NewWorkflowRegistry() *WorkflowRegistry {
	return &WorkflowRegistry{runners: make(map[string]RunWorkflowFunc)}
}

func RegisterWorkflowRunner[I any, O any](r *WorkflowRegistry, w workflowDefinition[I, O], fn WorkflowRunnerFunc[I, O]) {
	if r == nil {
		panic("nil workflow registry")
	}
	r.runnersMu.Lock()
	defer r.runnersMu.Unlock()
	if _, exists := r.runners[w.name]; exists {
		panic("agent workflow runner already registered")
	}

	r.runners[w.name] = func(ctx context.Context, task *ent.AgentTask, run *ent.AgentRun) RunWorkflowResult {
		var inp I
		if jsonErr := json.Unmarshal(task.Input, &inp); jsonErr != nil {
			return RunWorkflowResult{Error: fmt.Errorf("input: %w", jsonErr)}
		}
		out, runErr := fn(ctx, WorkflowRunContext{Task: task, Run: run}, inp)
		res := RunWorkflowResult{Error: runErr, Metadata: out.Metadata}
		if out != nil {
			// TODO: validate output

			encOut, jsonErr := json.Marshal(out)
			if jsonErr != nil {
				res.Error = errors.Join(res.Error, fmt.Errorf("encode output: %w", jsonErr))
			} else {
				res.EncodedOutput = encOut
			}
		}
		return res
	}
}

func (r *WorkflowRegistry) ValidateTaskInput(workflow string, input map[string]any) error {
	// TODO: validate
	return nil
}

func (r *WorkflowRegistry) ValidateAndEncodeTaskInput(workflow string, input map[string]any) ([]byte, error) {
	if validErr := r.ValidateTaskInput(workflow, input); validErr != nil {
		return nil, fmt.Errorf("validation: %w", validErr)
	}
	enc, jsonErr := json.Marshal(input)
	if jsonErr != nil {
		return nil, fmt.Errorf("encode: %w", jsonErr)
	}
	return enc, nil
}

func (r *WorkflowRegistry) GetWorkflowRunner(workflow string) (RunWorkflowFunc, error) {
	if r == nil {
		panic("nil workflow registry")
	}
	r.runnersMu.RLock()
	defer r.runnersMu.RUnlock()
	runnerFn, fnOk := r.runners[workflow]
	if !fnOk {
		return nil, fmt.Errorf("no registered workflow runner for type: %s", workflow)
	}
	return runnerFn, nil
}
