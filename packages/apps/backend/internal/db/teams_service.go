package db

import (
	"context"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"

	"github.com/rezible/rezible/ent"
	team "github.com/rezible/rezible/ent/team"
	tm "github.com/rezible/rezible/ent/teammembership"
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

func (s *TeamService) List(ctx context.Context, p rez.ListTeamsParams) (ent.Teams, error) {
	query := s.db.Team.Query()
	if len(p.TeamIds) > 0 {
		query = query.Where(team.IDIn(p.TeamIds...))
	}
	if len(p.UserIds) > 0 {
		query = query.Where(team.HasTeamMembershipsWith(tm.UserIDIn(p.UserIds...)))
	}
	return query.All(ctx)
}
