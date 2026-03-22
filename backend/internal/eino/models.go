package eino

import (
	"context"
	"errors"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/claude"
	"github.com/cloudwego/eino/components/model"
	rez "github.com/rezible/rezible"
)

type AnthropicConfig struct {
	ApiKey string `koanf:"api_key"`
}

func newClaudeLanguageModelProvider(ctx context.Context) (model.ToolCallingChatModel, error) {
	var cfg AnthropicConfig
	if cfgErr := rez.Config.Unmarshal("anthropic", &cfg); cfgErr != nil {
		return nil, errors.New("anthropic api key not set")
	}

	claudeCfg := &claude.Config{
		APIKey: cfg.ApiKey,
	}
	m, mErr := claude.NewChatModel(ctx, claudeCfg)
	if mErr != nil {
		return nil, fmt.Errorf("new anthropic model: %w", mErr)
	}
	return m, nil
}
