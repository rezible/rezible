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

type AiAgentSessionServiceSuite struct {
	test.Suite
}

func TestAiAgentSessionServiceSuite(t *testing.T) {
	suite.Run(t, &AiAgentSessionServiceSuite{
		Suite: test.NewSuite(),
	})
}

type agentSessionFixture struct {
	tdb     rez.Database
	jobs    *mocks.MockJobService
	msgs    *mocks.MockMessageQueue
	service *AiAgentSessionService
}

func (s *AiAgentSessionServiceSuite) newAgentSessionFixture() *agentSessionFixture {
	tdb := s.CreateTestDatabase()
	jobService := mocks.NewMockJobService(s.T())
	messageService := mocks.NewMockMessageQueue(s.T())
	messageService.EXPECT().
		Publish(mock.Anything, mock.IsType(rezai.AgentTurnUpdated{})).
		Return(nil).
		Maybe()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := &AiAgentSessionService{
		logger: logger,
		db:     tdb,
		jobs:   jobService,
		msgs:   messageService,
	}
	return &agentSessionFixture{
		tdb:     tdb,
		jobs:    jobService,
		msgs:    messageService,
		service: service,
	}
}

func (s *AiAgentSessionServiceSuite) createAgentSession(ctx context.Context, tdb rez.Database, input rez.ValidatingInput) *ent.AgentSession {
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

func (s *AiAgentSessionServiceSuite) createAgentTurn(ctx context.Context, tdb rez.Database, session *ent.AgentSession, seed agentTurnSeed) *ent.AgentTurn {
	if seed.id == uuid.Nil {
		seed.id = uuid.New()
	}
	if seed.input == nil {
		seed.input = &rez.AiAgentTurnInput{
			Message: ai.NewUserTextMessage("test"),
		}
	}

	input, inputErr := normalizeAgentTurnInput(seed.input)
	s.Require().NoError(inputErr)

	turn, txErr := ent.WithTxReturning(ctx, tdb, func(ctx context.Context, client *ent.Client) (*ent.AgentTurn, error) {
		sess, sessErr := client.AgentSession.Get(ctx, session.ID)
		s.Require().NoError(sessErr)

		numTurns, numTurnsErr := sess.QueryTurns().Count(ctx)
		s.Require().NoError(numTurnsErr)
		numMsgs, numMsgsErr := sess.QueryMessages().Count(ctx)
		s.Require().NoError(numMsgsErr)

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
		createdTurn, createdTurnErr := createTurn.Save(ctx)
		s.Require().NoError(createdTurnErr)

		if inputMsg := seed.input.Message; inputMsg != nil {
			createMsg := client.AgentMessage.Create().
				SetAgentSessionID(session.ID).
				SetAgentTurnID(createdTurn.ID).
				SetSequence(numMsgs + 1).
				SetRole(agentmessage.Role(inputMsg.Role)).
				SetContent(inputMsg.Content).
				SetMetadata(inputMsg.Metadata)
			createdMsg, createdMsgErr := createMsg.Save(ctx)
			s.Require().NoError(createdMsgErr)

			updateCreatedTurn := createdTurn.Update().
				SetInputMessageID(createdMsg.ID)
			var updateInputMessageErr error
			createdTurn, updateInputMessageErr = updateCreatedTurn.Save(ctx)
			s.Require().NoError(updateInputMessageErr)

			createdTurn.Edges.Messages = append(createdTurn.Edges.Messages, createdMsg)
		}

		return createdTurn, nil
	})
	s.Require().NoError(txErr)

	return turn
}

func (s *AiAgentSessionServiceSuite) createAgentMessage(ctx context.Context, tdb rez.Database, session *ent.AgentSession, turn *ent.AgentTurn, msg *ai.Message) *ent.AgentMessage {
	numMsgs, countErr := tdb.Client(ctx).AgentSession.QueryMessages(session).Count(ctx)
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
	return &rivertype.JobInsertResult{
		Job: &rivertype.JobRow{
			ID: id,
		},
	}
}

type testAgentInput struct {
	Foo string
}

func (i testAgentInput) Validate() error {
	return nil
}

func (s *AiAgentSessionServiceSuite) TestCreateAgentSessionCreatesQueuedStartAtomically() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionFixture()

	h.jobs.EXPECT().
		Insert(mock.Anything, mock.IsType(jobs.StartAgentSession{}), mock.Anything).
		Return(makeJobInsertResult(101), nil).
		Once()

	params := rez.CreateAiAgentSessionParams{
		AgentName: "test-agent",
		Input: testAgentInput{
			Foo: "bar",
		},
		Metadata: map[string]any{"baz": "123"},
	}

	session, createErr := h.service.CreateAgentSession(ctx, params)
	s.Require().NoError(createErr)
	s.Require().NotNil(session)
	s.Equal("test-agent", session.AgentName)
	s.Equal("123", session.Metadata["baz"])
}

