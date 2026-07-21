package ai

import (
	"embed"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/google/uuid"
)

//go:embed prompts
var PromptsDir embed.FS

type EventOnAgentTurnFinished struct {
	AgentSessionId       uuid.UUID
	AgentSessionMetadata map[string]any
	AgentTurnId          uuid.UUID
	FinishReason         aix.AgentFinishReason
	Response             *ai.Message
}
