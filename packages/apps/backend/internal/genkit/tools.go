package genkit

import (
	"context"
	"fmt"

	genkitx "github.com/firebase/genkit/go/genkit/exp"
	rez "github.com/rezible/rezible"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/execution"
)

type ToolRunner[Input any, Output any] interface {
	Definition() rezai.ToolDefinition[Input, Output]
	ToolFunc(context.Context, Input) (Output, error)
	//ToolOpts() []ai.ToolOption
}

func WithTool[I any, O any](t ToolRunner[I, O]) AiServiceOption {
	return AiServiceOption{
		kind: "tool",
		optFn: func(s *AiService) error {
			def := t.Definition()
			genkitx.DefineTool(s.gk, def.Name(), def.Description(), t.ToolFunc) //, t.ToolOpts()...)
			return nil
		},
	}
}

type SendChatMessageTool struct {
	msgs rez.MessageService
}

func NewSendChatMessageTool(msgs rez.MessageService) *SendChatMessageTool {
	return &SendChatMessageTool{msgs: msgs}
}

func (t *SendChatMessageTool) Definition() rezai.SendChatMessageToolDefinition {
	return rezai.SendChatMessageTool
}

func (t *SendChatMessageTool) ToolFunc(ctx context.Context, input rezai.SendChatMessageToolInput) (rezai.SendChatMessageToolOutput, error) {
	status := "message sent"
	if msgErr := t.publishMessageEvent(ctx, input); msgErr != nil {
		status = fmt.Sprintf("failed to send: %s", msgErr.Error())
	}
	return rezai.SendChatMessageToolOutput{Status: status}, nil
}

func (t *SendChatMessageTool) publishMessageEvent(ctx context.Context, input rezai.SendChatMessageToolInput) error {
	runId, idOk := execution.GetContext(ctx).AiAgentRunID()
	if !idOk {
		return fmt.Errorf("no agent run id in context")
	}
	evt := &rezai.EventSendChatMessageToolInvoked{
		AgentRunId: runId,
		Input:      input,
	}
	if msgErr := t.msgs.PublishEvent(ctx, evt); msgErr != nil {
		return fmt.Errorf("failed to send: %w", msgErr)
	}
	return nil
}
