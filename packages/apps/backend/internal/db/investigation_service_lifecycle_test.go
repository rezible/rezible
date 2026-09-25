package db

import (
	"context"
	"errors"
	"sync"
	"time"

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
	user := tdb.Client(ctx).User.Create().
		SetEmail(uuid.NewString() + "@example.com").
		SetName("Investigation user").
		SaveX(ctx)
	return execution.NewUserContext(ctx, &ent.UserAuthSession{
		UserID:    user.ID,
		TenantID:  user.TenantID,
		ExpiresAt: time.Now().Add(time.Hour),
	})
}

func (s *InvestigationServiceSuite) expectReconcileJobs(jobService *mocks.MockJobService, investigationID uuid.UUID, times int) {
	jobService.EXPECT().
		Insert(mock.Anything, jobs.ReconcileInvestigation{InvestigationID: investigationID}, (*river.InsertOpts)(nil)).
		Return(&rivertype.JobInsertResult{Job: &rivertype.JobRow{ID: 2001}}, nil).
		Times(times)
}

func (s *InvestigationServiceSuite) expectInvokeJobs(jobService *mocks.MockJobService, times int, insertErr error) {
	jobService.EXPECT().
		Insert(mock.Anything, mock.IsType(jobs.InvokeAgentTurn{}), (*river.InsertOpts)(nil)).
		Return(&rivertype.JobInsertResult{Job: &rivertype.JobRow{ID: 3001}}, insertErr).
		Times(times)
}

func (s *InvestigationServiceSuite) createLifecycleInvestigation(ctx context.Context, tdb rez.Database, jobService *mocks.MockJobService) (*InvestigationService, *ent.Investigation) {
	analysis := s.createAnalysis(tdb, ctx)
	s.expectStartJob(jobService, &rivertype.JobInsertResult{Job: &rivertype.JobRow{ID: 1001}}, nil)
	service := s.newService(tdb, jobService)
	investigation, createErr := service.CreateInvestigation(ctx, rez.CreateInvestigationParams{
		AnalysisID: analysis.ID,
		Query:      "Why are payment retries increasing?",
	})
	s.Require().NoError(createErr)
	return service, investigation
}

func (s *InvestigationServiceSuite) createTerminalTurn(ctx context.Context, tdb rez.Database, sessionID uuid.UUID, sequence int) *ent.AgentTurn {
	return s.createTurn(ctx, tdb, sessionID, sequence, agentturn.StatusCompleted)
}

func (s *InvestigationServiceSuite) createTurn(ctx context.Context, tdb rez.Database, sessionID uuid.UUID, sequence int, status agentturn.Status) *ent.AgentTurn {
	return tdb.Client(ctx).AgentTurn.Create().
		SetID(uuid.New()).
		SetAgentSessionID(sessionID).
		SetSequence(sequence).
		SetRiverJobID(int64(1000 + sequence)).
		SetStatus(status).
		SaveX(ctx)
}

func (s *InvestigationServiceSuite) TestInvestigationInputsAndEvidenceRevisionsDeduplicateByKeys() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
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

	inputs, inputsErr := service.ListInvestigationUserInputs(ctx, investigation.ID, ent.ListParams{OrderAsc: true})
	s.Require().NoError(inputsErr)
	s.Require().Len(inputs.Data, 1)
	s.Equal(firstInput.ID, inputs.Data[0].ID)
	requests, requestsErr := service.ListInvestigationEvidenceRevisions(ctx, investigation.ID, ent.ListParams{OrderAsc: true})
	s.Require().NoError(requestsErr)
	s.Require().Len(requests.Data, 1)
	s.Equal(firstEvidence.ID, requests.Data[0].ID)
}

