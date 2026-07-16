package ai

import (
	"embed"
	"encoding/json"
	"fmt"
	"hash/maphash"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/firebase/genkit/go/ai"
	aix "github.com/firebase/genkit/go/ai/exp"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	aars "github.com/rezible/rezible/ent/aiagentrunsnapshot"
)

//go:embed prompts
var PromptsDir embed.FS

func SessionSnapshotFromEnt[S SessionState](rs *ent.AiAgentRunSnapshot) (*aix.SessionSnapshot[S], error) {
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

func SessionSnapshotToEnt[S SessionState](s *aix.SessionSnapshot[S]) (*ent.AiAgentRunSnapshot, error) {
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

//func GetChangedSnapshotArtifactParts(updated *ent.AiAgentRunSnapshot, prev *ent.AiAgentRunSnapshot, artifactName string) ([]*ai.Part, error) {
//	getOutputArtifactParts := func(snap *ent.AiAgentRunSnapshot) ([]*ai.Part, error) {
//		if snap == nil || snap.State == nil {
//			return nil, nil
//		}
//		var sa struct {
//			State struct {
//				Artifacts []*aix.Artifact `json:"artifacts"`
//			} `json:"state"`
//		}
//		if jsonErr := json.Unmarshal(*snap.State, &sa); jsonErr != nil {
//			return nil, fmt.Errorf("unmarshal snapshot state: %w", jsonErr)
//		}
//		for _, a := range sa.State.Artifacts {
//			if a.Name == artifactName {
//				return a.Parts, nil
//			}
//		}
//		return nil, nil
//	}
//
//	newParts, newPartsErr := getOutputArtifactParts(updated)
//	if newPartsErr != nil {
//		return nil, fmt.Errorf("get new parts: %w", newPartsErr)
//	}
//	prevParts, prevPartsErr := getOutputArtifactParts(prev)
//	if prevPartsErr != nil {
//		return nil, fmt.Errorf("get new parts: %w", prevPartsErr)
//	}

func GetChangedSnapshotArtifactParts[S SessionState](prev *aix.SessionSnapshot[S], updated *aix.SessionSnapshot[S], artifactName string) ([]*ai.Part, error) {
	getOutputArtifactParts := func(snap *aix.SessionSnapshot[S]) []*ai.Part {
		if snap != nil && snap.State != nil {
			for _, a := range snap.State.Artifacts {
				if a.Name == artifactName {
					return a.Parts
				}
			}
		}
		return nil
	}

	prevParts := getOutputArtifactParts(prev)
	newParts := getOutputArtifactParts(updated)
	if len(prevParts) == 0 {
		return newParts, nil
	}

	hashSeed := maphash.MakeSeed()
	hashPart := func(p *ai.Part) (uint64, error) {
		pb, jsonErr := p.MarshalJSON()
		if jsonErr != nil {
			return 0, fmt.Errorf("marshal part: %w", jsonErr)
		}
		return maphash.Bytes(hashSeed, pb), nil
	}

	prevHashes := mapset.NewThreadUnsafeSet[uint64]()
	for _, p := range prevParts {
		hash, hashErr := hashPart(p)
		if hashErr != nil {
			return nil, fmt.Errorf("hash prev part: %w", hashErr)
		}
		prevHashes.Add(hash)
	}

	var changedParts []*ai.Part
	for _, p := range newParts {
		hash, hashErr := hashPart(p)
		if hashErr != nil {
			return nil, fmt.Errorf("hash new part: %w", hashErr)
		}
		if !prevHashes.Contains(hash) {
			changedParts = append(changedParts, p)
		}
	}
	return changedParts, nil
}
