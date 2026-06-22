package eino

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/gemini"
	einomodel "github.com/cloudwego/eino/components/model"
	"google.golang.org/genai"

	rez "github.com/rezible/rezible"
)

type ModelProvider interface {
	Model(context.Context) (einomodel.BaseChatModel, error)
	ModelMetadata() map[string]any
}

type configChatModel struct {
	cfg rez.AiConfig
}

func newChatModelProvider(cfg rez.AiConfig) ModelProvider {
	return &configChatModel{cfg: cfg}
}

func (f *configChatModel) ModelMetadata() map[string]any {
	return map[string]any{
		"provider": f.cfg.Provider,
		"model":    f.cfg.Model,
		"runtime":  "eino",
	}
}

func (f *configChatModel) Model(ctx context.Context) (einomodel.BaseChatModel, error) {
	if !f.cfg.Enabled {
		return nil, fmt.Errorf("agent model provider disabled")
	}
	switch f.cfg.Provider {
	case "gemini":
		if f.cfg.Gemini.APIKey == "" {
			return nil, fmt.Errorf("gemini api key is required")
		}
		client, clientErr := genai.NewClient(ctx, &genai.ClientConfig{APIKey: f.cfg.Gemini.APIKey})
		if clientErr != nil {
			return nil, fmt.Errorf("create gemini client: %w", clientErr)
		}
		model, modelErr := gemini.NewChatModel(ctx, &gemini.Config{
			Client: client,
			Model:  f.cfg.Model,
		})
		if modelErr != nil {
			return nil, fmt.Errorf("create eino gemini chat model: %w", modelErr)
		}
		return model, nil
	default:
		return nil, fmt.Errorf("unsupported agent model provider %q", f.cfg.Provider)
	}
}
