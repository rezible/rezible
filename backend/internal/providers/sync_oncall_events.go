package providers

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/oncallevent"
	"github.com/rs/zerolog/log"
	"time"
)

type oncallEventsDataSyncer struct {
	db       *ent.Client
	provider rez.OncallEventsDataProvider

	mutations []ent.Mutation
}

func newOncallEventsDataSyncer(db *ent.Client, provider rez.OncallEventsDataProvider) *oncallEventsDataSyncer {
	return &oncallEventsDataSyncer{db: db, provider: provider}
}

func (ds *oncallEventsDataSyncer) resetState() {
	ds.mutations = make([]ent.Mutation, 0)
}

func (ds *oncallEventsDataSyncer) SyncProviderData(ctx context.Context) error {
	start := time.Now()

	lastSync := getLastSyncTime(ctx, ds.db, "oncall_events")
	if ds.provider.Source() == "fake" && !lastSync.IsZero() {
		return nil
	}
	if lastSync.Add(time.Minute * 30).After(start) {
		return nil
	}

	if syncErr := ds.syncProviderEvents(ctx, lastSync); syncErr != nil {
		return fmt.Errorf("oncall events: %w", syncErr)
	}

	return nil
}

func (ds *oncallEventsDataSyncer) syncProviderEvents(ctx context.Context, from time.Time) error {
	var batch []*ent.OncallEvent

	start := time.Now()
	var numMutations int

	batchSize := 10
	for provTeam, pullErr := range ds.provider.PullEventsBetweenDates(ctx, from, time.Now()) {
		if pullErr != nil {
			return fmt.Errorf("pull oncall events: %w", pullErr)
		}
		batch = append(batch, provTeam)

		if len(batch) >= batchSize {
			batchMuts, syncErr := ds.syncBatch(ctx, batch)
			if syncErr != nil {
				return fmt.Errorf("failed to sync: %w", syncErr)
			}
			numMutations += batchMuts
			batch = make([]*ent.OncallEvent, 0)
		}
	}

	lastBatchMuts, batchErr := ds.syncBatch(ctx, batch)
	if batchErr != nil {
		return batchErr
	}
	numMutations += lastBatchMuts

	if saveErr := saveSyncHistory(ctx, ds.db, start, numMutations, "oncall_events"); saveErr != nil {
		log.Error().Err(saveErr).Msg("failed to save oncall events data sync history")
	}

	return nil
}

func (ds *oncallEventsDataSyncer) syncBatch(ctx context.Context, batch []*ent.OncallEvent) (int, error) {
	if len(batch) == 0 {
		return 0, nil
	}

	ds.resetState()
	syncErr := ds.createBatchSyncMutations(ctx, batch)
	if syncErr != nil {
		return 0, fmt.Errorf("building mutations: %w", syncErr)
	}

	if applyErr := applySyncMutations(ctx, ds.db, ds.mutations); applyErr != nil {
		return 0, fmt.Errorf("applying mutations: %w", applyErr)
	}
	numMutations := len(ds.mutations)
	ds.resetState()

	return numMutations, nil
}

func (ds *oncallEventsDataSyncer) createBatchSyncMutations(ctx context.Context, batch []*ent.OncallEvent) error {
	ids := make([]string, len(batch))
	for i, t := range batch {
		ids[i] = t.ProviderID
	}

	dbEvents, queryErr := ds.db.OncallEvent.Query().Where(oncallevent.ProviderIDIn(ids...)).All(ctx)
	if queryErr != nil {
		return fmt.Errorf("querying users: %w", queryErr)
	}
	dbProvMap := make(map[string]*ent.OncallEvent)
	for _, ev := range dbEvents {
		e := ev
		dbProvMap[e.ProviderID] = e
	}

	for _, provEvent := range batch {
		dbEvent, exists := dbProvMap[provEvent.ProviderID]
		if exists {

		}
		_, syncErr := ds.syncOncallEvent(ctx, dbEvent, provEvent)
		if syncErr != nil {
			return fmt.Errorf("syncing oncall event: %w", syncErr)
		}
	}

	return nil
}

func (ds *oncallEventsDataSyncer) syncOncallEvent(ctx context.Context, db, prov *ent.OncallEvent) (uuid.UUID, error) {
	var m *ent.OncallEventMutation
	var eventId uuid.UUID
	needsSync := true
	if db == nil {
		eventId = uuid.New()
		m = ds.db.OncallEvent.Create().SetID(eventId).Mutation()
	} else {
		eventId = db.ID
		m = ds.db.OncallEvent.UpdateOneID(eventId).Mutation()

		// TODO: get provider mapping support for fields
		needsSync = db.Title != prov.Title
	}

	m.SetProviderID(prov.ProviderID)
	m.SetTimestamp(prov.Timestamp)
	m.SetKind(prov.Kind)
	m.SetTitle(prov.Title)
	m.SetDescription(prov.Description)
	m.SetSource(prov.Source)

	if needsSync {
		ds.mutations = append(ds.mutations, m)
	}

	return eventId, nil
}
