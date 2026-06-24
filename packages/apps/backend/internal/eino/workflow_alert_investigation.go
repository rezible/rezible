package eino

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rezible/rezible/ent/agentcaseartifact"
	"github.com/rezible/rezible/ent/agentcasestep"
	"github.com/rezible/rezible/ent/agentrun"

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

func (w *alertInvestigationWorkflow) Kind() agentrun.WorkflowKind {
	return agentrun.WorkflowKindAlertInvestigation
}

func (w *alertInvestigationWorkflow) Validate(_ context.Context, run *ent.AgentRun) error {
	if run.WorkflowKind != w.Kind() {
		return fmt.Errorf("workflow/run kind mismatch: %s != %s", run.WorkflowKind, w.Kind())
	}
	if run.SubjectKind != "alert" {
		return fmt.Errorf("alert investigation requires alert subject, got %q", run.SubjectKind)
	}
	if run.SubjectID == nil {
		return fmt.Errorf("alert investigation requires subject id")
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

func (w *alertInvestigationWorkflow) Run(ctx context.Context, run *ent.AgentRun) (*AgentWorkflowResult, error) {
	inputArtifacts, artifactErr := w.alerts.GetInvestigationArtifacts(ctx, *run.SubjectID)
	if artifactErr != nil {
		return nil, fmt.Errorf("get alert investigation artifacts: %w", artifactErr)
	}
	artifacts := caseArtifactsFromInputs(inputArtifacts)
	contextPayload := retrievalContextPayload(artifacts)

	synthesis := fallbackAlertInvestigationSynthesis(artifacts)
	if w.aiEnabled {
		if modelSynthesis, modelErr := w.runModelSynthesis(ctx, inputArtifacts); modelErr == nil {
			synthesis = modelSynthesis
		} else {
			synthesis.Limitations = append(synthesis.Limitations, "AI synthesis unavailable: "+modelErr.Error())
		}
	}

	synthesisPayload, synthesisPayloadErr := toArtifactPayload(synthesis)
	if synthesisPayloadErr != nil {
		return nil, synthesisPayloadErr
	}
	conclusionPayload, conclusionPayloadErr := toArtifactPayload(rez.AgentCaseConclusionPayload[rez.AlertInvestigationFindings]{
		Summary: synthesis.Summary,
		Findings: rez.AlertInvestigationFindings{
			LikelyCause:     synthesis.LikelyCause,
			AffectedSystems: synthesis.AffectedSystems,
			SuggestedChecks: synthesis.SuggestedChecks,
			RecommendedNext: synthesis.RecommendedNext,
		},
		Confidence:         synthesis.Confidence,
		Limitations:        synthesis.Limitations,
		RecommendedActions: []string{synthesis.RecommendedNext},
	})
	if conclusionPayloadErr != nil {
		return nil, conclusionPayloadErr
	}
	return &AgentWorkflowResult{
		Summary: synthesis.Summary,
		Steps: []AgentCaseStep{
			{
				Kind:      agentcasestep.KindRetrieval,
				Title:     "Retrieved alert investigation context",
				Summary:   fmt.Sprintf("Retrieved %d subject(s), %d relationship(s), %d signal(s), and %d guide(s).", countArtifactsByRole(artifacts, "likely_subject"), countArtifactsByRole(artifacts, "neighbor"), countArtifactsByRole(artifacts, "recent_signal"), countArtifactsByRole(artifacts, "guide")),
				Output:    contextPayload,
				Artifacts: artifacts,
			},
			{
				Kind:    agentcasestep.KindConclusion,
				Title:   "Alert investigation conclusion",
				Summary: synthesis.Summary,
				Output:  synthesisPayload,
				Conclusions: []AgentCaseConclusion{{
					Kind:               "alert_investigation",
					Summary:            synthesis.Summary,
					Confidence:         synthesis.Confidence,
					RecommendedActions: []string{synthesis.RecommendedNext},
					Limitations:        synthesis.Limitations,
					Payload:            conclusionPayload,
				}},
			},
		},
	}, nil
}

func (w *alertInvestigationWorkflow) runModelSynthesis(ctx context.Context, artifacts []rez.AgentCaseArtifactInput) (alertInvestigationSynthesis, error) {
	promptBytes, promptErr := json.MarshalIndent(artifacts, "", "  ")
	if promptErr != nil {
		return alertInvestigationSynthesis{}, fmt.Errorf("marshal investigation context prompt: %w", promptErr)
	}
	out, runErr := runModelOnce(ctx, w.modelFactory, "alert-investigation", alertInvestigationInstruction, string(promptBytes))
	if runErr != nil {
		return alertInvestigationSynthesis{}, fmt.Errorf("run alert investigation agent: %w", runErr)
	}
	var synthesis alertInvestigationSynthesis
	if parseErr := json.Unmarshal([]byte(extractJSON(out.Text)), &synthesis); parseErr != nil {
		return alertInvestigationSynthesis{}, fmt.Errorf("parse alert investigation synthesis: %w", parseErr)
	}
	return synthesis, nil
}

func fallbackAlertInvestigationSynthesis(artifacts []AgentCaseArtifact) alertInvestigationSynthesis {
	contextPayload := retrievalContextPayload(artifacts)
	alertTitle := stringFromMap(contextPayload, "alertTitle")
	suggestedChecks := stringSliceFromMap(contextPayload, "suggestedChecks")
	limitations := stringSliceFromMap(contextPayload, "limitations")
	subjects := artifactsWithRole(artifacts, "likely_subject")
	signals := artifactsWithRole(artifacts, "recent_signal")
	guides := artifactsWithRole(artifacts, "guide")
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
		Summary:         fmt.Sprintf("%s needs investigation. Rezible matched %d likely affected system(s) and %d recent signal(s).", alertTitle, len(subjects), len(signals)),
		LikelyCause:     likelyCause,
		AffectedSystems: systems,
		SuggestedChecks: suggestedChecks,
		RecommendedNext: recommendedNext,
		Limitations:     limitations,
		Confidence:      fallbackConfidence(subjects, signals, guides),
	}
}

func fallbackConfidence(subjects, signals, guides []AgentCaseArtifact) string {
	if len(subjects) == 0 {
		return "low"
	}
	if len(signals) > 0 || len(guides) > 0 {
		return "medium"
	}
	return "low"
}

func retrievalContextPayload(artifacts []AgentCaseArtifact) map[string]any {
	for _, artifact := range artifacts {
		if artifact.Kind == agentcaseartifact.KindContext && artifact.Name == "retrieval_context" {
			return artifact.Payload
		}
	}
	return map[string]any{}
}

func artifactsWithRole(artifacts []AgentCaseArtifact, role string) []AgentCaseArtifact {
	result := make([]AgentCaseArtifact, 0)
	for _, artifact := range artifacts {
		if artifact.Role == role {
			result = append(result, artifact)
		}
	}
	return result
}

func countArtifactsByRole(artifacts []AgentCaseArtifact, role string) int {
	return len(artifactsWithRole(artifacts, role))
}

func stringFromMap(payload map[string]any, key string) string {
	value, _ := payload[key].(string)
	return value
}

func stringSliceFromMap(payload map[string]any, key string) []string {
	value, ok := payload[key].([]any)
	if !ok {
		if typed, typedOk := payload[key].([]string); typedOk {
			return typed
		}
		return []string{}
	}
	result := make([]string, 0, len(value))
	for _, item := range value {
		if text, textOk := item.(string); textOk {
			result = append(result, text)
		}
	}
	return result
}
