package eino

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/incident"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/pkg/agents"
)

type incidentTriageWorkflow struct {
	incidents    rez.IncidentService
	modelFactory ModelProvider
	incidentID   uuid.UUID
	input        agents.IncidentTriageInput
}

type incidentTriageSynthesis struct {
	Summary         string                         `json:"summary"`
	LikelyImpact    []agents.IncidentImpactFinding `json:"likelyImpact"`
	SuggestedChecks []string                       `json:"suggestedChecks"`
	Limitations     []string                       `json:"limitations"`
	Confidence      string                         `json:"confidence"`
}

var incidentTriageInstruction = strings.TrimSpace(`
You are Rezible's live incident triage agent.
Use only the supplied JSON context. Produce concise JSON with this schema:
{
  "summary": "short responder-oriented synthesis",
  "likelyImpact": [{"entityId":"uuid","displayName":"name","rationale":"why it matters","evidenceIds":["uuid"]}],
  "suggestedChecks": ["short verification step"],
  "limitations": ["known uncertainty"],
  "confidence": "low|medium|high"
}
Do not invent systems, incidents, alerts, or evidence that are not present in the context.
`)

func (w *incidentTriageWorkflow) run(ctx context.Context) (*agents.IncidentTriageOutput, error) {
	inc, incErr := w.incidents.Get(ctx, incident.ID(w.incidentID))
	if incErr != nil {
		return nil, fmt.Errorf("get incident: %w", incErr)
	}

	slog.DebugContext(ctx, "agent incident triage", "title", inc.Title)

	return nil, fmt.Errorf("not implemented")
}
