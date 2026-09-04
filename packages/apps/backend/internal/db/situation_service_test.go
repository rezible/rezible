package db

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/ent/situation"
	"github.com/rezible/rezible/ent/situationhazardassessment"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/test"
	"github.com/rezible/rezible/test/mocks"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
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
}

func (s *SituationServiceSuite) newHarness(tdb rez.Database) *situationServiceHarness {
	jobs := mocks.NewMockJobService(s.T())
	agents := &AgentSessionService{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		db:     tdb,
		jobs:   jobs,
	}
	return &situationServiceHarness{
		tdb:        tdb,
		jobs:       jobs,
		agents:     agents,
		situations: &SituationService{db: tdb, agents: agents},
		hazards:    &SystemHazardService{db: tdb},
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
	return client.AlertEpisode.Create().
		SetAlertDefinitionID(definition.ID).
		SetStartedAt(now).
		SetLastObservedAt(now).
		SaveX(ctx)
}

func (s *SituationServiceSuite) expectStartAgentSession(h *situationServiceHarness) {
	h.jobs.EXPECT().
		Insert(mock.Anything, mock.Anything, (*river.InsertOpts)(nil)).
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

func (s *SituationServiceSuite) TestAlertEpisodeLinkIsSetOnceAndClosedSituationsCannotReceiveNewLinks() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	h := s.newHarness(tdb)
	client := tdb.Client(ctx)
	sit := s.createSituation(ctx, h, "Checkout degradation")
	episode := s.createEpisode(ctx, client)

	linked, linkErr := h.situations.LinkAlertEpisodeToSituation(ctx, rez.LinkAlertEpisodeToSituationParams{
		AlertEpisodeID: episode.ID,
		SituationID:    sit.ID,
	})
	s.Require().NoError(linkErr)
	s.Equal(sit.ID, *linked.SituationID)
	_, repeatErr := h.situations.LinkAlertEpisodeToSituation(ctx, rez.LinkAlertEpisodeToSituationParams{
		AlertEpisodeID: episode.ID,
		SituationID:    sit.ID,
	})
	s.NoError(repeatErr)

	other := s.createSituation(ctx, h, "Another situation")
	_, moveErr := h.situations.LinkAlertEpisodeToSituation(ctx, rez.LinkAlertEpisodeToSituationParams{
		AlertEpisodeID: episode.ID,
		SituationID:    other.ID,
	})
	s.ErrorIs(moveErr, rez.ErrConflict)

	closed := s.createSituation(ctx, h, "Closed situation")
	_, closeErr := h.situations.CloseSituation(ctx, rez.CloseSituationParams{
		SituationID: closed.ID,
		Reason:      situation.CloseReasonDismissed,
	})
	s.Require().NoError(closeErr)
	closedEpisode := s.createEpisode(ctx, client)
	_, closedLinkErr := h.situations.LinkAlertEpisodeToSituation(ctx, rez.LinkAlertEpisodeToSituationParams{
		AlertEpisodeID: closedEpisode.ID,
		SituationID:    closed.ID,
	})
	s.ErrorIs(closedLinkErr, rez.ErrConflict)
}

func (s *SituationServiceSuite) TestSituationMayExistWithoutInvestigationAndInvestigationCreatesAnalysisAndSession() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	h := s.newHarness(tdb)
	sit := s.createSituation(ctx, h, "Checkout degradation")

	_, missingErr := h.situations.GetSituationInvestigation(ctx, sit.ID)
	s.True(ent.IsNotFound(missingErr))
	s.expectStartAgentSession(h)
	investigation, createErr := h.situations.CreateSituationInvestigation(ctx, rez.CreateSituationInvestigationParams{SituationID: sit.ID})
	s.Require().NoError(createErr)
	s.Equal(sit.ID, investigation.SituationID)
	s.Equal(1, tdb.Client(ctx).SituationInvestigation.Query().CountX(ctx))
	s.Equal(1, tdb.Client(ctx).SystemAnalysis.Query().CountX(ctx))
	s.Equal(1, tdb.Client(ctx).AgentSession.Query().CountX(ctx))

	session := tdb.Client(ctx).AgentSession.GetX(ctx, investigation.AgentSessionID)
	s.Require().NotNil(session.SystemAnalysisID)
	analysis := tdb.Client(ctx).SystemAnalysis.GetX(ctx, *session.SystemAnalysisID)
	s.Equal(sit.KnowledgeEntityID, *analysis.SubjectEntityID)

	repeated, repeatErr := h.situations.CreateSituationInvestigation(ctx, rez.CreateSituationInvestigationParams{SituationID: sit.ID})
	s.Require().NoError(repeatErr)
	s.Equal(investigation.ID, repeated.ID)
}

func (s *SituationServiceSuite) TestConcurrentSituationInvestigationCreationLeavesOneInvestigationAnalysisAndSession() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	h := s.newHarness(tdb)
	sit := s.createSituation(ctx, h, "Checkout degradation")
	s.expectStartAgentSession(h)

	results := make(chan *ent.SituationInvestigation, 2)
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			investigation, createErr := h.situations.CreateSituationInvestigation(ctx, rez.CreateSituationInvestigationParams{SituationID: sit.ID})
			results <- investigation
			errs <- createErr
		}()
	}
	wg.Wait()
	close(results)
	close(errs)
	for createErr := range errs {
		s.Require().NoError(createErr)
	}
	s.Equal(1, tdb.Client(ctx).SituationInvestigation.Query().CountX(ctx))
	s.Equal(1, tdb.Client(ctx).SystemAnalysis.Query().CountX(ctx))
	s.Equal(1, tdb.Client(ctx).AgentSession.Query().CountX(ctx))
}

func (s *SituationServiceSuite) TestSituationInvestigationReportPersistenceUsesRenamedArtifactAndType() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	h := s.newHarness(tdb)
	sit := s.createSituation(ctx, h, "Checkout degradation")
	s.expectStartAgentSession(h)
	investigation, createErr := h.situations.CreateSituationInvestigation(ctx, rez.CreateSituationInvestigationParams{SituationID: sit.ID})
	s.Require().NoError(createErr)
	report := schematypes.SituationInvestigationReport{Text: "The database is saturated."}
	reportJSON, jsonErr := json.Marshal(report)
	s.Require().NoError(jsonErr)
	tdb.Client(ctx).AgentArtifact.Create().
		SetAgentSessionID(investigation.AgentSessionID).
		SetName("situation_investigation_report").
		SetParts([]*ai.Part{ai.NewJSONPart(string(reportJSON))}).
		SaveX(ctx)

	s.NoError(h.situations.onAgentTurnFinished(ctx, &rezai.EventOnAgentTurnFinished{AgentSessionId: investigation.AgentSessionID}))
	updated := tdb.Client(ctx).SituationInvestigation.GetX(ctx, investigation.ID)
	s.Equal(report.Text, updated.Report.Text)
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

	retired, retireErr := h.hazards.RetireSystemHazard(ctx, rez.RetireSystemHazardParams{SystemHazardID: hazard.ID})
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
