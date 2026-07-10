package genkit

import (
	"context"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	gkapi "github.com/firebase/genkit/go/core/api"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	rezai "github.com/rezible/rezible/pkg/ai"
)

type AiService struct {
	sessions       rez.AiSessionStateService
	knowledgeGraph rez.KnowledgeGraphService

	plugins  []gkapi.Plugin
	toolRefs []ai.ToolRef
	gk       *genkit.Genkit

	agentWrappers    map[string]*agentWrapper
	workflowInvokers map[string]WorkflowInvokerFunc
}

func NewAiService(cfg rez.Config, sessions rez.AiSessionStateService, kg rez.KnowledgeGraphService) *AiService {
	s := &AiService{
		sessions:       sessions,
		knowledgeGraph: kg,

		plugins:          make([]gkapi.Plugin, 0),
		toolRefs:         make([]ai.ToolRef, 0),
		agentWrappers:    make(map[string]*agentWrapper),
		workflowInvokers: make(map[string]WorkflowInvokerFunc),
	}

	if cfg.AI.Gemini.Enabled {
		googleAiPlugin := &googlegenai.GoogleAI{APIKey: cfg.AI.Gemini.APIKey}
		s.plugins = append(s.plugins, googleAiPlugin)
	}

	return s
}

func (s *AiService) Init(ctx context.Context, opts ...AiServiceOption) error {
	s.gk = genkit.Init(ctx,
		genkit.WithPlugins(s.plugins...),
		genkit.WithDefaultModel(flashModel.Name()),
		genkit.WithExperimental(),
		genkit.WithPromptFS(rezai.PromptsDir),
	)

	for _, opt := range opts {
		if optErr := opt(s); optErr != nil {
			return fmt.Errorf("service init option: %w", optErr)
		}
	}

	return nil
}

type AiServiceOption = func(*AiService) error

func WithAgent[I rezai.AgentInput, S rezai.SessionState, O rezai.AgentOutput](r agentRunner[I, S, O]) AiServiceOption {
	return func(s *AiService) error {
		wrapper, wrapperErr := wrapAgentRunner(s, r)
		if wrapperErr != nil || wrapper == nil {
			return fmt.Errorf("wrap runner: %w", wrapperErr)
		}
		s.agentWrappers[r.definition().Name] = wrapper
		return nil
	}
}

func (s *AiService) getAgentWrapper(name string) (*agentWrapper, error) {
	wrapper, ok := s.agentWrappers[name]
	if !ok {
		return nil, fmt.Errorf("agent '%s' not found", name)
	}
	return wrapper, nil
}

func (s *AiService) GetAgentRunner(run *ent.AiAgentRun) (rez.AiAgentInvoker, error) {
	if run == nil {
		return nil, fmt.Errorf("nil run")
	}
	wrapper, wrapperErr := s.getAgentWrapper(run.AgentName)
	if wrapperErr != nil {
		return nil, wrapperErr
	}
	return wrapper.MakeRunner(run), nil
}

func (s *AiService) ValidateAgentRunInput(name string, input []byte) error {
	wrapper, wrapperErr := s.getAgentWrapper(name)
	if wrapperErr != nil {
		return wrapperErr
	}
	return wrapper.ValidateInput(input)
}

//func RegisterWorkflow[I rezai.WorkflowInput, O rezai.WorkflowOutput, S rezai.SessionState](s *AiService, w workflowRunner[I, O, S]) {
//	def := w.definition()
//	if def.Prompt != nil {
//		//ai.DefineDataPrompt[I, O](s.gk, *def.Prompt)
//	}
//	s.workflowInvokers[w.definition().Name] = makeWorkflowInvokerFunc(s.gk, makeSessionStore[S](s.state), w)
//}

func (s *AiService) GetWorkflowInvoker(name string) (rez.AiWorkflowInvoker, error) {
	invokerFn, ok := s.workflowInvokers[name]
	if !ok {
		return nil, fmt.Errorf("workflow %s not found", name)
	}
	return invokerFn(), nil
}
