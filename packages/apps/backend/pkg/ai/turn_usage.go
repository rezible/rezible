package ai

import (
	"context"
	"sync"
)

// TurnUsage is what one attempt of an agent turn used.
type TurnUsage struct {
	Model        string
	InputTokens  int
	OutputTokens int
	Reads        int
}

// turnTally collects one attempt's usage. Tool calls run concurrently, so it is locked.
type turnTally struct {
	mu    sync.Mutex
	usage TurnUsage
}

type turnTallyKey struct{}

// WithTurnUsage starts a tally for one attempt. totals returns what has been added so far.
func WithTurnUsage(ctx context.Context) (_ context.Context, totals func() TurnUsage) {
	tally := &turnTally{}
	totals = func() TurnUsage {
		tally.mu.Lock()
		defer tally.mu.Unlock()
		return tally.usage
	}
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

// AddModelUsage adds one model response to the tally in ctx. It does nothing outside a turn.
func AddModelUsage(ctx context.Context, model string, inputTokens, outputTokens int) {
	addTurnUsage(ctx, func(u *TurnUsage) {
		u.Model = model
		u.InputTokens += inputTokens
		u.OutputTokens += outputTokens
	})
}

// AddRead counts one provider read in the tally in ctx. It does nothing outside a turn.
func AddRead(ctx context.Context) {
	addTurnUsage(ctx, func(u *TurnUsage) {
		u.Reads++
	})
}
