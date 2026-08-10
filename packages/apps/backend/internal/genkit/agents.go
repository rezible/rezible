package genkit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/firebase/genkit/go/core"
	genkitx "github.com/firebase/genkit/go/genkit/exp"
	middlewarex "github.com/firebase/genkit/go/plugins/middleware/exp"
	rez "github.com/rezible/rezible"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/execution"
)

type (
	agentRunner[I rezai.AgentInput, S rezai.SessionState] interface {
		agentDefinition() rezai.AgentDefinition[I, S]
		makeInitialTurnInput(context.Context, I) (*rez.AiAgentTurnInput, error)
		transformState(context.Context, *aix.SessionState[S]) (*aix.SessionState[S], error)
		transformStreamChunk(context.Context, *aix.AgentStreamChunk) (*aix.AgentStreamChunk, error)
	}

	customAgentRunner[I rezai.AgentInput, S rezai.SessionState] interface {
		makeAgentFunc([]ai.Middleware) aix.AgentFunc[S]
	}

	initialContextSeeder[I rezai.AgentInput] interface {
		makeInitialContextSeed(context.Context, I) (string, error)
	}

	runnerMiddlewareProvider interface {
		makeMiddleware() []ai.Middleware
	}

	AgentWrapper interface {
		AgentConfig() rez.AiAgentConfig
		ValidateInput([]byte) (rez.AiAgentSessionInput, error)
		MakeInitialTurnInput(context.Context, []byte) (*rez.AiAgentTurnInput, error)
		Invoke(context.Context, rez.InvokeAgentTurnParams) (*rez.AiAgentInvocationResult, error)
	}
)

func makeAgentWrapper[I rezai.AgentInput, S rezai.SessionState](svc *AiService, runner agentRunner[I, S]) (AgentWrapper, error) {
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

	middleware := []ai.Middleware{
		&toolCallDisplayLabelMiddleware{},
	}

	if d.EnableKnowledgeGraph {
		middleware = append(middleware, newKnowledgeGraphMiddleware(svc.knowledge, runner))
	}

	if mp, ok := runner.(runnerMiddlewareProvider); ok {
		middleware = append(middleware, mp.makeMiddleware()...)
	}

	if d.EnableArtifacts {
		middleware = append(middleware, &middlewarex.Artifacts{})
	}

	var agent *aix.Agent[S]
	if cr, ok := runner.(customAgentRunner[I, S]); ok {
		agent = genkitx.DefineCustomAgent(svc.gk, d.Name, cr.makeAgentFunc(middleware), opts...)
	} else {
		prompt := aix.InlinePrompt{
			modelOpt,
			ai.WithSystem(d.SystemPrompt),
			ai.WithUse(middleware...),
		}
		agent = genkitx.DefineAgent(svc.gk, d.Name, prompt, opts...)
	}
	return &agentWrapper[I, S]{agent: agent, runner: runner}, nil
}

type agentWrapper[I rezai.AgentInput, S rezai.SessionState] struct {
	agent  *aix.Agent[S]
	runner agentRunner[I, S]
}

func (w *agentWrapper[I, S]) AgentConfig() rez.AiAgentConfig {
	d := w.runner.agentDefinition()
	return rez.AiAgentConfig{
		Name:        d.Name,
		DisplayName: "",
		Model:       "",
	}
}

func (w *agentWrapper[I, S]) ValidateInput(raw []byte) (rez.AiAgentSessionInput, error) {
	inp, validationErr := w.runner.agentDefinition().ValidateInput(raw)
	if validationErr != nil {
		return nil, validationErr
	} else if inp == nil {
		return nil, fmt.Errorf("nil input")
	}
	return *inp, nil
}

func (w *agentWrapper[I, S]) normalizeTurnInput(input *rez.AiAgentTurnInput) (*rez.AiAgentTurnInput, error) {
	if input == nil {
		return nil, rez.ErrInvalidInput
	}
	if input.Resume != nil && len(input.Resume.Respond)+len(input.Resume.Restart) == 0 {
		return &rez.AiAgentTurnInput{Message: input.Message}, nil
	}
	if input.Message == nil && input.Resume == nil {
		return nil, fmt.Errorf("%w: agent turn message or resume is required", rez.ErrInvalidInput)
	}
	return input, nil
}

