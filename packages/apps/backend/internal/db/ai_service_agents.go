package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	ars "github.com/rezible/rezible/ent/aiagentrunsnapshot"
)

type AiAgentRunSnapshotService struct {
	db rez.Database
}

func NewAgentRunSnapshotService(db rez.Database) (*AiAgentRunSnapshotService, error) {
	s := &AiAgentRunSnapshotService{
		db: db,
	}
	return s, nil
}

func (s *AiAgentRunSnapshotService) GetLatestAgentRunSnapshot(ctx context.Context, runId uuid.UUID) (*ent.AiAgentRunSnapshot, error) {
	query := s.db.Client(ctx).AiAgentRunSnapshot.Query().
		Where(ars.AiAgentRunID(runId)).
		Order(ars.ByCreatedAt()).
		Limit(1)
	res, resErr := query.Only(ctx)
	if resErr != nil && !ent.IsNotFound(resErr) {
		return nil, resErr
	}
	return res, nil
}

func (s *AiAgentRunSnapshotService) GetAgentRunSnapshot(ctx context.Context, id uuid.UUID) (*ent.AiAgentRunSnapshot, error) {
	res, resErr := s.db.Client(ctx).AiAgentRunSnapshot.Get(ctx, id)
	if resErr != nil && !ent.IsNotFound(resErr) {
		return nil, resErr
	}
	return res, nil
}

func (s *AiAgentRunSnapshotService) SetAgentRunSnapshot(ctx context.Context, id uuid.UUID, setFn func(*ent.AiAgentRunSnapshotMutation)) (*ent.AiAgentRunSnapshot, error) {
	var snapshot *ent.AiAgentRunSnapshot
	return snapshot, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		var mutator ent.EntityMutator[*ent.AiAgentRunSnapshot, *ent.AiAgentRunSnapshotMutation]
		if id != uuid.Nil {
			mutator = tx.AiAgentRunSnapshot.UpdateOneID(id)
		} else {
			mutator = tx.AiAgentRunSnapshot.Create()
		}
		setFn(mutator.Mutation())
		saved, saveErr := mutator.Save(ctx)
		if saveErr != nil {
			return fmt.Errorf("failed to save: %w", saveErr)
		}
		snapshot = saved.Unwrap()
		return nil
	})
}

func (s *AiAgentRunSnapshotService) UpdateAgentRunSnapshot(ctx context.Context, id uuid.UUID, setFn func(*ent.AiAgentRunSnapshot, *ent.AiAgentRunSnapshotMutation) error) (*ent.AiAgentRunSnapshot, error) {
	var snapshot *ent.AiAgentRunSnapshot
	return snapshot, s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		var curr *ent.AiAgentRunSnapshot
		var mutator ent.EntityMutator[*ent.AiAgentRunSnapshot, *ent.AiAgentRunSnapshotMutation]
		if id != uuid.Nil {
			var getErr error
			if curr, getErr = tx.AiAgentRunSnapshot.Get(ctx, id); getErr != nil && !ent.IsNotFound(getErr) {
				return fmt.Errorf("failed to lookup existing (%s): %w", id, getErr)
			}
		}
		if curr != nil {
			mutator = curr.Update()
		} else {
			mutator = tx.AiAgentRunSnapshot.Create()
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
