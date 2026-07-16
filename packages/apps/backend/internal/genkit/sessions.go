package genkit

import (
	"context"
	"encoding/json"
	"fmt"

	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	aars "github.com/rezible/rezible/ent/aiagentrunsnapshot"
	rezai "github.com/rezible/rezible/pkg/ai"
)

type sessionStore[S rezai.SessionState] struct {
	snapshots rez.AiAgentSnapshotService
}

func makeSessionStore[S rezai.SessionState](snapshots rez.AiAgentSnapshotService) *sessionStore[S] {
	return &sessionStore[S]{snapshots: snapshots}
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

func (s *sessionStore[S]) SaveSnapshot(
	ctx context.Context,
	sessionId string,
	updateFn func(*aix.SessionSnapshot[S]) (*aix.SessionSnapshot[S], error),
) (*aix.SessionSnapshot[S], error) {
	updateSnapshotTxFn := func(currSnap *ent.AiAgentRunSnapshot, m *ent.AiAgentRunSnapshotMutation) (*rez.AiAgentSnapshotDelta, error) {
		existing, convExistingErr := rezai.SessionSnapshotFromEnt[S](currSnap)
		if convExistingErr != nil {
			return nil, fmt.Errorf("convert existing snapshot: %w", convExistingErr)
		}

		updated, updateErr := updateFn(existing)
		if updateErr != nil {
			return nil, fmt.Errorf("update existing snapshot: %w", updateErr)
		}

		if updated == nil { // return nil = skip write
			return nil, nil
		}

		var delta rez.AiAgentSnapshotDelta

		if existing == nil {
			runId, runIdErr := uuid.Parse(updated.SessionID)
			if runIdErr != nil {
				return nil, fmt.Errorf("invalid session ID: %s", updated.SessionID)
			}
			m.SetAiAgentRunID(runId)

			if len(updated.ParentID) > 0 {
				parentId, parentIdErr := uuid.Parse(updated.ParentID)
				if parentIdErr != nil {
					return nil, fmt.Errorf("invalid parent ID: %s", updated.ParentID)
				}
				m.SetParentID(parentId)
			}
		}

		m.SetCreatedAt(updated.CreatedAt)
		m.SetUpdatedAt(updated.UpdatedAt)

		status := updated.Status
		if status == "" { // empty status is treated as completed
			status = aix.SnapshotStatusCompleted
		}
		if existing == nil || existing.Status != updated.Status {
			delta.Status = &status
		}
		m.SetStatus(aars.Status(status))
		m.SetFinishReason(string(updated.FinishReason))

		if updated.HeartbeatAt != nil {
			m.SetHeartbeatAt(*updated.HeartbeatAt)
		}

		if updated.Error != nil {
			errBytes, jsonErr := json.Marshal(updated.Error)
			if jsonErr != nil {
				return nil, fmt.Errorf("marshal snapshot error: %w", jsonErr)
			}
			m.SetError(errBytes)
		}

		if updated.State != nil {
			outputParts, outputPartsErr := rezai.GetChangedSnapshotArtifactParts(existing, updated, rezai.OutputArtifactName)
			if outputPartsErr != nil {
				return nil, fmt.Errorf("get changed output parts: %w", outputPartsErr)
			}
			delta.OutputParts = outputParts

			stateBytes, jsonErr := json.Marshal(updated.State)
			if jsonErr != nil {
				return nil, fmt.Errorf("marshal snapshot state: %w", jsonErr)
			}
			m.SetState(stateBytes)
		}

		return &delta, nil
	}

	var snapshotId uuid.UUID
	if sessionId != "" {
		var idErr error
		if snapshotId, idErr = uuid.Parse(sessionId); idErr != nil {
			return nil, fmt.Errorf("invalid ID: %s", sessionId)
		}
	}

	updated, updateErr := s.snapshots.UpdateAgentRunSnapshot(ctx, snapshotId, updateSnapshotTxFn)
	if updateErr != nil {
		return nil, fmt.Errorf("save snapshot: %w", updateErr)
	}

	return rezai.SessionSnapshotFromEnt[S](updated)
}

func (s *sessionStore[S]) OnSnapshotStatusChange(ctx context.Context, snapshotID string) <-chan aix.SnapshotStatus {
	id, _ := uuid.Parse(snapshotID)
	return s.snapshots.OnSnapshotStatusChange(ctx, id)
}
