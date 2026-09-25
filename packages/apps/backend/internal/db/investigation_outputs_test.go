package db

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/agentturn"
	"github.com/rezible/rezible/ent/investigationfinding"
	"github.com/rezible/rezible/ent/investigationfindingversion"
	"github.com/rezible/rezible/ent/investigationhypothesisversion"
	"github.com/rezible/rezible/ent/investigationoutputreference"
	"github.com/rezible/rezible/ent/investigationreport"
	"github.com/rezible/rezible/ent/investigationuserinput"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	kev "github.com/rezible/rezible/ent/knowledgeevidence"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/ent/schema/schematypes"
	"github.com/rezible/rezible/ent/systemanalysisentry"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/rezible/rezible/pkg/projections"
	"github.com/rezible/rezible/test/mocks"
	"github.com/riverqueue/river/rivertype"
	"github.com/stretchr/testify/mock"
)

func (s *InvestigationServiceSuite) outputFixture(ctx context.Context, tdb rez.Database, jobService *mocks.MockJobService) (*InvestigationService, *ent.Investigation, *ent.AgentTurn) {
	service, investigation := s.createLifecycleInvestigation(ctx, tdb, jobService)
	turn := s.createTurn(ctx, tdb, investigation.AgentSessionID, 1, agentturn.StatusRunning)
	return service, investigation, turn
}

func (s *InvestigationServiceSuite) outputCitationFixture(ctx context.Context, tdb rez.Database, analysisID uuid.UUID) (*ent.KnowledgeEntity, *ent.KnowledgeEvidence, *ent.KnowledgeEvidence, *ent.NormalizedEvent) {
	client := tdb.Client(ctx)
	now := time.Now().UTC().Truncate(time.Second)
	encodedAttributes, encodeErr := projections.EncodeAttributes(struct{}{})
	s.Require().NoError(encodeErr)
	event := client.NormalizedEvent.Create().
		SetProvider("investigation-tests").
		SetProviderNamespace("investigation-tests").
		SetProviderResourceRef(uuid.NewString()).
		SetProviderEventSource("investigation-tests").
		SetProviderEventRef(uuid.NewString()).
		SetKind("deployment").
		SetAttributes(encodedAttributes).
		SetOccurredAt(now).
		SetReceivedAt(now).
		SaveX(ctx)
	entity := client.KnowledgeEntity.Create().
		SetCategory(kne.CategoryContainer).
		SetKind("service").
		SaveX(ctx)
	client.SystemAnalysisEntity.Create().SetAnalysisID(analysisID).SetKnowledgeEntityID(entity.ID).ExecX(ctx)
	alias := client.KnowledgeSubjectAlias.Create().
		SetSubjectKind(ksa.SubjectKindEntity).
		SetProvider("investigation-tests").
		SetProviderNamespace("investigation-tests").
		SetProviderResourceRef(uuid.NewString()).
		SetEntityID(entity.ID).
		SaveX(ctx)
	evidence := client.KnowledgeEvidence.Create().
		SetEventID(event.ID).
		SetSubjectAliasID(alias.ID).
		SetKind(kev.KindDeleted).
		SetAssertion("source_deleted").
		SetEffectiveAt(now).
		SetSubjectState(schematypes.KnowledgeGraphSubjectState{DisplayName: "Payment service"}).
		SaveX(ctx)
	secondAlias := client.KnowledgeSubjectAlias.Create().
		SetSubjectKind(ksa.SubjectKindEntity).
		SetProvider("investigation-tests").
		SetProviderNamespace("investigation-tests").
		SetProviderResourceRef(uuid.NewString()).
		SetEntityID(entity.ID).
		SaveX(ctx)
	secondEvidence := client.KnowledgeEvidence.Create().
		SetEventID(event.ID).
		SetSubjectAliasID(secondAlias.ID).
		SetKind(kev.KindObserved).
		SetAssertion("source_observed").
		SetEffectiveAt(now).
		SetSubjectState(schematypes.KnowledgeGraphSubjectState{DisplayName: "Payment service"}).
		SaveX(ctx)
	entry := client.SystemAnalysisEntry.Create().
		SetAnalysisID(analysisID).
		SetKind(systemanalysisentry.KindObservation).
		SetTitle("The payment service changed").
		SetBody("A stored source record remains attached to this observation.").
		SaveX(ctx)
	client.SystemAnalysisEntrySubject.Create().SetEntryID(entry.ID).SetRole("primary").SetKnowledgeEntityID(entity.ID).ExecX(ctx)
	client.SystemAnalysisEntrySubject.Create().SetEntryID(entry.ID).SetRole("evidence_for").SetKnowledgeEvidenceID(evidence.ID).ExecX(ctx)
	client.SystemAnalysisEntrySubject.Create().SetEntryID(entry.ID).SetRole("evidence_for").SetKnowledgeEvidenceID(secondEvidence.ID).ExecX(ctx)
	return entity, evidence, secondEvidence, event
}

