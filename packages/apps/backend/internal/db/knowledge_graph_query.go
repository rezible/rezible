package db

import (
	"context"
	"fmt"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/ent/predicate"
)

type KnowledgeGraphQueryService struct {
	db rez.Database
}

func NewKnowledgeGraphQueryService(db rez.Database) (*KnowledgeGraphQueryService, error) {
	return &KnowledgeGraphQueryService{db: db}, nil
}

func (s *KnowledgeGraphQueryService) ListEntities(ctx context.Context, params rez.ListKnowledgeEntitiesParams) (*ent.ListResult[ent.KnowledgeEntity], error) {
	query := s.db.Client(ctx).KnowledgeEntity.Query().
		WithAliases().
		Order(kne.ByID(params.GetOrder()))

	if search := strings.TrimSpace(params.Search); search != "" {
		// explicitly do not support search for now
	}
	if len(params.Predicates) > 0 {
		query.Where(params.Predicates...)
	}

	return ent.DoListQuery[ent.KnowledgeEntity, *ent.KnowledgeEntityQuery](ctx, query, params.ListParams)
}

func (s *KnowledgeGraphQueryService) GetEntity(ctx context.Context, id uuid.UUID) (*ent.KnowledgeEntity, error) {
	return s.entityQueryWithEvidence(ctx, id).Only(ctx)
}

func (s *KnowledgeGraphQueryService) ListRelationships(ctx context.Context, params rez.ListKnowledgeRelationshipsParams) (*ent.ListResult[ent.KnowledgeRelationship], error) {
	query := s.db.Client(ctx).KnowledgeRelationship.Query().
		WithAliases().
		Order(knr.ByID(params.GetOrder()))
	if len(params.Predicates) > 0 {
		query.Where(params.Predicates...)
	}
	return ent.DoListQuery[ent.KnowledgeRelationship, *ent.KnowledgeRelationshipQuery](ctx, query, params.ListParams)
}

func (s *KnowledgeGraphQueryService) GetRelationship(ctx context.Context, id uuid.UUID) (*ent.KnowledgeRelationship, error) {
	return s.db.Client(ctx).KnowledgeRelationship.Query().
		Where(knr.ID(id)).
		WithAliases().
		WithSourceEntity(func(eq *ent.KnowledgeEntityQuery) {
			eq.WithAliases()
		}).
		WithTargetEntity(func(eq *ent.KnowledgeEntityQuery) {
			eq.WithAliases()
		}).
		Only(ctx)
}

func (s *KnowledgeGraphQueryService) relationshipQueryWithEvidence(ctx context.Context, id uuid.UUID, evP ...predicate.KnowledgeEvidence) *ent.KnowledgeRelationshipQuery {
	return s.db.Client(ctx).KnowledgeRelationship.Query().
		Where(knr.ID(id)).
		WithAliases(subjectAliasWithEvidence(evP...))
}

func (s *KnowledgeGraphQueryService) entityQueryWithEvidence(ctx context.Context, id uuid.UUID, evP ...predicate.KnowledgeEvidence) *ent.KnowledgeEntityQuery {
	return s.db.Client(ctx).KnowledgeEntity.Query().
		Where(kne.ID(id)).
		WithAliases(subjectAliasWithEvidence(evP...))
}

func subjectAliasWithEvidence(evidencePreds ...predicate.KnowledgeEvidence) func(*ent.KnowledgeSubjectAliasQuery) {
	return func(aq *ent.KnowledgeSubjectAliasQuery) {
		//aq.Where(subjectAliasPred)
		aq.WithEvidence(func(eq *ent.KnowledgeEvidenceQuery) {
			eq.Where(evidencePreds...)
		})
	}
}

func (s *KnowledgeGraphQueryService) ListSubjectAliases(ctx context.Context, params rez.ListKnowledgeSubjectAliasesParams) (*ent.ListResult[ent.KnowledgeSubjectAlias], error) {
	query := s.db.Client(ctx).KnowledgeSubjectAlias.Query().
		Where(params.Predicates...).
		Order(ksa.ByID(params.GetOrder()))
	if len(params.Predicates) > 0 {
		query.Where(params.Predicates...)
	}
	return ent.DoListQuery[ent.KnowledgeSubjectAlias, *ent.KnowledgeSubjectAliasQuery](ctx, query, params.ListParams)
}

