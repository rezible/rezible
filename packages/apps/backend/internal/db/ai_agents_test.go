package db

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/firebase/genkit/go/core"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	as "github.com/rezible/rezible/ent/agentsession"
	at "github.com/rezible/rezible/ent/agentturn"
	"github.com/rezible/rezible/ent/predicate"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/rezible/rezible/test"
	"github.com/rezible/rezible/test/mocks"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type AgentSessionServiceSuite struct {
	test.Suite
}

func TestAgentSessionServiceSuite(t *testing.T) {
	suite.Run(t, &AgentSessionServiceSuite{Suite: test.NewSuite()})
}

type agentSessionTestHarness struct {
	service  *AgentSessionService
	worker   *agentTurnWorker
	jobs     *mocks.MockJobService
	ai       *mocks.MockAiService
	messages *mocks.MockMessageService
}

type agentTurnSeed struct {
	id           uuid.UUID
	riverJobID   int64
	parentID     *uuid.UUID
	input        *rez.AgentTurnInput
	status       at.Status
	state        []byte
	error        []byte
	finishReason string
	createdAt    time.Time
	startedAt    *time.Time
	finishedAt   *time.Time
}

func (s *AgentSessionServiceSuite) newAgentSessionTestHarness() *agentSessionTestHarness {
	jobService := mocks.NewMockJobService(s.T())
	aiService := mocks.NewMockAiService(s.T())
	messageService := mocks.NewMockMessageService(s.T())
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	return &agentSessionTestHarness{
		service: &AgentSessionService{
			logger: logger,
			db:     s.Database(),
			jobs:   jobService,
			ai:     aiService,
		},
		worker: &agentTurnWorker{
			db:     s.Database(),
			ai:     aiService,
			msgs:   messageService,
			logger: logger,
		},
		jobs:     jobService,
		ai:       aiService,
		messages: messageService,
	}
}

func (s *AgentSessionServiceSuite) createAgentSession(ctx context.Context, ownerID uuid.UUID) *ent.AgentSession {
	createSession := s.Client(ctx).AgentSession.Create().
		SetAgentName("test-agent").
		SetOwnerUserID(ownerID)
	session, createErr := createSession.Save(ctx)
	s.Require().NoError(createErr)
	return session
}

func (s *AgentSessionServiceSuite) createAgentTurn(ctx context.Context, session *ent.AgentSession, seed agentTurnSeed) *ent.AgentTurn {
	if seed.id == uuid.Nil {
		seed.id = uuid.New()
	}
	if seed.input == nil {
		seed.input = &rez.AgentTurnInput{Message: ai.NewUserTextMessage("test")}
	}

	encodedInput, encodeErr := json.Marshal(seed.input)
	s.Require().NoError(encodeErr)

	createTurn := s.Client(ctx).AgentTurn.Create().
		SetID(seed.id).
		SetAgentSessionID(session.ID).
		SetRiverJobID(seed.riverJobID).
		SetInput(encodedInput).
		SetStatus(seed.status).
		SetNillableParentID(seed.parentID).
		SetNillableStartedAt(seed.startedAt).
		SetNillableFinishedAt(seed.finishedAt)
	if !seed.createdAt.IsZero() {
		createTurn.SetCreatedAt(seed.createdAt)
	}
	if seed.state != nil {
		createTurn.SetState(seed.state)
	}
	if seed.error != nil {
		createTurn.SetError(seed.error)
	}
	if seed.finishReason != "" {
		createTurn.SetFinishReason(seed.finishReason)
	}

	turn, createErr := createTurn.Save(ctx)
	s.Require().NoError(createErr)
	return turn
}

func (s *AgentSessionServiceSuite) userContext(user *ent.User) context.Context {
	authSession := &ent.UserAuthSession{
		TenantID:  s.SeedTenant.ID,
		UserID:    user.ID,
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}
	return execution.NewUserContext(s.T().Context(), authSession)
}

