package genkit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/agentmessage"
	"github.com/rezible/rezible/ent/agentturn"
	"github.com/rezible/rezible/test/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/test"
)

type AiServiceSuite struct {
	test.Suite
}

func TestAiServiceSuite(t *testing.T) {
	suite.Run(t, &AiServiceSuite{Suite: test.NewSuite()})
}

func (s *AiServiceSuite) checkSkip(name string) {
	if os.Getenv("AI_TESTS_ALL") != "true" && os.Getenv("AI_TESTS_"+name) != "true" {
		s.T().Skipf("Skipping live AI test '%s'", name)
	}
}

func (s *AiServiceSuite) makeService(opts ...AiServiceOption) *AiService {
	svc := NewAiService(s.Config())
	s.Require().NoError(svc.Init(s.T().Context(), opts...))

	return svc
}

func (s *AiServiceSuite) makeAgentSession(svc *AiService, name string, sessInput rezai.AgentInput) *ent.AgentSession {
	sessInputJson, sessInputJsonErr := json.Marshal(sessInput)
	s.Require().NoError(sessInputJsonErr)

	var session *ent.AgentSession
	txFn := func(ctx context.Context, tx *ent.Client) error {
		createSess := tx.AgentSession.Create().
			SetAgentName(name).
			SetInput(sessInputJson).
			SetOwnerUserID(s.SeedUser.ID)
		createdSession, saveSessionErr := createSess.Save(ctx)
		if saveSessionErr != nil {
			return fmt.Errorf("create session: %w", saveSessionErr)
		}
		session = createdSession.Unwrap()

		turnInput, inputErr := svc.MakeInitialAgentTurnInput(ctx, createdSession)
		if inputErr != nil {
			return fmt.Errorf("initial agent turn input: %w", inputErr)
		}

		createTurn := tx.AgentTurn.Create().
			SetID(uuid.New()).
			SetAgentSession(createdSession).
			SetSequence(1).
			SetRiverJobID(1). // synthetic test handle
			SetStatus(agentturn.StatusQueued).
			SetInputToolResume(turnInput.Resume)
		createdTurn, saveTurnErr := createTurn.Save(ctx)
		if saveTurnErr != nil {
			return fmt.Errorf("create turn: %w", saveTurnErr)
		}

		if turnInput.Message != nil {
			createMsg := tx.AgentMessage.Create().
				SetID(uuid.New()).
				SetAgentSessionID(createdSession.ID).
				SetAgentTurnID(createdTurn.ID).
				SetSequence(1).
				SetRole(agentmessage.RoleUser).
				SetContent(turnInput.Message.Content).
				SetMetadata(turnInput.Message.Metadata)
			createdMsg, saveMsgErr := createMsg.Save(ctx)
			if saveMsgErr != nil {
				return fmt.Errorf("create msg: %w", saveMsgErr)
			}
			createdTurn = createdTurn.Update().SetInputMessageID(createdMsg.ID).SaveX(ctx)

			session.Edges.Messages = append(session.Edges.Messages, createdMsg.Unwrap())
		}

		session.Edges.Turns = append(session.Edges.Turns, createdTurn.Unwrap())

		return nil
	}
	s.Require().NoError(s.Database().WithTx(s.SeedTenantContext(), txFn))
	return session
}

func (s *AiServiceSuite) makeInvokeSessionParams(sess *ent.AgentSession) rez.InvokeAgentTurnParams {
	s.Require().Greater(len(sess.Edges.Turns), 0)
	return rez.InvokeAgentTurnParams{
		Session: sess,
		Turn:    sess.Edges.Turns[0],
	}
}

func (s *AiServiceSuite) TestIntegrationToolsMiddlewareLoadsToolsPerTurn() {
	ta := makeTestAgent[testAgentState](ai.NewUserTextMessage("hello"))

	tool := ai.NewTool[any, map[string]any](
		"test_integration_lookup",
		"test integration lookup",
		func(ctx *ai.ToolContext, input any) (map[string]any, error) {
			return map[string]any{"ok": true}, nil
		},
	)

	intgs := mocks.NewMockIntegrationService(s.T())
	intgs.EXPECT().
		GetAvailableAgentTools(mock.Anything, rez.GetAvailableAgentToolsParams{AgentName: ta.def.Name}).
		Return([]ai.Tool{tool}, nil).
		Twice()

	ctx := s.T().Context()

	mw := newIntegrationToolsMiddleware(ta.def.Name, intgs)
	firstHooks, firstErr := mw.New(ctx)
	s.Require().NoError(firstErr, "first middleware init")
	s.Require().NotEmpty(firstHooks.Tools, "no tools supplied")
	s.Require().Equal(tool.Name(), firstHooks.Tools[0].Name(), "unexpected tool name")

	secondHooks, secondErr := mw.New(ctx)
	s.Require().NoError(secondErr, "second middleware init")
	s.Require().NotEmpty(secondHooks.Tools, "no tools supplied")
	s.Require().Equal(tool.Name(), secondHooks.Tools[0].Name(), "unexpected tool name")
}

