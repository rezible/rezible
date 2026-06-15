package adk

import (
	"context"
	"fmt"

	"google.golang.org/adk/model"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/genai"

	rez "github.com/rezible/rezible"
)

type LanguageModelFactory interface {
	Model(context.Context) (model.LLM, error)
	ModelMetadata() map[string]any
}

type configLanguageModelFactory struct {
	cfg rez.AiConfig
}

func newLanguageModelFactory(cfg rez.AiConfig) LanguageModelFactory {
	return &configLanguageModelFactory{cfg: cfg}
}

func (f *configLanguageModelFactory) ModelMetadata() map[string]any {
	return map[string]any{
		"provider": f.cfg.Provider,
		"model":    f.cfg.Model,
	}
}

func (f *configLanguageModelFactory) Model(ctx context.Context) (model.LLM, error) {
	if !f.cfg.Enabled {
		return nil, fmt.Errorf("agent model provider disabled")
	}
	switch f.cfg.Provider {
	case "gemini":
		if f.cfg.Gemini.APIKey == "" {
			return nil, fmt.Errorf("gemini api key is required")
		}
		return gemini.NewModel(ctx, f.cfg.Model, &genai.ClientConfig{APIKey: f.cfg.Gemini.APIKey})
	default:
		return nil, fmt.Errorf("unsupported agent model provider %q", f.cfg.Provider)
	}
}
