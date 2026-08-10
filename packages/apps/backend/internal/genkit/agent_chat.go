package genkit

import (
	"context"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	rez "github.com/rezible/rezible"
	rezai "github.com/rezible/rezible/pkg/ai"
)

type ChatAgent struct{}

func NewChatAgent() *ChatAgent {
	return &ChatAgent{}
}

func (a *ChatAgent) agentDefinition() rezai.ChatAgentDefinition {
	return rezai.ChatAgent
}

func (a *ChatAgent) makeInitialTurnInput(ctx context.Context, input rezai.ChatAgentInput) (*rez.AiAgentTurnInput, error) {
	return &rez.AiAgentTurnInput{Message: ai.NewUserTextMessage(input.Message)}, nil
}

func (a *ChatAgent) transformState(ctx context.Context, state *aix.SessionState[rezai.ChatAgentState]) (*aix.SessionState[rezai.ChatAgentState], error) {
	return state, nil
}

func (a *ChatAgent) transformStreamChunk(ctx context.Context, chunk *aix.AgentStreamChunk) (*aix.AgentStreamChunk, error) {
	return chunk, nil
}
