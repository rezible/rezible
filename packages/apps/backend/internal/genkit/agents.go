package genkit

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/firebase/genkit/go/core"
	"github.com/firebase/genkit/go/genkit"
	genkitx "github.com/firebase/genkit/go/genkit/exp"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/errs"
)

func WithAgent[I rez.ValidatingInput, S any](h agentHarness[I, S], mw ...AgentMiddlewareConstructorFn) AiRuntimeOption {
	return AiRuntimeOption{
		kind: AiRuntimeOptionKindAgent,
		runtimeFn: func(s *AiRuntime) error {
			return s.catalogue.register(s.gk, h, mw...)
		},
	}
}

type agentCatalogue struct {
	agents map[string]rez.AiAgentInvoker
}

func newAgentCatalogue() *agentCatalogue {
	return &agentCatalogue{
		agents: make(map[string]rez.AiAgentInvoker),
	}
}

func (c *agentCatalogue) GetConfigs() []rez.AiAgentConfig {
	var agents []rez.AiAgentConfig
	for _, agent := range c.agents {
		agents = append(agents, agent.Config())
	}
	return agents
}

func (c *agentCatalogue) getAgent(name string) (rez.AiAgentInvoker, error) {
	if a, ok := c.agents[name]; ok {
		return a, nil
	}
	return nil, fmt.Errorf("agent %q not found", name)
}

func (c *agentCatalogue) ValidateSessionInput(name string, input []byte) (rez.ValidatingInput, error) {
	a, agentErr := c.getAgent(name)
	if agentErr != nil {
		return nil, agentErr
	}
	return a.DecodeSessionInput(input)
}

func (c *agentCatalogue) MakeInitialTurnInput(ctx context.Context, sess *ent.AgentSession) (*rez.AiAgentTurnInput, error) {
	wrapper, wrapperErr := c.getAgent(sess.AgentName)
	if wrapperErr != nil {
		return nil, wrapperErr
	}
	return wrapper.MakeInitialTurnInput(ctx, sess)
}

type (
	agentHarness[SessionInput rez.ValidatingInput, State any] interface {
		agentDefinition() rezai.AgentDefinition[SessionInput]
		makeInitialTurnInput(context.Context, SessionInput) (*rez.AiAgentTurnInput, error)
		getCustomState(context.Context, *ent.AgentSession) (*State, error)
		transformState(context.Context, *aix.SessionState[State]) (*aix.SessionState[State], error)
		transformStreamChunk(context.Context, *aix.AgentStreamChunk) (*aix.AgentStreamChunk, error)
	}

	agentHarnessWithCustomFunc[SessionInput rez.ValidatingInput, State any] interface {
		runCustom(context.Context, aix.Responder, *aix.SessionRunner[State], *ai.Hooks) (*aix.AgentResult, error)
	}

	agentHarnessWithSystemPrompt[SessionInput rez.ValidatingInput, State any] interface {
		makeSystemPrompt(context.Context, SessionInput) (string, error)
	}

	agentHarnessWithMiddleware interface {
		makeMiddleware() []ai.Middleware
	}
)

