package ai

import (
	"embed"
	"encoding/json"
	"fmt"

	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	aars "github.com/rezible/rezible/ent/aiagentrunsnapshot"
)

//go:embed prompts
var PromptsDir embed.FS

func SessionSnapshotFromEnt[S SessionState](rs *ent.AiAgentRunSnapshot) (*aix.SessionSnapshot[S], error) {
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

func SessionSnapshotToEnt[S SessionState](s *aix.SessionSnapshot[S]) (*ent.AiAgentRunSnapshot, error) {
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
