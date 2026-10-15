package ai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddModelUsage(t *testing.T) {
	// Calls outside a turn have no tally.
	AddModelUsage(context.Background(), "test/model", 100, 60, 10)
	ctx, totals := WithTurnUsage(t.Context())
	AddModelUsage(ctx, "test/model", 100, 60, 10)
	AddModelUsage(ctx, "test/model", 200, 120, 20)
	require.Equal(t, TurnUsage{Model: "test/model", InputTokens: 300, CachedInputTokens: 180, OutputTokens: 30}, totals())
}
