package datasync

import (
	"context"
	"fmt"
	"iter"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/alert"
)

func syncAlerts(ctx context.Context, db *ent.Client, prov rez.AlertDataProvider) error {
	b := &alertsBatcher{db: db, provider: prov}
	s := newBatchedDataSyncer[*ent.Alert](db, "alerts", b)
	return s.Sync(ctx)
}

type alertsBatcher struct {
	db       *ent.Client
	provider rez.AlertDataProvider
}

func (b *alertsBatcher) setup(ctx context.Context) error {
	return nil
}

func (b *alertsBatcher) pullData(ctx context.Context) iter.Seq2[*ent.Alert, error] {
	return b.provider.PullAlerts(ctx)
}

func (b *alertsBatcher) createBatchMutations(ctx context.Context, batch []*ent.Alert) ([]ent.Mutation, error) {
	ids := make([]string, len(batch))
	for i, t := range batch {
		ids[i] = t.ProviderID
	}

	dbAlerts, queryErr := b.db.Alert.Query().Where(alert.ProviderIDIn(ids...)).All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("querying alerts: %w", queryErr)
	}
	dbProvMap := make(map[string]*ent.Alert)
	for _, al := range dbAlerts {
		a := al
		dbProvMap[a.ProviderID] = a
	}

	var muts []ent.Mutation
	for _, provAlert := range batch {
		dbAlert, exists := dbProvMap[provAlert.ProviderID]
		if exists {
		}

		var m *ent.AlertMutation

		if dbAlert == nil {
			m = b.db.Alert.Create().Mutation()
		} else {
			m = b.db.Alert.UpdateOneID(dbAlert.ID).Mutation()

			// TODO: get provider mapping support for fields
			needsSync := dbAlert.Title != provAlert.Title
			if !needsSync {
				continue
			}
		}

		m.SetProviderID(provAlert.ProviderID)
		m.SetTitle(provAlert.Title)

		muts = append(muts, m)
	}

	return muts, nil
}

func (b *alertsBatcher) getDeletionMutations() []ent.Mutation {
	return nil
}