func (s *AiServiceSuite) TestClientManagedTurnStateAndResumeRoundTrip() {
	ctx := s.SeedTenantContext()
	msg := ai.NewUserTextMessage("hello world")
	ta := makeTestAgent[testAgentState](msg)
	svc := s.makeService(WithAgent(ta))
	sess := s.makeAgentSession(svc, ta.def.Name, testAgentInput{})

	initParams := s.makeInvokeSessionParams(sess)
	initParams.Input = &rez.AiAgentTurnInput{Message: msg}

	initialRes, initialErr := svc.InvokeAgentTurn(ctx, initParams)
	s.Require().NoError(initialErr)
	s.Require().NotNil(initialRes)
	s.Require().NotEmpty(initialRes.State.Messages)
	s.Equal("hello world", initialRes.State.Messages[0].Text())

	resumeText := "resume"
	nextParams := s.makeInvokeSessionParams(sess)
	nextParams.Turn.ID = uuid.New()
	nextParams.State = initialRes.State
	nextParams.Input = &rez.AiAgentTurnInput{
		Resume: &aix.ToolResume{
			Respond: []*ai.Part{ai.NewTextPart(resumeText)},
		},
	}
	nextRes, nextErr := svc.InvokeAgentTurn(ctx, nextParams)
	s.Require().NoError(nextErr)
	s.Require().NotNil(nextRes)
	s.Require().GreaterOrEqual(len(nextRes.State.Messages), len(initialRes.State.Messages))
}

func (s *AiServiceSuite) TestSimpleGreetingAgent() {
	s.checkSkip("simple_greeting")

	ctx := s.SeedTenantContext()

	msg := ai.NewUserTextMessage("Reply with a one-word greeting.")
	ta := makeTestAgent[testAgentState](msg)
	svc := s.makeService(WithAgent(ta))

	sess := s.makeAgentSession(svc, ta.def.Name, testAgentInput{})
	s.T().Logf("Starting test agent session (id %s)", sess.ID)

	initParams := s.makeInvokeSessionParams(sess)
	initParams.Input = &rez.AiAgentTurnInput{Message: msg}

	result, invokeErr := svc.InvokeAgentTurn(ctx, initParams)
	s.Require().NoError(invokeErr)
	s.Require().NotNil(result)
	s.NotEmpty(result.State.Messages)
}

func makeTestAgent[S rezai.SessionState](msg *ai.Message) *testAgent[S] {
	taDef := testAgentDef[S]{
		Name:         "test_agent",
		Description:  "A simple agent",
		SystemPrompt: "You are an ai agent that follow user instructions exactly. Keep output concise",
	}
	return &testAgent[S]{def: taDef, userMessage: msg}
}

type (
	testAgentInput struct{}
	testAgentState struct {
		Foo string `json:"foo"`
	}

	testAgentDef[S rezai.SessionState] = rezai.AgentDefinition[testAgentInput, S]

	testAgent[S rezai.SessionState] struct {
		def         testAgentDef[S]
		customFn    func(S) S
		userMessage *ai.Message
	}
)

var _ agentRunner[testAgentInput, any] = &testAgent[any]{}

func (i testAgentInput) Validate() error {
	return nil
}

func (t *testAgent[S]) agentDefinition() testAgentDef[S] {
	return t.def
}

func (t *testAgent[S]) makeInitialTurnInput(ctx context.Context, input testAgentInput) (*rez.AiAgentTurnInput, error) {
	return &rez.AiAgentTurnInput{Message: ai.NewUserTextMessage(t.userMessage.Text())}, nil
}

func (t *testAgent[S]) getCustomState(ctx context.Context, sess *ent.AgentSession) (*S, error) {
	var s S
	return &s, nil
}

func (t *testAgent[S]) transformState(ctx context.Context, state *aix.SessionState[S]) (*aix.SessionState[S], error) {
	return state, nil
}

func (t *testAgent[S]) transformStreamChunk(ctx context.Context, chunk *aix.AgentStreamChunk) (*aix.AgentStreamChunk, error) {
	return chunk, nil
}
