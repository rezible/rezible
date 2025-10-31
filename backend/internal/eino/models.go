package eino

import (
	"context"
	"errors"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/claude"
	"github.com/cloudwego/eino/components/model"
	rez "github.com/rezible/rezible"
)

func newClaudeLanguageModelProvider(ctx context.Context) (model.ToolCallingChatModel, error) {
	apiKey := rez.Config.GetString("AI.ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, errors.New("anthropic api key not set")
	}

	claudeCfg := &claude.Config{
		APIKey: apiKey,
	}
	m, mErr := claude.NewChatModel(ctx, claudeCfg)
	if mErr != nil {
		return nil, fmt.Errorf("new anthropic model: %w", mErr)
	}
	return m, nil
}
