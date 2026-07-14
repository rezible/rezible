package genkit

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/maphash"
	"slices"
	"sync"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	aars "github.com/rezible/rezible/ent/aiagentrunsnapshot"
	rezai "github.com/rezible/rezible/pkg/ai"
)

type sessionStore[S rezai.SessionState] struct {
	snapshots    rez.AiAgentSnapshotService
	statusSubs   map[string][]chan aix.SnapshotStatus
	statusSubsMu sync.RWMutex
}

func makeSessionStore[S rezai.SessionState](snapshots rez.AiAgentSnapshotService) *sessionStore[S] {
	return &sessionStore[S]{
		snapshots:  snapshots,
		statusSubs: make(map[string][]chan aix.SnapshotStatus),
	}
}

func (s *sessionStore[S]) GetLatestSnapshot(ctx context.Context, sessionID string) (*aix.SessionSnapshot[S], error) {
	runId, idErr := uuid.Parse(sessionID)
	if idErr != nil {
		return nil, fmt.Errorf("invalid session ID: %s", sessionID)
	}
	rs, queryErr := s.snapshots.GetLatestSnapshotForRun(ctx, runId)
	if queryErr != nil {
		return nil, fmt.Errorf("lookup snapshot: %w", queryErr)
	}
	if rs == nil {
		return nil, nil
	}
	return rezai.SessionSnapshotFromEnt[S](rs)
}

func (s *sessionStore[S]) GetSnapshot(ctx context.Context, snapshotID string) (*aix.SessionSnapshot[S], error) {
	id, idErr := uuid.Parse(snapshotID)
	if idErr != nil {
		return nil, fmt.Errorf("invalid snapshot ID: %s", snapshotID)
	}
	rs, queryErr := s.snapshots.GetAgentRunSnapshot(ctx, id)
	if queryErr != nil {
		return nil, fmt.Errorf("lookup snapshot: %w", queryErr)
	}
	return rezai.SessionSnapshotFromEnt[S](rs)
}

func (s *sessionStore[S]) getOutputArtifact(snap *aix.SessionSnapshot[S]) *aix.Artifact {
	if snap == nil || snap.State == nil {
		return nil
	}
	for _, a := range snap.State.Artifacts {
		if a.Name == "output" {
			return a
		}
	}
	return nil
}

func (s *sessionStore[S]) getChangedParts(newSnap *aix.SessionSnapshot[S], prevSnap *aix.SessionSnapshot[S]) ([]*ai.Part, error) {
	currOutput := s.getOutputArtifact(newSnap)
	if currOutput == nil {
		return nil, nil
	}

	hashSeed := maphash.MakeSeed()
	prevHashes := mapset.NewThreadUnsafeSet[uint64]()
	if prev := s.getOutputArtifact(prevSnap); prev != nil {
		for _, art := range prev.Parts {
			pb, jsonErr := art.MarshalJSON()
			if jsonErr != nil {
				return nil, fmt.Errorf("marshal part: %w", jsonErr)
			}
			prevHashes.Add(maphash.Bytes(hashSeed, pb))
		}
	}

	var changed []*ai.Part
	for _, part := range currOutput.Parts {
		pb, jsonErr := part.MarshalJSON()
		if jsonErr != nil {
			return nil, fmt.Errorf("marshal part: %w", jsonErr)
		}
		if !prevHashes.Contains(maphash.Bytes(hashSeed, pb)) {
			changed = append(changed, part)
		}
	}
	return changed, nil
}

type snapshotWriteSetter[S rezai.SessionState] = func(*aix.SessionSnapshot[S]) (*aix.SessionSnapshot[S], error)

func (s *sessionStore[S]) SaveSnapshot(ctx context.Context, id string, setFn snapshotWriteSetter[S]) (*aix.SessionSnapshot[S], error) {
	var notifyStatus *aix.SnapshotStatus
	updateFn := func(rs *ent.AiAgentRunSnapshot, m *ent.AiAgentRunSnapshotMutation) ([]*ai.Part, error) {
		var existing *aix.SessionSnapshot[S]
		if rs != nil {
			var convErr error
			if existing, convErr = rezai.SessionSnapshotFromEnt[S](rs); convErr != nil {
				return nil, fmt.Errorf("convert existing snapshot: %w", convErr)
			}
		}

		snapshot, updateErr := setFn(existing)
		if updateErr != nil {
			return nil, fmt.Errorf("update existing snapshot: %w", updateErr)
		}
		if snapshot == nil {
			return nil, nil
		}

		runId, runIdErr := uuid.Parse(snapshot.SessionID)
		if runIdErr != nil {
			return nil, fmt.Errorf("invalid session ID: %s", snapshot.SessionID)
		}
		m.SetAiAgentRunID(runId)

		if existing == nil || existing.Status != snapshot.Status {
			notifyStatus = &snapshot.Status
		}

		m.SetStatus(aars.Status(snapshot.Status))
		m.SetFinishReason(string(snapshot.FinishReason))
		m.SetCreatedAt(snapshot.CreatedAt)
		m.SetUpdatedAt(snapshot.UpdatedAt)

		if len(snapshot.ParentID) > 0 {
			parentId, parentIdErr := uuid.Parse(snapshot.ParentID)
			if parentIdErr != nil {
				return nil, fmt.Errorf("invalid parent ID: %s", snapshot.ParentID)
			}
			m.SetParentID(parentId)
		}

		if snapshot.HeartbeatAt != nil {
			m.SetHeartbeatAt(*snapshot.HeartbeatAt)
		}

		if snapshot.Error != nil {
			sessErr, jsonErr := json.Marshal(snapshot.State)
			if jsonErr != nil {
				return nil, fmt.Errorf("marshal error: %w", jsonErr)
			}
			m.SetError(sessErr)
		}

		var outputs []*ai.Part
		if snapshot.State != nil {
			state, jsonErr := json.Marshal(snapshot.State)
			if jsonErr != nil {
				return nil, fmt.Errorf("marshal state: %w", jsonErr)
			}
			m.SetState(state)

			changedOutputs, changedErr := s.getChangedParts(snapshot, existing)
			if changedErr != nil {
				return nil, fmt.Errorf("changed outputs: %w", changedErr)
			}
			outputs = changedOutputs
		}

		return outputs, nil
	}

	var snapshotId uuid.UUID
	if id != "" {
		var idErr error
		if snapshotId, idErr = uuid.Parse(id); idErr != nil {
			return nil, fmt.Errorf("invalid ID: %s", id)
		}
	}

	updated, updateErr := s.snapshots.SetAgentRunSnapshot(ctx, snapshotId, updateFn)
	if updateErr != nil {
		return nil, fmt.Errorf("save snapshot: %w", updateErr)
	}

	if notifyStatus != nil {
		s.notifyLocked(updated.ID.String(), *notifyStatus)
	}

	return rezai.SessionSnapshotFromEnt[S](updated)
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
