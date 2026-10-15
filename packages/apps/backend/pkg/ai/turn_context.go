package ai

import (
	"encoding/json"
	"strings"

	"github.com/firebase/genkit/go/ai"
)

// PriorTurnToolResultLimit is the serialized size above which a tool result from an earlier turn is replaced
// by a note.
const PriorTurnToolResultLimit = 1000

// priorTurnToolResultNote is what the model reads in place of a large result from an earlier turn.
const priorTurnToolResultNote = "Result of a tool call in an earlier turn, removed to save space. Call the tool again if you need it."

// CompactPriorTurnMessages returns the messages of earlier turns as the model sees them: each tool response
// whose output serializes to more than PriorTurnToolResultLimit bytes is replaced by a note keeping its refs.
// It returns new messages and parts where it changes them and never mutates its input.
func CompactPriorTurnMessages(messages []*ai.Message) []*ai.Message {
	compacted := make([]*ai.Message, len(messages))
	for index, message := range messages {
		compacted[index] = compactPriorTurnMessage(message)
	}
	return compacted
}

func compactPriorTurnMessage(message *ai.Message) *ai.Message {
	if message == nil {
		return nil
	}
	var content []*ai.Part
	for index, part := range message.Content {
		note, replaced := priorTurnToolResultNoteFor(part)
		if !replaced {
			if content != nil {
				content = append(content, part)
			}
			continue
		}
		if content == nil {
			content = append(make([]*ai.Part, 0, len(message.Content)), message.Content[:index]...)
		}
		content = append(content, note)
	}
	if content == nil {
		return message
	}
	copied := *message
	copied.Content = content
	return &copied
}

// priorTurnToolResultNoteFor returns the note replacing the part, when it is a tool response too large to keep.
func priorTurnToolResultNoteFor(part *ai.Part) (*ai.Part, bool) {
	if !part.IsToolResponse() || part.ToolResponse == nil {
		return nil, false
	}
	encoded, encodeErr := json.Marshal(part.ToolResponse.Output)
	if encodeErr != nil || len(encoded) <= PriorTurnToolResultLimit {
		return nil, false
	}
	output := map[string]any{"omitted": priorTurnToolResultNote}
	var fields map[string]any
	if json.Unmarshal(encoded, &fields) == nil {
		for key, value := range fields {
			if text, isText := value.(string); isText && (key == "ref" || strings.HasSuffix(key, "_ref")) {
				output[key] = text
			}
		}
	}
	copied := *part
	copied.ToolResponse = &ai.ToolResponse{Name: part.ToolResponse.Name, Ref: part.ToolResponse.Ref, Output: output}
	return &copied, true
}
