package genkit

import (
	"context"
	"fmt"

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
	cfg    rez.AiConfig
	gk     *genkit.Genkit
	agents map[string]rez.WorkflowAgent
}

func NewAgentRegistry(ctx context.Context, cfg rez.Config) *AgentRegistry {
	gkOpts := []genkit.GenkitOption{
		genkit.WithExperimental(),
	}
	if cfg.AI.Gemini.Enabled {
		gkOpts = append(gkOpts, genkit.WithPlugins(&googlegenai.GoogleAI{
			APIKey: cfg.AI.Gemini.APIKey,
		}))
	}
	return &AgentRegistry{
		cfg:    cfg.AI,
		gk:     genkit.Init(ctx, gkOpts...),
		agents: make(map[string]rez.WorkflowAgent),
	}
}

func (r *AgentRegistry) Register(a rez.WorkflowAgent) {
	r.agents[a.WorkflowName()] = a
}

func (r *AgentRegistry) RegisterMultiple(snaps rez.AgentRunSnapshotService, was ...workflowAgent[any, any]) {
	for _, wa := range was {
		a := WrapWorkflowAgent(r, snaps, wa)
		r.agents[a.WorkflowName()] = a
	}
}

func (r *AgentRegistry) Get(workflow string) (rez.WorkflowAgent, bool) {
	a, ok := r.agents[workflow]
	return a, ok
}

type workflowAgent[S any, O any] interface {
	validateInput([]byte) error
	makeInitial(*ent.AgentRun) (*ai.Message, *S, error)
	workflow() agents.Workflow[S, O]
	agentFunc(*genkit.Genkit) aix.AgentFunc[S]
}

type stateTransformingWorkflowAgent[S any] interface {
	transformState(context.Context, *aix.SessionState[S]) (*aix.SessionState[S], error)
}

type streamChunkTransformingWorkflowAgent[S any] interface {
	transformStreamChunk(context.Context, *aix.AgentStreamChunk) (*aix.AgentStreamChunk, error)
}

func WrapWorkflowAgent[S any, O any](r *AgentRegistry, snaps rez.AgentRunSnapshotService, wa workflowAgent[S, O]) rez.WorkflowAgent {
	return wrapWorkflowAgent(r.gk, makeAgentSessionStore[S](snaps), wa)
}

func wrapWorkflowAgent[S any, O any](g *genkit.Genkit, s aix.SessionStore[S], wa workflowAgent[S, O]) rez.WorkflowAgent {
	w := wa.workflow()
	opts := []aix.AgentOption[S]{
		aix.WithSessionStore[S](s),
		aix.WithDescription[S](w.Description()),
	}
	if stwa, ok := wa.(stateTransformingWorkflowAgent[S]); ok {
		opts = append(opts, aix.WithStateTransform(stwa.transformState))
	}
	if sctwa, ok := wa.(streamChunkTransformingWorkflowAgent[S]); ok {
		opts = append(opts, aix.WithStreamTransform[S](sctwa.transformStreamChunk))
	}

	return &agentWrapper[S, O]{
		wfAgent: wa,
		agent:   genkitx.DefineCustomAgent[S](g, w.Name(), wa.agentFunc(g), opts...),
	}
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

func (w *agentWrapper[S, O]) Run(ctx context.Context, run *ent.AgentRun, opts *rez.RunAgentOpts) (uuid.UUID, error) {
	if run == nil {
		return uuid.Nil, fmt.Errorf("invalid task input")
	}

	msg, _, msgErr := w.wfAgent.makeInitial(run)
	if msgErr != nil {
		return uuid.Nil, fmt.Errorf("failed to make initial message: %w", msgErr)
	}

	input := &aix.AgentInput{Message: msg}
	if opts != nil {
		input.Detach = opts.Detach
	}
	out, runErr := w.agent.Run(ctx, input, aix.WithSessionID[S](run.ID.String()))
	if runErr != nil {
		return uuid.Nil, runErr
	}

	snapshotId, idErr := uuid.Parse(out.SnapshotID)
	if idErr != nil {
		return uuid.Nil, fmt.Errorf("invalid snapshot id: %w", idErr)
	}

	return snapshotId, nil
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
