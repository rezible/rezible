package db

import (
	"context"

	"github.com/google/uuid"

	"github.com/rezible/rezible/ent"
)

type TeamService struct {
	db *ent.Client
}

func NewTeamService(db *ent.Client) (*TeamService, error) {
	s := &TeamService{
		db: db,
	}

	return s, nil
}

func (s *TeamService) GetById(ctx context.Context, id uuid.UUID) (*ent.Team, error) {
	return s.db.Team.Get(ctx, id)
}
