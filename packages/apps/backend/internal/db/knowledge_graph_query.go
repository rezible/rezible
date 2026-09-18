package db

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
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
	"github.com/rezible/rezible/pkg/execution"
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

// TODO: clean this up

type knowledgeGraphLevelCursor struct {
	QueryHash string
	Offset    int
}

type knowledgeGraphLevelRecord struct {
	entity       *ent.KnowledgeEntity
	relationship *ent.KnowledgeRelationship
}

// selectKnowledgeGraph reads current canonical rows only. It loads the full tenant
// graph per request, returns direct relationships without descendant aggregation,
// and does not provide a consistency guarantee across concurrent changes between pages.
func (s *KnowledgeGraphQueryService) selectKnowledgeGraph(
	ctx context.Context,
	query rez.KnowledgeGraphLevelQuery,
) (map[uuid.UUID]*ent.KnowledgeEntity, []*ent.KnowledgeEntity, []*ent.KnowledgeRelationship, error) {
	tenantID, tenantOK := execution.GetContext(ctx).TenantID()
	if !tenantOK {
		return nil, nil, nil, rez.ErrTenantContextMissing
	}
	if query.Level < rez.KnowledgeGraphDetailLevelLandscape || query.Level > rez.KnowledgeGraphDetailLevelImplementation {
		return nil, nil, nil, fmt.Errorf("%w: knowledge graph detail level %d is outside 0-3", rez.ErrInvalidInput, query.Level)
	}

	rootIDs, normalizeErr := normalizeKnowledgeGraphRootIDs(query.Root.Lens.EntityIDs)
	if normalizeErr != nil {
		return nil, nil, nil, normalizeErr
	}
	if len(rootIDs) == 0 {
		return make(map[uuid.UUID]*ent.KnowledgeEntity), make([]*ent.KnowledgeEntity, 0), make([]*ent.KnowledgeRelationship, 0), nil
	}

	entityQuery := s.db.Client(ctx).KnowledgeEntity.Query().
		Where(kne.TenantID(tenantID)).
		WithAliases(func(aliasQuery *ent.KnowledgeSubjectAliasQuery) {
			aliasQuery.Where(ksa.TenantID(tenantID))
		}).
		Order(kne.ByID())
	entities, entityQueryErr := entityQuery.All(ctx)
	if entityQueryErr != nil {
		return nil, nil, nil, fmt.Errorf("load canonical knowledge graph entities: %w", entityQueryErr)
	}

	relationshipQuery := s.db.Client(ctx).KnowledgeRelationship.Query().
		Where(knr.TenantID(tenantID)).
		WithAliases(func(aliasQuery *ent.KnowledgeSubjectAliasQuery) {
			aliasQuery.Where(ksa.TenantID(tenantID))
		}).
		Order(knr.ByID())
	relationships, relationshipQueryErr := relationshipQuery.All(ctx)
	if relationshipQueryErr != nil {
		return nil, nil, nil, fmt.Errorf("load canonical knowledge graph relationships: %w", relationshipQueryErr)
	}

	entitiesByID := make(map[uuid.UUID]*ent.KnowledgeEntity, len(entities))
	for _, entity := range entities {
		entitiesByID[entity.ID] = entity
	}
	for _, rootID := range rootIDs {
		if _, rootExists := entitiesByID[rootID]; !rootExists {
			return nil, nil, nil, fmt.Errorf("%w: knowledge graph root %s", rez.ErrNotFound, rootID)
		}
	}

	childrenByParent := make(map[uuid.UUID][]uuid.UUID)
	for _, relationship := range relationships {
		if relationship.Predicate != knr.PredicateContains {
			continue
		}
		if _, sourceExists := entitiesByID[relationship.SourceEntityID]; !sourceExists {
			continue
		}
		if _, targetExists := entitiesByID[relationship.TargetEntityID]; !targetExists {
			continue
		}
		childrenByParent[relationship.SourceEntityID] = append(childrenByParent[relationship.SourceEntityID], relationship.TargetEntityID)
	}

	scopeIDs := mapset.NewSet[uuid.UUID]()
	queue := append(make([]uuid.UUID, 0, len(rootIDs)), rootIDs...)
	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]
		if !scopeIDs.Add(currentID) {
			continue
		}
		for _, childID := range childrenByParent[currentID] {
			if !scopeIDs.Contains(childID) {
				queue = append(queue, childID)
			}
		}
	}

	eligibleByID := make(map[uuid.UUID]*ent.KnowledgeEntity, scopeIDs.Cardinality())
	eligibleEntities := make([]*ent.KnowledgeEntity, 0, scopeIDs.Cardinality())
	for entityID := range scopeIDs.Iter() {
		entity := entitiesByID[entityID]
		entityLevel, mapped := rez.KnowledgeGraphDetailLevels[entity.Category]
		eligible := mapped && entityLevel <= query.Level
		if !eligible {
			continue
		}
		eligibleByID[entity.ID] = entity
		eligibleEntities = append(eligibleEntities, entity)
	}
	sort.Slice(eligibleEntities, func(i, j int) bool {
		return bytes.Compare(eligibleEntities[i].ID[:], eligibleEntities[j].ID[:]) < 0
	})

	selectedRelationships := make([]*ent.KnowledgeRelationship, 0, len(relationships))
	for _, relationship := range relationships {
		if _, sourceEligible := eligibleByID[relationship.SourceEntityID]; !sourceEligible {
			continue
		}
		if _, targetEligible := eligibleByID[relationship.TargetEntityID]; !targetEligible {
			continue
		}
		selectedRelationships = append(selectedRelationships, relationship)
	}
	sort.Slice(selectedRelationships, func(i, j int) bool {
		return bytes.Compare(selectedRelationships[i].ID[:], selectedRelationships[j].ID[:]) < 0
	})

	return eligibleByID, eligibleEntities, selectedRelationships, nil
}