func (s *InvestigationServiceSuite) TestInvestigationReportPublicationSelectionAndRetry() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	jobService := mocks.NewMockJobService(s.T())
	service, investigation, firstTurn := s.outputFixture(ctx, tdb, jobService)

	reportA, publishAErr := service.PublishInvestigationReport(ctx, rez.InvestigationPublicationScope{InvestigationID: investigation.ID, AgentTurnID: firstTurn.ID}, rez.PublishInvestigationReportParams{Text: " Initial report. "})
	s.Require().NoError(publishAErr)
	s.Equal("Initial report.", reportA.Text)
	s.NotNil(reportA.EvidenceIDs)
	s.Empty(reportA.EvidenceIDs)
	s.Equal(agentturn.StatusRunning, reportA.TurnStatus)
	tdb.Client(ctx).AgentTurn.UpdateOneID(firstTurn.ID).SetStatus(agentturn.StatusCompleted).ExecX(ctx)

	secondTurn := s.createTurn(ctx, tdb, investigation.AgentSessionID, 2, agentturn.StatusRunning)
	reportB, publishBErr := service.PublishInvestigationReport(ctx, rez.InvestigationPublicationScope{InvestigationID: investigation.ID, AgentTurnID: secondTurn.ID}, rez.PublishInvestigationReportParams{Text: " Revised report. "})
	s.Require().NoError(publishBErr)
	repeatedB, repeatBErr := service.PublishInvestigationReport(ctx, rez.InvestigationPublicationScope{InvestigationID: investigation.ID, AgentTurnID: secondTurn.ID}, rez.PublishInvestigationReportParams{Text: "Revised report."})
	s.Require().NoError(repeatBErr)
	s.Equal(reportB.ID, repeatedB.ID)
	s.Equal(agentturn.StatusRunning, repeatedB.TurnStatus)
	latest, latestErr := service.ReadInvestigationReport(ctx, investigation.ID, rez.ReadInvestigationReportParams{})
	s.Require().NoError(latestErr)
	s.Equal(reportB.ID, latest.ID)
	completed, completedErr := service.ReadInvestigationReport(ctx, investigation.ID, rez.ReadInvestigationReportParams{Selection: "completed"})
	s.Require().NoError(completedErr)
	s.Equal(reportA.ID, completed.ID)

	workerMessages := mocks.NewMockMessageQueue(s.T())
	workerMessages.EXPECT().Publish(mock.Anything, mock.IsType(rezai.AgentTurnUpdated{})).Return(nil).Maybe()
	worker := &InvokeAgentTurnWorker{db: tdb, msgs: workerMessages, logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	automaticRetryErr := worker.saveInvocationResult(ctx, makeAgentTurnJob(secondTurn, 1), nil, nil, errors.New("temporary model failure"))
	s.Require().NoError(automaticRetryErr)
	queued, queuedErr := service.ReadInvestigationReport(ctx, investigation.ID, rez.ReadInvestigationReportParams{})
	s.Require().NoError(queuedErr)
	s.Equal(reportA.ID, queued.ID)
	tdb.Client(ctx).AgentTurn.UpdateOneID(secondTurn.ID).SetStatus(agentturn.StatusRunning).ExecX(ctx)
	automaticRetryVisible, automaticRetryVisibleErr := service.ReadInvestigationReport(ctx, investigation.ID, rez.ReadInvestigationReportParams{})
	s.Require().NoError(automaticRetryVisibleErr)
	s.Equal(reportB.ID, automaticRetryVisible.ID)

	tdb.Client(ctx).AgentTurn.UpdateOneID(secondTurn.ID).SetStatus(agentturn.StatusFailed).ExecX(ctx)
	retryJobs := mocks.NewMockJobService(s.T())
	retryMessages := mocks.NewMockMessageQueue(s.T())
	retryMessages.EXPECT().Publish(mock.Anything, mock.IsType(rezai.AgentTurnUpdated{})).Return(nil).Once()
	retryJobs.EXPECT().Insert(mock.Anything, jobs.InvokeAgentTurn{AgentSessionID: investigation.AgentSessionID, AgentTurnID: secondTurn.ID}, mock.Anything).
		Return(&rivertype.JobInsertResult{Job: &rivertype.JobRow{ID: 5002}}, nil).Once()
	sessionService := &AiAgentSessionService{db: tdb, jobs: retryJobs, msgs: retryMessages}
	retriedTurn, retryErr := sessionService.RetryAgentTurn(ctx, secondTurn.ID)
	s.Require().NoError(retryErr)
	s.Equal(agentturn.StatusQueued, retriedTurn.Status)
	queuedAgain, queuedAgainErr := service.ReadInvestigationReport(ctx, investigation.ID, rez.ReadInvestigationReportParams{})
	s.Require().NoError(queuedAgainErr)
	s.Equal(reportA.ID, queuedAgain.ID)
	tdb.Client(ctx).AgentTurn.UpdateOneID(secondTurn.ID).SetStatus(agentturn.StatusRunning).ExecX(ctx)
	sharedRetryVisible, sharedRetryVisibleErr := service.ReadInvestigationReport(ctx, investigation.ID, rez.ReadInvestigationReportParams{})
	s.Require().NoError(sharedRetryVisibleErr)
	s.Equal(reportB.ID, sharedRetryVisible.ID)

	tdb.Client(ctx).AgentTurn.UpdateOneID(secondTurn.ID).SetStatus(agentturn.StatusCompleted).ExecX(ctx)
	completedB, completedBErr := service.ReadInvestigationReport(ctx, investigation.ID, rez.ReadInvestigationReportParams{Selection: "completed"})
	s.Require().NoError(completedBErr)
	s.Equal(reportB.ID, completedB.ID)
	thirdTurn := s.createTurn(ctx, tdb, investigation.AgentSessionID, 3, agentturn.StatusRunning)
	newTurnB, newTurnBErr := service.PublishInvestigationReport(ctx, rez.InvestigationPublicationScope{InvestigationID: investigation.ID, AgentTurnID: thirdTurn.ID}, rez.PublishInvestigationReportParams{Text: "Revised report."})
	s.Require().NoError(newTurnBErr)
	s.NotEqual(reportB.ID, newTurnB.ID)
}

