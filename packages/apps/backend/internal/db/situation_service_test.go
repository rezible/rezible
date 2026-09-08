package db

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/agentturn"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/ent/situation"
	"github.com/rezible/rezible/ent/situationhazardassessment"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/rezible/rezible/test"
	"github.com/rezible/rezible/test/mocks"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"golang.org/x/sync/errgroup"
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
	sits, _ := NewSituationService(tdb, jobSvc, agentsSvc, kg)
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

func (s *SituationServiceSuite) expectStartAgentSession(h *situationServiceHarness) {
	h.jobs.EXPECT().
		Insert(mock.Anything, mock.IsType(jobs.StartAgentSession{}), (*river.InsertOpts)(nil)).
		Return(&rivertype.JobInsertResult{Job: &rivertype.JobRow{ID: 1001}}, nil).
		Once()
}

func (s *SituationServiceSuite) expectReconciliation(h *situationServiceHarness) {
	h.jobs.EXPECT().
		Insert(mock.Anything, mock.IsType(jobs.ReconcileSituationInvestigation{}), (*river.InsertOpts)(nil)).
		Return(&rivertype.JobInsertResult{Job: &rivertype.JobRow{ID: 1001}}, nil).
		Once()
}

func (s *SituationServiceSuite) TestCreateSituationCreatesKnowledgeEntityInSameTransaction() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	h := s.newHarness(tdb)
	situation := s.createSituation(ctx, h, "Checkout degradation")

	s.Require().NotEqual(uuid.Nil, situation.KnowledgeEntityID)
	entity := tdb.Client(ctx).KnowledgeEntity.GetX(ctx, situation.KnowledgeEntityID)
	s.Equal("event", string(entity.Category))
	s.Equal("situation", entity.Kind)
	loaded, getErr := h.situations.GetSituation(ctx, situation.ID)
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
	_, closeErr := h.situations.CloseSituation(ctx, rez.CloseSituationParams{
		SituationID: closedSituation.ID,
		Reason:      situation.CloseReasonStabilized,
	})
	s.Require().NoError(closeErr)

	openList, openListErr := h.situations.ListSituations(ctx, rez.ListSituationsParams{
		Status: situation.StatusOpen,
	})
	s.Require().NoError(openListErr)
	openIds := make([]uuid.UUID, len(openList.Data))
	for i, sit := range openList.Data {
		openIds[i] = sit.ID
	}
	s.ElementsMatch([]uuid.UUID{openSituation.ID}, openIds)

	closedList, closedListErr := h.situations.ListSituations(ctx, rez.ListSituationsParams{
		Status: situation.StatusClosed,
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

func (s *SituationServiceSuite) TestSituationCloseIsIdempotentAndNeverReopens() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	h := s.newHarness(tdb)
	sit := s.createSituation(ctx, h, "Checkout degradation")

	closed, closeErr := h.situations.CloseSituation(ctx, rez.CloseSituationParams{
		SituationID: sit.ID,
		Reason:      situation.CloseReasonStabilized,
	})
	s.Require().NoError(closeErr)
	s.Equal(situation.StatusClosed, closed.Status)
	s.NotNil(closed.ClosedAt)
	s.Equal(situation.CloseReasonStabilized, *closed.CloseReason)

	repeated, repeatErr := h.situations.CloseSituation(ctx, rez.CloseSituationParams{
		SituationID: sit.ID,
		Reason:      situation.CloseReasonStabilized,
	})
	s.Require().NoError(repeatErr)
	s.Equal(closed.ID, repeated.ID)
	_, conflictErr := h.situations.CloseSituation(ctx, rez.CloseSituationParams{
		SituationID: sit.ID,
		Reason:      situation.CloseReasonDismissed,
	})
	s.ErrorIs(conflictErr, rez.ErrConflict)
}

func (s *SituationServiceSuite) TestSituationMayExistWithoutInvestigationAndInvestigationCreatesAnalysisAndSession() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	h := s.newHarness(tdb)
	sit := s.createSituation(ctx, h, "Checkout degradation")

	_, missingErr := h.situations.GetSituationInvestigation(ctx, sit.ID)
	s.True(ent.IsNotFound(missingErr))
	s.expectReconciliation(h)
	s.expectStartAgentSession(h)

	investigation, createErr := h.situations.CreateSituationInvestigation(ctx, sit.ID)
	s.Require().NoError(createErr)
	s.Equal(sit.ID, investigation.SituationID)
	s.Equal(1, tdb.Client(ctx).SituationInvestigation.Query().CountX(ctx))
	s.Equal(1, tdb.Client(ctx).SystemAnalysis.Query().CountX(ctx))
	s.Equal(1, tdb.Client(ctx).AgentSession.Query().CountX(ctx))

	session := tdb.Client(ctx).AgentSession.GetX(ctx, investigation.AgentSessionID)
	s.Require().NotNil(session.SystemAnalysisID)
	analysis := tdb.Client(ctx).SystemAnalysis.GetX(ctx, *session.SystemAnalysisID)
	s.Equal(sit.KnowledgeEntityID, *analysis.SubjectEntityID)

	repeated, repeatErr := h.situations.CreateSituationInvestigation(ctx, sit.ID)
	s.Require().NoError(repeatErr)
	s.Equal(investigation.ID, repeated.ID)
}

