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
	suite.Run(t, &SituationServiceSuite{
		Suite: test.NewSuite(),
	})
}

type situationServiceFixture struct {
	jobs           *mocks.MockJobService
	investigations *InvestigationService
	situations     *SituationService
}

func (s *SituationServiceSuite) newFixture(tdb rez.Database) *situationServiceFixture {
	jobService := mocks.NewMockJobService(s.T())
	agentService := &AiAgentSessionService{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		db:     tdb,
		jobs:   jobService,
	}
	knowledgeQuery, queryServiceErr := NewKnowledgeGraphQueryService(tdb)
	s.Require().NoError(queryServiceErr)

	analysisService, analysisServiceErr := NewSystemAnalysisService(tdb, knowledgeQuery)
	s.Require().NoError(analysisServiceErr)

	investigations := NewInvestigationService(tdb, agentService, jobService)
	situations, situationServiceErr := NewSituationService(tdb, investigations, analysisService, knowledgeQuery)
	s.Require().NoError(situationServiceErr)

	return &situationServiceFixture{
		jobs:           jobService,
		investigations: investigations,
		situations:     situations,
	}
}

func (h *situationServiceFixture) expectStartAgentSessionJobInserted(times int) {
	h.jobs.EXPECT().
		Insert(mock.Anything, mock.IsType(jobs.StartAgentSession{}), (*river.InsertOpts)(nil)).
		Return(&rivertype.JobInsertResult{
			Job: &rivertype.JobRow{
				ID: 1001,
			},
		}, nil).
		Times(times)
}

func (s *SituationServiceSuite) createSituation(ctx context.Context, tdb rez.Database, h *situationServiceFixture, title string) *ent.Situation {
	event := s.situationTestEvent(ctx, tdb)
	createdParams := rez.CreateSituationParams{
		Title:    title,
		Summary:  "checkout traffic is failing",
		OpenedAt: time.Now().UTC(),
		ObservationGroups: []rez.SituationObservationGroupParams{{
			Title:              "Source evidence",
			NormalizedEventIDs: []uuid.UUID{event.ID},
		}},
	}

	created, createErr := h.situations.CreateSituation(ctx, createdParams)
	s.Require().NoError(createErr)

	return created
}

func (s *SituationServiceSuite) situationTestEvent(ctx context.Context, database rez.Database) *ent.NormalizedEvent {
	now := time.Now().UTC()
	createNormalizedEvent := database.Client(ctx).NormalizedEvent.Create().
		SetProvider("test").
		SetProviderNamespace("situations").
		SetProviderResourceRef(uuid.NewString()).
		SetKind("alert").
		SetProviderEventSource("situation-test").
		SetProviderEventRef(uuid.NewString()).
		SetAttributes([]byte(`{"message":"checkout errors increased"}`)).
		SetOccurredAt(now).
		SetReceivedAt(now)
	normalizedEvent, normalizedEventErr := createNormalizedEvent.Save(ctx)
	s.Require().NoError(normalizedEventErr)

	return normalizedEvent
}

func (s *SituationServiceSuite) createEpisode(ctx context.Context, client *ent.Client) *ent.AlertEpisode {
	createDefinition := client.AlertDefinition.Create().
		SetTitle("Checkout alert")
	definition, definitionErr := createDefinition.Save(ctx)
	s.Require().NoError(definitionErr)

	now := time.Now().UTC()
	createEntity := client.KnowledgeEntity.Create().
		SetCategory(kne.CategoryEvent).
		SetKind("alert_episode")
	entity, entityErr := createEntity.Save(ctx)
	s.Require().NoError(entityErr)

	createAlertEpisode := client.AlertEpisode.Create().
		SetAlertDefinitionID(definition.ID).
		SetKnowledgeEntityID(entity.ID).
		SetStartedAt(now).
		SetLastObservedAt(now)
	alertEpisode, alertEpisodeErr := createAlertEpisode.Save(ctx)
	s.Require().NoError(alertEpisodeErr)

	return alertEpisode
}

