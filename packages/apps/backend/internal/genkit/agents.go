package genkit

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/firebase/genkit/go/genkit"
	genkitx "github.com/firebase/genkit/go/genkit/exp"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	rezai "github.com/rezible/rezible/pkg/ai"
)

type (
	AgentRunInvokerFunc func(run *ent.AiAgentRun) rez.AiAgentRunInvoker

	agentRunner[S rezai.AgentState] interface {
		definition() rezai.AgentDefinition[S]
		makeInitialState([]byte) (*ai.Message, *aix.SessionState[S], error)
		run(*genkit.Genkit) aix.AgentFunc[S]
	}

	agentWithStateTransformer[S rezai.AgentState] interface {
		transformState(context.Context, *aix.SessionState[S]) (*aix.SessionState[S], error)
	}

	agentWithStreamChunkTransformer[S rezai.AgentState] interface {
		transformStreamChunk(context.Context, *aix.AgentStreamChunk) (*aix.AgentStreamChunk, error)
	}
)

func makeAgentSessionInvokerFunc[S rezai.AgentState](g *genkit.Genkit, s aix.SessionStore[S], r agentRunner[S]) AgentRunInvokerFunc {
	name := r.definition().Name
	opts := []aix.AgentOption[S]{
		aix.WithSessionStore[S](s),
		aix.WithDescription[S]("Runner for agent " + name),
	}
	if sta, ok := r.(agentWithStateTransformer[S]); ok {
		opts = append(opts, aix.WithStateTransform(sta.transformState))
	}
	if scta, ok := r.(agentWithStreamChunkTransformer[S]); ok {
		opts = append(opts, aix.WithStreamTransform[S](scta.transformStreamChunk))
	}
	agent := genkitx.DefineCustomAgent[S](g, name, r.run(g), opts...)
	return func(run *ent.AiAgentRun) rez.AiAgentRunInvoker {
		return newAgentSessionInvoker(run, r, agent)
	}
}

type agentSessionInvoker[S rezai.AgentState] struct {
	sessionId    string
	sessionInput []byte
	runner       agentRunner[S]
	agent        *aix.Agent[S]
}

func newAgentSessionInvoker[S rezai.AgentState](run *ent.AiAgentRun, runner agentRunner[S], agent *aix.Agent[S]) *agentSessionInvoker[S] {
	return &agentSessionInvoker[S]{
		sessionId:    run.ID.String(),
		sessionInput: run.Input,
		agent:        agent,
		runner:       runner,
	}
}

type initialAgentState[S rezai.AgentState] struct {
	message  *ai.Message
	snapshot *aix.SessionSnapshot[S]
}

func (i *agentSessionInvoker[S]) createInitialState(ctx context.Context) (*initialAgentState[S], error) {
	msg, state, stateErr := i.runner.makeInitialState(i.sessionInput)
	if stateErr != nil {
		return nil, fmt.Errorf("initial state: %w", stateErr)
	}
	state.SessionID = i.sessionId

	createSnapshotFn := func(_ *aix.SessionSnapshot[S]) (*aix.SessionSnapshot[S], error) {
		return &aix.SessionSnapshot[S]{
			SessionID:    state.SessionID,
			FinishReason: aix.AgentFinishReasonStop,
			Status:       aix.SnapshotStatusCompleted,
			State:        state,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}, nil
	}
	snapshot, snapshotErr := i.agent.Store().SaveSnapshot(ctx, "", createSnapshotFn)
	if snapshotErr != nil {
		return nil, fmt.Errorf("initial snapshot: %w", snapshotErr)
	}
	return &initialAgentState[S]{message: msg, snapshot: snapshot}, nil
}

func (i *agentSessionInvoker[S]) getInvokeOpts(ctx context.Context, parentId *uuid.UUID) ([]aix.InvocationOption[S], error) {
	var snapshotId string
	if parentId != nil {
		snapshotId = parentId.String()
	} else {
		parent, parentErr := i.agent.Store().GetLatestSnapshot(ctx, i.sessionId)
		if parentErr != nil {
			return nil, fmt.Errorf("get parent snapshot: %w", parentErr)
		}
		snapshotId = parent.SnapshotID
	}
	invokeOpts := []aix.InvocationOption[S]{
		aix.WithSessionID[S](i.sessionId),
		aix.WithSnapshotID[S](snapshotId),
	}
	return invokeOpts, nil
}

func (i *agentSessionInvoker[S]) Start(ctx context.Context) (uuid.UUID, error) {
	initialState, initialStateErr := i.createInitialState(ctx)
	if initialStateErr != nil {
		return uuid.Nil, fmt.Errorf("initial state: %w", initialStateErr)
	}
	invokeOpts := []aix.InvocationOption[S]{aix.WithSessionID[S](i.sessionId)}
	if initialState.snapshot != nil {
		invokeOpts = append(invokeOpts, aix.WithSnapshotID[S](initialState.snapshot.SnapshotID))
	}
	input := &aix.AgentInput{Message: initialState.message}
	out, outErr := i.agent.Run(ctx, input, invokeOpts...)
	if outErr != nil {
		return uuid.Nil, fmt.Errorf("output: %w", outErr)
	}
	return uuid.Parse(out.SnapshotID)
}

func (i *agentSessionInvoker[S]) SendMessage(ctx context.Context, params rez.SendAgentRunMessageParams) (uuid.UUID, error) {
	if params.Message == nil {
		return uuid.Nil, fmt.Errorf("invalid message")
	}
	invokeOpts, invokeOptsErr := i.getInvokeOpts(ctx, params.ParentSnapshotID)
	if invokeOptsErr != nil {
		return uuid.Nil, fmt.Errorf("invoke opts: %w", invokeOptsErr)
	}
	out, outErr := i.agent.Run(ctx, &aix.AgentInput{Message: params.Message}, invokeOpts...)
	if outErr != nil {
		return uuid.Nil, fmt.Errorf("output: %w", outErr)
	}
	return uuid.Parse(out.SnapshotID)
}

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
