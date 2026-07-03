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
	rezai "github.com/rezible/rezible/pkg/ai"
)

type AlertInvestigationAgent struct {
	alerts rez.AlertService
}

func NewAlertInvestigationAgent(alerts rez.AlertService) *AlertInvestigationAgent {
	return &AlertInvestigationAgent{alerts: alerts}
}

func (a *AlertInvestigationAgent) definition() rezai.AgentDefinition[rezai.AlertInvestigationState] {
	return rezai.AlertInvestigationAgent
}

func (a *AlertInvestigationAgent) makeInitialState(jsonInput []byte) (*ai.Message, *aix.SessionState[rezai.AlertInvestigationState], error) {
	input, validErr := rezai.AlertInvestigationAgent.CreateInitialState(jsonInput)
	if validErr != nil {
		return nil, nil, validErr
	}
	state := &aix.SessionState[rezai.AlertInvestigationState]{
		Messages: []*ai.Message{ai.NewSystemTextMessage("foo bar")},
		Custom:   rezai.AlertInvestigationState{AlertID: input.AlertID},
	}
	msg := ai.NewUserTextMessage("baz")
	return msg, state, nil
}

func (a *AlertInvestigationAgent) run(g *genkit.Genkit) aix.AgentFunc[rezai.AlertInvestigationState] {
	return func(ctx context.Context, resp aix.Responder, sr *aix.SessionRunner[rezai.AlertInvestigationState]) (*aix.AgentResult, error) {
		alertId := sr.Custom().AlertID
		fmt.Printf("alert id: %+v\n", alertId)
		alrt, alrtErr := a.alerts.GetAlert(ctx, uuid.Nil)
		if alrtErr != nil {
			return nil, fmt.Errorf("get alert: %w", alrtErr)
		}

		slog.DebugContext(ctx, "agent alert investigation", "title", alrt.Title)
		//_ = &rezai.AlertInvestigationOutput{}

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
