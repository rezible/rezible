package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type PlaybookService struct {
	db *ent.Client
}

func NewPlaybookService(db *ent.Client) (*PlaybookService, error) {
	s := &PlaybookService{
		db: db,
	}

	return s, nil
}

func (s *PlaybookService) ListPlaybooks(ctx context.Context, params rez.ListPlaybooksParams) ([]*ent.Playbook, int, error) {
	query := s.db.Playbook.Query().
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
	return s.db.Playbook.Get(ctx, id)
}

type saveablePlaybookQuery interface {
	Save(context.Context) (*ent.Playbook, error)
}

func (s *PlaybookService) UpdatePlaybook(ctx context.Context, playbook *ent.Playbook) (*ent.Playbook, error) {
	var q saveablePlaybookQuery
	if playbook.ID == uuid.Nil {
		q = s.db.Playbook.Create().
			SetTitle(playbook.Title).
			SetContent(playbook.Content)
	} else {
		q = s.db.Playbook.UpdateOneID(playbook.ID).
			SetTitle(playbook.Title).
			SetContent(playbook.Content)
	}
	return q.Save(ctx)
}
