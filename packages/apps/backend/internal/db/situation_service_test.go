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

	"github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/alertinstance"
	inver "github.com/rezible/rezible/ent/investigationevidencerevision"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knev "github.com/rezible/rezible/ent/knowledgeevidence"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/ent/situation"
	sitha "github.com/rezible/rezible/ent/situationhazardassessment"
	siti "github.com/rezible/rezible/ent/situationinvestigation"
	sae "github.com/rezible/rezible/ent/systemanalysisentry"
	saes "github.com/rezible/rezible/ent/systemanalysisentrysubject"
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
	tdb            rez.Database
	jobs           *mocks.MockJobService
	agents         *AiAgentSessionService
	investigations *InvestigationService
	situations     *SituationService
	hazards        *SystemHazardService
	knowledge      *KnowledgeGraphQueryService
}

func (s *SituationServiceSuite) newHarness(tdb rez.Database) *situationServiceHarness {
	jobSvc := mocks.NewMockJobService(s.T())
	agentsSvc := &AiAgentSessionService{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		db:     tdb,
		jobs:   jobSvc,
	}
	kgQuery, _ := NewKnowledgeGraphQueryService(tdb)
	kgIngest, _ := NewKnowledgeGraphIngestionService(tdb)
	analysisService, _ := NewSystemAnalysisService(tdb, kgQuery)
	investigations := NewInvestigationService(tdb, agentsSvc, jobSvc)
	sits, _ := NewSituationService(tdb, investigations, analysisService, kgQuery)
	haz, _ := NewSystemHazardService(tdb, kgIngest)
	return &situationServiceHarness{
		tdb:            tdb,
		jobs:           jobSvc,
		agents:         agentsSvc,
		investigations: investigations,
		knowledge:      kgQuery,
		situations:     sits,
		hazards:        haz,
	}
}

func (h *situationServiceHarness) expectStartAgentSessionJobInserted(times int) {
	h.jobs.EXPECT().
		Insert(mock.Anything, mock.IsType(jobs.StartAgentSession{}), (*river.InsertOpts)(nil)).
		Return(&rivertype.JobInsertResult{Job: &rivertype.JobRow{ID: 1001}}, nil).
		Times(times)
}

func (s *SituationServiceSuite) createSituation(ctx context.Context, h *situationServiceHarness, title string) *ent.Situation {
	event := situationTestEvent(ctx, h.tdb)
	created, createErr := h.situations.CreateSituation(ctx, rez.CreateSituationParams{
		Title:    title,
		Summary:  "checkout traffic is failing",
		OpenedAt: time.Now().UTC(),
		ObservationGroups: []rez.SituationObservationGroupParams{{
			Title:              "Source evidence",
			NormalizedEventIDs: []uuid.UUID{event.ID},
		}},
	})
	s.Require().NoError(createErr)
	return created
}

func situationTestEvent(ctx context.Context, database rez.Database) *ent.NormalizedEvent {
	now := time.Now().UTC()
	return database.Client(ctx).NormalizedEvent.Create().
		SetProvider("test").SetProviderNamespace("situations").SetProviderResourceRef(uuid.NewString()).
		SetKind("alert").SetProviderEventSource("situation-test").SetProviderEventRef(uuid.NewString()).
		SetAttributes([]byte(`{"message":"checkout errors increased"}`)).
		SetOccurredAt(now).SetReceivedAt(now).
		SaveX(ctx)
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

func (s *SituationServiceSuite) TestCreateSituationCreatesInvestigationFromEvidence() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	h := s.newHarness(tdb)

	event := situationTestEvent(ctx, tdb)
	h.expectStartAgentSessionJobInserted(1)
	sit, createErr := h.situations.CreateSituation(ctx, rez.CreateSituationParams{
		Title:              "Checkout degradation",
		StartInvestigation: true,
		ObservationGroups: []rez.SituationObservationGroupParams{{
			Title: "Source evidence", NormalizedEventIDs: []uuid.UUID{event.ID},
		}},
	})
	s.Require().NoError(createErr)

	loaded, getErr := h.situations.GetSituation(ctx, sit.ID)
	s.Require().NoError(getErr)
	s.Require().NotNil(loaded.Edges.Investigation)
	investigation := loaded.Edges.Investigation.Edges.Investigation
	s.Require().NotNil(investigation)
	s.Equal(1, tdb.Client(ctx).SystemAnalysisEntry.Query().Where(sae.AnalysisID(investigation.SystemAnalysisID)).CountX(ctx))
	entry := tdb.Client(ctx).SystemAnalysisEntry.Query().Where(sae.AnalysisID(investigation.SystemAnalysisID)).OnlyX(ctx)
	s.Equal("normalized_event:"+event.ID.String(), *entry.Reference)
	detail, detailErr := h.investigations.ReadInvestigationDetail(ctx, investigation.ID)
	s.Require().NoError(detailErr)
	s.Equal(defaultSituationInvestigationQuestion, detail.Query)
}

