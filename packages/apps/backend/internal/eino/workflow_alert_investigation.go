package eino

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type alertInvestigationWorkflow struct {
	alerts       rez.AlertService
	modelFactory ModelProvider
	aiEnabled    bool
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

func (w *alertInvestigationWorkflow) Kind() rez.AgentWorkflowKind {
	return rez.AgentWorkflowKindAlertInvestigation
}

func (w *alertInvestigationWorkflow) Validate(_ context.Context, task *ent.AgentTask, _ *ent.AgentRun) error {
	if rez.AgentWorkflowKind(task.WorkflowKind) != w.Kind() {
		return fmt.Errorf("workflow/task kind mismatch: %s != %s", task.WorkflowKind, w.Kind())
	}
	if workflowInputSubject(task.WorkflowInput, "alert") == uuid.Nil {
		return fmt.Errorf("alert investigation requires alert subject")
	}
	if w.alerts == nil {
		return fmt.Errorf("alert service is required")
	}
	return nil
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

func (w *alertInvestigationWorkflow) Run(ctx context.Context, recorder *WorkflowRecorder, task *ent.AgentTask, _ *ent.AgentRun) (*AgentWorkflowResult, error) {
	alertID := workflowInputSubject(task.WorkflowInput, "alert")
	contextResult, contextErr := w.alerts.GetInvestigationContext(ctx, alertID)
	toolResult := map[string]any{"alertId": alertID}
	if contextResult != nil {
		toolResult["context"] = contextResult.Context
		toolResult["items"] = contextResult.Items
	}
	if _, recordErr := recorder.RecordToolCall(ctx, "alert.investigation_context", map[string]any{"alertId": alertID}, toolResult, contextErr); recordErr != nil {
		return nil, recordErr
	}
	if contextErr != nil {
		return nil, fmt.Errorf("get alert investigation context: %w", contextErr)
	}

	synthesis := fallbackAlertInvestigationSynthesis(contextResult)
	if w.aiEnabled {
		if modelSynthesis, modelErr := w.runModelSynthesis(ctx, recorder, contextResult); modelErr == nil {
			synthesis = modelSynthesis
		} else {
			synthesis.Limitations = append(synthesis.Limitations, "AI synthesis unavailable: "+modelErr.Error())
		}
	}

	return &AgentWorkflowResult{
		Content:   synthesis.Summary,
		Data:      alertInvestigationData(synthesis),
		Citations: contextResult.Citations,
		Findings:  alertInvestigationFindings(synthesis, contextResult),
	}, nil
}

func (w *alertInvestigationWorkflow) runModelSynthesis(ctx context.Context, recorder *WorkflowRecorder, contextResult *rez.AgentWorkflowContext) (alertInvestigationSynthesis, error) {
	prompt, promptErr := encodePromptContext(contextResult)
	if promptErr != nil {
		return alertInvestigationSynthesis{}, fmt.Errorf("marshal investigation context prompt: %w", promptErr)
	}
	out, runErr := runModelOnce(ctx, w.modelFactory, "alert-investigation", alertInvestigationInstruction, prompt)
	result := map[string]any{}
	if out != nil {
		result["text"] = out.Text
	}
	if _, recordErr := recorder.RecordToolCall(ctx, "model.generate", map[string]any{"workflow": "alert_investigation"}, result, runErr); recordErr != nil {
		return alertInvestigationSynthesis{}, recordErr
	}
	if runErr != nil {
		return alertInvestigationSynthesis{}, fmt.Errorf("run alert investigation agent: %w", runErr)
	}
	var synthesis alertInvestigationSynthesis
	if parseErr := json.Unmarshal([]byte(extractJSON(out.Text)), &synthesis); parseErr != nil {
		return alertInvestigationSynthesis{}, fmt.Errorf("parse alert investigation synthesis: %w", parseErr)
	}
	return synthesis, nil
}

func fallbackAlertInvestigationSynthesis(contextResult *rez.AgentWorkflowContext) alertInvestigationSynthesis {
	counts := countContextItemsByRole(contextResult.Items)
	alertTitle, _ := contextResult.Context["alertTitle"].(string)
	subjects := contextItemsWithRole(contextResult.Items, "likely_subject")
	signals := contextItemsWithRole(contextResult.Items, "recent_signal")
	guides := contextItemsWithRole(contextResult.Items, "guide")
	systems := make([]string, 0, len(subjects))
	for _, subject := range subjects {
		systems = append(systems, subject.Name)
	}
	likelyCause := "Insufficient evidence to isolate a cause."
	if len(signals) > 0 {
		likelyCause = "Recent related evidence exists for the likely affected systems; inspect the newest signals first."
	}
	recommendedNext := "monitor"
	if len(subjects) > 0 {
		recommendedNext = "declare_incident"
	}
	return alertInvestigationSynthesis{
		Summary:         fmt.Sprintf("%s needs investigation. Rezible matched %d likely affected system(s), %d recent signal(s), and %d guide(s).", alertTitle, counts["likely_subject"], counts["recent_signal"], counts["guide"]),
		LikelyCause:     likelyCause,
		AffectedSystems: systems,
		SuggestedChecks: contextResult.Suggested,
		RecommendedNext: recommendedNext,
		Limitations:     contextResult.Limitations,
		Confidence:      fallbackConfidence(subjects, signals, guides),
	}
}

func alertInvestigationData(s alertInvestigationSynthesis) map[string]any {
	return map[string]any{
		"schema": "alert_investigation.v1",
		"findings": rez.AlertInvestigationFindings{
			LikelyCause:     s.LikelyCause,
			AffectedSystems: s.AffectedSystems,
			SuggestedChecks: s.SuggestedChecks,
			RecommendedNext: s.RecommendedNext,
		},
		"limitations":        s.Limitations,
		"recommendedActions": []string{s.RecommendedNext},
	}
}

func alertInvestigationFindings(s alertInvestigationSynthesis, contextResult *rez.AgentWorkflowContext) []rez.AgentRunFindingInput {
	signalCitations := citationLinks(contextItemsWithRole(contextResult.Items, "recent_signal"), "supports")
	subjectCitations := citationLinks(contextItemsWithRole(contextResult.Items, "likely_subject"), "affected_entity")
	findings := []rez.AgentRunFindingInput{{
		FindingKind: "observation",
		Content:     s.Summary,
		Citations:   append(subjectCitations, signalCitations...),
	}, {
		FindingKind: "hypothesis",
		Content:     s.LikelyCause,
		Citations:   signalCitations,
	}, {
		FindingKind: "recommendation",
		Content:     s.RecommendedNext,
		Citations:   subjectCitations,
	}}
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
			Citations:   append(subjectCitations, signalCitations...),
		})
	}
	return findings
}