func (s *InvestigationServiceSuite) TestInvestigationAnswerOwnershipAndRevisions() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	jobService := mocks.NewMockJobService(s.T())
	service, investigation, turn := s.outputFixture(ctx, tdb, jobService)
	initialAnswer, initialAnswerErr := service.PublishInvestigationAnswer(ctx, rez.InvestigationPublicationScope{InvestigationID: investigation.ID, AgentTurnID: turn.ID}, rez.PublishInvestigationAnswerParams{
		Title: "Initial turn cannot answer", Body: "There is no directly assigned user input.",
	})
	s.Nil(initialAnswer)
	s.ErrorIs(initialAnswerErr, rez.ErrInvalidInput)
	userCtx := s.userContext(tdb, ctx)
	userID, userIDSet := execution.GetContext(userCtx).UserID()
	s.Require().True(userIDSet)
	firstInput := tdb.Client(ctx).InvestigationUserInput.Create().SetInvestigationID(investigation.ID).SetUserID(userID).
		SetText("Why did retries increase?").SetKey("question-one").SetAgentTurnID(turn.ID).SaveX(ctx)
	waitingInput := tdb.Client(ctx).InvestigationUserInput.Create().SetInvestigationID(investigation.ID).SetUserID(userID).
		SetText("What changed after deployment?").SetKey("question-two").SaveX(ctx)
	_, citationEvidence, secondEvidence, _ := s.outputCitationFixture(ctx, tdb, investigation.SystemAnalysisID)
	evidenceIDs := []uuid.UUID{secondEvidence.ID, citationEvidence.ID, citationEvidence.ID}
	answer, publishErr := service.PublishInvestigationAnswer(ctx, rez.InvestigationPublicationScope{InvestigationID: investigation.ID, AgentTurnID: turn.ID}, rez.PublishInvestigationAnswerParams{
		Title: "Retry increase", Body: "The supplied analysis has attached evidence records, including one deleted-kind record.", EvidenceIDs: evidenceIDs,
	})
	s.Require().NoError(publishErr)
	s.Equal(firstInput.ID, *answer.UserInputID)
	s.Equal("answer:"+firstInput.ID.String(), answer.Key)
	expectedEvidenceIDs := []uuid.UUID{citationEvidence.ID, secondEvidence.ID}
	sort.Slice(expectedEvidenceIDs, func(i, j int) bool { return expectedEvidenceIDs[i].String() < expectedEvidenceIDs[j].String() })
	s.Equal(expectedEvidenceIDs, answer.EvidenceIDs)
	repeatedAnswer, repeatErr := service.PublishInvestigationAnswer(ctx, rez.InvestigationPublicationScope{InvestigationID: investigation.ID, AgentTurnID: turn.ID}, rez.PublishInvestigationAnswerParams{
		Title: " Retry increase ", Body: "The supplied analysis has attached evidence records, including one deleted-kind record. ",
		EvidenceIDs: []uuid.UUID{citationEvidence.ID, secondEvidence.ID},
	})
	s.Require().NoError(repeatErr)
	s.Equal(answer.ID, repeatedAnswer.ID)
	changedAnswer, reviseErr := service.PublishInvestigationAnswer(ctx, rez.InvestigationPublicationScope{InvestigationID: investigation.ID, AgentTurnID: turn.ID}, rez.PublishInvestigationAnswerParams{
		Title: "Retry increase", Body: "The supplied evidence does not establish a cause.", EvidenceIDs: evidenceIDs,
	})
	s.Require().NoError(reviseErr)
	s.NotEqual(answer.ID, changedAnswer.ID)
	s.Equal(answer.FindingID, changedAnswer.FindingID)
	s.Equal(firstInput.ID, *changedAnswer.UserInputID)
	s.Equal(2, tdb.Client(ctx).InvestigationFindingVersion.Query().Where(investigationfindingversion.HasFindingWith(investigationfinding.UserInputID(firstInput.ID))).CountX(ctx))
	s.Equal(2, tdb.Client(ctx).InvestigationOutputReference.Query().Where(investigationoutputreference.FindingVersionID(answer.ID)).CountX(ctx))
	s.True(tdb.Client(ctx).InvestigationUserInput.Query().Where(
		investigationuserinput.ID(waitingInput.ID), investigationuserinput.AgentTurnIDIsNil(),
	).ExistX(ctx))

	ordinaryReservedKey, reservedKeyErr := service.PublishInvestigationFinding(ctx, rez.InvestigationPublicationScope{InvestigationID: investigation.ID, AgentTurnID: turn.ID}, rez.PublishInvestigationFindingParams{
		Key: "answer:" + waitingInput.ID.String(), Title: "Not an answer", Body: "The answer tool owns this key space.",
	})
	s.Nil(ordinaryReservedKey)
	s.ErrorIs(reservedKeyErr, rez.ErrInvalidInput)

	tdb.Client(ctx).AgentTurn.UpdateOneID(turn.ID).SetStatus(agentturn.StatusCompleted).ExecX(ctx)
	secondTurn := s.createTurn(ctx, tdb, investigation.AgentSessionID, 2, agentturn.StatusRunning)
	tdb.Client(ctx).InvestigationEvidenceRevision.Create().
		SetInvestigationID(investigation.ID).
		SetExplanation("Only evidence changed before this turn.").
		SetKey("evidence-only-turn").
		SetAgentTurnID(secondTurn.ID).
		SaveX(ctx)
	noQuestionAnswer, noQuestionErr := service.PublishInvestigationAnswer(ctx, rez.InvestigationPublicationScope{InvestigationID: investigation.ID, AgentTurnID: secondTurn.ID}, rez.PublishInvestigationAnswerParams{
		Title: "Unassigned", Body: "This evidence-only turn cannot pick a queued question.",
	})
	s.Nil(noQuestionAnswer)
	s.ErrorIs(noQuestionErr, rez.ErrInvalidInput)
	noQuestionLatest, noQuestionLatestErr := service.ListInvestigationFindings(ctx, investigation.ID, ent.ListParams{})
	s.Require().NoError(noQuestionLatestErr)
	s.Equal(1, noQuestionLatest.Total)
	s.True(tdb.Client(ctx).InvestigationUserInput.Query().Where(
		investigationuserinput.ID(waitingInput.ID), investigationuserinput.AgentTurnIDIsNil(),
	).ExistX(ctx))
}

