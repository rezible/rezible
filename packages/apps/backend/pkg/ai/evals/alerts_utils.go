package evals

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/rezible/rezible/ent/schema/schematypes"
	sae "github.com/rezible/rezible/ent/systemanalysisentity"
	saentry "github.com/rezible/rezible/ent/systemanalysisentry"
	saentries "github.com/rezible/rezible/ent/systemanalysisentrysubject"
	sar "github.com/rezible/rezible/ent/systemanalysisrelationship"
	rezai "github.com/rezible/rezible/pkg/ai"
)

type alertFixture struct {
	analysisID  uuid.UUID
	alertEntity *ent.KnowledgeEntity
}

type analysisState struct {
	includedEntityIDs       map[uuid.UUID]struct{}
	includedRelationshipIDs map[uuid.UUID]struct{}
	findingEntityIDs        map[uuid.UUID]struct{}
	findingRelationshipIDs  map[uuid.UUID]struct{}
	findingEvidenceIDs      map[uuid.UUID]struct{}
	findings                []subjectObservation
}

type subjectObservation struct {
	DisplayName string    `json:"displayName"`
	ID          uuid.UUID `json:"id"`
}

func seedBaseAlert(ctx context.Context, client *ent.Client, referenceTime time.Time) (alertFixture, rezai.EvalScenarioSeed, error) {
	alertEntity, entityErr := client.KnowledgeEntity.Create().
		SetCategory(kne.CategorySignal).
		SetKind("alert").
		Save(ctx)
	if entityErr != nil {
		return alertFixture{}, rezai.EvalScenarioSeed{}, fmt.Errorf("create alert knowledge entity: %w", entityErr)
	}

	description := fmt.Sprintf(
		"The production Checkout API 5xx error rate exceeded its warning threshold at %s.",
		referenceTime.Format(time.RFC3339),
	)
	createAlert := client.AlertDefinition.Create().
		SetTitle("Checkout API error rate is high").
		SetDescription(description).
		SetDefinition(`sum(rate(checkout_http_requests_total{environment="production",status=~"5.."}[5m])) > 1`).
		SetKnowledgeEntityID(alertEntity.ID)
	definition, alertErr := createAlert.Save(ctx)
	if alertErr != nil {
		return alertFixture{}, rezai.EvalScenarioSeed{}, fmt.Errorf("create alert: %w", alertErr)
	}

	event, eventErr := client.NormalizedEvent.Create().
		SetProvider("test").
		SetProviderNamespace("alert-evals").
		SetProviderResourceRef("checkout-alert").
		SetProviderEventSource("alerts").
		SetProviderEventRef("alert-event-" + uuid.NewString()).
		SetKind("alert_instance").
		SetAttributes([]byte("{}")).
		SetOccurredAt(referenceTime).
		SetReceivedAt(referenceTime).
		Save(ctx)
	if eventErr != nil {
		return alertFixture{}, rezai.EvalScenarioSeed{}, fmt.Errorf("create normalized event: %w", eventErr)
	}

	episode, episodeErr := client.AlertEpisode.Create().
		SetAlertDefinitionID(definition.ID).
		SetStartedAt(referenceTime).
		SetLastObservedAt(referenceTime).
		Save(ctx)
	if episodeErr != nil {
		return alertFixture{}, rezai.EvalScenarioSeed{}, fmt.Errorf("create alert episode: %w", episodeErr)
	}

	instance, instanceErr := client.AlertInstance.Create().
		SetAlertEpisodeID(episode.ID).
		SetNormalizedEventID(event.ID).
		Save(ctx)
	if instanceErr != nil {
		return alertFixture{}, rezai.EvalScenarioSeed{}, fmt.Errorf("create alert instance: %w", instanceErr)
	}

	createAnalysis := client.SystemAnalysis.Create().
		SetSubjectEntityID(alertEntity.ID).
		SetReferenceTime(referenceTime)
	analysis, analysisErr := createAnalysis.Save(ctx)
	if analysisErr != nil {
		return alertFixture{}, rezai.EvalScenarioSeed{}, fmt.Errorf("create system analysis: %w", analysisErr)
	}

	createAnalysisEntity := client.SystemAnalysisEntity.Create().
		SetAnalysisID(analysis.ID).
		SetKnowledgeEntityID(alertEntity.ID)
	if analysisEntityErr := createAnalysisEntity.Exec(ctx); analysisEntityErr != nil {
		return alertFixture{}, rezai.EvalScenarioSeed{}, fmt.Errorf("include alert entity: %w", analysisEntityErr)
	}

	fixture := alertFixture{analysisID: analysis.ID, alertEntity: alertEntity}
	seed := rezai.EvalScenarioSeed{
		Input:            rezai.AlertAgentInput{AlertInstanceID: instance.ID},
		SystemAnalysisID: &analysis.ID,
	}
	return fixture, seed, nil
}

