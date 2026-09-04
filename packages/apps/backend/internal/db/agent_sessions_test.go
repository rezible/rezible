package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/agentmessage"
	asb "github.com/rezible/rezible/ent/agentsessionbinding"
	at "github.com/rezible/rezible/ent/agentturn"
	"github.com/rezible/rezible/ent/predicate"
	rezai "github.com/rezible/rezible/pkg/ai"
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
	tdb        rez.Database
	jobs       *mocks.MockJobService
	ai         *mocks.MockAiService
	msgs       *mocks.MockMessageService
	service    *AgentSessionService
	sessWorker *StartAgentSessionWorker
	turnWorker *InvokeAgentTurnWorker
}

func (s *AgentSessionServiceSuite) newAgentSessionTestHarness() *agentSessionTestHarness {
	tdb := s.CreateTestDatabase()
	jobService := mocks.NewMockJobService(s.T())
	aiService := mocks.NewMockAiService(s.T())
	messageService := mocks.NewMockMessageService(s.T())
	messageService.EXPECT().
		Publish(mock.Anything, mock.IsType(rezai.AgentTurnUpdated{})).Return(nil).Maybe()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	svc := &AgentSessionService{
		logger: logger,
		db:     tdb,
		jobs:   jobService,
		msgs:   messageService,
	}

	sessWorker := &StartAgentSessionWorker{
		db:     tdb,
		ai:     aiService,
		aiSess: svc,
		logger: logger,
	}

	turnWorker := &InvokeAgentTurnWorker{
		db:     tdb,
		ai:     aiService,
		msgs:   messageService,
		logger: logger,
	}

	return &agentSessionTestHarness{
		tdb:        tdb,
		jobs:       jobService,
		ai:         aiService,
		msgs:       messageService,
		service:    svc,
		sessWorker: sessWorker,
		turnWorker: turnWorker,
	}
}

func (s *AgentSessionServiceSuite) createAgentSession(ctx context.Context, tdb rez.Database, input rez.ValidatingInput) *ent.AgentSession {
	inputJson, jsonErr := json.Marshal(input)
	s.Require().NoError(jsonErr)
	createSession := tdb.Client(ctx).AgentSession.Create().
		SetAgentName("test-agent").
		SetInput(inputJson)
	session, createErr := createSession.Save(ctx)
	s.Require().NoError(createErr)
	return session
}

type agentTurnSeed struct {
	id           uuid.UUID
	riverJobID   int64
	input        *rez.AiAgentTurnInput
	status       at.Status
	error        error
	finishReason string
	createdAt    time.Time
	startedAt    *time.Time
	finishedAt   *time.Time
}

func (s *AgentSessionServiceSuite) createAgentTurn(ctx context.Context, tdb rez.Database, session *ent.AgentSession, seed agentTurnSeed) *ent.AgentTurn {
	if seed.id == uuid.Nil {
		seed.id = uuid.New()
	}
	if seed.input == nil {
		seed.input = &rez.AiAgentTurnInput{Message: ai.NewUserTextMessage("test")}
	}

	input, inputErr := normalizeAgentTurnInput(seed.input)
	s.Require().NoError(inputErr)

	var turn *ent.AgentTurn
	txErr := tdb.WithTx(ctx, func(ctx context.Context, client *ent.Client) error {
		sess := client.AgentSession.GetX(ctx, session.ID)

		numTurns := sess.QueryTurns().CountX(ctx)
		numMsgs := sess.QueryMessages().CountX(ctx)

		createTurn := client.AgentTurn.Create().
			SetID(seed.id).
			SetAgentSessionID(session.ID).
			SetRiverJobID(seed.riverJobID).
			SetSequence(numTurns + 1).
			SetStatus(seed.status).
			SetNillableStartedAt(seed.startedAt).
			SetNillableFinishedAt(seed.finishedAt)
		if !seed.createdAt.IsZero() {
			createTurn.SetCreatedAt(seed.createdAt)
		}

		if input.Resume != nil {
			createTurn.SetInputToolResume(input.Resume)
		}
		if seed.error != nil {
			createTurn.SetError(seed.error.Error())
		}
		if seed.finishReason != "" {
			createTurn.SetFinishReason(seed.finishReason)
		}
		createdTurn := createTurn.SaveX(ctx)

		if inputMsg := seed.input.Message; inputMsg != nil {
			createMsg := client.AgentMessage.Create().
				SetAgentSessionID(session.ID).
				SetAgentTurnID(createdTurn.ID).
				SetSequence(numMsgs + 1).
				SetRole(agentmessage.Role(inputMsg.Role)).
				SetContent(inputMsg.Content).
				SetMetadata(inputMsg.Metadata)
			createdMsg := createMsg.SaveX(ctx)

			createdTurn = createdTurn.Update().SetInputMessageID(createdMsg.ID).SaveX(ctx)

			createdTurn.Edges.Messages = append(createdTurn.Edges.Messages, createdMsg.Unwrap())
		}

		turn = createdTurn.Unwrap()

		return nil
	})
	s.Require().NoError(txErr)

	return turn
}

