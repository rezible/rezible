package eino

import (
	"context"
	"fmt"

	"github.com/rezible/rezible/ent/agentcasestep"
	"github.com/rezible/rezible/ent/agentrun"

	"github.com/rezible/rezible/ent"
)

type stubWorkflow struct {
	kind agentrun.WorkflowKind
}

func (w stubWorkflow) Kind() agentrun.WorkflowKind {
	return w.kind
}

func (w stubWorkflow) Validate(_ context.Context, run *ent.AgentRun) error {
	if run.WorkflowKind != w.kind {
		return fmt.Errorf("workflow/run kind mismatch: %s != %s", run.WorkflowKind, w.kind)
	}
	return nil
}

func (w stubWorkflow) Run(_ context.Context, run *ent.AgentRun) (*AgentWorkflowResult, error) {
	return &AgentWorkflowResult{
		Summary: "Workflow not implemented.",
		Steps: []AgentCaseStep{
			{
				Kind:    agentcasestep.KindConclusion,
				Title:   "Workflow scaffold",
				Summary: "Workflow not implemented.",
				Output: map[string]any{
					"status": "not_implemented",
				},
			},
		},
	}, nil
}
