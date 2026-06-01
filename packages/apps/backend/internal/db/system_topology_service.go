package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ts "github.com/rezible/rezible/ent/systemtopologysnapshot"
)

type SystemTopologyService struct {
	db rez.Database
}

func NewSystemTopologyService(db rez.Database) (*SystemTopologyService, error) {
	return &SystemTopologyService{db: db}, nil
}

func (s *SystemTopologyService) ListEntities(ctx context.Context, params rez.ListSystemTopologyEntitiesParams) (*ent.ListResult[ent.KnowledgeEntity], error) {
	query := s.db.Client(ctx).KnowledgeEntity.Query().
		WithAliases().
		WithSourceRelationships(func(q *ent.KnowledgeRelationshipQuery) {
			q.Select(knr.FieldID, knr.FieldKind, knr.FieldSourceEntityID, knr.FieldTargetEntityID)
		}).
		WithTargetRelationships(func(q *ent.KnowledgeRelationshipQuery) {
			q.Select(knr.FieldID, knr.FieldKind, knr.FieldSourceEntityID, knr.FieldTargetEntityID)
		})
	if params.Search != "" {
		query.Where(kne.DisplayNameContainsFold(params.Search))
	}
	if len(params.Kinds) > 0 {
		query.Where(kne.KindIn(params.Kinds...))
	}
	return ent.DoListQuery[ent.KnowledgeEntity, *ent.KnowledgeEntityQuery](ctx, query, params.ListParams)
}

func (s *SystemTopologyService) GetEntity(ctx context.Context, id uuid.UUID) (*ent.KnowledgeEntity, error) {
	return s.db.Client(ctx).KnowledgeEntity.Query().
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

func (s *SystemTopologyService) ListRelationships(ctx context.Context, params rez.ListSystemTopologyRelationshipsParams) (*ent.ListResult[ent.KnowledgeRelationship], error) {
	query := s.db.Client(ctx).KnowledgeRelationship.Query().
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
	return ent.DoListQuery[ent.KnowledgeRelationship, *ent.KnowledgeRelationshipQuery](ctx, query, params.ListParams)
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
	entities, entityErr := s.db.Client(ctx).KnowledgeEntity.Query().
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
	return nil, nil
}

func (s *SystemTopologyService) GetSnapshot(ctx context.Context, id uuid.UUID) (*ent.SystemTopologySnapshot, error) {
	return s.db.Client(ctx).SystemTopologySnapshot.Query().
		Where(ts.ID(id)).
		WithEntities().
		WithRelationships(func(q *ent.SystemTopologySnapshotRelationshipQuery) {
			q.WithSourceSnapshotEntity()
			q.WithTargetSnapshotEntity()
		}).
		Only(ctx)
}

func (s *SystemTopologyService) relationshipsTouching(ctx context.Context, entityIDs []uuid.UUID, kinds []string) ([]*ent.KnowledgeRelationship, error) {
	if len(entityIDs) == 0 {
		return nil, nil
	}
	query := s.db.Client(ctx).KnowledgeRelationship.Query().
		Where(knr.Or(knr.SourceEntityIDIn(entityIDs...), knr.TargetEntityIDIn(entityIDs...))).
		WithSourceEntity().
		WithTargetEntity()
	if len(kinds) > 0 {
		query.Where(knr.KindIn(kinds...))
	}
	return query.All(ctx)
}