func (c *agentCatalogue) register[SessionInput rez.ValidatingInput, State any](gk *genkit.Genkit, harness agentHarness[SessionInput, State], mwFuncs ...AgentMiddlewareConstructorFn) error {
	d := harness.agentDefinition()
	opts := []aix.AgentOption[State]{
		aix.WithDescription[State](d.Description),
		aix.WithStateTransform[State](harness.transformState),
		aix.WithStreamTransform[State](harness.transformStreamChunk),
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
	if mwRunner, ok := harness.(agentHarnessWithMiddleware); ok {
		middleware = append(middleware, mwRunner.makeMiddleware()...)
	}

	withSystemPrompt := ai.WithSystem(d.SystemPrompt)
	if spr, hasPromptFn := harness.(agentHarnessWithSystemPrompt[SessionInput, State]); hasPromptFn {
		withSystemPrompt = ai.WithSystemFn(func(ctx context.Context, i any) (string, error) {
			input, inputOk := i.(SessionInput)
			if !inputOk {
				return "", fmt.Errorf("invalid input type %T", i)
			}
			return spr.makeSystemPrompt(ctx, input)
		})
	}

	var agent *aix.Agent[State]
	if fh, hasFunc := harness.(agentHarnessWithCustomFunc[SessionInput, State]); hasFunc {
		agentFunc := func(ctx context.Context, resp aix.Responder, sess *aix.SessionRunner[State]) (*aix.AgentResult, error) {
			hooks := make([]*ai.Hooks, len(middleware))
			for i, mw := range middleware {
				mwHooks, mwErr := mw.New(ctx)
				if mwErr != nil {
					return nil, fmt.Errorf("middleware %s error: %w", mw.Name(), mwErr)
				}
				hooks[i] = mwHooks
			}
			// TODO: combine all hooks
			combinedHooks := &ai.Hooks{}
			return fh.runCustom(ctx, resp, sess, combinedHooks)
		}
		agent = genkitx.DefineCustomAgent(gk, d.Name, agentFunc, opts...)
	} else {
		prompt := aix.InlinePrompt{
			ai.WithModelName(modelName),
			ai.WithUse(middleware...),
			withSystemPrompt,
		}
		if d.MaxToolIterations != 0 {
			prompt = append(prompt, ai.WithMaxTurns(d.MaxToolIterations))
		}
		agent = genkitx.DefineAgent(gk, d.Name, prompt, opts...)
	}

	c.agents[d.Name] = &wrappedAgent[SessionInput, State]{
		agent:   agent,
		harness: harness,
		config: rez.AiAgentConfig{
			Name:        d.Name,
			DisplayName: d.Name,
			Model:       modelName,
		},
	}

	return nil
}

type wrappedAgent[SessionInput rez.ValidatingInput, State any] struct {
	agent   *aix.Agent[State]
	harness agentHarness[SessionInput, State]
	config  rez.AiAgentConfig
}

func (w *wrappedAgent[SessionInput, State]) Config() rez.AiAgentConfig {
	return w.config
}

func (w *wrappedAgent[SessionInput, State]) DecodeSessionInput(raw []byte) (rez.ValidatingInput, error) {
	input, validationErr := w.harness.agentDefinition().DecodeSessionInput(raw)
	if validationErr != nil {
		return nil, validationErr
	} else if input == nil {
		return nil, fmt.Errorf("nil input")
	}
	return *input, nil
}

func (w *wrappedAgent[SessionInput, State]) normalizeTurnInput(input *rez.AiAgentTurnInput, state *rez.AiAgentTurnState, continueFromState bool) (*aix.AgentInput, error) {
	if input == nil {
		return nil, errs.ErrInvalidInput
	}
	if continueFromState {
		if state == nil || len(state.Messages) == 0 {
			return nil, fmt.Errorf("%w: continuation requires committed messages", errs.ErrInvalidInput)
		}
		return &aix.AgentInput{}, nil
	}
	if input.Message != nil {
		return &aix.AgentInput{Message: input.Message}, nil
	}
	if input.Resume != nil && len(input.Resume.Respond)+len(input.Resume.Restart) > 0 {
		return &aix.AgentInput{Resume: input.Resume}, nil
	}
	return nil, fmt.Errorf("%w: agent turn message or resume is required", errs.ErrInvalidInput)
}

func (w *wrappedAgent[SessionInput, State]) MakeInitialTurnInput(ctx context.Context, sess *ent.AgentSession) (*rez.AiAgentTurnInput, error) {
	validated, validateErr := w.harness.agentDefinition().DecodeSessionInput(sess.Input)
	if validateErr != nil || validated == nil {
		return nil, fmt.Errorf("input: %w", validateErr)
	}
	input := *validated

	turnInput, turnErr := w.harness.makeInitialTurnInput(ctx, input)
	if turnErr != nil {
		return nil, fmt.Errorf("make turn: %w", turnErr)
	} else if turnInput == nil {
		return nil, errs.ErrInvalidInput
	}

	normalized, normalizeErr := w.normalizeTurnInput(turnInput, nil, false)
	if normalizeErr != nil {
		return nil, fmt.Errorf("normalize initial input: %w", normalizeErr)
	}

	return &rez.AiAgentTurnInput{Message: normalized.Message, Resume: normalized.Resume}, nil
}

func (w *wrappedAgent[SessionInput, State]) getTurnSessionState(ctx context.Context, params rez.InvokeAiAgentTurnParams) (*aix.SessionState[State], error) {
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

	custom, customErr := w.harness.getCustomState(ctx, params.Session)
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
	Input   *rez.AiAgentTurnInput
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
	invCtx := &agentInvocationContext{
		Session: params.Session,
		Turn:    params.Turn,
		Input:   params.Input,
	}
	ctx = context.WithValue(ctx, agentInvocationContextKey{}, invCtx)

	turnInput, inputErr := w.normalizeTurnInput(params.Input, &params.State, params.ContinueFromState)
	if inputErr != nil {
		return nil, fmt.Errorf("turn input: %w", inputErr)
	}

	state, stateErr := w.getTurnSessionState(ctx, params)
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
		for chunk, receiveErr := range conn.Receive() {
			if receiveErr != nil {
				slog.Warn("error receiving chunk", "error", receiveErr.Error())
			}
			if chunk == nil {
				continue
			}
			turnChunk := rez.AiAgentTurnChunk{
				Artifact:   chunk.Artifact,
				ModelChunk: chunk.ModelChunk,
			}
			if chunk.TurnEnd != nil {
				turnChunk.TurnEndFinishReason = &chunk.TurnEnd.FinishReason
			}
			params.OnChunk(turnChunk)
		}
	}

	out, outputErr := conn.Output()
	if out == nil {
		if outputErr == nil {
			outputErr = fmt.Errorf("agent returned nil output")
		}
		return nil, fmt.Errorf("output: %w", outputErr)
	}

	if out.SessionID != params.Session.ID.String() {
		return nil, fmt.Errorf("output session ID %q does not match %q", out.SessionID, params.Session.ID)
	}

	result, wrapErr := w.wrapInvocationOutput(out)
	if wrapErr != nil {
		return nil, wrapErr
	}
	if outputErr != nil && result.Error == nil {
		result.FinishReason = aix.AgentFinishReasonFailed
		result.Error = outputErr
	}
	return result, nil
}

func (w *wrappedAgent[SessionInput, State]) wrapInvocationOutput(out *aix.AgentOutput[State]) (*rez.AiAgentInvocationResult, error) {
	if out == nil {
		return nil, fmt.Errorf("agent returned nil output")
	}
	if out.FinishReason == aix.AgentFinishReasonDetached {
		return nil, fmt.Errorf("agent detached")
	}
	if out.SnapshotID != "" {
		return nil, fmt.Errorf("agent returned snapshot ID %q", out.SnapshotID)
	}

	result := &rez.AiAgentInvocationResult{
		Response:     out.Message,
		FinishReason: out.FinishReason,
	}
	if out.Error != nil || result.FinishReason == aix.AgentFinishReasonFailed {
		result.FinishReason = aix.AgentFinishReasonFailed
		if out.Error != nil {
			result.Error = out.Error
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
