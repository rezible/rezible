package db

import (
	"cmp"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"entgo.io/ent/dialect/sql"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knea "github.com/rezible/rezible/ent/knowledgeentityancestry"
	kner "github.com/rezible/rezible/ent/knowledgeentityrepresentation"
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

func (s *KnowledgeGraphQueryService) QueryEntities(ctx context.Context, params rez.QueryKnowledgeGraphParams) (*rez.KnowledgeGraphEntitiesPage, error) {
	scopeIds := mapset.NewSet(params.EntityIDs...).ToSlice()
	slices.SortFunc(scopeIds, func(id1 uuid.UUID, id2 uuid.UUID) int {
		return cmp.Compare(id1.String(), id2.String())
	})

	limit := 200
	if params.Limit <= 500 && params.Limit > 0 {
		limit = params.Limit
	}

	tenantId, tenantOK := execution.GetContext(ctx).TenantID()
	if !tenantOK {
		return nil, rez.ErrInvalidTenant
	}

	fingerprint := graphQueryFingerprint{
		TenantID:  tenantId,
		QueryType: "entities",
		RootIDs:   scopeIds,
		Level:     params.DetailLevel,
	}
	fpHash, hashErr := fingerprint.makeHash()
	if hashErr != nil {
		return nil, fmt.Errorf("make graph query fingerprint: %w", hashErr)
	}

	var cursor graphQueryCursor
	if params.Cursor != nil {
		if cursorErr := cursor.decode(*params.Cursor); cursorErr != nil {
			return nil, cursorErr
		}
		if fpHash != cursor.QueryFingerprintHash {
			return nil, fmt.Errorf("invalid fingerprint")
		}
	}

	candidates := s.db.Client(ctx).KnowledgeEntityRepresentation.Query().
		Where(kner.DetailLevelEQ(int(params.DetailLevel)))

	if len(scopeIds) > 0 {
		entityHasAncestorInScope := kne.HasAncestryLinksWith(knea.AncestorIDIn(scopeIds...))
		candidates.Where(kner.HasEntityWith(entityHasAncestorInScope))
	}

	entitiesQuery := candidates.QueryRepresentative().
		Unique(true)

	pageQuery := entitiesQuery.Clone().
		Order(kne.ByID(sql.OrderAsc())).
		Limit(limit + 1)

	if cursor.AfterRepresentativeID != nil {
		pageQuery.Where(kne.IDGT(*cursor.AfterRepresentativeID))
	}

	ids, idsErr := pageQuery.IDs(ctx)
	if idsErr != nil {
		return nil, fmt.Errorf("query representative ids: %w", idsErr)
	}

	hasMore := len(ids) > limit
	if hasMore {
		ids = ids[:limit]
	}

	result, pageErr := s.queryEntitiesPage(ctx, ids)
	if pageErr != nil {
		return nil, pageErr
	}

	totalCount, countErr := entitiesQuery.Clone().Count(ctx)
	if countErr != nil {
		return nil, fmt.Errorf("count representatives: %w", countErr)
	}
	result.TotalCount = totalCount

	if hasMore {
		nextCursor := graphQueryCursor{
			QueryFingerprintHash:  fpHash,
			AfterRepresentativeID: new(ids[len(ids)-1]),
		}
		cursorEnc, encErr := nextCursor.encode()
		if encErr != nil {
			return nil, fmt.Errorf("encode entity cursor: %w", encErr)
		}
		result.NextCursor = &cursorEnc
	}

	return result, nil
}

func (s *KnowledgeGraphQueryService) queryEntitiesPage(ctx context.Context, ids []uuid.UUID) (*rez.KnowledgeGraphEntitiesPage, error) {
	if len(ids) == 0 {
		return &rez.KnowledgeGraphEntitiesPage{}, nil
	}

	const maxAncestorContext = 2_000

	client := s.db.Client(ctx)

	queryAncestors := client.KnowledgeEntityAncestry.Query().
		Where(knea.DescendantIDIn(ids...), knea.DepthGT(0)).
		QueryAncestor().
		Where(kne.IDNotIn(ids...)).
		Unique(true).
		Order(ent.Asc(kne.FieldID)).
		Limit(maxAncestorContext + 1)
	ancestorIDs, ancestorsErr := queryAncestors.IDs(ctx)
	if ancestorsErr != nil {
		return nil, fmt.Errorf("query ancestor IDs: %w", ancestorsErr)
	}
	if len(ancestorIDs) > maxAncestorContext {
		return nil, fmt.Errorf("ancestor context exceeds %d entities; narrow the query scope", maxAncestorContext)
	}

	allIDs := make([]uuid.UUID, 0, len(ids)+len(ancestorIDs))
	allIDs = append(allIDs, ids...)
	allIDs = append(allIDs, ancestorIDs...)

	queryMeta := client.KnowledgeEntity.Query().
		Where(kne.IDIn(allIDs...)).
		Select(kne.FieldID, kne.FieldCategory, kne.FieldKind)
	records, recordsErr := queryMeta.All(ctx)
	if recordsErr != nil {
		return nil, fmt.Errorf("query entity metadata: %w", recordsErr)
	}

	// Do not silently return incomplete hierarchy context.
	if len(records) != len(allIDs) {
		return nil, fmt.Errorf("incomplete entity metadata: expected %d records, got %d", len(allIDs), len(records))
	}

	byID := make(map[uuid.UUID]*rez.KnowledgeGraphViewEntity, len(records))
	for _, record := range records {
		byID[record.ID] = &rez.KnowledgeGraphViewEntity{
			ID:        record.ID,
			Name:      record.Kind,
			Category:  record.Category,
			ParentIDs: []uuid.UUID{},
		}
	}

	queryParents := client.KnowledgeRelationship.Query().
		Where(
			knr.PredicateEQ(knr.PredicateContains),
			knr.SourceEntityIDIn(allIDs...),
			knr.TargetEntityIDIn(allIDs...),
		).
		Unique(true).
		Order(ent.Asc(knr.FieldTargetEntityID), ent.Asc(knr.FieldSourceEntityID)).
		Select(knr.FieldSourceEntityID, knr.FieldTargetEntityID)

	var parentLinks []struct {
		SourceEntityID uuid.UUID `json:"source_entity_id"`
		TargetEntityID uuid.UUID `json:"target_entity_id"`
	}
	if scanParentLinksErr := queryParents.Scan(ctx, &parentLinks); scanParentLinksErr != nil {
		return nil, fmt.Errorf("query direct parents: %w", scanParentLinksErr)
	}

	for _, link := range parentLinks {
		child := byID[link.TargetEntityID]
		child.ParentIDs = append(child.ParentIDs, link.SourceEntityID)
	}

	countAncestors := client.KnowledgeEntityAncestry.Query().
		Where(knea.DescendantIDIn(allIDs...), knea.DepthGT(0)).
		GroupBy(knea.FieldDescendantID).
		Aggregate(ent.Count())

	var ancestorCounts []struct {
		DescendantID uuid.UUID `json:"descendant_id"`
		Count        int64     `json:"count"`
	}
	if countAncestorsErr := countAncestors.Scan(ctx, &ancestorCounts); countAncestorsErr != nil {
		return nil, fmt.Errorf("count ancestors: %w", countAncestorsErr)
	}

	for _, row := range ancestorCounts {
		byID[row.DescendantID].AncestorCount = row.Count
	}

	countDescendants := client.KnowledgeEntityAncestry.Query().
		Where(knea.AncestorIDIn(allIDs...), knea.DepthGT(0)).
		GroupBy(knea.FieldAncestorID).
		Aggregate(ent.Count())

	var descendantCounts []struct {
		AncestorID uuid.UUID `json:"ancestor_id"`
		Count      int64     `json:"count"`
	}
	if descendantsErr := countDescendants.Scan(ctx, &descendantCounts); descendantsErr != nil {
		return nil, fmt.Errorf("count descendants: %w", descendantsErr)
	}

	for _, row := range descendantCounts {
		byID[row.AncestorID].DescendantCount = row.Count
	}

	result := &rez.KnowledgeGraphEntitiesPage{
		Entities:  make([]rez.KnowledgeGraphViewEntity, len(ids)),
		Ancestors: make([]rez.KnowledgeGraphViewEntity, len(ancestorIDs)),
	}

	for i, id := range ids {
		result.Entities[i] = *byID[id]
	}

	for i, id := range ancestorIDs {
		result.Ancestors[i] = *byID[id]
	}

	return result, nil
}

func (s *KnowledgeGraphQueryService) QueryConnections(ctx context.Context, params rez.QueryKnowledgeGraphParams) (*rez.KnowledgeGraphConnectionsPage, error) {
	// 1. Validate and normalize the request.
	//    - Resolve the graph/tenant and authorization scope from context.
	//    - Validate the detail level and require an expected graph timestamp.
	//    - Read the explicit endpoint representative IDs from params.
	//      These are NOT hierarchy scopes: do not expand them to descendants.
	//    - Sort and deduplicate endpoint IDs; enforce the configured list limit.
	//    - Empty endpoint IDs mean no connections, not the entire graph.
	//    - Default the page size to 500; require a value between 1 and 2,000.

	// 2. Validate the cursor.
	//    - Build a query fingerprint from the method, graph, authorization
	//      scope, detail level, normalized endpoint IDs, and any filters.
	//    - Decode the cursor, if present, and validate its fingerprint.
	//    - Extract its graph timestamp and last (source, target, predicate).
	//    - Its timestamp must agree with the expected timestamp in params.
	//    - Changing the endpoint set requires starting with an empty cursor.
	//    - Return ErrInvalidKnowledgeGraphQuery for invalid/mismatched inputs.

	// 3. Open a read-only REPEATABLE READ Ent transaction.
	//    - Defer rollback and execute every subsequent read through this tx.
	//    - Read the graph timestamp and compare it with the expected timestamp.
	//    - Return ErrKnowledgeGraphChanged on mismatch.
	//    - If the endpoint set is empty, commit and return an empty page with
	//      the validated timestamp.

	// 4. Validate the requested endpoints.
	//    - Confirm each endpoint is authorized and represents itself at the
	//      requested detail level in KnowledgeEntityRepresentation.
	//    - Reject invalid endpoints rather than silently changing the query.

	// 5. Construct the underlying relationship query.
	//    - Join KnowledgeEntityRepresentation twice, for source and target,
	//      using the requested detail level for both joins.
	//    - Retain relationships only when BOTH resulting representative IDs
	//      belong to the requested endpoint set.
	//    - Exclude contains and same-representative relationships.
	//    - Apply graph/tenant, authorization, and requested relationship filters.
	//    - Explicitly authorize underlying relationships and their endpoints;
	//      authorizing only the representative IDs is insufficient.
	//    - Execute any raw SQL through the transaction, with explicit access
	//      predicates rather than relying on generated Ent privacy behavior.

	// 6. Aggregate the complete matching relationship set.
	//    - GROUP BY source representative, target representative, predicate.
	//    - Count each underlying relationship once.
	//    - RelationshipCount is the complete count for that aggregate at this
	//      graph timestamp, not a page-local increment.
	//    - Do not limit raw relationships before aggregation.
	//    - Direction matters: A -> B and B -> A are separate connections.

	// 7. Paginate the aggregated connections.
	//    - Apply the cursor to the aggregated results using the tuple
	//      (source_id, target_id, predicate) > the cursor's last tuple.
	//    - ORDER BY the same tuple, with identical types and ordering rules.
	//    - Fetch limit + 1 aggregates; exclude the extra one from the response.
	//    - The limit bounds returned connections, not underlying relationships
	//      counted or necessarily the database work required.

	// 8. Assemble the response and finish the transaction.
	//    - Return Connections and GraphUpdatedAt.
	//    - A connection's identity is (detail level, source, target, predicate);
	//      it is not the ID of an individual stored relationship.
	//    - If another page exists, encode a cursor containing the query
	//      fingerprint, graph timestamp, and last RETURNED connection tuple.
	//    - Otherwise, leave NextCursor empty.
	//    - Commit the read transaction before returning; propagate errors.

	panic("implement me")
}

type graphQueryCursor struct {
	QueryFingerprintHash  string     `json:"h"`
	AfterRepresentativeID *uuid.UUID `json:"after"`
}

func (c *graphQueryCursor) encode() (string, error) {
	encodedCursor, marshalErr := json.Marshal(c)
	if marshalErr != nil {
		return "", fmt.Errorf("%w: encode knowledge graph cursor: %w", rez.ErrInvalidInput, marshalErr)
	}
	return base64.RawURLEncoding.EncodeToString(encodedCursor), nil
}

func (c *graphQueryCursor) decode(encoded string) error {
	decodedCursor, decodeErr := base64.RawURLEncoding.DecodeString(encoded)
	if decodeErr != nil {
		decodedCursor, decodeErr = base64.URLEncoding.DecodeString(encoded)
	}
	if decodeErr != nil {
		return fmt.Errorf("%w: malformed page cursor", rez.ErrInvalidInput)
	}
	if unmarshalErr := json.Unmarshal(decodedCursor, c); unmarshalErr != nil || c.QueryFingerprintHash == "" {
		return fmt.Errorf("%w: malformed page cursor", rez.ErrInvalidInput)
	}
	return nil
}

type graphQueryFingerprint struct {
	TenantID  int                           `json:"tenant_id"`
	QueryType string                        `json:"query_type"`
	RootIDs   []uuid.UUID                   `json:"root_ids"`
	Level     rez.KnowledgeGraphDetailLevel `json:"level"`
}

func (f *graphQueryFingerprint) makeHash() (string, error) {
	slices.SortFunc(f.RootIDs, func(a, b uuid.UUID) int {
		return cmp.Compare(a.String(), b.String())
	})
	encodedInput, marshalErr := json.Marshal(f)
	if marshalErr != nil {
		return "", fmt.Errorf("hash knowledge graph query: %w", marshalErr)
	}
	return fmt.Sprintf("%x", sha256.Sum256(encodedInput)), nil
}
