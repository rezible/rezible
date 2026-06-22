package eino

import (
	"context"
	"fmt"
	"strings"

	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"github.com/rezible/rezible"
	"github.com/rezible/rezible/ent/agentrunartifact"
)

type modelRunOutput struct {
	Text string
}

func runModelOnce(ctx context.Context, modelFactory ModelProvider, name, instruction, userInput string) (*modelRunOutput, error) {
	chatModel, modelErr := modelFactory.Model(ctx)
	if modelErr != nil {
		return nil, modelErr
	}

	msgs := []*schema.Message{
		{
			Role:    schema.System,
			Content: instruction,
			Name:    name,
		},
		{
			Role:    schema.User,
			Content: userInput,
		},
	}
	response, runErr := chatModel.Generate(ctx, msgs, einomodel.WithTemperature(0))
	if runErr != nil {
		return nil, fmt.Errorf("run eino model: %w", runErr)
	}
	if response == nil || strings.TrimSpace(response.Content) == "" {
		return nil, fmt.Errorf("eino model returned empty response")
	}
	return &modelRunOutput{Text: response.Content}, nil
}

func redactedModelArtifact(name string, cfg rez.AiConfig) AgentRunArtifact {
	payload := map[string]any{
		"provider": cfg.Provider,
		"model":    cfg.Model,
	}
	if !cfg.StoreRawModelPayloads {
		payload["raw_payloads"] = "redacted"
	}
	return AgentRunArtifact{
		Kind:     agentrunartifact.KindModel,
		Name:     name,
		Payload:  payload,
		Redacted: !cfg.StoreRawModelPayloads,
	}
}
