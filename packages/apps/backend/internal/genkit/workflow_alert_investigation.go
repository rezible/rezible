package genkit

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/pkg/agents"
)

type alertInvestigationAgent struct {
	alerts rez.AlertService
}

func newAlertInvestigationAgent(alerts rez.AlertService) *alertInvestigationAgent {
	return &alertInvestigationAgent{alerts: alerts}
}

func (a *alertInvestigationAgent) workflow() agents.Workflow[agents.AlertInvestigationState, agents.AlertInvestigationOutput] {
	return agents.WorkflowAlertInvestigation
}

func (a *alertInvestigationAgent) validateInput(input []byte) error {
	if input == nil || len(input) == 0 {
		return fmt.Errorf("empty input")
	}
	return nil
}

func (a *alertInvestigationAgent) makeInitialMessage(run *ent.AgentRun) (*ai.Message, error) {
	if validErr := a.validateInput(run.Input); validErr != nil {
		return nil, validErr
	}
	inp := ai.NewUserTextMessage("foo bar")
	return inp, nil
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

func (a *alertInvestigationAgent) agentFunc() aix.AgentFunc[agents.AlertInvestigationState] {
	return func(ctx context.Context, resp aix.Responder, sess *aix.SessionRunner[agents.AlertInvestigationState]) (*aix.AgentResult, error) {
		alrt, alrtErr := a.alerts.GetAlert(ctx, uuid.Nil)
		if alrtErr != nil {
			return nil, fmt.Errorf("get alert: %w", alrtErr)
		}

		slog.DebugContext(ctx, "agent alert investigation", "title", alrt.Title)
		_ = &agents.AlertInvestigationOutput{}

		return nil, fmt.Errorf("not implemented")
	}
}
