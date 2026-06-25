package eino

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/pkg/agents"
)

type AgentRunner struct {
	cfg          rez.AiConfig
	incidents    rez.IncidentService
	alerts       rez.AlertService
	modelFactory ModelProvider
}

func NewAgentWorkflowRunner(cfg rez.Config, incidents rez.IncidentService, alerts rez.AlertService) (*AgentRunner, error) {
	r := &AgentRunner{
		cfg:          cfg.AI,
		incidents:    incidents,
		alerts:       alerts,
		modelFactory: newChatModelProvider(cfg.AI),
	}
	return r, nil
}

func (r *AgentRunner) RegisterWorkflows(reg *agents.WorkflowRegistry) {
	agents.RegisterWorkflowRunner(reg, agents.WorkflowAlertInvestigation, r.runAlertInvestigation)
	agents.RegisterWorkflowRunner(reg, agents.WorkflowIncidentTriage, r.runIncidentTriage)
}

type workflowRunner[O any] interface {
	run(context.Context) (*O, error)
}

func runWorkflow[O any](ctx context.Context, r workflowRunner[O], md map[string]string) (*agents.WorkflowRunnerResult[O], error) {
	out, err := r.run(ctx)
	return &agents.WorkflowRunnerResult[O]{Output: out, Error: err, Metadata: md}, nil
}

func (r *AgentRunner) runAlertInvestigation(
	ctx context.Context,
	rc agents.WorkflowRunContext,
	input agents.AlertInvestigationInput,
) (*agents.WorkflowRunnerResult[agents.AlertInvestigationOutput], error) {
	alertID, idErr := rc.GetSubjectEntityId("alert")
	if idErr != nil {
		return nil, fmt.Errorf("alert investigation has no alert subject: %w", idErr)
	}
	workflow := &alertInvestigationWorkflow{
		alerts:       r.alerts,
		modelFactory: r.modelFactory,
		input:        input,
		alertID:      alertID,
	}
	return runWorkflow(ctx, workflow, nil)
}

func (r *AgentRunner) runIncidentTriage(
	ctx context.Context,
	rc agents.WorkflowRunContext,
	input agents.IncidentTriageInput,
) (*agents.WorkflowRunnerResult[agents.IncidentTriageOutput], error) {
	incID, idErr := rc.GetSubjectEntityId("incident")
	if idErr != nil {
		return nil, fmt.Errorf("no incident subject: %w", idErr)
	}
	workflow := &incidentTriageWorkflow{
		incidents:    r.incidents,
		modelFactory: r.modelFactory,
		input:        input,
		incidentID:   incID,
	}
	return runWorkflow(ctx, workflow, nil)
}
