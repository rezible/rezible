package datasyncer

import (
	"context"
	"fmt"
	"iter"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/oncallevent"
)

func syncOncallEvents(ctx context.Context, db *ent.Client, alerts rez.AlertDataProvider) error {
	b := &oncallEventsBatcher{db: db, alerts: alerts}
	s := newBatchedDataSyncer[*ent.OncallEvent](db, "oncall_events", b)
	return s.Sync(ctx)
}

type oncallEventsBatcher struct {
	db     *ent.Client
	alerts rez.AlertDataProvider
}

func (b *oncallEventsBatcher) setup(ctx context.Context) error {
	return nil
}

func (b *oncallEventsBatcher) pullData(ctx context.Context) iter.Seq2[*ent.OncallEvent, error] {
	from := time.Now().Add(-24 * time.Hour)
	return func(yield func(*ent.OncallEvent, error) bool) {
		for provAlert, pullErr := range b.alerts.PullAlertEventsBetweenDates(ctx, from, time.Now()) {
			if !yield(provAlert, pullErr) {
				return
			}
		}
		// other event data sources
	}
}

func (b *oncallEventsBatcher) createBatchMutations(ctx context.Context, batch []*ent.OncallEvent) ([]ent.Mutation, error) {
	ids := make([]string, len(batch))
	for i, t := range batch {
		ids[i] = t.ProviderID
	}

	dbEvents, queryErr := b.db.OncallEvent.Query().Where(oncallevent.ProviderIDIn(ids...)).All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("querying db events: %w", queryErr)
	}
	dbProvMap := make(map[string]*ent.OncallEvent)
	for _, ev := range dbEvents {
		e := ev
		dbProvMap[e.ProviderID] = e
	}

	var mutations []ent.Mutation
	for _, provEvent := range batch {
		dbEvent, exists := dbProvMap[provEvent.ProviderID]
		if exists {
		}

		evMut := b.syncOncallEvent(dbEvent, provEvent)
		if evMut != nil {
			mutations = append(mutations, evMut)
		}
	}

	return mutations, nil
}

func (b *oncallEventsBatcher) syncOncallEvent(db, prov *ent.OncallEvent) *ent.OncallEventMutation {
	var m *ent.OncallEventMutation
	if db == nil {
		m = b.db.OncallEvent.Create().Mutation()
	} else {
		m = b.db.OncallEvent.UpdateOneID(db.ID).Mutation()

		// TODO: get provider mapping support for fields
		needsSync := db.Title != prov.Title
		if !needsSync {
			return nil
		}
	}

	m.SetProviderID(prov.ProviderID)
	m.SetTimestamp(prov.Timestamp)
	m.SetKind(prov.Kind)
	m.SetTitle(prov.Title)
	m.SetDescription(prov.Description)
	m.SetSource(prov.Source)

	return m
}

func (b *oncallEventsBatcher) getDeletionMutations() []ent.Mutation {
	return nil
}