func decodeInvestigationReport(result *rez.AiAgentInvocationResult) (*schematypes.AlertInvestigationReport, rezai.EvalCheck) {
	check := rezai.EvalCheck{
		ID:       "report_artifact",
		Expected: "valid investigation_report JSON artifact",
	}
	if result == nil {
		check.Summary = "The agent returned no invocation result."
		return nil, check
	}

	for _, artifact := range result.State.Artifacts {
		if artifact == nil || artifact.Name != "investigation_report" {
			continue
		}
		for _, part := range artifact.Parts {
			if part == nil || strings.TrimSpace(part.Text) == "" {
				continue
			}
			rawReport := strings.TrimSpace(part.Text)
			var report schematypes.AlertInvestigationReport
			if decodeErr := json.Unmarshal([]byte(rawReport), &report); decodeErr != nil {
				check.Summary = "The investigation_report artifact contains malformed JSON."
				check.Observed = rawReport
				return nil, check
			}
			normalizeReport(&report)
			check.Passed = true
			check.Summary = "The investigation_report artifact contains valid JSON."
			return &report, check
		}
		check.Summary = "The investigation_report artifact contains no nonblank content."
		return nil, check
	}

	check.Summary = "The investigation_report artifact was not created."
	return nil, check
}

func normalizeReport(report *schematypes.AlertInvestigationReport) {
	report.Text = strings.TrimSpace(report.Text)
	report.LikelyCause = strings.TrimSpace(report.LikelyCause)
	report.BestNextStep = strings.TrimSpace(report.BestNextStep)
	report.Limitations = nonblankStrings(report.Limitations)
	report.SuggestedChecks = nonblankStrings(report.SuggestedChecks)
	report.RecommendedActions = nonblankStrings(report.RecommendedActions)
}

func nonblankStrings(values []string) []string {
	trimmed := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			trimmed = append(trimmed, value)
		}
	}
	return trimmed
}

func reportTextCheck(report *schematypes.AlertInvestigationReport) rezai.EvalCheck {
	passed := report.Text != ""
	summary := "The investigation report contains user-facing text."
	if !passed {
		summary = "The investigation report text is blank."
	}
	return rezai.EvalCheck{ID: "report_text", Passed: passed, Summary: summary, Expected: "nonblank text", Observed: report.Text}
}

func nextActionCheck(report *schematypes.AlertInvestigationReport) rezai.EvalCheck {
	observed := map[string]any{
		"suggestedChecks":    report.SuggestedChecks,
		"recommendedActions": report.RecommendedActions,
		"bestNextStep":       report.BestNextStep,
	}
	passed := len(report.SuggestedChecks)+len(report.RecommendedActions) > 0 || report.BestNextStep != ""
	summary := "The investigation report proposes a next check or action."
	if !passed {
		summary = "The investigation report proposes no nonblank next check or action."
	}
	return rezai.EvalCheck{ID: "next_action", Passed: passed, Summary: summary, Expected: "at least one nonblank check or action", Observed: observed}
}

func queryAnalysisState(ctx context.Context, client *ent.Client, analysisID uuid.UUID) (analysisState, error) {
	state := analysisState{
		includedEntityIDs:       make(map[uuid.UUID]struct{}),
		includedRelationshipIDs: make(map[uuid.UUID]struct{}),
		findingEntityIDs:        make(map[uuid.UUID]struct{}),
		findingRelationshipIDs:  make(map[uuid.UUID]struct{}),
		findingEvidenceIDs:      make(map[uuid.UUID]struct{}),
	}

	entitiesQuery := client.SystemAnalysisEntity.Query().Where(sae.AnalysisID(analysisID))
	includedEntities, entitiesErr := entitiesQuery.All(ctx)
	if entitiesErr != nil {
		return analysisState{}, fmt.Errorf("query included analysis entities: %w", entitiesErr)
	}
	for _, included := range includedEntities {
		state.includedEntityIDs[included.KnowledgeEntityID] = struct{}{}
	}

	relationshipsQuery := client.SystemAnalysisRelationship.Query().Where(sar.AnalysisID(analysisID))
	includedRelationships, relationshipsErr := relationshipsQuery.All(ctx)
	if relationshipsErr != nil {
		return analysisState{}, fmt.Errorf("query included analysis relationships: %w", relationshipsErr)
	}
	for _, included := range includedRelationships {
		state.includedRelationshipIDs[included.KnowledgeRelationshipID] = struct{}{}
	}

	findingsQuery := client.SystemAnalysisEntry.Query().Where(
		saentry.AnalysisID(analysisID),
		saentry.KindEQ(saentry.KindFinding),
	)
	findings, findingsErr := findingsQuery.All(ctx)
	if findingsErr != nil {
		return analysisState{}, fmt.Errorf("query analysis findings: %w", findingsErr)
	}
	for _, finding := range findings {
		state.findings = append(state.findings, subjectObservation{DisplayName: finding.Title, ID: finding.ID})
		subjectsQuery := client.SystemAnalysisEntrySubject.Query().Where(saentries.EntryID(finding.ID))
		subjects, subjectsErr := subjectsQuery.All(ctx)
		if subjectsErr != nil {
			return analysisState{}, fmt.Errorf("query subjects for finding %s: %w", finding.ID, subjectsErr)
		}
		for _, subject := range subjects {
			if subject.KnowledgeEntityID != nil {
				state.findingEntityIDs[*subject.KnowledgeEntityID] = struct{}{}
			}
			if subject.KnowledgeRelationshipID != nil {
				state.findingRelationshipIDs[*subject.KnowledgeRelationshipID] = struct{}{}
			}
			if subject.KnowledgeEvidenceID != nil {
				state.findingEvidenceIDs[*subject.KnowledgeEvidenceID] = struct{}{}
			}
		}
	}
	return state, nil
}
