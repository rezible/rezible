package genkit

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"sync"

	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	aars "github.com/rezible/rezible/ent/aiagentrunsnapshot"
	rezai "github.com/rezible/rezible/pkg/ai"
)

func convertToSessionSnapshot[S rezai.SessionState](rs *ent.AiAgentRunSnapshot) (*aix.SessionSnapshot[S], error) {
	if rs == nil {
		return nil, nil
	}
	snapshot := &aix.SessionSnapshot[S]{
		SessionID:    rs.AiAgentRunID.String(),
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

func convertFromSessionSnapshot[S rezai.SessionState](s *aix.SessionSnapshot[S]) (*ent.AiAgentRunSnapshot, error) {
	if s == nil {
		return nil, nil
	}
	sessId, sessIdErr := uuid.Parse(s.SessionID)
	if sessIdErr != nil {
		return nil, sessIdErr
	}
	snapshotId, snapshotIdErr := uuid.Parse(s.SnapshotID)
	if snapshotIdErr != nil {
		return nil, snapshotIdErr
	}
	var parentId *uuid.UUID
	if s.ParentID != "" {
		id, parseErr := uuid.Parse(s.ParentID)
		if parseErr != nil {
			return nil, parseErr
		}
		parentId = &id
	}
	snap := &ent.AiAgentRunSnapshot{
		ID:           snapshotId,
		CreatedAt:    s.CreatedAt,
		UpdatedAt:    s.UpdatedAt,
		AiAgentRunID: sessId,
		ParentID:     parentId,
		Status:       aars.Status(s.Status),
		FinishReason: string(s.FinishReason),
		HeartbeatAt:  s.HeartbeatAt,
	}
	if s.State != nil {
		enc, encErr := json.Marshal(s.State)
		if encErr != nil {
			return nil, encErr
		}
		snap.State = &enc
	}
	if s.Error != nil {
		enc, encErr := json.Marshal(s.Error)
		if encErr != nil {
			return nil, encErr
		}
		snap.Error = &enc
	}
	return snap, nil
}

type sessionStore[S rezai.SessionState] struct {
	state        rez.AiSessionStateService
	statusSubs   map[string][]chan aix.SnapshotStatus
	statusSubsMu sync.RWMutex
}

func makeSessionStore[S rezai.SessionState](state rez.AiSessionStateService) *sessionStore[S] {
	return &sessionStore[S]{
		state:      state,
		statusSubs: make(map[string][]chan aix.SnapshotStatus),
	}
}

func (s *sessionStore[S]) GetLatestSnapshot(ctx context.Context, sessionID string) (*aix.SessionSnapshot[S], error) {
	runId, idErr := uuid.Parse(sessionID)
	if idErr != nil {
		return nil, fmt.Errorf("invalid session ID: %s", sessionID)
	}
	rs, queryErr := s.state.GetLatestAgentRunSnapshot(ctx, runId)
	if queryErr != nil {
		return nil, fmt.Errorf("lookup snapshot: %w", queryErr)
	}
	return convertToSessionSnapshot[S](rs)
}

func (s *sessionStore[S]) GetSnapshot(ctx context.Context, snapshotID string) (*aix.SessionSnapshot[S], error) {
	id, idErr := uuid.Parse(snapshotID)
	if idErr != nil {
		return nil, fmt.Errorf("invalid snapshot ID: %s", snapshotID)
	}
	rs, queryErr := s.state.GetAgentRunSnapshot(ctx, id)
	if queryErr != nil {
		return nil, fmt.Errorf("lookup snapshot: %w", queryErr)
	}
	return convertToSessionSnapshot[S](rs)
}

func (s *sessionStore[S]) SaveSnapshot(
	ctx context.Context,
	id string,
	setFn func(*aix.SessionSnapshot[S]) (*aix.SessionSnapshot[S], error),
) (*aix.SessionSnapshot[S], error) {
	var snapshotId uuid.UUID
	if id != "" {
		var idErr error
		if snapshotId, idErr = uuid.Parse(id); idErr != nil {
			return nil, fmt.Errorf("invalid ID: %s", id)
		}
	}

	var shouldNotify bool
	var notifyStatus aix.SnapshotStatus
	updateFn := func(rs *ent.AiAgentRunSnapshot, m *ent.AiAgentRunSnapshotMutation) error {
		existing, convErr := convertToSessionSnapshot[S](rs)
		if convErr != nil {
			return fmt.Errorf("convert existing snapshot: %w", convErr)
		}

		snapshot, updateErr := setFn(existing)
		if updateErr != nil {
			return fmt.Errorf("update existing snapshot: %w", updateErr)
		} else if snapshot == nil {
			return nil
		}
		shouldNotify = existing == nil || existing.Status != snapshot.Status
		notifyStatus = snapshot.Status

		runId, runIdErr := uuid.Parse(snapshot.SessionID)
		if runIdErr != nil {
			return fmt.Errorf("invalid session ID: %s", snapshot.SessionID)
		}
		m.SetAiAgentRunID(runId)

		m.SetStatus(aars.Status(snapshot.Status))
		m.SetFinishReason(string(snapshot.FinishReason))
		m.SetCreatedAt(snapshot.CreatedAt)
		m.SetUpdatedAt(snapshot.UpdatedAt)

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
	updated, updateErr := s.state.UpdateAgentRunSnapshot(ctx, snapshotId, updateFn)
	if updateErr != nil {
		return nil, fmt.Errorf("save snapshot: %w", updateErr)
	}
	if shouldNotify {
		s.notifyLocked(updated.ID.String(), notifyStatus)
	}
	return convertToSessionSnapshot[S](updated)
}

func (s *sessionStore[S]) OnSnapshotStatusChange(ctx context.Context, snapshotID string) <-chan aix.SnapshotStatus {
	ch := make(chan aix.SnapshotStatus, 1)

	s.statusSubsMu.Lock()
	snap, snapErr := s.GetSnapshot(ctx, snapshotID)
	if snapErr != nil {
		s.statusSubsMu.Unlock()
		close(ch)
		return ch
	}
	ch <- snap.Status
	s.statusSubs[snapshotID] = append(s.statusSubs[snapshotID], ch)
	s.statusSubsMu.Unlock()

	context.AfterFunc(ctx, func() {
		s.removeSub(snapshotID, ch)
	})

	return ch
}

func (s *sessionStore[State]) removeSub(snapshotID string, ch chan aix.SnapshotStatus) {
	s.statusSubsMu.Lock()
	defer s.statusSubsMu.Unlock()
	subs := s.statusSubs[snapshotID]
	i := slices.Index(subs, ch)
	if i < 0 {
		return
	}
	subs = slices.Delete(subs, i, i+1)
	if len(subs) == 0 {
		delete(s.statusSubs, snapshotID)
	} else {
		s.statusSubs[snapshotID] = subs
	}
	close(ch)
}

func (s *sessionStore[State]) notifyLocked(snapshotID string, status aix.SnapshotStatus) {
	s.statusSubsMu.Lock()
	defer s.statusSubsMu.Unlock()
	for _, ch := range s.statusSubs[snapshotID] {
		select {
		case <-ch:
		default:
		}
		select {
		case ch <- status:
		default:
		}
	}
}