func makeAgentTurnJob(turn *ent.AgentTurn, attempt int) *river.Job[jobs.InvokeAgentTurn] {
	return &river.Job[jobs.InvokeAgentTurn]{
		JobRow: &rivertype.JobRow{
			ID:          turn.RiverJobID,
			Attempt:     attempt,
			MaxAttempts: 3,
		},
		Args: jobs.InvokeAgentTurn{
			AgentSessionID: turn.AgentSessionID,
			AgentTurnID:    turn.ID,
		},
	}
}

func makeJobInsertResult(id int64) *rivertype.JobInsertResult {
	return &rivertype.JobInsertResult{Job: &rivertype.JobRow{ID: id}}
}

func (s *AgentSessionServiceSuite) TestCreateAgentSessionCreatesQueuedRootAtomically() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionTestHarness()
	definitionInput := map[string]any{"alertId": "alert-1"}
	initialInput := &rez.AgentTurnInput{Message: ai.NewUserTextMessage("initial")}

	h.ai.EXPECT().
		MakeInitialAgentTurnInput(mock.Anything, "test-agent", definitionInput).
		Return(initialInput, nil).
		Once()
	h.jobs.EXPECT().
		Insert(mock.Anything, mock.Anything, (*river.InsertOpts)(nil)).
		Return(makeJobInsertResult(101), nil).
		Once()

	params := rez.CreateAgentSessionParams{
		AgentName:        "  test-agent  ",
		OwnerUserID:      s.SeedUser.ID,
		PermissionScopes: []string{"alerts:read"},
		Input:            definitionInput,
		Metadata:         map[string]any{"channel": "C123"},
	}
	session, createErr := h.service.CreateAgentSession(ctx, params)
	s.Require().NoError(createErr)
	s.Require().NotNil(session)
	s.Equal("test-agent", session.AgentName)
	s.Equal(s.SeedUser.ID, session.OwnerUserID)
	s.Equal([]string{"alerts:read"}, session.DefaultScopes)
	s.Equal("C123", session.Metadata["channel"])
	s.Require().Len(session.Edges.Turns, 1)

	turn := session.Edges.Turns[0]
	s.Equal(at.StatusQueued, turn.Status)
	s.Equal(int64(101), turn.RiverJobID)
	s.Nil(turn.ParentID)

	var storedInput rez.AgentTurnInput
	decodeErr := json.Unmarshal(turn.Input, &storedInput)
	s.Require().NoError(decodeErr)
	s.Require().NotNil(storedInput.Message)
	s.Equal("initial", storedInput.Message.Text())

	s.Require().Len(h.jobs.Calls, 1)
	insertedArgs, argsOK := h.jobs.Calls[0].Arguments.Get(1).(jobs.InvokeAgentTurn)
	s.Require().True(argsOK)
	s.Equal(session.ID, insertedArgs.AgentSessionID)
	s.Equal(turn.ID, insertedArgs.AgentTurnID)
}

func (s *AgentSessionServiceSuite) TestCreateAgentSessionRollsBackWhenJobInsertFails() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionTestHarness()
	initialInput := &rez.AgentTurnInput{Message: ai.NewUserTextMessage("initial")}
	insertErr := errors.New("job insert failed")

	querySessionsBefore := s.Client(ctx).AgentSession.Query()
	sessionsBefore, countSessionsErr := querySessionsBefore.Count(ctx)
	s.Require().NoError(countSessionsErr)
	queryTurnsBefore := s.Client(ctx).AgentTurn.Query()
	turnsBefore, countTurnsErr := queryTurnsBefore.Count(ctx)
	s.Require().NoError(countTurnsErr)

	h.ai.EXPECT().
		MakeInitialAgentTurnInput(mock.Anything, "test-agent", mock.Anything).
		Return(initialInput, nil).
		Once()
	h.jobs.EXPECT().
		Insert(mock.Anything, mock.Anything, (*river.InsertOpts)(nil)).
		Return(nil, insertErr).
		Once()

	params := rez.CreateAgentSessionParams{
		AgentName:   "test-agent",
		OwnerUserID: s.SeedUser.ID,
	}
	session, createErr := h.service.CreateAgentSession(ctx, params)
	s.Nil(session)
	s.ErrorIs(createErr, insertErr)

	querySessionsAfter := s.Client(ctx).AgentSession.Query()
	sessionsAfter, countSessionsErr := querySessionsAfter.Count(ctx)
	s.Require().NoError(countSessionsErr)
	queryTurnsAfter := s.Client(ctx).AgentTurn.Query()
	turnsAfter, countTurnsErr := queryTurnsAfter.Count(ctx)
	s.Require().NoError(countTurnsErr)
	s.Equal(sessionsBefore, sessionsAfter)
	s.Equal(turnsBefore, turnsAfter)
}