func (s *AgentSessionServiceSuite) createAgentMessage(ctx context.Context, tdb rez.Database, session *ent.AgentSession, turn *ent.AgentTurn, msg *ai.Message) *ent.AgentMessage {
	numMsgs, countErr := session.QueryMessages().Count(ctx)
	s.Require().NoError(countErr)

	createMsg := tdb.Client(ctx).AgentMessage.Create().
		SetAgentSessionID(session.ID).
		SetAgentTurnID(turn.ID).
		SetSequence(numMsgs + 1).
		SetRole(agentmessage.Role(msg.Role)).
		SetContent(msg.Content).
		SetMetadata(msg.Metadata)
	createdMsg, createErr := createMsg.Save(ctx)
	s.Require().NoError(createErr)
	return createdMsg
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

type testAgentInput struct {
	Foo string
}

func (i testAgentInput) Validate() error {
	return nil
}

func (s *AgentSessionServiceSuite) TestCreateAgentSessionCreatesQueuedStartAtomically() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionTestHarness()
	analysis := h.tdb.Client(ctx).SystemAnalysis.Create().SaveX(ctx)

	h.jobs.EXPECT().
		Insert(mock.Anything, mock.IsType(jobs.StartAgentSession{}), mock.Anything).
		Return(makeJobInsertResult(101), nil).
		Once()

	params := rez.CreateAgentSessionParams{
		AgentName:        "test-agent",
		Input:            testAgentInput{Foo: "bar"},
		SystemAnalysisID: &analysis.ID,
		Metadata:         map[string]any{"baz": "123"},
	}

	session, createErr := h.service.CreateAgentSession(ctx, params)
	s.Require().NoError(createErr)
	s.Require().NotNil(session)
	s.Equal("test-agent", session.AgentName)
	s.Require().NotNil(session.SystemAnalysisID)
	s.Equal(analysis.ID, *session.SystemAnalysisID)
	s.Equal("123", session.Metadata["baz"])
}

func (s *AgentSessionServiceSuite) TestCreateAgentSessionCreatesRequestedBindings() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionTestHarness()
	resourceRef := uuid.NewString()

	h.jobs.EXPECT().
		Insert(mock.Anything, mock.IsType(jobs.StartAgentSession{}), mock.Anything).
		Return(makeJobInsertResult(101), nil).
		Once()

	bindingRef := rez.ProviderResourceRef{
		Provider:    "rezible",
		ResourceRef: resourceRef,
	}
	bindingParams := rez.AgentSessionBindingParams{
		ProviderResourceRef: bindingRef,
		Metadata:            map[string]any{"origin": "test"},
	}
	params := rez.CreateAgentSessionParams{
		AgentName: "test-agent",
		Input:     testAgentInput{Foo: "bar"},
		Bindings:  []rez.AgentSessionBindingParams{bindingParams},
	}

	session, createErr := h.service.CreateAgentSession(ctx, params)
	s.Require().NoError(createErr)
	s.Require().NotNil(session)

	bindingPreds := []predicate.AgentSessionBinding{
		asb.Provider(bindingParams.Provider),
		asb.ProviderNamespace(bindingParams.ProviderNamespace),
		asb.ProviderResourceRef(resourceRef),
	}
	binding, bindingErr := h.service.LookupAgentSessionBinding(ctx, bindingPreds...)
	s.Require().NoError(bindingErr)
	s.Equal(session.ID, binding.AgentSessionID)
	s.Nil(binding.IntegrationID)
	s.Equal("test", binding.Metadata["origin"])
}

