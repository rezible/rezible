package evals

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/agentturn"
	"github.com/rezible/rezible/ent/investigationreport"
	"github.com/rezible/rezible/ent/situationinvestigation"
	"github.com/rezible/rezible/ent/systemanalysisentry"
	rezai "github.com/rezible/rezible/pkg/ai"
)

type investigationFixture struct {
	analysisID  uuid.UUID
	situationID uuid.UUID
}

func seedBaseInvestigation(ctx context.Context, client *ent.Client, referenceTime time.Time) (investigationFixture, rezai.EvalScenarioSeed, error) {
	description := fmt.Sprintf(
		"The production Checkout API 5xx error rate exceeded its warning threshold at %s.",
		referenceTime.Format(time.RFC3339),
	)
	createEvent := client.NormalizedEvent.Create().
		SetProvider("evaluation").
		SetProviderNamespace("investigation").
		SetProviderResourceRef("insufficient-context-" + uuid.NewString()).
		SetKind("alert").
		SetProviderEventSource("evaluation").
		SetProviderEventRef(uuid.NewString()).
		SetAttributes([]byte(`{"message":"checkout errors increased"}`)).
		SetOccurredAt(referenceTime).
		SetReceivedAt(referenceTime)
	event, eventErr := createEvent.Save(ctx)
	if eventErr != nil {
		return investigationFixture{}, rezai.EvalScenarioSeed{}, fmt.Errorf("create normalized event: %w", eventErr)
	}

	createEvidence := client.KnowledgeEvidence.Create().
		SetEventID(event.ID)
	evidence, evidenceErr := createEvidence.Save(ctx)
	if evidenceErr != nil {
		return investigationFixture{}, rezai.EvalScenarioSeed{}, fmt.Errorf("create evidence: %w", eventErr)
	}

	createSituation := client.Situation.Create().
		SetTitle("Checkout API degradation").
		SetSummary(description).
		SetOpenedAt(referenceTime)
	createdSituation, situationErr := createSituation.Save(ctx)
	if situationErr != nil {
		return investigationFixture{}, rezai.EvalScenarioSeed{}, fmt.Errorf("create situation: %w", situationErr)
	}

	createAnalysis := client.SystemAnalysis.Create().
		SetReferenceTime(referenceTime)
	analysis, analysisErr := createAnalysis.Save(ctx)
	if analysisErr != nil {
		return investigationFixture{}, rezai.EvalScenarioSeed{}, fmt.Errorf("create system analysis: %w", analysisErr)
	}

	createGroup := client.SituationObservationGroup.Create().
		SetSituationID(createdSituation.ID).
		SetTitle("Initial normalized event").
		AddEventIDs(event.ID)
	if groupErr := createGroup.Exec(ctx); groupErr != nil {
		return investigationFixture{}, rezai.EvalScenarioSeed{}, fmt.Errorf("create situation evidence group: %w", groupErr)
	}
	createAnalysisEntry := client.SystemAnalysisEntry.Create().
		SetAnalysisID(analysis.ID).
		SetReference("normalized_event:" + event.ID.String()).
		SetKind(systemanalysisentry.KindObservation).
		SetOccurredAt(referenceTime).
		SetSequence(1).
		SetTitle("alert event").
		SetBody("provider_event_source: evaluation\nprovider_event_ref: " + event.ProviderEventRef)
	entry, entryErr := createAnalysisEntry.Save(ctx)
	if entryErr != nil {
		return investigationFixture{}, rezai.EvalScenarioSeed{}, fmt.Errorf("create system analysis observation: %w", entryErr)
	}
	createEntrySubject := client.SystemAnalysisEntrySubject.Create().
		SetEntryID(entry.ID).
		SetRole("evidence").
		SetKnowledgeEvidence(evidence)
	if subjectErr := createEntrySubject.Exec(ctx); subjectErr != nil {
		return investigationFixture{}, rezai.EvalScenarioSeed{}, fmt.Errorf("attach source event to system analysis: %w", subjectErr)
	}

	fixture := investigationFixture{analysisID: analysis.ID, situationID: createdSituation.ID}
	investigationID := uuid.New()

	input := rezai.InvestigationAgentSessionInput{
		Query: "Assess the operational situation.",
	}
	raw, err := json.Marshal(input)
	if err != nil {
		return fixture, rezai.EvalScenarioSeed{}, err
	}
	createSession := client.AgentSession.Create().SetAgentName(rezai.InvestigationAgent.Name).SetInput(raw)
	session, err := createSession.Save(ctx)
	if err != nil {
		return fixture, rezai.EvalScenarioSeed{}, err
	}
	createTurn := client.AgentTurn.Create().SetID(uuid.New()).SetAgentSessionID(session.ID).SetSequence(1).SetRiverJobID(0).SetStatus(agentturn.StatusRunning)
	turn, err := createTurn.Save(ctx)
	if err != nil {
		return fixture, rezai.EvalScenarioSeed{}, err
	}
	createInvestigation := client.Investigation.Create().SetID(investigationID).SetSystemAnalysisID(analysis.ID).SetAgentSessionID(session.ID)
	createdInvestigation, investigationErr := createInvestigation.Save(ctx)
	if investigationErr != nil {
		return fixture, rezai.EvalScenarioSeed{}, investigationErr
	}
	createSituationInvestigation := client.SituationInvestigation.Create().SetSituationID(fixture.situationID).SetInvestigationID(createdInvestigation.ID)
	if joinErr := createSituationInvestigation.Exec(ctx); joinErr != nil {
		return fixture, rezai.EvalScenarioSeed{}, joinErr
	}
	return fixture, rezai.EvalScenarioSeed{Session: session, Turn: turn}, nil
}

