package eino

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudwego/eino-ext/components/model/claude"
	"github.com/cloudwego/eino/components/model"
)

const (
	anthropicApiKeyEnvVar = "ANTHROPIC_API_KEY"
)

func newClaudeLanguageModelProvider(ctx context.Context) (model.ToolCallingChatModel, error) {
	apiKey := os.Getenv(anthropicApiKeyEnvVar)
	if apiKey == "" {
		return nil, fmt.Errorf("%s not set", anthropicApiKeyEnvVar)
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
