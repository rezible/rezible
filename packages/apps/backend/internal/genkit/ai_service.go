package genkit

import (
	"cmp"
	"context"
	"fmt"
	"slices"

	"github.com/firebase/genkit/go/ai"
	gkapi "github.com/firebase/genkit/go/core/api"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/evaluators"
	"github.com/firebase/genkit/go/plugins/googlegenai"

	rez "github.com/rezible/rezible"
)

type AiRuntime struct {
	cfg rez.AiConfig

	toolRefs []ai.ToolRef
	gk       *genkit.Genkit

	catalogue *agentCatalogue
}

func NewAiRuntime(cfg rez.Config) *AiRuntime {
	return &AiRuntime{
		cfg:       cfg.AI,
		toolRefs:  make([]ai.ToolRef, 0),
		catalogue: newAgentCatalogue(),
	}
}

func (s *AiRuntime) Init(ctx context.Context, opts ...AiRuntimeOption) error {
	plugins, pluginsErr := s.makePlugins()
	if pluginsErr != nil {
		return fmt.Errorf("plugins: %w", pluginsErr)
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

func (s *AiRuntime) makePlugins() ([]gkapi.Plugin, error) {
	var plugins []gkapi.Plugin
	if geminiCfg := s.cfg.Gemini; geminiCfg.Enabled {
		plugins = append(plugins, &googlegenai.GoogleAI{APIKey: geminiCfg.APIKey})
	}

	if IsDevMode() {
		plugins = append(plugins, &evaluators.GenkitEval{
			Metrics: []evaluators.MetricConfig{{MetricType: evaluators.EvaluatorDeepEqual}},
		})
	}

	return plugins, nil
}

func IsDevMode() bool {
	return gkapi.CurrentEnvironment() == gkapi.EnvironmentDev
}

func (s *AiRuntime) applyOptions(opts []AiRuntimeOption) error {
	slices.SortFunc(opts, func(a, b AiRuntimeOption) int {
		return cmp.Compare(a.kind, b.kind)
	})
	for _, opt := range opts {
		if optErr := opt.optFn(s); optErr != nil {
			return fmt.Errorf("service init option: %w", optErr)
		}
	}
	return nil
}

type AiRuntimeOptionKind int

const (
	AiRuntimeOptionKindModel AiRuntimeOptionKind = iota
	AiRuntimeOptionKindTool
	AiRuntimeOptionKindAgent
)

type AiRuntimeOption struct {
	kind  AiRuntimeOptionKind
	optFn func(*AiRuntime) error
}

func (s *AiRuntime) AgentCatalogue() rez.AiAgentCatalogue {
	return s.catalogue
}

func (s *AiRuntime) InvokeAgentTurn(ctx context.Context, params rez.InvokeAiAgentTurnParams) (*rez.AiAgentInvocationResult, error) {
	if params.Session == nil {
		return nil, fmt.Errorf("agent session is required")
	}
	agent, agentErr := s.catalogue.getAgent(params.Session.AgentName)
	if agentErr != nil {
		return nil, agentErr
	}
	return agent.Invoke(ctx, params)
}
