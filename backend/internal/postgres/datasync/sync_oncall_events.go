package datasync

import (
	"context"
	"fmt"
	"iter"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/alert"
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
	alertProvIds := make([]string, 0, len(batch))
	for i, t := range batch {
		ids[i] = t.ProviderID
		if a := t.Edges.Alert; a != nil && a.ProviderID != "" {
			alertProvIds = append(alertProvIds, a.ProviderID)
		}
	}

	dbEvents, eventsQueryErr := b.db.OncallEvent.Query().
		Where(oncallevent.ProviderIDIn(ids...)).
		All(ctx)
	if eventsQueryErr != nil {
		return nil, fmt.Errorf("querying db events: %w", eventsQueryErr)
	}
	dbProvMap := make(map[string]*ent.OncallEvent)
	for _, ev := range dbEvents {
		e := ev
		dbProvMap[e.ProviderID] = e
	}

	dbProvAlertMap := make(map[string]*ent.Alert)
	if len(alertProvIds) > 0 {
		dbAlerts, alertsQueryErr := b.db.Alert.Query().
			Where(alert.ProviderIDIn(alertProvIds...)).
			All(ctx)
		if alertsQueryErr != nil {
			return nil, fmt.Errorf("querying db alerts: %w", alertsQueryErr)
		}
		for _, a := range dbAlerts {
			dbProvAlertMap[a.ProviderID] = a
		}
	}

	var mutations []ent.Mutation
	for _, provEvent := range batch {
		dbEvent, exists := dbProvMap[provEvent.ProviderID]
		if exists {
		}

		if provAlert := provEvent.Edges.Alert; provAlert != nil {
			if dbAlert, alertExists := dbProvAlertMap[provAlert.ProviderID]; alertExists {
				provEvent.AlertID = dbAlert.ID
			}
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
	if prov.AlertID != uuid.Nil {
		m.SetAlertID(prov.AlertID)
	}
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
