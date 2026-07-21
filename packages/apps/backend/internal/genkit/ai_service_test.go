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
	"github.com/rezible/rezible/ent/agentturn"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/internal/db"
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
	kg, kgErr := db.NewKnowledgeGraphService(s.Database())
	s.Require().NoError(kgErr)

	svc := NewAiService(s.Config(), kg)
	s.Require().NoError(svc.Init(s.T().Context(), opts...))

	return svc
}

func (s *AiServiceSuite) makeAgentSession(name string, input rezai.AgentInput) (*ent.AgentSession, *ent.AgentTurn) {
	initialInput, inputErr := json.Marshal(input)
	s.Require().NoError(inputErr)

	var session *ent.AgentSession
	var initialTurn *ent.AgentTurn
	txFn := func(ctx context.Context, tx *ent.Client) error {
		createSess := tx.AgentSession.Create().
			SetAgentName(name).
			SetOwnerUserID(s.SeedUser.ID)
		createdSession, saveSessionErr := createSess.Save(ctx)
		if saveSessionErr != nil {
			return fmt.Errorf("create session: %w", saveSessionErr)
		}

		createTurn := tx.AgentTurn.Create().
			SetID(uuid.New()).
			SetAgentSession(createdSession).
			SetRiverJobID(1). // synthetic test handle
			SetStatus(agentturn.StatusQueued).
			SetInput(initialInput)
		createdTurn, saveTurnErr := createTurn.Save(ctx)
		if saveTurnErr != nil {
			return fmt.Errorf("create turn: %w", saveTurnErr)
		}

		session = createdSession.Unwrap()
		initialTurn = createdTurn.Unwrap()

		return nil
	}
	s.Require().NoError(s.Database().WithTx(s.SeedTenantContext(), txFn))
	return session, initialTurn
}

type (
	testAgentInput struct{}
	testAgentState struct {
		Foo string `json:"foo"`
	}
	testAgentOutput struct {
		Greeting string `json:"greeting"`
	}
)

func (i testAgentInput) Validate() error {
	return nil
}

func (o testAgentOutput) Validate() error {
	if len(o.Greeting) == 0 {
		return fmt.Errorf("empty greeting")
	}
	return nil
}

func makeTestAgent[S rezai.SessionState](userMessage string) *testAgent[S] {
	taDef := testAgentDef[S]{
		Name:         "test_agent",
		Description:  "A simple agent",
		SystemPrompt: "You are an ai agent that follow user instructions exactly. Keep output concise",
	}
	return &testAgent[S]{
		def:         taDef,
		userMessage: userMessage,
	}
}

func (s *AiServiceSuite) TestClientManagedTurnStateAndResumeRoundTrip() {
	ctx := s.SeedTenantContext()
	ta := makeTestAgent[testAgentState]("initial")
	ta.customFn = func(state testAgentState) testAgentState {
		state.Foo += "-updated"
		return state
	}
	svc := s.makeService(WithAgent(ta))

	sess, initialTurn := s.makeAgentSession(ta.def.Name, testAgentInput{})

	initialInput := &rez.AgentTurnInput{Message: ai.NewUserTextMessage("root")}
	initialRes, initialErr := svc.InvokeAgentTurn(ctx, sess, initialTurn, nil, initialInput)
	s.Require().NoError(initialErr)
	s.Require().NotNil(initialRes)
	s.Equal(sess.ID.String(), ta.receivedSessionID)
	s.False(ta.receivedDetach)
	s.Empty(ta.receivedSnapshotID)
	var rootState aix.SessionState[testAgentState]
	s.Require().NoError(json.Unmarshal(initialRes.State, &rootState))
	s.Equal(sess.ID.String(), rootState.SessionID)
	s.Equal("-updated", rootState.Custom.Foo)

	resumePart := ai.NewTextPart("resume")
	resumeInput := &rez.AgentTurnInput{Resume: &ai.GenerateActionResume{
		Respond: []*ai.Part{resumePart},
	}}
	next, nextErr := svc.InvokeAgentTurn(ctx, sess, initialTurn, initialRes.State, resumeInput)
	s.Require().NoError(nextErr)
	s.Require().NotNil(next)
	s.Equal("-updated", ta.receivedCustom.Foo)
	s.False(ta.receivedDetach)
	s.Require().NotNil(ta.receivedResume)
	s.Require().Len(ta.receivedResume.Respond, 1)
	s.Equal(resumePart.Text, ta.receivedResume.Respond[0].Text)
	s.Empty(ta.receivedSnapshotID)
	var nextState aix.SessionState[testAgentState]
	s.Require().NoError(json.Unmarshal(next.State, &nextState))
	s.Equal(sess.ID.String(), nextState.SessionID)
	s.Equal("-updated-updated", nextState.Custom.Foo)
}

