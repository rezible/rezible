package genkit

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/agentmessage"
	"github.com/rezible/rezible/ent/agentturn"
	rezai "github.com/rezible/rezible/pkg/ai"
)

func (s *AiRuntimeSuite) makeAgentSession(svc *AiRuntime, tdb rez.Database, name string, sessInput rez.ValidatingInput) *ent.AgentSession {
	sessInputJson, sessInputJsonErr := json.Marshal(sessInput)
	s.Require().NoError(sessInputJsonErr)

	txFn := func(ctx context.Context, tx *ent.Client) (*ent.AgentSession, error) {
		createSess := tx.AgentSession.Create().
			SetAgentName(name).
			SetInput(sessInputJson)
		createdSession, saveSessionErr := createSess.Save(ctx)
		if saveSessionErr != nil {
			return nil, fmt.Errorf("create session: %w", saveSessionErr)
		}
		session := createdSession

		turnInput, inputErr := svc.catalogue.MakeInitialTurnInput(ctx, createdSession)
		if inputErr != nil {
			return nil, fmt.Errorf("initial agent turn input: %w", inputErr)
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
			return nil, fmt.Errorf("create turn: %w", saveTurnErr)
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
				return nil, fmt.Errorf("create msg: %w", saveMsgErr)
			}
			updateCreatedTurn := createdTurn.Update().SetInputMessageID(createdMsg.ID)
			createdTurn = updateCreatedTurn.SaveX(ctx)

			session.Edges.Messages = append(session.Edges.Messages, createdMsg)
		}

		session.Edges.Turns = append(session.Edges.Turns, createdTurn)

		return session, nil
	}
	session, txErr := ent.WithTxReturning(s.SeedTenantContext(), tdb, txFn)
	s.Require().NoError(txErr)
	return session
}

func withTestAgent[S any](agent *testAgent[S]) []AiRuntimeOption {
	resp := &ai.ModelResponse{Message: ai.NewModelTextMessage("foo")}
	model := makeTestOutputModel(resp)
	agent.def.Model = model.Name
	return []AiRuntimeOption{WithDefinedModel(model), WithAgent(agent)}
}

func (s *AiRuntimeSuite) makeInvokeAgentSessionParams(sess *ent.AgentSession) rez.InvokeAiAgentTurnParams {
	s.Require().Greater(len(sess.Edges.Turns), 0)
	return rez.InvokeAiAgentTurnParams{Session: sess, Turn: sess.Edges.Turns[0]}
}