func loadSituationInvestigationReport(ctx context.Context, client *ent.Client, situationID uuid.UUID) (*ent.InvestigationReport, rezai.EvalCheck, error) {
	check := rezai.EvalCheck{ID: "accepted_report", Expected: "a persisted investigation report"}
	query := client.SituationInvestigation.Query().Where(situationinvestigation.SituationID(situationID))
	inv, queryErr := query.WithInvestigation().Only(ctx)
	if queryErr != nil {
		return nil, check, queryErr
	}
	reportQuery := client.InvestigationReport.Query().Where(investigationreport.InvestigationID(inv.Edges.Investigation.ID)).WithAgentTurn()
	reports, reportErr := reportQuery.All(ctx)
	if reportErr != nil {
		return nil, check, reportErr
	}
	eligible := make([]*ent.InvestigationReport, 0, len(reports))
	for _, report := range reports {
		if report.Edges.AgentTurn != nil && (report.Edges.AgentTurn.Status == agentturn.StatusRunning || report.Edges.AgentTurn.Status == agentturn.StatusCompleted) {
			eligible = append(eligible, report)
		}
	}
	sort.Slice(eligible, func(i, j int) bool {
		if eligible[i].Edges.AgentTurn.Sequence != eligible[j].Edges.AgentTurn.Sequence {
			return eligible[i].Edges.AgentTurn.Sequence > eligible[j].Edges.AgentTurn.Sequence
		}
		if !eligible[i].CreatedAt.Equal(eligible[j].CreatedAt) {
			return eligible[i].CreatedAt.After(eligible[j].CreatedAt)
		}
		return eligible[i].ID.String() > eligible[j].ID.String()
	})
	if len(eligible) == 0 {
		check.Summary = "No eligible investigation report was published."
		return nil, check, nil
	}
	report := eligible[0]
	check.Passed = true
	check.Summary = "The investigation report was accepted."
	return report, check, nil
}

func normalizeReport(report *ent.InvestigationReport) {
	report.Text = strings.TrimSpace(report.Text)
}

func reportTextCheck(report *ent.InvestigationReport) rezai.EvalCheck {
	passed := report.Text != ""
	summary := "The investigation report contains user-facing text."
	if !passed {
		summary = "The investigation report text is blank."
	}
	return rezai.EvalCheck{ID: "report_text", Passed: passed, Summary: summary, Expected: "nonblank text", Observed: report.Text}
}
