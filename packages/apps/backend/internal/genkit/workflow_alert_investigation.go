package genkit

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/firebase/genkit/go/genkit"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/pkg/agents"
)

type AlertInvestigationAgent struct {
	alerts rez.AlertService
}

func NewAlertInvestigationAgent(alerts rez.AlertService) *AlertInvestigationAgent {
	return &AlertInvestigationAgent{alerts: alerts}
}

func (a *AlertInvestigationAgent) workflow() agents.Workflow[agents.AlertInvestigationState, agents.AlertInvestigationOutput] {
	return agents.WorkflowAlertInvestigation
}

func (a *AlertInvestigationAgent) validateInput(input []byte) error {
	return nil
}

func (a *AlertInvestigationAgent) makeInitialState(run *ent.AgentRun) (*aix.SessionState[agents.AlertInvestigationState], error) {
	if validErr := a.validateInput(run.Input); validErr != nil {
		return nil, validErr
	}
	alertId, idErr := run.Edges.GetSubjectEntityId("alert")
	if idErr != nil {
		return nil, fmt.Errorf("id error: %w", idErr)
	}
	initial := &aix.SessionState[agents.AlertInvestigationState]{
		SessionID: run.ID.String(),
		Messages:  []*ai.Message{ai.NewUserTextMessage("foo bar")},
		Custom:    agents.AlertInvestigationState{AlertID: alertId},
	}
	return initial, nil
}

func (a *AlertInvestigationAgent) agentFunc(g *genkit.Genkit) aix.AgentFunc[agents.AlertInvestigationState] {
	return func(ctx context.Context, resp aix.Responder, sr *aix.SessionRunner[agents.AlertInvestigationState]) (*aix.AgentResult, error) {
		alertId := sr.Custom().AlertID
		fmt.Printf("alert id: %+v\n", alertId)
		alrt, alrtErr := a.alerts.GetAlert(ctx, uuid.Nil)
		if alrtErr != nil {
			return nil, fmt.Errorf("get alert: %w", alrtErr)
		}

		slog.DebugContext(ctx, "agent alert investigation", "title", alrt.Title)
		_ = &agents.AlertInvestigationOutput{}

		return nil, fmt.Errorf("not implemented")
	}
}

type alertInvestigationSynthesis struct {
	Summary         string   `json:"summary"`
	LikelyCause     string   `json:"likelyCause"`
	AffectedSystems []string `json:"affectedSystems"`
	SuggestedChecks []string `json:"suggestedChecks"`
	RecommendedNext string   `json:"recommendedNext"`
	Limitations     []string `json:"limitations"`
	Confidence      string   `json:"confidence"`
}

var alertInvestigationInstruction = strings.TrimSpace(`
You are Rezible's alert investigation agent.
Use only the supplied JSON context. Produce concise JSON with this schema:
{
  "summary": "short responder-oriented synthesis",
  "likelyCause": "best current hypothesis from supplied evidence",
  "affectedSystems": ["system name"],
  "suggestedChecks": ["short verification step"],
  "recommendedNext": "monitor|declare_incident|attach_to_existing_incident|escalate",
  "limitations": ["known uncertainty"],
  "confidence": "low|medium|high"
}
Do not invent systems, incidents, alerts, or evidence that are not present in the context.
`)
