package datasync

import (
	"context"
	"fmt"
	"iter"
	"log/slog"
	"reflect"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/systemcomponent"
	scr "github.com/rezible/rezible/ent/systemcomponentrelationship"
)

const systemComponentProviderSource = "datasync/system_components"

func syncSystemComponents(ctx context.Context, db *ent.Client, prov rez.SystemComponentsDataProvider, providerName string, opts SyncOptions, met *metrics) error {
	b := &systemComponentsBatcher{db: db, provider: prov, providerName: providerName}
	s := newBatchedDataSyncer[ent.SystemComponent](db, "system_component", b, opts, met)
	return s.Sync(ctx)
}

type systemComponentsBatcher struct {
	db           *ent.Client
	provider     rez.SystemComponentsDataProvider
	providerName string
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
		ids[i] = c.ExternalID
	}

	query := b.db.SystemComponent.Query().Where(systemcomponent.ExternalIDIn(ids...))
	dbComponents, queryErr := query.All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("querying db system components: %w", queryErr)
	}

	dbIdMap := make(map[string]*ent.SystemComponent)
	for _, dbCmp := range dbComponents {
		c := dbCmp
		dbIdMap[c.ExternalID] = c
	}

	var mutations []ent.Mutation
	for _, provCmp := range batch {
		dbCmp, exists := dbIdMap[provCmp.ExternalID]
		if exists {
			provCmp.ID = dbCmp.ID
		} else {
			provCmp.ID = uuid.New()
			dbIdMap[provCmp.ExternalID] = &ent.SystemComponent{
				ID:         provCmp.ID,
				ExternalID: provCmp.ExternalID,
			}
		}
		mutations = append(mutations, b.syncComponent(dbCmp, provCmp)...)
	}

	relMutations, relErr := b.syncRelationships(ctx, batch, dbIdMap)
	if relErr != nil {
		return nil, fmt.Errorf("sync relationships: %w", relErr)
	}
	mutations = append(mutations, relMutations...)

	return mutations, nil
}

func (b *systemComponentsBatcher) syncComponent(db, prov *ent.SystemComponent) []ent.Mutation {
	var mutations []ent.Mutation

	var m *ent.SystemComponentMutation
	needsSync := true
	if db == nil {
		m = b.db.SystemComponent.Create().SetID(prov.ID).Mutation()
	} else {
		m = b.db.SystemComponent.UpdateOneID(db.ID).Mutation()

		// TODO: get provider mapping support for fields
		needsSync = db.Name != prov.Name ||
			db.Description != prov.Description ||
			db.KindID != prov.KindID ||
			!reflect.DeepEqual(db.Properties, prov.Properties)
	}

	if prov.KindID != uuid.Nil {
		m.SetKindID(prov.KindID)
	} else if prov.Edges.Kind != nil && prov.Edges.Kind.ID != uuid.Nil {
		m.SetKindID(prov.Edges.Kind.ID)
	}

	m.SetExternalID(prov.ExternalID)
	m.SetName(prov.Name)
	m.SetProperties(prov.Properties)
	m.SetDescription(prov.Description)

	if needsSync {
		mutations = append(mutations, m)
	}

	return mutations
}

type componentRelationshipRef struct {
	sourceID    uuid.UUID
	targetID    uuid.UUID
	targetRef   string
	description string
}

