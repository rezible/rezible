package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
)

const (
	maxKnowledgeViewDepth         = 4
	maxKnowledgeViewEntities      = 200
	maxKnowledgeViewRelationships = 400
	maxKnowledgeViewEvidence      = 1000
)

type KnowledgeGraphService struct {
	db rez.Database
}

func NewKnowledgeGraphService(db rez.Database) (*KnowledgeGraphService, error) {
	if db == nil {
		return nil, fmt.Errorf("database is required")
	}
	return &KnowledgeGraphService{db: db}, nil
}

func (s *KnowledgeGraphService) ListEntities(ctx context.Context, params rez.ListKnowledgeGraphEntitiesParams) (*ent.ListResult[ent.KnowledgeEntity], error) {
	query := s.db.Client(ctx).KnowledgeEntity.Query().
		WithAliases().
		WithSourceRelationships(func(query *ent.KnowledgeRelationshipQuery) {
			query.Select(knr.FieldID, knr.FieldKind, knr.FieldSourceEntityID, knr.FieldTargetEntityID)
		}).
		WithTargetRelationships(func(query *ent.KnowledgeRelationshipQuery) {
			query.Select(knr.FieldID, knr.FieldKind, knr.FieldSourceEntityID, knr.FieldTargetEntityID)
		})
	//if params.Search != "" {
	//	query.Where(kne.DisplayNameContainsFold(params.Search))
	//}
	if len(params.Predicates) > 0 {
		query.Where(params.Predicates...)
	}
	return ent.DoListQuery[ent.KnowledgeEntity, *ent.KnowledgeEntityQuery](ctx, query, params.ListParams)
}

func (s *KnowledgeGraphService) GetEntity(ctx context.Context, id uuid.UUID) (*ent.KnowledgeEntity, error) {
	query := s.db.Client(ctx).KnowledgeEntity.Query().
		Where(kne.ID(id)).
		WithAliases().
		WithSourceRelationships(func(query *ent.KnowledgeRelationshipQuery) {
			query.WithTargetEntity()
		}).
		WithTargetRelationships(func(query *ent.KnowledgeRelationshipQuery) {
			query.WithSourceEntity()
		})
	return query.Only(ctx)
}

func (s *KnowledgeGraphService) ListRelationships(ctx context.Context, params rez.ListKnowledgeGraphRelationshipsParams) (*ent.ListResult[ent.KnowledgeRelationship], error) {
	query := s.db.Client(ctx).KnowledgeRelationship.Query().
		WithSourceEntity().
		WithTargetEntity().
		WithAliases()
	if len(params.Predicates) > 0 {
		query.Where(params.Predicates...)
	}
	return ent.DoListQuery[ent.KnowledgeRelationship, *ent.KnowledgeRelationshipQuery](ctx, query, params.ListParams)
}

func (s *KnowledgeGraphService) GetRelationship(ctx context.Context, id uuid.UUID) (*ent.KnowledgeRelationship, error) {
	query := s.db.Client(ctx).KnowledgeRelationship.Query().
		Where(knr.ID(id)).
		WithAliases().
		WithSourceEntity().
		WithTargetEntity()
	return query.Only(ctx)
}

func (s *KnowledgeGraphService) GetEntityAt(ctx context.Context, id uuid.UUID, referencedAt time.Time) (*ent.KnowledgeEntity, error) {
	queryEntity := s.db.Client(ctx).KnowledgeEntity.Query().
		Where(kne.ID(id)).
		WithAliases()
	entity, queryErr := queryEntity.Only(ctx)
	if queryErr != nil {
		return nil, queryErr
	} else if referencedAt.IsZero() {
		return entity, nil
	}

	queryEvidence := s.db.Client(ctx).KnowledgeEvidence.Query().
		Where(ke.KindNEQ(ke.KindDeleted), ke.EffectiveAtLTE(referencedAt), ke.HasSubjectAliasWith(ksa.EntityID(id))).
		WithSubjectAlias().
		Order(ent.Desc(ke.FieldEffectiveAt), ent.Desc(ke.FieldCreatedAt))
	evidence, evidenceErr := queryEvidence.First(ctx)
	if evidenceErr != nil {
		return nil, fmt.Errorf("query evidence: %w", evidenceErr)
	}
	fmt.Printf("use evidence state: %v\n", evidence.SubjectState)
	return entity, nil
}

func (s *KnowledgeGraphService) GetRelationshipAt(ctx context.Context, id uuid.UUID, referencedAt time.Time) (*ent.KnowledgeRelationship, error) {
	return nil, fmt.Errorf("todo")
}

func (s *KnowledgeGraphService) GetView(ctx context.Context, rootID uuid.UUID, params rez.GetKnowledgeGraphViewParams) (*rez.KnowledgeGraphView, error) {
	return nil, fmt.Errorf("todo")
}
