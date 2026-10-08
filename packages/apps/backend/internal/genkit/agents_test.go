package genkit

import (
	"context"
	"encoding/json"
	"errors"
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
	"github.com/rezible/rezible/pkg/errs"
)

func (s *AiRuntimeSuite) makeAgentSession(ctx context.Context, svc *AiRuntime, tdb rez.Database, name string, sessInput rez.ValidatingInput) *ent.AgentSession {
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
	session, txErr := ent.WithTxReturning(ctx, tdb, txFn)
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
	ctx, tdb := s.SetupTestDatabase()
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
	sess := s.makeAgentSession(ctx, svc, tdb, agent.def.Name, testAgentInput{})

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

// committedStepsModel calls check_service once, then fails or ends the next generation through failSecond, then
// answers from the tool response.
type committedStepsModel struct {
	mu          sync.Mutex
	input       string
	modelCalls  int
	toolCalls   int
	failSecond  func(context.Context) error
	checkTool   ai.Tool
	finalAnswer string
}

func newCommittedStepsModel(input string, failSecond func(context.Context) error) *committedStepsModel {
	m := &committedStepsModel{input: input, failSecond: failSecond, finalAnswer: "The service is healthy"}
	m.checkTool = ai.NewTool[struct{}, string](
		"check_service",
		"Check the service",
		func(*ai.ToolContext, struct{}) (string, error) {
			m.mu.Lock()
			defer m.mu.Unlock()
			m.toolCalls++
			return "healthy", nil
		},
	)
	return m
}

func (m *committedStepsModel) generate(ctx context.Context, req *ai.ModelRequest, _ any, _ ai.ModelStreamCallback) (*ai.ModelResponse, error) {
	m.mu.Lock()
	m.modelCalls++
	step := m.modelCalls
	m.mu.Unlock()

	switch step {
	case 1:
		call := &ai.ToolRequest{Name: m.checkTool.Name(), Ref: "check-1", Input: map[string]any{}}
		return &ai.ModelResponse{Message: ai.NewModelMessage(ai.NewToolRequestPart(call))}, nil
	case 2:
		return nil, m.failSecond(ctx)
	case 3:
		invCtx := getAgentInvocationContext(ctx)
		if invCtx == nil || invCtx.Input == nil || invCtx.Input.Message == nil || invCtx.Input.Message.Text() != m.input {
			return nil, fmt.Errorf("resumed invocation lost its original assignment")
		}
		inputs, responses := 0, 0
		for _, message := range req.Messages {
			if message.Role == ai.RoleUser && message.Text() == m.input {
				inputs++
			}
			for _, part := range message.Content {
				if part.ToolResponse != nil && part.ToolResponse.Ref == "check-1" {
					responses++
				}
			}
		}
		if inputs != 1 || responses != 1 {
			return nil, fmt.Errorf("committed history: inputs=%d responses=%d", inputs, responses)
		}
		return &ai.ModelResponse{
			Message:      ai.NewModelTextMessage(m.finalAnswer),
			FinishReason: ai.FinishReasonStop,
		}, nil
	}
	return nil, fmt.Errorf("unexpected model call %d", step)
}

// runFailAndResume fails a turn after its tool step, then resumes from the state the failure returned.
func (s *AiRuntimeSuite) runFailAndResume(failSecond func(context.Context) error, firstCtx func(context.Context) context.Context) {
	ctx, tdb := s.SetupTestDatabase()
	msg := ai.NewUserTextMessage("Check the service")
	agent := makeTestAgent[testAgentState](msg)
	script := newCommittedStepsModel(msg.Text(), failSecond)

	model := makeTestOutputModel(nil)
	model.opts.Supports.Tools = true
	model.fn = script.generate
	agent.def.Model = model.Name
	middleware := func(string) ai.Middleware {
		return &testToolsMiddleware{tools: []ai.Tool{script.checkTool}}
	}
	svc := s.makeRuntime(ctx, WithDefinedModel(model), WithAgent(agent, middleware))
	sess := s.makeAgentSession(ctx, svc, tdb, agent.def.Name, testAgentInput{})

	first := s.makeInvokeAgentSessionParams(sess)
	first.Input = &rez.AiAgentTurnInput{Message: msg}
	failed, failedErr := svc.InvokeAgentTurn(firstCtx(ctx), first)
	s.Require().NoError(failedErr)
	s.Require().NotNil(failed)
	s.Equal(aix.AgentFinishReasonFailed, failed.FinishReason)
	s.Require().Error(failed.Error)
	s.Require().Len(failed.State.Messages, 3, "the failure returns the committed tool step")
	s.Require().NotNil(failed.State.Messages[2].Content[0].ToolResponse)

	resume := s.makeInvokeAgentSessionParams(sess)
	resume.Input = &rez.AiAgentTurnInput{Message: msg}
	resume.ContinueFromState = true
	resume.State = failed.State
	result, resumeErr := svc.InvokeAgentTurn(ctx, resume)
	s.Require().NoError(resumeErr)
	s.Require().NotNil(result)
	s.Require().NoError(result.Error)
	s.Equal(aix.AgentFinishReasonStop, result.FinishReason)
	s.Require().Len(result.State.Messages, 4)
	s.Equal(script.finalAnswer, result.State.Messages[3].Text())

	script.mu.Lock()
	defer script.mu.Unlock()
	s.Equal(3, script.modelCalls)
	s.Equal(1, script.toolCalls, "the committed tool call is not repeated")
}

func (s *AiRuntimeSuite) TestContinuationRequiresExplicitIntentAndState() {
	agent := &wrappedAgent[testAgentInput, testAgentState]{}
	input := &rez.AiAgentTurnInput{Message: ai.NewUserTextMessage("Check the service")}
	state := &rez.AiAgentTurnState{Messages: []*ai.Message{input.Message}}

	_, missingStateErr := agent.normalizeTurnInput(input, &rez.AiAgentTurnState{}, true)
	s.ErrorIs(missingStateErr, errs.ErrInvalidInput)

	_, emptyInputErr := agent.normalizeTurnInput(&rez.AiAgentTurnInput{}, state, false)
	s.ErrorIs(emptyInputErr, errs.ErrInvalidInput)

	normalized, inputErr := agent.normalizeTurnInput(input, state, false)
	s.Require().NoError(inputErr)
	s.Same(input.Message, normalized.Message, "history alone must not suppress a new input")
}

func (s *AiRuntimeSuite) TestResumeAfterModelErrorDoesNotRepeatCommittedSteps() {
	failSecond := func(context.Context) error { return errors.New("model unavailable") }
	s.runFailAndResume(failSecond, func(ctx context.Context) context.Context { return ctx })
}

func (s *AiRuntimeSuite) TestResumeAfterCancelledContextDoesNotRepeatCommittedSteps() {
	var cancel context.CancelFunc
	firstCtx := func(ctx context.Context) context.Context {
		var cancelCtx context.Context
		cancelCtx, cancel = context.WithCancel(ctx)
		return cancelCtx
	}
	// The job's context ends while the model is generating after the tool step.
	failSecond := func(ctx context.Context) error {
		cancel()
		<-ctx.Done()
		return ctx.Err()
	}
	s.runFailAndResume(failSecond, firstCtx)
}

func (s *AiRuntimeSuite) TestToolIterationLimitFailsWithMaxTurnsError() {
	ctx, tdb := s.SetupTestDatabase()
	msg := ai.NewUserTextMessage("Keep checking")
	agent := makeTestAgent[testAgentState](msg)
	agent.def.MaxToolIterations = 2
	checkTool := ai.NewTool[struct{}, string](
		"check_service",
		"Check the service",
		func(*ai.ToolContext, struct{}) (string, error) {
			return "unknown", nil
		},
	)

	var mu sync.Mutex
	step := 0
	model := makeTestOutputModel(nil)
	model.opts.Supports.Tools = true
	model.fn = func(context.Context, *ai.ModelRequest, any, ai.ModelStreamCallback) (*ai.ModelResponse, error) {
		mu.Lock()
		defer mu.Unlock()
		step++
		call := &ai.ToolRequest{Name: checkTool.Name(), Ref: fmt.Sprintf("check-%d", step), Input: map[string]any{}}
		return &ai.ModelResponse{Message: ai.NewModelMessage(ai.NewToolRequestPart(call))}, nil
	}
	agent.def.Model = model.Name

	middleware := func(string) ai.Middleware {
		return &testToolsMiddleware{tools: []ai.Tool{checkTool}}
	}
	svc := s.makeRuntime(ctx, WithDefinedModel(model), WithAgent(agent, middleware))
	sess := s.makeAgentSession(ctx, svc, tdb, agent.def.Name, testAgentInput{})

	params := s.makeInvokeAgentSessionParams(sess)
	params.Input = &rez.AiAgentTurnInput{Message: msg}
	result, invokeErr := svc.InvokeAgentTurn(ctx, params)
	s.Require().NoError(invokeErr)
	s.Require().NotNil(result)
	s.Equal(aix.AgentFinishReasonFailed, result.FinishReason)
	s.Require().Error(result.Error)
	s.ErrorIs(result.Error, ai.ErrMaxTurnsExceeded)
	s.NotEmpty(result.State.Messages, "committed steps are the resume point")
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
