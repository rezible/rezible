package genkit

import (
	"context"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	rez "github.com/rezible/rezible"
)

type (
	AgentMiddlewareConstructorFn = func(agentName string) ai.Middleware
)

type wrapGenerateFn = func(context.Context, *ai.GenerateParams, ai.GenerateNext) (*ai.ModelResponse, error)

func makeSystemTextInjectorFn(marker, text string) wrapGenerateFn {
	markedPart := ai.NewTextPart(text)
	markedPart.Metadata = map[string]any{marker: true}

	injectRequest := func(req *ai.ModelRequest) *ai.ModelRequest {
		newReq := *req
		newReq.Messages = append([]*ai.Message(nil), req.Messages...)
		reqPart := markedPart.Clone()
		for i, message := range newReq.Messages {
			if message == nil {
				continue
			}
			for j, part := range message.Content {
				if part == nil || !part.IsText() || part.Metadata == nil || part.Metadata[marker] != true {
					continue
				}
				if part.Text == text {
					return &newReq
				}
				msgCopy := message.Clone()
				msgCopy.Content[j] = reqPart
				newReq.Messages[i] = msgCopy
				return &newReq
			}
		}
		for i, message := range newReq.Messages {
			if message == nil || message.Role != ai.RoleSystem {
				continue
			}
			msgCopy := message.Clone()
			msgCopy.Content = append(msgCopy.Content, reqPart)
			newReq.Messages[i] = msgCopy
			return &newReq
		}
		newReq.Messages = append([]*ai.Message{ai.NewSystemMessage(reqPart)}, newReq.Messages...)
		return &newReq
	}
	return func(ctx context.Context, params *ai.GenerateParams, next ai.GenerateNext) (*ai.ModelResponse, error) {
		p := *params
		p.Request = injectRequest(params.Request)
		return next(ctx, &p)
	}
}

type agentDebugMiddleware struct{}

func (m *agentDebugMiddleware) Name() string {
	return "agent_debug"
}

func (m *agentDebugMiddleware) New(ctx context.Context) (*ai.Hooks, error) {
	return &ai.Hooks{
		WrapTool: func(ctx context.Context, params *ai.ToolParams, next ai.ToolNext) (*ai.MultipartToolResponse, error) {
			//pretty.Println("tool request", params.Request)
			resp, respErr := next(ctx, params)
			//pretty.Println("tool response", resp, "error", respErr)
			return resp, respErr
		},
		WrapGenerate: func(ctx context.Context, params *ai.GenerateParams, next ai.GenerateNext) (*ai.ModelResponse, error) {
			//pretty.Println("model request", params.Request)
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

func WithIntegrationToolsAgentMiddleware(integrations rez.IntegrationService) AgentMiddlewareConstructorFn {
	return func(agentName string) ai.Middleware {
		return newIntegrationToolsMiddleware(agentName, integrations)
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