func (s *AgentSessionServiceSuite) TestCreateAgentSessionRejectsExternalBindingWithoutNamespace() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionTestHarness()
	params := rez.CreateAgentSessionParams{
		AgentName: "test-agent",
		Input:     testAgentInput{Foo: "bar"},
		Bindings: []rez.AgentSessionBindingParams{{
			ProviderResourceRef: rez.ProviderResourceRef{Provider: "slack", ResourceRef: "thread:C123:123.456"},
		}},
	}

	session, createErr := h.service.CreateAgentSession(ctx, params)
	s.Nil(session)
	s.ErrorIs(createErr, rez.ErrInvalidInput)
	s.Zero(h.tdb.Client(ctx).AgentSession.Query().CountX(ctx))
	s.Zero(h.tdb.Client(ctx).AgentSessionBinding.Query().CountX(ctx))
}

func (s *AgentSessionServiceSuite) TestCreateAgentSessionRejectsIntegrationProviderMismatchAtomically() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionTestHarness()
	intg := h.tdb.Client(ctx).Integration.Create().
		SetProvider("github").
		SetName("github").
		SetDisplayName("GitHub").
		SetProviderInstallationRef("org-1").
		SetInstallationConfig([]byte(`{}`)).
		SaveX(ctx)
	params := rez.CreateAgentSessionParams{
		AgentName: "test-agent",
		Input:     testAgentInput{Foo: "bar"},
		Bindings: []rez.AgentSessionBindingParams{{
			ProviderResourceRef: rez.ProviderResourceRef{
				Provider:          "slack",
				ProviderNamespace: "T123",
				ResourceRef:       "thread:C123:123.456",
			},
			IntegrationID: &intg.ID,
		}},
	}

	session, createErr := h.service.CreateAgentSession(ctx, params)
	s.Nil(session)
	s.ErrorIs(createErr, rez.ErrInvalidInput)
	s.Zero(h.tdb.Client(ctx).AgentSession.Query().CountX(ctx))
	s.Zero(h.tdb.Client(ctx).AgentSessionBinding.Query().CountX(ctx))
}

func (s *AgentSessionServiceSuite) TestSetAgentSessionBindingCreatesBindingAfterSession() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionTestHarness()
	session := s.createAgentSession(ctx, h.tdb, testAgentInput{Foo: "bar"})
	intg := h.tdb.Client(ctx).Integration.Create().
		SetProvider("slack").
		SetName("slack_agent").
		SetDisplayName("Slack Agent").
		SetProviderInstallationRef("team:T123").
		SetInstallationConfig([]byte(`{}`)).
		SaveX(ctx)
	opaqueRef := " thread:C123:123.456 "

	binding, setErr := h.service.SetAgentSessionBinding(ctx, uuid.Nil, func(m *ent.AgentSessionBindingMutation) {
		m.SetAgentSessionID(session.ID)
		m.SetIntegrationID(intg.ID)
		m.SetProvider("slack")
		m.SetProviderNamespace("T123")
		m.SetProviderResourceRef(opaqueRef)
		m.SetMetadata(map[string]any{"origin": "follow-up"})
	})
	s.Require().NoError(setErr)
	s.Require().NotNil(binding)
	s.Equal(session.ID, binding.AgentSessionID)
	s.Equal(opaqueRef, binding.ProviderResourceRef)
	s.Equal("follow-up", binding.Metadata["origin"])

	updated, updateErr := h.service.SetAgentSessionBinding(ctx, binding.ID, func(m *ent.AgentSessionBindingMutation) {
		m.SetMetadata(map[string]any{"origin": "updated"})
	})
	s.Require().NoError(updateErr)
	s.Equal("updated", updated.Metadata["origin"])
	s.Equal(opaqueRef, updated.ProviderResourceRef)
}

