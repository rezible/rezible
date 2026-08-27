package db

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	return &KnowledgeGraphService{db: db}, nil
}

func knowledgeAliasWithEvidenceQuery(evP ...predicate.KnowledgeEvidence) func(*ent.KnowledgeSubjectAliasQuery) {
	return func(aq *ent.KnowledgeSubjectAliasQuery) {
		aq.WithEvidence(func(eq *ent.KnowledgeEvidenceQuery) {
			eq.Where(evP...)
		})
	}
}

func (s *KnowledgeGraphService) ListEntities(ctx context.Context, params rez.ListKnowledgeGraphEntitiesParams) (*ent.ListResult[ent.KnowledgeEntity], error) {
	query := s.db.Client(ctx).KnowledgeEntity.Query().
		WithAliases(knowledgeAliasWithEvidenceQuery())

	if search := strings.TrimSpace(params.Search); search != "" {
		query.Where(kne.KindContainsFold(search))
	}
	if len(params.Predicates) > 0 {
		query.Where(params.Predicates...)
	}

	return ent.DoListQuery[ent.KnowledgeEntity, *ent.KnowledgeEntityQuery](ctx, query, params.ListParams)
}

func (s *KnowledgeGraphService) entityQueryWithEvidence(ctx context.Context, id uuid.UUID, evP ...predicate.KnowledgeEvidence) *ent.KnowledgeEntityQuery {
	evidencePreds := append(evP, ke.HasSubjectAliasWith(ksa.HasEntityWith(kne.ID(id))))
	return s.db.Client(ctx).KnowledgeEntity.Query().
		Where(kne.ID(id)).
		WithAliases(knowledgeAliasWithEvidenceQuery(evidencePreds...))
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
		WithAliases(knowledgeAliasWithEvidenceQuery(append(evP, ke.HasSubjectAliasWith(ksa.HasRelationshipWith(knr.ID(id))))...))
}

