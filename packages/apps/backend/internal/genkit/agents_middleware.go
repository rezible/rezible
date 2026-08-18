package genkit

import (
	"context"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	rez "github.com/rezible/rezible"
)

type (
	AgentDetails struct {
		Name string
	}
	AgentMiddlewareConstructorFn = func(AgentDetails) ai.Middleware
)

type agentDebugMiddleware struct{}

func (m *agentDebugMiddleware) Name() string {
	return "agent_debug"
}

func (m *agentDebugMiddleware) New(ctx context.Context) (*ai.Hooks, error) {
	return &ai.Hooks{
		WrapGenerate: func(ctx context.Context, params *ai.GenerateParams, next ai.GenerateNext) (*ai.ModelResponse, error) {
			resp, respErr := next(ctx, params)
			//pretty.Println("model response", resp, "error", respErr)
			return resp, respErr
		},
	}, nil
}

type toolCallDisplayLabelMiddleware struct{}

func (m *toolCallDisplayLabelMiddleware) Name() string {
	return "toolcall_display_label"
}

func (m *toolCallDisplayLabelMiddleware) New(ctx context.Context) (*ai.Hooks, error) {
	return &ai.Hooks{
		WrapTool: func(ctx context.Context, params *ai.ToolParams, next ai.ToolNext) (*ai.MultipartToolResponse, error) {
			// TODO: wrap tool call params to add user-facing display text field
			return next(ctx, params)
		},
	}, nil
}

func WithIntegrationToolsMiddleware(integrations rez.IntegrationService) AgentMiddlewareConstructorFn {
	return func(d AgentDetails) ai.Middleware {
		return newIntegrationToolsMiddleware(d.Name, integrations)
	}
}

type integrationToolsMiddleware struct {
	agentName    string
	integrations rez.IntegrationService
}

func newIntegrationToolsMiddleware(agentName string, integrations rez.IntegrationService) *integrationToolsMiddleware {
	return &integrationToolsMiddleware{agentName: agentName, integrations: integrations}
}

func (m *integrationToolsMiddleware) Name() string {
	return "integration_tools"
}

func (m *integrationToolsMiddleware) New(ctx context.Context) (*ai.Hooks, error) {
	params := rez.GetAvailableAgentToolsParams{AgentName: m.agentName}
	tools, toolsErr := m.integrations.GetAvailableAgentTools(ctx, params)
	if toolsErr != nil {
		return nil, fmt.Errorf("get available integration agent tools: %w", toolsErr)
	}
	return &ai.Hooks{Tools: tools}, nil
}