func (s *AgentSessionServiceSuite) TestRequestAgentTurnValidatesInputAndCompletedRoot() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionTestHarness()
	session := s.createAgentSession(ctx, s.SeedUser.ID)
	root := s.createAgentTurn(ctx, session, agentTurnSeed{
		riverJobID: 201,
		status:     at.StatusQueued,
	})
	validParams := &rez.RequestAgentTurnParams{
		Input: &rez.AgentTurnInput{Message: ai.NewUserTextMessage("follow up")},
	}

	turn, requestErr := h.service.RequestAgentTurn(ctx, session.ID, validParams)
	s.Nil(turn)
	s.ErrorIs(requestErr, rez.ErrConflict)

	completeRoot := root.Update().
		SetStatus(at.StatusCompleted).
		SetState([]byte(`{"root":true}`)).
		SetFinishReason(string(aix.AgentFinishReasonStop)).
		SetFinishedAt(time.Now().UTC())
	root, completeErr := completeRoot.Save(ctx)
	s.Require().NoError(completeErr)

	turn, requestErr = h.service.RequestAgentTurn(ctx, session.ID, nil)
	s.Nil(turn)
	s.ErrorIs(requestErr, rez.ErrInvalidInput)

	emptyParams := &rez.RequestAgentTurnParams{
		Input: &rez.AgentTurnInput{Resume: &ai.GenerateActionResume{}},
	}
	turn, requestErr = h.service.RequestAgentTurn(ctx, session.ID, emptyParams)
	s.Nil(turn)
	s.ErrorIs(requestErr, rez.ErrInvalidInput)

	h.jobs.EXPECT().
		Insert(mock.Anything, mock.Anything, (*river.InsertOpts)(nil)).
		Return(makeJobInsertResult(202), nil).
		Once()
	turn, requestErr = h.service.RequestAgentTurn(ctx, session.ID, validParams)
	s.Require().NoError(requestErr)
	s.Require().NotNil(turn)
	s.Equal(at.StatusQueued, turn.Status)
	s.Equal(int64(202), turn.RiverJobID)
	s.Nil(turn.ParentID)
	s.Equal(root.AgentSessionID, turn.AgentSessionID)
}

func (s *AgentSessionServiceSuite) TestAgentSessionAccessIsOwnerScoped() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionTestHarness()
	session := s.createAgentSession(ctx, s.SeedUser.ID)
	turn := s.createAgentTurn(ctx, session, agentTurnSeed{
		riverJobID: 301,
		status:     at.StatusQueued,
	})

	createOtherUser := s.Client(ctx).User.Create().
		SetEmail("other+" + uuid.NewString() + "@example.com").
		SetName("Other User")
	otherUser, createUserErr := createOtherUser.Save(ctx)
	s.Require().NoError(createUserErr)

	ownerCtx := s.userContext(s.SeedUser)
	otherCtx := s.userContext(otherUser)
	ownedSession, getSessionErr := h.service.GetAgentSession(ownerCtx, session.ID)
	s.Require().NoError(getSessionErr)
	s.Equal(session.ID, ownedSession.ID)

	_, getSessionErr = h.service.GetAgentSession(otherCtx, session.ID)
	s.True(ent.IsNotFound(getSessionErr))
	_, getTurnErr := h.service.GetAgentTurn(otherCtx, turn.ID)
	s.True(ent.IsNotFound(getTurnErr))

	listSessionsParams := rez.ListAgentSessionsParams{
		ListParams: ent.ListParams{Count: true},
		Predicates: []predicate.AgentSession{as.ID(session.ID)},
	}
	listedSessions, listSessionsErr := h.service.ListAgentSessions(otherCtx, listSessionsParams)
	s.Require().NoError(listSessionsErr)
	s.Empty(listedSessions.Data)
	s.Zero(listedSessions.Count)

	_, abortErr := h.service.AbortAgentTurn(otherCtx, turn.ID)
	s.True(ent.IsNotFound(abortErr))
}

