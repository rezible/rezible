package projections

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/rezible/rezible/ent"
)

func TestChangePointEventAfter(t *testing.T) {
	at := time.Date(2026, 10, 7, 14, 0, 0, 0, time.UTC)
	tests := []struct {
		name                          string
		kind                          string
		incomingStatus, currentStatus string
		incomingAt, currentAt         time.Time
		incomingID, currentID         byte
		after                         bool
	}{
		{"later time wins over a terminal status", KindDeployment, "started", "succeeded", at.Add(time.Minute), at, 1, 2, true},
		{"earlier time loses to a started status", KindDeployment, "succeeded", "started", at, at.Add(time.Minute), 2, 1, false},
		{"terminal wins over started with a lower ID", KindDeployment, "failed", "started", at, at, 1, 2, true},
		{"started loses to terminal with a greater ID", KindDeployment, "started", "succeeded", at, at, 2, 1, false},
		{"two terminal statuses fall to the greater ID", KindDeployment, "failed", "succeeded", at, at, 2, 1, true},
		{"two terminal statuses, lower ID", KindDeployment, "succeeded", "failed", at, at, 1, 2, false},
		{"two started statuses fall to the greater ID", KindDeployment, "started", "started", at, at, 2, 1, true},
		{"the same event", KindDeployment, "succeeded", "succeeded", at, at, 1, 1, false},
		{"equal merge times fall to the greater ID", KindCodeChange, "", "", at, at, 2, 1, true},
		{"equal merge times, lower ID", KindCodeChange, "", "", at, at, 1, 2, false},
		{"earlier merge", KindCodeChange, "", "", at, at.Add(time.Minute), 2, 1, false},
		{"status counts only for deployments", KindCodeChange, "succeeded", "started", at, at, 1, 2, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			incoming := &ent.NormalizedEvent{
				ID:         uuid.UUID{15: tc.incomingID},
				Kind:       tc.kind,
				Attributes: fmt.Appendf(nil, `{"status":%q}`, tc.incomingStatus),
			}
			current := &ent.NormalizedEvent{
				ID:         uuid.UUID{15: tc.currentID},
				Kind:       tc.kind,
				Attributes: fmt.Appendf(nil, `{"status":%q}`, tc.currentStatus),
			}
			require.Equal(t, tc.after, ChangePointEventAfter(incoming, tc.incomingAt, current, tc.currentAt))
		})
	}
}
