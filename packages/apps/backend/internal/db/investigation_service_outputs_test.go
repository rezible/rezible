package db

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/agentturn"
	"github.com/rezible/rezible/ent/investigationfinding"
	"github.com/rezible/rezible/ent/investigationfindingversion"
	"github.com/rezible/rezible/ent/investigationfindingversionlink"
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

type investigationCitationFixture struct {
	entity         *ent.KnowledgeEntity
	evidence       *ent.KnowledgeEvidence
	secondEvidence *ent.KnowledgeEvidence
	event          *ent.NormalizedEvent
}

func (s *InvestigationServiceSuite) outputCitationFixture(ctx context.Context, tdb rez.Database, analysisID uuid.UUID) investigationCitationFixture {
	client := tdb.Client(ctx)
	now := time.Now().UTC().Truncate(time.Second)
	encodedAttributes, encodeErr := projections.EncodeAttributes(struct{}{})
	s.Require().NoError(encodeErr)

	createEvent := client.NormalizedEvent.Create().
		SetProvider("investigation-tests").
		SetProviderNamespace("investigation-tests").
		SetProviderResourceRef(uuid.NewString()).
		SetProviderEventSource("investigation-tests").
		SetProviderEventRef(uuid.NewString()).
		SetKind("deployment").
		SetAttributes(encodedAttributes).
		SetOccurredAt(now).
		SetReceivedAt(now)
	event, eventErr := createEvent.Save(ctx)
	s.Require().NoError(eventErr)

	createEntity := client.KnowledgeEntity.Create().
		SetCategory(kne.CategoryContainer).
		SetKind("service")
	entity, entityErr := createEntity.Save(ctx)
	s.Require().NoError(entityErr)

	attachAnalysisEntity := client.SystemAnalysisEntity.Create().
		SetAnalysisID(analysisID).
		SetKnowledgeEntityID(entity.ID)
	attachAnalysisEntityErr := attachAnalysisEntity.Exec(ctx)
	s.Require().NoError(attachAnalysisEntityErr)

	createAlias := client.KnowledgeSubjectAlias.Create().
		SetSubjectKind(ksa.SubjectKindEntity).
		SetProvider("investigation-tests").
		SetProviderNamespace("investigation-tests").
		SetProviderResourceRef(uuid.NewString()).
		SetEntityID(entity.ID)
	alias, aliasErr := createAlias.Save(ctx)
	s.Require().NoError(aliasErr)

	createEvidence := client.KnowledgeEvidence.Create().
		SetEventID(event.ID).
		SetSubjectAliasID(alias.ID).
		SetKind(kev.KindDeleted).
		SetAssertion("source_deleted").
		SetEffectiveAt(now).
		SetSubjectState(schematypes.KnowledgeGraphSubjectState{
			DisplayName: "Payment service",
		})
	evidence, evidenceErr := createEvidence.Save(ctx)
	s.Require().NoError(evidenceErr)

	createSecondAlias := client.KnowledgeSubjectAlias.Create().
		SetSubjectKind(ksa.SubjectKindEntity).
		SetProvider("investigation-tests").
		SetProviderNamespace("investigation-tests").
		SetProviderResourceRef(uuid.NewString()).
		SetEntityID(entity.ID)
	secondAlias, secondAliasErr := createSecondAlias.Save(ctx)
	s.Require().NoError(secondAliasErr)

	createSecondEvidence := client.KnowledgeEvidence.Create().
		SetEventID(event.ID).
		SetSubjectAliasID(secondAlias.ID).
		SetKind(kev.KindObserved).
		SetAssertion("source_observed").
		SetEffectiveAt(now).
		SetSubjectState(schematypes.KnowledgeGraphSubjectState{
			DisplayName: "Payment service",
		})
	secondEvidence, secondEvidenceErr := createSecondEvidence.Save(ctx)
	s.Require().NoError(secondEvidenceErr)

	createEntry := client.SystemAnalysisEntry.Create().
		SetAnalysisID(analysisID).
		SetKind(systemanalysisentry.KindObservation).
		SetTitle("The payment service changed").
		SetBody("A stored source record remains attached to this observation.")
	entry, entryErr := createEntry.Save(ctx)
	s.Require().NoError(entryErr)

	attachEntryEntity := client.SystemAnalysisEntrySubject.Create().
		SetEntryID(entry.ID).
		SetRole("primary").
		SetKnowledgeEntityID(entity.ID)
	attachEntryEntityErr := attachEntryEntity.Exec(ctx)
	s.Require().NoError(attachEntryEntityErr)

	attachEvidence := client.SystemAnalysisEntrySubject.Create().
		SetEntryID(entry.ID).
		SetRole("evidence_for").
		SetKnowledgeEvidenceID(evidence.ID)
	attachEvidenceErr := attachEvidence.Exec(ctx)
	s.Require().NoError(attachEvidenceErr)

	attachSecondEvidence := client.SystemAnalysisEntrySubject.Create().
		SetEntryID(entry.ID).
		SetRole("evidence_for").
		SetKnowledgeEvidenceID(secondEvidence.ID)
	attachSecondEvidenceErr := attachSecondEvidence.Exec(ctx)
	s.Require().NoError(attachSecondEvidenceErr)

	return investigationCitationFixture{
		entity:         entity,
		evidence:       evidence,
		secondEvidence: secondEvidence,
		event:          event,
	}
}

