package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	knf "github.com/rezible/rezible/ent/knowledgefact"
	knfr "github.com/rezible/rezible/ent/knowledgefactrelationship"
	ts "github.com/rezible/rezible/ent/systemtopologysnapshot"
)

type SystemTopologyService struct {
	db *ent.Client
}

func NewSystemTopologyService(db *ent.Client) (*SystemTopologyService, error) {
	return &SystemTopologyService{db: db}, nil
}

func (s *SystemTopologyService) ListEntities(ctx context.Context, params rez.ListSystemTopologyEntitiesParams) (*ent.ListResult[*ent.KnowledgeFact], error) {
	query := s.db.KnowledgeFact.Query().
		WithAliases().
		WithSourceRelationships(func(q *ent.KnowledgeFactRelationshipQuery) {
			q.Select(knfr.FieldID, knfr.FieldKind, knfr.FieldSourceFactID, knfr.FieldTargetFactID)
		}).
		WithTargetRelationships(func(q *ent.KnowledgeFactRelationshipQuery) {
			q.Select(knfr.FieldID, knfr.FieldKind, knfr.FieldSourceFactID, knfr.FieldKind)
		})
	if params.Search != "" {
		query.Where(knf.DisplayNameContainsFold(params.Search))
	}
	if len(params.Kinds) > 0 {
		query.Where(knf.KindIn(params.Kinds...))
	}
	return ent.DoListQuery[*ent.KnowledgeFact, *ent.KnowledgeFactQuery](ctx, query, params.ListParams)
}

func (s *SystemTopologyService) GetEntity(ctx context.Context, id uuid.UUID) (*ent.KnowledgeFact, error) {
	return s.db.KnowledgeFact.Query().
		Where(knf.ID(id)).
		WithAliases().
		WithSourceRelationships(func(q *ent.KnowledgeFactRelationshipQuery) {
			q.WithTargetFact()
		}).
		WithTargetRelationships(func(q *ent.KnowledgeFactRelationshipQuery) {
			q.WithSourceFact()
		}).
		Only(ctx)
}

func (s *SystemTopologyService) ListRelationships(ctx context.Context, params rez.ListSystemTopologyRelationshipsParams) (*ent.ListResult[*ent.KnowledgeFactRelationship], error) {
	query := s.db.KnowledgeFactRelationship.Query().
		WithSourceFact().
		WithTargetFact()
	if len(params.Kinds) > 0 {
		query.Where(knfr.KindIn(params.Kinds...))
	}
	if params.SourceFactID != uuid.Nil {
		query.Where(knfr.SourceFactID(params.SourceFactID))
	}
	if params.TargetFactID != uuid.Nil {
		query.Where(knfr.TargetFactID(params.TargetFactID))
	}
	if params.FactID != uuid.Nil {
		query.Where(knfr.Or(knfr.SourceFactID(params.FactID), knfr.TargetFactID(params.FactID)))
	}
	return ent.DoListQuery[*ent.KnowledgeFactRelationship, *ent.KnowledgeFactRelationshipQuery](ctx, query, params.ListParams)
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
	relationshipsByID := make(map[uuid.UUID]*ent.KnowledgeFactRelationship)

	for range depth {
		rels, queryErr := s.relationshipsTouching(ctx, frontier, params.RelationshipKinds)
		if queryErr != nil {
			return nil, fmt.Errorf("query relationship neighborhood: %w", queryErr)
		}
		next := make([]uuid.UUID, 0)
		for _, rel := range rels {
			relationshipsByID[rel.ID] = rel
			for _, entityID := range []uuid.UUID{rel.SourceFactID, rel.TargetFactID} {
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
	entities, entityErr := s.db.KnowledgeFact.Query().
		Where(knf.IDIn(ids...)).
		WithAliases().
		All(ctx)
	if entityErr != nil {
		return nil, fmt.Errorf("query neighborhood entities: %w", entityErr)
	}

	relationships := make([]*ent.KnowledgeFactRelationship, 0, len(relationshipsByID))
	for _, rel := range relationshipsByID {
		relationships = append(relationships, rel)
	}
	return &rez.SystemTopologyGraph{Entities: entities, Relationships: relationships}, nil
}

func (s *SystemTopologyService) CreateSnapshot(ctx context.Context, params rez.CreateSystemTopologySnapshotParams) (*ent.SystemTopologySnapshot, error) {
	return nil, nil
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

func (s *SystemTopologyService) relationshipsTouching(ctx context.Context, entityIDs []uuid.UUID, kinds []string) ([]*ent.KnowledgeFactRelationship, error) {
	if len(entityIDs) == 0 {
		return nil, nil
	}
	query := s.db.KnowledgeFactRelationship.Query().
		Where(knfr.Or(knfr.SourceFactIDIn(entityIDs...), knfr.TargetFactIDIn(entityIDs...))).
		WithSourceFact().
		WithTargetFact()
	if len(kinds) > 0 {
		query.Where(knfr.KindIn(kinds...))
	}
	return query.All(ctx)
}
