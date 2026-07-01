package genkit

import (
	"context"
	"fmt"
	"time"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/firebase/genkit/go/genkit"
	genkitx "github.com/firebase/genkit/go/genkit/exp"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/pkg/agents"
)

type AgentRegistry struct {
	cfg       rez.AiConfig
	gk        *genkit.Genkit
	snapshots rez.AgentRunSnapshotService
	agents    map[string]rez.WorkflowAgent
}

func NewAgentRegistry(ctx context.Context, cfg rez.Config, snapshots rez.AgentRunSnapshotService) *AgentRegistry {
	gkOpts := []genkit.GenkitOption{
		genkit.WithExperimental(),
	}
	if cfg.AI.Gemini.Enabled {
		gkOpts = append(gkOpts, genkit.WithPlugins(&googlegenai.GoogleAI{
			APIKey: cfg.AI.Gemini.APIKey,
		}))
	}
	return &AgentRegistry{
		cfg:       cfg.AI,
		gk:        genkit.Init(ctx, gkOpts...),
		snapshots: snapshots,
		agents:    make(map[string]rez.WorkflowAgent),
	}
}

func (r *AgentRegistry) Register(a rez.WorkflowAgent) {
	r.agents[a.WorkflowName()] = a
}

func RegisterWorkflowAgent[S any, O any](r *AgentRegistry, wa workflowAgent[S, O]) {
	r.Register(wrapWorkflowAgent(r.gk, makeAgentSessionStore[S](r.snapshots), wa))
}

func (r *AgentRegistry) RegisterMultiple(was ...workflowAgent[any, any]) {
	for _, wa := range was {
		RegisterWorkflowAgent(r, wa)
	}
}

func (r *AgentRegistry) Get(workflow string) (rez.WorkflowAgent, bool) {
	a, ok := r.agents[workflow]
	return a, ok
}

type workflowAgent[S any, O any] interface {
	validateInput([]byte) error
	makeInitialState(*ent.AgentRun) (*aix.SessionState[S], error)
	workflow() agents.Workflow[S, O]
	agentFunc(*genkit.Genkit) aix.AgentFunc[S]
}

type stateTransformingWorkflowAgent[S any] interface {
	transformState(context.Context, *aix.SessionState[S]) (*aix.SessionState[S], error)
}

type streamChunkTransformingWorkflowAgent[S any] interface {
	transformStreamChunk(context.Context, *aix.AgentStreamChunk) (*aix.AgentStreamChunk, error)
}

func wrapWorkflowAgent[S any, O any](g *genkit.Genkit, s aix.SessionStore[S], wa workflowAgent[S, O]) rez.WorkflowAgent {
	w := wa.workflow()
	agentOpts := []aix.AgentOption[S]{
		aix.WithSessionStore[S](s),
		aix.WithDescription[S](w.Description()),
	}
	if stwa, ok := wa.(stateTransformingWorkflowAgent[S]); ok {
		agentOpts = append(agentOpts, aix.WithStateTransform(stwa.transformState))
	}
	if sctwa, ok := wa.(streamChunkTransformingWorkflowAgent[S]); ok {
		agentOpts = append(agentOpts, aix.WithStreamTransform[S](sctwa.transformStreamChunk))
	}
	agent := genkitx.DefineCustomAgent[S](g, w.Name(), wa.agentFunc(g), agentOpts...)
	return &agentWrapper[S, O]{wfAgent: wa, agent: agent}
}

type agentWrapper[S any, O any] struct {
	wfAgent workflowAgent[S, O]
	agent   *aix.Agent[S]
}

func (w *agentWrapper[S, O]) WorkflowName() string {
	return w.wfAgent.workflow().Name()
}

func (w *agentWrapper[S, O]) ValidateInput(input []byte) error {
	return w.wfAgent.validateInput(input)
}

func (w *agentWrapper[S, O]) createInitialSnapshot(ctx context.Context, run *ent.AgentRun) (*aix.SessionSnapshot[S], error) {
	init, initErr := w.wfAgent.makeInitialState(run)
	if initErr != nil {
		return nil, fmt.Errorf("initial state: %w", initErr)
	}
	saveSnapshotFn := func(_ *aix.SessionSnapshot[S]) (*aix.SessionSnapshot[S], error) {
		snap := &aix.SessionSnapshot[S]{
			SessionID: run.ID.String(),
			//SnapshotID:   snapshotId.String(),
			FinishReason: aix.AgentFinishReasonStop,
			Status:       aix.SnapshotStatusCompleted,
			State:        init,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		return snap, nil
	}
	return w.agent.Store().SaveSnapshot(ctx, "", saveSnapshotFn)
}

func (w *agentWrapper[S, O]) getInvokeOpts(ctx context.Context, run *ent.AgentRun) ([]aix.InvocationOption[S], error) {
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

func (w *agentWrapper[S, O]) Invoke(ctx context.Context, run *ent.AgentRun, msg *ai.Message) (uuid.UUID, error) {
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

func (w *agentWrapper[S, O]) GetStatus(ctx context.Context, taskId uuid.UUID) error {
	snapshot, snapshotErr := w.agent.GetLatestSnapshot(ctx, taskId.String())
	if snapshotErr != nil {
		return snapshotErr
	}
	status := string(snapshot.Status)
	fmt.Printf("get latest snapshot status: %s\n", status)
	return nil
}
