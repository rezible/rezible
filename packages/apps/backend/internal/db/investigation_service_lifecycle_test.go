package db

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/agentturn"
	"github.com/rezible/rezible/ent/investigationevidencerevision"
	"github.com/rezible/rezible/ent/investigationuserinput"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/rezible/rezible/test/mocks"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/stretchr/testify/mock"
)

func (s *InvestigationServiceSuite) userContext(tdb rez.Database, ctx context.Context) context.Context {
	createUser := tdb.Client(ctx).User.Create().
		SetEmail(uuid.NewString() + "@example.com").
		SetName("Investigation user")
	user, userErr := createUser.Save(ctx)
	s.Require().NoError(userErr)

	return execution.NewUserContext(ctx, &ent.UserAuthSession{
		UserID:    user.ID,
		TenantID:  user.TenantID,
		ExpiresAt: time.Now().Add(time.Hour),
	})
}

func (s *InvestigationServiceSuite) expectReconcileJobs(jobService *mocks.MockJobService, investigationID uuid.UUID, times int) {
	jobService.EXPECT().
		Insert(mock.Anything, jobs.ReconcileInvestigation{
			InvestigationID: investigationID,
		}, (*river.InsertOpts)(nil)).
		Return(&rivertype.JobInsertResult{
			Job: &rivertype.JobRow{
				ID: 2001,
			},
		}, nil).
		Times(times)
}

func (s *InvestigationServiceSuite) expectInvokeJobs(jobService *mocks.MockJobService, times int, insertErr error) {
	jobService.EXPECT().
		Insert(mock.Anything, mock.IsType(jobs.InvokeAgentTurn{}), (*river.InsertOpts)(nil)).
		Return(&rivertype.JobInsertResult{
			Job: &rivertype.JobRow{
				ID: 3001,
			},
		}, insertErr).
		Times(times)
}

func (s *InvestigationServiceSuite) createLifecycleInvestigation(ctx context.Context, tdb rez.Database, jobService *mocks.MockJobService) (*InvestigationService, *ent.Investigation) {
	analysis := s.createAnalysis(tdb, ctx)
	s.expectStartJob(jobService, &rivertype.JobInsertResult{
		Job: &rivertype.JobRow{
			ID: 1001,
		},
	}, nil)
	service := s.newService(tdb, jobService)
	investigationParams := rez.CreateInvestigationParams{
		AnalysisID: analysis.ID,
		Query:      "Why are payment retries increasing?",
	}

	investigation, createErr := service.CreateInvestigation(ctx, investigationParams)
	s.Require().NoError(createErr)

	return service, investigation
}

func (s *InvestigationServiceSuite) createTerminalTurn(ctx context.Context, tdb rez.Database, sessionID uuid.UUID, sequence int) *ent.AgentTurn {
	return s.createTurn(ctx, tdb, sessionID, sequence, agentturn.StatusCompleted)
}

func (s *InvestigationServiceSuite) createTurn(ctx context.Context, tdb rez.Database, sessionID uuid.UUID, sequence int, status agentturn.Status) *ent.AgentTurn {
	createAgentTurn := tdb.Client(ctx).AgentTurn.Create().
		SetID(uuid.New()).
		SetAgentSessionID(sessionID).
		SetSequence(sequence).
		SetRiverJobID(int64(1000 + sequence)).
		SetStatus(status)
	agentTurn, agentTurnErr := createAgentTurn.Save(ctx)
	s.Require().NoError(agentTurnErr)

	return agentTurn
}