func normalizeKnowledgeGraphRootIDs(rootIDs []uuid.UUID) ([]uuid.UUID, error) {
	normalized := make([]uuid.UUID, 0, len(rootIDs))
	seen := make(map[uuid.UUID]struct{}, len(rootIDs))
	for _, rootID := range rootIDs {
		if rootID == uuid.Nil {
			return nil, fmt.Errorf("%w: knowledge graph root ID cannot be nil", rez.ErrInvalidInput)
		}
		if _, duplicate := seen[rootID]; duplicate {
			continue
		}
		seen[rootID] = struct{}{}
		normalized = append(normalized, rootID)
	}
	sort.Slice(normalized, func(i, j int) bool {
		return bytes.Compare(normalized[i][:], normalized[j][:]) < 0
	})
	return normalized, nil
}

func knowledgeGraphQueryHash(tenantID int, rootIDs []uuid.UUID, level rez.KnowledgeGraphDetailLevel) (string, error) {
	rootStrings := make([]string, 0, len(rootIDs))
	for _, rootID := range rootIDs {
		rootStrings = append(rootStrings, rootID.String())
	}
	hashInput := struct {
		TenantID int      `json:"tenant_id"`
		RootIDs  []string `json:"root_ids"`
		Level    int      `json:"level"`
	}{
		TenantID: tenantID,
		RootIDs:  rootStrings,
		Level:    int(level),
	}
	encodedInput, marshalErr := json.Marshal(hashInput)
	if marshalErr != nil {
		return "", fmt.Errorf("hash knowledge graph query: %w", marshalErr)
	}
	digest := sha256.Sum256(encodedInput)
	return fmt.Sprintf("%x", digest), nil
}

func encodeKnowledgeGraphLevelCursor(cursor knowledgeGraphLevelCursor) (rez.KnowledgeGraphPageCursor, error) {
	encodedCursor, marshalErr := json.Marshal(cursor)
	if marshalErr != nil {
		return "", fmt.Errorf("%w: encode knowledge graph cursor: %w", rez.ErrInvalidInput, marshalErr)
	}
	return rez.KnowledgeGraphPageCursor(base64.RawURLEncoding.EncodeToString(encodedCursor)), nil
}

func decodeKnowledgeGraphLevelCursor(encodedCursor rez.KnowledgeGraphPageCursor) (knowledgeGraphLevelCursor, error) {
	var cursor knowledgeGraphLevelCursor
	decodedCursor, decodeErr := base64.RawURLEncoding.DecodeString(string(encodedCursor))
	if decodeErr != nil {
		decodedCursor, decodeErr = base64.URLEncoding.DecodeString(string(encodedCursor))
	}
	if decodeErr != nil {
		return cursor, fmt.Errorf("%w: malformed knowledge graph cursor", rez.ErrInvalidInput)
	}
	if unmarshalErr := json.Unmarshal(decodedCursor, &cursor); unmarshalErr != nil || cursor.QueryHash == "" {
		return cursor, fmt.Errorf("%w: malformed knowledge graph cursor", rez.ErrInvalidInput)
	}
	return cursor, nil
}

