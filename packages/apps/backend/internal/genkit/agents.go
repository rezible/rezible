package genkit

import (
	"context"
	"fmt"
	"log/slog"

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
	agents    map[string]rez.Agent
}

func NewAgentRegistry(ctx context.Context, cfg rez.Config, snapshots rez.AgentRunSnapshotService) *AgentRegistry {
	gkOpts := []genkit.GenkitOption{
		genkit.WithPlugins(&googlegenai.GoogleAI{}),
		genkit.WithExperimental(),
	}
	return &AgentRegistry{
		cfg:       cfg.AI,
		gk:        genkit.Init(ctx, gkOpts...),
		snapshots: snapshots,
		agents:    make(map[string]rez.Agent),
	}
}

func (r *AgentRegistry) RegisterAlertInvestigationAgent(alerts rez.AlertService) {
	a := makeWorkflowAgentWrapper(r, newAlertInvestigationAgent(alerts))
	r.agents[a.WorkflowName()] = a
}

func (r *AgentRegistry) GetAgent(workflowName string) (rez.Agent, bool) {
	a, ok := r.agents[workflowName]
	return a, ok
}

func (r *AgentRegistry) ValidateWorkflowInput(workflowName string, input []byte) error {
	agent, ok := r.agents[workflowName]
	if !ok {
		return fmt.Errorf("workflow %s not found", workflowName)
	}
	return agent.ValidateInput(input)
}

type workflowAgent[S any, O any] interface {
	validateInput([]byte) error
	makeInitialMessage(*ent.AgentRun) (*ai.Message, error)
	workflow() agents.Workflow[S, O]
	agentFunc() aix.AgentFunc[S]
}

type stateTransformingWorkflowAgent[S any] interface {
	transformState(context.Context, *aix.SessionState[S]) (*aix.SessionState[S], error)
}

type streamChunkTransformingWorkflowAgent[S any] interface {
	transformStreamChunk(context.Context, *aix.AgentStreamChunk) (*aix.AgentStreamChunk, error)
}

func makeWorkflowAgentWrapper[S any, O any](r *AgentRegistry, wa workflowAgent[S, O]) rez.Agent {
	s := makeAgentSessionStore[S](r.snapshots)
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
		agent:   genkitx.DefineCustomAgent[S](r.gk, w.Name(), wa.agentFunc(), opts...),
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

	msg, msgErr := w.wfAgent.makeInitialMessage(run)
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
		slog.ErrorContext(ctx, "failed to parse snapshot ID", "error", idErr)
	}

	fmt.Printf("run agent output: %+v\n", out)

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