func (s *InvestigationServiceSuite) TestInvestigationFindingAndHypothesisSelectors() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	jobService := mocks.NewMockJobService(s.T())
	service, investigation, firstTurn := s.outputFixture(ctx, tdb, jobService)
	citationEntity, citationEvidence, _, citationEvent := s.outputCitationFixture(ctx, tdb, investigation.SystemAnalysisID)
	initial, initialErr := service.PublishInvestigationFinding(ctx, rez.InvestigationPublicationScope{InvestigationID: investigation.ID, AgentTurnID: firstTurn.ID}, rez.PublishInvestigationFindingParams{
		Key: "initial", Title: "Errors increased", Body: "The payment error rate rose.",
		EvidenceIDs: []uuid.UUID{citationEvidence.ID},
	})
	s.Require().NoError(initialErr)
	outsideAnalysisEntity := tdb.Client(ctx).KnowledgeEntity.Create().SetCategory(kne.CategoryContainer).SetKind("database").SaveX(ctx)
	_, outsideCitationErr := service.PublishInvestigationFinding(ctx, rez.InvestigationPublicationScope{InvestigationID: investigation.ID, AgentTurnID: firstTurn.ID}, rez.PublishInvestigationFindingParams{
		Key: "outside", Title: "Outside", Body: "This subject was not supplied in the analysis.",
		EvidenceIDs: []uuid.UUID{outsideAnalysisEntity.ID},
	})
	s.Error(outsideCitationErr)
	relationship := tdb.Client(ctx).KnowledgeRelationship.Create().
		SetPredicate("depends_on").
		SetSourceEntityID(citationEntity.ID).
		SetTargetEntityID(outsideAnalysisEntity.ID).
		SaveX(ctx)
	_, relationshipCitationErr := service.PublishInvestigationFinding(ctx, rez.InvestigationPublicationScope{InvestigationID: investigation.ID, AgentTurnID: firstTurn.ID}, rez.PublishInvestigationFindingParams{
		Key: "relationship-citation", Title: "Relationship is not a citation", Body: "Graph membership remains source data.",
		EvidenceIDs: []uuid.UUID{relationship.ID},
	})
	s.Error(relationshipCitationErr)
	_, eventCitationErr := service.PublishInvestigationFinding(ctx, rez.InvestigationPublicationScope{InvestigationID: investigation.ID, AgentTurnID: firstTurn.ID}, rez.PublishInvestigationFindingParams{
		Key: "event-citation", Title: "Event is not a citation", Body: "Events remain analysis source data, not publication citations.",
		EvidenceIDs: []uuid.UUID{citationEvent.ID},
	})
	s.Error(eventCitationErr)

	_, missingEvidenceErr := service.PublishInvestigationFinding(ctx, rez.InvestigationPublicationScope{InvestigationID: investigation.ID, AgentTurnID: firstTurn.ID}, rez.PublishInvestigationFindingParams{
		Key: "missing", Title: "Missing", Body: "This evidence is not attached to the analysis.",
		EvidenceIDs: []uuid.UUID{uuid.New()},
	})
	s.Error(missingEvidenceErr)
	_, zeroEvidenceErr := service.PublishInvestigationFinding(ctx, rez.InvestigationPublicationScope{InvestigationID: investigation.ID, AgentTurnID: firstTurn.ID}, rez.PublishInvestigationFindingParams{
		Key: "zero", Title: "Zero evidence ID", Body: "An empty UUID is invalid.",
		EvidenceIDs: []uuid.UUID{uuid.Nil},
	})
	s.Error(zeroEvidenceErr)
	s.Equal(1, tdb.Client(ctx).InvestigationFinding.Query().Where(investigationfinding.InvestigationID(investigation.ID)).CountX(ctx))
	s.Equal(1, tdb.Client(ctx).InvestigationFindingVersion.Query().Where(investigationfindingversion.HasFindingWith(investigationfinding.InvestigationID(investigation.ID))).CountX(ctx))

	statuses := []investigationhypothesisversion.Status{
		investigationhypothesisversion.StatusOpen,
		investigationhypothesisversion.StatusSupported,
		investigationhypothesisversion.StatusDisproven,
		investigationhypothesisversion.StatusInconclusive,
	}
	hypothesisVersions := make([]*rez.InvestigationHypothesisVersion, 0, len(statuses))
	for _, status := range statuses {
		version, publishErr := service.PublishInvestigationHypothesis(ctx, rez.InvestigationPublicationScope{InvestigationID: investigation.ID, AgentTurnID: firstTurn.ID}, rez.PublishInvestigationHypothesisParams{
			Key: "hypothesis-" + string(status), Title: "Possible explanation", Justification: "Evidence is limited, so this remains uncertain.", Status: status,
		})
		s.Require().NoError(publishErr)
		hypothesisVersions = append(hypothesisVersions, version)
	}
	invalidStatus, invalidStatusErr := service.PublishInvestigationHypothesis(ctx, rez.InvestigationPublicationScope{InvestigationID: investigation.ID, AgentTurnID: firstTurn.ID}, rez.PublishInvestigationHypothesisParams{
		Key: "bad-status", Title: "Possible explanation", Justification: "No evidence yet.", Status: "maybe",
	})
	s.Nil(invalidStatus)
	s.Error(invalidStatusErr)
	for _, version := range hypothesisVersions {
		readVersion, readErr := service.GetInvestigationHypothesisVersion(ctx, investigation.ID, version.ID)
		s.Require().NoError(readErr)
		s.Equal(version.Status, readVersion.Status)
		s.Equal(agentturn.StatusRunning, readVersion.TurnStatus)
	}
	hypothesisPage, pageErr := service.ListInvestigationHypotheses(ctx, investigation.ID, ent.ListParams{Page: 1, PageSize: 2})
	s.Require().NoError(pageErr)
	s.Equal(4, hypothesisPage.Total)
	s.Len(hypothesisPage.Data, 2)
	s.Less(hypothesisPage.Page*hypothesisPage.PageSize, hypothesisPage.Total)

	tdb.Client(ctx).AgentTurn.UpdateOneID(firstTurn.ID).SetStatus(agentturn.StatusCompleted).ExecX(ctx)
	secondTurn := s.createTurn(ctx, tdb, investigation.AgentSessionID, 2, agentturn.StatusRunning)
	invalidator, invalidatorErr := service.PublishInvestigationFinding(ctx, rez.InvestigationPublicationScope{InvestigationID: investigation.ID, AgentTurnID: secondTurn.ID}, rez.PublishInvestigationFindingParams{
		Key: "reasoning", Title: "Candidate explanation", Body: "The deployment may have contributed.",
		FindingReferences: []rez.FindingVersionReference{{VersionID: initial.ID, Relation: "invalidates"}},
	})
	s.Require().NoError(invalidatorErr)
	selected, selectedErr := service.ListInvestigationFindings(ctx, investigation.ID, ent.ListParams{Page: 1, PageSize: 1})
	s.Require().NoError(selectedErr)
	s.Equal(2, selected.Total)
	s.Less(selected.Page*selected.PageSize, selected.Total)
	firstVersion, firstVersionErr := service.GetInvestigationFindingVersion(ctx, investigation.ID, initial.ID)
	s.Require().NoError(firstVersionErr)
	s.Equal([]uuid.UUID{invalidator.ID}, firstVersion.InvalidatedByVersionIDs)
	s.Equal([]rez.FindingVersionReference{{VersionID: initial.ID, Relation: "invalidates"}}, invalidator.FindingReferences)

	tdb.Client(ctx).AgentTurn.UpdateOneID(secondTurn.ID).SetStatus(agentturn.StatusCompleted).ExecX(ctx)
	thirdTurn := s.createTurn(ctx, tdb, investigation.AgentSessionID, 3, agentturn.StatusRunning)
	currentReasoning, reasoningRevErr := service.PublishInvestigationFinding(ctx, rez.InvestigationPublicationScope{InvestigationID: investigation.ID, AgentTurnID: thirdTurn.ID}, rez.PublishInvestigationFindingParams{
		Key: "reasoning", Title: "Candidate explanation", Body: "Evidence is insufficient to decide.",
	})
	s.Require().NoError(reasoningRevErr)
	s.NotEqual(invalidator.ID, currentReasoning.ID)
	firstVersionAfterRevision, firstVersionAfterRevisionErr := service.GetInvestigationFindingVersion(ctx, investigation.ID, initial.ID)
	s.Require().NoError(firstVersionAfterRevisionErr)
	s.Empty(firstVersionAfterRevision.InvalidatedByVersionIDs)
	currentPage, currentPageErr := service.ListInvestigationFindings(ctx, investigation.ID, ent.ListParams{Page: 1, PageSize: 25})
	s.Require().NoError(currentPageErr)
	s.Equal(2, currentPage.Total)
	var currentReasoningID uuid.UUID
	for _, item := range currentPage.Data {
		if item.Key == "reasoning" {
			currentReasoningID = item.ID
		}
	}
	s.Equal(currentReasoning.ID, currentReasoningID)
	_, wrongParentErr := service.GetInvestigationFindingVersion(ctx, uuid.New(), initial.ID)
	s.Error(wrongParentErr)
}

