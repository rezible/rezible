package ai

import (
	"strings"
	"testing"

	"github.com/firebase/genkit/go/ai"
	"github.com/stretchr/testify/require"
)

func TestCompactPriorTurnMessages(t *testing.T) {
	large := map[string]any{
		"provider_read_ref": "read-1",
		"service_ref":       "service-1",
		"ref":               "other-1",
		"count_ref":         7,
		"lines":             strings.Repeat("upstream timeout ", 100),
	}
	small := map[string]any{"ref": "small-1", "text": "entity small-1 payments-api"}
	largeResponse := &ai.Part{
		Kind:         ai.PartToolResponse,
		ToolResponse: &ai.ToolResponse{Name: "search_logs", Ref: "call-1", Output: large, Content: []*ai.Part{ai.NewTextPart("detail")}},
		Metadata:     map[string]any{"kept": true},
	}
	smallResponse := ai.NewToolResponsePart(&ai.ToolResponse{Name: "list_services", Ref: "call-2", Output: small})
	request := ai.NewToolRequestPart(&ai.ToolRequest{Name: "search_logs", Ref: "call-1", Input: map[string]any{"contains": "timeout"}})
	text := ai.NewTextPart("Looking at the logs.")
	text.Metadata = map[string]any{"signature": "abc"}
	notObject := ai.NewToolResponsePart(&ai.ToolResponse{Name: "list", Ref: "call-3", Output: []any{strings.Repeat("x", PriorTurnToolResultLimit)}})

	modelMessage := ai.NewModelMessage(text, request)
	toolMessage := &ai.Message{Role: ai.RoleTool, Content: []*ai.Part{largeResponse, smallResponse, notObject}}
	messages := []*ai.Message{modelMessage, toolMessage}

	compacted := CompactPriorTurnMessages(messages)

	require.Len(t, compacted, 2)
	require.Same(t, modelMessage, compacted[0], "a message without large results is kept as it is")
	require.NotSame(t, toolMessage, compacted[1])

	note := compacted[1].Content[0]
	require.Equal(t, "search_logs", note.ToolResponse.Name)
	require.Equal(t, "call-1", note.ToolResponse.Ref)
	require.Empty(t, note.ToolResponse.Content)
	require.Equal(t, map[string]any{
		"omitted":           priorTurnToolResultNote,
		"provider_read_ref": "read-1",
		"service_ref":       "service-1",
		"ref":               "other-1",
	}, note.ToolResponse.Output)
	require.Equal(t, map[string]any{"kept": true}, note.Metadata)

	require.Same(t, smallResponse, compacted[1].Content[1], "a result at or under the limit is kept whole")
	require.Equal(t, map[string]any{"omitted": priorTurnToolResultNote}, compacted[1].Content[2].ToolResponse.Output,
		"an output that is not an object keeps no fields")

	require.Same(t, largeResponse, toolMessage.Content[0], "the input message is not mutated")
	require.Equal(t, large, largeResponse.ToolResponse.Output, "the input part is not mutated")
	require.Len(t, largeResponse.ToolResponse.Content, 1)
}

func TestCompactPriorTurnMessagesKeepsResultAtTheLimit(t *testing.T) {
	// {"text":"…"} is 11 bytes around the text.
	atLimit := ai.NewToolResponsePart(&ai.ToolResponse{Name: "read", Ref: "call-1", Output: map[string]any{"text": strings.Repeat("x", PriorTurnToolResultLimit-11)}})
	message := &ai.Message{Role: ai.RoleTool, Content: []*ai.Part{atLimit}}
	require.Same(t, message, CompactPriorTurnMessages([]*ai.Message{message})[0])
}
