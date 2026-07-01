package genkit

import (
	"context"
	"encoding/json"
	"fmt"

	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ars "github.com/rezible/rezible/ent/agentrunsnapshot"
)

type agentSessionStore[S any] struct {
	snapshots rez.AgentRunSnapshotService
}

func makeAgentSessionStore[S any](snapshots rez.AgentRunSnapshotService) *agentSessionStore[S] {
	return &agentSessionStore[S]{snapshots: snapshots}
}

func asSnapshot[S any](rs *ent.AgentRunSnapshot) (*aix.SessionSnapshot[S], error) {
	if rs == nil {
		return nil, nil
	}
	snapshot := &aix.SessionSnapshot[S]{
		SessionID:    rs.AgentRunID.String(),
		SnapshotID:   rs.ID.String(),
		FinishReason: aix.AgentFinishReason(rs.FinishReason),
		Status:       aix.SnapshotStatus(rs.Status.String()),
		HeartbeatAt:  rs.HeartbeatAt,
		CreatedAt:    rs.CreatedAt,
		UpdatedAt:    rs.UpdatedAt,
	}
	if parentId := rs.ParentID; parentId != nil && *parentId != uuid.Nil {
		snapshot.ParentID = (*parentId).String()
	}
	if rs.State != nil && len(*rs.State) > 0 {
		if jsonErr := json.Unmarshal(*rs.State, &snapshot.State); jsonErr != nil {
			return nil, fmt.Errorf("unmarshal session state: %w", jsonErr)
		}
	}
	if rs.Error != nil && len(*rs.Error) > 0 {
		if jsonErr := json.Unmarshal(*rs.Error, &snapshot.Error); jsonErr != nil {
			return nil, fmt.Errorf("unmarshal session error: %w", jsonErr)
		}
	}
	return snapshot, nil
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
	return asSnapshot[S](rs)
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
	return asSnapshot[S](rs)
}

func (s *agentSessionStore[S]) SaveSnapshot(ctx context.Context, id string, setFn func(*aix.SessionSnapshot[S]) (*aix.SessionSnapshot[S], error)) (*aix.SessionSnapshot[S], error) {
	var snapshotId uuid.UUID
	if id != "" {
		var idErr error
		if snapshotId, idErr = uuid.Parse(id); idErr != nil {
			return nil, fmt.Errorf("invalid ID: %s", id)
		}
	}
	setSnapshotDataFn := func(rs *ent.AgentRunSnapshot, m *ent.AgentRunSnapshotMutation) error {
		existing, convErr := asSnapshot[S](rs)
		if convErr != nil {
			return fmt.Errorf("convert existing snapshot: %w", convErr)
		}

		snapshot, updateErr := setFn(existing)
		if updateErr != nil {
			return fmt.Errorf("update existing snapshot: %w", updateErr)
		} else if snapshot == nil {
			return nil
		}

		m.SetStatus(ars.Status(snapshot.Status))
		m.SetFinishReason(string(snapshot.FinishReason))
		m.SetCreatedAt(snapshot.CreatedAt)
		m.SetUpdatedAt(snapshot.UpdatedAt)

		runId, runIdErr := uuid.Parse(snapshot.SessionID)
		if runIdErr != nil {
			return fmt.Errorf("invalid session ID: %s", snapshot.SessionID)
		}
		m.SetAgentRunID(runId)

		if len(snapshot.ParentID) > 0 {
			parentId, parentIdErr := uuid.Parse(snapshot.ParentID)
			if parentIdErr != nil {
				return fmt.Errorf("invalid parent ID: %s", snapshot.ParentID)
			}
			m.SetParentID(parentId)
		}

		if snapshot.HeartbeatAt != nil {
			m.SetHeartbeatAt(*snapshot.HeartbeatAt)
		}

		if snapshot.State != nil {
			state, jsonErr := json.Marshal(snapshot.State)
			if jsonErr != nil {
				return fmt.Errorf("marshal state: %w", jsonErr)
			}
			m.SetState(state)
		}
		if snapshot.Error != nil {
			sessErr, jsonErr := json.Marshal(snapshot.State)
			if jsonErr != nil {
				return fmt.Errorf("marshal error: %w", jsonErr)
			}
			m.SetError(sessErr)
		}
		return nil
	}
	updated, updateErr := s.snapshots.UpdateSnapshot(ctx, snapshotId, setSnapshotDataFn)
	if updateErr != nil {
		return nil, fmt.Errorf("save snapshot: %w", updateErr)
	}
	return asSnapshot[S](updated)
}
