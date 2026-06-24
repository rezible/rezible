package eino

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/pkg/agents"
)

type incidentContextPackWorkflow struct {
	incidents    rez.IncidentService
	modelFactory ModelProvider
	input        agents.IncidentContextPackInput
	incidentID   uuid.UUID
}

type incidentContextPackSynthesis struct {
	Summary         string                         `json:"summary"`
	LikelyImpact    []agents.IncidentImpactFinding `json:"likelyImpact"`
	SuggestedChecks []string                       `json:"suggestedChecks"`
	Limitations     []string                       `json:"limitations"`
	Confidence      string                         `json:"confidence"`
}

var incidentContextPackInstruction = strings.TrimSpace(`
You are Rezible's live incident context-pack agent.
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

func (w *incidentContextPackWorkflow) Run(ctx context.Context) (*agents.RunResult, error) {
	if w.incidents == nil {
		return nil, fmt.Errorf("incident service is required")
	}
	contextResult, contextErr := w.incidents.GetIncidentContext(ctx, w.incidentID)
	if contextErr != nil {
		return nil, fmt.Errorf("get incident context: %w", contextErr)
	}

	prompt, promptErr := encodeWorkflowContext(contextResult)
	if promptErr != nil {
		return nil, fmt.Errorf("marshal context pack prompt: %w", promptErr)
	}
	out, runErr := runModelOnce(ctx, w.modelFactory, "incident-context-pack", incidentContextPackInstruction, prompt)
	if runErr != nil {
		return nil, fmt.Errorf("run context pack agent: %w", runErr)
	}

	var synthesis incidentContextPackSynthesis
	if parseErr := json.Unmarshal([]byte(extractJSON(out.Text)), &synthesis); parseErr != nil {
		return nil, fmt.Errorf("parse context pack synthesis: %w", parseErr)
	}
	if len(synthesis.Limitations) == 0 {
		synthesis.Limitations = contextResult.Limitations
	}
	if len(synthesis.SuggestedChecks) == 0 {
		synthesis.SuggestedChecks = contextResult.Suggested
	}
	data, dataErr := incidentContextPackData(synthesis)
	if dataErr != nil {
		return nil, dataErr
	}

	return &agents.RunResult{
		Content:   synthesis.Summary,
		Data:      data,
		Citations: contextResult.Citations,
		Findings:  incidentContextPackFindings(synthesis, contextResult),
	}, nil
}

func incidentContextPackData(s incidentContextPackSynthesis) (map[string]any, error) {
	return agents.EncodeOutput(agents.IncidentContextPackOutput{
		Findings: agents.IncidentTriageFindings{
			LikelyImpact:    s.LikelyImpact,
			SuggestedChecks: s.SuggestedChecks,
		},
		Limitations:        s.Limitations,
		RecommendedActions: s.SuggestedChecks,
	})
}

func incidentContextPackFindings(s incidentContextPackSynthesis, contextResult *agents.WorkflowContext) []agents.RunFindingInput {
	impactCitations := citationLinks(append(contextItemsWithRole(contextResult.Items, "explicit_impact"), contextItemsWithRole(contextResult.Items, "inferred_impact")...), "affected_entity")
	alertCitations := citationLinks(contextItemsWithRole(contextResult.Items, "active_alert"), "supports")
	evidenceCitations := citationLinks(contextItemsWithRole(contextResult.Items, "recent_evidence"), "supports")
	findings := []agents.RunFindingInput{{
		FindingKind: "observation",
		Content:     s.Summary,
		Citations:   append(append(impactCitations, alertCitations...), evidenceCitations...),
	}}
	for _, impact := range s.LikelyImpact {
		content := impact.DisplayName
		if impact.Rationale != "" {
			content += ": " + impact.Rationale
		}
		findings = append(findings, agents.RunFindingInput{
			FindingKind: "hypothesis",
			Content:     content,
			Citations:   impactCitations,
		})
	}
	for _, limitation := range s.Limitations {
		findings = append(findings, agents.RunFindingInput{
			FindingKind: "limitation",
			Content:     limitation,
		})
	}
	for _, check := range s.SuggestedChecks {
		findings = append(findings, agents.RunFindingInput{
			FindingKind: "recommendation",
			Content:     check,
			Citations:   append(impactCitations, evidenceCitations...),
		})
	}
	return findings
}