func (s *AgentSessionServiceSuite) TestClaimAgentTurnUsesFIFOAndSuccessfulLeafState() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionTestHarness()
	session := s.createAgentSession(ctx, s.SeedUser.ID)
	baseTime := time.Now().UTC().Add(-3 * time.Minute)
	rootState := []byte(`{"root":true}`)
	root := s.createAgentTurn(ctx, session, agentTurnSeed{
		riverJobID:   401,
		status:       at.StatusCompleted,
		state:        rootState,
		finishReason: string(aix.AgentFinishReasonStop),
		createdAt:    baseTime,
	})
	older := s.createAgentTurn(ctx, session, agentTurnSeed{
		riverJobID: 402,
		status:     at.StatusQueued,
		createdAt:  baseTime.Add(time.Minute),
	})
	younger := s.createAgentTurn(ctx, session, agentTurnSeed{
		riverJobID: 403,
		status:     at.StatusQueued,
		createdAt:  baseTime.Add(2 * time.Minute),
	})

	claim, claimErr := h.worker.claimAgentTurn(ctx, makeAgentTurnJob(younger, 1))
	s.Nil(claim)
	s.True(errors.Is(claimErr, &river.JobSnoozeError{}))

	queryYounger := s.Client(ctx).AgentTurn.Query().Where(at.ID(younger.ID))
	younger, queryYoungerErr := queryYounger.Only(ctx)
	s.Require().NoError(queryYoungerErr)
	s.Equal(at.StatusQueued, younger.Status)

	claim, claimErr = h.worker.claimAgentTurn(ctx, makeAgentTurnJob(older, 1))
	s.Require().NoError(claimErr)
	s.Require().NotNil(claim)
	s.Equal(rootState, claim.parent.State)
	s.Require().NotNil(claim.turn.ParentID)
	s.Equal(root.ID, *claim.turn.ParentID)
	s.Equal(at.StatusRunning, claim.turn.Status)
	s.NotNil(claim.turn.StartedAt)
}

func (s *AgentSessionServiceSuite) TestWorkerPersistsSuccessfulResultAndPublishesEvent() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionTestHarness()
	session := s.createAgentSession(ctx, s.SeedUser.ID)
	rootState := []byte(`{"root":true}`)
	resultState := []byte(`{"messages":["done"]}`)
	root := s.createAgentTurn(ctx, session, agentTurnSeed{
		riverJobID:   501,
		status:       at.StatusCompleted,
		state:        rootState,
		finishReason: string(aix.AgentFinishReasonStop),
	})
	input := &rez.AgentTurnInput{Message: ai.NewUserTextMessage("continue")}
	turn := s.createAgentTurn(ctx, session, agentTurnSeed{
		riverJobID: 502,
		status:     at.StatusQueued,
		input:      input,
	})
	response := ai.NewModelTextMessage("done")
	result := &rez.AgentInvocationResult{
		State:        resultState,
		Response:     response,
		FinishReason: aix.AgentFinishReasonStop,
	}

	h.ai.EXPECT().
		InvokeAgentTurn(mock.Anything, mock.Anything).
		Return(result, nil).
		Once()
	h.messages.EXPECT().
		Publish(mock.Anything, mock.Anything).
		Return(nil).
		Once()

	workErr := h.worker.Work(ctx, makeAgentTurnJob(turn, 1))
	s.Require().NoError(workErr)

	queryTurn := s.Client(ctx).AgentTurn.Query().Where(at.ID(turn.ID))
	turn, queryTurnErr := queryTurn.Only(ctx)
	s.Require().NoError(queryTurnErr)
	s.Equal(at.StatusCompleted, turn.Status)
	s.Equal(resultState, turn.State)
	s.Nil(turn.Error)
	s.Equal(string(aix.AgentFinishReasonStop), turn.FinishReason)
	s.NotNil(turn.FinishedAt)
	s.Require().NotNil(turn.ParentID)
	s.Equal(root.ID, *turn.ParentID)

	s.Require().Len(h.messages.Calls, 1)
	published, eventOK := h.messages.Calls[0].Arguments.Get(1).(*rezai.EventOnAgentTurnFinished)
	s.Require().True(eventOK)
	s.Equal(session.ID, published.AgentSessionId)
	s.Equal(turn.ID, published.AgentTurnId)
	s.Equal(response, published.Response)
}