func (s *SituationServiceSuite) TestConcurrentSituationInvestigationCreationLeavesOneInvestigationAnalysisAndSession() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	h := s.newHarness(tdb)
	sit := s.createSituation(ctx, h, "Checkout degradation")

	s.expectReconciliation(h)
	s.expectStartAgentSession(h)

	results := make(chan *ent.SituationInvestigation, 2)
	var g errgroup.Group
	for range 2 {
		g.Go(func() error {
			investigation, createErr := h.situations.CreateSituationInvestigation(ctx, sit.ID)
			results <- investigation
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
	sit := s.createSituation(ctx, h, "Checkout degradation")
	s.expectReconciliation(h)
	s.expectStartAgentSession(h)

	inv, err := h.situations.CreateSituationInvestigation(ctx, sit.ID)
	s.Require().NoError(err)

	createTurn := tdb.Client(ctx).AgentTurn.Create().
		SetID(uuid.New()).
		SetAgentSessionID(inv.AgentSessionID).
		SetSequence(1).
		SetRiverJobID(1002).
		SetStatus(agentturn.StatusRunning)
	turn := createTurn.SaveX(ctx)
	tdb.Client(ctx).SituationInvestigation.UpdateOneID(inv.ID).
		SetRequestedTurnID(turn.ID).
		SetRequestedRevision(1).
		SaveX(ctx)
	report := schematypes.SituationInvestigationReport{Text: "The database is saturated."}
	updated, reportErr := h.situations.SetSituationInvestigationReport(ctx, rez.SetSituationInvestigationReportParams{AgentTurnID: turn.ID, Report: report})
	s.Require().NoError(reportErr)
	s.Equal(report.Text, updated.Report.Text)
	s.Equal(1, updated.CompletedRevision)
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
	sit := s.createSituation(ctx, h, "Checkout degradation")
	hazard, hazardErr := h.hazards.CreateSystemHazard(ctx, rez.CreateSystemHazardParams{Title: "Checkout unavailable"})
	s.Require().NoError(hazardErr)
	user := tdb.Client(ctx).User.Query().FirstX(ctx)

	first, firstErr := h.situations.AddSituationHazardAssessment(ctx, rez.AddSituationHazardAssessmentParams{
		SituationID:    sit.ID,
		SystemHazardID: hazard.ID,
		Status:         situationhazardassessment.StatusSuspected,
		Summary:        "The observed symptoms may represent this hazard.",
		UserID:         &user.ID,
		AssessedAt:     time.Now().UTC(),
	})
	s.Require().NoError(firstErr)
	second, secondErr := h.situations.AddSituationHazardAssessment(ctx, rez.AddSituationHazardAssessmentParams{
		SituationID:    sit.ID,
		SystemHazardID: hazard.ID,
		Status:         situationhazardassessment.StatusDisproven,
		Summary:        "The symptoms do not match the hazard after review.",
		UserID:         &user.ID,
		AssessedAt:     time.Now().UTC().Add(time.Minute),
	})
	s.Require().NoError(secondErr)
	s.Equal(1, first.Revision)
	s.Equal(2, second.Revision)

	_, noAssessorErr := h.situations.AddSituationHazardAssessment(ctx, rez.AddSituationHazardAssessmentParams{
		SituationID:    sit.ID,
		SystemHazardID: hazard.ID,
		Status:         situationhazardassessment.StatusConfirmed,
		Summary:        "Missing provenance.",
	})
	s.ErrorIs(noAssessorErr, rez.ErrInvalidInput)
	_, twoAssessorsErr := h.situations.AddSituationHazardAssessment(ctx, rez.AddSituationHazardAssessmentParams{
		SituationID:    sit.ID,
		SystemHazardID: hazard.ID,
		Status:         situationhazardassessment.StatusConfirmed,
		Summary:        "Two provenance sources are not supported.",
		UserID:         &user.ID,
		AgentTurnID:    &uuid.Nil,
	})
	s.ErrorIs(twoAssessorsErr, rez.ErrInvalidInput)
}

func (s *SituationServiceSuite) TestSituationServiceTenantIsolation() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	h := s.newHarness(tdb)
	sit := s.createSituation(ctx, h, "Tenant one situation")
	s.Require().NoError(tdb.Client(s.SystemContext()).Tenant.Create().Exec(s.SystemContext()))
	otherTenantID := 2
	otherUserID := uuid.New()
	otherCtx := execution.SetContext(s.T().Context(), execution.Context{
		ActorKind: execution.KindUser,
		Auth: execution.Auth{
			TenantID: &otherTenantID,
			UserID:   &otherUserID,
		},
	})

	_, getErr := h.situations.GetSituation(otherCtx, sit.ID)
	s.Error(getErr)
	_, closeErr := h.situations.CloseSituation(otherCtx, rez.CloseSituationParams{SituationID: sit.ID, Reason: situation.CloseReasonDismissed})
	s.Error(closeErr)
	_, hazardErr := h.hazards.GetSystemHazard(otherCtx, uuid.New())
	s.True(ent.IsNotFound(hazardErr))
}
