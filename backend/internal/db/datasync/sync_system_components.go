package datasync

import (
	"context"
	"fmt"
	"iter"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/systemcomponent"
)

func syncSystemComponents(ctx context.Context, db *ent.Client, prov rez.SystemComponentsDataProvider) error {
	b := &systemComponentsBatcher{db: db, provider: prov}
	s := newBatchedDataSyncer[*ent.SystemComponent](db, "system_component", b)
	return s.Sync(ctx)
}

type systemComponentsBatcher struct {
	db       *ent.Client
	provider rez.SystemComponentsDataProvider
}

func (b *systemComponentsBatcher) setup(ctx context.Context) error {
	return nil
}

func (b *systemComponentsBatcher) pullData(ctx context.Context) iter.Seq2[*ent.SystemComponent, error] {
	return b.provider.PullSystemComponents(ctx)
}

func (b *systemComponentsBatcher) getDeletionMutations() []ent.Mutation {
	return nil
}

func (b *systemComponentsBatcher) createBatchMutations(ctx context.Context, batch []*ent.SystemComponent) ([]ent.Mutation, error) {
	ids := make([]string, len(batch))
	for i, c := range batch {
		ids[i] = c.ProviderID
	}

	query := b.db.SystemComponent.Query().Where(systemcomponent.ProviderIDIn(ids...))
	dbComponents, queryErr := query.All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("querying db system components: %w", queryErr)
	}

	dbIdMap := make(map[string]*ent.SystemComponent)
	for _, dbCmp := range dbComponents {
		c := dbCmp
		dbIdMap[c.ProviderID] = c
	}

	var mutations []ent.Mutation
	for _, provCmp := range batch {
		dbCmp, exists := dbIdMap[provCmp.ProviderID]
		if exists {
			// don't delete this component
		}
		mutations = append(mutations, b.syncComponent(dbCmp, provCmp)...)
	}

	// TODO: sync component relationships

	return mutations, nil
}

func (b *systemComponentsBatcher) syncComponent(db, prov *ent.SystemComponent) []ent.Mutation {
	var mutations []ent.Mutation

	var m *ent.SystemComponentMutation
	needsSync := true
	if db == nil {
		m = b.db.SystemComponent.Create().Mutation()
	} else {
		m = b.db.SystemComponent.UpdateOneID(db.ID).Mutation()

		// TODO: get provider mapping support for fields
		needsSync = db.Name != prov.Name
	}

	if prov.KindID != uuid.Nil {
		m.SetKindID(prov.KindID)
	}

	m.SetProviderID(prov.ProviderID)
	m.SetName(prov.Name)
	m.SetProperties(prov.Properties)
	m.SetDescription(prov.Description)

	if needsSync {
		mutations = append(mutations, m)
	}

	return mutations
}

/*
func (b *systemComponentsBatcher) syncComponentKind(prov *ent.SystemComponentKind) *ent.SystemComponentKindMutation {
	var m *ent.SystemComponentKindMutation
	var kindId uuid.UUID

	dbKind, dbExists := b.componentKindProvIds[prov.ProviderID]

	if !dbExists {
		kindId = uuid.New()
		b.componentKindProvIds[prov.ProviderID] = &ent.SystemComponentKind{
			ID:          kindId,
			ProviderID:  prov.ProviderID,
			Label:       prov.Label,
			Description: prov.Description,
		}
		m = b.db.SystemComponentKind.Create().SetID(kindId).Mutation()
	} else {
		kindId = dbKind.ID
		m = b.db.SystemComponentKind.UpdateOneID(kindId).Mutation()

		needsSync := dbKind.Label != prov.Label
		if !needsSync {
			return nil
		}
	}

	m.SetLabel(prov.Label)
	m.SetDescription(prov.Description)
	m.SetProviderID(prov.ProviderID)

	return m
}
*/