func (s *AiAgentSessionServiceSuite) TestCreateAgentSessionCreatesRequestedBindings() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionFixture()
	resourceRef := uuid.NewString()

	h.jobs.EXPECT().
		Insert(mock.Anything, mock.IsType(jobs.StartAgentSession{}), mock.Anything).
		Return(makeJobInsertResult(101), nil).
		Once()

	bindingRef := rez.ProviderResourceRef{
		Provider:    "rezible",
		ResourceRef: resourceRef,
	}
	bindingParams := rez.AiAgentSessionBindingParams{
		ProviderResourceRef: bindingRef,
		Metadata:            map[string]any{"origin": "test"},
	}
	params := rez.CreateAiAgentSessionParams{
		AgentName: "test-agent",
		Input: testAgentInput{
			Foo: "bar",
		},
		Bindings: []rez.AiAgentSessionBindingParams{bindingParams},
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

func (s *AiAgentSessionServiceSuite) TestCreateAgentSessionRejectsExternalBindingWithoutNamespace() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionFixture()
	params := rez.CreateAiAgentSessionParams{
		AgentName: "test-agent",
		Input: testAgentInput{
			Foo: "bar",
		},
		Bindings: []rez.AiAgentSessionBindingParams{{
			ProviderResourceRef: rez.ProviderResourceRef{
				Provider:    "slack",
				ResourceRef: "thread:C123:123.456",
			},
		}},
	}

	session, createErr := h.service.CreateAgentSession(ctx, params)
	s.Nil(session)
	s.ErrorIs(createErr, rez.ErrInvalidInput)
	agentSessionCount, agentSessionCountErr := h.tdb.Client(ctx).AgentSession.Query().Count(ctx)
	s.Require().NoError(agentSessionCountErr)

	s.Zero(agentSessionCount)
	agentSessionBindingCount, agentSessionBindingCountErr := h.tdb.Client(ctx).AgentSessionBinding.Query().Count(ctx)
	s.Require().NoError(agentSessionBindingCountErr)

	s.Zero(agentSessionBindingCount)
}