func (s *KnowledgeGraphService) ListRelationships(ctx context.Context, params rez.ListKnowledgeGraphRelationshipsParams) (*ent.ListResult[ent.KnowledgeRelationship], error) {
	query := s.db.Client(ctx).KnowledgeRelationship.Query().
		WithAliases(knowledgeAliasWithEvidenceQuery()).
		Order(knr.ByID(params.GetOrder()))
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

func (s *KnowledgeGraphService) ListEvidence(ctx context.Context, params rez.ListKnowledgeGraphEvidenceParams) (*ent.ListResult[ent.KnowledgeEvidence], error) {
	query := s.db.Client(ctx).KnowledgeEvidence.Query().
		Where(params.Predicates...).
		Order(ke.ByID(params.GetOrder()))
	if len(params.Predicates) > 0 {
		query.Where(params.Predicates...)
	}
	return ent.DoListQuery[ent.KnowledgeEvidence, *ent.KnowledgeEvidenceQuery](ctx, query, params.ListParams)
}

func (s *KnowledgeGraphService) GetEvidence(ctx context.Context, id uuid.UUID) (*ent.KnowledgeEvidence, error) {
	return s.db.Client(ctx).KnowledgeEvidence.Query().
		Where(ke.ID(id)).
		WithEvent().
		WithSubjectAlias().
		Only(ctx)
}

func (s *KnowledgeGraphService) QueryEntityNeighborhood(ctx context.Context, params rez.QueryKnowledgeEntityNeighborhoodParams) (*rez.KnowledgeGraphNeighborhoodSlice, error) {
	query := s.db.Client(ctx).KnowledgeRelationship.Query().
		Order(knr.ByID())

	rootEntityId, entityPreds, entityPredsErr := s.getRelationshipEntityPredicate(params)
	if entityPredsErr != nil {
		return nil, fmt.Errorf("get neighbor: %w", entityPredsErr)
	}

	query.Where(entityPreds)

	relationshipPredicates := mapset.NewSet[knr.Predicate]()
	for _, value := range params.RelationshipPredicates {
		rp := knr.Predicate(strings.TrimSpace(value))
		if predicateErr := knr.PredicateValidator(rp); predicateErr != nil {
			return nil, fmt.Errorf("invalid relationship predicate: %s", predicateErr)
		}
		relationshipPredicates.Add(rp)
	}
	if !relationshipPredicates.IsEmpty() {
		query.Where(knr.PredicateIn(relationshipPredicates.ToSlice()...))
	}

	numRels, numRelsErr := query.Count(ctx)
	if numRelsErr != nil {
		return nil, fmt.Errorf("query count: %w", numRelsErr)
	}

	limit := 20
	if params.Limit > 0 && params.Limit < 50 {
		limit = params.Limit
	}

	res := &rez.KnowledgeGraphNeighborhoodSlice{
		RootEntityID:      rootEntityId,
		Entities:          nil,
		Relationships:     nil,
		RelationshipCount: numRels,
	}

	query = query.Limit(limit).Offset(params.Offset)

	rels, relsErr := query.All(ctx)
	if relsErr != nil {
		return nil, fmt.Errorf("query neighborhood relationships: %w", relsErr)
	}
	res.Relationships = rels

	entityIds := mapset.NewSetWithSize[uuid.UUID](len(rels) * 2)
	for _, rel := range rels {
		entityIds.Append(rel.SourceEntityID, rel.TargetEntityID)
	}
	if !entityIds.IsEmpty() {
		queryEntities := s.db.Client(ctx).KnowledgeEntity.Query().
			Where(kne.IDIn(entityIds.ToSlice()...))
		ents, entsErr := queryEntities.All(ctx)
		if entsErr != nil {
			return nil, fmt.Errorf("query entities: %w", entsErr)
		}
		res.Entities = ents
	}

	return res, nil
}

func (s *KnowledgeGraphService) getRelationshipEntityPredicate(params rez.QueryKnowledgeEntityNeighborhoodParams) (uuid.UUID, predicate.KnowledgeRelationship, error) {
	neighborCategories := mapset.NewSet[kne.Category]()
	for _, value := range params.NeighborEntityCategories {
		category := kne.Category(strings.TrimSpace(value))
		if categoryErr := kne.CategoryValidator(category); categoryErr != nil {
			return uuid.Nil, nil, fmt.Errorf("invalid neighbor entity category: %s", categoryErr)
		}
		neighborCategories.Add(category)
	}
	neighborCategoriesPred := kne.CategoryIn(neighborCategories.ToSlice()...)

	sourceId := params.SourceEntityID
	targetId := params.TargetEntityID
	if params.EntityID != nil {
		sourceId = params.EntityID
		targetId = params.EntityID
	}
	var sourcePreds []predicate.KnowledgeEntity
	var targetPreds []predicate.KnowledgeEntity
	if sourceId != nil {
		sourcePreds = append(sourcePreds, kne.ID(*sourceId))
		if !neighborCategories.IsEmpty() {
			sourcePreds = append(sourcePreds, neighborCategoriesPred)
		}
	}
	if targetId != nil {
		targetPreds = append(targetPreds, kne.ID(*targetId))
		if !neighborCategories.IsEmpty() {
			targetPreds = append(targetPreds, neighborCategoriesPred)
		}
	}

	if params.EntityID != nil {
		return *params.EntityID, knr.Or(knr.HasSourceEntityWith(sourcePreds...), knr.HasTargetEntityWith(targetPreds...)), nil
	} else if sourceId != nil {
		return *sourceId, knr.HasSourceEntityWith(sourcePreds...), nil
	} else if targetId != nil {
		return *targetId, knr.HasTargetEntityWith(targetPreds...), nil
	}
	return uuid.Nil, nil, fmt.Errorf("no entity id")
}

func (s *KnowledgeGraphService) queryNeighborhoodGroupSummary(ctx context.Context, p predicate.KnowledgeRelationship) (map[string]rez.KnowledgeGraphNeighborhoodGroupSummary, error) {
	var rows []struct {
		Predicate knr.Predicate `json:"predicate"`
		Count     int           `json:"count"`
	}
	query := s.db.Client(ctx).KnowledgeRelationship.Query().
		Where(p).
		GroupBy(knr.FieldPredicate).
		Aggregate(ent.Count())
	if scanErr := query.Scan(ctx, &rows); scanErr != nil {
		return nil, scanErr
	}
	groups := map[string]rez.KnowledgeGraphNeighborhoodGroupSummary{}
	for _, row := range rows {
		groups[row.Predicate.String()] = rez.KnowledgeGraphNeighborhoodGroupSummary{
			Count: row.Count,
		}
	}
	return groups, nil
}

func (s *KnowledgeGraphService) SummarizeEntityNeighborhood(ctx context.Context, entityID uuid.UUID) (*rez.KnowledgeGraphEntityNeighborhoodSummary, error) {
	if entityID == uuid.Nil {
		return nil, fmt.Errorf("%w: entity ID is required", rez.ErrInvalidInput)
	}

	incomingGroups, queryIncomingErr := s.queryNeighborhoodGroupSummary(ctx, knr.TargetEntityID(entityID))
	if queryIncomingErr != nil {
		return nil, fmt.Errorf("query incoming: %w", queryIncomingErr)
	}

	outgoingGroups, queryOutgoingErr := s.queryNeighborhoodGroupSummary(ctx, knr.SourceEntityID(entityID))
	if queryOutgoingErr != nil {
		return nil, fmt.Errorf("query outgoing: %w", queryOutgoingErr)
	}

	summary := &rez.KnowledgeGraphEntityNeighborhoodSummary{
		IncomingRelationships: incomingGroups,
		OutgoingRelationships: outgoingGroups,
	}

	return summary, nil
}

func (s *KnowledgeGraphService) GetView(ctx context.Context, params rez.GetKnowledgeGraphViewParams) (*rez.KnowledgeGraphView, error) {
	depth := min(maxKnowledgeViewDepth, max(1, params.Depth))

	rootId := params.EntityID

	if rootId == uuid.Nil {
		// TODO: find entity with the most relationships
		queryPopular := s.db.Client(ctx).KnowledgeRelationship.Query().
			Where()
		rel, relErr := queryPopular.First(ctx)
		if relErr != nil {
			return nil, fmt.Errorf("query popular graph relationship: %w", relErr)
		}
		rootId = rel.SourceEntityID
	}

	rootEntity, rootErr := s.entityQueryWithEvidence(ctx, rootId).Only(ctx)
	if rootErr != nil {
		return nil, fmt.Errorf("query root entity: %w", rootErr)
	}

	entityIDs := mapset.NewSet[uuid.UUID](rootEntity.ID)
	entities := ent.KnowledgeEntities{rootEntity}

	relationshipIDs := mapset.NewSet[uuid.UUID]()
	relationships := ent.KnowledgeRelationships{}

	truncated := false

	queryFrontier := func(ids []uuid.UUID) ([]uuid.UUID, error) {
		queryRels := s.db.Client(ctx).KnowledgeRelationship.Query().
			Where(knr.Or(knr.SourceEntityIDIn(ids...), knr.TargetEntityIDIn(ids...))).
			WithAliases(knowledgeAliasWithEvidenceQuery()).
			Limit(maxKnowledgeViewRelationships + 1)
		if len(params.RelationshipPredicates) > 0 {
			predicates := make([]knr.Predicate, len(params.RelationshipPredicates))
			for i, value := range params.RelationshipPredicates {
				predicates[i] = knr.Predicate(value)
				if predicateErr := knr.PredicateValidator(predicates[i]); predicateErr != nil {
					return nil, predicateErr
				}
			}
			queryRels.Where(knr.PredicateIn(predicates...))
		}
		matchedRelationships, queryErr := queryRels.All(ctx)
		if queryErr != nil {
			return nil, fmt.Errorf("query graph relationships: %w", queryErr)
		}
		nextEntityIDs := mapset.NewSet[uuid.UUID]()
		for _, rel := range matchedRelationships {
			if relationshipIDs.Cardinality() >= maxKnowledgeViewRelationships {
				truncated = true
				break
			}
			if relationshipIDs.Add(rel.ID) {
				relationships = append(relationships, rel)
				if !entityIDs.Contains(rel.SourceEntityID) {
					nextEntityIDs.Add(rel.SourceEntityID)
				}
				if !entityIDs.Contains(rel.TargetEntityID) {
					nextEntityIDs.Add(rel.TargetEntityID)
				}
				if entityIDs.Cardinality()+nextEntityIDs.Cardinality() >= maxKnowledgeViewEntities {
					truncated = true
					break
				}
			}
		}

		nextIDs := nextEntityIDs.ToSlice()
		if len(nextIDs) > 0 {
			queryEntities := s.db.Client(ctx).KnowledgeEntity.Query().
				Where(kne.IDIn(nextIDs...)).
				WithAliases(knowledgeAliasWithEvidenceQuery())
			nextEntities, queryEntitiesErr := queryEntities.All(ctx)
			if queryEntitiesErr != nil {
				return nil, fmt.Errorf("query entities: %w", queryEntitiesErr)
			}
			for _, e := range nextEntities {
				if entityIDs.Add(e.ID) {
					entities = append(entities, e)
				}
			}
		}

		return nextIDs, nil
	}

	frontierIDs := []uuid.UUID{rootId}
	for level := 0; level <= depth && len(frontierIDs) > 0; level++ {
		nextIDs, queryFrontierErr := queryFrontier(frontierIDs)
		if queryFrontierErr != nil {
			return nil, queryFrontierErr
		}
		frontierIDs = nextIDs
	}

	return &rez.KnowledgeGraphView{
		RootID:        rootId,
		Entities:      entities,
		Relationships: relationships,
		Truncated:     truncated,
	}, nil
}