func (s *InvestigationServiceSuite) TestReconcileRetainsFailedTurnAssignmentsAndContinuesWaitingWork() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	jobService := mocks.NewMockJobService(s.T())
	service, investigation := s.createLifecycleInvestigation(ctx, tdb, jobService)
	userCtx := s.userContext(tdb, ctx)
	initialTurn := s.createTurn(ctx, tdb, investigation.AgentSessionID, 1, agentturn.StatusRunning)
	s.expectReconcileJobs(jobService, investigation.ID, 5)

	_, submitOneErr := service.SubmitInvestigationUserInput(userCtx, rez.SubmitInvestigationUserInputParams{
		InvestigationID: investigation.ID,
		Text:            "Question one",
		SubmissionKey:   "question-one",
	})
	s.Require().NoError(submitOneErr)
	_, submitTwoErr := service.SubmitInvestigationUserInput(userCtx, rez.SubmitInvestigationUserInputParams{
		InvestigationID: investigation.ID,
		Text:            "Question two",
		SubmissionKey:   "question-two",
	})
	s.Require().NoError(submitTwoErr)
	_, evidenceOneErr := service.RecordInvestigationEvidenceRevision(ctx, rez.RecordInvestigationEvidenceRevisionParams{
		InvestigationID: investigation.ID,
		Explanation:     "First evidence change",
		CallerKey:       "evidence-one",
	})
	s.Require().NoError(evidenceOneErr)
	_, evidenceTwoErr := service.RecordInvestigationEvidenceRevision(ctx, rez.RecordInvestigationEvidenceRevisionParams{
		InvestigationID: investigation.ID,
		Explanation:     "Second evidence change",
		CallerKey:       "evidence-two",
	})
	s.Require().NoError(evidenceTwoErr)
	busyReconcileErr := service.ReconcileInvestigation(ctx, investigation.ID)
	s.Require().NoError(busyReconcileErr)
	s.Equal(1, tdb.Client(ctx).AgentTurn.Query().Where(agentturn.AgentSessionID(investigation.AgentSessionID)).CountX(ctx))
	tdb.Client(ctx).AgentTurn.UpdateOneID(initialTurn.ID).SetStatus(agentturn.StatusCompleted).ExecX(ctx)

	orderedInputs, orderedInputsErr := service.ListInvestigationUserInputs(ctx, investigation.ID, ent.ListParams{})
	s.Require().NoError(orderedInputsErr)
	orderedEvidence, orderedEvidenceErr := service.ListInvestigationEvidenceRevisions(ctx, investigation.ID, ent.ListParams{})
	s.Require().NoError(orderedEvidenceErr)
	s.Require().Len(orderedInputs.Data, 2)
	s.Require().Len(orderedEvidence.Data, 2)
	s.expectInvokeJobs(jobService, 2, nil)

	firstReconcileErr := service.ReconcileInvestigation(ctx, investigation.ID)
	s.Require().NoError(firstReconcileErr)
	firstFollowUpTurn := tdb.Client(ctx).AgentTurn.Query().
		Where(agentturn.AgentSessionID(investigation.AgentSessionID), agentturn.Sequence(2)).
		OnlyX(ctx)
	firstInputMessage := tdb.Client(ctx).AgentMessage.GetX(ctx, *firstFollowUpTurn.InputMessageID)
	s.Equal("User question:\n"+orderedInputs.Data[0].Text+"\n\nEvidence revised:\n"+orderedEvidence.Data[0].Explanation, firstInputMessage.Content[0].Text)

	assignedInput := tdb.Client(ctx).InvestigationUserInput.GetX(ctx, orderedInputs.Data[0].ID)
	s.Equal(firstFollowUpTurn.ID, *assignedInput.AgentTurnID)
	assignedEvidence := tdb.Client(ctx).InvestigationEvidenceRevision.GetX(ctx, orderedEvidence.Data[0].ID)
	s.Equal(firstFollowUpTurn.ID, *assignedEvidence.AgentTurnID)
	waitingEvidence := tdb.Client(ctx).InvestigationEvidenceRevision.GetX(ctx, orderedEvidence.Data[1].ID)
	s.Nil(waitingEvidence.AgentTurnID)
	waitingInput := tdb.Client(ctx).InvestigationUserInput.GetX(ctx, orderedInputs.Data[1].ID)
	s.Nil(waitingInput.AgentTurnID)

	queuedMessage := firstInputMessage.Content[0].Text
	_, thirdInputErr := service.SubmitInvestigationUserInput(userCtx, rez.SubmitInvestigationUserInputParams{
		InvestigationID: investigation.ID,
		Text:            "Question while a turn is queued",
		SubmissionKey:   "question-three",
	})
	s.Require().NoError(thirdInputErr)
	queuedInputMessage := tdb.Client(ctx).AgentMessage.GetX(ctx, *firstFollowUpTurn.InputMessageID)
	s.Equal(queuedMessage, queuedInputMessage.Content[0].Text)
	queuedReconcileErr := service.ReconcileInvestigation(ctx, investigation.ID)
	s.Require().NoError(queuedReconcileErr)
	s.Equal(2, tdb.Client(ctx).AgentTurn.Query().Where(agentturn.AgentSessionID(investigation.AgentSessionID)).CountX(ctx))
	s.expectReconcileJobs(jobService, investigation.ID, 1)
	replayedAssignedInput, replayAssignedErr := service.SubmitInvestigationUserInput(userCtx, rez.SubmitInvestigationUserInputParams{
		InvestigationID: investigation.ID,
		Text:            "Question one",
		SubmissionKey:   "question-one",
	})
	s.Require().NoError(replayAssignedErr)
	s.Equal(firstFollowUpTurn.ID, *replayedAssignedInput.AgentTurnID)

	tdb.Client(ctx).AgentTurn.UpdateOneID(firstFollowUpTurn.ID).SetStatus(agentturn.StatusFailed).ExecX(ctx)
	secondReconcileErr := service.ReconcileInvestigation(ctx, investigation.ID)
	s.Require().NoError(secondReconcileErr)
	secondFollowUpTurn := tdb.Client(ctx).AgentTurn.Query().
		Where(agentturn.AgentSessionID(investigation.AgentSessionID), agentturn.Sequence(3)).
		OnlyX(ctx)
	secondInputMessage := tdb.Client(ctx).AgentMessage.GetX(ctx, *secondFollowUpTurn.InputMessageID)
	s.Equal("User question:\n"+orderedInputs.Data[1].Text+"\n\nEvidence revised:\n"+orderedEvidence.Data[1].Explanation, secondInputMessage.Content[0].Text)

	assignedInput = tdb.Client(ctx).InvestigationUserInput.GetX(ctx, orderedInputs.Data[0].ID)
	assignedEvidence = tdb.Client(ctx).InvestigationEvidenceRevision.GetX(ctx, orderedEvidence.Data[0].ID)
	s.Equal(firstFollowUpTurn.ID, *assignedInput.AgentTurnID)
	s.Equal(firstFollowUpTurn.ID, *assignedEvidence.AgentTurnID)
}

