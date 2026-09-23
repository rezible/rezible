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
	"github.com/rezible/rezible/pkg/execution"
	"google.golang.org/genai"
)

type AiRuntime struct {
	cfg       rez.AiConfig
	catalogue *agentCatalogue
	gk        *genkit.Genkit
}

func NewAiRuntime(cfg rez.Config) *AiRuntime {
	return &AiRuntime{
		cfg:       cfg.AI,
		catalogue: newAgentCatalogue(),
	}
}

func (r *AiRuntime) Init(ctx context.Context, opts ...AiRuntimeOption) error {
	var plugins []gkapi.Plugin
	var defaultModelName string
	for _, opt := range opts {
		if len(opt.plugins) > 0 {
			plugins = append(plugins, opt.plugins...)
		}
		if opt.defaultModelName != nil {
			fmt.Printf("opt: %+v\n", opt)
			if defaultModelName != "" {
				return fmt.Errorf("multiple default models set")
			}
			defaultModelName = *opt.defaultModelName
		}
	}

	gkOpts := []genkit.GenkitOption{
		genkit.WithExperimental(),
		genkit.WithPlugins(plugins...),
	}
	if defaultModelName != "" {
		gkOpts = append(gkOpts, genkit.WithDefaultModel(defaultModelName))
	}

	r.gk = genkit.Init(ctx, gkOpts...)

	if optsErr := r.applyOptions(opts); optsErr != nil {
		return fmt.Errorf("apply service options: %w", optsErr)
	}

	return nil
}

func (r *AiRuntime) applyOptions(opts []AiRuntimeOption) error {
	slices.SortFunc(opts, func(a, b AiRuntimeOption) int {
		return cmp.Compare(a.kind, b.kind)
	})
	for _, opt := range opts {
		if opt.runtimeFn == nil {
			continue
		}
		if optErr := opt.runtimeFn(r); optErr != nil {
			return fmt.Errorf("service init option: %w", optErr)
		}
	}
	return nil
}

func (r *AiRuntime) AgentCatalogue() rez.AiAgentCatalogue {
	return r.catalogue
}

func (r *AiRuntime) InvokeAgentTurn(ctx context.Context, params rez.InvokeAiAgentTurnParams) (*rez.AiAgentInvocationResult, error) {
	if params.Session == nil {
		return nil, fmt.Errorf("agent session is required")
	}
	agent, agentErr := r.catalogue.getAgent(params.Session.AgentName)
	if agentErr != nil {
		return nil, agentErr
	}
	return agent.Invoke(execution.NewAiAgentContext(ctx, params.Session, params.Turn), params)
}

type AiRuntimeOption struct {
	kind             AiRuntimeOptionKind
	runtimeFn        func(*AiRuntime) error
	defaultModelName *string
	plugins          []gkapi.Plugin
}

// TODO: remove these, don't use option kinds (just different functions eg `opt.agentFn`)
type AiRuntimeOptionKind int

const (
	AiRuntimeOptionKindPlugin AiRuntimeOptionKind = iota
	AiRuntimeOptionKindModel  AiRuntimeOptionKind = iota
	AiRuntimeOptionKindTool   AiRuntimeOptionKind = iota
	AiRuntimeOptionKindAgent  AiRuntimeOptionKind = iota
)

type ModelDefinition[Config any] struct {
	Name      string
	IsDefault bool
	opts      *ai.ModelOptions
	fn        ai.ModelActionFunc[Config]
}

func WithDefinedModel[Config any](def ModelDefinition[Config]) AiRuntimeOption {
	opt := AiRuntimeOption{
		kind: AiRuntimeOptionKindModel,
		runtimeFn: func(s *AiRuntime) error {
			genkit.DefineModelAction(s.gk, def.Name, def.opts, def.fn)
			return nil
		},
	}
	if def.IsDefault {
		opt.defaultModelName = &def.Name
	}
	return opt
}

var geminiFlashModel = googlegenai.ModelRef("googleai/gemini-flash-latest", &genai.GenerateContentConfig{
	ThinkingConfig: &genai.ThinkingConfig{ThinkingLevel: genai.ThinkingLevelMinimal},
})

func WithGeminiPlugin(cfg rez.AiProviderConfigGemini) AiRuntimeOption {
	opt := AiRuntimeOption{kind: AiRuntimeOptionKindPlugin}
	if cfg.Enabled {
		opt.plugins = append(opt.plugins, &googlegenai.GoogleAI{APIKey: cfg.APIKey})
	}
	return opt
}