func (s *KnowledgeGraphQueryService) QueryGraphLevel(
	ctx context.Context,
	query rez.KnowledgeGraphLevelQuery,
	pageParams rez.KnowledgeGraphPageParams,
) (rez.KnowledgeGraphLevelQueryResult, error) {
	result := rez.KnowledgeGraphLevelQueryResult{
		Entities:    make(ent.KnowledgeEntities, 0),
		Connections: make([]rez.KnowledgeGraphConnectionAggregate, 0),
	}
	if pageParams.Limit < 0 {
		return result, fmt.Errorf("%w: knowledge graph page limit cannot be negative", rez.ErrInvalidInput)
	}

	eligibleByID, eligibleEntities, relationships, selectionErr := s.selectKnowledgeGraph(ctx, query)
	if selectionErr != nil {
		return result, selectionErr
	}
	tenantID, tenantOK := execution.GetContext(ctx).TenantID()
	if !tenantOK {
		return result, rez.ErrTenantContextMissing
	}
	rootIDs, normalizeErr := normalizeKnowledgeGraphRootIDs(query.Root.Lens.EntityIDs)
	if normalizeErr != nil {
		return result, normalizeErr
	}
	queryHash, hashErr := knowledgeGraphQueryHash(tenantID, rootIDs, query.Level)
	if hashErr != nil {
		return result, hashErr
	}

	offset := 0
	if pageParams.Cursor != "" {
		cursor, decodeErr := decodeKnowledgeGraphLevelCursor(pageParams.Cursor)
		if decodeErr != nil {
			return result, decodeErr
		}
		if cursor.QueryHash != queryHash || cursor.Offset < 0 {
			return result, fmt.Errorf("%w: knowledge graph cursor does not match query", rez.ErrInvalidInput)
		}
		offset = cursor.Offset
	}

	records := make([]knowledgeGraphLevelRecord, 0, len(eligibleEntities)+len(relationships))
	for _, entity := range eligibleEntities {
		records = append(records, knowledgeGraphLevelRecord{entity: entity})
	}
	for _, relationship := range relationships {
		records = append(records, knowledgeGraphLevelRecord{relationship: relationship})
	}
	total := len(records)
	if offset > total {
		return result, fmt.Errorf("%w: knowledge graph cursor offset is beyond result", rez.ErrInvalidInput)
	}
	limit := pageParams.Limit
	if limit == 0 {
		limit = 100
	}
	take := limit
	if remaining := total - offset; take > remaining {
		take = remaining
	}
	end := offset + take

	pageEntities := make(map[uuid.UUID]*ent.KnowledgeEntity, take)
	for _, record := range records[offset:end] {
		if record.entity != nil {
			pageEntities[record.entity.ID] = record.entity
			continue
		}
		relationship := record.relationship
		result.Connections = append(result.Connections, rez.KnowledgeGraphConnectionAggregate{
			Key: rez.KnowledgeGraphConnection{
				SourceRepresentativeID: relationship.SourceEntityID,
				TargetRepresentativeID: relationship.TargetEntityID,
				Predicate:              relationship.Predicate,
			},
			RelationshipCount: 1,
		})
		pageEntities[relationship.SourceEntityID] = eligibleByID[relationship.SourceEntityID]
		pageEntities[relationship.TargetEntityID] = eligibleByID[relationship.TargetEntityID]
	}
	for _, entity := range pageEntities {
		result.Entities = append(result.Entities, entity)
	}
	sort.Slice(result.Entities, func(i, j int) bool {
		return bytes.Compare(result.Entities[i].ID[:], result.Entities[j].ID[:]) < 0
	})
	if end < total {
		nextCursor, cursorErr := encodeKnowledgeGraphLevelCursor(knowledgeGraphLevelCursor{
			QueryHash: queryHash,
			Offset:    end,
		})
		if cursorErr != nil {
			return result, cursorErr
		}
		result.NextCursor = nextCursor
	}
	return result, nil
}

func (s *KnowledgeGraphQueryService) ListKnowledgeGraphConnectionAggregateRelationships(
	ctx context.Context,
	params rez.ListKnowledgeGraphConnectionAggregateRelationshipsParams,
) (*ent.ListResult[ent.KnowledgeRelationship], error) {
	if params.Connection.SourceRepresentativeID == uuid.Nil || params.Connection.TargetRepresentativeID == uuid.Nil {
		return nil, fmt.Errorf("%w: connection endpoint IDs cannot be nil", rez.ErrInvalidInput)
	}
	if predicateErr := knr.PredicateValidator(params.Connection.Predicate); predicateErr != nil {
		return nil, fmt.Errorf("%w: invalid connection predicate: %w", rez.ErrInvalidInput, predicateErr)
	}

	_, _, relationships, selectionErr := s.selectKnowledgeGraph(ctx, params.Query)
	if selectionErr != nil {
		return nil, selectionErr
	}
	var selectedRelationshipID uuid.UUID
	for _, relationship := range relationships {
		if relationship.SourceEntityID != params.Connection.SourceRepresentativeID ||
			relationship.TargetEntityID != params.Connection.TargetRepresentativeID ||
			relationship.Predicate != params.Connection.Predicate {
			continue
		}
		selectedRelationshipID = relationship.ID
		break
	}

	page := params.ListParams.GetPage()
	pageSize := params.ListParams.GetPageSize()
	if selectedRelationshipID == uuid.Nil {
		return &ent.ListResult[ent.KnowledgeRelationship]{
			Data:     make([]*ent.KnowledgeRelationship, 0),
			Page:     page,
			PageSize: pageSize,
			Total:    0,
		}, nil
	}
	relationshipQuery := s.db.Client(ctx).KnowledgeRelationship.Query().
		Where(knr.ID(selectedRelationshipID)).
		WithAliases(func(aliasQuery *ent.KnowledgeSubjectAliasQuery) {
			aliasQuery.Where()
		}).
		WithSourceEntity(func(entityQuery *ent.KnowledgeEntityQuery) {
			entityQuery.WithAliases()
		}).
		WithTargetEntity(func(entityQuery *ent.KnowledgeEntityQuery) {
			entityQuery.WithAliases()
		}).
		Order(knr.ByID(params.GetOrder()))
	return ent.DoListQuery[ent.KnowledgeRelationship, *ent.KnowledgeRelationshipQuery](ctx, relationshipQuery, params.ListParams)
}