func (s *InvestigationServiceSuite) TestReconcileLeavesWorkForStartupAndRollsBackWhenTurnJobFails() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	jobService := mocks.NewMockJobService(s.T())
	service, investigation := s.createLifecycleInvestigation(ctx, tdb, jobService)
	userCtx := s.userContext(tdb, ctx)
	s.expectReconcileJobs(jobService, investigation.ID, 1)
	input, submitErr := service.SubmitInvestigationUserInput(userCtx, rez.SubmitInvestigationUserInputParams{
		InvestigationID: investigation.ID,
		Text:            "Waiting for initial execution",
		SubmissionKey:   "waiting-input",
	})
	s.Require().NoError(submitErr)

	noTurnReconcileErr := service.ReconcileInvestigation(ctx, investigation.ID)
	s.Require().NoError(noTurnReconcileErr)
	s.Zero(tdb.Client(ctx).AgentTurn.Query().Where(agentturn.AgentSessionID(investigation.AgentSessionID)).CountX(ctx))
	s.Nil(tdb.Client(ctx).InvestigationUserInput.GetX(ctx, input.ID).AgentTurnID)

	s.createTerminalTurn(ctx, tdb, investigation.AgentSessionID, 1)
	queueFailure := errors.New("agent turn queue unavailable")
	s.expectInvokeJobs(jobService, 1, queueFailure)
	failedReconcileErr := service.ReconcileInvestigation(ctx, investigation.ID)
	s.ErrorContains(failedReconcileErr, "agent turn queue unavailable")
	s.Equal(1, tdb.Client(ctx).AgentTurn.Query().Where(agentturn.AgentSessionID(investigation.AgentSessionID)).CountX(ctx))
	s.Nil(tdb.Client(ctx).InvestigationUserInput.GetX(ctx, input.ID).AgentTurnID)
}

func (s *InvestigationServiceSuite) TestEvidenceAnalysisChangeAndRevisionShareCallerTransaction() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
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
		_, requestErr := service.RecordInvestigationEvidenceRevision(txCtx, rez.RecordInvestigationEvidenceRevisionParams{
			InvestigationID: investigation.ID,
			Explanation:     "New evidence was linked to the analysis.",
			CallerKey:       "analysis-update-rollback",
		})
		if requestErr != nil {
			return requestErr
		}
		return rollbackErr
	})
	s.ErrorIs(transactionErr, rollbackErr)

	analysis := tdb.Client(ctx).SystemAnalysis.GetX(ctx, investigation.SystemAnalysisID)
	s.Nil(analysis.ReferenceTime)
	s.Zero(tdb.Client(ctx).InvestigationEvidenceRevision.Query().Where(
		investigationevidencerevision.InvestigationID(investigation.ID),
	).CountX(ctx))
}

func (s *InvestigationServiceSuite) TestTerminalEventAndInputRaceCanDispatchDuplicateJobsOnlyOnce() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
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
		_, submitErr := service.SubmitInvestigationUserInput(userCtx, rez.SubmitInvestigationUserInputParams{
			InvestigationID: investigation.ID,
			Text:            "Question submitted as the previous turn completes",
			SubmissionKey:   "terminal-race",
		})
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

	s.Equal(2, tdb.Client(ctx).AgentTurn.Query().Where(agentturn.AgentSessionID(investigation.AgentSessionID)).CountX(ctx))
	inputs := tdb.Client(ctx).InvestigationUserInput.Query().
		Where(investigationuserinput.InvestigationID(investigation.ID), investigationuserinput.AgentTurnIDNotNil()).
		AllX(ctx)
	s.Require().Len(inputs, 1)
	s.Equal(2, tdb.Client(ctx).AgentTurn.GetX(ctx, *inputs[0].AgentTurnID).Sequence)
}
