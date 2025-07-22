package providers

import (
	"context"
	"fmt"
	"github.com/rezible/rezible/ent/alert"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type alertDataSyncer struct {
	db       *ent.Client
	provider rez.AlertDataProvider

	mutations []ent.Mutation
}

func newAlertDataSyncer(db *ent.Client, prov rez.AlertDataProvider) *alertDataSyncer {
	ds := &alertDataSyncer{db: db, provider: prov}
	ds.resetState()
	return ds
}

func (ds *alertDataSyncer) resetState() {
	ds.mutations = make([]ent.Mutation, 0)
}

func (ds *alertDataSyncer) SyncProviderData(ctx context.Context) error {
	start := time.Now()

	lastSync := getLastSyncTime(ctx, ds.db, "alerts")
	if lastSync.Add(time.Minute * 30).After(start) {
		return nil
	}

	if usersErr := ds.syncAllProviderAlerts(ctx); usersErr != nil {
		return fmt.Errorf("alerts: %w", usersErr)
	}
	log.Info().Msg("alerts data sync complete")

	return nil
}

func (ds *alertDataSyncer) syncAllProviderAlerts(ctx context.Context) error {
	var batch []*ent.Alert

	start := time.Now()
	var numMutations int

	batchSize := 10
	for provAlert, pullErr := range ds.provider.PullAlerts(ctx) {
		if pullErr != nil {
			return fmt.Errorf("pull alerts: %w", pullErr)
		}
		batch = append(batch, provAlert)

		if len(batch) >= batchSize {
			batchMuts, syncErr := ds.syncBatch(ctx, batch)
			if syncErr != nil {
				return syncErr
			}
			numMutations += batchMuts
			batch = make([]*ent.Alert, 0)
		}
	}

	lastBatchMuts, batchErr := ds.syncBatch(ctx, batch)
	if batchErr != nil {
		return batchErr
	}
	numMutations += lastBatchMuts

	if saveErr := saveSyncHistory(ctx, ds.db, start, numMutations, "alerts"); saveErr != nil {
		log.Error().Err(saveErr).Msg("failed to save alerts data sync history")
	}

	return nil
}

func (ds *alertDataSyncer) syncBatch(ctx context.Context, batch []*ent.Alert) (int, error) {
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

func (ds *alertDataSyncer) createBatchSyncMutations(ctx context.Context, batch []*ent.Alert) error {
	ids := make([]string, len(batch))
	for i, t := range batch {
		ids[i] = t.ProviderID
	}

	dbAlerts, queryErr := ds.db.Alert.Query().Where(alert.ProviderIDIn(ids...)).All(ctx)
	if queryErr != nil {
		return fmt.Errorf("querying alerts: %w", queryErr)
	}
	dbProvMap := make(map[string]*ent.Alert)
	for _, al := range dbAlerts {
		a := al
		dbProvMap[a.ProviderID] = a
	}

	for _, provAlert := range batch {
		dbAlert, exists := dbProvMap[provAlert.ProviderID]
		if exists {
			// don't delete this user
		}
		_, syncErr := ds.syncAlert(ctx, dbAlert, provAlert)
		if syncErr != nil {
			return fmt.Errorf("syncing alert: %w", syncErr)
		}
	}

	return nil
}

func (ds *alertDataSyncer) syncAlert(ctx context.Context, db, prov *ent.Alert) (uuid.UUID, error) {
	var m *ent.AlertMutation
	var id uuid.UUID
	needsSync := true
	if db == nil {
		id = uuid.New()
		m = ds.db.Alert.Create().SetID(id).Mutation()
	} else {
		id = db.ID
		m = ds.db.Alert.UpdateOneID(id).Mutation()

		// TODO: get provider mapping support for fields
		needsSync = db.Title != prov.Title
	}

	m.SetProviderID(prov.ProviderID)
	m.SetTitle(prov.Title)

	if needsSync {
		ds.mutations = append(ds.mutations, m)
	}

	return id, nil
}