func (s *InvestigationServiceSuite) TestInvestigationInputsAndEvidenceRevisionsDeduplicateByKeys() {
	ctx, tdb := s.SetupTestDatabase()
	jobService := mocks.NewMockJobService(s.T())
	service, investigation := s.createLifecycleInvestigation(ctx, tdb, jobService)
	userCtx := s.userContext(tdb, ctx)
	s.expectReconcileJobs(jobService, investigation.ID, 3)

	inputParams := rez.SubmitInvestigationUserInputParams{
		InvestigationID: investigation.ID,
		Text:            "  Why did retries spike?  ",
		SubmissionKey:   "  request-42  ",
	}
	firstInput, firstInputErr := service.SubmitInvestigationUserInput(userCtx, inputParams)
	s.Require().NoError(firstInputErr)

	repeatedInput, repeatedInputErr := service.SubmitInvestigationUserInput(userCtx, inputParams)
	s.Require().NoError(repeatedInputErr)
	s.Equal(firstInput.ID, repeatedInput.ID)
	s.Equal("Why did retries spike?", repeatedInput.Text)

	changedInputParams := inputParams
	changedInputParams.Text = "Why did latency spike?"
	_, changedInputErr := service.SubmitInvestigationUserInput(userCtx, changedInputParams)
	s.ErrorIs(changedInputErr, rez.ErrConflict)

	evidenceParams := rez.RecordInvestigationEvidenceRevisionParams{
		InvestigationID: investigation.ID,
		Explanation:     "  A new database alert episode was linked to the analysis.  ",
		CallerKey:       "  alert-episode-42  ",
	}
	firstEvidence, firstEvidenceErr := service.RecordInvestigationEvidenceRevision(ctx, evidenceParams)
	s.Require().NoError(firstEvidenceErr)

	repeatedEvidence, repeatedEvidenceErr := service.RecordInvestigationEvidenceRevision(ctx, evidenceParams)
	s.Require().NoError(repeatedEvidenceErr)
	s.Equal(firstEvidence.ID, repeatedEvidence.ID)
	s.Equal("A new database alert episode was linked to the analysis.", repeatedEvidence.Explanation)

	changedEvidenceParams := evidenceParams
	changedEvidenceParams.Explanation = "A different alert episode was linked."
	_, changedEvidenceErr := service.RecordInvestigationEvidenceRevision(ctx, changedEvidenceParams)
	s.ErrorIs(changedEvidenceErr, rez.ErrConflict)

	inputsParams := rez.ListInvestigationUserInputsParams{
		ListParams:      ent.ListParams{OrderAsc: true},
		InvestigationID: investigation.ID,
	}

	inputs, inputsErr := service.ListInvestigationUserInputs(ctx, inputsParams)
	s.Require().NoError(inputsErr)
	s.Require().Len(inputs.Data, 1)
	s.Equal(firstInput.ID, inputs.Data[0].ID)
	requestsParams := rez.ListInvestigationEvidenceRevisionsParams{
		ListParams:      ent.ListParams{OrderAsc: true},
		InvestigationID: investigation.ID,
	}

	requests, requestsErr := service.ListInvestigationEvidenceRevisions(ctx, requestsParams)
	s.Require().NoError(requestsErr)
	s.Require().Len(requests.Data, 1)
	s.Equal(firstEvidence.ID, requests.Data[0].ID)
}

