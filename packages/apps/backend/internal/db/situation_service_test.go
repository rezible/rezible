package db

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"golang.org/x/sync/errgroup"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	agt "github.com/rezible/rezible/ent/agentturn"
	invr "github.com/rezible/rezible/ent/investigationreport"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/rezible/rezible/ent/situation"
	sitha "github.com/rezible/rezible/ent/situationhazardassessment"
	siti "github.com/rezible/rezible/ent/situationinvestigation"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/rezible/rezible/test"
	"github.com/rezible/rezible/test/mocks"
)

type SituationServiceSuite struct {
	test.Suite
}

func TestSituationServiceSuite(t *testing.T) {
	suite.Run(t, &SituationServiceSuite{Suite: test.NewSuite()})
}

type situationServiceHarness struct {
	tdb        rez.Database
	jobs       *mocks.MockJobService
	agents     *AgentSessionService
	situations *SituationService
	hazards    *SystemHazardService
	knowledge  *KnowledgeGraphService
}

func (s *SituationServiceSuite) newHarness(tdb rez.Database) *situationServiceHarness {
	jobSvc := mocks.NewMockJobService(s.T())
	agentsSvc := &AgentSessionService{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		db:     tdb,
		jobs:   jobSvc,
	}
	kg, _ := NewKnowledgeGraphService(tdb)
	investigations := NewInvestigationService(tdb, agentsSvc)
	sits, _ := NewSituationService(tdb, jobSvc, kg, investigations)
	haz, _ := NewSystemHazardService(tdb, kg)
	return &situationServiceHarness{
		tdb:        tdb,
		jobs:       jobSvc,
		agents:     agentsSvc,
		knowledge:  kg,
		situations: sits,
		hazards:    haz,
	}
}

func (h *situationServiceHarness) expectStartAgentSessionJobInserted(times int) {
	h.jobs.EXPECT().
		Insert(mock.Anything, mock.IsType(jobs.StartAgentSession{}), (*river.InsertOpts)(nil)).
		Return(&rivertype.JobInsertResult{Job: &rivertype.JobRow{ID: 1001}}, nil).
		Times(times)
}

func (h *situationServiceHarness) expectReconciliationJobInserted(times int) {
	h.jobs.EXPECT().
		Insert(mock.Anything, mock.IsType(jobs.BumpSituationInvestigation{}), (*river.InsertOpts)(nil)).
		Return(&rivertype.JobInsertResult{Job: &rivertype.JobRow{ID: 1001}}, nil).
		Times(times)
}

func (h *situationServiceHarness) expectSessionCreationJobsInserted(times int) {
	h.expectStartAgentSessionJobInserted(times)
	h.expectReconciliationJobInserted(times)
}

func (s *SituationServiceSuite) createSituation(ctx context.Context, h *situationServiceHarness, title string) *ent.Situation {
	created, createErr := h.situations.CreateSituation(ctx, rez.CreateSituationParams{
		Title:    title,
		Summary:  "checkout traffic is failing",
		OpenedAt: time.Now().UTC(),
	})
	s.Require().NoError(createErr)
	return created
}

func (s *SituationServiceSuite) createEpisode(ctx context.Context, client *ent.Client) *ent.AlertEpisode {
	definition := client.AlertDefinition.Create().SetTitle("Checkout alert").SaveX(ctx)
	now := time.Now().UTC()
	entity := client.KnowledgeEntity.Create().
		SetCategory(kne.CategoryEvent).
		SetKind("alert_episode").
		SaveX(ctx)
	return client.AlertEpisode.Create().
		SetAlertDefinitionID(definition.ID).
		SetKnowledgeEntityID(entity.ID).
		SetStartedAt(now).
		SetLastObservedAt(now).
		SaveX(ctx)
}

func (s *SituationServiceSuite) TestCreateSituationCreatesKnowledgeEntityInSameTransaction() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	h := s.newHarness(tdb)

	sit := s.createSituation(ctx, h, "Checkout degradation")
	s.Require().NotEqual(uuid.Nil, sit.KnowledgeEntityID)

	entity := tdb.Client(ctx).KnowledgeEntity.GetX(ctx, sit.KnowledgeEntityID)
	s.Equal("event", string(entity.Category))
	s.Equal("situation", entity.Kind)

	loaded, getErr := h.situations.GetSituation(ctx, sit.ID)
	s.Require().NoError(getErr)

	_, edgeErr := loaded.Edges.KnowledgeEntityOrErr()
	s.NoError(edgeErr)
}

