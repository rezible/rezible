package db

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/rezible/rezible/ent"
)

func TestAlertEpisodeStateSettle(t *testing.T) {
	start := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	at := func(minutes int) time.Time { return start.Add(time.Duration(minutes) * time.Minute) }
	window := func(key string, firedAt, lastObservedAt int, resolvedAt *int) *ent.AlertInstance {
		w := &ent.AlertInstance{ID: uuid.New(), InstanceKey: key, FiredAt: at(firedAt), LastObservedAt: at(lastObservedAt)}
		if resolvedAt != nil {
			w.ResolvedAt = new(at(*resolvedAt))
		}
		return w
	}
	cases := []struct {
		name     string
		timeout  time.Duration
		windows  []*ent.AlertInstance
		asOf     int
		ends     []string
		closedAt string
		deadline string
	}{
		{
			name:     "a reported resolution takes precedence",
			timeout:  10 * time.Minute,
			windows:  []*ent.AlertInstance{window("a", 0, 0, new(4)), window("a", 2, 2, nil)},
			asOf:     3,
			ends:     []string{"resolved@4", ""},
			deadline: "12",
		},
		{
			name:     "a pending supersession is the next deadline",
			timeout:  10 * time.Minute,
			windows:  []*ent.AlertInstance{window("a", 0, 0, nil), window("a", 3, 3, nil), window("b", 0, 0, nil)},
			asOf:     1,
			ends:     []string{"", "", ""},
			deadline: "3",
		},
		{
			name:     "supersession takes precedence over an earlier timeout",
			timeout:  10 * time.Minute,
			windows:  []*ent.AlertInstance{window("a", 0, 0, nil), window("a", 12, 12, nil)},
			asOf:     12,
			ends:     []string{"superseded@12", ""},
			deadline: "22",
		},
		{
			name:     "a timeout ends a window and the episode waits out the grace",
			timeout:  10 * time.Minute,
			windows:  []*ent.AlertInstance{window("a", 0, 0, nil)},
			asOf:     10,
			ends:     []string{"timeout@10"},
			deadline: "15",
		},
		{
			name:     "the episode closes at its last end plus the grace",
			timeout:  10 * time.Minute,
			windows:  []*ent.AlertInstance{window("a", 0, 0, nil), window("b", 1, 1, new(13))},
			asOf:     18,
			ends:     []string{"timeout@10", "resolved@13"},
			closedAt: "18",
		},
		{
			name:    "without a timeout an unresolved window has no deadline",
			windows: []*ent.AlertInstance{window("a", 0, 0, nil)},
			asOf:    600,
			ends:    []string{""},
		},
	}
	minutes := func(t *time.Time) string {
		if t == nil {
			return ""
		}
		return fmt.Sprint(int(t.Sub(start) / time.Minute))
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			settlement := newAlertEpisodeState(tc.windows).settle(tc.timeout, 5*time.Minute, at(tc.asOf))
			ends := make([]string, len(settlement.ends))
			for i, end := range settlement.ends {
				if end != nil {
					ends[i] = fmt.Sprintf("%s@%s", end.reason, minutes(&end.at))
				}
			}
			assert.Equal(t, tc.ends, ends)
			assert.Equal(t, tc.closedAt, minutes(settlement.closedAt))
			assert.Equal(t, tc.deadline, minutes(settlement.nextDeadline))
		})
	}
}