func (s *InvestigationServiceSuite) TestConcurrentInvestigationReportWritesDeduplicate() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	jobService := mocks.NewMockJobService(s.T())
	service, investigation, turn := s.outputFixture(ctx, tdb, jobService)
	_, citationEvidence, _, _ := s.outputCitationFixture(ctx, tdb, investigation.SystemAnalysisID)
	params := rez.PublishInvestigationReportParams{
		Text: "Same report", EvidenceIDs: []uuid.UUID{citationEvidence.ID},
	}
	type writeResult struct {
		publication *rez.InvestigationReportResult
		publishErr  error
	}
	results := make(chan writeResult, 2)
	var workers sync.WaitGroup
	for range 2 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			publication, publishErr := service.PublishInvestigationReport(ctx, rez.InvestigationPublicationScope{InvestigationID: investigation.ID, AgentTurnID: turn.ID}, params)
			results <- writeResult{publication: publication, publishErr: publishErr}
		}()
	}
	workers.Wait()
	close(results)
	var publicationID uuid.UUID
	for result := range results {
		s.Require().NoError(result.publishErr)
		s.Require().NotNil(result.publication)
		if publicationID == uuid.Nil {
			publicationID = result.publication.ID
		}
		s.Equal(publicationID, result.publication.ID)
	}
	s.Equal(1, tdb.Client(ctx).InvestigationReport.Query().Where(investigationreport.InvestigationID(investigation.ID)).CountX(ctx))
	s.Equal(1, tdb.Client(ctx).InvestigationOutputReference.Query().Where(investigationoutputreference.ReportID(publicationID)).CountX(ctx))
}

