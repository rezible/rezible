package providers

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/systemcomponent"
)

const (
	systemComponentsDataType = "system_components"
)

type systemComponentsDataSyncer struct {
	db       *ent.Client
	provider rez.SystemComponentsDataProvider

	mutations []ent.Mutation

	componentKindProvIds map[string]*ent.SystemComponentKind
}

func newSystemComponentsDataSyncer(db *ent.Client, prov rez.SystemComponentsDataProvider) *systemComponentsDataSyncer {
	ds := &systemComponentsDataSyncer{db: db, provider: prov}
	ds.resetState()
	return ds
}

func (ds *systemComponentsDataSyncer) resetState() {
	ds.mutations = make([]ent.Mutation, 0)
}

func (ds *systemComponentsDataSyncer) saveSyncHistory(ctx context.Context, start time.Time, num int, dataType string) {
	historyErr := ds.db.ProviderSyncHistory.Create().
		SetStartedAt(start).
		SetFinishedAt(time.Now()).
		SetNumMutations(num).
		SetDataType(dataType).
		Exec(ctx)
	if historyErr != nil {
		log.Error().Err(historyErr).Msg("failed to save sync history")
	}
}

func (ds *systemComponentsDataSyncer) SyncProviderData(ctx context.Context) error {
	start := time.Now()

	lastSync := getLastSyncTime(ctx, ds.db, systemComponentsDataType)
	if lastSync.Add(time.Minute * 30).After(start) {
		return nil
	}

	if kindsErr := ds.loadComponentKindsMap(ctx); kindsErr != nil {
		return fmt.Errorf("loading component kinds map: %w", kindsErr)
	}

	if componentsErr := ds.syncAllProviderComponents(ctx); componentsErr != nil {
		return fmt.Errorf("system components: %w", componentsErr)
	}
	log.Info().
		Msg("system components data sync complete")

	return nil
}

func (ds *systemComponentsDataSyncer) loadComponentKindsMap(ctx context.Context) error {
	provIds := make(map[string]*ent.SystemComponentKind)

	kinds, queryErr := ds.db.SystemComponentKind.Query().All(ctx)
	if queryErr != nil {
		return fmt.Errorf("querying component kinds: %w", queryErr)
	}
	for _, kind := range kinds {
		if kind.ProviderID == "" {
			continue
		}
		k := kind
		provIds[k.ProviderID] = k
	}
	ds.componentKindProvIds = provIds
	return nil
}

func (ds *systemComponentsDataSyncer) syncAllProviderComponents(ctx context.Context) error {
	var batch []*ent.SystemComponent

	start := time.Now()
	var numMutations int

	batchSize := 10
	for cmp, pullErr := range ds.provider.PullSystemComponents(ctx) {
		if pullErr != nil {
			return fmt.Errorf("pull system components: %w", pullErr)
		}

		batch = append(batch, cmp)

		if len(batch) >= batchSize {
			batchMuts, syncErr := ds.syncBatch(ctx, batch)
			if syncErr != nil {
				return syncErr
			}
			numMutations += batchMuts
			batch = make([]*ent.SystemComponent, 0)
		}
	}

	lastBatchMuts, batchErr := ds.syncBatch(ctx, batch)
	if batchErr != nil {
		return batchErr
	}
	numMutations += lastBatchMuts

	ds.saveSyncHistory(ctx, start, numMutations, systemComponentsDataType)

	return nil
}

func (ds *systemComponentsDataSyncer) syncBatch(ctx context.Context, batch []*ent.SystemComponent) (int, error) {
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

func (ds *systemComponentsDataSyncer) createBatchSyncMutations(ctx context.Context, batch []*ent.SystemComponent) error {
	ids := make([]string, len(batch))
	for i, c := range batch {
		ids[i] = c.ProviderID
	}

	query := ds.db.SystemComponent.Query().Where(systemcomponent.ProviderIDIn(ids...))
	dbComponents, queryErr := query.All(ctx)
	if queryErr != nil {
		return fmt.Errorf("querying system components: %w", queryErr)
	}

	dbIdMap := make(map[string]*ent.SystemComponent)
	for _, dbCmp := range dbComponents {
		c := dbCmp
		dbIdMap[c.ProviderID] = c
	}

	for _, provCmp := range batch {
		if provCmp.Edges.Kind != nil {
			provCmp.KindID = ds.syncComponentKind(provCmp.Edges.Kind)
		}

		dbCmp, exists := dbIdMap[provCmp.ProviderID]
		if exists {
			// don't delete this component
		}
		_ = ds.syncComponent(dbCmp, provCmp)
	}

	// TODO: sync component relationships

	return nil
}

func (ds *systemComponentsDataSyncer) syncComponentKind(prov *ent.SystemComponentKind) uuid.UUID {
	var m *ent.SystemComponentKindMutation
	var kindId uuid.UUID

	dbKind, dbExists := ds.componentKindProvIds[prov.ProviderID]

	needsSync := true
	if !dbExists {
		kindId = uuid.New()
		ds.componentKindProvIds[prov.ProviderID] = &ent.SystemComponentKind{
			ID:          kindId,
			ProviderID:  prov.ProviderID,
			Label:       prov.Label,
			Description: prov.Description,
		}
		m = ds.db.SystemComponentKind.Create().SetID(kindId).Mutation()
	} else {
		kindId = dbKind.ID
		m = ds.db.SystemComponentKind.UpdateOneID(kindId).Mutation()

		needsSync = dbKind.Label != prov.Label
	}

	m.SetLabel(prov.Label)
	m.SetDescription(prov.Description)
	m.SetProviderID(prov.ProviderID)

	if needsSync {
		ds.mutations = append(ds.mutations, m)
	}

	return kindId
}

func (ds *systemComponentsDataSyncer) syncComponent(db, prov *ent.SystemComponent) uuid.UUID {
	var m *ent.SystemComponentMutation
	var componentId uuid.UUID
	needsSync := true
	if db == nil {
		componentId = uuid.New()
		m = ds.db.SystemComponent.Create().SetID(componentId).Mutation()
	} else {
		componentId = db.ID
		m = ds.db.SystemComponent.UpdateOneID(componentId).Mutation()

		// TODO: get provider mapping support for fields
		needsSync = db.Name != prov.Name
	}

	m.SetProviderID(prov.ProviderID)
	m.SetName(prov.Name)
	if prov.KindID != uuid.Nil {
		m.SetKindID(prov.KindID)
	}
	m.SetProperties(prov.Properties)
	m.SetDescription(prov.Description)

	if needsSync {
		ds.mutations = append(ds.mutations, m)
	}

	return componentId
}
