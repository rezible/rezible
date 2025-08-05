package anthropic

import (
	"context"
	"errors"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/claude"
	rez "github.com/rezible/rezible"
)

var (
	ErrNoApiToken = errors.New("no anthropic api token provided")
)

type LanguageModelProvider struct {
	model *claude.ChatModel
}

type Config struct {
	ApiToken string `json:"api_token"`
}

func NewClaudeLanguageModelProvider(ctx context.Context, cfg Config) (*LanguageModelProvider, error) {
	if cfg.ApiToken == "" {
		return nil, ErrNoApiToken
	}

	claudeCfg := &claude.Config{
		APIKey: cfg.ApiToken,
	}
	m, mErr := claude.NewChatModel(ctx, claudeCfg)
	if mErr != nil {
		return nil, fmt.Errorf("new anthropic model: %w", mErr)
	}
	return &LanguageModelProvider{model: m}, nil
}

func (p *LanguageModelProvider) Model() rez.AiLanguageModel {
	return p.model
}