func (s *InvestigationServiceSuite) TestInvestigationPublicationCannotRaceTurnCompletion() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	jobService := mocks.NewMockJobService(s.T())
	service, investigation, turn := s.outputFixture(ctx, tdb, jobService)
	start := make(chan struct{})
	var workers sync.WaitGroup
	var publication *rez.InvestigationReportResult
	var publishErr error
	var completionErr error
	workers.Add(2)
	go func() {
		defer workers.Done()
		<-start
		publication, publishErr = service.PublishInvestigationReport(ctx, rez.InvestigationPublicationScope{InvestigationID: investigation.ID, AgentTurnID: turn.ID}, rez.PublishInvestigationReportParams{Text: "Complete race."})
	}()
	go func() {
		defer workers.Done()
		<-start
		completionErr = tdb.Client(ctx).AgentTurn.UpdateOneID(turn.ID).SetStatus(agentturn.StatusCompleted).Exec(ctx)
	}()
	close(start)
	workers.Wait()
	s.Require().NoError(completionErr)
	s.Equal(agentturn.StatusCompleted, tdb.Client(ctx).AgentTurn.GetX(ctx, turn.ID).Status)
	publicationCount := tdb.Client(ctx).InvestigationReport.Query().Where(investigationreport.InvestigationID(investigation.ID)).CountX(ctx)
	if publishErr == nil {
		s.Require().NotNil(publication)
		s.Equal(1, publicationCount)
	} else {
		s.ErrorIs(publishErr, rez.ErrConflict)
		s.Nil(publication)
		s.Zero(publicationCount)
	}
}
