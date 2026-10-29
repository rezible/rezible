package ai

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/firebase/genkit/go/core/tracing"
)

// TurnUsage is what one attempt of an agent turn used.
type TurnUsage struct {
	Model             string
	InputTokens       int
	CachedInputTokens int
	OutputTokens      int
	Reads             int
	// ModelCalls is how many model calls the attempt made; PeakInputTokens is the largest input of any one.
	ModelCalls      int
	PeakInputTokens int
}

// turnTally collects one attempt's usage. Tool calls run concurrently, so it is locked.
type turnTally struct {
	mu    sync.Mutex
	usage TurnUsage
}

type turnTallyKey struct{}

// WithTurnUsage starts a tally for one attempt. totals returns what has been added so far. Every model call made
// under the returned context is added to it; thinking tokens are billed as output.
func WithTurnUsage(ctx context.Context) (_ context.Context, totals func() TurnUsage) {
	tally := &turnTally{}
	totals = func() TurnUsage {
		tally.mu.Lock()
		defer tally.mu.Unlock()
		return tally.usage
	}
	countModelCall := func(callCtx context.Context, call *aix.ModelCall) {
		// An enclosing tally's observer sees this tally's calls too, since scoped observers can't be removed. Only
		// the innermost tally counts a call, as AddModelUsage does.
		if inner, _ := callCtx.Value(turnTallyKey{}).(*turnTally); inner != tally {
			return
		}
		if usage := call.Response.Usage; usage != nil {
			AddModelUsage(callCtx, call.Model, usage.InputTokens, usage.CachedContentTokens, usage.OutputTokens+usage.ThoughtsTokens)
		}
	}
	ctx = tracing.WithInstrumentation(ctx, aix.Observer{ModelDone: countModelCall})
	return context.WithValue(ctx, turnTallyKey{}, tally), totals
}

func addTurnUsage(ctx context.Context, add func(*TurnUsage)) {
	tally, ok := ctx.Value(turnTallyKey{}).(*turnTally)
	if !ok {
		return
	}
	tally.mu.Lock()
	defer tally.mu.Unlock()
	add(&tally.usage)
}

// AddModelUsage adds one model response to the tally in ctx. cachedInputTokens is the part of inputTokens the
// provider served from its cache, as the model plugin reports it. It does nothing outside a turn.
func AddModelUsage(ctx context.Context, model string, inputTokens, cachedInputTokens, outputTokens int) {
	addTurnUsage(ctx, func(u *TurnUsage) {
		u.Model = model
		u.InputTokens += inputTokens
		u.CachedInputTokens += cachedInputTokens
		u.OutputTokens += outputTokens
		u.ModelCalls++
		u.PeakInputTokens = max(u.PeakInputTokens, inputTokens)
	})
}

// AddRead counts one provider read in the tally in ctx. It does nothing outside a turn.
func AddRead(ctx context.Context) {
	addTurnUsage(ctx, func(u *TurnUsage) {
		u.Reads++
	})
}

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