func (s *InvestigationServiceSuite) TestReconcileRetainsFailedTurnAssignmentsAndContinuesWaitingWork() {
	ctx, tdb := s.SetupTestDatabase()
	jobService := mocks.NewMockJobService(s.T())
	service, investigation := s.createLifecycleInvestigation(ctx, tdb, jobService)
	userCtx := s.userContext(tdb, ctx)
	initialTurn := s.createTurn(ctx, tdb, investigation.AgentSessionID, 1, agentturn.StatusRunning)
	s.expectReconcileJobs(jobService, investigation.ID, 5)

	submitOneParams := rez.SubmitInvestigationUserInputParams{
		InvestigationID: investigation.ID,
		Text:            "Question one",
		SubmissionKey:   "question-one",
	}

	_, submitOneErr := service.SubmitInvestigationUserInput(userCtx, submitOneParams)
	s.Require().NoError(submitOneErr)

	submitTwoParams := rez.SubmitInvestigationUserInputParams{
		InvestigationID: investigation.ID,
		Text:            "Question two",
		SubmissionKey:   "question-two",
	}

	_, submitTwoErr := service.SubmitInvestigationUserInput(userCtx, submitTwoParams)
	s.Require().NoError(submitTwoErr)

	evidenceOneParams := rez.RecordInvestigationEvidenceRevisionParams{
		InvestigationID: investigation.ID,
		Explanation:     "First evidence change",
		CallerKey:       "evidence-one",
	}

	_, evidenceOneErr := service.RecordInvestigationEvidenceRevision(ctx, evidenceOneParams)
	s.Require().NoError(evidenceOneErr)

	evidenceTwoParams := rez.RecordInvestigationEvidenceRevisionParams{
		InvestigationID: investigation.ID,
		Explanation:     "Second evidence change",
		CallerKey:       "evidence-two",
	}

	_, evidenceTwoErr := service.RecordInvestigationEvidenceRevision(ctx, evidenceTwoParams)
	s.Require().NoError(evidenceTwoErr)

	busyReconcileErr := service.ReconcileInvestigation(ctx, investigation.ID)
	s.Require().NoError(busyReconcileErr)

	queryAgentTurn := tdb.Client(ctx).AgentTurn.Query().
		Where(agentturn.AgentSessionID(investigation.AgentSessionID))
	agentTurnCount, agentTurnCountErr := queryAgentTurn.Count(ctx)
	s.Require().NoError(agentTurnCountErr)

	s.Equal(1, agentTurnCount)
	completeInitialTurn := tdb.Client(ctx).AgentTurn.UpdateOneID(initialTurn.ID).
		SetStatus(agentturn.StatusCompleted)
	completeInitialTurnErr := completeInitialTurn.Exec(ctx)
	s.Require().NoError(completeInitialTurnErr)

	orderedInputs, orderedInputsErr := service.ListInvestigationUserInputs(ctx, rez.ListInvestigationUserInputsParams{InvestigationID: investigation.ID})
	s.Require().NoError(orderedInputsErr)

	orderedEvidence, orderedEvidenceErr := service.ListInvestigationEvidenceRevisions(ctx, rez.ListInvestigationEvidenceRevisionsParams{InvestigationID: investigation.ID})
	s.Require().NoError(orderedEvidenceErr)
	s.Require().Len(orderedInputs.Data, 2)
	s.Require().Len(orderedEvidence.Data, 2)
	s.expectInvokeJobs(jobService, 2, nil)

	firstReconcileErr := service.ReconcileInvestigation(ctx, investigation.ID)
	s.Require().NoError(firstReconcileErr)

	queryFirstFollowUpTurn := tdb.Client(ctx).AgentTurn.Query().
		Where(agentturn.AgentSessionID(investigation.AgentSessionID), agentturn.Sequence(2))
	firstFollowUpTurn, firstFollowUpTurnErr := queryFirstFollowUpTurn.Only(ctx)
	s.Require().NoError(firstFollowUpTurnErr)

	firstInputMessage, firstInputMessageErr := tdb.Client(ctx).AgentMessage.Get(ctx, *firstFollowUpTurn.InputMessageID)
	s.Require().NoError(firstInputMessageErr)
	expectedFirstMessage := "User question:\n" + orderedInputs.Data[0].Text +
		"\n\nEvidence revised:\n" + orderedEvidence.Data[0].Explanation +
		"\n" + orderedEvidence.Data[1].Explanation
	s.Equal(expectedFirstMessage, firstInputMessage.Content[0].Text)

	assignedInput, assignedInputErr := tdb.Client(ctx).InvestigationUserInput.Get(ctx, orderedInputs.Data[0].ID)
	s.Require().NoError(assignedInputErr)
	s.Equal(firstFollowUpTurn.ID, *assignedInput.AgentTurnID)
	for _, revision := range orderedEvidence.Data {
		assignedEvidence, assignedEvidenceErr := tdb.Client(ctx).InvestigationEvidenceRevision.Get(ctx, revision.ID)
		s.Require().NoError(assignedEvidenceErr)
		s.Equal(firstFollowUpTurn.ID, *assignedEvidence.AgentTurnID)
	}
	waitingInput, waitingInputErr := tdb.Client(ctx).InvestigationUserInput.Get(ctx, orderedInputs.Data[1].ID)
	s.Require().NoError(waitingInputErr)
	s.Nil(waitingInput.AgentTurnID)

	queuedMessage := firstInputMessage.Content[0].Text
	thirdInputParams := rez.SubmitInvestigationUserInputParams{
		InvestigationID: investigation.ID,
		Text:            "Question while a turn is queued",
		SubmissionKey:   "question-three",
	}

	_, thirdInputErr := service.SubmitInvestigationUserInput(userCtx, thirdInputParams)
	s.Require().NoError(thirdInputErr)

	queuedInputMessage, queuedInputMessageErr := tdb.Client(ctx).AgentMessage.Get(ctx, *firstFollowUpTurn.InputMessageID)
	s.Require().NoError(queuedInputMessageErr)
	s.Equal(queuedMessage, queuedInputMessage.Content[0].Text)
	queuedReconcileErr := service.ReconcileInvestigation(ctx, investigation.ID)
	s.Require().NoError(queuedReconcileErr)

	queryAgentTurn2 := tdb.Client(ctx).AgentTurn.Query().
		Where(agentturn.AgentSessionID(investigation.AgentSessionID))
	agentTurn2Count, agentTurn2CountErr := queryAgentTurn2.Count(ctx)
	s.Require().NoError(agentTurn2CountErr)

	s.Equal(2, agentTurn2Count)
	s.expectReconcileJobs(jobService, investigation.ID, 1)
	replayedAssignedInputParams := rez.SubmitInvestigationUserInputParams{
		InvestigationID: investigation.ID,
		Text:            "Question one",
		SubmissionKey:   "question-one",
	}

	replayedAssignedInput, replayAssignedErr := service.SubmitInvestigationUserInput(userCtx, replayedAssignedInputParams)
	s.Require().NoError(replayAssignedErr)
	s.Equal(firstFollowUpTurn.ID, *replayedAssignedInput.AgentTurnID)

	failFollowUpTurn := tdb.Client(ctx).AgentTurn.UpdateOneID(firstFollowUpTurn.ID).
		SetStatus(agentturn.StatusFailed)
	failFollowUpTurnErr := failFollowUpTurn.Exec(ctx)
	s.Require().NoError(failFollowUpTurnErr)

	secondReconcileErr := service.ReconcileInvestigation(ctx, investigation.ID)
	s.Require().NoError(secondReconcileErr)

	querySecondFollowUpTurn := tdb.Client(ctx).AgentTurn.Query().
		Where(agentturn.AgentSessionID(investigation.AgentSessionID), agentturn.Sequence(3))
	secondFollowUpTurn, secondFollowUpTurnErr := querySecondFollowUpTurn.Only(ctx)
	s.Require().NoError(secondFollowUpTurnErr)

	secondInputMessage, secondInputMessageErr := tdb.Client(ctx).AgentMessage.Get(ctx, *secondFollowUpTurn.InputMessageID)
	s.Require().NoError(secondInputMessageErr)
	s.Equal("User question:\n"+orderedInputs.Data[1].Text, secondInputMessage.Content[0].Text)

	var reloadAssignedInputErr error
	assignedInput, reloadAssignedInputErr = tdb.Client(ctx).InvestigationUserInput.Get(ctx, orderedInputs.Data[0].ID)
	s.Require().NoError(reloadAssignedInputErr)
	s.Equal(firstFollowUpTurn.ID, *assignedInput.AgentTurnID)
	for _, revision := range orderedEvidence.Data {
		retainedEvidence, retainedEvidenceErr := tdb.Client(ctx).InvestigationEvidenceRevision.Get(ctx, revision.ID)
		s.Require().NoError(retainedEvidenceErr)
		s.Equal(firstFollowUpTurn.ID, *retainedEvidence.AgentTurnID)
	}
}

