package genkit

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/firebase/genkit/go/ai"
	gkapi "github.com/firebase/genkit/go/core/api"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	rezai "github.com/rezible/rezible/pkg/ai"
)

type AiService struct {
	cfg rez.AiConfig

	sessions       rez.AiAgentSnapshotService
	knowledgeGraph rez.KnowledgeGraphService

	toolRefs []ai.ToolRef
	gk       *genkit.Genkit

	agentWrappers    map[string]*agentWrapper
	workflowInvokers map[string]WorkflowInvokerFunc
}

func NewAiService(cfg rez.Config, sessions rez.AiAgentSnapshotService, kg rez.KnowledgeGraphService) *AiService {
	s := &AiService{
		cfg: cfg.AI,

		sessions:       sessions,
		knowledgeGraph: kg,

		toolRefs:         make([]ai.ToolRef, 0),
		agentWrappers:    make(map[string]*agentWrapper),
		workflowInvokers: make(map[string]WorkflowInvokerFunc),
	}

	return s
}

func (s *AiService) Init(ctx context.Context, opts ...AiServiceOption) error {
	var plugins []gkapi.Plugin

	if geminiCfg := s.cfg.Gemini; geminiCfg.Enabled {
		plugins = append(plugins, &googlegenai.GoogleAI{APIKey: geminiCfg.APIKey})
	}

	s.gk = genkit.Init(ctx,
		genkit.WithPlugins(plugins...),
		genkit.WithDefaultModel(flashModel.Name()),
		genkit.WithExperimental(),
		genkit.WithPromptFS(rezai.PromptsDir),
	)

	// apply tool registrations first
	slices.SortFunc(opts, func(a, b AiServiceOption) int {
		if a.kind == b.kind {
			return 0
		}
		if a.kind == "tool" {
			return -1
		}
		return 1
	})

	for _, o := range opts {
		slog.Debug("ai init opt", "kind", o.kind)
		if optErr := o.optFn(s); optErr != nil {
			return fmt.Errorf("service init option: %w", optErr)
		}
	}

	return nil
}

type AiServiceOption struct {
	kind  string
	optFn func(*AiService) error
}

func WithAgent[I rezai.AgentInput, S rezai.SessionState, O rezai.AgentOutput](r agentRunner[I, S, O]) AiServiceOption {
	return AiServiceOption{
		kind: "agent",
		optFn: func(s *AiService) error {
			wrapper, wrapperErr := wrapAgentRunner(s, r)
			if wrapperErr != nil || wrapper == nil {
				return fmt.Errorf("wrap runner: %w", wrapperErr)
			}
			s.agentWrappers[r.agentDefinition().Name] = wrapper
			return nil
		},
	}
}

func (s *AiService) getAgentWrapper(name string) (*agentWrapper, error) {
	wrapper, ok := s.agentWrappers[name]
	if !ok {
		return nil, fmt.Errorf("agent '%s' not found", name)
	}
	return wrapper, nil
}

func (s *AiService) ValidateAgentRunInput(name string, input []byte) error {
	wrapper, wrapperErr := s.getAgentWrapper(name)
	if wrapperErr != nil {
		return wrapperErr
	}
	return wrapper.ValidateInput(input)
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

func (s *AiService) getRegisteredTools(refs []ai.ToolRef) ([]ai.ToolRef, []ai.ToolRef) {
	toolMap := make(map[string]ai.ToolRef)
	for _, ref := range genkit.ListTools(s.gk) {
		toolMap[ref.Name()] = ref
	}
	var registered []ai.ToolRef
	var missing []ai.ToolRef
	for _, t := range refs {
		if ref, ok := toolMap[t.Name()]; ok {
			registered = append(registered, ref)
		} else {
			missing = append(missing, t)
		}
	}
	return registered, missing
}