func (s *AiAgentSessionServiceSuite) TestCreateAgentSessionRejectsIntegrationProviderMismatchAtomically() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionFixture()
	createIntg := h.tdb.Client(ctx).Integration.Create().
		SetProvider("github").
		SetName("github").
		SetDisplayName("GitHub").
		SetProviderInstallationRef("org-1").
		SetInstallationConfig([]byte(`{}`))
	intg, intgErr := createIntg.Save(ctx)
	s.Require().NoError(intgErr)

	params := rez.CreateAiAgentSessionParams{
		AgentName: "test-agent",
		Input: testAgentInput{
			Foo: "bar",
		},
		Bindings: []rez.AiAgentSessionBindingParams{{
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
	agentSessionCount, agentSessionCountErr := h.tdb.Client(ctx).AgentSession.Query().Count(ctx)
	s.Require().NoError(agentSessionCountErr)

	s.Zero(agentSessionCount)
	agentSessionBindingCount, agentSessionBindingCountErr := h.tdb.Client(ctx).AgentSessionBinding.Query().Count(ctx)
	s.Require().NoError(agentSessionBindingCountErr)

	s.Zero(agentSessionBindingCount)
}

func (s *AiAgentSessionServiceSuite) TestSetAgentSessionBindingCreatesBindingAfterSession() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionFixture()
	session := s.createAgentSession(ctx, h.tdb, testAgentInput{
		Foo: "bar",
	})
	createIntg := h.tdb.Client(ctx).Integration.Create().
		SetProvider("slack").
		SetName("slack_agent").
		SetDisplayName("Slack Agent").
		SetProviderInstallationRef("team:T123").
		SetInstallationConfig([]byte(`{}`))
	intg, intgErr := createIntg.Save(ctx)
	s.Require().NoError(intgErr)

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

func (s *AiAgentSessionServiceSuite) TestSetAgentSessionBindingValidatesNewBinding() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionFixture()
	session := s.createAgentSession(ctx, h.tdb, testAgentInput{
		Foo: "bar",
	})

	binding, setErr := h.service.SetAgentSessionBinding(ctx, uuid.Nil, func(m *ent.AgentSessionBindingMutation) {
		m.SetAgentSessionID(session.ID)
		m.SetProvider("slack")
		m.SetProviderResourceRef("thread:C123:123.456")
	})
	s.Nil(binding)
	s.ErrorIs(setErr, rez.ErrInvalidInput)
	agentSessionBindingCount, agentSessionBindingCountErr := h.tdb.Client(ctx).AgentSessionBinding.Query().Count(ctx)
	s.Require().NoError(agentSessionBindingCountErr)

	s.Zero(agentSessionBindingCount)
}

func (s *AiAgentSessionServiceSuite) TestSetAgentSessionBindingKeepsNamespacesDistinct() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionFixture()
	firstSession := s.createAgentSession(ctx, h.tdb, testAgentInput{
		Foo: "first",
	})
	secondSession := s.createAgentSession(ctx, h.tdb, testAgentInput{
		Foo: "second",
	})
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
	agentSessionBindingCount, agentSessionBindingCountErr := h.tdb.Client(ctx).AgentSessionBinding.Query().Count(ctx)
	s.Require().NoError(agentSessionBindingCountErr)

	s.Equal(2, agentSessionBindingCount)
}

