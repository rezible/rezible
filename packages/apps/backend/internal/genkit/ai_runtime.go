package genkit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/firebase/genkit/go/ai"
	gkapi "github.com/firebase/genkit/go/core/api"
	"github.com/firebase/genkit/go/core/tracing"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/anthropic"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"github.com/firebase/genkit/go/plugins/otelgenai"
	rez "github.com/rezible/rezible"
	rezai "github.com/rezible/rezible/pkg/ai"
)

type AiRuntime struct {
	cfg         rez.AiConfig
	extraModels []*ai.ModelAction
	gk          *genkit.Genkit
	// captureAiContent records prompts, tool results and provider data on Genkit's spans.
	captureAiContent bool
}

// NewAiRuntime creates a runtime serving the enabled providers' models and any extra models, which serve under
// their own names.
func NewAiRuntime(cfg rez.Config, extraModels ...*ai.ModelAction) *AiRuntime {
	return &AiRuntime{
		cfg:              cfg.AI,
		extraModels:      extraModels,
		captureAiContent: cfg.Telemetry.Tracing.CaptureAiContent,
	}
}

func (r *AiRuntime) Init(ctx context.Context) error {
	providers := enabledProviders(r.cfg.Providers)
	if r.cfg.Models.Default == "" {
		switch len(providers) {
		case 0:
			// Without a provider there is no model, and the AI features need none.
		case 1:
			r.cfg.Models.Default = providers[0].defaultModel
		default:
			return fmt.Errorf("ai.models.default is required when more than one provider is enabled")
		}
	}

	var plugins []gkapi.Plugin
	if len(r.extraModels) > 0 {
		plugins = append(plugins, extraModelsPlugin(r.extraModels))
	}
	for _, provider := range providers {
		plugins = append(plugins, provider.plugin(r.cfg.Models))
	}
	gkOpts := []genkit.GenkitOption{
		genkit.WithExperimental(),
		genkit.WithPlugins(plugins...),
	}
	if r.cfg.Models.Default != "" {
		gkOpts = append(gkOpts, genkit.WithDefaultModel(r.cfg.Models.Default))
	}
	// Genkit's spans follow the OpenTelemetry GenAI conventions, on the global tracer provider. Content is recorded
	// on them only when it may be captured.
	telemetry := &otelgenai.Options{ContentCapturingMode: otelgenai.NoContent}
	if r.captureAiContent {
		telemetry = &otelgenai.Options{ContentCapturingMode: otelgenai.SpanOnly, CaptureActionIO: true}
	}
	tracing.SetInstrumentation(otelgenai.New(telemetry))
	r.gk = genkit.Init(ctx, gkOpts...)
	return r.validateConfig()
}

// validateConfig checks that each configured model's provider is enabled. A provider can still reject a model
// it does not serve, which surfaces when the model is called.
func (r *AiRuntime) validateConfig() error {
	for _, name := range r.cfg.Models.Names() {
		if genkit.LookupModel(r.gk, name) == nil {
			return fmt.Errorf("model %q has no enabled provider: name it as provider/model and enable its provider under ai.providers", name)
		}
	}
	return nil
}

// modelFor is the model of the named agent or workflow.
func (r *AiRuntime) modelFor(name string) string {
	return rezai.ModelFor(r.cfg.Models, name)
}

// extraModelsPlugin serves models that come from no provider.
type extraModelsPlugin []*ai.ModelAction

func (p extraModelsPlugin) Name() string {
	return "extra"
}

func (p extraModelsPlugin) Init(context.Context) []gkapi.Action {
	actions := make([]gkapi.Action, len(p))
	for i, model := range p {
		actions[i] = model
	}
	return actions
}

// modelProvider is an enabled model provider.
type modelProvider struct {
	// defaultModel is the model used when ai.models.default is empty and this is the only enabled provider.
	defaultModel string
	plugin       func(rez.AiModelsConfig) gkapi.Plugin
}

func enabledProviders(cfg rez.AiProvidersConfig) []modelProvider {
	var providers []modelProvider
	if cfg.GoogleAI.Enabled {
		providers = append(providers, modelProvider{
			defaultModel: "googleai/gemini-flash-latest",
			plugin: func(rez.AiModelsConfig) gkapi.Plugin {
				return &googlegenai.GoogleAI{APIKey: cfg.GoogleAI.APIKey}
			},
		})
	}
	if cfg.Anthropic.Enabled {
		providers = append(providers, modelProvider{
			defaultModel: "anthropic/claude-haiku-5-5",
			plugin: func(models rez.AiModelsConfig) gkapi.Plugin {
				return anthropicPlugin(cfg.Anthropic, models.Names())
			},
		})
	}
	return providers
}

// anthropicPlugin serves Claude models. The plugin only assumes native structured output for the models it lists,
// so each configured model it serves is declared with it; current models support it.
func anthropicPlugin(cfg rez.AiProviderConfigAnthropic, modelNames []string) *anthropic.Anthropic {
	plugin := &anthropic.Anthropic{
		APIKey: cfg.APIKey,
		Opts:   []option.RequestOption{option.WithMiddleware(cacheAnthropicMessages)},
		Models: make(map[string]ai.ModelOptions),
	}
	structured := ai.ModelOptions{Supports: &ai.ModelSupports{
		Multiturn:   true,
		Tools:       true,
		SystemRole:  true,
		Media:       true,
		Output:      []string{"text", "json"},
		Constrained: ai.ConstrainedSupportAll,
	}}
	for _, name := range modelNames {
		if strings.HasPrefix(name, plugin.Name()+"/") {
			plugin.Models[name] = structured
		}
	}
	return plugin
}

// cacheAnthropicMessages turns on the API's automatic prompt caching for each Messages request, so a tool loop's
// later calls read the conversation so far from the cache. The plugin sets no cache control of its own.
func cacheAnthropicMessages(req *http.Request, next option.MiddlewareNext) (*http.Response, error) {
	if req.Method != http.MethodPost || !strings.HasSuffix(req.URL.Path, "/v1/messages") || req.Body == nil {
		return next(req)
	}
	body, readErr := io.ReadAll(req.Body)
	if readErr != nil {
		return nil, fmt.Errorf("read anthropic request body: %w", readErr)
	}
	fields := map[string]json.RawMessage{}
	if decodeErr := json.Unmarshal(body, &fields); decodeErr == nil {
		if _, set := fields["cache_control"]; !set {
			fields["cache_control"] = json.RawMessage(`{"type":"ephemeral"}`)
			if encoded, encodeErr := json.Marshal(fields); encodeErr == nil {
				body = encoded
			}
		}
	}
	req.Body = io.NopCloser(bytes.NewReader(body))
	req.ContentLength = int64(len(body))
	req.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(body)), nil }
	return next(req)
}
