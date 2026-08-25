package ai

import (
	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent/agentturn"
)

type EventOnAgentTurnFinished struct {
	AgentSessionId       uuid.UUID
	AgentSessionMetadata map[string]any
	AgentTurnId          uuid.UUID
	FinishReason         aix.AgentFinishReason
	Response             *ai.Message
}

type AgentTurnUpdated struct {
	AgentSessionId uuid.UUID
	AgentTurnId    uuid.UUID
	Status         agentturn.Status
	FinishReason   string
}

func (e AgentTurnUpdated) MessageScopes() []string {
	return []string{"agent_session:" + e.AgentSessionId.String(), "agent_turn:" + e.AgentTurnId.String()}
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
