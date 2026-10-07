package integrations

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestReadStep(t *testing.T) {
	end := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)
	start := end.Add(-time.Hour)

	t.Run("a fractional step below the minimum is raised and recorded", func(t *testing.T) {
		step, limit := ReadStep(start, end, 11500*time.Millisecond)

		require.Equal(t, 12*time.Second, step)
		require.Equal(t, "step raised from 11.5s to 12s to stay within 300 points", limit)
	})

	t.Run("a fractional step above the minimum is rounded up without a limit", func(t *testing.T) {
		step, limit := ReadStep(start, end, 12500*time.Millisecond)

		require.Equal(t, 13*time.Second, step)
		require.Empty(t, limit)
	})
}
