package genkit

import (
	"context"
	"encoding/json"
	"fmt"

	middlewarex "github.com/firebase/genkit/go/plugins/middleware/exp"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"

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
		definition() rezai.AgentDefinition[I, S, O]
		transformState(context.Context, *aix.SessionState[S]) (*aix.SessionState[S], error)
		transformStreamChunk(context.Context, *aix.AgentStreamChunk) (*aix.AgentStreamChunk, error)
		makeInitialUserMessage(context.Context, I) (*ai.Message, error)
	}

	customAgentRunner[I rezai.AgentInput, S rezai.SessionState, O rezai.AgentOutput] interface {
		makeAgentFunc([]ai.Middleware, []ai.ToolRef) aix.AgentFunc[S]
		//run(context.Context, aix.Responder, *aix.SessionRunner[S]) (*aix.AgentResult, error)
	}
)

func wrapAgentRunner[I rezai.AgentInput, S rezai.SessionState, O rezai.AgentOutput](svc *AiService, runner agentRunner[I, S, O]) (*agentWrapper, error) {
	def := runner.definition()
	opts := []aix.AgentOption[S]{
		aix.WithSessionStore[S](makeSessionStore[S](svc.sessions)),
		aix.WithDescription[S](def.Description),
		aix.WithStateTransform[S](runner.transformState),
		aix.WithStreamTransform[S](runner.transformStreamChunk),
	}

	//model, modelErr := svc.getModel()

	tools, toolsErr := svc.getRequiredToolRefs(def.RequiredTools)
	if toolsErr != nil {
		return nil, fmt.Errorf("tools: %w", toolsErr)
	}
	middleware := []ai.Middleware{
		newAgentRunOutputWriter[S, O](svc.sessions),
	}
	if def.EnableArtifacts {
		middleware = append(middleware, &middlewarex.Artifacts{})
	}

	var agent *aix.Agent[S]
	if cr, ok := runner.(customAgentRunner[I, S, O]); ok {
		runFunc := cr.makeAgentFunc(middleware, tools)
		agent = genkitx.DefineCustomAgent(svc.gk, def.Name, runFunc, opts...)
	} else {
		systemPrompt := fmt.Sprintf("%s\n\nRemember to write outputs using the 'write_output' tool!!", def.SystemPrompt)
		prompt := aix.InlinePrompt{
			ai.WithModel(flashModel),
			ai.WithSystem(systemPrompt),
			ai.WithUse(middleware...),
			ai.WithTools(tools...),
		}
		agent = genkitx.DefineAgent(svc.gk, def.Name, prompt, opts...)
	}

	validateInputFunc := func(raw []byte) (*I, error) {
		var input I
		if jsonErr := json.Unmarshal(raw, &input); jsonErr != nil {
			return nil, fmt.Errorf("unmarshal: %w", jsonErr)
		}
		if validationErr := input.Validate(); validationErr != nil {
			return nil, fmt.Errorf("validate: %w", validationErr)
		}
		return &input, nil
	}

	return &agentWrapper{
		inputValidatorFunc: func(input []byte) error {
			_, err := validateInputFunc(input)
			return err
		},
		makeRunnerFunc: func(run *ent.AiAgentRun) rez.AiAgentInvoker {
			return &agentSessionInvoker[I, S]{
				agent:     agent,
				sessionId: run.ID.String(),
				makeInitialUserMessageFn: func(ctx context.Context) (*ai.Message, error) {
					input, inputErr := validateInputFunc(run.Input)
					if inputErr != nil || input == nil {
						return nil, fmt.Errorf("input: %w", inputErr)
					}
					return runner.makeInitialUserMessage(ctx, *input)
				},
			}
		},
	}, nil
}

type agentWrapper struct {
	inputValidatorFunc func([]byte) error
	makeRunnerFunc     func(run *ent.AiAgentRun) rez.AiAgentInvoker
}

func (w *agentWrapper) MakeRunner(run *ent.AiAgentRun) rez.AiAgentInvoker {
	return w.makeRunnerFunc(run)
}

func (w *agentWrapper) ValidateInput(raw []byte) error {
	return w.inputValidatorFunc(raw)
}

