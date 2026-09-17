package genkit

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	genkitx "github.com/firebase/genkit/go/genkit/exp"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/execution"
)

func WithAgent[I rez.ValidatingInput, S any](r agentRunner[I, S], mwFuncs ...AgentMiddlewareConstructorFn) AiRuntimeOption {
	return AiRuntimeOption{
		kind: AiRuntimeOptionKindAgent,
		optFn: func(s *AiRuntime) error {
			return s.catalogue.addAgent(s.gk, r, mwFuncs...)
		},
	}
}

type agentInvoker interface {
	Config() rez.AiAgentConfig
	DecodeSessionInput([]byte) (rez.ValidatingInput, error)
	MakeInitialTurnInput(context.Context, *ent.AgentSession) (*rez.AiAgentTurnInput, error)
	Invoke(context.Context, rez.InvokeAiAgentTurnParams) (*rez.AiAgentInvocationResult, error)
}

type agentCatalogue struct {
	agents map[string]agentInvoker
}

func newAgentCatalogue() *agentCatalogue {
	return &agentCatalogue{
		agents: make(map[string]agentInvoker),
	}
}

func (c *agentCatalogue) GetAgents() []rez.AiAgentConfig {
	var agents []rez.AiAgentConfig
	for _, agent := range c.agents {
		agents = append(agents, agent.Config())
	}
	return agents
}

func (c *agentCatalogue) getAgent(name string) (agentInvoker, error) {
	if a, ok := c.agents[name]; ok {
		return a, nil
	}
	return nil, fmt.Errorf("agent %q not found", name)
}

func (c *agentCatalogue) ValidateAgentSessionInput(name string, input []byte) (rez.ValidatingInput, error) {
	a, agentErr := c.getAgent(name)
	if agentErr != nil {
		return nil, agentErr
	}
	return a.DecodeSessionInput(input)
}

func (c *agentCatalogue) MakeInitialAgentTurnInput(ctx context.Context, sess *ent.AgentSession) (*rez.AiAgentTurnInput, error) {
	wrapper, wrapperErr := c.getAgent(sess.AgentName)
	if wrapperErr != nil {
		return nil, wrapperErr
	}
	return wrapper.MakeInitialTurnInput(ctx, sess)
}

type (
	agentRunner[SessionInput rez.ValidatingInput, State any] interface {
		agentDefinition() rezai.AiAgentDefinition[SessionInput]
		makeInitialTurnInput(context.Context, SessionInput) (*rez.AiAgentTurnInput, error)
		getCustomState(context.Context, *ent.AgentSession) (*State, error)
		transformState(context.Context, *aix.SessionState[State]) (*aix.SessionState[State], error)
		transformStreamChunk(context.Context, *aix.AgentStreamChunk) (*aix.AgentStreamChunk, error)
	}

	customAgentRunner[SessionInput rez.ValidatingInput, State any] interface {
		makeAgentFunc([]ai.Middleware) aix.AgentFunc[State]
	}

	systemPromptFuncAgentRunner[SessionInput rez.ValidatingInput] interface {
		makeSystemPrompt(context.Context, SessionInput) (string, error)
	}

	runnerWithInitialTurnMessage[SessionInput rez.ValidatingInput] interface {
		updateInitialTurnMessage(context.Context, SessionInput) (string, error)
	}

	runnerWithMiddleware interface {
		makeMiddleware() []ai.Middleware
	}
)

func (c *agentCatalogue) addAgent[SessionInput rez.ValidatingInput, State any](gk *genkit.Genkit, runner agentRunner[SessionInput, State], mwFuncs ...AgentMiddlewareConstructorFn) error {
	d := runner.agentDefinition()
	opts := []aix.AgentOption[State]{
		aix.WithDescription[State](d.Description),
		aix.WithStateTransform[State](runner.transformState),
		aix.WithStreamTransform[State](runner.transformStreamChunk),
	}

	var modelName string
	if d.Model != "" {
		if model := genkit.LookupModel(gk, d.Model); model != nil {
			modelName = model.Name()
		} else {
			return fmt.Errorf("invalid model '%s'", d.Model)
		}
	}

	middleware := []ai.Middleware{
		&agentDebugMiddleware{},
		&toolCallDisplayLabelMiddleware{},
	}
	for _, mwFn := range mwFuncs {
		middleware = append(middleware, mwFn(d.Name))
	}
	if mwRunner, ok := runner.(runnerWithMiddleware); ok {
		middleware = append(middleware, mwRunner.makeMiddleware()...)
	}

	withSystemPrompt := ai.WithSystem(d.SystemPrompt)
	if spr, hasPromptFn := runner.(systemPromptFuncAgentRunner[SessionInput]); hasPromptFn {
		withSystemPrompt = ai.WithSystemFn(func(ctx context.Context, i any) (string, error) {
			input, inputOk := i.(SessionInput)
			if !inputOk {
				return "", fmt.Errorf("invalid input type %T", i)
			}
			return spr.makeSystemPrompt(ctx, input)
		})
	}

	var agent *aix.Agent[State]
	if cr, ok := runner.(customAgentRunner[SessionInput, State]); ok {
		agent = genkitx.DefineCustomAgent(gk, d.Name, cr.makeAgentFunc(middleware), opts...)
	} else {
		prompt := aix.InlinePrompt{
			ai.WithModelName(modelName),
			ai.WithUse(middleware...),
			withSystemPrompt,
		}
		agent = genkitx.DefineAgent(gk, d.Name, prompt, opts...)
	}

	cfg := rez.AiAgentConfig{
		Name:        d.Name,
		DisplayName: d.Name,
		Model:       modelName,
	}

	c.agents[d.Name] = &wrappedAgent[SessionInput, State]{agent: agent, runner: runner, config: cfg}

	return nil
}

