package v1

import (
	"testing"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/stretchr/testify/require"
)

func TestAgentTurnModelChunkPreservesAttemptAndContent(t *testing.T) {
	startedAt := time.Date(2026, 10, 7, 0, 0, 0, 123456000, time.UTC)
	sessionID, turnID := uuid.New(), uuid.New()
	chunk := rez.AiAgentTurnChunk{ModelChunk: &ai.ModelResponseChunk{
		Role: ai.RoleModel, Index: 2, Content: []*ai.Part{ai.NewTextPart("prefix")},
	}}
	event, convertErr := AgentTurnChunkEventFromRez(sessionID, turnID, &startedAt, chunk)
	require.NoError(t, convertErr)
	require.Equal(t, &startedAt, event.StartedAt)
	require.Equal(t, sessionID, event.SessionId)
	require.Equal(t, turnID, event.TurnId)
	require.NotNil(t, event.Model)
	require.Equal(t, 2, event.Model.Index)
	require.Len(t, event.Model.Parts, 1)
	require.NotNil(t, event.Model.Parts[0].Text)
	require.Equal(t, "prefix", *event.Model.Parts[0].Text)
}
