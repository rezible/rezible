package eino

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/agentcaseartifact"
	"github.com/rezible/rezible/ent/agentcasestep"
	"github.com/rezible/rezible/ent/agentrun"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

func (s *AgentService) onIncidentUpdated(ctx context.Context, ev *rez.EventOnIncidentUpdated) error {
	return s.requestIncidentContextPackRun(ctx, ev.IncidentId, "incident_updated")
}

func (s *AgentService) onIncidentImpactsUpdated(ctx context.Context, ev *rez.EventOnIncidentImpactsUpdated) error {
	return s.requestIncidentContextPackRun(ctx, ev.IncidentId, "incident_impacts_updated")
}

func (s *AgentService) requestIncidentContextPackRun(ctx context.Context, incidentID uuid.UUID, trigger string) error {
	if !s.cfg.Enabled || incidentID == uuid.Nil {
		return nil
	}
	bucket := time.Now().UTC().Truncate(5 * time.Minute).Format(time.RFC3339)
	_, reqErr := s.CreateCase(ctx, rez.AgentCaseRequest{
		Title:        "Incident context pack",
		Query:        "Build current triage context for the incident.",
		WorkflowKind: agentrun.WorkflowKindIncidentContextPack,
		SubjectKind:  "incident",
		SubjectID:    incidentID,
		TriggerMetadata: map[string]any{
			"trigger": trigger,
			"bucket":  bucket,
		},
	})
	return reqErr
}

type incidentContextPackWorkflow struct {
	incidents    rez.IncidentService
	modelFactory ModelProvider
}

type incidentContextPackSynthesis struct {
	Summary         string                       `json:"summary"`
	LikelyImpact    []incidentContextPackFinding `json:"likelyImpact"`
	SuggestedChecks []string                     `json:"suggestedChecks"`
	Limitations     []string                     `json:"limitations"`
	Confidence      string                       `json:"confidence"`
}

type incidentContextPackFinding struct {
	EntityID    string   `json:"entityId"`
	DisplayName string   `json:"displayName"`
	Rationale   string   `json:"rationale"`
	EvidenceIDs []string `json:"evidenceIds"`
}

func (w *incidentContextPackWorkflow) Kind() agentrun.WorkflowKind {
	return agentrun.WorkflowKindIncidentContextPack
}

func (w *incidentContextPackWorkflow) Validate(_ context.Context, run *ent.AgentRun) error {
	if run.WorkflowKind != w.Kind() {
		return fmt.Errorf("workflow/run kind mismatch: %s != %s", run.WorkflowKind, w.Kind())
	}
	if run.SubjectKind != "incident" {
		return fmt.Errorf("incident context pack requires incident subject, got %q", run.SubjectKind)
	}
	if run.SubjectID == nil {
		return fmt.Errorf("incident context pack requires subject id")
	}
	if w.incidents == nil {
		return fmt.Errorf("incident service is required")
	}
	return nil
}

