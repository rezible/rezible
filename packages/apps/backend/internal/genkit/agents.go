package genkit

import (
	"context"
	"fmt"

	middlewarex "github.com/firebase/genkit/go/plugins/middleware/exp"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/pkg/execution"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	genkitx "github.com/firebase/genkit/go/genkit/exp"
	rez "github.com/rezible/rezible"
	rezai "github.com/rezible/rezible/pkg/ai"
)

type (
	agentRunner[I rezai.AgentInput, S rezai.SessionState, O rezai.AgentOutput] interface {
		agentDefinition() rezai.AgentDefinition[I, S, O]
		transformState(context.Context, *aix.SessionState[S]) (*aix.SessionState[S], error)
		transformStreamChunk(context.Context, *aix.AgentStreamChunk) (*aix.AgentStreamChunk, error)
		makeInitialUserMessage(context.Context, I) (*ai.Message, error)
	}

	customAgentRunner[I rezai.AgentInput, S rezai.SessionState, O rezai.AgentOutput] interface {
		makeAgentFunc([]ai.Middleware, []ai.ToolRef) aix.AgentFunc[S]
	}

	AgentWrapper interface {
		ValidateInput([]byte) error
		MakeInvoker(*ent.AiAgentRun) rez.AiAgentRunInvoker
	}
)

func makeAgentWrapper[I rezai.AgentInput, S rezai.SessionState, O rezai.AgentOutput](svc *AiService, runner agentRunner[I, S, O]) (AgentWrapper, error) {
	d := runner.agentDefinition()
	opts := []aix.AgentOption[S]{
		aix.WithSessionStore[S](makeSessionStore[S](svc.sessions)),
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
		//} else if d.Prompt != "" {
		//genkitx.DefinePromptAgent(svc.gk, d.Name)
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

func (w *agentWrapper[I, S, O]) ValidateInput(raw []byte) error {
	_, err := w.runner.agentDefinition().ValidateInput(raw)
	return err
}

func (w *agentWrapper[I, S, O]) MakeInvoker(run *ent.AiAgentRun) rez.AiAgentRunInvoker {
	return &agentSessionInvoker[I, S, O]{agent: w.agent, runner: w.runner, run: run}
}

type agentSessionInvoker[I rezai.AgentInput, S rezai.SessionState, O rezai.AgentOutput] struct {
	agent  *aix.Agent[S]
	runner agentRunner[I, S, O]
	run    *ent.AiAgentRun
}

func (i *agentSessionInvoker[I, S, O]) Invoke(ctx context.Context, parentId *uuid.UUID, msg *ai.Message, resume *ai.GenerateActionResume) (uuid.UUID, error) {
	ctx = execution.NewAiAgentRunContext(ctx, i.run)

	input := &aix.AgentInput{Message: msg}
	if resume != nil {
		input.Resume = &aix.ToolResume{Respond: resume.Respond, Restart: resume.Restart}
	}

	invokeOpts, optsErr := i.getInvokeOpts(ctx, parentId, input)
	if optsErr != nil {
		return uuid.Nil, fmt.Errorf("get invoke opts: %w", optsErr)
	}

	out, runErr := i.agent.Run(ctx, input, invokeOpts...)
	if runErr != nil {
		return uuid.Nil, fmt.Errorf("run agent: %w", runErr)
	}

	return i.checkRunOutput(out)
}

func (i *agentSessionInvoker[I, S, O]) getInvokeOpts(ctx context.Context, parentId *uuid.UUID, input *aix.AgentInput) ([]aix.InvocationOption[S], error) {
	invokeOpts := []aix.InvocationOption[S]{
		aix.WithSessionID[S](i.run.ID.String()),
	}
	if parentId != nil {
		invokeOpts = append(invokeOpts, aix.WithSnapshotID[S](parentId.String()))
	} else {
		parent, parentErr := i.agent.Store().GetLatestSnapshot(ctx, i.run.ID.String())
		if parentErr != nil {
			return nil, fmt.Errorf("get parent snapshot: %w", parentErr)
		}

		if parent != nil {
			invokeOpts = append(invokeOpts, aix.WithSnapshotID[S](parent.SnapshotID))
		} else if input.Message == nil && input.Resume == nil {
			// no parent, no message == session start
			runInput, runInputErr := i.runner.agentDefinition().ValidateInput(i.run.Input)
			if runInputErr != nil || runInput == nil {
				return nil, fmt.Errorf("input: %w", runInputErr)
			}
			initialMsg, initialMsgErr := i.runner.makeInitialUserMessage(ctx, *runInput)
			if initialMsgErr != nil {
				return nil, fmt.Errorf("create initial user message: %w", initialMsgErr)
			}
			input.Message = initialMsg
		}
	}
	return invokeOpts, nil
}

func (i *agentSessionInvoker[I, S, O]) checkRunOutput(out *aix.AgentOutput[S]) (uuid.UUID, error) {
	if out.FinishReason == aix.AgentFinishReasonFailed {
		return uuid.Nil, fmt.Errorf("agent failed: %s", out.Error.Error())
	}
	snapshotId, idErr := uuid.Parse(out.SnapshotID)
	if idErr != nil {
		return uuid.Nil, fmt.Errorf("parse snapshot id: %w", idErr)
	}
	return snapshotId, nil
}

/*
func (i *agentSessionInvoker[S]) Resume(ctx context.Context, params rez.ResumeAgentRunParams) (uuid.UUID, error) {
	if (len(params.Respond) + len(params.Restart)) == 0 {
		return uuid.Nil, fmt.Errorf("invalid resume params")
	}
	invokeOpts, invokeOptsErr := i.getInvokeOpts(ctx, params.ParentSnapshotID)
	if invokeOptsErr != nil {
		return uuid.Nil, fmt.Errorf("invoke opts: %w", invokeOptsErr)
	}
	conn, connErr := i.agent.Connect(ctx, invokeOpts...)
	if connErr != nil {
		return uuid.Nil, fmt.Errorf("connect: %w", connErr)
	}
	resume := &aix.ToolResume{Respond: params.Respond, Restart: params.Restart}
	if sendErr := conn.SendResume(resume); sendErr != nil {
		return uuid.Nil, fmt.Errorf("send resume: %w", sendErr)
	}
	out, outErr := conn.Output()
	if outErr != nil {
		return uuid.Nil, fmt.Errorf("output: %w", outErr)
	}
	return uuid.Parse(out.SnapshotID)
}
*/