func (s *AgentSessionServiceSuite) TestWorkerTerminalizesRunningRedeliveryWithoutInvokingAgent() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionTestHarness()
	session := s.createAgentSession(ctx, s.SeedUser.ID)
	root := s.createAgentTurn(ctx, session, agentTurnSeed{
		riverJobID:   601,
		status:       at.StatusCompleted,
		state:        []byte(`{"root":true}`),
		finishReason: string(aix.AgentFinishReasonStop),
	})
	turn := s.createAgentTurn(ctx, session, agentTurnSeed{
		riverJobID: 602,
		parentID:   &root.ID,
		status:     at.StatusRunning,
		state:      []byte(`{"partial":true}`),
		startedAt:  new(time.Now().UTC().Add(-time.Minute)),
	})

	workErr := h.worker.Work(ctx, makeAgentTurnJob(turn, 2))
	s.Require().NoError(workErr)

	queryTurn := s.Client(ctx).AgentTurn.Query().Where(at.ID(turn.ID))
	turn, queryTurnErr := queryTurn.Only(ctx)
	s.Require().NoError(queryTurnErr)
	s.Equal(at.StatusFailed, turn.Status)
	s.Nil(turn.State)
	s.True(json.Valid(turn.Error))
	s.Equal(string(aix.AgentFinishReasonFailed), turn.FinishReason)
	s.NotNil(turn.FinishedAt)
	s.Empty(h.ai.Calls)
	s.Empty(h.messages.Calls)
}

func (s *AgentSessionServiceSuite) TestWorkerPersistsFailedResultWithLastGoodState() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionTestHarness()
	session := s.createAgentSession(ctx, s.SeedUser.ID)
	rootState := []byte(`{"root":true}`)
	lastGoodState := []byte(`{"lastGood":true}`)
	s.createAgentTurn(ctx, session, agentTurnSeed{
		riverJobID:   651,
		status:       at.StatusCompleted,
		state:        rootState,
		finishReason: string(aix.AgentFinishReasonStop),
	})
	turn := s.createAgentTurn(ctx, session, agentTurnSeed{
		riverJobID: 652,
		status:     at.StatusQueued,
	})
	result := &rez.AgentInvocationResult{
		State:        lastGoodState,
		FinishReason: aix.AgentFinishReasonFailed,
		Error:        core.AsGenkitError(errors.New("model failed")),
	}

	h.ai.EXPECT().
		InvokeAgentTurn(mock.Anything, mock.Anything).
		Return(result, nil).
		Once()
	h.messages.EXPECT().
		Publish(mock.Anything, mock.Anything).
		Return(nil).
		Once()

	workErr := h.worker.Work(ctx, makeAgentTurnJob(turn, 1))
	s.Require().NoError(workErr)

	queryTurn := s.Client(ctx).AgentTurn.Query().Where(at.ID(turn.ID))
	turn, queryTurnErr := queryTurn.Only(ctx)
	s.Require().NoError(queryTurnErr)
	s.Equal(at.StatusFailed, turn.Status)
	s.Equal(lastGoodState, turn.State)
	s.True(json.Valid(turn.Error))
	s.Equal(string(aix.AgentFinishReasonFailed), turn.FinishReason)
	s.NotNil(turn.FinishedAt)
}