func (s *InvestigationServiceSuite) TestReconcileLeavesWorkForStartupAndRollsBackWhenTurnJobFails() {
	ctx, tdb := s.SetupTestDatabase()
	jobService := mocks.NewMockJobService(s.T())
	service, investigation := s.createLifecycleInvestigation(ctx, tdb, jobService)
	userCtx := s.userContext(tdb, ctx)
	s.expectReconcileJobs(jobService, investigation.ID, 1)
	inputParams := rez.SubmitInvestigationUserInputParams{
		InvestigationID: investigation.ID,
		Text:            "Waiting for initial execution",
		SubmissionKey:   "waiting-input",
	}

	input, submitErr := service.SubmitInvestigationUserInput(userCtx, inputParams)
	s.Require().NoError(submitErr)

	noTurnReconcileErr := service.ReconcileInvestigation(ctx, investigation.ID)
	s.Require().NoError(noTurnReconcileErr)

	queryAgentTurn := tdb.Client(ctx).AgentTurn.Query().
		Where(agentturn.AgentSessionID(investigation.AgentSessionID))
	agentTurnCount, agentTurnCountErr := queryAgentTurn.Count(ctx)
	s.Require().NoError(agentTurnCountErr)

	s.Zero(agentTurnCount)
	waitingInput, waitingInputErr := tdb.Client(ctx).InvestigationUserInput.Get(ctx, input.ID)
	s.Require().NoError(waitingInputErr)

	s.Nil(waitingInput.AgentTurnID)

	s.createTerminalTurn(ctx, tdb, investigation.AgentSessionID, 1)
	queueFailure := errors.New("agent turn queue unavailable")
	s.expectInvokeJobs(jobService, 1, queueFailure)
	failedReconcileErr := service.ReconcileInvestigation(ctx, investigation.ID)
	s.ErrorContains(failedReconcileErr, "agent turn queue unavailable")
	queryAgentTurn2 := tdb.Client(ctx).AgentTurn.Query().
		Where(agentturn.AgentSessionID(investigation.AgentSessionID))
	agentTurn2Count, agentTurn2CountErr := queryAgentTurn2.Count(ctx)
	s.Require().NoError(agentTurn2CountErr)

	s.Equal(1, agentTurn2Count)
	afterFailureInput, afterFailureInputErr := tdb.Client(ctx).InvestigationUserInput.Get(ctx, input.ID)
	s.Require().NoError(afterFailureInputErr)

	s.Nil(afterFailureInput.AgentTurnID)
}

