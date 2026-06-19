package eino

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rezible/rezible/ent/agentrun"
	"github.com/rezible/rezible/ent/agentrunartifact"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type incidentContextPackWorkflow struct {
	incidents    rez.IncidentService
	modelFactory ChatModelFactory
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
	var artifacts []AgentRunArtifact

	pack, packErr := w.incidents.GetIncidentContextPack(ctx, *run.SubjectID)
	if packErr != nil {
		return nil, fmt.Errorf("get incident context pack: %w", packErr)
	}

	packPayload, packPayloadErr := toArtifactPayload(pack)
	if packPayloadErr != nil {
		return nil, packPayloadErr
	}
	artifacts = append(artifacts, AgentRunArtifact{
		Kind:    agentrunartifact.KindContext,
		Name:    "retrieval_context",
		Payload: packPayload,
	})

	promptBytes, promptErr := json.MarshalIndent(pack, "", "  ")
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
	artifacts = append(artifacts, AgentRunArtifact{
		Kind:    agentrunartifact.KindResult,
		Name:    "context_pack_synthesis",
		Payload: synthesisPayload,
	})

	return &AgentWorkflowResult{Artifacts: artifacts}, nil
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