func (s *AgentSessionServiceSuite) TestSetAgentSessionBindingValidatesNewBinding() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionTestHarness()
	session := s.createAgentSession(ctx, h.tdb, testAgentInput{Foo: "bar"})

	binding, setErr := h.service.SetAgentSessionBinding(ctx, uuid.Nil, func(m *ent.AgentSessionBindingMutation) {
		m.SetAgentSessionID(session.ID)
		m.SetProvider("slack")
		m.SetProviderResourceRef("thread:C123:123.456")
	})
	s.Nil(binding)
	s.ErrorIs(setErr, rez.ErrInvalidInput)
	s.Zero(h.tdb.Client(ctx).AgentSessionBinding.Query().CountX(ctx))
}

func (s *AgentSessionServiceSuite) TestSetAgentSessionBindingKeepsNamespacesDistinct() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionTestHarness()
	firstSession := s.createAgentSession(ctx, h.tdb, testAgentInput{Foo: "first"})
	secondSession := s.createAgentSession(ctx, h.tdb, testAgentInput{Foo: "second"})
	createBinding := func(sessionID uuid.UUID, namespace string) (*ent.AgentSessionBinding, error) {
		return h.service.SetAgentSessionBinding(ctx, uuid.Nil, func(m *ent.AgentSessionBindingMutation) {
			m.SetAgentSessionID(sessionID)
			m.SetProvider("slack")
			m.SetProviderNamespace(namespace)
			m.SetProviderResourceRef("thread:C123:123.456")
		})
	}

	first, firstErr := createBinding(firstSession.ID, "T123")
	s.Require().NoError(firstErr)
	second, secondErr := createBinding(secondSession.ID, "T456")
	s.Require().NoError(secondErr)
	s.NotEqual(first.ID, second.ID)
	s.Equal(2, h.tdb.Client(ctx).AgentSessionBinding.Query().CountX(ctx))
}

func (s *AgentSessionServiceSuite) TestCreateAgentSessionRollsBackWhenJobInsertFails() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionTestHarness()
	client := h.tdb.Client(ctx)

	insertJobErr := errors.New("job insert failed")
	h.jobs.EXPECT().
		Insert(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, insertJobErr).
		Once()

	sessionsBefore, countSessBeforeErr := client.AgentSession.Query().Count(ctx)
	s.Require().NoError(countSessBeforeErr)

	turnsBefore, countTurnsBeforeErr := client.AgentTurn.Query().Count(ctx)
	s.Require().NoError(countTurnsBeforeErr)

	params := rez.CreateAgentSessionParams{
		AgentName: "test-agent",
		Input:     testAgentInput{Foo: "bar"},
	}
	session, createErr := h.service.CreateAgentSession(ctx, params)
	s.Nil(session)
	s.ErrorIs(createErr, insertJobErr)

	sessionsAfter, countSessAfterErr := client.AgentSession.Query().Count(ctx)
	s.Require().NoError(countSessAfterErr)

	turnsAfter, countTurnsErr := client.AgentTurn.Query().Count(ctx)
	s.Require().NoError(countTurnsErr)

	s.Equal(sessionsBefore, sessionsAfter)
	s.Equal(turnsBefore, turnsAfter)
}