func (b *systemComponentsBatcher) syncRelationships(ctx context.Context, batch []*ent.SystemComponent, dbIdMap map[string]*ent.SystemComponent) ([]ent.Mutation, error) {
	refs := make([]componentRelationshipRef, 0)
	for _, provCmp := range batch {
		source, ok := dbIdMap[provCmp.ExternalID]
		if !ok || source.ID == uuid.Nil {
			slog.Warn("skipping system component relationship with unresolved source", "sourceExternalId", provCmp.ExternalID)
			continue
		}
		for _, rel := range provCmp.Edges.ComponentRelationships {
			if rel == nil {
				continue
			}
			targetRef := rel.ExternalID
			target, ok := dbIdMap[targetRef]
			if !ok && targetRef != "" {
				dbTarget, targetErr := b.db.SystemComponent.Query().
					Where(systemcomponent.ExternalID(targetRef)).
					Only(ctx)
				if targetErr != nil && !ent.IsNotFound(targetErr) {
					return nil, fmt.Errorf("querying relationship target component: %w", targetErr)
				}
				if dbTarget != nil {
					target = dbTarget
					dbIdMap[targetRef] = dbTarget
					ok = true
				}
			}
			if !ok || target == nil || target.ID == uuid.Nil {
				slog.Warn("skipping system component relationship with unresolved target",
					"sourceExternalId", provCmp.ExternalID,
					"targetExternalId", targetRef,
				)
				continue
			}
			refs = append(refs, componentRelationshipRef{
				sourceID:    source.ID,
				targetID:    target.ID,
				targetRef:   targetRef,
				description: rel.Description,
			})
		}
	}
	if len(refs) == 0 {
		return nil, nil
	}

	sourceIDs := make([]uuid.UUID, 0, len(refs))
	targetIDs := make([]uuid.UUID, 0, len(refs))
	for _, ref := range refs {
		sourceIDs = append(sourceIDs, ref.sourceID)
		targetIDs = append(targetIDs, ref.targetID)
	}
	existingRels, queryErr := b.db.SystemComponentRelationship.Query().
		Where(scr.SourceIDIn(sourceIDs...), scr.TargetIDIn(targetIDs...)).
		All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("querying db system component relationships: %w", queryErr)
	}

	existingByPair := make(map[string]*ent.SystemComponentRelationship, len(existingRels))
	for _, rel := range existingRels {
		existingByPair[relationshipPairKey(rel.SourceID, rel.TargetID)] = rel
	}

	mutations := make([]ent.Mutation, 0, len(refs))
	for _, ref := range refs {
		if existing := existingByPair[relationshipPairKey(ref.sourceID, ref.targetID)]; existing != nil {
			m := b.db.SystemComponentRelationship.UpdateOneID(existing.ID).Mutation()
			m.SetExternalID(ref.targetRef)
			m.SetDescription(ref.description)
			mutations = append(mutations, m)
			continue
		}
		m := b.db.SystemComponentRelationship.Create().
			SetID(uuid.New()).
			SetSourceID(ref.sourceID).
			SetTargetID(ref.targetID).
			SetExternalID(ref.targetRef).
			SetDescription(ref.description).
			Mutation()
		mutations = append(mutations, m)
	}

	return mutations, nil
}

func relationshipPairKey(sourceID uuid.UUID, targetID uuid.UUID) string {
	return sourceID.String() + ":" + targetID.String()
}

// TODO: this will all get replaced by the new provider observation backfill flow anyway
/*
func (b *systemComponentsBatcher) afterBatchApplied(ctx context.Context, batch []*ent.SystemComponent) error {
		if b.messages == nil {
			return nil
		}
		providerName := b.providerName
		if providerName == "" {
			providerName = "unknown"
		}
		for _, cmp := range batch {
			if publishErr := b.messages.PublishEvent(ctx, b.observedEvent(providerName, cmp)); publishErr != nil {
				slog.Error("failed to publish system component observed event", "error", publishErr, "externalId", cmp.ExternalID)
			}
		}
	return nil
}

func (b *systemComponentsBatcher) observedEvent(providerName string, cmp *ent.SystemComponent) *rez.SystemComponentObserved {
	event := &rez.SystemComponentObserved{
		Provider:       providerName,
		ProviderSource: systemComponentProviderSource,
		ExternalID:     cmp.ExternalID,
		Name:           cmp.Name,
		Description:    cmp.Description,
		Properties:     cmp.Properties,
		Relationships:  make([]rez.SystemComponentRelationshipObserved, 0, len(cmp.Edges.ComponentRelationships)),
	}
	for _, rel := range cmp.Edges.ComponentRelationships {
		if rel == nil {
			continue
		}
		event.Relationships = append(event.Relationships, rez.SystemComponentRelationshipObserved{
			TargetExternalID: rel.ExternalID,
			Description:      rel.Description,
		})
	}
	return event
}
*/
