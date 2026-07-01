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

func NewAgentRegistry(ctx context.Context, cfg rez.Config) (*AgentRegistry, error) {
	r := &AgentRegistry{
		cfg:    cfg.AI,
		agents: make(map[string]rez.WorkflowAgent),
		gk: genkit.Init(ctx,
			genkit.WithPlugins(&googlegenai.GoogleAI{}),
			genkit.WithExperimental(),
		),
	}

	for _, a := range r.makeAgents() {
		r.agents[a.WorkflowName()] = a
	}
	return r, nil
}

func (r *AgentRegistry) GetWorkflowAgent(workflowName string) (rez.WorkflowAgent, bool) {
	a, ok := r.agents[workflowName]
	return a, ok
}

func (r *AgentRegistry) makeAgents() []rez.WorkflowAgent {
	return []rez.WorkflowAgent{
		makeWorkflowAgentWrapper(r, newAlertInvestigationAgent(nil)),
	}
}

type workflowAgent[S any, O any] interface {
	makeInitialMessage(*ent.AgentTask) (*ai.Message, error)
	workflow() agents.Workflow[S, O]
	agentFunc() aix.AgentFunc[S]
}

type stateTransformingWorkflowAgent[S any] interface {
	transformState(context.Context, *aix.SessionState[S]) (*aix.SessionState[S], error)
}

type streamChunkTransformingWorkflowAgent[S any] interface {
	transformStreamChunk(context.Context, *aix.AgentStreamChunk) (*aix.AgentStreamChunk, error)
}

func makeWorkflowAgentWrapper[S any, O any](r *AgentRegistry, wa workflowAgent[S, O]) rez.WorkflowAgent {
	return &agentWrapper[S, O]{wa: wa, agent: defineAgent(r, wa)}
}

func defineAgent[S any, O any](r *AgentRegistry, wa workflowAgent[S, O]) *aix.Agent[S] {
	s := makeAgentSessionStore[S](nil)
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
	return genkitx.DefineCustomAgent[S](r.gk, w.Name(), wa.agentFunc(), opts...)
}

type agentWrapper[S any, O any] struct {
	wa    workflowAgent[S, O]
	agent *aix.Agent[S]
}

func (w *agentWrapper[S, O]) WorkflowName() string {
	return w.wa.workflow().Name()
}

func (w *agentWrapper[S, O]) RunTask(ctx context.Context, task *ent.AgentTask) (string, error) {
	if task == nil {
		return "", fmt.Errorf("invalid task input")
	}
	msg, msgErr := w.wa.makeInitialMessage(task)
	if msgErr != nil {
		return "", fmt.Errorf("failed to make initial message: %w", msgErr)
	}

	input := &aix.AgentInput{
		Detach:  true,
		Message: msg,
	}
	out, runErr := w.agent.Run(ctx, input, aix.WithSessionID[S](task.ID.String()))
	if runErr != nil {
		return "", runErr
	}

	fmt.Printf("output: %+v\n", out)
	return out.SnapshotID, nil
}

func (w *agentWrapper[S, O]) GetStatus(ctx context.Context, taskId uuid.UUID) (string, error) {
	snapshot, snapshotErr := w.agent.GetLatestSnapshot(ctx, taskId.String())
	if snapshotErr != nil {
		return "", snapshotErr
	}
	return string(snapshot.Status), nil
}