func (s *SituationServiceSuite) TestListSituationsFiltersByStatusAndSearch() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	h := s.newHarness(tdb)

	openSituation := s.createSituation(ctx, h, "Checkout degradation")
	closedSituation := s.createSituation(ctx, h, "Payment provider outage")
	closeErr := h.situations.CloseSituation(ctx, rez.CloseSituationParams{
		SituationID: closedSituation.ID,
		Reason:      situation.CloseReasonStabilized,
	})
	s.Require().NoError(closeErr)

	openList, openListErr := h.situations.ListSituations(ctx, rez.ListSituationsParams{
		Active: new(true),
	})
	s.Require().NoError(openListErr)
	openIds := make([]uuid.UUID, len(openList.Data))
	for i, sit := range openList.Data {
		openIds[i] = sit.ID
	}
	s.ElementsMatch([]uuid.UUID{openSituation.ID}, openIds)

	closedList, closedListErr := h.situations.ListSituations(ctx, rez.ListSituationsParams{
		Active: new(false),
	})
	s.Require().NoError(closedListErr)
	s.Len(closedList.Data, 1)
	s.Equal(closedSituation.ID, closedList.Data[0].ID)

	searchList, searchErr := h.situations.ListSituations(ctx, rez.ListSituationsParams{
		Search: "checkout",
	})
	s.Require().NoError(searchErr)
	s.Len(searchList.Data, 1)
	s.Equal(openSituation.ID, searchList.Data[0].ID)
}

func (s *SituationServiceSuite) TestListSituationsActiveIncludesInvestigatingAndExcludesClosed() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	h := s.newHarness(tdb)

	observed := s.createSituation(ctx, h, "Observed")
	investigating := s.createSituation(ctx, h, "Investigating")
	closed := s.createSituation(ctx, h, "Closed")

	closeErr := h.situations.CloseSituation(ctx, rez.CloseSituationParams{
		SituationID: closed.ID,
		Reason:      situation.CloseReasonStabilized,
	})
	s.Require().NoError(closeErr)

	result, listErr := h.situations.ListSituations(ctx, rez.ListSituationsParams{
		Active: new(true),
	})
	s.Require().NoError(listErr)
	s.ElementsMatch([]uuid.UUID{observed.ID, investigating.ID}, []uuid.UUID{result.Data[0].ID, result.Data[1].ID})
}

func (s *SituationServiceSuite) TestSituationCloseIsIdempotentAndNeverReopens() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	h := s.newHarness(tdb)

	sit := s.createSituation(ctx, h, "Checkout degradation")
	closeErr := h.situations.CloseSituation(ctx, rez.CloseSituationParams{
		SituationID: sit.ID,
		Reason:      situation.CloseReasonStabilized,
	})
	s.Require().NoError(closeErr)

	repeatErr := h.situations.CloseSituation(ctx, rez.CloseSituationParams{
		SituationID: sit.ID,
		Reason:      situation.CloseReasonStabilized,
	})
	s.Require().NoError(repeatErr)

	conflictErr := h.situations.CloseSituation(ctx, rez.CloseSituationParams{
		SituationID: sit.ID,
		Reason:      situation.CloseReasonDismissed,
	})
	s.ErrorIs(conflictErr, rez.ErrConflict)
}

func (s *SituationServiceSuite) TestSituationMayExistWithoutInvestigationAndInvestigationCreatesAnalysisAndSession() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()

	h := s.newHarness(tdb)

	h.expectSessionCreationJobsInserted(1)

	_, missingErr := h.situations.LookupSituationInvestigation(ctx, siti.ID(uuid.New()))
	s.True(ent.IsNotFound(missingErr))

	sit := s.createSituation(ctx, h, "Checkout degradation")
	impactParams := rez.CreateSituationInvestigationParams{
		SituationID: sit.ID,
		Query:       new("Assess operational impact."),
	}
	impactSitInv, impactInvErr := h.situations.CreateSituationInvestigation(ctx, impactParams)
	s.Require().NoError(impactInvErr)

	s.Equal(1, tdb.Client(ctx).SituationInvestigation.Query().CountX(ctx))
	s.Equal(1, tdb.Client(ctx).Investigation.Query().CountX(ctx))
	s.Equal(1, tdb.Client(ctx).SystemAnalysis.Query().CountX(ctx))
	s.Equal(1, tdb.Client(ctx).AgentSession.Query().CountX(ctx))

	s.Require().NotNil(impactSitInv.Edges.Investigation)

	analysis := tdb.Client(ctx).SystemAnalysis.GetX(ctx, impactSitInv.Edges.Investigation.SystemAnalysisID)
	s.Equal(sit.KnowledgeEntityID, *analysis.SubjectEntityID)

	factorsParams := rez.CreateSituationInvestigationParams{
		SituationID: sit.ID,
		Query:       new("Assess contributing factors."),
	}
	factorsInv, factorsInvErr := h.situations.CreateSituationInvestigation(ctx, factorsParams)
	s.Require().NoError(factorsInvErr)
	s.Equal(impactSitInv.ID, factorsInv.ID)

	s.Equal(1, tdb.Client(ctx).SituationInvestigation.Query().CountX(ctx))
}

