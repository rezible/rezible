package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ars "github.com/rezible/rezible/ent/agentrunsnapshot"
)

type AgentRunSnapshotService struct {
	db rez.Database
}

func NewAgentRunSnapshotService(db rez.Database) (*AgentRunSnapshotService, error) {
	s := &AgentRunSnapshotService{
		db: db,
	}
	return s, nil
}

func (s *AgentRunSnapshotService) GetLatestSnapshot(ctx context.Context, runId uuid.UUID) (*ent.AgentRunSnapshot, error) {
	query := s.db.Client(ctx).AgentRunSnapshot.Query().
		Where(ars.AgentRunID(runId)).
		Order(ars.ByCreatedAt()).
		Limit(1)
	res, resErr := query.Only(ctx)
	if resErr != nil && !ent.IsNotFound(resErr) {
		return nil, resErr
	}
	return res, nil
}

func (s *AgentRunSnapshotService) GetSnapshot(ctx context.Context, id uuid.UUID) (*ent.AgentRunSnapshot, error) {
	res, resErr := s.db.Client(ctx).AgentRunSnapshot.Get(ctx, id)
	if resErr != nil && !ent.IsNotFound(resErr) {
		return nil, resErr
	}
	return res, nil
}

func (s *AgentRunSnapshotService) UpdateSnapshot(ctx context.Context, id uuid.UUID, setFn func(*ent.AgentRunSnapshot, *ent.AgentRunSnapshotMutation) error) (*ent.AgentRunSnapshot, error) {
	var snapshot *ent.AgentRunSnapshot
	return snapshot, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		var curr *ent.AgentRunSnapshot
		var mutator ent.EntityMutator[*ent.AgentRunSnapshot, *ent.AgentRunSnapshotMutation]
		if id != uuid.Nil {
			var getErr error
			if curr, getErr = tx.AgentRunSnapshot.Get(ctx, id); getErr != nil {
				return fmt.Errorf("failed to lookup existing snapshot: %w", getErr)
			}
			mutator = curr.Update()
		} else {
			mutator = tx.AgentRunSnapshot.Create()
		}
		m := mutator.Mutation()
		if setErr := setFn(curr, m); setErr != nil {
			return setErr
		}
		if len(m.Fields()) > 0 {
			saved, saveErr := mutator.Save(ctx)
			if saveErr != nil {
				return fmt.Errorf("failed to save: %w", saveErr)
			}
			snapshot = saved.Unwrap()
		} else {
			if curr == nil {
				return fmt.Errorf("no fields changed, no existing snapshot")
			}
			snapshot = curr.Unwrap()
		}
		return nil
	})
}
