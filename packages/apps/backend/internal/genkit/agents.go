package genkit

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/firebase/genkit/go/core"
	genkitx "github.com/firebase/genkit/go/genkit/exp"
	middlewarex "github.com/firebase/genkit/go/plugins/middleware/exp"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/execution"
)

type (
	agentRunner[I rezai.AgentInput, S rezai.SessionState, O rezai.AgentOutput] interface {
		agentDefinition() rezai.AgentDefinition[I, S, O]
		makeInitialTurnInput(context.Context, I) (*rez.AgentTurnInput, error)
		transformState(context.Context, *aix.SessionState[S]) (*aix.SessionState[S], error)
		transformStreamChunk(context.Context, *aix.AgentStreamChunk) (*aix.AgentStreamChunk, error)
	}

	customAgentRunner[I rezai.AgentInput, S rezai.SessionState, O rezai.AgentOutput] interface {
		makeAgentFunc([]ai.Middleware, []ai.ToolRef) aix.AgentFunc[S]
	}

	AgentWrapper interface {
		ValidateAndEncodeInput(any) ([]byte, error)
		MakeInitialTurnInput(context.Context, []byte) (*rez.AgentTurnInput, error)
		Invoke(context.Context, *ent.AgentSession, *ent.AgentTurn, []byte, *rez.AgentTurnInput) (*rez.AgentTurnResult, error)
	}
)

func makeAgentWrapper[I rezai.AgentInput, S rezai.SessionState, O rezai.AgentOutput](svc *AiService, runner agentRunner[I, S, O]) (AgentWrapper, error) {
	d := runner.agentDefinition()
	opts := []aix.AgentOption[S]{
		aix.WithDescription[S](d.Description),
		aix.WithStateTransform[S](runner.transformState),
		aix.WithStreamTransform[S](runner.transformStreamChunk),
	}

	modelOpt := ai.WithModel(svc.getDefaultModel())
	if d.Model != "" {
		modelOpt = ai.WithModelName(d.Model)
	}
	registeredTools, dynamicTools := svc.getRegisteredTools(d.RequiredTools)
	agentMw, mwErr := newAgentRunnerMiddleware(runner, dynamicTools)
	if mwErr != nil {
		return nil, fmt.Errorf("agent runner middleware: %w", mwErr)
	}
	middleware := []ai.Middleware{
		agentMw,
		newKnowledgeGraphMiddleware[S](svc.knowledgeGraph),
	}
	if d.EnableArtifacts {
		middleware = append(middleware, &middlewarex.Artifacts{})
	}

	var agent *aix.Agent[S]
	if cr, ok := runner.(customAgentRunner[I, S, O]); ok {
		agent = genkitx.DefineCustomAgent(svc.gk, d.Name, cr.makeAgentFunc(middleware, registeredTools), opts...)
	} else {
		prompt := aix.InlinePrompt{
			modelOpt,
			ai.WithSystem(d.SystemPrompt),
			ai.WithUse(middleware...),
			ai.WithTools(registeredTools...),
		}
		agent = genkitx.DefineAgent(svc.gk, d.Name, prompt, opts...)
	}
	return &agentWrapper[I, S, O]{agent: agent, runner: runner}, nil
}

type agentWrapper[I rezai.AgentInput, S rezai.SessionState, O rezai.AgentOutput] struct {
	agent  *aix.Agent[S]
	runner agentRunner[I, S, O]
}

func (w *agentWrapper[I, S, O]) ValidateAndEncodeInput(input any) ([]byte, error) {
	raw, rawOK := input.([]byte)
	if !rawOK {
		var marshalErr error
		raw, marshalErr = json.Marshal(input)
		if marshalErr != nil {
			return nil, fmt.Errorf("marshal input: %w", marshalErr)
		}
	}
	if _, validationErr := w.runner.agentDefinition().ValidateInput(raw); validationErr != nil {
		return nil, validationErr
	}
	return raw, nil
}