func (w *agentWrapper[I, S]) MakeInitialTurnInput(ctx context.Context, raw []byte) (*rez.AiAgentTurnInput, error) {
	input, inputErr := w.runner.agentDefinition().ValidateInput(raw)
	if inputErr != nil || input == nil {
		return nil, fmt.Errorf("input: %w", inputErr)
	}
	turnInput, turnErr := w.runner.makeInitialTurnInput(ctx, *input)
	if turnErr != nil {
		return nil, turnErr
	} else if turnInput == nil {
		return nil, rez.ErrInvalidInput
	}
	if seeder, ok := w.runner.(initialContextSeeder[I]); ok {
		seed, seedErr := seeder.makeInitialContextSeed(ctx, *input)
		if seedErr != nil {
			return nil, seedErr
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

	return normalized, nil
}

func (w *agentWrapper[I, S]) Invoke(ctx context.Context, params rez.InvokeAgentTurnParams) (*rez.AiAgentInvocationResult, error) {
	sess := params.Session
	ctx = execution.NewAiAgentContext(ctx, sess, params.Turn)

	input, inputErr := w.normalizeTurnInput(params.Input)
	if inputErr != nil {
		return nil, inputErr
	}

	state := &aix.SessionState[S]{SessionID: sess.ID.String()}
	if params.Parent != nil {
		if unmarshalErr := json.Unmarshal(params.Parent.State, state); unmarshalErr != nil {
			return nil, fmt.Errorf("decode parent state: %w", unmarshalErr)
		}
		if state.SessionID != sess.ID.String() {
			return nil, fmt.Errorf("parent state session ID %q does not match %q", state.SessionID, sess.ID)
		}
	}

	agentInput := &aix.AgentInput{
		Message: input.Message,
	}
	if input.Resume != nil {
		agentInput.Resume = &aix.ToolResume{
			Respond: input.Resume.Respond,
			Restart: input.Resume.Restart,
		}
	}

	fmt.Printf("running: %+v\n", agentInput.Message.Text())

	conn, connErr := w.agent.Connect(ctx, aix.WithState(state))
	if connErr != nil {
		return nil, fmt.Errorf("connect: %w", connErr)
	}
	if sendErr := conn.Send(agentInput); sendErr != nil && !errors.Is(sendErr, core.ErrActionCompleted) {
		return nil, sendErr
	}

	if closeErr := conn.Close(); closeErr != nil {
		slog.Warn("error closing input connection", "error", closeErr.Error())
	}

	for chunk, receiveErr := range conn.Receive() {
		if receiveErr != nil {
			slog.Warn("error receiving chunk", "error", receiveErr.Error())
		}
		if chunk != nil && params.OnChunk != nil {
			var finishReason *aix.AgentFinishReason
			if chunk.TurnEnd != nil {
				finishReason = &chunk.TurnEnd.FinishReason
			}
			params.OnChunk(rez.AiAgentTurnChunk{
				Artifact:            chunk.Artifact,
				ModelChunk:          chunk.ModelChunk,
				TurnEndFinishReason: finishReason,
			})
		}
	}

	out, outputErr := conn.Output()
	if outputErr != nil {
		return nil, fmt.Errorf("output: %w", outputErr)
	}

	if out.SessionID != sess.ID.String() {
		return nil, fmt.Errorf("output session ID %q does not match %q", out.SessionID, sess.ID)
	}

	result, wrapErr := w.getOutputResult(out)
	if wrapErr != nil {
		return nil, wrapErr
	}
	return result, nil
}

func (w *agentWrapper[I, S]) getOutputResult(out *aix.AgentOutput[S]) (*rez.AiAgentInvocationResult, error) {
	if out == nil {
		return nil, fmt.Errorf("agent returned nil output")
	}
	if out.FinishReason == aix.AgentFinishReasonDetached {
		return nil, fmt.Errorf("client-managed agent detached")
	}
	if out.SnapshotID != "" {
		return nil, fmt.Errorf("client-managed agent returned snapshot ID %q", out.SnapshotID)
	}

	citations, citationsErr := getAgentKnowledgeCitations(out.Artifacts)
	if citationsErr != nil {
		return nil, fmt.Errorf("unable to get agent knowledge citations: %w", citationsErr)
	}

	result := &rez.AiAgentInvocationResult{
		Response:           out.Message,
		FinishReason:       out.FinishReason,
		Error:              out.Error,
		KnowledgeCitations: citations,
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
