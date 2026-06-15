package adk

import (
	"context"
	"fmt"

	"github.com/rezible/rezible/ent/agentrunartifact"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/adk/tool"
	"google.golang.org/genai"

	rez "github.com/rezible/rezible"
)

type agentRunOutput struct {
	Text string
}

func runAgentOnce(ctx context.Context, modelFactory LanguageModelFactory, name, instruction, userInput string, tools []tool.Tool) (*agentRunOutput, error) {
	llm, modelErr := modelFactory.Model(ctx)
	if modelErr != nil {
		return nil, modelErr
	}

	a, agentErr := llmagent.New(llmagent.Config{
		Name:            name,
		Description:     "Rezible operational memory workflow agent.",
		Model:           llm,
		Instruction:     instruction,
		Tools:           tools,
		IncludeContents: llmagent.IncludeContentsNone,
	})
	if agentErr != nil {
		return nil, fmt.Errorf("create adk agent: %w", agentErr)
	}

	r, runnerErr := runner.New(runner.Config{
		AppName:           "rezible",
		Agent:             a,
		SessionService:    session.InMemoryService(),
		AutoCreateSession: true,
	})
	if runnerErr != nil {
		return nil, fmt.Errorf("create adk runner: %w", runnerErr)
	}

	sessionID := name + "-session"
	userID := "rezible-agent"
	msg := genai.NewContentFromText(userInput, genai.RoleUser)
	var out agentRunOutput
	for ev, evErr := range r.Run(ctx, userID, sessionID, msg, agent.RunConfig{}) {
		if evErr != nil {
			return nil, evErr
		}
		if ev == nil || ev.LLMResponse.Content == nil {
			continue
		}
		for _, part := range ev.LLMResponse.Content.Parts {
			out.Text += part.Text
		}
	}
	if out.Text == "" {
		return nil, fmt.Errorf("adk agent returned empty response")
	}
	return &out, nil
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
