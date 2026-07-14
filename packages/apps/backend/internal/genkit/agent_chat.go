package genkit

import (
	"context"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	rezai "github.com/rezible/rezible/pkg/ai"
)

type (
	ChatAgent struct {
	}
	chatAgentSessionState = aix.SessionState[rezai.ChatAgentState]
)

func NewChatAgent() *ChatAgent {
	return &ChatAgent{}
}

func (a *ChatAgent) definition() rezai.ChatAgentDefinition {
	return rezai.ChatAgent
}

func (a *ChatAgent) transformState(ctx context.Context, state *chatAgentSessionState) (*chatAgentSessionState, error) {
	return state, nil
}

func (a *ChatAgent) transformStreamChunk(ctx context.Context, chunk *aix.AgentStreamChunk) (*aix.AgentStreamChunk, error) {
	return chunk, nil
}

func (a *ChatAgent) makeInitialUserMessage(ctx context.Context, input rezai.ChatAgentInput) (*ai.Message, error) {
	return ai.NewUserTextMessage(input.Message), nil
}

//func (a *ChatAgent) makeOutputTool() *aix.Tool[rezai.ChatAgentOutput, AgentOutputToolResult] {
//	return aix.NewTool("send_message", "Reply with a chat message",
//		func(ctx context.Context, out rezai.ChatAgentOutput) (AgentOutputToolResult, error) {
//			status := "Message sent successfully"
//			if msgErr := a.sendChatMessage(ctx, out); msgErr != nil {
//				status = fmt.Sprintf("Error sending message: %s", msgErr.Error())
//			}
//			return AgentOutputToolResult{Status: status}, nil
//		},
//	)
//}
//
//func (a *ChatAgent) sendChatMessage(ctx context.Context, out rezai.ChatAgentOutput) error {
//	sess := aix.SessionFromContext[rezai.ChatAgentState](ctx)
//	if sess == nil {
//		return fmt.Errorf("session not found in context")
//	}
//	fmt.Printf("send chat message: %+v\n", out)
//	return nil
//}