func contextItemsWithRole(items []rez.AgentWorkflowContextItem, role string) []rez.AgentWorkflowContextItem {
	result := make([]rez.AgentWorkflowContextItem, 0)
	for _, item := range items {
		if item.Role == role {
			result = append(result, item)
		}
	}
	return result
}

func countContextItemsByRole(items []rez.AgentWorkflowContextItem) map[string]int {
	counts := make(map[string]int)
	for _, item := range items {
		if item.Role != "" {
			counts[item.Role]++
		}
	}
	return counts
}

func citationLinks(items []rez.AgentWorkflowContextItem, supportKind string) []rez.AgentRunFindingCitationInput {
	result := make([]rez.AgentRunFindingCitationInput, 0)
	seen := make(map[int]struct{})
	for _, item := range items {
		if item.Citation <= 0 {
			continue
		}
		if _, ok := seen[item.Citation]; ok {
			continue
		}
		seen[item.Citation] = struct{}{}
		result = append(result, rez.AgentRunFindingCitationInput{
			CitationIndex: item.Citation,
			SupportKind:   supportKind,
		})
	}
	return result
}

func fallbackConfidence(subjects, signals, guides []rez.AgentWorkflowContextItem) string {
	if len(subjects) == 0 {
		return "low"
	}
	if len(signals) > 0 || len(guides) > 0 {
		return "medium"
	}
	return "low"
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
