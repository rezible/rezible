package db

import (
	"context"
	"strings"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/alert"
	"github.com/rezible/rezible/ent/playbook"
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

func (s *PlaybookService) ListPlaybooks(ctx context.Context, params rez.ListPlaybooksParams) (*ent.ListResult[ent.Playbook], error) {
	query := s.db.Client(ctx).Playbook.Query().
		Order(playbook.ByID(params.GetOrder()))
	if search := strings.TrimSpace(params.Search); search != "" {
		query.Where(playbook.TitleContainsFold(search))
	}
	if params.AlertID != uuid.Nil {
		query.Where(playbook.HasAlertsWith(alert.ID(params.AlertID)))
	}
	return ent.DoListQuery[ent.Playbook, *ent.PlaybookQuery](ctx, query, params.ListParams)
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
