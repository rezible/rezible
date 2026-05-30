package adk

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	"google.golang.org/adk/model"
)

type AnthropicConfig struct {
	ApiKey string `cfg:"api_key"`
}

func getConfigModels(ctx context.Context) ([]model.LLM, error) {
	var models []model.LLM

	return models, nil
}

func newClaudeLanguageModel(ctx context.Context, cl rez.ConfigLoader) (model.LLM, error) {
	var cfg AnthropicConfig
	if cfgErr := cl.Unmarshal("anthropic", &cfg); cfgErr != nil {
		return nil, fmt.Errorf("anthropic config: %w", cfgErr)
	}

	//claudeCfg := &claude.Config{
	//	APIKey: cfg.ApiKey,
	//}
	//m, mErr := claude.NewChatModel(ctx, claudeCfg)
	//if mErr != nil {
	//	return nil, fmt.Errorf("new anthropic model: %w", mErr)
	//}
	//return m, nil
	return nil, nil
}