func (s *InvestigationServiceSuite) TestEvidenceAnalysisChangeAndRevisionShareCallerTransaction() {
	ctx, tdb := s.SetupTestDatabase()
	jobService := mocks.NewMockJobService(s.T())
	service, investigation := s.createLifecycleInvestigation(ctx, tdb, jobService)
	s.expectReconcileJobs(jobService, investigation.ID, 1)
	queryService, queryServiceErr := NewKnowledgeGraphQueryService(tdb)
	s.Require().NoError(queryServiceErr)

	analysisService, analysisServiceErr := NewSystemAnalysisService(tdb, queryService)
	s.Require().NoError(analysisServiceErr)

	rollbackErr := errors.New("rollback analysis and evidence revision")
	referenceTime := time.Now().UTC()
	transactionErr := tdb.WithTx(ctx, func(txCtx context.Context, _ *ent.Client) error {
		_, updateErr := analysisService.SetSystemAnalysis(txCtx, investigation.SystemAnalysisID, func(mutation *ent.SystemAnalysisMutation) {
			mutation.SetReferenceTime(referenceTime)
		})
		if updateErr != nil {
			return updateErr
		}
		requestParams := rez.RecordInvestigationEvidenceRevisionParams{
			InvestigationID: investigation.ID,
			Explanation:     "New evidence was linked to the analysis.",
			CallerKey:       "analysis-update-rollback",
		}

		_, requestErr := service.RecordInvestigationEvidenceRevision(txCtx, requestParams)
		if requestErr != nil {
			return requestErr
		}
		return rollbackErr
	})
	s.ErrorIs(transactionErr, rollbackErr)

	analysis, analysisErr := tdb.Client(ctx).SystemAnalysis.Get(ctx, investigation.SystemAnalysisID)
	s.Require().NoError(analysisErr)
	s.Nil(analysis.ReferenceTime)
	queryInvestigationEvidenceRevision := tdb.Client(ctx).InvestigationEvidenceRevision.Query().
		Where(
			investigationevidencerevision.InvestigationID(investigation.ID),
		)
	investigationEvidenceRevisionCount, investigationEvidenceRevisionCountErr := queryInvestigationEvidenceRevision.Count(ctx)
	s.Require().NoError(investigationEvidenceRevisionCountErr)

	s.Zero(investigationEvidenceRevisionCount)
}

