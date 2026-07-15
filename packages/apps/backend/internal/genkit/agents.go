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
	AgentInvoker interface {
		ValidateInput([]byte) error
		MakeRunner(*ent.AiAgentRun) rez.AiAgentInvoker
	}

	agentRunner[I rezai.AgentInput, S rezai.SessionState, O rezai.AgentOutput] interface {
		agentDefinition() rezai.AgentDefinition[I, S, O]
		transformState(context.Context, *aix.SessionState[S]) (*aix.SessionState[S], error)
		transformStreamChunk(context.Context, *aix.AgentStreamChunk) (*aix.AgentStreamChunk, error)
		makeInitialUserMessage(context.Context, I) (*ai.Message, error)
	}

	customAgentRunner[I rezai.AgentInput, S rezai.SessionState, O rezai.AgentOutput] interface {
		makeAgentFunc([]ai.Middleware, []ai.ToolRef) aix.AgentFunc[S]
	}
)

func wrapAgentRunner[I rezai.AgentInput, S rezai.SessionState, O rezai.AgentOutput](svc *AiService, ar agentRunner[I, S, O]) (*agentWrapper, error) {
	d := ar.agentDefinition()
	opts := []aix.AgentOption[S]{
		aix.WithSessionStore[S](makeSessionStore[S](svc.sessions)),
		aix.WithDescription[S](d.Description),
		aix.WithStateTransform[S](ar.transformState),
		aix.WithStreamTransform[S](ar.transformStreamChunk),
	}

	registeredTools, dynamicTools := svc.getRegisteredTools(d.RequiredTools)

	agentMw, mwErr := newAgentRunnerMiddleware(ar, dynamicTools)
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
	if cr, ok := ar.(customAgentRunner[I, S, O]); ok {
		runFunc := cr.makeAgentFunc(middleware, registeredTools)
		agent = genkitx.DefineCustomAgent(svc.gk, d.Name, runFunc, opts...)
	} else {
		systemPrompt := fmt.Sprintf("%s\n\nRemember to write outputs using the 'write_output' tool!!", d.SystemPrompt)
		prompt := aix.InlinePrompt{
			ai.WithModel(flashModel),
			ai.WithSystem(systemPrompt),
			ai.WithUse(middleware...),
			ai.WithTools(registeredTools...),
		}
		agent = genkitx.DefineAgent(svc.gk, d.Name, prompt, opts...)
	}

	validateInputFunc := func(input []byte) error {
		_, err := d.ValidateInput(input)
		return err
	}

	makeInvokerFunc := func(run *ent.AiAgentRun) rez.AiAgentInvoker {
		return &agentSessionInvoker[I, S, O]{
			agent:  agent,
			run:    run,
			runner: ar,
		}
	}

	return &agentWrapper{
		inputValidatorFunc: validateInputFunc,
		makeInvokerFunc:    makeInvokerFunc,
	}, nil
}

type agentWrapper struct {
	inputValidatorFunc func([]byte) error
	makeInvokerFunc    func(run *ent.AiAgentRun) rez.AiAgentInvoker
}

func (w *agentWrapper) MakeRunner(run *ent.AiAgentRun) rez.AiAgentInvoker {
	return w.makeInvokerFunc(run)
}

func (w *agentWrapper) ValidateInput(raw []byte) error {
	return w.inputValidatorFunc(raw)
}

type agentSessionInvoker[I rezai.AgentInput, S rezai.SessionState, O rezai.AgentOutput] struct {
	agent  *aix.Agent[S]
	run    *ent.AiAgentRun
	runner agentRunner[I, S, O]
}

func (i *agentSessionInvoker[I, S, O]) Invoke(ctx context.Context, parentId *uuid.UUID, msg *ai.Message, resume *ai.GenerateActionResume) (uuid.UUID, error) {
	input := &aix.AgentInput{Message: msg}
	if resume != nil {
		input.Resume = &aix.ToolResume{
			Respond: resume.Respond,
			Restart: resume.Restart,
		}
	}

	ctx = execution.NewAiAgentRunContext(ctx, i.run)

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
