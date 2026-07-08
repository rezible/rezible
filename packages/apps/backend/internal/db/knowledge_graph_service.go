package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	kgs "github.com/rezible/rezible/ent/knowledgegraphsnapshot"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
)

type KnowledgeGraphService struct {
	db rez.Database
}

func NewKnowledgeGraphService(db rez.Database) (*KnowledgeGraphService, error) {
	return &KnowledgeGraphService{db: db}, nil
}

func (s *KnowledgeGraphService) ListEntities(ctx context.Context, params rez.ListKnowledgeGraphEntitiesParams) (*ent.ListResult[ent.KnowledgeEntity], error) {
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
	if len(params.Predicates) > 0 {
		query.Where(params.Predicates...)
	}
	return ent.DoListQuery[ent.KnowledgeEntity, *ent.KnowledgeEntityQuery](ctx, query, params.ListParams)
}

func (s *KnowledgeGraphService) GetEntity(ctx context.Context, id uuid.UUID) (*ent.KnowledgeEntity, error) {
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

func (s *KnowledgeGraphService) ListRelationships(ctx context.Context, params rez.ListKnowledgeGraphRelationshipsParams) (*ent.ListResult[ent.KnowledgeRelationship], error) {
	query := s.db.Client(ctx).KnowledgeRelationship.Query().
		WithSourceEntity().
		WithTargetEntity()
	if len(params.Predicates) > 0 {
		query.Where(params.Predicates...)
	}
	return ent.DoListQuery[ent.KnowledgeRelationship, *ent.KnowledgeRelationshipQuery](ctx, query, params.ListParams)
}

func (s *KnowledgeGraphService) GetView(ctx context.Context, u uuid.UUID, params rez.GetKnowledgeGraphViewParams) (*rez.KnowledgeGraphView, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *KnowledgeGraphService) CreateSnapshot(ctx context.Context, params rez.CreateKnowledgeGraphSnapshotParams) (*ent.KnowledgeGraphSnapshot, error) {
	return nil, nil
}

func (s *KnowledgeGraphService) GetSnapshot(ctx context.Context, id uuid.UUID) (*ent.KnowledgeGraphSnapshot, error) {
	return s.db.Client(ctx).KnowledgeGraphSnapshot.Query().
		Where(kgs.ID(id)).
		WithEntities().
		WithRelationships(func(q *ent.KnowledgeGraphSnapshotRelationshipQuery) {
			q.WithSourceSnapshotEntity()
			q.WithTargetSnapshotEntity()
		}).
		Only(ctx)
}
