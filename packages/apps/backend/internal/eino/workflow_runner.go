package eino

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/agents"
	"github.com/rezible/rezible/ent"
)

type WorkflowRunner struct {
	cfg          rez.AiConfig
	incidents    rez.IncidentService
	alerts       rez.AlertService
	modelFactory ModelProvider
}

func NewWorkflowRunner(cfg rez.Config, reg *agents.WorkflowRegistry, incidents rez.IncidentService, alerts rez.AlertService) (*WorkflowRunner, error) {
	r := &WorkflowRunner{
		cfg:          cfg.AI,
		incidents:    incidents,
		alerts:       alerts,
		modelFactory: newChatModelProvider(cfg.AI),
	}

	reg.Register(agents.WorkflowKindIncidentContextPack, r.runIncidentContextPack)
	reg.Register(agents.WorkflowKindAlertInvestigation, r.runAlertInvestigation)

	return r, nil
}

func (r *WorkflowRunner) runIncidentContextPack(ctx context.Context, task *ent.AgentTask, run *ent.AgentRun) (*agents.RunResult, error) {
	input, inputErr := agents.DecodeInput[agents.IncidentContextPackInput](task.WorkflowInput)
	if inputErr != nil {
		return nil, fmt.Errorf("decode incident context pack input: %w", inputErr)
	}
	incidentID := agents.FindSubjectRefID(input.Subjects, "incident")
	if incidentID == uuid.Nil {
		return nil, fmt.Errorf("incident context pack requires incident subject")
	}
	workflow := incidentContextPackWorkflow{
		incidents:    r.incidents,
		modelFactory: r.modelFactory,
		input:        *input,
		incidentID:   incidentID,
	}
	return workflow.Run(ctx)
}

func (r *WorkflowRunner) runAlertInvestigation(ctx context.Context, task *ent.AgentTask, run *ent.AgentRun) (*agents.RunResult, error) {
	input, inputErr := agents.DecodeInput[agents.AlertInvestigationInput](task.WorkflowInput)
	if inputErr != nil {
		return nil, fmt.Errorf("decode alert investigation input: %w", inputErr)
	}
	alertID := agents.FindSubjectRefID(input.Subjects, "alert")
	if alertID == uuid.Nil {
		return nil, fmt.Errorf("alert investigation requires alert subject")
	}
	workflow := alertInvestigationWorkflow{
		alerts:       r.alerts,
		modelFactory: r.modelFactory,
		aiEnabled:    r.cfg.Enabled,
		input:        *input,
		alertID:      alertID,
	}
	return workflow.Run(ctx)
}

func encodeWorkflowContext(ctx *agents.WorkflowContext) (string, error) {
	payload := map[string]any{
		"generatedAt": ctx.GeneratedAt,
		"context":     ctx.Context,
		"items":       ctx.Items,
	}
	bytes, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