type agentSessionInvoker[I rezai.AgentInput, S rezai.SessionState] struct {
	agent                    *aix.Agent[S]
	sessionId                string
	makeInitialUserMessageFn func(context.Context) (*ai.Message, error)
}

func (i *agentSessionInvoker[I, S]) checkAgentOutput(out *aix.AgentOutput[S]) (uuid.UUID, error) {
	if out.FinishReason == aix.AgentFinishReasonFailed {
		return uuid.Nil, fmt.Errorf("agent failed: %s", out.Error.Error())
	}
	snapshotId, idErr := uuid.Parse(out.SnapshotID)
	if idErr != nil {
		return uuid.Nil, fmt.Errorf("parse snapshot id: %w", idErr)
	}
	return snapshotId, nil
}

func (i *agentSessionInvoker[I, S]) Start(ctx context.Context) (uuid.UUID, error) {
	initialMsg, initialMsgErr := i.makeInitialUserMessageFn(ctx)
	if initialMsgErr != nil {
		return uuid.Nil, fmt.Errorf("create initial user message: %w", initialMsgErr)
	}
	input := &aix.AgentInput{
		Message: initialMsg,
		Detach:  false,
	}
	out, runErr := i.agent.Run(ctx, input, aix.WithSessionID[S](i.sessionId))
	if runErr != nil {
		return uuid.Nil, fmt.Errorf("run agent: %w", runErr)
	}
	return i.checkAgentOutput(out)
}

func (i *agentSessionInvoker[I, S]) Invoke(ctx context.Context, parentId *uuid.UUID, msg *ai.Message, resume *ai.GenerateActionResume) (uuid.UUID, error) {
	invokeOpts := []aix.InvocationOption[S]{
		aix.WithSessionID[S](i.sessionId),
	}
	input := &aix.AgentInput{
		Message: msg,
	}
	if resume != nil {
		input.Resume = &aix.ToolResume{
			Respond: resume.Respond,
			Restart: resume.Restart,
		}
	}

	if parentId != nil {
		invokeOpts = append(invokeOpts, aix.WithSnapshotID[S](parentId.String()))
	} else {
		parent, parentErr := i.agent.Store().GetLatestSnapshot(ctx, i.sessionId)
		if parentErr != nil {
			return uuid.Nil, fmt.Errorf("get parent snapshot: %w", parentErr)
		}
		if parent != nil {
			invokeOpts = append(invokeOpts, aix.WithSnapshotID[S](parent.SnapshotID))
		} else if msg == nil {
			// no parent, no message == session start
			initialMsg, initialMsgErr := i.makeInitialUserMessageFn(ctx)
			if initialMsgErr != nil {
				return uuid.Nil, fmt.Errorf("create initial user message: %w", initialMsgErr)
			}
			input.Message = initialMsg
		}
	}
	out, runErr := i.agent.Run(ctx, input, invokeOpts...)
	if runErr != nil {
		return uuid.Nil, fmt.Errorf("run agent: %w", runErr)
	}
	return i.checkAgentOutput(out)
}

/*
func (i *agentSessionInvoker[S]) createInitialSessionSnapshot(ctx context.Context) (*aix.SessionSnapshot[S], error) {
	var input I
	if jsonErr := json.Unmarshal(i.run.Input, &input); jsonErr != nil {
		return nil, fmt.Errorf("unmarshal input: %w", jsonErr)
	}
	state, stateErr := i.runner.makeInitialSessionState(ctx, input)
	if stateErr != nil {
		return nil, fmt.Errorf("initial state: %w", stateErr)
	}
	createSnapshotFn := func(_ *aix.SessionSnapshot[S]) (*aix.SessionSnapshot[S], error) {
		return &aix.SessionSnapshot[S]{
			SessionID:    i.sessionId(),
			FinishReason: aix.AgentFinishReasonStop,
			Status:       aix.SnapshotStatusCompleted,
			State:        state,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}, nil
	}
	snapshot, snapshotErr := i.agent.Store().SaveSnapshot(ctx, "", createSnapshotFn)
	if snapshotErr != nil {
		return nil, fmt.Errorf("save snapshot: %w", snapshotErr)
	}
	return snapshot, nil
}
*/

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