func (s *SituationServiceSuite) TestCreateSituationCreatesInvestigationFromEvidence() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)

	event := s.situationTestEvent(ctx, tdb)
	h.expectStartAgentSessionJobInserted(1)
	sitParams := rez.CreateSituationParams{
		Title:              "Checkout degradation",
		StartInvestigation: true,
		ObservationGroups: []rez.SituationObservationGroupParams{{
			Title:              "Source evidence",
			NormalizedEventIDs: []uuid.UUID{event.ID},
		}},
	}

	sit, createErr := h.situations.CreateSituation(ctx, sitParams)
	s.Require().NoError(createErr)

	loaded, getErr := h.situations.GetSituation(ctx, sit.ID)
	s.Require().NoError(getErr)
	s.Require().NotNil(loaded.Edges.Investigation)
	investigation := loaded.Edges.Investigation.Edges.Investigation
	s.Require().NotNil(investigation)
	querySystemAnalysisEntry := tdb.Client(ctx).SystemAnalysisEntry.Query().
		Where(sae.AnalysisID(investigation.SystemAnalysisID))
	systemAnalysisEntryCount, systemAnalysisEntryCountErr := querySystemAnalysisEntry.Count(ctx)
	s.Require().NoError(systemAnalysisEntryCountErr)

	s.Equal(1, systemAnalysisEntryCount)
	queryEntry := tdb.Client(ctx).SystemAnalysisEntry.Query().
		Where(sae.AnalysisID(investigation.SystemAnalysisID))
	entry, entryErr := queryEntry.Only(ctx)
	s.Require().NoError(entryErr)
	s.Equal("normalized_event:"+event.ID.String(), *entry.Reference)
	detail, detailErr := h.investigations.ReadInvestigationDetail(ctx, investigation.ID)
	s.Require().NoError(detailErr)
	s.Equal(defaultSituationInvestigationQuestion, detail.Query)
}

func (s *SituationServiceSuite) TestSituationInvestigationMaterializationNormalizesAndDeduplicatesSubjects() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	client := tdb.Client(ctx)
	event := s.situationTestEvent(ctx, tdb)
	createEntity := client.KnowledgeEntity.Create().
		SetCategory(kne.CategoryContainer).
		SetKind("service")
	entity, entityErr := createEntity.Save(ctx)
	s.Require().NoError(entityErr)

	createAlias := client.KnowledgeSubjectAlias.Create().
		SetSubjectKind(ksa.SubjectKindEntity).
		SetProvider(event.Provider).
		SetProviderNamespace(event.ProviderNamespace).
		SetProviderResourceRef(event.ProviderResourceRef).
		SetEntityID(entity.ID)
	alias, aliasErr := createAlias.Save(ctx)
	s.Require().NoError(aliasErr)

	createEvidence := client.KnowledgeEvidence.Create().
		SetEventID(event.ID).
		SetSubjectAliasID(alias.ID).
		SetKind(knev.KindObserved).
		SetAssertion("service_observed").
		SetEffectiveAt(event.OccurredAt).
		SetSubjectState(schematypes.KnowledgeGraphSubjectState{
			DisplayName: "Checkout service",
		})
	evidence, evidenceErr := createEvidence.Save(ctx)
	s.Require().NoError(evidenceErr)

	h.expectStartAgentSessionJobInserted(1)
	sitParams := rez.CreateSituationParams{
		Title:              "Checkout degradation with graph evidence",
		StartInvestigation: true,
		ObservationGroups: []rez.SituationObservationGroupParams{{
			Title:              "Source evidence",
			NormalizedEventIDs: []uuid.UUID{event.ID},
		}},
	}

	sit, createErr := h.situations.CreateSituation(ctx, sitParams)
	s.Require().NoError(createErr)

	loaded, getErr := h.situations.GetSituation(ctx, sit.ID)
	s.Require().NoError(getErr)
	s.Require().NotNil(loaded.Edges.Investigation)
	investigation := loaded.Edges.Investigation.Edges.Investigation
	s.Require().NotNil(investigation)
	analysisID := investigation.SystemAnalysisID
	entryQuery := client.SystemAnalysisEntry.Query().
		Where(sae.AnalysisID(analysisID))
	entry, entryErr := entryQuery.Only(ctx)
	s.Require().NoError(entryErr)

	entitySubjectQuery := client.SystemAnalysisEntrySubject.Query().
		Where(
			saes.EntryID(entry.ID), saes.KnowledgeEntityID(entity.ID),
		)
	entitySubject, entitySubjectErr := entitySubjectQuery.Only(ctx)
	s.Require().NoError(entitySubjectErr)
	s.Equal("context", entitySubject.Role)
	evidenceSubjectQuery := client.SystemAnalysisEntrySubject.Query().
		Where(
			saes.EntryID(entry.ID), saes.KnowledgeEvidenceID(evidence.ID),
		)
	evidenceSubject, evidenceSubjectErr := evidenceSubjectQuery.Only(ctx)
	s.Require().NoError(evidenceSubjectErr)
	s.Equal("supports", evidenceSubject.Role)
	entryCountQuery := client.SystemAnalysisEntry.Query().
		Where(sae.AnalysisID(analysisID))
	entryCount, entryCountErr := entryCountQuery.Count(ctx)
	s.Require().NoError(entryCountErr)

	s.Equal(1, entryCount)
	entrySubjectCountQuery := client.SystemAnalysisEntrySubject.Query().
		Where(saes.EntryID(entry.ID))
	entrySubjectCount, entrySubjectCountErr := entrySubjectCountQuery.Count(ctx)
	s.Require().NoError(entrySubjectCountErr)

	s.Equal(2, entrySubjectCount)

	for range 2 {
		materializeErr := h.situations.prepareOrRefreshSituationAnalysis(ctx, analysisID, []uuid.UUID{event.ID}, nil)
		s.Require().NoError(materializeErr)
	}
	entryCountQuery = client.SystemAnalysisEntry.Query().
		Where(sae.AnalysisID(analysisID))
	refreshedEntryCount, refreshedEntryCountErr := entryCountQuery.Count(ctx)
	s.Require().NoError(refreshedEntryCountErr)

	s.Equal(1, refreshedEntryCount)
	entrySubjectCountQuery = client.SystemAnalysisEntrySubject.Query().
		Where(saes.EntryID(entry.ID))
	refreshedSubjectCount, refreshedSubjectCountErr := entrySubjectCountQuery.Count(ctx)
	s.Require().NoError(refreshedSubjectCountErr)

	s.Equal(2, refreshedSubjectCount)
}