func (s *AiAgentSessionServiceSuite) TestCreateAgentSessionRollsBackWhenJobInsertFails() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionFixture()
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

	params := rez.CreateAiAgentSessionParams{
		AgentName: "test-agent",
		Input: testAgentInput{
			Foo: "bar",
		},
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

func (s *AiAgentSessionServiceSuite) TestRequestAgentTurnValidatesInputAndCompletedRoot() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionFixture()
	session := s.createAgentSession(ctx, h.tdb, testAgentInput{
		Foo: "bar",
	})

	initialTurn := s.createAgentTurn(ctx, h.tdb, session, agentTurnSeed{
		riverJobID: 201,
		status:     at.StatusQueued,
	})

	validParams := &rez.RequestAiAgentTurnParams{
		Input: &rez.AiAgentTurnInput{
			Message: ai.NewUserTextMessage("follow up"),
		},
	}

	turn, requestErr := h.service.RequestAgentTurn(ctx, session.ID, validParams)
	s.Nil(turn)
	s.ErrorIs(requestErr, rez.ErrConflict)

	setInitialCompleted := h.tdb.Client(ctx).AgentTurn.UpdateOneID(initialTurn.ID).
		SetStatus(at.StatusCompleted).
		SetFinishReason(string(aix.AgentFinishReasonStop)).
		SetFinishedAt(time.Now().UTC())
	s.Require().NoError(setInitialCompleted.Exec(ctx))

	turn, requestErr = h.service.RequestAgentTurn(ctx, session.ID, nil)
	s.Nil(turn)
	s.ErrorIs(requestErr, rez.ErrInvalidInput)

	emptyParams := &rez.RequestAiAgentTurnParams{
		Input: &rez.AiAgentTurnInput{
			Resume: &aix.ToolResume{},
		},
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

	inputMsg, inputMsgErr := h.tdb.Client(ctx).AgentTurn.QueryInputMessage(turn).Only(ctx)
	s.Require().NoError(inputMsgErr)
	s.Equal(agentmessage.RoleUser, inputMsg.Role)
	s.Equal("follow up", inputMsg.MakeGenkitMessage().Text())
}

func (s *AiAgentSessionServiceSuite) TestWorkerPersistsSuccessfulResultAndPublishesEvent() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionFixture()
	runtime := mocks.NewMockAiAgentRuntime(s.T())
	worker := &InvokeAgentTurnWorker{
		db:     h.tdb,
		agents: runtime,
		msgs:   h.msgs,
		logger: h.service.logger,
	}
	session := s.createAgentSession(ctx, h.tdb, testAgentInput{
		Foo: "bar",
	})

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
		input: &rez.AiAgentTurnInput{
			Message: followUpInput,
		},
	})
	s.Require().NotNil(followUp)

	followUpReply := ai.NewModelTextMessage("done")
	artifact := &aix.Artifact{
		Name:     "report",
		Parts:    []*ai.Part{ai.NewTextPart("new report")},
		Metadata: map[string]any{"version": "new"},
	}
	createAgentArtifact := h.tdb.Client(ctx).AgentArtifact.Create().
		SetAgentSessionID(session.ID).
		SetAgentTurnID(initialTurn.ID).
		SetName(artifact.Name).
		SetParts([]*ai.Part{ai.NewTextPart("old report")}).
		SetMetadata(map[string]any{"version": "old"})

	execAgentArtifactErr := createAgentArtifact.Exec(ctx)
	s.Require().NoError(execAgentArtifactErr)

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

	runtime.EXPECT().
		InvokeAgentTurn(mock.Anything, mock.Anything).
		Run(func(_ context.Context, params rez.InvokeAiAgentTurnParams) {
			s.Equal(session.ID, params.Session.ID, "invocation session")
			s.Equal(followUp.ID, params.Turn.ID, "invocation turn")
			s.Require().NotNil(params.Input)
			s.Require().NotNil(params.Input.Message)
			s.Equal("baz", params.Input.Message.Text(), "follow-up input")
			s.Require().Len(params.State.Messages, 2)
			s.Equal("test", params.State.Messages[0].Text(), "initial input history")
			s.Equal("ready", params.State.Messages[1].Text(), "initial reply history")
		}).
		Return(turnResult, nil).
		Once()

	h.msgs.EXPECT().
		Publish(mock.Anything, mock.IsType(&rezai.EventOnAgentTurnFinished{})).
		Run(func(_ context.Context, value any) {
			event := value.(*rezai.EventOnAgentTurnFinished)
			s.Equal(session.ID, event.AgentSessionId, "published session")
			s.Equal(followUp.ID, event.AgentTurnId, "published turn")
			s.Equal(aix.AgentFinishReasonStop, event.FinishReason)
			s.Equal(followUpReply, event.Response)
		}).
		Return(nil).
		Once()

	s.Require().NoError(worker.Work(ctx, makeAgentTurnJob(followUp, 1)))

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

	queryTurnMessages := h.tdb.Client(ctx).AgentTurn.QueryMessages(turn).Order(agentmessage.BySequence())
	turnMsgs, queryMsgsErr := queryTurnMessages.All(ctx)
	s.Require().NoError(queryMsgsErr)
	s.Require().Len(turnMsgs, 2)
	s.Equal("baz", turnMsgs[0].MakeGenkitMessage().Text())
	s.Equal("done", turnMsgs[1].MakeGenkitMessage().Text())
	s.Equal(3, turnMsgs[0].Sequence)
	s.Equal(4, turnMsgs[1].Sequence)

	updatedArtifact, artifactErr := h.tdb.Client(ctx).AgentSession.QueryArtifacts(session).Only(ctx)
	s.Require().NoError(artifactErr)
	s.Equal(followUp.ID, *updatedArtifact.LastAgentTurnID)
	s.Equal("new", updatedArtifact.Metadata["version"])
	s.Require().Len(updatedArtifact.Parts, 1)
	s.Equal("new report", updatedArtifact.Parts[0].Text)
}

