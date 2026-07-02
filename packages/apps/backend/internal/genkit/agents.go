package genkit

import (
	"context"
	"fmt"
	"time"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/firebase/genkit/go/genkit"
	genkitx "github.com/firebase/genkit/go/genkit/exp"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type (
	agent[S any] interface {
		makeInitialState(*ent.AgentRun) (*aix.SessionState[S], error)
		workflowName() string
		agentFunc(*genkit.Genkit) aix.AgentFunc[S]
	}

	agentWithStateTransformer[S any] interface {
		transformState(context.Context, *aix.SessionState[S]) (*aix.SessionState[S], error)
	}

	agentWithStreamChunkTransformer[S any] interface {
		transformStreamChunk(context.Context, *aix.AgentStreamChunk) (*aix.AgentStreamChunk, error)
	}
)

func wrapAgent[S any](g *genkit.Genkit, s aix.SessionStore[S], a agent[S]) rez.Agent {
	workflow := a.workflowName()
	agentFunc := a.agentFunc(g)
	agentOpts := []aix.AgentOption[S]{
		aix.WithSessionStore[S](s),
		aix.WithDescription[S]("Runner for workflow " + workflow),
	}
	if stwa, ok := a.(agentWithStateTransformer[S]); ok {
		agentOpts = append(agentOpts, aix.WithStateTransform(stwa.transformState))
	}
	if sctwa, ok := a.(agentWithStreamChunkTransformer[S]); ok {
		agentOpts = append(agentOpts, aix.WithStreamTransform[S](sctwa.transformStreamChunk))
	}
	return &wrappedAgent[S]{
		workflow:           workflow,
		makeInitialStateFn: a.makeInitialState,
		agent:              genkitx.DefineCustomAgent[S](g, workflow, agentFunc, agentOpts...),
	}
}

type wrappedAgent[S any] struct {
	workflow           string
	makeInitialStateFn func(*ent.AgentRun) (*aix.SessionState[S], error)
	agent              *aix.Agent[S]
}

func (w *wrappedAgent[S]) Workflow() string {
	return w.workflow
}

func (w *wrappedAgent[S]) createInitialSnapshot(ctx context.Context, run *ent.AgentRun) (uuid.UUID, error) {
	if run == nil {
		return uuid.Nil, fmt.Errorf("nil run")
	}
	createSnapshotFn := func(_ *aix.SessionSnapshot[S]) (*aix.SessionSnapshot[S], error) {
		state, stateErr := w.makeInitialStateFn(run)
		if stateErr != nil {
			return nil, fmt.Errorf("initial state: %w", stateErr)
		}
		return &aix.SessionSnapshot[S]{
			SessionID:    run.ID.String(),
			FinishReason: aix.AgentFinishReasonStop,
			Status:       aix.SnapshotStatusCompleted,
			State:        state,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}, nil
	}
	snapshot, snapshotErr := w.agent.Store().SaveSnapshot(ctx, "", createSnapshotFn)
	if snapshotErr != nil {
		return uuid.Nil, fmt.Errorf("initial snapshot: %w", snapshotErr)
	}
	return uuid.Parse(snapshot.SnapshotID)
}

func (w *wrappedAgent[S]) getInvokeOpts(ctx context.Context, runId uuid.UUID, parentId *uuid.UUID) ([]aix.InvocationOption[S], error) {
	var parent *aix.SessionSnapshot[S]
	var parentErr error
	if parentId != nil {
		parent, parentErr = w.agent.Store().GetSnapshot(ctx, parentId.String())
	} else {
		parent, parentErr = w.agent.Store().GetLatestSnapshot(ctx, runId.String())
	}
	if parentErr != nil {
		return nil, fmt.Errorf("get parent snapshot: %w", parentErr)
	}
	snapshotId, parseErr := uuid.Parse(parent.SnapshotID)
	if parseErr != nil {
		return nil, fmt.Errorf("invalid parent snapshot id: %w", parseErr)
	}
	invokeOpts := []aix.InvocationOption[S]{
		aix.WithSessionID[S](runId.String()),
		aix.WithSnapshotID[S](snapshotId.String()),
	}
	return invokeOpts, nil
}

func (w *wrappedAgent[S]) Start(ctx context.Context, run *ent.AgentRun, msg *ai.Message) (uuid.UUID, error) {
	if run == nil {
		return uuid.Nil, fmt.Errorf("invalid run")
	}
	parentId, parentErr := w.createInitialSnapshot(ctx, run)
	if parentErr != nil {
		return uuid.Nil, fmt.Errorf("make initial state: %w", parentErr)
	}
	invokeOpts := []aix.InvocationOption[S]{
		aix.WithSessionID[S](run.ID.String()),
		aix.WithSnapshotID[S](parentId.String()),
	}
	input := &aix.AgentInput{Message: msg}
	out, outErr := w.agent.Run(ctx, input, invokeOpts...)
	if outErr != nil {
		return uuid.Nil, fmt.Errorf("output: %w", outErr)
	}
	return uuid.Parse(out.SnapshotID)
}

func (w *wrappedAgent[S]) SendMessage(ctx context.Context, run *ent.AgentRun, params rez.SendAgentRunMessageParams) (uuid.UUID, error) {
	if run == nil {
		return uuid.Nil, fmt.Errorf("invalid run")
	} else if params.Message == nil {
		return uuid.Nil, fmt.Errorf("invalid message")
	}
	invokeOpts, invokeOptsErr := w.getInvokeOpts(ctx, run.ID, params.ParentSnapshotID)
	if invokeOptsErr != nil {
		return uuid.Nil, fmt.Errorf("invoke opts: %w", invokeOptsErr)
	}
	out, outErr := w.agent.Run(ctx, &aix.AgentInput{Message: params.Message}, invokeOpts...)
	if outErr != nil {
		return uuid.Nil, fmt.Errorf("output: %w", outErr)
	}
	return uuid.Parse(out.SnapshotID)
}

func (w *wrappedAgent[S]) Resume(ctx context.Context, run *ent.AgentRun, params rez.ResumeAgentRunParams) (uuid.UUID, error) {
	if run == nil {
		return uuid.Nil, fmt.Errorf("invalid run")
	} else if (len(params.Respond) + len(params.Restart)) == 0 {
		return uuid.Nil, fmt.Errorf("invalid resume params")
	}
	invokeOpts, invokeOptsErr := w.getInvokeOpts(ctx, run.ID, params.ParentSnapshotID)
	if invokeOptsErr != nil {
		return uuid.Nil, fmt.Errorf("invoke opts: %w", invokeOptsErr)
	}
	conn, connErr := w.agent.Connect(ctx, invokeOpts...)
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
