package eino

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

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
	_, reqErr := s.CreateTask(ctx, rez.CreateAgentTaskRequest{
		WorkflowKind: rez.AgentWorkflowKindIncidentContextPack,
		WorkflowInput: map[string]any{
			"schema": "incident_context_pack.v1",
			"subjects": []map[string]any{{
				"type": "incident",
				"id":   incidentID.String(),
			}},
			"objectives": []string{"build current triage context for the incident"},
		},
		TriggerKind: "event",
		TriggerPayload: map[string]any{
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
	Summary         string                      `json:"summary"`
	LikelyImpact    []rez.IncidentImpactFinding `json:"likelyImpact"`
	SuggestedChecks []string                    `json:"suggestedChecks"`
	Limitations     []string                    `json:"limitations"`
	Confidence      string                      `json:"confidence"`
}

func (w *incidentContextPackWorkflow) Kind() rez.AgentWorkflowKind {
	return rez.AgentWorkflowKindIncidentContextPack
}

func (w *incidentContextPackWorkflow) Validate(_ context.Context, task *ent.AgentTask, _ *ent.AgentRun) error {
	if rez.AgentWorkflowKind(task.WorkflowKind) != w.Kind() {
		return fmt.Errorf("workflow/task kind mismatch: %s != %s", task.WorkflowKind, w.Kind())
	}
	if workflowInputSubject(task.WorkflowInput, "incident") == uuid.Nil {
		return fmt.Errorf("incident context pack requires incident subject")
	}
	if w.incidents == nil {
		return fmt.Errorf("incident service is required")
	}
	return nil
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

func (w *incidentContextPackWorkflow) Run(ctx context.Context, recorder *WorkflowRecorder, task *ent.AgentTask, _ *ent.AgentRun) (*AgentWorkflowResult, error) {
	incidentID := workflowInputSubject(task.WorkflowInput, "incident")
	contextResult, contextErr := w.incidents.GetIncidentContext(ctx, incidentID)
	toolResult := map[string]any{"incidentId": incidentID}
	if contextResult != nil {
		toolResult["context"] = contextResult.Context
		toolResult["items"] = contextResult.Items
	}
	if _, recordErr := recorder.RecordToolCall(ctx, "incident.context_pack", map[string]any{"incidentId": incidentID}, toolResult, contextErr); recordErr != nil {
		return nil, recordErr
	}
	if contextErr != nil {
		return nil, fmt.Errorf("get incident context: %w", contextErr)
	}

	prompt, promptErr := encodePromptContext(contextResult)
	if promptErr != nil {
		return nil, fmt.Errorf("marshal context pack prompt: %w", promptErr)
	}
	out, runErr := runModelOnce(ctx, w.modelFactory, "incident-context-pack", incidentContextPackInstruction, prompt)
	modelResult := map[string]any{}
	if out != nil {
		modelResult["text"] = out.Text
	}
	if _, recordErr := recorder.RecordToolCall(ctx, "model.generate", map[string]any{"workflow": "incident_context_pack"}, modelResult, runErr); recordErr != nil {
		return nil, recordErr
	}
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

	return &AgentWorkflowResult{
		Content:   synthesis.Summary,
		Data:      incidentContextPackData(synthesis),
		Citations: contextResult.Citations,
		Findings:  incidentContextPackFindings(synthesis, contextResult),
	}, nil
}

func incidentContextPackData(s incidentContextPackSynthesis) map[string]any {
	return map[string]any{
		"schema": "incident_context_pack.v1",
		"findings": rez.IncidentTriageFindings{
			LikelyImpact:    s.LikelyImpact,
			SuggestedChecks: s.SuggestedChecks,
		},
		"limitations":        s.Limitations,
		"recommendedActions": s.SuggestedChecks,
	}
}

func incidentContextPackFindings(s incidentContextPackSynthesis, contextResult *rez.AgentWorkflowContext) []rez.AgentRunFindingInput {
	impactCitations := citationLinks(append(contextItemsWithRole(contextResult.Items, "explicit_impact"), contextItemsWithRole(contextResult.Items, "inferred_impact")...), "affected_entity")
	alertCitations := citationLinks(contextItemsWithRole(contextResult.Items, "active_alert"), "supports")
	evidenceCitations := citationLinks(contextItemsWithRole(contextResult.Items, "recent_evidence"), "supports")
	findings := []rez.AgentRunFindingInput{{
		FindingKind: "observation",
		Content:     s.Summary,
		Citations:   append(append(impactCitations, alertCitations...), evidenceCitations...),
	}}
	for _, impact := range s.LikelyImpact {
		content := impact.DisplayName
		if impact.Rationale != "" {
			content += ": " + impact.Rationale
		}
		findings = append(findings, rez.AgentRunFindingInput{
			FindingKind: "hypothesis",
			Content:     content,
			Citations:   impactCitations,
		})
	}
	for _, limitation := range s.Limitations {
		findings = append(findings, rez.AgentRunFindingInput{
			FindingKind: "limitation",
			Content:     limitation,
		})
	}
	for _, check := range s.SuggestedChecks {
		findings = append(findings, rez.AgentRunFindingInput{
			FindingKind: "recommendation",
			Content:     check,
			Citations:   append(impactCitations, evidenceCitations...),
		})
	}
	return findings
}