func (s *AiRuntimeSuite) TestClientManagedTurnStateAndResumeRoundTrip() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	msg := ai.NewUserTextMessage("Confirm the next check")
	agent := makeTestAgent[testAgentState](msg)
	interruptTool := ai.NewTool[struct{}, string](
		"confirm_check",
		"Ask for confirmation",
		func(tc *ai.ToolContext, _ struct{}) (string, error) {
			return "", tc.Interrupt(&ai.InterruptOptions{})
		},
	)

	var mu sync.Mutex
	step := 0
	model := makeTestOutputModel(nil)
	model.opts.Supports.Tools = true
	model.fn = func(
		ctx context.Context,
		req *ai.ModelRequest,
		_ any,
		_ ai.ModelStreamCallback,
	) (*ai.ModelResponse, error) {
		mu.Lock()
		defer mu.Unlock()
		step++
		if step == 1 {
			call := &ai.ToolRequest{
				Name:  interruptTool.Name(),
				Ref:   "confirm-1",
				Input: map[string]any{},
			}
			return &ai.ModelResponse{
				Message: ai.NewModelMessage(ai.NewToolRequestPart(call)),
			}, nil
		}
		if step != 2 {
			return nil, fmt.Errorf("unexpected model step %d", step)
		}
		inputCount := 0
		requestCount := 0
		responseCount := 0
		for _, message := range req.Messages {
			if message.Role == ai.RoleUser && message.Text() == msg.Text() {
				inputCount++
			}
			for _, part := range message.Content {
				call := part.ToolRequest
				if call != nil && call.Name == interruptTool.Name() && call.Ref == "confirm-1" {
					requestCount++
				}
				response := part.ToolResponse
				if response == nil || response.Name != interruptTool.Name() || response.Ref != "confirm-1" {
					continue
				}
				if response.Output == "approved" {
					responseCount++
				}
			}
		}
		if inputCount != 1 || requestCount != 1 || responseCount != 1 {
			return nil, fmt.Errorf(
				"resume history: inputs=%d requests=%d matching responses=%d",
				inputCount, requestCount, responseCount,
			)
		}
		return &ai.ModelResponse{
			Message:      ai.NewModelTextMessage("Check approved"),
			FinishReason: ai.FinishReasonStop,
		}, nil
	}

	middleware := func(string) ai.Middleware {
		return &testToolsMiddleware{tools: []ai.Tool{interruptTool}}
	}
	svc := s.makeRuntime(ctx, WithDefinedModel(model), WithAgent(agent, middleware))
	sess := s.makeAgentSession(svc, tdb, agent.def.Name, testAgentInput{})

	// The first invocation must stop at the confirmation tool.
	initial := s.makeInvokeAgentSessionParams(sess)
	initial.Input = &rez.AiAgentTurnInput{Message: msg}
	first, firstErr := svc.InvokeAgentTurn(ctx, initial)
	s.Require().NoError(firstErr)
	s.Require().NotNil(first)
	s.Require().NoError(first.Error)
	s.Require().Equal(aix.AgentFinishReasonInterrupted, first.FinishReason)

	// Resume that exact tool call with an approval response.
	next := s.makeInvokeAgentSessionParams(sess)
	next.Turn = &ent.AgentTurn{
		ID:             uuid.New(),
		AgentSessionID: sess.ID,
	}
	next.State = first.State
	response := &ai.ToolResponse{
		Name:   interruptTool.Name(),
		Ref:    "confirm-1",
		Output: "approved",
	}
	next.Input = &rez.AiAgentTurnInput{
		Resume: &aix.ToolResume{
			Respond: []*ai.Part{ai.NewToolResponsePart(response)},
		},
	}
	result, resumeErr := svc.InvokeAgentTurn(ctx, next)
	s.Require().NoError(resumeErr)
	s.Require().NotNil(result)
	s.Require().NoError(result.Error)
	s.Equal(aix.AgentFinishReasonStop, result.FinishReason)
	s.Require().NotNil(result.Response)
	s.Equal("Check approved", result.Response.Text())

	mu.Lock()
	s.Equal(2, step)
	mu.Unlock()
	inputCount := 0
	for _, message := range result.State.Messages {
		if message.Role == ai.RoleUser && message.Text() == msg.Text() {
			inputCount++
		}
	}
	s.Equal(1, inputCount)
}

func makeTestAgent[S any](msg *ai.Message) *testAgent[S] {
	taDef := testAgentDef{
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

	testAgentDef = rezai.AgentDefinition[testAgentInput]

	testAgent[S any] struct {
		def         testAgentDef
		customFn    func(S) S
		userMessage *ai.Message
	}
)

var _ agentHarness[testAgentInput, any] = &testAgent[any]{}

func (i testAgentInput) Validate() error {
	return nil
}

func (t *testAgent[S]) agentDefinition() testAgentDef {
	return t.def
}

func (t *testAgent[S]) makeInitialTurnInput(ctx context.Context, input testAgentInput) (*rez.AiAgentTurnInput, error) {
	return &rez.AiAgentTurnInput{Message: ai.NewUserTextMessage(t.userMessage.Text())}, nil
}

func (t *testAgent[S]) getCustomState(ctx context.Context, sess *ent.AgentSession) (*S, error) {
	return new(S), nil
}

func (t *testAgent[S]) transformState(ctx context.Context, state *aix.SessionState[S]) (*aix.SessionState[S], error) {
	return state, nil
}

func (t *testAgent[S]) transformStreamChunk(ctx context.Context, chunk *aix.AgentStreamChunk) (*aix.AgentStreamChunk, error) {
	return chunk, nil
}

// Keep tools private: Genkit serializes the prompt's middleware configuration.
type testToolsMiddleware struct {
	tools []ai.Tool
}

func (*testToolsMiddleware) Name() string {
	return "test_tools"
}

func (m *testToolsMiddleware) New(context.Context) (*ai.Hooks, error) {
	return &ai.Hooks{Tools: m.tools}, nil
}