func (s *KnowledgeGraphQueryService) GetSubjectAlias(ctx context.Context, id uuid.UUID) (*ent.KnowledgeSubjectAlias, error) {
	return s.db.Client(ctx).KnowledgeSubjectAlias.Query().
		Where(ksa.ID(id)).
		Only(ctx)
}

func (s *KnowledgeGraphQueryService) ListEvidence(ctx context.Context, params rez.ListKnowledgeEvidenceParams) (*ent.ListResult[ent.KnowledgeEvidence], error) {
	query := s.db.Client(ctx).KnowledgeEvidence.Query().
		Where(params.Predicates...).
		Order(ke.ByID(params.GetOrder()))
	if len(params.Predicates) > 0 {
		query.Where(params.Predicates...)
	}
	return ent.DoListQuery[ent.KnowledgeEvidence, *ent.KnowledgeEvidenceQuery](ctx, query, params.ListParams)
}

func (s *KnowledgeGraphQueryService) GetEvidence(ctx context.Context, id uuid.UUID) (*ent.KnowledgeEvidence, error) {
	return s.db.Client(ctx).KnowledgeEvidence.Query().
		Where(ke.ID(id)).
		WithSubjectAlias().
		Only(ctx)
}

const (
	maxStructureEntities      = 10_000
	maxStructureRelationships = 100_000
)

func (s *KnowledgeGraphQueryService) GetGraphStructureSnapshot(ctx context.Context) (*rez.KnowledgeGraphStructure, error) {
	structureCats := mapset.NewSetFromMapKeys(rez.KnowledgeGraphCategoryStructureLevels).ToSlice()

	queryStructureEnts := s.db.Client(ctx).KnowledgeEntity.Query().
		Where(kne.CategoryIn(structureCats...)).
		Limit(maxStructureEntities+1).
		Select(
			kne.FieldID,
			kne.FieldCategory,
			kne.FieldKind,
		)
	structureEnts, queryEntsErr := queryStructureEnts.All(ctx)
	if queryEntsErr != nil {
		return nil, fmt.Errorf("fetch structure entities: %w", queryEntsErr)
	}
	if len(structureEnts) > maxStructureEntities {
		return nil, fmt.Errorf("graph exceeds %d structural entities", maxStructureEntities)
	}

	structure := &rez.KnowledgeGraphStructure{}

	structure.Entities = make([]rez.KnowledgeGraphStructureEntity, len(structureEnts))
	ids := make([]uuid.UUID, len(structureEnts))
	for i, e := range structureEnts {
		ids[i] = e.ID
		structure.Entities[i] = rez.KnowledgeGraphStructureEntity{
			ID:       e.ID,
			Category: e.Category,
			Kind:     e.Kind,
		}
	}

	if len(ids) > 0 {
		queryStructureRels := s.db.Client(ctx).KnowledgeRelationship.Query().
			Where(knr.SourceEntityIDIn(ids...), knr.TargetEntityIDIn(ids...)).
			Limit(maxStructureRelationships+1).
			Select(
				knr.FieldID,
				knr.FieldSourceEntityID,
				knr.FieldTargetEntityID,
				knr.FieldPredicate,
			)
		structureRels, queryRelsErr := queryStructureRels.All(ctx)
		if queryRelsErr != nil {
			return nil, fmt.Errorf("query structure relationships: %w", queryRelsErr)
		}
		if len(structureRels) > maxStructureRelationships {
			return nil, fmt.Errorf("graph exceeds %d relationships", maxStructureRelationships)
		}
		structure.Relationships = make([]rez.KnowledgeGraphStructureRelationship, len(structureRels))
		for i, r := range structureRels {
			structure.Relationships[i] = rez.KnowledgeGraphStructureRelationship{
				ID:        r.ID,
				SourceID:  r.SourceEntityID,
				TargetID:  r.TargetEntityID,
				Predicate: r.Predicate,
			}
		}
	}

	return structure, nil
}