var (
	incidentContextPackInstruction = strings.TrimSpace(`
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
)

func (w *incidentContextPackWorkflow) Run(ctx context.Context, run *ent.AgentRun) (*AgentWorkflowResult, error) {
	inputArtifacts, artifactErr := w.incidents.GetIncidentContextArtifacts(ctx, *run.SubjectID)
	if artifactErr != nil {
		return nil, fmt.Errorf("get incident context artifacts: %w", artifactErr)
	}
	artifacts := caseArtifactsFromInputs(inputArtifacts)
	packPayload := retrievalContextPayload(artifacts)

	promptBytes, promptErr := json.MarshalIndent(inputArtifacts, "", "  ")
	if promptErr != nil {
		return nil, fmt.Errorf("marshal context pack prompt: %w", promptErr)
	}

	out, runErr := runModelOnce(ctx, w.modelFactory, "incident-context-pack", incidentContextPackInstruction, string(promptBytes))
	if runErr != nil {
		return nil, fmt.Errorf("run context pack agent: %w", runErr)
	}

	var synthesis incidentContextPackSynthesis
	if parseErr := json.Unmarshal([]byte(extractJSON(out.Text)), &synthesis); parseErr != nil {
		return nil, fmt.Errorf("parse context pack synthesis: %w", parseErr)
	}

	synthesisPayload, synthesisPayloadErr := toArtifactPayload(synthesis)
	if synthesisPayloadErr != nil {
		return nil, synthesisPayloadErr
	}
	conclusionPayload, conclusionPayloadErr := toArtifactPayload(rez.AgentCaseConclusionPayload[rez.IncidentTriageFindings]{
		Summary: synthesis.Summary,
		Findings: rez.IncidentTriageFindings{
			LikelyImpact:    convertIncidentContextPackFindings(synthesis.LikelyImpact, artifacts),
			SuggestedChecks: synthesis.SuggestedChecks,
		},
		Confidence:         synthesis.Confidence,
		Limitations:        synthesis.Limitations,
		RecommendedActions: synthesis.SuggestedChecks,
	})
	if conclusionPayloadErr != nil {
		return nil, conclusionPayloadErr
	}
	return &AgentWorkflowResult{
		Summary: synthesis.Summary,
		Steps: []AgentCaseStep{
			{
				Kind:      agentcasestep.KindRetrieval,
				Title:     "Retrieved incident context",
				Summary:   fmt.Sprintf("Retrieved %d explicit impact(s), %d inferred impact(s), %d active alert(s), and %d related incident(s).", countArtifactsByRole(artifacts, "explicit_impact"), countArtifactsByRole(artifacts, "inferred_impact"), countArtifactsByRole(artifacts, "active_alert"), countArtifactsByRole(artifacts, "related_incident")),
				Output:    packPayload,
				Artifacts: artifacts,
			},
			{
				Kind:    agentcasestep.KindConclusion,
				Title:   "Incident triage conclusion",
				Summary: synthesis.Summary,
				Output:  synthesisPayload,
				Conclusions: []AgentCaseConclusion{{
					Kind:               "incident_triage",
					Summary:            synthesis.Summary,
					Confidence:         synthesis.Confidence,
					RecommendedActions: synthesis.SuggestedChecks,
					Limitations:        synthesis.Limitations,
					Payload:            conclusionPayload,
				}},
			},
		},
	}, nil
}

func convertIncidentContextPackFindings(findings []incidentContextPackFinding, artifacts []AgentCaseArtifact) []rez.AgentEntityRef {
	byID := make(map[string]rez.AgentEntityRef)
	for _, artifact := range artifacts {
		if artifact.Kind != agentcaseartifact.KindEntityRef {
			continue
		}
		ref := rez.AgentEntityRef{}
		bytes, marshalErr := json.Marshal(artifact.Payload)
		if marshalErr != nil {
			continue
		}
		if unmarshalErr := json.Unmarshal(bytes, &ref); unmarshalErr != nil {
			continue
		}
		if ref.EntityID != uuid.Nil {
			byID[ref.EntityID.String()] = ref
		}
	}
	result := make([]rez.AgentEntityRef, len(findings))
	for i, finding := range findings {
		if ref, ok := byID[finding.EntityID]; ok {
			ref.Reason = finding.Rationale
			result[i] = ref
			continue
		}
		result[i] = rez.AgentEntityRef{
			DisplayName: finding.DisplayName,
			Reason:      finding.Rationale,
		}
	}
	return result
}

func toArtifactPayload(v any) (map[string]any, error) {
	bytes, marshalErr := json.Marshal(v)
	if marshalErr != nil {
		return nil, fmt.Errorf("marshal artifact payload: %w", marshalErr)
	}
	var payload map[string]any
	if unmarshalErr := json.Unmarshal(bytes, &payload); unmarshalErr != nil {
		return nil, fmt.Errorf("unmarshal artifact payload: %w", unmarshalErr)
	}
	return payload, nil
}

func extractJSON(text string) string {
	trimmed := strings.TrimSpace(text)
	if strings.HasPrefix(trimmed, "```") {
		trimmed = strings.TrimPrefix(trimmed, "```json")
		trimmed = strings.TrimPrefix(trimmed, "```")
		trimmed = strings.TrimSuffix(trimmed, "```")
	}
	return strings.TrimSpace(trimmed)
}
