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
	return &rez.AgentTurnInput{
		Message: ai.NewUserTextMessage(fmt.Sprintf("Investigate alert instance %s.", input.AlertID)),
	}, nil
}

func (a *AlertsAgent) makeInitialContextSeed(ctx context.Context, input rezai.AlertAgentInput) (string, error) {
	inst, instErr := a.alerts.GetAlertInstance(ctx, input.AlertID)
	if instErr != nil {
		return "", fmt.Errorf("get alert instance: %w", instErr)
	}

	alrt, alrtErr := inst.Edges.AlertOrErr()
	if alrtErr != nil {
		return "", fmt.Errorf("get alert: %w", alrtErr)
	}

	seed := fmt.Sprintf(`Alert instance ID: %s
Title: %s
Description: %s
Definition: %s`, input.AlertID, alrt.Title, alrt.Description, alrt.Definition)
	if alrt.KnowledgeEntityID != nil {
		seed += fmt.Sprintf("\nKnowledge graph entity ID: %s", *alrt.KnowledgeEntityID)
	}
	return seed, nil
}

func (a *AlertsAgent) transformState(ctx context.Context, state *aix.SessionState[rezai.AlertAgentState]) (*aix.SessionState[rezai.AlertAgentState], error) {
	return state, nil
}

func (a *AlertsAgent) transformStreamChunk(ctx context.Context, chunk *aix.AgentStreamChunk) (*aix.AgentStreamChunk, error) {
	return chunk, nil
}
