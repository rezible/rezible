package genkit

import (
	"context"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	rez "github.com/rezible/rezible"
	rezai "github.com/rezible/rezible/pkg/ai"
)

type (
	AlertsAgent struct {
		alerts rez.AlertService
	}
)

func NewAlertsAgent(alerts rez.AlertService) *AlertsAgent {
	return &AlertsAgent{alerts: alerts}
}

func (a *AlertsAgent) agentDefinition() rezai.AlertsAgentDefinition {
	return rezai.AlertsAgent
}

func (a *AlertsAgent) makeInitialTurnInput(ctx context.Context, input rezai.AlertAgentInput) (*rez.AgentTurnInput, error) {
	inst, instErr := a.alerts.GetAlertInstance(ctx, input.AlertID)
	if instErr != nil {
		return nil, fmt.Errorf("get alert instance: %w", instErr)
	}

	alrt, alrtErr := inst.Edges.AlertOrErr()
	if alrtErr != nil {
		return nil, fmt.Errorf("get alert: %w", alrtErr)
	}

	msgText := fmt.Sprintf(`You are an ai agent built to help software engineering teams investigate & triage alerts.
Title: %s
Description: %s
Definition: %s`, alrt.Title, alrt.Description, alrt.Definition)

	return &rez.AgentTurnInput{Message: ai.NewUserTextMessage(msgText)}, nil
}

func (a *AlertsAgent) transformState(ctx context.Context, state *aix.SessionState[rezai.AlertAgentState]) (*aix.SessionState[rezai.AlertAgentState], error) {
	return state, nil
}

func (a *AlertsAgent) transformStreamChunk(ctx context.Context, chunk *aix.AgentStreamChunk) (*aix.AgentStreamChunk, error) {
	return chunk, nil
}