type wrappedAgent[SessionInput rez.ValidatingInput, State any] struct {
	agent  *aix.Agent[State]
	config rez.AiAgentConfig
	runner agentRunner[SessionInput, State]
}

func (w *wrappedAgent[SessionInput, State]) Config() rez.AiAgentConfig {
	return w.config
}

func (w *wrappedAgent[SessionInput, State]) DecodeSessionInput(raw []byte) (rez.ValidatingInput, error) {
	input, validationErr := w.runner.agentDefinition().DecodeSessionInput(raw)
	if validationErr != nil {
		return nil, validationErr
	} else if input == nil {
		return nil, fmt.Errorf("nil input")
	}
	return *input, nil
}

func (w *wrappedAgent[SessionInput, State]) normalizeTurnInput(input *rez.AiAgentTurnInput) (*aix.AgentInput, error) {
	if input == nil {
		return nil, rez.ErrInvalidInput
	}
	if input.Message != nil {
		return &aix.AgentInput{Message: input.Message}, nil
	}
	if input.Resume != nil && len(input.Resume.Respond)+len(input.Resume.Restart) > 0 {
		return &aix.AgentInput{Resume: input.Resume}, nil
	}
	return nil, fmt.Errorf("%w: agent turn message or resume is required", rez.ErrInvalidInput)
}

func (w *wrappedAgent[SessionInput, State]) MakeInitialTurnInput(ctx context.Context, sess *ent.AgentSession) (*rez.AiAgentTurnInput, error) {
	validated, validateErr := w.runner.agentDefinition().DecodeSessionInput(sess.Input)
	if validateErr != nil || validated == nil {
		return nil, fmt.Errorf("input: %w", validateErr)
	}
	input := *validated

	turnInput, turnErr := w.runner.makeInitialTurnInput(ctx, input)
	if turnErr != nil {
		return nil, fmt.Errorf("make turn: %w", turnErr)
	} else if turnInput == nil {
		return nil, rez.ErrInvalidInput
	}

	if msgSetter, setMsg := w.runner.(runnerWithInitialTurnMessage[SessionInput]); setMsg {
		seed, msgErr := msgSetter.updateInitialTurnMessage(ctx, input)
		if msgErr != nil {
			return nil, fmt.Errorf("update message: %w", msgErr)
		}

		if strings.TrimSpace(seed) != "" {
			task := ""
			if turnInput.Message != nil {
				task = strings.TrimSpace(turnInput.Message.Text())
			}
			turnInput.Message = ai.NewUserTextMessage(strings.TrimSpace("Context:\n" + seed + "\n\nTask:\n" + task))
		}
	}

	normalized, normalizeErr := w.normalizeTurnInput(turnInput)
	if normalizeErr != nil {
		return nil, fmt.Errorf("normalize initial input: %w", normalizeErr)
	}

	return &rez.AiAgentTurnInput{Message: normalized.Message, Resume: normalized.Resume}, nil
}

func (w *wrappedAgent[SessionInput, State]) getTurnState(ctx context.Context, params rez.InvokeAiAgentTurnParams) (*aix.SessionState[State], error) {
	state := &aix.SessionState[State]{
		SessionID: params.Session.ID.String(),
		Messages:  make([]*ai.Message, len(params.State.Messages)),
		Artifacts: make([]*aix.Artifact, len(params.State.Artifacts)),
	}

	for i, m := range params.State.Messages {
		state.Messages[i] = m.Clone()
	}

	for i, a := range params.State.Artifacts {
		state.Artifacts[i] = &aix.Artifact{Name: a.Name, Metadata: a.Metadata, Parts: a.Parts}
	}

	custom, customErr := w.runner.getCustomState(ctx, params.Session)
	if customErr != nil {
		return nil, fmt.Errorf("custom state: %w", customErr)
	} else if custom != nil {
		state.Custom = *custom
	}

	return state, nil
}

