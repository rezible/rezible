package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type PlaybookService struct {
	db rez.Database
}

func NewPlaybookService(db rez.Database) (*PlaybookService, error) {
	s := &PlaybookService{
		db: db,
	}

	return s, nil
}

func (s *PlaybookService) ListPlaybooks(ctx context.Context, params rez.ListPlaybooksParams) ([]*ent.Playbook, int, error) {
	query := s.db.Client(ctx).Playbook.Query().
		Where()

	qCtx := params.GetQueryContext(ctx)
	count, queryErr := query.Count(qCtx)
	if queryErr != nil {
		return nil, 0, fmt.Errorf("count: %w", queryErr)
	}
	playbooks := make([]*ent.Playbook, 0)
	if count > 0 {
		playbooks, queryErr = query.All(qCtx)
	}
	if queryErr != nil {
		return nil, 0, fmt.Errorf("query: %w", queryErr)
	}
	return playbooks, count, nil
}

func (s *PlaybookService) GetPlaybook(ctx context.Context, id uuid.UUID) (*ent.Playbook, error) {
	return s.db.Client(ctx).Playbook.Get(ctx, id)
}

type saveablePlaybookQuery interface {
	Save(context.Context) (*ent.Playbook, error)
}

func (s *PlaybookService) SetPlaybook(ctx context.Context, playbook *ent.Playbook) (*ent.Playbook, error) {
	var q saveablePlaybookQuery
	if playbook.ID == uuid.Nil {
		q = s.db.Client(ctx).Playbook.Create().
			SetTitle(playbook.Title).
			SetContent(playbook.Content)
	} else {
		q = s.db.Client(ctx).Playbook.UpdateOneID(playbook.ID).
			SetTitle(playbook.Title).
			SetContent(playbook.Content)
	}
	return q.Save(ctx)
}
