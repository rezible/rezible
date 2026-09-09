package evals

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/agentturn"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/ent/situation"
	"github.com/rezible/rezible/ent/situationinvestigation"
	sae "github.com/rezible/rezible/ent/systemanalysisentity"
	saentry "github.com/rezible/rezible/ent/systemanalysisentry"
	saentries "github.com/rezible/rezible/ent/systemanalysisentrysubject"
	sar "github.com/rezible/rezible/ent/systemanalysisrelationship"
	rezai "github.com/rezible/rezible/pkg/ai"
)

type investigationFixture struct {
	analysisID  uuid.UUID
	situationID uuid.UUID
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

func seedBaseInvestigation(ctx context.Context, client *ent.Client, referenceTime time.Time) (investigationFixture, rezai.EvalScenarioSeed, error) {
	situationEntity, entityErr := client.KnowledgeEntity.Create().
		SetCategory(kne.CategoryEvent).
		SetKind("situation").
		Save(ctx)
	if entityErr != nil {
		return investigationFixture{}, rezai.EvalScenarioSeed{}, fmt.Errorf("create situation knowledge entity: %w", entityErr)
	}

	description := fmt.Sprintf(
		"The production Checkout API 5xx error rate exceeded its warning threshold at %s.",
		referenceTime.Format(time.RFC3339),
	)
	createSituation := client.Situation.Create().
		SetKnowledgeEntityID(situationEntity.ID).
		SetTitle("Checkout API degradation").
		SetSummary(description).
		SetStatus(situation.StatusOpen).
		SetOpenedAt(referenceTime)
	createdSituation, situationErr := createSituation.Save(ctx)
	if situationErr != nil {
		return investigationFixture{}, rezai.EvalScenarioSeed{}, fmt.Errorf("create situation: %w", situationErr)
	}

	createAnalysis := client.SystemAnalysis.Create().
		SetSubjectEntityID(situationEntity.ID).
		SetReferenceTime(referenceTime)
	analysis, analysisErr := createAnalysis.Save(ctx)
	if analysisErr != nil {
		return investigationFixture{}, rezai.EvalScenarioSeed{}, fmt.Errorf("create system analysis: %w", analysisErr)
	}

	createAnalysisEntity := client.SystemAnalysisEntity.Create().
		SetAnalysisID(analysis.ID).
		SetKnowledgeEntityID(situationEntity.ID)
	if analysisEntityErr := createAnalysisEntity.Exec(ctx); analysisEntityErr != nil {
		return investigationFixture{}, rezai.EvalScenarioSeed{}, fmt.Errorf("include situation entity: %w", analysisEntityErr)
	}

	fixture := investigationFixture{analysisID: analysis.ID, situationID: createdSituation.ID}
	raw, err := json.Marshal(rezai.InvestigationAgentInput{SituationID: fixture.situationID})
	if err != nil {
		return fixture, rezai.EvalScenarioSeed{}, err
	}
	createSession := client.AgentSession.Create().SetAgentName(rezai.InvestigationAgent.Name).SetInput(raw).SetSystemAnalysisID(analysis.ID)
	session, err := createSession.Save(ctx)
	if err != nil {
		return fixture, rezai.EvalScenarioSeed{}, err
	}
	createTurn := client.AgentTurn.Create().SetID(uuid.New()).SetAgentSessionID(session.ID).SetSequence(1).SetRiverJobID(0).SetStatus(agentturn.StatusRunning)
	turn, err := createTurn.Save(ctx)
	if err != nil {
		return fixture, rezai.EvalScenarioSeed{}, err
	}
	createInvestigation := client.SituationInvestigation.Create().SetSituationID(fixture.situationID).
		SetSystemAnalysisID(analysis.ID).SetAgentSessionID(session.ID).SetRequestedRevision(1).SetRequestedTurnID(turn.ID)
	if err := createInvestigation.Exec(ctx); err != nil {
		return fixture, rezai.EvalScenarioSeed{}, err
	}
	return fixture, rezai.EvalScenarioSeed{Session: session, Turn: turn}, nil
}

func loadSituationInvestigationReport(ctx context.Context, client *ent.Client, situationID uuid.UUID) (*schematypes.SituationInvestigationReport, rezai.EvalCheck, error) {
	check := rezai.EvalCheck{ID: "accepted_report", Expected: "a persisted investigation report"}
	query := client.SituationInvestigation.Query().Where(situationinvestigation.SituationID(situationID))
	inv, err := query.Only(ctx)
	if err != nil {
		return nil, check, err
	}
	if inv.CompletedRevision == 0 {
		check.Summary = "No investigation report was accepted."
		return nil, check, nil
	}
	normalizeReport(inv.Report)
	check.Passed = true
	check.Summary = "The investigation report was accepted."
	return inv.Report, check, nil
}

func normalizeReport(report *schematypes.SituationInvestigationReport) {
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

func reportTextCheck(report *schematypes.SituationInvestigationReport) rezai.EvalCheck {
	passed := report.Text != ""
	summary := "The investigation report contains user-facing text."
	if !passed {
		summary = "The investigation report text is blank."
	}
	return rezai.EvalCheck{ID: "report_text", Passed: passed, Summary: summary, Expected: "nonblank text", Observed: report.Text}
}

func nextActionCheck(report *schematypes.SituationInvestigationReport) rezai.EvalCheck {
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
