package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
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

const (
	maxKnowledgeViewDepth         = 4
	maxKnowledgeViewEntities      = 200
	maxKnowledgeViewRelationships = 400
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

func (s *KnowledgeGraphService) aliasWithEvidenceQuery(evP ...predicate.KnowledgeEvidence) func(*ent.KnowledgeSubjectAliasQuery) {
	return func(aq *ent.KnowledgeSubjectAliasQuery) {
		aq.WithEvidence(func(eq *ent.KnowledgeEvidenceQuery) {
			eq.Where(evP...)
			eq.Order(ent.Desc(ke.FieldEffectiveAt), ent.Desc(ke.FieldCreatedAt), ent.Desc(ke.FieldID))
			eq.WithEvent()
			eq.Limit(1)
		})
	}
}

func (s *KnowledgeGraphService) ListEntities(ctx context.Context, params rez.ListKnowledgeGraphEntitiesParams) (*ent.ListResult[ent.KnowledgeEntity], error) {
	query := s.db.Client(ctx).KnowledgeEntity.Query().
		WithAliases(s.aliasWithEvidenceQuery())

	if search := strings.TrimSpace(params.Search); search != "" {
		searchMatch := kne.Or(
			kne.KindContainsFold(search),
			kne.HasAliasesWith(
				ksa.HasEvidenceWith(func(selector *sql.Selector) {
					selector.Where(sql.P(func(builder *sql.Builder) {
						builder.Ident(selector.C(ke.FieldSubjectState)).
							WriteString("::text ILIKE ").
							Arg("%" + search + "%")
					}))
				}),
			),
		)
		query.Where(searchMatch)
	}
	if len(params.Predicates) > 0 {
		query.Where(params.Predicates...)
	}

	return ent.DoListQuery[ent.KnowledgeEntity, *ent.KnowledgeEntityQuery](ctx, query, params.ListParams)
}

func (s *KnowledgeGraphService) entityQueryWithEvidence(ctx context.Context, id uuid.UUID, evP ...predicate.KnowledgeEvidence) *ent.KnowledgeEntityQuery {
	return s.db.Client(ctx).KnowledgeEntity.Query().
		Where(kne.ID(id)).
		WithAliases(s.aliasWithEvidenceQuery(append(evP, ke.HasSubjectAliasWith(ksa.HasEntityWith(kne.ID(id))))...))
}

func (s *KnowledgeGraphService) GetEntity(ctx context.Context, id uuid.UUID) (*ent.KnowledgeEntity, error) {
	return s.entityQueryWithEvidence(ctx, id).Only(ctx)
}

func (s *KnowledgeGraphService) GetEntityAt(ctx context.Context, id uuid.UUID, referencedAt time.Time) (*ent.KnowledgeEntity, error) {
	return s.entityQueryWithEvidence(ctx, id, ke.EffectiveAtLTE(referencedAt)).Only(ctx)
}

func (s *KnowledgeGraphService) relationshipQueryWithEvidence(ctx context.Context, id uuid.UUID, evP ...predicate.KnowledgeEvidence) *ent.KnowledgeRelationshipQuery {
	return s.db.Client(ctx).KnowledgeRelationship.Query().
		Where(knr.ID(id)).
		WithAliases(s.aliasWithEvidenceQuery(append(evP, ke.HasSubjectAliasWith(ksa.HasRelationshipWith(knr.ID(id))))...))
}

func (s *KnowledgeGraphService) GetEvidence(ctx context.Context, id uuid.UUID) (*ent.KnowledgeEvidence, error) {
	return s.db.Client(ctx).KnowledgeEvidence.Query().
		Where(ke.ID(id)).
		WithEvent().
		WithSubjectAlias().
		Only(ctx)
}

func (s *KnowledgeGraphService) ListRelationships(ctx context.Context, params rez.ListKnowledgeGraphRelationshipsParams) (*ent.ListResult[ent.KnowledgeRelationship], error) {
	query := s.db.Client(ctx).KnowledgeRelationship.Query().
		WithAliases(s.aliasWithEvidenceQuery())
	if len(params.Predicates) > 0 {
		query.Where(params.Predicates...)
	}
	return ent.DoListQuery[ent.KnowledgeRelationship, *ent.KnowledgeRelationshipQuery](ctx, query, params.ListParams)
}

func (s *KnowledgeGraphService) GetRelationship(ctx context.Context, id uuid.UUID) (*ent.KnowledgeRelationship, error) {
	return s.relationshipQueryWithEvidence(ctx, id).Only(ctx)
}

func (s *KnowledgeGraphService) GetRelationshipAt(ctx context.Context, id uuid.UUID, referencedAt time.Time) (*ent.KnowledgeRelationship, error) {
	return s.relationshipQueryWithEvidence(ctx, id, ke.EffectiveAtLTE(referencedAt)).Only(ctx)
}

func (s *KnowledgeGraphService) GetView(ctx context.Context, params rez.GetKnowledgeGraphViewParams) (*rez.KnowledgeGraphView, error) {
	depth := min(maxKnowledgeViewDepth, max(1, params.Depth))

	view := &rez.KnowledgeGraphView{RootID: params.EntityID}

	if view.RootID == uuid.Nil {
		// TODO: find entity with the most relationships
		queryPopular := s.db.Client(ctx).KnowledgeRelationship.Query().
			Where()
		rel, relErr := queryPopular.First(ctx)
		if relErr != nil {
			if ent.IsNotFound(relErr) {
				return view, nil
			}
			return nil, fmt.Errorf("query graph relationship: %w", relErr)
		}
		view.RootID = rel.SourceEntityID
	}

	entityIDs := mapset.NewSet[uuid.UUID]()
	relationshipIDs := mapset.NewSet[uuid.UUID]()
	queryFrontier := func(ids []uuid.UUID) ([]uuid.UUID, error) {
		queryRels := s.db.Client(ctx).KnowledgeRelationship.Query().
			Where(knr.Or(knr.SourceEntityIDIn(ids...), knr.TargetEntityIDIn(ids...))).
			WithAliases(s.aliasWithEvidenceQuery())
		if len(params.RelationshipKinds) > 0 {
			queryRels.Where(knr.KindIn(params.RelationshipKinds...))
		}
		queryRels.Limit(maxKnowledgeViewRelationships + 1)

		relationships, queryErr := queryRels.All(ctx)
		if queryErr != nil {
			return nil, fmt.Errorf("query graph relationships: %w", queryErr)
		}

		nextEntityIDs := mapset.NewSet[uuid.UUID]()
		for _, rel := range relationships {
			if relationshipIDs.Cardinality() >= maxKnowledgeViewRelationships {
				view.Truncated = true
				break
			}
			if relationshipIDs.Add(rel.ID) {
				view.Relationships = append(view.Relationships, rel)
				if !entityIDs.Contains(rel.SourceEntityID) {
					nextEntityIDs.Add(rel.SourceEntityID)
				}
				if !entityIDs.Contains(rel.TargetEntityID) {
					nextEntityIDs.Add(rel.TargetEntityID)
				}
				if entityIDs.Cardinality()+nextEntityIDs.Cardinality() >= maxKnowledgeViewEntities {
					view.Truncated = true
					break
				}
			}
		}

		nextIDs := nextEntityIDs.ToSlice()
		if len(nextIDs) > 0 {
			queryEntities := s.db.Client(ctx).KnowledgeEntity.Query().
				Where(kne.IDIn(nextIDs...)).
				WithAliases(s.aliasWithEvidenceQuery())
			nextEntities, queryEntitiesErr := queryEntities.All(ctx)
			if queryEntitiesErr != nil {
				return nil, fmt.Errorf("query entities: %w", queryEntitiesErr)
			}
			for _, e := range nextEntities {
				if entityIDs.Add(e.ID) {
					view.Entities = append(view.Entities, e)
				}
			}
		}

		return nextIDs, nil
	}

	frontierIDs := []uuid.UUID{view.RootID}
	for level := 0; level <= depth && len(frontierIDs) > 0; level++ {
		nextIDs, queryFrontierErr := queryFrontier(frontierIDs)
		if queryFrontierErr != nil {
			return nil, queryFrontierErr
		}
		frontierIDs = nextIDs
	}

	return view, nil
}
