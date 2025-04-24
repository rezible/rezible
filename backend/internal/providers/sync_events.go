package providers

import (
	"context"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type oncallEventsDataSyncer struct {
	db       *ent.Client
	provider rez.OncallEventsDataProvider

	mutations []ent.Mutation
}

func newOncallEventsDataSyncer(db *ent.Client, provider rez.OncallEventsDataProvider) *oncallEventsDataSyncer {
	return &oncallEventsDataSyncer{db: db, provider: provider}
}

func (as *oncallEventsDataSyncer) SyncProviderData(ctx context.Context) error {
	
	return nil
}
