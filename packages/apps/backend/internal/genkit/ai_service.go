package genkit

import (
	"context"
	"fmt"

	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	rezai "github.com/rezible/rezible/pkg/ai"
)

type AiService struct {
	cfg              rez.AiConfig
	gk               *genkit.Genkit
	snapshots        rez.AiStateService
	agentInvokers    map[string]AgentRunInvokerFunc
	workflowInvokers map[string]WorkflowInvokerFunc
}

func NewAiService(ctx context.Context, cfg rez.Config, snapshots rez.AiStateService) *AiService {
	gkOpts := []genkit.GenkitOption{
		genkit.WithExperimental(),
	}
	if cfg.AI.Gemini.Enabled {
		gkOpts = append(gkOpts, genkit.WithPlugins(&googlegenai.GoogleAI{
			APIKey: cfg.AI.Gemini.APIKey,
		}))
	}
	return &AiService{
		cfg:              cfg.AI,
		gk:               genkit.Init(ctx, gkOpts...),
		snapshots:        snapshots,
		agentInvokers:    make(map[string]AgentRunInvokerFunc),
		workflowInvokers: make(map[string]WorkflowInvokerFunc),
	}
}

func RegisterAgent[S rezai.AgentState](s *AiService, a agentRunner[S]) {
	s.agentInvokers[a.definition().Name] = makeAgentSessionInvokerFunc(s.gk, makeAgentSessionStore[S](s.snapshots), a)
}

func (s *AiService) GetAgentRunInvoker(run *ent.AiAgentRun) (rez.AiAgentRunInvoker, error) {
	if run == nil {
		return nil, fmt.Errorf("nil run")
	}
	invokerFn, ok := s.agentInvokers[run.AgentName]
	if !ok {
		return nil, fmt.Errorf("agent %s not found", run.AgentName)
	}
	return invokerFn(run), nil
}

func RegisterWorkflow[I rezai.WorkflowInput, S any, O rezai.WorkflowOutput](s *AiService, w workflowRunner[I, S, O]) {

}

func (s *AiService) GetWorkflowInvoker(name string) (rez.AiWorkflowInvoker, error) {
	invokerFn, ok := s.workflowInvokers[name]
	if !ok {
		return nil, fmt.Errorf("workflow %s not found", name)
	}
	return invokerFn(), nil
}

//func (s *AiSessionService) GetWorkflow(name string) (rez.AiAgentInvoker, bool) {
//
//}