func (s *AgentSessionServiceSuite) TestRequestAgentTurnValidatesInputAndCompletedRoot() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionTestHarness()
	session := s.createAgentSession(ctx, h.tdb, testAgentInput{Foo: "bar"})

	initialTurn := s.createAgentTurn(ctx, h.tdb, session, agentTurnSeed{
		riverJobID: 201,
		status:     at.StatusQueued,
	})

	validParams := &rez.RequestAgentTurnParams{
		Input: &rez.AiAgentTurnInput{
			Message: ai.NewUserTextMessage("follow up"),
		},
	}

	turn, requestErr := h.service.RequestAgentTurn(ctx, session.ID, validParams)
	s.Nil(turn)
	s.ErrorIs(requestErr, rez.ErrConflict)

	setInitialCompleted := initialTurn.Update().
		SetStatus(at.StatusCompleted).
		SetFinishReason(string(aix.AgentFinishReasonStop)).
		SetFinishedAt(time.Now().UTC())
	s.Require().NoError(setInitialCompleted.Exec(ctx))

	turn, requestErr = h.service.RequestAgentTurn(ctx, session.ID, nil)
	s.Nil(turn)
	s.ErrorIs(requestErr, rez.ErrInvalidInput)

	emptyParams := &rez.RequestAgentTurnParams{
		Input: &rez.AiAgentTurnInput{Resume: &aix.ToolResume{}},
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
	s.Equal(2, turn.Sequence)
	s.Equal(int64(202), turn.RiverJobID)
	s.Equal(initialTurn.AgentSessionID, turn.AgentSessionID)
	s.Require().NotNil(turn.InputMessageID)

	inputMsg, inputMsgErr := turn.QueryInputMessage().Only(ctx)
	s.Require().NoError(inputMsgErr)
	s.Equal(agentmessage.RoleUser, inputMsg.Role)
	s.Equal("follow up", inputMsg.MakeGenkitMessage().Text())
}

func (s *AgentSessionServiceSuite) TestWorkerPersistsSuccessfulResultAndPublishesEvent() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionTestHarness()
	session := s.createAgentSession(ctx, h.tdb, testAgentInput{Foo: "bar"})

	initialTurn := s.createAgentTurn(ctx, h.tdb, session, agentTurnSeed{
		riverJobID:   501,
		status:       at.StatusCompleted,
		finishReason: string(aix.AgentFinishReasonStop),
	})
	s.Require().NotNil(initialTurn)
	initialReply := ai.NewModelTextMessage("ready")
	s.createAgentMessage(ctx, h.tdb, session, initialTurn, initialReply)

	followUpInput := ai.NewUserTextMessage("baz")
	followUp := s.createAgentTurn(ctx, h.tdb, session, agentTurnSeed{
		riverJobID: 502,
		status:     at.StatusQueued,
		input:      &rez.AiAgentTurnInput{Message: followUpInput},
	})
	s.Require().NotNil(followUp)

	followUpReply := ai.NewModelTextMessage("done")
	artifact := &aix.Artifact{
		Name:     "report",
		Parts:    []*ai.Part{ai.NewTextPart("new report")},
		Metadata: map[string]any{"version": "new"},
	}
	h.tdb.Client(ctx).AgentArtifact.Create().
		SetAgentSessionID(session.ID).
		SetAgentTurnID(initialTurn.ID).
		SetName(artifact.Name).
		SetParts([]*ai.Part{ai.NewTextPart("old report")}).
		SetMetadata(map[string]any{"version": "old"}).
		ExecX(ctx)

	turnResult := &rez.AiAgentInvocationResult{
		State: rez.AiAgentTurnState{
			Messages: []*ai.Message{
				initialTurn.Edges.Messages[0].MakeGenkitMessage(),
				initialReply,
				followUpInput,
				followUpReply,
			},
			Artifacts: []*aix.Artifact{artifact},
		},
		Response:     followUpReply,
		FinishReason: aix.AgentFinishReasonStop,
	}

	h.ai.EXPECT().
		InvokeAgentTurn(mock.Anything, mock.MatchedBy(func(params rez.InvokeAgentTurnParams) bool {
			return params.Session.ID == session.ID &&
				params.Turn.ID == followUp.ID &&
				params.Input != nil &&
				params.Input.Message != nil &&
				params.Input.Message.Text() == "baz" &&
				len(params.State.Messages) == 2 &&
				params.State.Messages[0].Text() == "test" &&
				params.State.Messages[1].Text() == "ready"
		})).
		Return(turnResult, nil).
		Once()

	h.msgs.EXPECT().
		Publish(mock.Anything, mock.MatchedBy(func(ev *rezai.EventOnAgentTurnFinished) bool {
			return ev.AgentSessionId == session.ID &&
				ev.AgentTurnId == followUp.ID &&
				ev.FinishReason == aix.AgentFinishReasonStop &&
				ev.Response == followUpReply
		})).
		Return(nil).
		Once()

	s.Require().NoError(h.turnWorker.Work(ctx, makeAgentTurnJob(followUp, 1)))

	queryTurn := h.tdb.Client(ctx).AgentTurn.Query().
		Where(at.ID(followUp.ID)).
		WithMessages().
		WithArtifacts()
	turn, queryTurnErr := queryTurn.Only(ctx)
	s.Require().NoError(queryTurnErr)
	s.Equal(at.StatusCompleted, turn.Status)
	s.Nil(turn.Error)
	s.Equal(string(aix.AgentFinishReasonStop), turn.FinishReason)
	s.NotNil(turn.FinishedAt)

	turnMsgs, queryMsgsErr := turn.QueryMessages().Order(agentmessage.BySequence()).All(ctx)
	s.Require().NoError(queryMsgsErr)
	s.Require().Len(turnMsgs, 2)
	s.Equal("baz", turnMsgs[0].MakeGenkitMessage().Text())
	s.Equal("done", turnMsgs[1].MakeGenkitMessage().Text())
	s.Equal(3, turnMsgs[0].Sequence)
	s.Equal(4, turnMsgs[1].Sequence)

	updatedArtifact, artifactErr := session.QueryArtifacts().Only(ctx)
	s.Require().NoError(artifactErr)
	s.Equal(followUp.ID, *updatedArtifact.LastAgentTurnID)
	s.Equal("new", updatedArtifact.Metadata["version"])
	s.Require().Len(updatedArtifact.Parts, 1)
	s.Equal("new report", updatedArtifact.Parts[0].Text)
}

