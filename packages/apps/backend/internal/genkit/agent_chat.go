package genkit

import (
	"context"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	rezai "github.com/rezible/rezible/pkg/ai"
)

type ChatAgent struct{}

func NewChatAgent() *ChatAgent {
	return &ChatAgent{}
}

func (a *ChatAgent) agentDefinition() rezai.ChatAgentDefinition {
	return rezai.ChatAgent
}

func (a *ChatAgent) transformState(ctx context.Context, state *aix.SessionState[rezai.ChatAgentState]) (*aix.SessionState[rezai.ChatAgentState], error) {
	return state, nil
}

func (a *ChatAgent) transformStreamChunk(ctx context.Context, chunk *aix.AgentStreamChunk) (*aix.AgentStreamChunk, error) {
	return chunk, nil
}

func (a *ChatAgent) makeInitialUserMessage(ctx context.Context, input rezai.ChatAgentInput) (*ai.Message, error) {
	return ai.NewUserTextMessage(input.Message), nil
}