func (s *SituationServiceSuite) TestListSituationsFiltersByStatusAndSearch() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)

	openSituation := s.createSituation(ctx, tdb, h, "Checkout degradation")
	closedSituation := s.createSituation(ctx, tdb, h, "Payment provider outage")
	closeErr := h.situations.CloseSituation(ctx, closedSituation.ID, situation.CloseReasonStabilized)
	s.Require().NoError(closeErr)

	openListParams := rez.ListSituationsParams{
		Active: new(true),
	}

	openList, openListErr := h.situations.ListSituations(ctx, openListParams)
	s.Require().NoError(openListErr)

	openIds := make([]uuid.UUID, len(openList.Data))
	for i, sit := range openList.Data {
		openIds[i] = sit.ID
	}
	s.ElementsMatch([]uuid.UUID{openSituation.ID}, openIds)

	closedListParams := rez.ListSituationsParams{
		Active: new(false),
	}

	closedList, closedListErr := h.situations.ListSituations(ctx, closedListParams)
	s.Require().NoError(closedListErr)
	s.Len(closedList.Data, 1)
	s.Equal(closedSituation.ID, closedList.Data[0].ID)

	searchListParams := rez.ListSituationsParams{
		Search: "checkout",
	}

	searchList, searchErr := h.situations.ListSituations(ctx, searchListParams)
	s.Require().NoError(searchErr)
	s.Len(searchList.Data, 1)
	s.Equal(openSituation.ID, searchList.Data[0].ID)
}

func (s *SituationServiceSuite) TestSituationCloseIsIdempotentAndNeverReopens() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)

	sit := s.createSituation(ctx, tdb, h, "Checkout degradation")
	closeErr := h.situations.CloseSituation(ctx, sit.ID, situation.CloseReasonStabilized)
	s.Require().NoError(closeErr)

	repeatErr := h.situations.CloseSituation(ctx, sit.ID, situation.CloseReasonStabilized)
	s.Require().NoError(repeatErr)

	conflictErr := h.situations.CloseSituation(ctx, sit.ID, situation.CloseReasonDismissed)
	s.ErrorIs(conflictErr, rez.ErrConflict)
}