func (s *AgentSessionServiceSuite) TestWorkerFailsTurnWhenAgentReturnsNilResult() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionTestHarness()
	session := s.createAgentSession(ctx, h.tdb, testAgentInput{})

	_ = s.createAgentTurn(ctx, h.tdb, session, agentTurnSeed{
		riverJobID:   551,
		status:       at.StatusCompleted,
		finishReason: string(aix.AgentFinishReasonStop),
	})
	turn := s.createAgentTurn(ctx, h.tdb, session, agentTurnSeed{
		riverJobID: 552,
		status:     at.StatusQueued,
	})

	h.ai.EXPECT().
		InvokeAgentTurn(mock.Anything, mock.Anything).
		Return(nil, nil).
		Once()

	workErr := h.turnWorker.Work(ctx, makeAgentTurnJob(turn, 1))
	s.Require().NoError(workErr)

	turn, turnErr := h.tdb.Client(ctx).AgentTurn.Get(ctx, turn.ID)
	s.Require().NoError(turnErr)
	s.Equal(at.StatusFailed, turn.Status)
	s.Require().NotNil(turn.Error)
	s.Contains(*turn.Error, "agent returned no result")
	s.Equal(string(aix.AgentFinishReasonFailed), turn.FinishReason)
	s.NotNil(turn.FinishedAt)
}

func (s *AgentSessionServiceSuite) TestWorkerTerminalizesRunningRedeliveryWithoutInvokingAgent() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionTestHarness()
	session := s.createAgentSession(ctx, h.tdb, testAgentInput{})

	_ = s.createAgentTurn(ctx, h.tdb, session, agentTurnSeed{
		riverJobID:   601,
		status:       at.StatusCompleted,
		finishReason: string(aix.AgentFinishReasonStop),
	})

	turn := s.createAgentTurn(ctx, h.tdb, session, agentTurnSeed{
		riverJobID: 602,
		status:     at.StatusRunning,
		startedAt:  new(time.Now().UTC().Add(-time.Minute)),
	})

	workErr := h.turnWorker.Work(ctx, makeAgentTurnJob(turn, 2))
	s.Require().NoError(workErr)

	var turnErr error
	turn, turnErr = h.tdb.Client(ctx).AgentTurn.Get(ctx, turn.ID)
	s.Require().NoError(turnErr)
	s.Equal(at.StatusFailed, turn.Status)
	s.NotNil(turn.Error)
	s.Equal(string(aix.AgentFinishReasonFailed), turn.FinishReason)
	s.NotNil(turn.FinishedAt)
	s.Empty(h.ai.Calls)
	s.Require().Len(h.msgs.Calls, 1)
	updated, isTurnUpdate := h.msgs.Calls[0].Arguments.Get(1).(rezai.AgentTurnUpdated)
	s.True(isTurnUpdate)
	s.Equal(at.StatusFailed, updated.Status)
	s.Equal(string(aix.AgentFinishReasonFailed), updated.FinishReason)
}

