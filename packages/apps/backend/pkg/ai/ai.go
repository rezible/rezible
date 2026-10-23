package ai

import (
	"cmp"
	"time"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent/agentturn"
)

func makeMessageEventScopes(sessId, turnId uuid.UUID) rez.MessageEventScopes {
	return rez.MessageEventScopes{
		"agent_session:" + sessId.String(),
		"agent_turn:" + turnId.String(),
	}
}

type EventOnAgentTurnFinished struct {
	AgentSessionId       uuid.UUID
	AgentSessionMetadata map[string]any
	AgentTurnId          uuid.UUID
	FinishReason         aix.AgentFinishReason
	Response             *ai.Message
}

func (EventOnAgentTurnFinished) MessageName() string {
	return "agent.turn-finished.v1"
}

func (e EventOnAgentTurnFinished) MessageScopes() rez.MessageEventScopes {
	return makeMessageEventScopes(e.AgentSessionId, e.AgentTurnId)
}

type AgentTurnUpdated struct {
	StartedAt      *time.Time
	AgentSessionId uuid.UUID
	AgentTurnId    uuid.UUID
	Status         agentturn.Status
	FinishReason   string
}

func (AgentTurnUpdated) MessageName() string {
	return "agent.turn-updated.v1"
}

func (e AgentTurnUpdated) MessageScopes() rez.MessageEventScopes {
	return makeMessageEventScopes(e.AgentSessionId, e.AgentTurnId)
}

type EventOnAgentTurnChunk struct {
	StartedAt      *time.Time
	AgentSessionId uuid.UUID
	AgentTurnId    uuid.UUID
	Chunk          rez.AiAgentTurnChunk
}

func (EventOnAgentTurnChunk) MessageName() string {
	return "agent.turn-chunk.v1"
}

func (e EventOnAgentTurnChunk) MessageScopes() rez.MessageEventScopes {
	return makeMessageEventScopes(e.AgentSessionId, e.AgentTurnId)
}

// ModelFor is the model configured for the named agent or workflow, or the default model; "" when neither is set.
func ModelFor(models rez.AiModelsConfig, name string) string {
	var model string
	switch name {
	case InvestigationAgent.Name:
		model = models.Investigation
	case ChatAgent.Name:
		model = models.Chat
	case JudgeSituationCandidateDefinition.Name:
		model = models.SituationJudge
	case ClassifyAgentThreadResponseDefinition.Name:
		model = models.ThreadClassifier
	}
	return cmp.Or(model, models.Default)
}