func (s *InvestigationServiceSuite) TestTerminalEventAndInputRaceCanDispatchDuplicateJobsOnlyOnce() {
	ctx, tdb := s.SetupTestDatabase()
	jobService := mocks.NewMockJobService(s.T())
	service, investigation := s.createLifecycleInvestigation(ctx, tdb, jobService)
	userCtx := s.userContext(tdb, ctx)
	initialTurn := s.createTerminalTurn(ctx, tdb, investigation.AgentSessionID, 1)
	s.expectReconcileJobs(jobService, investigation.ID, 2)
	s.expectInvokeJobs(jobService, 1, nil)
	unrelatedEventErr := service.onAgentTurnUpdated(ctx, &rezai.AgentTurnUpdated{
		AgentSessionId: uuid.New(),
		AgentTurnId:    uuid.New(),
		Status:         agentturn.StatusCompleted,
	})
	s.Require().NoError(unrelatedEventErr)

	var wait sync.WaitGroup
	operationErrors := make(chan error, 2)
	wait.Add(2)
	go func() {
		defer wait.Done()
		submitParams := rez.SubmitInvestigationUserInputParams{
			InvestigationID: investigation.ID,
			Text:            "Question submitted as the previous turn completes",
			SubmissionKey:   "terminal-race",
		}

		_, submitErr := service.SubmitInvestigationUserInput(userCtx, submitParams)
		operationErrors <- submitErr
	}()
	go func() {
		defer wait.Done()
		operationErrors <- service.onAgentTurnUpdated(ctx, &rezai.AgentTurnUpdated{
			AgentSessionId: investigation.AgentSessionID,
			AgentTurnId:    initialTurn.ID,
			Status:         agentturn.StatusCompleted,
		})
	}()
	wait.Wait()
	close(operationErrors)
	for operationErr := range operationErrors {
		s.Require().NoError(operationErr)
	}

	var reconcileWait sync.WaitGroup
	reconcileErrors := make(chan error, 2)
	reconcileWait.Add(2)
	for range 2 {
		go func() {
			defer reconcileWait.Done()
			reconcileErrors <- service.ReconcileInvestigation(ctx, investigation.ID)
		}()
	}
	reconcileWait.Wait()
	close(reconcileErrors)
	for reconcileErr := range reconcileErrors {
		s.Require().NoError(reconcileErr)
	}

	queryAgentTurn := tdb.Client(ctx).AgentTurn.Query().
		Where(agentturn.AgentSessionID(investigation.AgentSessionID))
	agentTurnCount, agentTurnCountErr := queryAgentTurn.Count(ctx)
	s.Require().NoError(agentTurnCountErr)

	s.Equal(2, agentTurnCount)
	queryInputs := tdb.Client(ctx).InvestigationUserInput.Query().
		Where(investigationuserinput.InvestigationID(investigation.ID), investigationuserinput.AgentTurnIDNotNil())
	inputs, inputsErr := queryInputs.All(ctx)
	s.Require().NoError(inputsErr)
	s.Require().Len(inputs, 1)
	assignedTurn, assignedTurnErr := tdb.Client(ctx).AgentTurn.Get(ctx, *inputs[0].AgentTurnID)
	s.Require().NoError(assignedTurnErr)

	s.Equal(2, assignedTurn.Sequence)
}

func (s *InvestigationServiceSuite) recordEvidenceRevision(ctx context.Context, tdb rez.Database, service *InvestigationService, investigationID uuid.UUID, key string) *ent.InvestigationEvidenceRevision {
	revisionParams := rez.RecordInvestigationEvidenceRevisionParams{
		InvestigationID: investigationID,
		Explanation:     "Evidence changed: " + key + ".",
		CallerKey:       key,
	}
	revision, recordErr := service.RecordInvestigationEvidenceRevision(ctx, revisionParams)
	s.Require().NoError(recordErr)

	stored, storedErr := tdb.Client(ctx).InvestigationEvidenceRevision.Get(ctx, revision.ID)
	s.Require().NoError(storedErr)

	return stored
}

func (s *InvestigationServiceSuite) latestTurn(ctx context.Context, tdb rez.Database, sessionID uuid.UUID) *ent.AgentTurn {
	queryLatestTurn := tdb.Client(ctx).AgentTurn.Query().
		Where(agentturn.AgentSessionID(sessionID)).
		Order(agentturn.BySequence(sql.OrderDesc()))
	turn, turnErr := queryLatestTurn.First(ctx)
	s.Require().NoError(turnErr)

	return turn
}

func (s *InvestigationServiceSuite) finishTurn(ctx context.Context, tdb rez.Database, turn *ent.AgentTurn, status agentturn.Status) {
	finishTurn := tdb.Client(ctx).AgentTurn.UpdateOneID(turn.ID).
		SetStatus(status)
	s.Require().NoError(finishTurn.Exec(ctx))
}

func (s *InvestigationServiceSuite) countTurns(ctx context.Context, tdb rez.Database, sessionID uuid.UUID) int {
	queryTurns := tdb.Client(ctx).AgentTurn.Query().
		Where(agentturn.AgentSessionID(sessionID))
	turnCount, countErr := queryTurns.Count(ctx)
	s.Require().NoError(countErr)

	return turnCount
}

