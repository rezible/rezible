package genkit

import (
	"context"
	"fmt"

	"github.com/firebase/genkit/go/ai"
	rez "github.com/rezible/rezible"
	rezai "github.com/rezible/rezible/pkg/ai"
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

// turnUsageMiddleware adds each model call's tokens to the turn's tally. Thinking tokens are billed as output.
// The tool loop runs later iterations inside WrapGenerate's next, so only WrapModel sees one call's own response;
// WrapGenerate passes it the model name through the context.
type turnUsageMiddleware struct{}

type turnUsageModelKey struct{}

func (m *turnUsageMiddleware) Name() string {
	return "turn_usage"
}

func (m *turnUsageMiddleware) New(ctx context.Context) (*ai.Hooks, error) {
	return &ai.Hooks{
		WrapGenerate: func(ctx context.Context, params *ai.GenerateParams, next ai.GenerateNext) (*ai.ModelResponse, error) {
			return next(context.WithValue(ctx, turnUsageModelKey{}, params.Options.Model), params)
		},
		WrapModel: func(ctx context.Context, params *ai.ModelParams, next ai.ModelNext) (*ai.ModelResponse, error) {
			resp, respErr := next(ctx, params)
			if resp != nil && resp.Usage != nil {
				model, _ := ctx.Value(turnUsageModelKey{}).(string)
				usage := resp.Usage
				rezai.AddModelUsage(ctx, model, usage.InputTokens, usage.OutputTokens+usage.ThoughtsTokens)
			}
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
	params := rez.GetAvailableAiAgentToolsParams{AgentName: m.agentName}
	tools, toolsErr := m.integrations.GetAvailableAgentTools(ctx, params)
	if toolsErr != nil {
		return nil, fmt.Errorf("get available integration agent tools: %w", toolsErr)
	}
	return &ai.Hooks{Tools: tools}, nil
}
