package postgres

import (
	"context"
	"fmt"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type SystemComponentsService struct {
	db     *ent.Client
	loader rez.ProviderLoader
}

func NewSystemComponentsService(db *ent.Client, pl rez.ProviderLoader) (*SystemComponentsService, error) {
	s := &SystemComponentsService{
		db:     db,
		loader: pl,
	}

	return s, nil
}

func (s *SystemComponentsService) SyncData(ctx context.Context) error {
	prov, provErr := s.loader.LoadSystemComponentsDataProvider(ctx)
	if provErr != nil {
		return fmt.Errorf("loading system components data provider: %w", provErr)
	}
	syncer := newSystemComponentsDataSyncer(s.db, prov)
	return syncer.syncProviderData(ctx)
}