func (s *SituationServiceSuite) TestCreateSituationRejectsEvidenceFreeSourcesAtomically() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)

	createParams := rez.CreateSituationParams{
		Title: "No evidence",
	}

	_, createErr := h.situations.CreateSituation(ctx, createParams)
	s.Error(createErr)
	situationCount, situationCountErr := tdb.Client(ctx).Situation.Query().Count(ctx)
	s.Require().NoError(situationCountErr)

	s.Zero(situationCount)
	systemAnalysisCount, systemAnalysisCountErr := tdb.Client(ctx).SystemAnalysis.Query().Count(ctx)
	s.Require().NoError(systemAnalysisCountErr)

	s.Zero(systemAnalysisCount)
	investigationCount, investigationCountErr := tdb.Client(ctx).Investigation.Query().Count(ctx)
	s.Require().NoError(investigationCountErr)

	s.Zero(investigationCount)
	agentSessionCount, agentSessionCountErr := tdb.Client(ctx).AgentSession.Query().Count(ctx)
	s.Require().NoError(agentSessionCountErr)

	s.Zero(agentSessionCount)
}

func (s *SituationServiceSuite) TestSituationCanStartFromEpisodeWithoutInstances() {
	ctx, tdb := s.SetupTestDatabase()
	h := s.newFixture(tdb)
	episode := s.createEpisode(ctx, tdb.Client(ctx))
	h.expectStartAgentSessionJobInserted(1)

	sitParams := rez.CreateSituationParams{
		Title:              "Checkout alert",
		StartInvestigation: true,
		ObservationGroups: []rez.SituationObservationGroupParams{{
			Title:           "Alert episode",
			AlertEpisodeIDs: []uuid.UUID{episode.ID},
		}},
	}

	sit, createErr := h.situations.CreateSituation(ctx, sitParams)
	s.Require().NoError(createErr)

	situationInvestigationCount, situationInvestigationCountErr := tdb.Client(ctx).SituationInvestigation.Query().Count(ctx)
	s.Require().NoError(situationInvestigationCountErr)

	s.Equal(1, situationInvestigationCount)
	investigationCount, investigationCountErr := tdb.Client(ctx).Investigation.Query().Count(ctx)
	s.Require().NoError(investigationCountErr)

	s.Equal(1, investigationCount)
	systemAnalysisCount, systemAnalysisCountErr := tdb.Client(ctx).SystemAnalysis.Query().Count(ctx)
	s.Require().NoError(systemAnalysisCountErr)

	s.Equal(1, systemAnalysisCount)
	agentSessionCount, agentSessionCountErr := tdb.Client(ctx).AgentSession.Query().Count(ctx)
	s.Require().NoError(agentSessionCountErr)

	s.Equal(1, agentSessionCount)

	queryLink := tdb.Client(ctx).SituationInvestigation.Query().
		Where(siti.SituationID(sit.ID)).
		WithInvestigation()
	link, linkErr := queryLink.Only(ctx)
	s.Require().NoError(linkErr)

	queryEntry := tdb.Client(ctx).SystemAnalysisEntry.Query().
		Where(sae.AnalysisID(link.Edges.Investigation.SystemAnalysisID))
	entry, entryErr := queryEntry.Only(ctx)
	s.Require().NoError(entryErr)
	s.Equal("alert_episode:"+episode.ID.String(), *entry.Reference)
	queryInvestigationEvidenceRevision := tdb.Client(ctx).InvestigationEvidenceRevision.Query().
		Where(inver.InvestigationID(link.InvestigationID))
	investigationEvidenceRevisionCount, investigationEvidenceRevisionCountErr := queryInvestigationEvidenceRevision.Count(ctx)
	s.Require().NoError(investigationEvidenceRevisionCountErr)

	s.Zero(investigationEvidenceRevisionCount)
	queryAlertInstance := tdb.Client(ctx).AlertInstance.Query().
		Where(alertinstance.AlertEpisodeID(episode.ID))
	alertInstanceCount, alertInstanceCountErr := queryAlertInstance.Count(ctx)
	s.Require().NoError(alertInstanceCountErr)

	s.Equal(0, alertInstanceCount)
}

