package ai

import (
	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
)

type EventOnAgentTurnFinished struct {
	AgentSessionId       uuid.UUID
	AgentSessionMetadata map[string]any
	AgentTurnId          uuid.UUID
	FinishReason         aix.AgentFinishReason
	Response             *ai.Message
}

func (e EventOnAgentTurnFinished) MessageScopes() []string {
	return []string{"agent_session:" + e.AgentSessionId.String(), "agent_turn:" + e.AgentTurnId.String()}
}

type EventOnAgentTurnChunk struct {
	AgentSessionId uuid.UUID
	AgentTurnId    uuid.UUID
	Chunk          rez.AiAgentTurnChunk
}

func (e EventOnAgentTurnChunk) MessageScopes() []string {
	return []string{"agent_session:" + e.AgentSessionId.String(), "agent_turn:" + e.AgentTurnId.String()}
}