func (s *AgentSessionServiceSuite) TestAbortAgentTurnIsIdempotent() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionTestHarness()
	session := s.createAgentSession(ctx, s.SeedUser.ID)
	turn := s.createAgentTurn(ctx, session, agentTurnSeed{
		riverJobID: 701,
		status:     at.StatusQueued,
		state:      []byte(`{"stale":true}`),
		error:      []byte(`{"message":"stale"}`),
	})

	h.jobs.EXPECT().
		Cancel(mock.Anything, int64(701)).
		Return(nil).
		Once()

	aborted, abortErr := h.service.AbortAgentTurn(ctx, turn.ID)
	s.Require().NoError(abortErr)
	s.Equal(at.StatusAborted, aborted.Status)
	s.Equal(string(aix.AgentFinishReasonAborted), aborted.FinishReason)
	s.Nil(aborted.State)
	s.Nil(aborted.Error)
	s.NotNil(aborted.FinishedAt)

	abortedAgain, abortErr := h.service.AbortAgentTurn(ctx, turn.ID)
	s.Require().NoError(abortErr)
	s.Equal(aborted.ID, abortedAgain.ID)
	s.Equal(at.StatusAborted, abortedAgain.Status)
}

func (s *AgentSessionServiceSuite) TestRetryAgentTurnRequeuesSameTurnAndClearsTerminalState() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionTestHarness()
	session := s.createAgentSession(ctx, s.SeedUser.ID)
	root := s.createAgentTurn(ctx, session, agentTurnSeed{
		riverJobID:   801,
		status:       at.StatusCompleted,
		state:        []byte(`{"root":true}`),
		finishReason: string(aix.AgentFinishReasonStop),
	})
	originalInput := &rez.AgentTurnInput{Message: ai.NewUserTextMessage("retry me")}
	turn := s.createAgentTurn(ctx, session, agentTurnSeed{
		riverJobID:   802,
		parentID:     &root.ID,
		input:        originalInput,
		status:       at.StatusFailed,
		state:        []byte(`{"lastGood":true}`),
		error:        []byte(`{"message":"failed"}`),
		finishReason: string(aix.AgentFinishReasonFailed),
		startedAt:    new(time.Now().UTC().Add(-2 * time.Minute)),
		finishedAt:   new(time.Now().UTC().Add(-time.Minute)),
	})
	originalEncodedInput := append([]byte(nil), turn.Input...)

	h.jobs.EXPECT().
		Insert(mock.Anything, mock.Anything, (*river.InsertOpts)(nil)).
		Return(makeJobInsertResult(803), nil).
		Once()

	retried, retryErr := h.service.RetryAgentTurn(ctx, turn.ID)
	s.Require().NoError(retryErr)
	s.Equal(turn.ID, retried.ID)
	s.Equal(at.StatusQueued, retried.Status)
	s.Equal(int64(803), retried.RiverJobID)
	s.Equal(originalEncodedInput, retried.Input)
	s.Nil(retried.State)
	s.Nil(retried.Error)
	s.Nil(retried.StartedAt)
	s.Nil(retried.FinishedAt)
	s.Empty(retried.FinishReason)

	s.Require().Len(h.jobs.Calls, 1)
	insertedArgs, argsOK := h.jobs.Calls[0].Arguments.Get(1).(jobs.InvokeAgentTurn)
	s.Require().True(argsOK)
	s.Equal(session.ID, insertedArgs.AgentSessionID)
	s.Equal(turn.ID, insertedArgs.AgentTurnID)
}
