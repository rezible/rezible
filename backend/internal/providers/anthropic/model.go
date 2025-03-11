package anthropic

import (
	"errors"
	"fmt"
	rez "github.com/rezible/rezible"

	"github.com/tmc/langchaingo/llms/anthropic"
)

var (
	ErrNoApiToken = errors.New("no anthropic api token provided")
)

type ModelProvider struct {
	model *anthropic.LLM
}

type Config struct {
	ApiToken string `json:"api_token"`
}

func NewClaudeAiModelProvider(cfg Config) (*ModelProvider, error) {
	if cfg.ApiToken == "" {
		return nil, ErrNoApiToken
	}
	m, mErr := anthropic.New(anthropic.WithToken(cfg.ApiToken))
	if mErr != nil {
		return nil, fmt.Errorf("new anthropic model: %w", mErr)
	}
	return &ModelProvider{model: m}, nil
}

func (p *ModelProvider) Model() rez.AiModel {
	return p.model
}