func (s *AiAgentSessionServiceSuite) TestWorkerFailsTurnWhenAgentReturnsNilResult() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionFixture()
	runtime := mocks.NewMockAiAgentRuntime(s.T())
	worker := &InvokeAgentTurnWorker{
		db:     h.tdb,
		agents: runtime,
		msgs:   h.msgs,
		logger: h.service.logger,
	}
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

	runtime.EXPECT().
		InvokeAgentTurn(mock.Anything, mock.Anything).
		Return(nil, nil).
		Times(3)

	for attempt := 1; attempt <= 3; attempt++ {
		workErr := worker.Work(ctx, makeAgentTurnJob(turn, attempt))
		s.Require().ErrorContains(workErr, "agent returned no result")
		persisted, persistedErr := h.tdb.Client(ctx).AgentTurn.Get(ctx, turn.ID)
		s.Require().NoError(persistedErr)
		if attempt < 3 {
			s.Equal(at.StatusQueued, persisted.Status)
			s.Nil(persisted.FinishedAt)
		}
	}

	turn, turnErr := h.tdb.Client(ctx).AgentTurn.Get(ctx, turn.ID)
	s.Require().NoError(turnErr)
	s.Equal(at.StatusFailed, turn.Status)
	s.Require().NotNil(turn.Error)
	s.Contains(*turn.Error, "agent returned no result")
	s.Equal(string(aix.AgentFinishReasonFailed), turn.FinishReason)
	s.NotNil(turn.FinishedAt)
}

func (s *AiAgentSessionServiceSuite) TestWorkerDoesNotMutateConcurrentRunningDelivery() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionFixture()
	runtime := mocks.NewMockAiAgentRuntime(s.T())
	worker := &InvokeAgentTurnWorker{
		db:     h.tdb,
		agents: runtime,
		msgs:   h.msgs,
		logger: h.service.logger,
	}
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

	workErr := worker.Work(ctx, makeAgentTurnJob(turn, 2))
	s.Require().ErrorIs(workErr, errAgentTurnAlreadyRunning)

	var turnErr error
	turn, turnErr = h.tdb.Client(ctx).AgentTurn.Get(ctx, turn.ID)
	s.Require().NoError(turnErr)
	s.Equal(at.StatusRunning, turn.Status)
	s.Nil(turn.Error)
	s.Nil(turn.FinishedAt)
	s.Empty(runtime.Calls)
	s.Empty(h.msgs.Calls)
}

func (s *AiAgentSessionServiceSuite) TestAbortAgentTurnIsIdempotent() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionFixture()
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

func (s *AiAgentSessionServiceSuite) TestRetryAgentTurnRequeuesSameTurnAndClearsTerminalState() {
	ctx := s.SeedTenantContext()
	h := s.newAgentSessionFixture()
	session := s.createAgentSession(ctx, h.tdb, testAgentInput{})

	root := s.createAgentTurn(ctx, h.tdb, session, agentTurnSeed{
		riverJobID:   801,
		status:       at.StatusCompleted,
		finishReason: string(aix.AgentFinishReasonStop),
	})
	s.Require().NotNil(root)

	inputMsg := ai.NewUserTextMessage("retry me")

	turn := s.createAgentTurn(ctx, h.tdb, session, agentTurnSeed{
		riverJobID: 802,
		input: &rez.AiAgentTurnInput{
			Message: inputMsg,
		},
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

	msgsBefore, msgsBeforeErr := h.tdb.Client(ctx).AgentTurn.QueryMessages(turn).Count(ctx)
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

	msgsAfter, msgsAfterErr := h.tdb.Client(ctx).AgentTurn.QueryMessages(retried).Count(ctx)
	s.Require().NoError(msgsAfterErr)
	s.Require().Equal(msgsBefore, msgsAfter)
}