func (s *AiServiceSuite) TestSimpleGreetingAgent() {
	s.checkSkip("simple_greeting")

	s.SeedTestEntities()

	ctx := s.SeedTenantContext()

	ta := makeTestAgent[testAgentState]("Write a one-word greeting to result, then reply with a simple 'done'.")
	reg := s.makeService(WithAgent(ta))

	session, initialTurn := s.makeAgentSession(ta.def.Name, testAgentInput{})
	s.T().Logf("Starting test agent session (id %s)", session.ID)

	initialInput := &rez.AgentTurnInput{Message: ai.NewUserTextMessage(ta.userMessage)}
	result, startErr := reg.InvokeAgentTurn(ctx, session, initialTurn, nil, initialInput)

	s.Require().NoError(startErr)
	s.Require().NotNil(result)
	s.Require().NotEmpty(result.State)
	var state aix.SessionState[testAgentState]
	s.Require().NoError(json.Unmarshal(result.State, &state))
	s.Equal(session.ID.String(), state.SessionID)
	s.Require().NotEmpty(state.Messages)
}

type (
	testAgentDef[S rezai.SessionState] = rezai.AgentDefinition[testAgentInput, S, testAgentOutput]

	testAgent[S rezai.SessionState] struct {
		def                testAgentDef[S]
		customFn           func(S) S
		userMessage        string
		receivedCustom     S
		receivedDetach     bool
		receivedResume     *aix.ToolResume
		receivedSessionID  string
		receivedSnapshotID string
	}
)

func (t *testAgent[S]) agentDefinition() testAgentDef[S] {
	return t.def
}

func (t *testAgent[S]) makeInitialTurnInput(ctx context.Context, input testAgentInput) (*rez.AgentTurnInput, error) {
	return &rez.AgentTurnInput{Message: ai.NewUserTextMessage(t.userMessage)}, nil
}

func (t *testAgent[S]) makeAgentFunc([]ai.Middleware, []ai.ToolRef) aix.AgentFunc[S] {
	return func(ctx context.Context, _ aix.Responder, session *aix.SessionRunner[S]) (*aix.AgentResult, error) {
		t.receivedSessionID = session.SessionID()
		turnErr := session.Run(ctx, func(ctx context.Context, input *aix.AgentInput) (*aix.TurnResult, error) {
			t.receivedCustom = session.Custom()
			t.receivedDetach = input.Detach
			t.receivedResume = input.Resume
			if turnCtx := aix.TurnContextFromContext(ctx); turnCtx != nil {
				t.receivedSnapshotID = turnCtx.SnapshotID
			}
			if t.customFn != nil {
				session.UpdateCustom(t.customFn)
			}
			session.AddMessages(ai.NewModelTextMessage("done"))
			return &aix.TurnResult{FinishReason: aix.AgentFinishReasonStop}, nil
		})
		return session.Result(), turnErr
	}
}

func (t *testAgent[S]) transformState(ctx context.Context, state *aix.SessionState[S]) (*aix.SessionState[S], error) {
	return state, nil
}

func (t *testAgent[S]) transformStreamChunk(ctx context.Context, chunk *aix.AgentStreamChunk) (*aix.AgentStreamChunk, error) {
	return chunk, nil
}

//func (t *testAgent[S]) run(ctx context.Context, resp aix.Responder, sess *aix.SessionRunner[S]) (*aix.AgentResult, error) {
//	if t.customFn != nil {
//		sess.UpdateCustom(t.customFn)
//	}
//
//	runSessTurnFn := func(ctx context.Context, input *aix.AgentInput) (*aix.TurnResult, error) {
//		var finishReason aix.AgentFinishReason
//		var msg *ai.Message
//
//		if t.fakeCall {
//			finishReason = aix.AgentFinishReasonStop
//			msg = ai.NewModelTextMessage("response")
//		} else {
//			gen, genErr := genkit.Generate(ctx, genkit.FromContext(ctx),
//				ai.WithModelName("googleai/gemini-flash-latest"),
//				ai.WithSystem("You are a concise assistant."),
//				ai.WithMessages(sess.Messages()...),
//			)
//			if genErr != nil {
//				return nil, fmt.Errorf("generate err: %w", genErr)
//			}
//			msg = gen.Message
//			finishReason = aix.AgentFinishReason(gen.FinishReason)
//		}
//		sess.AddMessages(msg)
//
//		return &aix.TurnResult{FinishReason: finishReason}, nil
//	}
//
//	if turnErr := sess.Run(ctx, runSessTurnFn); turnErr != nil {
//		return nil, fmt.Errorf("run turn: %w", turnErr)
//	}
//
//	return sess.Result(), nil
//}
