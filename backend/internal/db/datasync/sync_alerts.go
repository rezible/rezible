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
	"github.com/rezible/rezible/ent/alertinstance"
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
		ids[i] = t.ExternalID
	}

	dbAlerts, queryErr := b.db.Alert.Query().Where(alert.ExternalIDIn(ids...)).All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("querying alerts: %w", queryErr)
	}
	dbProvMap := make(map[string]*ent.Alert)
	for _, al := range dbAlerts {
		a := al
		dbProvMap[a.ExternalID] = a
	}

	var muts []ent.Mutation
	for _, provAlert := range batch {
		dbAlert, exists := dbProvMap[provAlert.ExternalID]
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

		m.SetExternalID(provAlert.ExternalID)
		m.SetTitle(provAlert.Title)

		muts = append(muts, m)
	}

	return muts, nil
}

func (b *alertsBatcher) getDeletionMutations() []ent.Mutation {
	return nil
}

func syncAlertInstances(ctx context.Context, db *ent.Client, alerts rez.AlertDataProvider) error {
	b := &alertInstancesBatcher{db: db, alerts: alerts}
	s := newBatchedDataSyncer[*ent.AlertInstance](db, "alert_instance", b)
	s.setSyncInterval(time.Hour * 12)
	return s.Sync(ctx)
}

type alertInstancesBatcher struct {
	db     *ent.Client
	alerts rez.AlertDataProvider
}

func (b *alertInstancesBatcher) setup(ctx context.Context) error {
	return nil
}

func (b *alertInstancesBatcher) pullData(ctx context.Context) iter.Seq2[*ent.AlertInstance, error] {
	from := time.Now().Add(-24 * time.Hour)
	// TODO: pass in last sync info to pullData
	return b.alerts.PullAlertInstancesBetweenDates(ctx, from, time.Now())
}

func (b *alertInstancesBatcher) createBatchMutations(ctx context.Context, batch []*ent.AlertInstance) ([]ent.Mutation, error) {
	provIds := make([]string, len(batch))
	alertProvIds := make([]string, 0, len(batch))
	for i, t := range batch {
		provIds[i] = t.ExternalID
		if a := t.Edges.Alert; a != nil && a.ExternalID != "" {
			alertProvIds = append(alertProvIds, a.ExternalID)
		}
	}

	dbProvAlertMap := make(map[string]*ent.Alert)
	if len(alertProvIds) > 0 {
		dbAlerts, alertsQueryErr := b.db.Alert.Query().
			Where(alert.ExternalIDIn(alertProvIds...)).
			All(ctx)
		if alertsQueryErr != nil {
			return nil, fmt.Errorf("querying db alerts: %w", alertsQueryErr)
		}
		for _, al := range dbAlerts {
			a := al
			dbProvAlertMap[a.ExternalID] = a
		}
	}

	dbInstances, instancesQueryErr := b.db.AlertInstance.Query().
		Where(alertinstance.ExternalIDIn(provIds...)).
		WithEvent().
		All(ctx)
	if instancesQueryErr != nil {
		return nil, fmt.Errorf("querying db instances: %w", instancesQueryErr)
	}

	dbProvMap := make(map[string]*ent.AlertInstance)
	for _, ins := range dbInstances {
		i := ins
		dbProvMap[i.ExternalID] = ins
	}

	var mutations []ent.Mutation
	for _, prov := range batch {
		db, exists := dbProvMap[prov.ExternalID]
		if exists {
		}

		if provAlert := prov.Edges.Alert; provAlert != nil {
			if dbAlert, alertExists := dbProvAlertMap[provAlert.ExternalID]; alertExists {
				prov.AlertID = dbAlert.ID
			}
		}

		evMut, evId := b.syncEvent(db.Edges.Event, prov.Edges.Event)
		if evMut != nil {
			mutations = append(mutations, evMut)
		}
		prov.EventID = evId

		if mut := b.syncInstance(db, prov); mut != nil {
			mutations = append(mutations, mut)
		}
	}

	return mutations, nil
}

func (b *alertInstancesBatcher) syncInstance(db, prov *ent.AlertInstance) *ent.AlertInstanceMutation {
	var m *ent.AlertInstanceMutation
	if db == nil {
		m = b.db.AlertInstance.Create().Mutation()
	} else {
		m = b.db.AlertInstance.UpdateOneID(db.ID).Mutation()

		// TODO: get provider mapping support for fields
		needsSync := false
		if !needsSync {
			return nil
		}
	}

	m.SetExternalID(prov.ExternalID)
	if prov.AlertID != uuid.Nil {
		m.SetAlertID(prov.AlertID)
	}
	m.SetEventID(prov.EventID)

	return m
}

func (b *alertInstancesBatcher) syncEvent(db, prov *ent.Event) (*ent.EventMutation, uuid.UUID) {
	var m *ent.EventMutation
	var id uuid.UUID
	if db == nil {
		id = uuid.New()
		m = b.db.Event.Create().SetID(id).Mutation()
	} else {
		id = db.ID
		m = b.db.Event.UpdateOneID(id).Mutation()

		// TODO: get provider mapping support for fields
		needsSync := db.Title != prov.Title || db.Description != prov.Description
		if !needsSync {
			return nil, id
		}
	}
	m.SetTimestamp(prov.Timestamp)
	m.SetKind(prov.Kind)
	m.SetTitle(prov.Title)
	m.SetDescription(prov.Description)
	m.SetSource(prov.Source)

	return m, id
}

func (b *alertInstancesBatcher) getDeletionMutations() []ent.Mutation {
	return nil
}