func (s *SituationServiceSuite) TestConcurrentSituationInvestigationCreation() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()

	h := s.newHarness(tdb)
	h.expectSessionCreationJobsInserted(1)

	sit := s.createSituation(ctx, h, "Checkout degradation")
	params := rez.CreateSituationInvestigationParams{SituationID: sit.ID, Query: new("Assess operational impact.")}

	results := make(chan *ent.Investigation, 2)
	var g errgroup.Group
	for range 2 {
		g.Go(func() error {
			investigation, createErr := h.situations.CreateSituationInvestigation(ctx, params)
			results <- investigation.Edges.Investigation
			return createErr
		})
	}
	err := g.Wait()
	close(results)
	s.Require().NoError(err)
	client := tdb.Client(ctx)
	s.Equal(1, client.SituationInvestigation.Query().CountX(ctx))
	s.Equal(1, client.SystemAnalysis.Query().CountX(ctx))
	s.Equal(1, client.AgentSession.Query().CountX(ctx))
}

func (s *SituationServiceSuite) TestSituationInvestigationReportPersistenceWhileRunning() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()

	h := s.newHarness(tdb)
	h.expectSessionCreationJobsInserted(1)

	sit := s.createSituation(ctx, h, "Checkout degradation")

	params := rez.CreateSituationInvestigationParams{
		SituationID: sit.ID,
		Query:       new("Assess operational impact."),
	}
	sitInv, createErr := h.situations.CreateSituationInvestigation(ctx, params)
	s.Require().NoError(createErr)

	queryReport := tdb.Client(ctx).InvestigationReport.Query().Where(invr.InvestigationID(sitInv.ID))

	inv := sitInv.Edges.Investigation
	s.Require().NotNil(inv)

	createTurn := tdb.Client(ctx).AgentTurn.Create().
		SetID(uuid.New()).
		SetAgentSessionID(inv.AgentSessionID).
		SetSequence(1).
		SetRiverJobID(1002).
		SetStatus(agt.StatusRunning)
	turn := createTurn.SaveX(ctx)

	sitInv.Update().
		SetRequestedTurnID(turn.ID).
		SetRequestedRevision(1).
		SaveX(ctx)

	reportInput := rez.InvestigationReportInput{
		Text:               "The database is saturated.",
		Limitations:        []string{"Only the primary database was checked."},
		RecommendedActions: []string{"Scale the database pool."},
		SuggestedChecks:    []string{"Inspect connection wait time."},
	}
	setReportParams := rez.SetSituationInvestigationReportParams{
		AgentTurnID: turn.ID,
		Report:      reportInput,
	}
	updated, reportErr := h.situations.SetSituationInvestigationReport(ctx, setReportParams)
	s.Require().NoError(reportErr)

	report := queryReport.OnlyX(ctx)

	firstCreatedAt := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	report.Update().SetCreatedAt(firstCreatedAt).SaveX(ctx)

	s.Equal(reportInput.Text, report.Text)
	s.Equal(reportInput.Limitations, report.Limitations)
	s.Equal(reportInput.RecommendedActions, report.RecommendedActions)
	s.Equal(reportInput.SuggestedChecks, report.SuggestedChecks)
	s.Equal(1, updated.CompletedRevision)

	repeatedSetReportParams := rez.SetSituationInvestigationReportParams{
		AgentTurnID: turn.ID,
		Report: rez.InvestigationReportInput{
			Text: "This repeated report must not be inserted.",
		},
	}
	repeated, repeatErr := h.situations.SetSituationInvestigationReport(ctx, repeatedSetReportParams)
	s.Require().NoError(repeatErr)
	s.Equal(updated.ID, repeated.ID)
	s.Equal(1, queryReport.CountX(ctx))
}

