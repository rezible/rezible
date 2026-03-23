package adk

import (
	"context"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rotisserie/eris"
	"google.golang.org/adk/model"
)

type AnthropicConfig struct {
	ApiKey string `koanf:"api_key"`
}

func getConfigModels(ctx context.Context) ([]model.LLM, error) {
	var models []model.LLM

	return models, nil
}

func newClaudeLanguageModel(ctx context.Context) (model.LLM, error) {
	var cfg AnthropicConfig
	if cfgErr := rez.Config.Unmarshal("anthropic", &cfg); cfgErr != nil {
		return nil, eris.New("anthropic api key not set")
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