func (s *AgentSessionServiceSuite) TestAbortAgentTurnIsIdempotent() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionTestHarness()
	session := s.createAgentSession(ctx, h.tdb, testAgentInput{})
	turn := s.createAgentTurn(ctx, h.tdb, session, agentTurnSeed{
		riverJobID: 701,
		status:     at.StatusQueued,
	})

	h.jobs.EXPECT().
		Cancel(mock.Anything, int64(701)).
		Return(nil).
		Once()

	aborted, abortErr := h.service.AbortAgentTurn(ctx, turn.ID)
	s.Require().NoError(abortErr)
	s.Equal(at.StatusAborted, aborted.Status)
	s.Equal(string(aix.AgentFinishReasonAborted), aborted.FinishReason)
	s.Nil(aborted.Error)
	s.NotNil(aborted.FinishedAt)

	abortedAgain, abortAgainErr := h.service.AbortAgentTurn(ctx, turn.ID)
	s.Require().NoError(abortAgainErr)
	s.Equal(aborted.ID, abortedAgain.ID)
	s.Equal(at.StatusAborted, abortedAgain.Status)
}

func (s *AgentSessionServiceSuite) TestRetryAgentTurnRequeuesSameTurnAndClearsTerminalState() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionTestHarness()
	session := s.createAgentSession(ctx, h.tdb, testAgentInput{})

	root := s.createAgentTurn(ctx, h.tdb, session, agentTurnSeed{
		riverJobID:   801,
		status:       at.StatusCompleted,
		finishReason: string(aix.AgentFinishReasonStop),
	})
	s.Require().NotNil(root)

	inputMsg := ai.NewUserTextMessage("retry me")

	turn := s.createAgentTurn(ctx, h.tdb, session, agentTurnSeed{
		riverJobID:   802,
		input:        &rez.AiAgentTurnInput{Message: inputMsg},
		status:       at.StatusFailed,
		error:        fmt.Errorf("failed"),
		finishReason: string(aix.AgentFinishReasonFailed),
		startedAt:    new(time.Now().UTC().Add(-2 * time.Minute)),
		finishedAt:   new(time.Now().UTC().Add(-time.Minute)),
	})

	args := jobs.InvokeAgentTurn{
		AgentTurnID:    turn.ID,
		AgentSessionID: session.ID,
	}
	h.jobs.EXPECT().
		Insert(mock.Anything, args, mock.Anything).
		Return(makeJobInsertResult(803), nil).
		Once()

	msgsBefore, msgsBeforeErr := turn.QueryMessages().Count(ctx)
	s.Require().NoError(msgsBeforeErr)

	retried, retryErr := h.service.RetryAgentTurn(ctx, turn.ID)
	s.Require().NoError(retryErr)
	s.Equal(turn.ID, retried.ID)
	s.Equal(at.StatusQueued, retried.Status)
	s.Equal(int64(803), retried.RiverJobID)
	s.Nil(retried.Error)
	s.Nil(retried.StartedAt)
	s.Nil(retried.FinishedAt)
	s.Empty(retried.FinishReason)

	msgsAfter, msgsAfterErr := retried.QueryMessages().Count(ctx)
	s.Require().NoError(msgsAfterErr)
	s.Require().Equal(msgsBefore, msgsAfter)
}