func (w *agentWrapper[I, S, O]) normalizeTurnInput(input *rez.AgentTurnInput) (*rez.AgentTurnInput, error) {
	if input == nil {
		return nil, rez.ErrInvalidInput
	}
	if input.Resume != nil && len(input.Resume.Respond)+len(input.Resume.Restart) == 0 {
		return &rez.AgentTurnInput{Message: input.Message}, nil
	}
	if input.Message == nil && input.Resume == nil {
		return nil, fmt.Errorf("%w: agent turn message or resume is required", rez.ErrInvalidInput)
	}
	return input, nil
}

func (w *agentWrapper[I, S, O]) MakeInitialTurnInput(ctx context.Context, raw []byte) (*rez.AgentTurnInput, error) {
	input, inputErr := w.runner.agentDefinition().ValidateInput(raw)
	if inputErr != nil || input == nil {
		return nil, fmt.Errorf("input: %w", inputErr)
	}
	turnInput, turnErr := w.runner.makeInitialTurnInput(ctx, *input)
	if turnErr != nil {
		return nil, turnErr
	}
	return w.normalizeTurnInput(turnInput)
}

func (w *agentWrapper[I, S, O]) Invoke(ctx context.Context, session *ent.AgentSession, turn *ent.AgentTurn, parentState []byte, input *rez.AgentTurnInput) (*rez.AgentTurnResult, error) {
	ctx = execution.NewAiAgentContext(ctx, session, turn)

	var inputErr error
	if input, inputErr = w.normalizeTurnInput(input); inputErr != nil {
		return nil, inputErr
	}

	state := &aix.SessionState[S]{SessionID: session.ID.String()}
	if len(parentState) > 0 {
		if unmarshalErr := json.Unmarshal(parentState, state); unmarshalErr != nil {
			return nil, fmt.Errorf("decode parent state: %w", unmarshalErr)
		}
		if state.SessionID != session.ID.String() {
			return nil, fmt.Errorf("parent state session ID %q does not match %q", state.SessionID, session.ID)
		}
	}

	var agentResume *aix.ToolResume
	if input.Resume != nil {
		agentResume = &aix.ToolResume{Respond: input.Resume.Respond, Restart: input.Resume.Restart}
	}
	agentInput := &aix.AgentInput{Message: input.Message, Resume: agentResume}
	out, runErr := w.agent.Run(ctx, agentInput, aix.WithState(state))
	if runErr != nil {
		return nil, fmt.Errorf("run agent: %w", runErr)
	}
	return w.wrapAgentOutput(session.ID, out)
}

func (w *agentWrapper[I, S, O]) wrapAgentOutput(sessId uuid.UUID, out *aix.AgentOutput[S]) (*rez.AgentTurnResult, error) {
	if out == nil {
		return nil, fmt.Errorf("agent returned nil output")
	}
	if out.FinishReason == aix.AgentFinishReasonDetached {
		return nil, fmt.Errorf("client-managed agent detached")
	}
	if out.SnapshotID != "" {
		return nil, fmt.Errorf("client-managed agent returned snapshot ID %q", out.SnapshotID)
	}
	if out.SessionID != sessId.String() {
		return nil, fmt.Errorf("output session ID %q does not match %q", out.SessionID, sessId)
	}
	if out.State != nil && out.State.SessionID != sessId.String() {
		return nil, fmt.Errorf("output state session ID %q does not match %q", out.State.SessionID, sessId)
	}

	result := &rez.AgentTurnResult{
		FinishReason: out.FinishReason,
		Error:        out.Error,
	}
	if out.Error != nil || result.FinishReason == aix.AgentFinishReasonFailed {
		result.FinishReason = aix.AgentFinishReasonFailed
		if result.Error == nil {
			result.Error = core.NewError(core.INTERNAL, "agent returned failed finish reason without an error")
		}
	} else if out.State == nil {
		return nil, fmt.Errorf("successful agent output has nil state")
	}

	if out.State != nil {
		stateJson, stateJsonErr := json.Marshal(out.State)
		if stateJsonErr != nil {
			return nil, fmt.Errorf("encode output state: %w", stateJsonErr)
		}
		result.State = stateJson
	}
	return result, nil
}