func (s *SituationServiceSuite) TestSystemHazardRetirementAndRiskAssessmentRevisions() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()

	h := s.newHarness(tdb)

	hazard, createErr := h.hazards.CreateSystemHazard(ctx, rez.CreateSystemHazardParams{
		Title:                 "Database unavailable",
		Description:           "A required datastore cannot serve requests.",
		PotentialConsequences: "Requests fail or become unavailable.",
	})
	s.Require().NoError(createErr)

	first, firstErr := h.hazards.AddSystemHazardRiskAssessment(ctx, rez.AddSystemHazardRiskAssessmentParams{
		SystemHazardID: hazard.ID,
		Likelihood:     "possible",
		Consequence:    "major",
		RiskLevel:      "high",
		AssessedAt:     time.Now().UTC(),
	})
	s.Require().NoError(firstErr)

	second, secondErr := h.hazards.AddSystemHazardRiskAssessment(ctx, rez.AddSystemHazardRiskAssessmentParams{
		SystemHazardID: hazard.ID,
		Likelihood:     "likely",
		Consequence:    "major",
		RiskLevel:      "critical",
		Rationale:      "Recent capacity data increased confidence.",
		AssessedAt:     time.Now().UTC().Add(time.Minute),
	})
	s.Require().NoError(secondErr)
	s.Equal(1, first.Revision)
	s.Equal(2, second.Revision)

	latest, latestErr := h.hazards.GetLatestSystemHazardRiskAssessment(ctx, hazard.ID)
	s.Require().NoError(latestErr)
	s.Equal(second.ID, latest.ID)

	retired, retireErr := h.hazards.RetireSystemHazard(ctx, hazard.ID)
	s.Require().NoError(retireErr)
	s.Equal("retired", string(retired.Status))
}

func (s *SituationServiceSuite) TestSituationHazardAssessmentRevisionsAndAssessorConstraint() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()

	h := s.newHarness(tdb)

	hazard, hazardErr := h.hazards.CreateSystemHazard(ctx, rez.CreateSystemHazardParams{Title: "Checkout unavailable"})
	s.Require().NoError(hazardErr)

	sit := s.createSituation(ctx, h, "Checkout degradation")
	userId := tdb.Client(ctx).User.Query().FirstIDX(ctx)

	first, firstErr := h.situations.AddSituationHazardAssessment(ctx, rez.AddSituationHazardAssessmentParams{
		SituationID:    sit.ID,
		SystemHazardID: hazard.ID,
		Status:         sitha.StatusSuspected,
		Summary:        "The observed symptoms may represent this hazard.",
		UserID:         &userId,
		AssessedAt:     time.Now().UTC(),
	})
	s.Require().NoError(firstErr)

	second, secondErr := h.situations.AddSituationHazardAssessment(ctx, rez.AddSituationHazardAssessmentParams{
		SituationID:    sit.ID,
		SystemHazardID: hazard.ID,
		Status:         sitha.StatusDisproven,
		Summary:        "The symptoms do not match the hazard after review.",
		UserID:         &userId,
		AssessedAt:     time.Now().UTC().Add(time.Minute),
	})
	s.Require().NoError(secondErr)
	s.Equal(1, first.Revision)
	s.Equal(2, second.Revision)

	_, noAssessorErr := h.situations.AddSituationHazardAssessment(ctx, rez.AddSituationHazardAssessmentParams{
		SituationID:    sit.ID,
		SystemHazardID: hazard.ID,
		Status:         sitha.StatusConfirmed,
		Summary:        "Missing provenance.",
	})
	s.ErrorIs(noAssessorErr, rez.ErrInvalidInput)

	_, twoAssessorsErr := h.situations.AddSituationHazardAssessment(ctx, rez.AddSituationHazardAssessmentParams{
		SituationID:    sit.ID,
		SystemHazardID: hazard.ID,
		Status:         sitha.StatusConfirmed,
		Summary:        "Two provenance sources are not supported.",
		UserID:         &userId,
		AgentTurnID:    new(uuid.New()),
	})
	s.ErrorIs(twoAssessorsErr, rez.ErrInvalidInput)
}
