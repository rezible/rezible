package genkit

import (
	"context"
	"fmt"
	"slices"

	"github.com/firebase/genkit/go/ai"
	gkapi "github.com/firebase/genkit/go/core/api"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"google.golang.org/genai"
)

type AiService struct {
	cfg rez.AiConfig

	toolRefs []ai.ToolRef
	gk       *genkit.Genkit

	agentWrappers map[string]AgentWrapper
}

func NewAiService(cfg rez.Config) *AiService {
	return &AiService{
		cfg:           cfg.AI,
		toolRefs:      make([]ai.ToolRef, 0),
		agentWrappers: make(map[string]AgentWrapper),
	}
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
	)

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

func (s *AiService) GetAgents() []rez.AiAgentConfig {
	var agents []rez.AiAgentConfig
	for _, wrapper := range s.agentWrappers {
		agents = append(agents, wrapper.AgentConfig())
	}
	return agents
}

func (s *AiService) getAgentWrapper(name string) (AgentWrapper, error) {
	if wrapper, ok := s.agentWrappers[name]; ok {
		return wrapper, nil
	}
	return nil, fmt.Errorf("agent %q not found", name)
}

func (s *AiService) ValidateAgentSessionInput(name string, input []byte) (rez.AiAgentSessionInput, error) {
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

var flashModel = googlegenai.ModelRef("googleai/gemini-flash-latest", &genai.GenerateContentConfig{
	ThinkingConfig: &genai.ThinkingConfig{ThinkingLevel: genai.ThinkingLevelMinimal},
})

func (s *AiService) getModel(name string) ai.ModelRef { return flashModel }
func (s *AiService) getDefaultModel() ai.ModelRef     { return flashModel }
