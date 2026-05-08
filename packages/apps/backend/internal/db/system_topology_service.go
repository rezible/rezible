package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knea "github.com/rezible/rezible/ent/knowledgeentityalias"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/ent/predicate"
	ts "github.com/rezible/rezible/ent/systemtopologysnapshot"
)

type SystemTopologyService struct {
	db *ent.Client
}

func NewSystemTopologyService(db *ent.Client) (*SystemTopologyService, error) {
	return &SystemTopologyService{db: db}, nil
}

func (s *SystemTopologyService) ListEntities(ctx context.Context, params rez.ListSystemTopologyEntitiesParams) (*ent.ListResult[*ent.KnowledgeEntity], error) {
	query := s.db.KnowledgeEntity.Query().
		WithAliases().
		WithSourceRelationships(func(q *ent.KnowledgeRelationshipQuery) {
			q.Select(knr.FieldID, knr.FieldSourceEntityID, knr.FieldTargetEntityID, knr.FieldKind)
		}).
		WithTargetRelationships(func(q *ent.KnowledgeRelationshipQuery) {
			q.Select(knr.FieldID, knr.FieldSourceEntityID, knr.FieldTargetEntityID, knr.FieldKind)
		})
	if params.Search != "" {
		query.Where(kne.DisplayNameContainsFold(params.Search))
	}
	if len(params.Kinds) > 0 {
		query.Where(kne.KindIn(params.Kinds...))
	}
	if params.Provider != "" || params.ProviderSource != "" || params.SubjectKind != "" {
		query.Where(kne.HasAliasesWith(aliasPredicates(params.Provider, params.ProviderSource, params.SubjectKind)...))
	}
	return ent.DoListQuery[*ent.KnowledgeEntity, *ent.KnowledgeEntityQuery](ctx, query, params.ListParams)
}

func (s *SystemTopologyService) GetEntity(ctx context.Context, id uuid.UUID) (*ent.KnowledgeEntity, error) {
	return s.db.KnowledgeEntity.Query().
		Where(kne.ID(id)).
		WithAliases().
		WithSourceRelationships(func(q *ent.KnowledgeRelationshipQuery) {
			q.WithTargetEntity()
		}).
		WithTargetRelationships(func(q *ent.KnowledgeRelationshipQuery) {
			q.WithSourceEntity()
		}).
		Only(ctx)
}

func (s *SystemTopologyService) ListRelationships(ctx context.Context, params rez.ListSystemTopologyRelationshipsParams) (*ent.ListResult[*ent.KnowledgeRelationship], error) {
	query := s.db.KnowledgeRelationship.Query().
		WithSourceEntity().
		WithTargetEntity()
	if len(params.Kinds) > 0 {
		query.Where(knr.KindIn(params.Kinds...))
	}
	if params.SourceEntityID != uuid.Nil {
		query.Where(knr.SourceEntityID(params.SourceEntityID))
	}
	if params.TargetEntityID != uuid.Nil {
		query.Where(knr.TargetEntityID(params.TargetEntityID))
	}
	if params.EntityID != uuid.Nil {
		query.Where(knr.Or(knr.SourceEntityID(params.EntityID), knr.TargetEntityID(params.EntityID)))
	}
	return ent.DoListQuery[*ent.KnowledgeRelationship, *ent.KnowledgeRelationshipQuery](ctx, query, params.ListParams)
}

func (s *SystemTopologyService) GetNeighborhood(ctx context.Context, id uuid.UUID, params rez.SystemTopologyNeighborhoodParams) (*rez.SystemTopologyGraph, error) {
	depth := params.Depth
	if depth < 1 {
		depth = 1
	}
	if depth > 4 {
		depth = 4
	}

	entityIDs := map[uuid.UUID]struct{}{id: {}}
	frontier := []uuid.UUID{id}
	relationshipsByID := make(map[uuid.UUID]*ent.KnowledgeRelationship)

	for range depth {
		rels, queryErr := s.relationshipsTouching(ctx, frontier, params.RelationshipKinds)
		if queryErr != nil {
			return nil, fmt.Errorf("query relationship neighborhood: %w", queryErr)
		}
		next := make([]uuid.UUID, 0)
		for _, rel := range rels {
			relationshipsByID[rel.ID] = rel
			for _, entityID := range []uuid.UUID{rel.SourceEntityID, rel.TargetEntityID} {
				if _, seen := entityIDs[entityID]; seen {
					continue
				}
				entityIDs[entityID] = struct{}{}
				next = append(next, entityID)
			}
		}
		frontier = next
		if len(frontier) == 0 {
			break
		}
	}

	ids := make([]uuid.UUID, 0, len(entityIDs))
	for entityID := range entityIDs {
		ids = append(ids, entityID)
	}
	entities, entityErr := s.db.KnowledgeEntity.Query().
		Where(kne.IDIn(ids...)).
		WithAliases().
		All(ctx)
	if entityErr != nil {
		return nil, fmt.Errorf("query neighborhood entities: %w", entityErr)
	}

	relationships := make([]*ent.KnowledgeRelationship, 0, len(relationshipsByID))
	for _, rel := range relationshipsByID {
		relationships = append(relationships, rel)
	}
	return &rez.SystemTopologyGraph{Entities: entities, Relationships: relationships}, nil
}

