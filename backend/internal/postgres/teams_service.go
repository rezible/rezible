package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type TeamService struct {
	db     *ent.Client
	loader rez.ProviderLoader
}

func NewTeamService(db *ent.Client, pl rez.ProviderLoader) (*TeamService, error) {
	s := &TeamService{
		db:     db,
		loader: pl,
	}

	return s, nil
}

func (s *TeamService) SyncData(ctx context.Context) error {
	prov, provErr := s.loader.LoadTeamDataProvider(ctx)
	if provErr != nil {
		return fmt.Errorf("loading user data provider: %w", provErr)
	}
	syncer := newTeamDataSyncer(s.db, prov)
	return syncer.syncProviderData(ctx)
}

func (s *TeamService) GetById(ctx context.Context, id uuid.UUID) (*ent.Team, error) {
	return s.db.Team.Get(ctx, id)
}