func (s *InvestigationServiceSuite) TestReconcileAssignsAllPendingEvidenceToOneTurn() {
	ctx, tdb := s.SetupTestDatabase()
	jobService := mocks.NewMockJobService(s.T())
	service, investigation := s.createLifecycleInvestigation(ctx, tdb, jobService)
	userCtx := s.userContext(tdb, ctx)
	runningTurn := s.createTurn(ctx, tdb, investigation.AgentSessionID, 1, agentturn.StatusRunning)
	s.expectReconcileJobs(jobService, investigation.ID, 5)

	firstQuestionParams := rez.SubmitInvestigationUserInputParams{
		InvestigationID: investigation.ID,
		Text:            "Did the deploy cause this?",
		SubmissionKey:   "question-one",
	}
	firstQuestion, firstQuestionErr := service.SubmitInvestigationUserInput(userCtx, firstQuestionParams)
	s.Require().NoError(firstQuestionErr)

	secondQuestionParams := rez.SubmitInvestigationUserInputParams{
		InvestigationID: investigation.ID,
		Text:            "Which customers are affected?",
		SubmissionKey:   "question-two",
	}
	secondQuestion, secondQuestionErr := service.SubmitInvestigationUserInput(userCtx, secondQuestionParams)
	s.Require().NoError(secondQuestionErr)

	revisions := []*ent.InvestigationEvidenceRevision{
		s.recordEvidenceRevision(ctx, tdb, service, investigation.ID, "evidence-one"),
		s.recordEvidenceRevision(ctx, tdb, service, investigation.ID, "evidence-two"),
		s.recordEvidenceRevision(ctx, tdb, service, investigation.ID, "evidence-three"),
	}

	s.Require().NoError(service.ReconcileInvestigation(ctx, investigation.ID))
	s.Equal(1, s.countTurns(ctx, tdb, investigation.AgentSessionID))

	s.finishTurn(ctx, tdb, runningTurn, agentturn.StatusCompleted)
	s.expectInvokeJobs(jobService, 1, nil)

	s.Require().NoError(service.ReconcileInvestigation(ctx, investigation.ID))

	nextTurn := s.latestTurn(ctx, tdb, investigation.AgentSessionID)
	s.Equal(2, nextTurn.Sequence)
	for _, revision := range revisions {
		assigned, assignedErr := tdb.Client(ctx).InvestigationEvidenceRevision.Get(ctx, revision.ID)
		s.Require().NoError(assignedErr)
		s.Equal(nextTurn.ID, *assigned.AgentTurnID)
	}

	assignedQuestion, assignedQuestionErr := tdb.Client(ctx).InvestigationUserInput.Get(ctx, firstQuestion.ID)
	s.Require().NoError(assignedQuestionErr)
	s.Equal(nextTurn.ID, *assignedQuestion.AgentTurnID)

	waitingQuestion, waitingQuestionErr := tdb.Client(ctx).InvestigationUserInput.Get(ctx, secondQuestion.ID)
	s.Require().NoError(waitingQuestionErr)
	s.Nil(waitingQuestion.AgentTurnID)
}

func (s *InvestigationServiceSuite) TestEvidenceTurnsStopAtLimit() {
	ctx, tdb := s.SetupTestDatabase()
	jobService := mocks.NewMockJobService(s.T())
	service, investigation := s.createLifecycleInvestigation(ctx, tdb, jobService)
	userCtx := s.userContext(tdb, ctx)
	sessionID := investigation.AgentSessionID
	s.createTerminalTurn(ctx, tdb, sessionID, 1)
	s.expectReconcileJobs(jobService, investigation.ID, MaxEvidenceTurns+3)
	s.expectInvokeJobs(jobService, MaxEvidenceTurns+2, nil)

	for i := range MaxEvidenceTurns {
		s.recordEvidenceRevision(ctx, tdb, service, investigation.ID, fmt.Sprintf("evidence-%d", i))
		s.Require().NoError(service.ReconcileInvestigation(ctx, investigation.ID))

		turn := s.latestTurn(ctx, tdb, sessionID)
		s.Equal(i+2, turn.Sequence)
		s.finishTurn(ctx, tdb, turn, agentturn.StatusCompleted)
	}

	s.recordEvidenceRevision(ctx, tdb, service, investigation.ID, "evidence-past-limit")
	s.Require().NoError(service.ReconcileInvestigation(ctx, investigation.ID))
	s.Equal(MaxEvidenceTurns+1, s.countTurns(ctx, tdb, sessionID))

	paused, pausedErr := service.ReadInvestigationDetail(ctx, investigation.ID)
	s.Require().NoError(pausedErr)
	s.Equal(1, paused.PendingEvidenceRevisions)
	s.True(paused.AutomaticUpdatesPaused)

	s.Require().NoError(service.UpdateInvestigation(ctx, investigation.ID))
	updateTurn := s.latestTurn(ctx, tdb, sessionID)
	s.Equal(MaxEvidenceTurns+2, updateTurn.Sequence)
	s.finishTurn(ctx, tdb, updateTurn, agentturn.StatusCompleted)

	updated, updatedErr := service.ReadInvestigationDetail(ctx, investigation.ID)
	s.Require().NoError(updatedErr)
	s.Zero(updated.PendingEvidenceRevisions)
	s.False(updated.AutomaticUpdatesPaused)

	s.recordEvidenceRevision(ctx, tdb, service, investigation.ID, "evidence-after-update")
	s.Require().NoError(service.ReconcileInvestigation(ctx, investigation.ID))
	s.Equal(MaxEvidenceTurns+2, s.countTurns(ctx, tdb, sessionID))

	questionParams := rez.SubmitInvestigationUserInputParams{
		InvestigationID: investigation.ID,
		Text:            "Is the error rate still rising?",
		SubmissionKey:   "question-past-limit",
	}
	_, questionErr := service.SubmitInvestigationUserInput(userCtx, questionParams)
	s.Require().NoError(questionErr)
	s.Require().NoError(service.ReconcileInvestigation(ctx, investigation.ID))

	questionTurn := s.latestTurn(ctx, tdb, sessionID)
	s.Equal(MaxEvidenceTurns+3, questionTurn.Sequence)

	answered, answeredErr := service.ReadInvestigationDetail(ctx, investigation.ID)
	s.Require().NoError(answeredErr)
	s.Zero(answered.PendingEvidenceRevisions)
}

