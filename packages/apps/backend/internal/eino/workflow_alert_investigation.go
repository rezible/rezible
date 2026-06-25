package eino

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/pkg/agents"
)

type alertInvestigationWorkflow struct {
	alerts       rez.AlertService
	modelFactory ModelProvider
	input        agents.AlertInvestigationInput
	alertID      uuid.UUID
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

func (w *alertInvestigationWorkflow) run(ctx context.Context) (*agents.AlertInvestigationOutput, error) {
	alrt, alrtErr := w.alerts.GetAlert(ctx, w.alertID)
	if alrtErr != nil {
		return nil, fmt.Errorf("get alert: %w", alrtErr)
	}

	slog.DebugContext(ctx, "agent alert investigation", "title", alrt.Title)
	//out := &agents.AlertInvestigationOutput{}

	return nil, fmt.Errorf("workflow not implemented")
}