func (s *SituationServiceSuite) TestSituationInvestigationMaterializationNormalizesAndDeduplicatesSubjects() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	h := s.newHarness(tdb)
	client := tdb.Client(ctx)
	event := situationTestEvent(ctx, tdb)
	createEntity := client.KnowledgeEntity.Create().
		SetCategory(kne.CategoryContainer).
		SetKind("service")
	entity := createEntity.SaveX(ctx)
	createAlias := client.KnowledgeSubjectAlias.Create().
		SetSubjectKind(ksa.SubjectKindEntity).
		SetProvider(event.Provider).
		SetProviderNamespace(event.ProviderNamespace).
		SetProviderResourceRef(event.ProviderResourceRef).
		SetEntityID(entity.ID)
	alias := createAlias.SaveX(ctx)
	createEvidence := client.KnowledgeEvidence.Create().
		SetEventID(event.ID).
		SetSubjectAliasID(alias.ID).
		SetKind(knev.KindObserved).
		SetAssertion("service_observed").
		SetEffectiveAt(event.OccurredAt).
		SetSubjectState(schematypes.KnowledgeGraphSubjectState{DisplayName: "Checkout service"})
	evidence := createEvidence.SaveX(ctx)

	h.expectStartAgentSessionJobInserted(1)
	sit, createErr := h.situations.CreateSituation(ctx, rez.CreateSituationParams{
		Title:              "Checkout degradation with graph evidence",
		StartInvestigation: true,
		ObservationGroups: []rez.SituationObservationGroupParams{{
			Title: "Source evidence", NormalizedEventIDs: []uuid.UUID{event.ID},
		}},
	})
	s.Require().NoError(createErr)
	loaded, getErr := h.situations.GetSituation(ctx, sit.ID)
	s.Require().NoError(getErr)
	s.Require().NotNil(loaded.Edges.Investigation)
	investigation := loaded.Edges.Investigation.Edges.Investigation
	s.Require().NotNil(investigation)
	analysisID := investigation.SystemAnalysisID
	entryQuery := client.SystemAnalysisEntry.Query().Where(sae.AnalysisID(analysisID))
	entry := entryQuery.OnlyX(ctx)

	entitySubjectQuery := client.SystemAnalysisEntrySubject.Query().Where(
		saes.EntryID(entry.ID), saes.KnowledgeEntityID(entity.ID),
	)
	entitySubject := entitySubjectQuery.OnlyX(ctx)
	s.Equal("context", entitySubject.Role)
	evidenceSubjectQuery := client.SystemAnalysisEntrySubject.Query().Where(
		saes.EntryID(entry.ID), saes.KnowledgeEvidenceID(evidence.ID),
	)
	evidenceSubject := evidenceSubjectQuery.OnlyX(ctx)
	s.Equal("supports", evidenceSubject.Role)
	entryCountQuery := client.SystemAnalysisEntry.Query().Where(sae.AnalysisID(analysisID))
	s.Equal(1, entryCountQuery.CountX(ctx))
	entrySubjectCountQuery := client.SystemAnalysisEntrySubject.Query().Where(saes.EntryID(entry.ID))
	s.Equal(2, entrySubjectCountQuery.CountX(ctx))

	for range 2 {
		materializeErr := h.situations.prepareOrRefreshSituationAnalysis(ctx, analysisID, []uuid.UUID{event.ID}, nil)
		s.Require().NoError(materializeErr)
	}
	entryCountQuery = client.SystemAnalysisEntry.Query().Where(sae.AnalysisID(analysisID))
	s.Equal(1, entryCountQuery.CountX(ctx))
	entrySubjectCountQuery = client.SystemAnalysisEntrySubject.Query().Where(saes.EntryID(entry.ID))
	s.Equal(2, entrySubjectCountQuery.CountX(ctx))
}