func (s *SystemTopologyService) CreateSnapshot(ctx context.Context, params rez.CreateSystemTopologySnapshotParams) (*ent.SystemTopologySnapshot, error) {
	graph, graphErr := s.snapshotGraph(ctx, params)
	if graphErr != nil {
		return nil, graphErr
	}
	asOf := params.AsOf
	if asOf.IsZero() {
		asOf = time.Now()
	}
	scope := params.Scope
	if scope == "" {
		scope = ts.ScopeExplicitEntities.String()
		if len(params.RootEntityIDs) > 0 {
			scope = ts.ScopeRootEntities.String()
		}
	}
	scopeProperties := params.ScopeProperties
	if scopeProperties == nil {
		scopeProperties = map[string]any{
			"entityIds":   uuidStrings(params.EntityIDs),
			"rootIds":     uuidStrings(params.RootEntityIDs),
			"depth":       params.Depth,
			"entityKinds": params.EntityKinds,
			"relKinds":    params.RelationshipKinds,
		}
	}

	var snapshot *ent.SystemTopologySnapshot
	if txErr := ent.WithTx(ctx, s.db, func(tx *ent.Tx) error {
		create := tx.SystemTopologySnapshot.Create().
			SetAsOf(asOf).
			SetScope(ts.Scope(scope)).
			SetScopeProperties(scopeProperties)
		if params.Name != "" {
			create.SetName(params.Name)
		}
		var createErr error
		snapshot, createErr = create.Save(ctx)
		if createErr != nil {
			return fmt.Errorf("create snapshot: %w", createErr)
		}

		snapshotEntities := make(map[uuid.UUID]*ent.SystemTopologySnapshotEntity, len(graph.Entities))
		for _, entity := range graph.Entities {
			aliases := snapshotAliases(entity.Edges.Aliases)
			snapEntity, entityErr := tx.SystemTopologySnapshotEntity.Create().
				SetSnapshotID(snapshot.ID).
				SetKnowledgeEntityID(entity.ID).
				SetEntityKind(entity.Kind).
				SetDisplayName(entity.DisplayName).
				SetDescription(entity.Description).
				SetProperties(entity.Properties).
				SetAliases(aliases).
				Save(ctx)
			if entityErr != nil {
				return fmt.Errorf("create snapshot entity: %w", entityErr)
			}
			snapshotEntities[entity.ID] = snapEntity
		}

		for _, rel := range graph.Relationships {
			source, sourceOK := snapshotEntities[rel.SourceEntityID]
			target, targetOK := snapshotEntities[rel.TargetEntityID]
			if !sourceOK || !targetOK {
				continue
			}
			_, relErr := tx.SystemTopologySnapshotRelationship.Create().
				SetSnapshotID(snapshot.ID).
				SetKnowledgeRelationshipID(rel.ID).
				SetSourceSnapshotEntityID(source.ID).
				SetTargetSnapshotEntityID(target.ID).
				SetRelationshipKind(rel.Kind).
				SetDisplayName(rel.DisplayName).
				SetDescription(rel.Description).
				SetProperties(rel.Properties).
				Save(ctx)
			if relErr != nil {
				return fmt.Errorf("create snapshot relationship: %w", relErr)
			}
		}
		return nil
	}); txErr != nil {
		return nil, txErr
	}

	return s.GetSnapshot(ctx, snapshot.ID)
}

func (s *SystemTopologyService) GetSnapshot(ctx context.Context, id uuid.UUID) (*ent.SystemTopologySnapshot, error) {
	return s.db.SystemTopologySnapshot.Query().
		Where(ts.ID(id)).
		WithEntities().
		WithRelationships(func(q *ent.SystemTopologySnapshotRelationshipQuery) {
			q.WithSourceSnapshotEntity()
			q.WithTargetSnapshotEntity()
		}).
		Only(ctx)
}

