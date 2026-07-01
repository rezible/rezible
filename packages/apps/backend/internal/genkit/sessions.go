package genkit

import (
	"context"
	"encoding/json"
	"fmt"

	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type agentSessionStore[S any] struct {
	snapshots rez.AgentRunSnapshotService
}

func makeAgentSessionStore[S any](snapshots rez.AgentRunSnapshotService) *agentSessionStore[S] {
	return &agentSessionStore[S]{snapshots: snapshots}
}

func (s *agentSessionStore[S]) asSnapshot(rs *ent.AgentRunSnapshot) (*aix.SessionSnapshot[S], error) {
	if rs == nil {
		return nil, nil
	}
	var snapshot aix.SessionSnapshot[S]
	if jsonErr := json.Unmarshal(rs.Data, &snapshot); jsonErr != nil {
		return nil, fmt.Errorf("unable to unmarshal session snapshot %q: %w", rs.ID, jsonErr)
	}
	return &snapshot, nil
}

func (s *agentSessionStore[S]) GetLatestSnapshot(ctx context.Context, sessionID string) (*aix.SessionSnapshot[S], error) {
	runId, idErr := uuid.Parse(sessionID)
	if idErr != nil {
		return nil, fmt.Errorf("invalid session ID: %s", sessionID)
	}
	rs, queryErr := s.snapshots.GetLatestSnapshot(ctx, runId)
	if queryErr != nil {
		return nil, fmt.Errorf("lookup snapshot: %w", queryErr)
	}
	return s.asSnapshot(rs)
}

func (s *agentSessionStore[S]) GetSnapshot(ctx context.Context, snapshotID string) (*aix.SessionSnapshot[S], error) {
	id, idErr := uuid.Parse(snapshotID)
	if idErr != nil {
		return nil, fmt.Errorf("invalid snapshot ID: %s", snapshotID)
	}
	rs, queryErr := s.snapshots.GetSnapshot(ctx, id)
	if queryErr != nil {
		return nil, fmt.Errorf("lookup snapshot: %w", queryErr)
	}
	return s.asSnapshot(rs)
}

func (s *agentSessionStore[S]) SaveSnapshot(ctx context.Context, snapshotID string, setFn func(*aix.SessionSnapshot[S]) (*aix.SessionSnapshot[S], error)) (*aix.SessionSnapshot[S], error) {
	var id uuid.UUID
	if snapshotID != "" {
		var idErr error
		id, idErr = uuid.Parse(snapshotID)
		if idErr != nil {
			return nil, fmt.Errorf("invalid snapshot ID: %s", snapshotID)
		}
	}
	setSnapshotDataFn := func(rs *ent.AgentRunSnapshot, m *ent.AgentRunSnapshotMutation) error {
		existing, convErr := s.asSnapshot(rs)
		if convErr != nil {
			return fmt.Errorf("convert existing snapshot: %w", convErr)
		}
		updated, updateErr := setFn(existing)
		if updateErr != nil {
			return fmt.Errorf("update existing snapshot: %w", updateErr)
		}
		if updated != nil {
			data, dataErr := json.Marshal(updated)
			if dataErr != nil {
				return fmt.Errorf("marshal updated snapshot: %w", dataErr)
			}
			m.SetData(data)
		}
		return nil
	}
	updated, updateErr := s.snapshots.UpdateSnapshot(ctx, id, setSnapshotDataFn)
	if updateErr != nil {
		return nil, fmt.Errorf("update existing snapshot: %w", updateErr)
	}
	return s.asSnapshot(updated)
}
