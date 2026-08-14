package genkit

import (
	"cmp"
	"context"
	"fmt"
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

	toolRefs []ai.ToolRef
	gk       *genkit.Genkit

	agentWrappers    map[string]rezai.AgentWrapper
	workflowWrappers map[string]rezai.WorkflowWrapper
}

func NewAiService(cfg rez.Config) *AiService {
	return &AiService{
		cfg:              cfg.AI,
		toolRefs:         make([]ai.ToolRef, 0),
		agentWrappers:    make(map[string]rezai.AgentWrapper),
		workflowWrappers: make(map[string]rezai.WorkflowWrapper),
	}
}

func (s *AiService) Init(ctx context.Context, opts ...AiServiceOption) error {
	var plugins []gkapi.Plugin
	if geminiCfg := s.cfg.Gemini; geminiCfg.Enabled {
		plugins = append(plugins, &googlegenai.GoogleAI{APIKey: geminiCfg.APIKey})
	}

	defaultModel := s.getDefaultModel()
	if defaultModel == nil {
		return fmt.Errorf("no default model found")
	}

	s.gk = genkit.Init(ctx,
		genkit.WithPlugins(plugins...),
		genkit.WithDefaultModel(defaultModel.Name()),
		genkit.WithExperimental(),
	)

	if optsErr := s.applyOptions(opts); optsErr != nil {
		return fmt.Errorf("apply service options: %w", optsErr)
	}

	return nil
}

func (s *AiService) applyOptions(opts []AiServiceOption) error {
	slices.SortFunc(opts, func(a, b AiServiceOption) int {
		return cmp.Compare(a.kind, b.kind)
	})
	for _, opt := range opts {
		if optErr := opt.optFn(s); optErr != nil {
			return fmt.Errorf("service init option: %w", optErr)
		}
	}
	return nil
}

type AiServiceOptionKind int

const (
	AiServiceOptionKindModel    AiServiceOptionKind = iota
	AiServiceOptionKindTool     AiServiceOptionKind = iota
	AiServiceOptionKindWorkflow AiServiceOptionKind = iota
	AiServiceOptionKindAgent    AiServiceOptionKind = iota
)

type AiServiceOption struct {
	kind  AiServiceOptionKind
	optFn func(*AiService) error
}

func (s *AiService) GetAgents() []rez.AiAgentConfig {
	var agents []rez.AiAgentConfig
	for _, wrapper := range s.agentWrappers {
		agents = append(agents, wrapper.Config())
	}
	return agents
}

func (s *AiService) getAgentWrapper(name string) (rezai.AgentWrapper, error) {
	if wrapper, ok := s.agentWrappers[name]; ok {
		return wrapper, nil
	}
	return nil, fmt.Errorf("agent %q not found", name)
}

func (s *AiService) ValidateAgentSessionInput(name string, input []byte) (rez.ValidatingInput, error) {
	wrapper, wrapperErr := s.getAgentWrapper(name)
	if wrapperErr != nil {
		return nil, wrapperErr
	}
	return wrapper.ValidateInput(input)
}

func (s *AiService) MakeInitialAgentTurnInput(ctx context.Context, sess *ent.AgentSession) (*rez.AiAgentTurnInput, error) {
	wrapper, wrapperErr := s.getAgentWrapper(sess.AgentName)
	if wrapperErr != nil {
		return nil, wrapperErr
	}
	return wrapper.MakeInitialTurnInput(ctx, sess)
}

func (s *AiService) InvokeAgentTurn(ctx context.Context, params rez.InvokeAgentTurnParams) (*rez.AiAgentInvocationResult, error) {
	if params.Session == nil {
		return nil, fmt.Errorf("agent session is required")
	}
	wrapper, wrapperErr := s.getAgentWrapper(params.Session.AgentName)
	if wrapperErr != nil {
		return nil, wrapperErr
	}
	return wrapper.Invoke(ctx, params)
}

func (s *AiService) GetWorkflowRunner(name string) (rez.AiWorkflowRunner, error) {
	wr, ok := s.workflowWrappers[name]
	if !ok {
		return nil, fmt.Errorf("workflow not found: %s", name)
	}
	return wr, nil
}
