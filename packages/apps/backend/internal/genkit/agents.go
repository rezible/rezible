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

func (w *wrappedAgent[S]) createInitialSnapshot(ctx context.Context, run *ent.AgentRun) (*aix.SessionSnapshot[S], error) {
	state, stateErr := w.makeInitialStateFn(run)
	if stateErr != nil {
		return nil, fmt.Errorf("initial state: %w", stateErr)
	}
	return w.agent.Store().SaveSnapshot(ctx, "", func(_ *aix.SessionSnapshot[S]) (*aix.SessionSnapshot[S], error) {
		return &aix.SessionSnapshot[S]{
			SessionID:    run.ID.String(),
			FinishReason: aix.AgentFinishReasonStop,
			Status:       aix.SnapshotStatusCompleted,
			State:        state,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}, nil
	})
}

func (w *wrappedAgent[S]) getInvokeOpts(ctx context.Context, run *ent.AgentRun) ([]aix.InvocationOption[S], error) {
	if run == nil {
		return nil, fmt.Errorf("invalid task input")
	}

	parent, parentErr := w.agent.Store().GetLatestSnapshot(ctx, run.ID.String())
	if parentErr != nil {
		return nil, fmt.Errorf("get latest snapshot: %w", parentErr)
	} else if parent == nil {
		parent, parentErr = w.createInitialSnapshot(ctx, run)
		if parentErr != nil {
			return nil, fmt.Errorf("make initial state: %w", parentErr)
		}
	}
	parentId, parentIdErr := uuid.Parse(parent.SnapshotID)
	if parentIdErr != nil {
		return nil, fmt.Errorf("parse parent snapshot id: %w", parentIdErr)
	}

	invokeOpts := []aix.InvocationOption[S]{
		aix.WithSessionID[S](run.ID.String()),
		aix.WithSnapshotID[S](parentId.String()),
	}
	return invokeOpts, nil
}

func (w *wrappedAgent[S]) Invoke(ctx context.Context, run *ent.AgentRun, msg *ai.Message) (uuid.UUID, error) {
	invokeOpts, invokeOptsErr := w.getInvokeOpts(ctx, run)
	if invokeOptsErr != nil {
		return uuid.Nil, fmt.Errorf("invoke opts: %w", invokeOptsErr)
	}
	input := &aix.AgentInput{Message: msg}
	out, outErr := w.agent.Run(ctx, input, invokeOpts...)
	/*
		conn, connErr := w.agent.Connect(ctx, invokeOpts...)
		if connErr != nil {
			return nil, fmt.Errorf("connect: %w", connErr)
		}
		if invokeErr := conn.Send(input); invokeErr != nil && !errors.Is(invokeErr, core.ErrActionCompleted) {
			return nil, fmt.Errorf("invoke: %w", invokeErr)
		}
		out, outErr := conn.Output()
	*/
	if outErr != nil {
		return uuid.Nil, fmt.Errorf("output: %w", outErr)
	}
	return uuid.Parse(out.SnapshotID)
}

func (w *wrappedAgent[S]) GetStatus(ctx context.Context, taskId uuid.UUID) error {
	snapshot, snapshotErr := w.agent.GetLatestSnapshot(ctx, taskId.String())
	if snapshotErr != nil {
		return snapshotErr
	}
	status := string(snapshot.Status)
	fmt.Printf("get latest snapshot status: %s\n", status)
	return nil
}