type agentInvocationContext struct {
	Session *ent.AgentSession
	Turn    *ent.AgentTurn
}

type agentInvocationContextKey struct{}

func getAgentInvocationContext(ctx context.Context) *agentInvocationContext {
	invCtx, ok := ctx.Value(agentInvocationContextKey{}).(*agentInvocationContext)
	if !ok || invCtx == nil || invCtx.Session == nil {
		return nil
	}
	return invCtx
}

func (w *wrappedAgent[SessionInput, State]) Invoke(ctx context.Context, params rez.InvokeAiAgentTurnParams) (*rez.AiAgentInvocationResult, error) {
	ctx = execution.NewAiAgentContext(ctx, params.Session, params.Turn)

	invCtx := &agentInvocationContext{Session: params.Session, Turn: params.Turn}
	ctx = context.WithValue(ctx, agentInvocationContextKey{}, invCtx)

	turnInput, inputErr := w.normalizeTurnInput(params.Input)
	if inputErr != nil {
		return nil, fmt.Errorf("turn input: %w", inputErr)
	}

	state, stateErr := w.getTurnState(ctx, params)
	if stateErr != nil {
		return nil, fmt.Errorf("turn state: %w", stateErr)
	}

	conn, connErr := w.agent.Connect(ctx, aix.WithState(state))
	if connErr != nil {
		return nil, fmt.Errorf("connect: %w", connErr)
	}

	if sendErr := conn.Send(turnInput); sendErr != nil && !errors.Is(sendErr, core.ErrActionCompleted) {
		return nil, sendErr
	}

	if closeErr := conn.Close(); closeErr != nil {
		slog.Warn("error closing input connection", "error", closeErr.Error())
	}

	if params.OnChunk != nil {
		w.handleChunks(conn, params.OnChunk)
	}

	out, outputErr := conn.Output()
	if outputErr != nil {
		return nil, fmt.Errorf("output: %w", outputErr)
	}

	if out.SessionID != params.Session.ID.String() {
		return nil, fmt.Errorf("output session ID %q does not match %q", out.SessionID, params.Session.ID)
	}

	return w.wrapInvocationOutput(out)
}

func (w *wrappedAgent[SessionInput, State]) handleChunks(conn *aix.AgentConnection[State], emitFn func(rez.AiAgentTurnChunk)) {
	for chunk, receiveErr := range conn.Receive() {
		if receiveErr != nil {
			slog.Warn("error receiving chunk", "error", receiveErr.Error())
		}
		if emitFn == nil || chunk == nil {
			continue
		}
		turnChunk := rez.AiAgentTurnChunk{
			Artifact:   chunk.Artifact,
			ModelChunk: chunk.ModelChunk,
		}
		if chunk.TurnEnd != nil {
			turnChunk.TurnEndFinishReason = &chunk.TurnEnd.FinishReason
		}
		emitFn(turnChunk)
	}
}

func (w *wrappedAgent[SessionInput, State]) wrapInvocationOutput(out *aix.AgentOutput[State]) (*rez.AiAgentInvocationResult, error) {
	if out == nil {
		return nil, fmt.Errorf("agent returned nil output")
	}
	if out.FinishReason == aix.AgentFinishReasonDetached {
		return nil, fmt.Errorf("client-managed agent detached")
	}
	if out.SnapshotID != "" {
		return nil, fmt.Errorf("client-managed agent returned snapshot ID %q", out.SnapshotID)
	}

	result := &rez.AiAgentInvocationResult{
		Response:     out.Message,
		FinishReason: out.FinishReason,
	}
	if out.Error != nil || result.FinishReason == aix.AgentFinishReasonFailed {
		result.FinishReason = aix.AgentFinishReasonFailed
		if out.Error != nil {
			result.Error = out.Error.Unwrap()
		} else {
			result.Error = fmt.Errorf("agent failed with no error")
		}
	} else if out.State == nil {
		return nil, fmt.Errorf("successful agent output has nil state")
	}

	if out.State != nil {
		result.State = rez.AiAgentTurnState{
			Messages:  out.State.Messages,
			Artifacts: out.State.Artifacts,
		}
	}
	return result, nil
}