func (s *InvestigationServiceSuite) TestInvestigationEvidenceCurrency() {
	ctx, tdb := s.SetupTestDatabase()
	jobService := mocks.NewMockJobService(s.T())
	service, investigation := s.createLifecycleInvestigation(ctx, tdb, jobService)
	sessionID := investigation.AgentSessionID
	s.createTerminalTurn(ctx, tdb, sessionID, 1)
	s.expectReconcileJobs(jobService, investigation.ID, 3)
	s.expectInvokeJobs(jobService, 3, nil)

	readDetail := func() *rez.InvestigationDetail {
		detail, detailErr := service.ReadInvestigationDetail(ctx, investigation.ID)
		s.Require().NoError(detailErr)
		return detail
	}

	initial := readDetail()
	s.Zero(initial.PendingEvidenceRevisions)
	s.WithinDuration(investigation.CreatedAt, initial.EvidenceCurrentAsOf, 0)

	completedRevision := s.recordEvidenceRevision(ctx, tdb, service, investigation.ID, "evidence-completed")
	recorded := readDetail()
	s.Equal(1, recorded.PendingEvidenceRevisions)
	s.False(recorded.AutomaticUpdatesPaused)
	s.WithinDuration(investigation.CreatedAt, recorded.EvidenceCurrentAsOf, 0)

	s.Require().NoError(service.ReconcileInvestigation(ctx, investigation.ID))
	running := readDetail()
	s.Zero(running.PendingEvidenceRevisions)
	s.WithinDuration(investigation.CreatedAt, running.EvidenceCurrentAsOf, 0)

	s.finishTurn(ctx, tdb, s.latestTurn(ctx, tdb, sessionID), agentturn.StatusCompleted)
	completed := readDetail()
	s.WithinDuration(completedRevision.CreatedAt, completed.EvidenceCurrentAsOf, 0)

	s.recordEvidenceRevision(ctx, tdb, service, investigation.ID, "evidence-failed")
	s.Require().NoError(service.ReconcileInvestigation(ctx, investigation.ID))
	s.finishTurn(ctx, tdb, s.latestTurn(ctx, tdb, sessionID), agentturn.StatusFailed)
	failed := readDetail()
	s.Zero(failed.PendingEvidenceRevisions)
	s.WithinDuration(completedRevision.CreatedAt, failed.EvidenceCurrentAsOf, 0)

	newerRevision := s.recordEvidenceRevision(ctx, tdb, service, investigation.ID, "evidence-after-failure")
	afterFailure := readDetail()
	s.Equal(1, afterFailure.PendingEvidenceRevisions)

	s.Require().NoError(service.ReconcileInvestigation(ctx, investigation.ID))
	s.finishTurn(ctx, tdb, s.latestTurn(ctx, tdb, sessionID), agentturn.StatusCompleted)
	caughtUp := readDetail()
	s.Zero(caughtUp.PendingEvidenceRevisions)
	s.WithinDuration(newerRevision.CreatedAt, caughtUp.EvidenceCurrentAsOf, 0)
}