func (s *InvestigationServiceSuite) TestInvestigationReportPublicationSelectionAndRetry() {
	ctx, tdb := s.SetupTestDatabase()
	jobService := mocks.NewMockJobService(s.T())
	service, investigation, firstTurn := s.outputFixture(ctx, tdb, jobService)

	firstTurnScope := rez.InvestigationPublicationScope{
		InvestigationID: investigation.ID,
		AgentTurnID:     firstTurn.ID,
	}
	reportAParams := rez.PublishInvestigationReportParams{
		Text: " Initial report. ",
	}

	reportA, publishAErr := service.PublishInvestigationReport(ctx, firstTurnScope, reportAParams)
	s.Require().NoError(publishAErr)
	s.Equal("Initial report.", reportA.Text)
	s.Empty(reportA.Edges.OutputReferences)
	s.Equal(agentturn.StatusRunning, reportA.Edges.AgentTurn.Status)
	firstTurnCompleted := tdb.Client(ctx).AgentTurn.UpdateOneID(firstTurn.ID).
		SetStatus(agentturn.StatusCompleted)
	firstTurnCompletedErr := firstTurnCompleted.Exec(ctx)
	s.Require().NoError(firstTurnCompletedErr)

	secondTurn := s.createTurn(ctx, tdb, investigation.AgentSessionID, 2, agentturn.StatusRunning)
	secondTurnScope := rez.InvestigationPublicationScope{
		InvestigationID: investigation.ID,
		AgentTurnID:     secondTurn.ID,
	}
	reportBParams := rez.PublishInvestigationReportParams{
		Text: " Revised report. ",
	}

	reportB, publishBErr := service.PublishInvestigationReport(ctx, secondTurnScope, reportBParams)
	s.Require().NoError(publishBErr)

	repeatedBParams := rez.PublishInvestigationReportParams{
		Text: "Revised report.",
	}

	repeatedB, repeatBErr := service.PublishInvestigationReport(ctx, secondTurnScope, repeatedBParams)
	s.Require().NoError(repeatBErr)
	s.Equal(reportB.ID, repeatedB.ID)
	s.Equal(agentturn.StatusRunning, repeatedB.Edges.AgentTurn.Status)
	latest, latestErr := service.ReadInvestigationReport(ctx, rez.ReadInvestigationReportParams{InvestigationID: investigation.ID})
	s.Require().NoError(latestErr)
	s.Equal(reportB.ID, latest.ID)
	completedParams := rez.ReadInvestigationReportParams{
		InvestigationID: investigation.ID,
		Selection:       rez.InvestigationReportSelectionCompleted,
	}

	completed, completedErr := service.ReadInvestigationReport(ctx, completedParams)
	s.Require().NoError(completedErr)
	s.Equal(reportA.ID, completed.ID)

	workerMessages := mocks.NewMockMessageQueue(s.T())
	workerMessages.EXPECT().
		Publish(mock.Anything, mock.IsType(rezai.AgentTurnUpdated{})).
		Return(nil).
		Maybe()
	worker := &InvokeAgentTurnWorker{
		db:     tdb,
		msgs:   workerMessages,
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	automaticRetryErr := worker.saveInvocationResult(ctx, makeAgentTurnJob(secondTurn, 1), nil, nil, errors.New("temporary model failure"))
	s.Require().NoError(automaticRetryErr)

	queued, queuedErr := service.ReadInvestigationReport(ctx, rez.ReadInvestigationReportParams{InvestigationID: investigation.ID})
	s.Require().NoError(queuedErr)
	s.Equal(reportA.ID, queued.ID)
	secondTurnRunning := tdb.Client(ctx).AgentTurn.UpdateOneID(secondTurn.ID).
		SetStatus(agentturn.StatusRunning)
	secondTurnRunningErr := secondTurnRunning.Exec(ctx)
	s.Require().NoError(secondTurnRunningErr)

	automaticRetryVisible, automaticRetryVisibleErr := service.ReadInvestigationReport(ctx, rez.ReadInvestigationReportParams{InvestigationID: investigation.ID})
	s.Require().NoError(automaticRetryVisibleErr)
	s.Equal(reportB.ID, automaticRetryVisible.ID)

	secondTurnFailed := tdb.Client(ctx).AgentTurn.UpdateOneID(secondTurn.ID).
		SetStatus(agentturn.StatusFailed)
	secondTurnFailedErr := secondTurnFailed.Exec(ctx)
	s.Require().NoError(secondTurnFailedErr)

	retryJobs := mocks.NewMockJobService(s.T())
	retryMessages := mocks.NewMockMessageQueue(s.T())
	retryMessages.EXPECT().
		Publish(mock.Anything, mock.IsType(rezai.AgentTurnUpdated{})).
		Return(nil).
		Once()
	retryJobs.EXPECT().
		Insert(mock.Anything, jobs.InvokeAgentTurn{
			AgentSessionID: investigation.AgentSessionID,
			AgentTurnID:    secondTurn.ID,
		}, mock.Anything).
		Return(&rivertype.JobInsertResult{
			Job: &rivertype.JobRow{
				ID: 5002,
			},
		}, nil).
		Once()
	sessionService := &AiAgentSessionService{
		db:   tdb,
		jobs: retryJobs,
		msgs: retryMessages,
	}
	retriedTurn, retryErr := sessionService.RetryAgentTurn(ctx, secondTurn.ID)
	s.Require().NoError(retryErr)
	s.Equal(agentturn.StatusQueued, retriedTurn.Status)
	queuedAgain, queuedAgainErr := service.ReadInvestigationReport(ctx, rez.ReadInvestigationReportParams{InvestigationID: investigation.ID})
	s.Require().NoError(queuedAgainErr)
	s.Equal(reportA.ID, queuedAgain.ID)
	startRetriedTurn := tdb.Client(ctx).AgentTurn.UpdateOneID(secondTurn.ID).
		SetStatus(agentturn.StatusRunning)
	startRetriedTurnErr := startRetriedTurn.Exec(ctx)
	s.Require().NoError(startRetriedTurnErr)

	sharedRetryVisible, sharedRetryVisibleErr := service.ReadInvestigationReport(ctx, rez.ReadInvestigationReportParams{InvestigationID: investigation.ID})
	s.Require().NoError(sharedRetryVisibleErr)
	s.Equal(reportB.ID, sharedRetryVisible.ID)

	secondTurnCompleted := tdb.Client(ctx).AgentTurn.UpdateOneID(secondTurn.ID).
		SetStatus(agentturn.StatusCompleted)
	secondTurnCompletedErr := secondTurnCompleted.Exec(ctx)
	s.Require().NoError(secondTurnCompletedErr)

	completedBParams := rez.ReadInvestigationReportParams{
		InvestigationID: investigation.ID,
		Selection:       rez.InvestigationReportSelectionCompleted,
	}

	completedB, completedBErr := service.ReadInvestigationReport(ctx, completedBParams)
	s.Require().NoError(completedBErr)
	s.Equal(reportB.ID, completedB.ID)
	thirdTurn := s.createTurn(ctx, tdb, investigation.AgentSessionID, 3, agentturn.StatusRunning)
	thirdTurnScope := rez.InvestigationPublicationScope{
		InvestigationID: investigation.ID,
		AgentTurnID:     thirdTurn.ID,
	}
	newTurnBParams := rez.PublishInvestigationReportParams{
		Text: "Revised report.",
	}

	newTurnB, newTurnBErr := service.PublishInvestigationReport(ctx, thirdTurnScope, newTurnBParams)
	s.Require().NoError(newTurnBErr)
	s.NotEqual(reportB.ID, newTurnB.ID)
}

func (s *InvestigationServiceSuite) TestInvestigationReportSummary() {
	ctx, tdb := s.SetupTestDatabase()
	jobService := mocks.NewMockJobService(s.T())
	service, investigation, turn := s.outputFixture(ctx, tdb, jobService)
	scope := rez.InvestigationPublicationScope{
		InvestigationID: investigation.ID,
		AgentTurnID:     turn.ID,
	}

	tooLong := rez.PublishInvestigationReportParams{
		Text:    "# Report",
		Summary: strings.Repeat("é", 401),
	}
	_, tooLongErr := service.PublishInvestigationReport(ctx, scope, tooLong)
	s.Require().ErrorIs(tooLongErr, rez.ErrInvalidInput)

	first, firstErr := service.PublishInvestigationReport(ctx, scope, rez.PublishInvestigationReportParams{
		Text:    "# Report\n\nRedis pressure rose after the deployment.",
		Summary: " Redis pressure is the strongest signal; the cause is unconfirmed. ",
	})
	s.Require().NoError(firstErr)
	s.Equal("Redis pressure is the strongest signal; the cause is unconfirmed.", first.Summary)

	read, readErr := service.ReadInvestigationReport(ctx, rez.ReadInvestigationReportParams{InvestigationID: investigation.ID})
	s.Require().NoError(readErr)
	s.Equal(first.ID, read.ID)
	s.Equal(first.Summary, read.Summary)

	revised, revisedErr := service.PublishInvestigationReport(ctx, scope, rez.PublishInvestigationReportParams{
		Text:    "# Report\n\nRedis pressure rose after the deployment.",
		Summary: "The deployment is the likely cause.",
	})
	s.Require().NoError(revisedErr)
	s.NotEqual(first.ID, revised.ID)
	s.Equal("The deployment is the likely cause.", revised.Summary)
}

func (s *InvestigationServiceSuite) TestInvestigationAnswerOwnershipAndRevisions() {
	ctx, tdb := s.SetupTestDatabase()
	jobService := mocks.NewMockJobService(s.T())
	service, investigation, turn := s.outputFixture(ctx, tdb, jobService)
	turnScope := rez.InvestigationPublicationScope{
		InvestigationID: investigation.ID,
		AgentTurnID:     turn.ID,
	}
	initialAnswerParams := rez.PublishInvestigationAnswerParams{
		Title: "Initial turn cannot answer",
		Body:  "There is no directly assigned user input.",
	}

	initialAnswer, initialAnswerErr := service.PublishInvestigationAnswer(ctx, turnScope, initialAnswerParams)
	s.Nil(initialAnswer)
	s.ErrorIs(initialAnswerErr, rez.ErrInvalidInput)
	userCtx := s.userContext(tdb, ctx)
	userID, userIDSet := execution.GetContext(userCtx).UserID()
	s.Require().True(userIDSet)
	createFirstInput := tdb.Client(ctx).InvestigationUserInput.Create().
		SetInvestigationID(investigation.ID).
		SetUserID(userID).
		SetText("Why did retries increase?").
		SetKey("question-one").
		SetAgentTurnID(turn.ID)
	firstInput, firstInputErr := createFirstInput.Save(ctx)
	s.Require().NoError(firstInputErr)

	createWaitingInput := tdb.Client(ctx).InvestigationUserInput.Create().
		SetInvestigationID(investigation.ID).
		SetUserID(userID).
		SetText("What changed after deployment?").
		SetKey("question-two")
	waitingInput, waitingInputErr := createWaitingInput.Save(ctx)
	s.Require().NoError(waitingInputErr)

	citations := s.outputCitationFixture(ctx, tdb, investigation.SystemAnalysisID)
	evidenceIDs := []uuid.UUID{citations.secondEvidence.ID, citations.evidence.ID, citations.evidence.ID}
	answerParams := rez.PublishInvestigationAnswerParams{
		Title:       "Retry increase",
		Body:        "The supplied analysis has attached evidence records, including one deleted-kind record.",
		EvidenceIDs: evidenceIDs,
	}

	answer, publishErr := service.PublishInvestigationAnswer(ctx, turnScope, answerParams)
	s.Require().NoError(publishErr)
	s.Equal(firstInput.ID, *answer.Edges.Finding.UserInputID)
	s.Equal("answer:"+firstInput.ID.String(), answer.Edges.Finding.Key)
	expectedEvidenceIDs := []uuid.UUID{citations.evidence.ID, citations.secondEvidence.ID}
	sort.Slice(expectedEvidenceIDs, func(i, j int) bool {
		return expectedEvidenceIDs[i].String() < expectedEvidenceIDs[j].String()
	})
	s.Equal(expectedEvidenceIDs, ent.InvestigationOutputReferences(answer.Edges.OutputReferences).KnowledgeEvidenceIDs())
	repeatedAnswerParams := rez.PublishInvestigationAnswerParams{
		Title:       " Retry increase ",
		Body:        "The supplied analysis has attached evidence records, including one deleted-kind record. ",
		EvidenceIDs: []uuid.UUID{citations.evidence.ID, citations.secondEvidence.ID},
	}

	repeatedAnswer, repeatErr := service.PublishInvestigationAnswer(ctx, turnScope, repeatedAnswerParams)
	s.Require().NoError(repeatErr)
	s.Equal(answer.ID, repeatedAnswer.ID)
	changedAnswerParams := rez.PublishInvestigationAnswerParams{
		Title:       "Retry increase",
		Body:        "The supplied evidence does not establish a cause.",
		EvidenceIDs: evidenceIDs,
	}

	changedAnswer, reviseErr := service.PublishInvestigationAnswer(ctx, turnScope, changedAnswerParams)
	s.Require().NoError(reviseErr)
	s.NotEqual(answer.ID, changedAnswer.ID)
	s.Equal(answer.FindingID, changedAnswer.FindingID)
	s.Equal(firstInput.ID, *changedAnswer.Edges.Finding.UserInputID)
	queryInvestigationFindingVersion := tdb.Client(ctx).InvestigationFindingVersion.Query().
		Where(investigationfindingversion.HasFindingWith(investigationfinding.UserInputID(firstInput.ID)))
	investigationFindingVersionCount, investigationFindingVersionCountErr := queryInvestigationFindingVersion.Count(ctx)
	s.Require().NoError(investigationFindingVersionCountErr)

	s.Equal(2, investigationFindingVersionCount)
	queryInvestigationOutputReference := tdb.Client(ctx).InvestigationOutputReference.Query().
		Where(investigationoutputreference.FindingVersionID(answer.ID))
	investigationOutputReferenceCount, investigationOutputReferenceCountErr := queryInvestigationOutputReference.Count(ctx)
	s.Require().NoError(investigationOutputReferenceCountErr)

	s.Equal(2, investigationOutputReferenceCount)
	queryInvestigationUserInput := tdb.Client(ctx).InvestigationUserInput.Query().
		Where(
			investigationuserinput.ID(waitingInput.ID), investigationuserinput.AgentTurnIDIsNil(),
		)
	investigationUserInputExists, investigationUserInputExistsErr := queryInvestigationUserInput.Exist(ctx)
	s.Require().NoError(investigationUserInputExistsErr)

	s.True(investigationUserInputExists)

	ordinaryReservedKeyParams := rez.PublishInvestigationFindingParams{
		Key:   "answer:" + waitingInput.ID.String(),
		Title: "Not an answer",
		Body:  "The answer tool owns this key space.",
	}

	ordinaryReservedKey, reservedKeyErr := service.PublishInvestigationFinding(ctx, turnScope, ordinaryReservedKeyParams)
	s.Nil(ordinaryReservedKey)
	s.ErrorIs(reservedKeyErr, rez.ErrInvalidInput)

	turnCompleted := tdb.Client(ctx).AgentTurn.UpdateOneID(turn.ID).
		SetStatus(agentturn.StatusCompleted)
	turnCompletedErr := turnCompleted.Exec(ctx)
	s.Require().NoError(turnCompletedErr)

	secondTurn := s.createTurn(ctx, tdb, investigation.AgentSessionID, 2, agentturn.StatusRunning)
	createInvestigationEvidenceRevision := tdb.Client(ctx).InvestigationEvidenceRevision.Create().
		SetInvestigationID(investigation.ID).
		SetExplanation("Only evidence changed before this turn.").
		SetKey("evidence-only-turn").
		SetAgentTurnID(secondTurn.ID)
	createEvidenceRevisionErr := createInvestigationEvidenceRevision.Exec(ctx)
	s.Require().NoError(createEvidenceRevisionErr)

	secondTurnScope := rez.InvestigationPublicationScope{
		InvestigationID: investigation.ID,
		AgentTurnID:     secondTurn.ID,
	}
	noQuestionAnswerParams := rez.PublishInvestigationAnswerParams{
		Title: "Unassigned",
		Body:  "This evidence-only turn cannot pick a queued question.",
	}

	noQuestionAnswer, noQuestionErr := service.PublishInvestigationAnswer(ctx, secondTurnScope, noQuestionAnswerParams)
	s.Nil(noQuestionAnswer)
	s.ErrorIs(noQuestionErr, rez.ErrInvalidInput)
	noQuestionLatest, noQuestionLatestErr := service.ListInvestigationFindings(ctx, rez.ListInvestigationFindingsParams{InvestigationID: investigation.ID})
	s.Require().NoError(noQuestionLatestErr)
	s.Equal(1, noQuestionLatest.Total)
	queryInvestigationUserInput2 := tdb.Client(ctx).InvestigationUserInput.Query().
		Where(
			investigationuserinput.ID(waitingInput.ID), investigationuserinput.AgentTurnIDIsNil(),
		)
	investigationUserInput2Exists, investigationUserInput2ExistsErr := queryInvestigationUserInput2.Exist(ctx)
	s.Require().NoError(investigationUserInput2ExistsErr)

	s.True(investigationUserInput2Exists)
}

func (s *InvestigationServiceSuite) TestInvestigationFindingCitationValidation() {
	ctx, tdb := s.SetupTestDatabase()
	jobService := mocks.NewMockJobService(s.T())
	service, investigation, firstTurn := s.outputFixture(ctx, tdb, jobService)
	citations := s.outputCitationFixture(ctx, tdb, investigation.SystemAnalysisID)
	firstTurnScope := rez.InvestigationPublicationScope{
		InvestigationID: investigation.ID,
		AgentTurnID:     firstTurn.ID,
	}
	initialParams := rez.PublishInvestigationFindingParams{
		Key:         "initial",
		Title:       "Errors increased",
		Body:        "The payment error rate rose.",
		EvidenceIDs: []uuid.UUID{citations.evidence.ID},
	}

	_, initialErr := service.PublishInvestigationFinding(ctx, firstTurnScope, initialParams)
	s.Require().NoError(initialErr)

	createOutsideAnalysisEntity := tdb.Client(ctx).KnowledgeEntity.Create().
		SetCategory(kne.CategoryContainer).
		SetKind("database")
	outsideAnalysisEntity, outsideAnalysisEntityErr := createOutsideAnalysisEntity.Save(ctx)
	s.Require().NoError(outsideAnalysisEntityErr)

	outsideCitationParams := rez.PublishInvestigationFindingParams{
		Key:         "outside",
		Title:       "Outside",
		Body:        "This subject was not supplied in the analysis.",
		EvidenceIDs: []uuid.UUID{outsideAnalysisEntity.ID},
	}

	_, outsideCitationErr := service.PublishInvestigationFinding(ctx, firstTurnScope, outsideCitationParams)
	s.Error(outsideCitationErr)
	createRelationship := tdb.Client(ctx).KnowledgeRelationship.Create().
		SetPredicate("depends_on").
		SetSourceEntityID(citations.entity.ID).
		SetTargetEntityID(outsideAnalysisEntity.ID)
	relationship, relationshipErr := createRelationship.Save(ctx)
	s.Require().NoError(relationshipErr)

	relationshipCitationParams := rez.PublishInvestigationFindingParams{
		Key:         "relationship-citation",
		Title:       "Relationship is not a citation",
		Body:        "Graph membership remains source data.",
		EvidenceIDs: []uuid.UUID{relationship.ID},
	}

	_, relationshipCitationErr := service.PublishInvestigationFinding(ctx, firstTurnScope, relationshipCitationParams)
	s.Error(relationshipCitationErr)
	eventCitationParams := rez.PublishInvestigationFindingParams{
		Key:         "event-citation",
		Title:       "Event is not a citation",
		Body:        "Events remain analysis source data, not publication citations.",
		EvidenceIDs: []uuid.UUID{citations.event.ID},
	}

	_, eventCitationErr := service.PublishInvestigationFinding(ctx, firstTurnScope, eventCitationParams)
	s.Error(eventCitationErr)

	missingEvidenceParams := rez.PublishInvestigationFindingParams{
		Key:         "missing",
		Title:       "Missing",
		Body:        "This evidence is not attached to the analysis.",
		EvidenceIDs: []uuid.UUID{uuid.New()},
	}

	_, missingEvidenceErr := service.PublishInvestigationFinding(ctx, firstTurnScope, missingEvidenceParams)
	s.Error(missingEvidenceErr)
	zeroEvidenceParams := rez.PublishInvestigationFindingParams{
		Key:         "zero",
		Title:       "Zero evidence ID",
		Body:        "An empty UUID is invalid.",
		EvidenceIDs: []uuid.UUID{uuid.Nil},
	}

	_, zeroEvidenceErr := service.PublishInvestigationFinding(ctx, firstTurnScope, zeroEvidenceParams)
	s.Error(zeroEvidenceErr)
	queryInvestigationFinding := tdb.Client(ctx).InvestigationFinding.Query().
		Where(investigationfinding.InvestigationID(investigation.ID))
	investigationFindingCount, investigationFindingCountErr := queryInvestigationFinding.Count(ctx)
	s.Require().NoError(investigationFindingCountErr)

	s.Equal(1, investigationFindingCount)
	queryInvestigationFindingVersion := tdb.Client(ctx).InvestigationFindingVersion.Query().
		Where(investigationfindingversion.HasFindingWith(investigationfinding.InvestigationID(investigation.ID)))
	investigationFindingVersionCount, investigationFindingVersionCountErr := queryInvestigationFindingVersion.Count(ctx)
	s.Require().NoError(investigationFindingVersionCountErr)

	s.Equal(1, investigationFindingVersionCount)
}

func (s *InvestigationServiceSuite) TestInvestigationHypothesisStatusValidationAndSelection() {
	ctx, tdb := s.SetupTestDatabase()
	jobService := mocks.NewMockJobService(s.T())
	service, investigation, firstTurn := s.outputFixture(ctx, tdb, jobService)
	firstTurnScope := rez.InvestigationPublicationScope{
		InvestigationID: investigation.ID,
		AgentTurnID:     firstTurn.ID,
	}
	statuses := []investigationhypothesisversion.Status{
		investigationhypothesisversion.StatusOpen,
		investigationhypothesisversion.StatusSupported,
		investigationhypothesisversion.StatusDisproven,
		investigationhypothesisversion.StatusInconclusive,
	}
	hypothesisVersions := make([]*ent.InvestigationHypothesisVersion, 0, len(statuses))
	for _, status := range statuses {
		versionParams := rez.PublishInvestigationHypothesisParams{
			Key:           "hypothesis-" + string(status),
			Title:         "Possible explanation",
			Justification: "Evidence is limited, so this remains uncertain.",
			Status:        status,
		}

		version, publishErr := service.PublishInvestigationHypothesis(ctx, firstTurnScope, versionParams)
		s.Require().NoError(publishErr)
		hypothesisVersions = append(hypothesisVersions, version)
	}
	invalidStatusParams := rez.PublishInvestigationHypothesisParams{
		Key:           "bad-status",
		Title:         "Possible explanation",
		Justification: "No evidence yet.",
		Status:        "maybe",
	}

	invalidStatus, invalidStatusErr := service.PublishInvestigationHypothesis(ctx, firstTurnScope, invalidStatusParams)
	s.Nil(invalidStatus)
	s.Error(invalidStatusErr)
	for _, version := range hypothesisVersions {
		readVersion, readErr := service.GetInvestigationHypothesisVersion(ctx, investigation.ID, version.ID)
		s.Require().NoError(readErr)
		s.Equal(version.Status, readVersion.Status)
		s.Equal(agentturn.StatusRunning, readVersion.Edges.AgentTurn.Status)
	}
	hypothesisPageParams := rez.ListInvestigationHypothesesParams{
		ListParams:      ent.ListParams{Page: 1, PageSize: 2},
		InvestigationID: investigation.ID,
	}

	hypothesisPage, pageErr := service.ListInvestigationHypotheses(ctx, hypothesisPageParams)
	s.Require().NoError(pageErr)
	s.Equal(4, hypothesisPage.Total)
	s.Len(hypothesisPage.Data, 2)
	s.Less(hypothesisPage.Page*hypothesisPage.PageSize, hypothesisPage.Total)
}

func (s *InvestigationServiceSuite) TestInvestigationFindingPaginationInvalidationAndRevision() {
	ctx, tdb := s.SetupTestDatabase()
	jobService := mocks.NewMockJobService(s.T())
	service, investigation, firstTurn := s.outputFixture(ctx, tdb, jobService)
	citations := s.outputCitationFixture(ctx, tdb, investigation.SystemAnalysisID)
	firstTurnScope := rez.InvestigationPublicationScope{
		InvestigationID: investigation.ID,
		AgentTurnID:     firstTurn.ID,
	}
	initialParams := rez.PublishInvestigationFindingParams{
		Key:         "initial",
		Title:       "Errors increased",
		Body:        "The payment error rate rose.",
		EvidenceIDs: []uuid.UUID{citations.evidence.ID},
	}

	initial, initialErr := service.PublishInvestigationFinding(ctx, firstTurnScope, initialParams)
	s.Require().NoError(initialErr)

	firstTurnCompleted := tdb.Client(ctx).AgentTurn.UpdateOneID(firstTurn.ID).
		SetStatus(agentturn.StatusCompleted)
	firstTurnCompletedErr := firstTurnCompleted.Exec(ctx)
	s.Require().NoError(firstTurnCompletedErr)

	secondTurn := s.createTurn(ctx, tdb, investigation.AgentSessionID, 2, agentturn.StatusRunning)
	secondTurnScope := rez.InvestigationPublicationScope{
		InvestigationID: investigation.ID,
		AgentTurnID:     secondTurn.ID,
	}
	invalidatorParams := rez.PublishInvestigationFindingParams{
		Key:   "reasoning",
		Title: "Candidate explanation",
		Body:  "The deployment may have contributed.",
		FindingReferences: []rez.InvestigationFindingVersionReference{{
			VersionID: initial.ID,
			Relation:  "invalidates",
		}},
	}

	invalidator, invalidatorErr := service.PublishInvestigationFinding(ctx, secondTurnScope, invalidatorParams)
	s.Require().NoError(invalidatorErr)

	selectedParams := rez.ListInvestigationFindingsParams{
		ListParams:      ent.ListParams{Page: 1, PageSize: 1},
		InvestigationID: investigation.ID,
	}

	selected, selectedErr := service.ListInvestigationFindings(ctx, selectedParams)
	s.Require().NoError(selectedErr)
	s.Equal(2, selected.Total)
	s.Less(selected.Page*selected.PageSize, selected.Total)
	firstVersion, firstVersionErr := service.GetInvestigationFindingVersion(ctx, investigation.ID, initial.ID)
	s.Require().NoError(firstVersionErr)
	s.Equal([]uuid.UUID{invalidator.ID}, firstVersion.InvalidatedByVersionIDs())
	s.Require().Len(invalidator.Edges.OutgoingLinks, 1)
	s.Equal(initial.ID, invalidator.Edges.OutgoingLinks[0].TargetVersionID)
	s.Equal(investigationfindingversionlink.RelationInvalidates, invalidator.Edges.OutgoingLinks[0].Relation)

	secondTurnCompleted := tdb.Client(ctx).AgentTurn.UpdateOneID(secondTurn.ID).
		SetStatus(agentturn.StatusCompleted)
	secondTurnCompletedErr := secondTurnCompleted.Exec(ctx)
	s.Require().NoError(secondTurnCompletedErr)

	thirdTurn := s.createTurn(ctx, tdb, investigation.AgentSessionID, 3, agentturn.StatusRunning)
	thirdTurnScope := rez.InvestigationPublicationScope{
		InvestigationID: investigation.ID,
		AgentTurnID:     thirdTurn.ID,
	}
	currentReasoningParams := rez.PublishInvestigationFindingParams{
		Key:   "reasoning",
		Title: "Candidate explanation",
		Body:  "Evidence is insufficient to decide.",
	}

	currentReasoning, reasoningRevErr := service.PublishInvestigationFinding(ctx, thirdTurnScope, currentReasoningParams)
	s.Require().NoError(reasoningRevErr)
	s.NotEqual(invalidator.ID, currentReasoning.ID)
	firstVersionAfterRevision, firstVersionAfterRevisionErr := service.GetInvestigationFindingVersion(ctx, investigation.ID, initial.ID)
	s.Require().NoError(firstVersionAfterRevisionErr)
	s.Empty(firstVersionAfterRevision.InvalidatedByVersionIDs())
	currentPageParams := rez.ListInvestigationFindingsParams{
		ListParams:      ent.ListParams{Page: 1, PageSize: 25},
		InvestigationID: investigation.ID,
	}

	currentPage, currentPageErr := service.ListInvestigationFindings(ctx, currentPageParams)
	s.Require().NoError(currentPageErr)
	s.Equal(2, currentPage.Total)
	var currentReasoningID uuid.UUID
	for _, item := range currentPage.Data {
		if item.Edges.Finding.Key == "reasoning" {
			currentReasoningID = item.ID
		}
	}
	s.Equal(currentReasoning.ID, currentReasoningID)
	_, wrongParentErr := service.GetInvestigationFindingVersion(ctx, uuid.New(), initial.ID)
	s.Error(wrongParentErr)
}

func (s *InvestigationServiceSuite) TestConcurrentInvestigationReportWritesDeduplicate() {
	ctx, tdb := s.SetupTestDatabase()
	jobService := mocks.NewMockJobService(s.T())
	service, investigation, turn := s.outputFixture(ctx, tdb, jobService)
	citations := s.outputCitationFixture(ctx, tdb, investigation.SystemAnalysisID)
	params := rez.PublishInvestigationReportParams{
		Text:        "Same report",
		EvidenceIDs: []uuid.UUID{citations.evidence.ID},
	}
	type writeResult struct {
		publication *ent.InvestigationReport
		publishErr  error
	}
	results := make(chan writeResult, 2)
	var workers sync.WaitGroup
	for range 2 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			turnScope := rez.InvestigationPublicationScope{
				InvestigationID: investigation.ID,
				AgentTurnID:     turn.ID,
			}

			publication, publishErr := service.PublishInvestigationReport(ctx, turnScope, params)
			results <- writeResult{
				publication: publication,
				publishErr:  publishErr,
			}
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
	queryInvestigationReport := tdb.Client(ctx).InvestigationReport.Query().
		Where(investigationreport.InvestigationID(investigation.ID))
	investigationReportCount, investigationReportCountErr := queryInvestigationReport.Count(ctx)
	s.Require().NoError(investigationReportCountErr)

	s.Equal(1, investigationReportCount)
	queryInvestigationOutputReference := tdb.Client(ctx).InvestigationOutputReference.Query().
		Where(investigationoutputreference.ReportID(publicationID))
	investigationOutputReferenceCount, investigationOutputReferenceCountErr := queryInvestigationOutputReference.Count(ctx)
	s.Require().NoError(investigationOutputReferenceCountErr)

	s.Equal(1, investigationOutputReferenceCount)
}

func (s *InvestigationServiceSuite) TestInvestigationPublicationCannotRaceTurnCompletion() {
	ctx, tdb := s.SetupTestDatabase()
	jobService := mocks.NewMockJobService(s.T())
	service, investigation, turn := s.outputFixture(ctx, tdb, jobService)
	start := make(chan struct{})
	var workers sync.WaitGroup
	var publication *ent.InvestigationReport
	var publishErr error
	var completionErr error
	workers.Add(2)
	go func() {
		defer workers.Done()
		<-start
		turnScope := rez.InvestigationPublicationScope{
			InvestigationID: investigation.ID,
			AgentTurnID:     turn.ID,
		}
		publicationParams := rez.PublishInvestigationReportParams{
			Text: "Complete race.",
		}

		publication, publishErr = service.PublishInvestigationReport(ctx, turnScope, publicationParams)
	}()
	go func() {
		defer workers.Done()
		<-start
		turnCompleted := tdb.Client(ctx).AgentTurn.UpdateOneID(turn.ID).
			SetStatus(agentturn.StatusCompleted)

		completionErr = turnCompleted.Exec(ctx)
	}()
	close(start)
	workers.Wait()
	s.Require().NoError(completionErr)

	completedTurn, completedTurnErr := tdb.Client(ctx).AgentTurn.Get(ctx, turn.ID)
	s.Require().NoError(completedTurnErr)

	s.Equal(agentturn.StatusCompleted, completedTurn.Status)
	queryPublicationCount := tdb.Client(ctx).InvestigationReport.Query().
		Where(investigationreport.InvestigationID(investigation.ID))
	publicationCount, publicationCountErr := queryPublicationCount.Count(ctx)
	s.Require().NoError(publicationCountErr)
	if publishErr == nil {
		s.Require().NotNil(publication)
		s.Equal(1, publicationCount)
	} else {
		s.ErrorIs(publishErr, rez.ErrConflict)
		s.Nil(publication)
		s.Zero(publicationCount)
	}
}