func (s *SituationServiceSuite) TestListSituationsFiltersByStatusAndSearch() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	h := s.newHarness(tdb)

	openSituation := s.createSituation(ctx, h, "Checkout degradation")
	closedSituation := s.createSituation(ctx, h, "Payment provider outage")
	closeErr := h.situations.CloseSituation(ctx, closedSituation.ID, situation.CloseReasonStabilized)
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

func (s *SituationServiceSuite) TestSituationCloseIsIdempotentAndNeverReopens() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	h := s.newHarness(tdb)

	sit := s.createSituation(ctx, h, "Checkout degradation")
	closeErr := h.situations.CloseSituation(ctx, sit.ID, situation.CloseReasonStabilized)
	s.Require().NoError(closeErr)

	repeatErr := h.situations.CloseSituation(ctx, sit.ID, situation.CloseReasonStabilized)
	s.Require().NoError(repeatErr)

	conflictErr := h.situations.CloseSituation(ctx, sit.ID, situation.CloseReasonDismissed)
	s.ErrorIs(conflictErr, rez.ErrConflict)
}

func (s *SituationServiceSuite) TestCreateSituationRejectsEvidenceFreeSourcesAtomically() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	h := s.newHarness(tdb)

	_, createErr := h.situations.CreateSituation(ctx, rez.CreateSituationParams{Title: "No evidence"})
	s.Error(createErr)
	s.Zero(tdb.Client(ctx).Situation.Query().CountX(ctx))
	s.Zero(tdb.Client(ctx).SystemAnalysis.Query().CountX(ctx))
	s.Zero(tdb.Client(ctx).Investigation.Query().CountX(ctx))
	s.Zero(tdb.Client(ctx).AgentSession.Query().CountX(ctx))
}

func (s *SituationServiceSuite) TestSituationCanStartFromEpisodeWithoutInstances() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	h := s.newHarness(tdb)
	episode := s.createEpisode(ctx, tdb.Client(ctx))
	h.expectStartAgentSessionJobInserted(1)

	sit, createErr := h.situations.CreateSituation(ctx, rez.CreateSituationParams{
		Title:              "Checkout alert",
		StartInvestigation: true,
		ObservationGroups: []rez.SituationObservationGroupParams{{
			Title: "Alert episode", AlertEpisodeIDs: []uuid.UUID{episode.ID},
		}},
	})
	s.Require().NoError(createErr)
	s.Equal(1, tdb.Client(ctx).SituationInvestigation.Query().CountX(ctx))
	s.Equal(1, tdb.Client(ctx).Investigation.Query().CountX(ctx))
	s.Equal(1, tdb.Client(ctx).SystemAnalysis.Query().CountX(ctx))
	s.Equal(1, tdb.Client(ctx).AgentSession.Query().CountX(ctx))

	link := tdb.Client(ctx).SituationInvestigation.Query().Where(siti.SituationID(sit.ID)).WithInvestigation().OnlyX(ctx)
	entry := tdb.Client(ctx).SystemAnalysisEntry.Query().Where(sae.AnalysisID(link.Edges.Investigation.SystemAnalysisID)).OnlyX(ctx)
	s.Equal("alert_episode:"+episode.ID.String(), *entry.Reference)
	s.Zero(tdb.Client(ctx).InvestigationEvidenceRevision.Query().Where(inver.InvestigationID(link.InvestigationID)).CountX(ctx))
	s.Equal(0, tdb.Client(ctx).AlertInstance.Query().Where(alertinstance.AlertEpisodeID(episode.ID)).CountX(ctx))
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