func (s *SituationServiceSuite) TestSystemHazardRetirementAndRiskAssessmentRevisions() {
	ctx, tdb := s.SetupTestDatabase()

	knowledge, knowledgeServiceErr := NewKnowledgeGraphIngestionService(tdb)
	s.Require().NoError(knowledgeServiceErr)

	hazards, hazardServiceErr := NewSystemHazardService(tdb, knowledge)
	s.Require().NoError(hazardServiceErr)

	hazardParams := rez.CreateSystemHazardParams{
		Title:                 "Database unavailable",
		Description:           "A required datastore cannot serve requests.",
		PotentialConsequences: "Requests fail or become unavailable.",
	}

	hazard, createErr := hazards.CreateSystemHazard(ctx, hazardParams)
	s.Require().NoError(createErr)

	firstParams := rez.AddSystemHazardRiskAssessmentParams{
		SystemHazardID: hazard.ID,
		Likelihood:     "possible",
		Consequence:    "major",
		RiskLevel:      "high",
		AssessedAt:     time.Now().UTC(),
	}

	first, firstErr := hazards.AddSystemHazardRiskAssessment(ctx, firstParams)
	s.Require().NoError(firstErr)

	secondParams := rez.AddSystemHazardRiskAssessmentParams{
		SystemHazardID: hazard.ID,
		Likelihood:     "likely",
		Consequence:    "major",
		RiskLevel:      "critical",
		Rationale:      "Recent capacity data increased confidence.",
		AssessedAt:     time.Now().UTC().Add(time.Minute),
	}

	second, secondErr := hazards.AddSystemHazardRiskAssessment(ctx, secondParams)
	s.Require().NoError(secondErr)
	s.Equal(1, first.Revision)
	s.Equal(2, second.Revision)

	latest, latestErr := hazards.GetLatestSystemHazardRiskAssessment(ctx, hazard.ID)
	s.Require().NoError(latestErr)
	s.Equal(second.ID, latest.ID)

	retired, retireErr := hazards.RetireSystemHazard(ctx, hazard.ID)
	s.Require().NoError(retireErr)
	s.Equal("retired", string(retired.Status))
}

func (s *SituationServiceSuite) TestSituationHazardAssessmentRevisionsAndAssessorConstraint() {
	ctx, tdb := s.SetupTestDatabase()

	knowledge, knowledgeServiceErr := NewKnowledgeGraphIngestionService(tdb)
	s.Require().NoError(knowledgeServiceErr)

	hazards, hazardServiceErr := NewSystemHazardService(tdb, knowledge)
	s.Require().NoError(hazardServiceErr)

	h := s.newFixture(tdb)

	hazardParams := rez.CreateSystemHazardParams{
		Title: "Checkout unavailable",
	}

	hazard, hazardErr := hazards.CreateSystemHazard(ctx, hazardParams)
	s.Require().NoError(hazardErr)

	sit := s.createSituation(ctx, tdb, h, "Checkout degradation")
	userId, userIdErr := tdb.Client(ctx).User.Query().FirstID(ctx)
	s.Require().NoError(userIdErr)

	firstParams := rez.AddSituationHazardAssessmentParams{
		SituationID:    sit.ID,
		SystemHazardID: hazard.ID,
		Status:         sitha.StatusSuspected,
		Summary:        "The observed symptoms may represent this hazard.",
		UserID:         &userId,
		AssessedAt:     time.Now().UTC(),
	}

	first, firstErr := h.situations.AddSituationHazardAssessment(ctx, firstParams)
	s.Require().NoError(firstErr)

	secondParams := rez.AddSituationHazardAssessmentParams{
		SituationID:    sit.ID,
		SystemHazardID: hazard.ID,
		Status:         sitha.StatusDisproven,
		Summary:        "The symptoms do not match the hazard after review.",
		UserID:         &userId,
		AssessedAt:     time.Now().UTC().Add(time.Minute),
	}

	second, secondErr := h.situations.AddSituationHazardAssessment(ctx, secondParams)
	s.Require().NoError(secondErr)
	s.Equal(1, first.Revision)
	s.Equal(2, second.Revision)

	noAssessorParams := rez.AddSituationHazardAssessmentParams{
		SituationID:    sit.ID,
		SystemHazardID: hazard.ID,
		Status:         sitha.StatusConfirmed,
		Summary:        "Missing provenance.",
	}

	_, noAssessorErr := h.situations.AddSituationHazardAssessment(ctx, noAssessorParams)
	s.ErrorIs(noAssessorErr, rez.ErrInvalidInput)

	twoAssessorsParams := rez.AddSituationHazardAssessmentParams{
		SituationID:    sit.ID,
		SystemHazardID: hazard.ID,
		Status:         sitha.StatusConfirmed,
		Summary:        "Two provenance sources are not supported.",
		UserID:         &userId,
		AgentTurnID:    new(uuid.New()),
	}

	_, twoAssessorsErr := h.situations.AddSituationHazardAssessment(ctx, twoAssessorsParams)
	s.ErrorIs(twoAssessorsErr, rez.ErrInvalidInput)
}