func (s *SystemTopologyService) snapshotGraph(ctx context.Context, params rez.CreateSystemTopologySnapshotParams) (*rez.SystemTopologyGraph, error) {
	if len(params.RootEntityIDs) > 0 {
		graph := &rez.SystemTopologyGraph{}
		seen := make(map[uuid.UUID]*ent.KnowledgeEntity)
		seenRelationships := make(map[uuid.UUID]*ent.KnowledgeRelationship)
		for _, rootID := range params.RootEntityIDs {
			partial, neighborhoodErr := s.GetNeighborhood(ctx, rootID, rez.SystemTopologyNeighborhoodParams{
				Depth:             params.Depth,
				RelationshipKinds: params.RelationshipKinds,
			})
			if neighborhoodErr != nil {
				return nil, neighborhoodErr
			}
			for _, entity := range partial.Entities {
				seen[entity.ID] = entity
			}
			for _, rel := range partial.Relationships {
				seenRelationships[rel.ID] = rel
			}
		}
		for _, entity := range seen {
			graph.Entities = append(graph.Entities, entity)
		}
		for _, rel := range seenRelationships {
			graph.Relationships = append(graph.Relationships, rel)
		}
		return graph, nil
	}

	if len(params.EntityIDs) == 0 {
		return nil, fmt.Errorf("snapshot requires entity IDs or root entity IDs")
	}
	entities, entityErr := s.db.KnowledgeEntity.Query().
		Where(kne.IDIn(params.EntityIDs...)).
		WithAliases().
		All(ctx)
	if entityErr != nil {
		return nil, fmt.Errorf("query snapshot entities: %w", entityErr)
	}
	rels, relErr := s.relationshipsTouching(ctx, params.EntityIDs, params.RelationshipKinds)
	if relErr != nil {
		return nil, fmt.Errorf("query snapshot relationships: %w", relErr)
	}
	entitySet := make(map[uuid.UUID]struct{}, len(entities))
	for _, entity := range entities {
		entitySet[entity.ID] = struct{}{}
	}
	filteredRels := make([]*ent.KnowledgeRelationship, 0, len(rels))
	for _, rel := range rels {
		_, sourceOK := entitySet[rel.SourceEntityID]
		_, targetOK := entitySet[rel.TargetEntityID]
		if sourceOK && targetOK {
			filteredRels = append(filteredRels, rel)
		}
	}
	return &rez.SystemTopologyGraph{Entities: entities, Relationships: filteredRels}, nil
}

func (s *SystemTopologyService) relationshipsTouching(ctx context.Context, entityIDs []uuid.UUID, kinds []string) ([]*ent.KnowledgeRelationship, error) {
	if len(entityIDs) == 0 {
		return nil, nil
	}
	query := s.db.KnowledgeRelationship.Query().
		Where(knr.Or(knr.SourceEntityIDIn(entityIDs...), knr.TargetEntityIDIn(entityIDs...))).
		WithSourceEntity().
		WithTargetEntity()
	if len(kinds) > 0 {
		query.Where(knr.KindIn(kinds...))
	}
	return query.All(ctx)
}

func aliasPredicates(provider, providerSource, subjectKind string) []predicate.KnowledgeEntityAlias {
	predicates := make([]predicate.KnowledgeEntityAlias, 0, 3)
	if provider != "" {
		predicates = append(predicates, knea.Provider(provider))
	}
	if providerSource != "" {
		predicates = append(predicates, knea.ProviderSource(providerSource))
	}
	if subjectKind != "" {
		predicates = append(predicates, knea.SubjectKind(subjectKind))
	}
	return predicates
}

func snapshotAliases(aliases []*ent.KnowledgeEntityAlias) []map[string]any {
	res := make([]map[string]any, len(aliases))
	for i, alias := range aliases {
		res[i] = map[string]any{
			"id":             alias.ID.String(),
			"provider":       alias.Provider,
			"providerSource": alias.ProviderSource,
			"subjectKind":    alias.SubjectKind,
			"subjectRef":     alias.SubjectRef,
			"firstSeenAt":    alias.FirstSeenAt,
			"lastSeenAt":     alias.LastSeenAt,
		}
	}
	return res
}

func uuidStrings(ids []uuid.UUID) []string {
	res := make([]string, len(ids))
	for i, id := range ids {
		res[i] = id.String()
	}
	return res
}
