package db

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	kela "github.com/rezible/rezible/ent/knowledgeentitylinkingattribute"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/knowledgegraph"
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
	if len(params.Categories) > 0 {
		query.Where(kne.CategoryIn(params.Categories...))
	}
	if params.LinkingAttribute != "" {
		linking := []predicate.KnowledgeEntityLinkingAttribute{kela.Attribute(params.LinkingAttribute)}
		if params.LinkingValueContains != "" {
			linking = append(linking, kela.ValueContains(params.LinkingValueContains))
		}
		query.Where(kne.HasLinkingAttributesWith(linking...))
	}

	return ent.DoListQuery[ent.KnowledgeEntity, *ent.KnowledgeEntityQuery](ctx, query, params.ListParams)
}

func (s *KnowledgeGraphQueryService) GetEntity(ctx context.Context, id uuid.UUID, expand rez.KnowledgeEntityExpand) (*ent.KnowledgeEntity, error) {
	query := s.db.Client(ctx).KnowledgeEntity.Query().
		Where(kne.ID(id))
	if expand.AliasEvidence {
		query.WithAliases(subjectAliasWithEvidence())
	} else {
		query.WithAliases()
	}
	if expand.LinkingAttributes {
		query.WithLinkingAttributes()
	}
	return query.Only(ctx)
}

func (s *KnowledgeGraphQueryService) ListRelationships(ctx context.Context, params rez.ListKnowledgeRelationshipsParams) (*ent.ListResult[ent.KnowledgeRelationship], error) {
	query := s.db.Client(ctx).KnowledgeRelationship.Query().
		WithAliases()
	if len(params.Predicates) > 0 {
		query.Where(params.Predicates...)
	}
	if params.EntityID != uuid.Nil {
		var neighbour []predicate.KnowledgeEntity
		if len(params.NeighbourCategories) > 0 {
			neighbour = append(neighbour, kne.CategoryIn(params.NeighbourCategories...))
		}
		if len(params.ExcludedNeighbourCategories) > 0 {
			neighbour = append(neighbour, kne.CategoryNotIn(params.ExcludedNeighbourCategories...))
		}
		query.Where(knr.Or(
			knr.And(knr.SourceEntityID(params.EntityID), knr.HasTargetEntityWith(neighbour...)),
			knr.And(knr.TargetEntityID(params.EntityID), knr.HasSourceEntityWith(neighbour...)),
		))
	} else if len(params.NeighbourCategories)+len(params.ExcludedNeighbourCategories) > 0 {
		return nil, fmt.Errorf("%w: neighbour categories need an entity ID", errs.ErrInvalidInput)
	}
	if params.NewestEvidenceFirst {
		query.Order(s.byLatestEvidence, knr.ByID())
	} else {
		query.Order(knr.ByID(params.GetOrder()))
	}
	if params.Expand.Endpoints {
		query.WithSourceEntity().WithTargetEntity()
	}
	return ent.DoListQuery[ent.KnowledgeRelationship, *ent.KnowledgeRelationshipQuery](ctx, query, params.ListParams)
}

// byLatestEvidence orders relationships by the effective time of their latest evidence, newest first. Ent's
// edge ordering can't take the maximum across two hops (aliases, then evidence), so this is a correlated
// subquery.
func (s *KnowledgeGraphQueryService) byLatestEvidence(sel *sql.Selector) {
	aliases := sql.Table(ksa.Table)
	evidence := sql.Table(ke.Table)
	latest := sql.Dialect(sel.Dialect()).
		Select(sql.Max(evidence.C(ke.FieldEffectiveAt))).
		From(evidence).
		Join(aliases).On(evidence.C(ke.FieldSubjectAliasID), aliases.C(ksa.FieldID)).
		Where(sql.ColumnsEQ(aliases.C(ksa.FieldRelationshipID), sel.C(knr.FieldID)))
	sel.OrderExprFunc(func(b *sql.Builder) {
		b.WriteString("(").Join(latest).WriteString(") DESC NULLS LAST")
	})
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

func (s *KnowledgeGraphQueryService) GetEvidence(ctx context.Context, id uuid.UUID, expand rez.KnowledgeEvidenceExpand) (*ent.KnowledgeEvidence, error) {
	query := s.db.Client(ctx).KnowledgeEvidence.Query().
		Where(ke.ID(id))
	if expand.Subject {
		query.WithSubjectAlias(withAliasSubject)
	} else {
		query.WithSubjectAlias()
	}
	return query.Only(ctx)
}

// withAliasSubject loads the alias's entity, or its relationship and the relationship's endpoints.
func withAliasSubject(aq *ent.KnowledgeSubjectAliasQuery) {
	aq.WithEntity().WithRelationship(func(rq *ent.KnowledgeRelationshipQuery) {
		rq.WithSourceEntity().WithTargetEntity()
	})
}

// SummarizeEvidence summarizes nothing when given neither an entity nor aliases.
func (s *KnowledgeGraphQueryService) SummarizeEvidence(ctx context.Context, params rez.SummarizeKnowledgeEvidenceParams) (*ent.ListResult[rez.KnowledgeEvidenceSummary], error) {
	page, pageSize := params.GetPage(), params.GetPageSize()
	result := &ent.ListResult[rez.KnowledgeEvidenceSummary]{Data: make([]*rez.KnowledgeEvidenceSummary, 0), Page: page, PageSize: pageSize}
	if params.EntityID == uuid.Nil && len(params.AliasIDs) == 0 {
		return result, nil
	}
	client := s.db.Client(ctx)

	var predicates []predicate.KnowledgeEvidence
	if params.EntityID != uuid.Nil {
		queryEntity := client.KnowledgeEntity.Query().
			Where(kne.ID(params.EntityID))
		entityExists, existsErr := queryEntity.Exist(ctx)
		if existsErr != nil {
			return nil, fmt.Errorf("check knowledge entity: %w", existsErr)
		}
		if !entityExists {
			return nil, fmt.Errorf("%w: no knowledge entity %s", errs.ErrNotFound, params.EntityID)
		}
		onEntityOrRelationships := ksa.Or(
			ksa.EntityID(params.EntityID),
			ksa.HasRelationshipWith(knr.Or(knr.SourceEntityID(params.EntityID), knr.TargetEntityID(params.EntityID))),
		)
		predicates = append(predicates, ke.HasSubjectAliasWith(onEntityOrRelationships))
	}
	if len(params.AliasIDs) > 0 {
		predicates = append(predicates, ke.SubjectAliasIDIn(params.AliasIDs...))
	}
	if !params.Start.IsZero() {
		predicates = append(predicates, ke.EffectiveAtGTE(params.Start))
	}
	if !params.End.IsZero() {
		predicates = append(predicates, ke.EffectiveAtLTE(params.End))
	}
	groupColumns := []string{ke.FieldSubjectAliasID}
	if params.ByAssertion {
		groupColumns = append(groupColumns, ke.FieldAssertion)
	}

	// Ent's group-by has no count of groups or order by an aggregate, so both are written as SQL expressions.
	countGroups := client.KnowledgeEvidence.Query().
		Where(predicates...).
		Aggregate(func(sel *sql.Selector) string {
			columns := make([]string, len(groupColumns))
			for index, column := range groupColumns {
				columns[index] = sel.C(column)
			}
			return fmt.Sprintf("count(DISTINCT (%s))", strings.Join(columns, ", "))
		})
	total, countErr := countGroups.Int(ctx)
	if countErr != nil {
		return nil, fmt.Errorf("count evidence groups: %w", countErr)
	}
	result.Total = total
	if total == 0 {
		return result, nil
	}

	newestGroupFirst := func(sel *sql.Selector) {
		sel.OrderExpr(sql.Expr(fmt.Sprintf("max(%s) DESC", sel.C(ke.FieldEffectiveAt))))
		for _, column := range groupColumns {
			sel.OrderBy(sel.C(column))
		}
	}
	// Postgres has no aggregate for the row holding the greatest value, so this takes the first of the
	// ordered IDs.
	latestID := func(sel *sql.Selector) string {
		ordered := fmt.Sprintf("(array_agg(%s ORDER BY %s DESC, %s DESC))[1]",
			sel.C(ke.FieldID), sel.C(ke.FieldEffectiveAt), sel.C(ke.FieldID))
		return sql.As(ordered, "latest_id")
	}
	groupEvidence := client.KnowledgeEvidence.Query().
		Where(predicates...).
		Order(newestGroupFirst).
		Offset((page-1)*pageSize).
		Limit(pageSize).
		GroupBy(groupColumns[0], groupColumns[1:]...).
		Aggregate(
			ent.Count(),
			ent.As(ent.Min(ke.FieldEffectiveAt), "earliest_at"),
			ent.As(ent.Max(ke.FieldEffectiveAt), "latest_at"),
			latestID,
		)
	var groups []struct {
		SubjectAliasID uuid.UUID `json:"subject_alias_id"`
		Assertion      string    `json:"assertion"`
		Count          int       `json:"count"`
		EarliestAt     time.Time `json:"earliest_at"`
		LatestAt       time.Time `json:"latest_at"`
		LatestID       uuid.UUID `json:"latest_id"`
	}
	if scanErr := groupEvidence.Scan(ctx, &groups); scanErr != nil {
		return nil, fmt.Errorf("group evidence: %w", scanErr)
	}

	aliasIDs := make([]uuid.UUID, len(groups))
	latestIDs := make([]uuid.UUID, len(groups))
	for index, group := range groups {
		aliasIDs[index], latestIDs[index] = group.SubjectAliasID, group.LatestID
	}
	queryLatest := client.KnowledgeEvidence.Query().
		Where(ke.IDIn(latestIDs...)).
		Select(ke.FieldID, ke.FieldSubjectState)
	latest, latestErr := queryLatest.All(ctx)
	if latestErr != nil {
		return nil, fmt.Errorf("query latest evidence: %w", latestErr)
	}
	latestByID := make(map[uuid.UUID]*ent.KnowledgeEvidence, len(latest))
	for _, evidence := range latest {
		latestByID[evidence.ID] = evidence
	}
	var aliasesByID map[uuid.UUID]*ent.KnowledgeSubjectAlias
	if params.Expand.Alias {
		queryAliases := client.KnowledgeSubjectAlias.Query().
			Where(ksa.IDIn(aliasIDs...)).
			WithRelationship(func(rq *ent.KnowledgeRelationshipQuery) {
				rq.WithSourceEntity().WithTargetEntity()
			})
		aliases, aliasesErr := queryAliases.All(ctx)
		if aliasesErr != nil {
			return nil, fmt.Errorf("query evidence aliases: %w", aliasesErr)
		}
		aliasesByID = make(map[uuid.UUID]*ent.KnowledgeSubjectAlias, len(aliases))
		for _, alias := range aliases {
			aliasesByID[alias.ID] = alias
		}
	}

	for _, group := range groups {
		latestEvidence := latestByID[group.LatestID]
		alias := aliasesByID[group.SubjectAliasID]
		if latestEvidence == nil || (params.Expand.Alias && alias == nil) {
			return nil, fmt.Errorf("evidence of alias %s changed while it was summarized", group.SubjectAliasID)
		}
		result.Data = append(result.Data, &rez.KnowledgeEvidenceSummary{
			SubjectAliasID: group.SubjectAliasID,
			Alias:          alias,
			Assertion:      group.Assertion,
			Count:          group.Count,
			EarliestAt:     group.EarliestAt.UTC(),
			LatestAt:       group.LatestAt.UTC(),
			LatestID:       group.LatestID,
			LatestState:    latestEvidence.SubjectState,
		})
	}
	return result, nil
}

const (
	defaultEntityQueryLimit = 100
	maxEntityQueryLimit     = 500
)

func (s *KnowledgeGraphQueryService) SelectGraphEntities(ctx context.Context, params rez.SelectKnowledgeGraphEntitiesParams) (*rez.KnowledgeGraphEntitiesPage, error) {
	selection := &knowledgeEntitySelection{
		Categories: params.Filter.Categories,
		Kinds:      params.Filter.Kinds,
		Limit:      defaultEntityQueryLimit,
	}
	if params.Limit != nil {
		selection.Limit = *params.Limit
	}
	if normalizeErr := selection.normalize(); normalizeErr != nil {
		return nil, fmt.Errorf("normalize entity selection: %w", normalizeErr)
	}

	fingerprint, fingerprintErr := selection.makeEntityFingerprint()
	if fingerprintErr != nil {
		return nil, fmt.Errorf("fingerprint entity selection: %w", fingerprintErr)
	}

	if params.Cursor != nil {
		cursor, cursorErr := s.decodeGraphQueryCursor(string(*params.Cursor), fingerprint, true)
		if cursorErr != nil {
			return nil, fmt.Errorf("decode entity selection cursor: %w", cursorErr)
		}
		selection.AfterID = cursor.AfterID
	}

	reference, encodeErr := s.encodeEntitySelectionRef(selection)
	if encodeErr != nil {
		return nil, fmt.Errorf("encode entity selection reference: %w", encodeErr)
	}

	page := &rez.KnowledgeGraphEntitiesPage{
		Entities:           make([]*ent.KnowledgeEntity, 0),
		EntitySelectionRef: reference,
	}

	selectEntitySummaryFields := s.selectedEntitiesQuery(ctx, selection, +1).
		Select(kne.FieldID, kne.FieldCategory, kne.FieldKind)
	entities, queryErr := selectEntitySummaryFields.All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("select knowledge graph entities: %w", queryErr)
	}
	if len(entities) > selection.Limit {
		page.Entities = entities[:selection.Limit]
		descriptor := knowledgeGraphQueryCursor{
			Version:       knowledgeGraphQueryVersion,
			IsEntityQuery: true,
			AfterID:       new(page.Entities[len(page.Entities)-1].ID),
			Fingerprint:   fingerprint,
		}
		encodedCursor, cursorErr := descriptor.encode()
		if cursorErr != nil {
			return nil, fmt.Errorf("encode entity selection cursor: %w", cursorErr)
		}
		page.NextCursor = new(rez.EntitySelectionCursor(encodedCursor))
	} else if entities != nil {
		page.Entities = entities
	}
	return page, nil
}

func (s *KnowledgeGraphQueryService) selectedEntitiesQuery(ctx context.Context, selection *knowledgeEntitySelection, extraAmount int) *ent.KnowledgeEntityQuery {
	query := s.db.Client(ctx).KnowledgeEntity.Query()
	if len(selection.Categories) > 0 {
		query.Where(kne.CategoryIn(selection.Categories...))
	}
	if len(selection.Kinds) > 0 {
		query.Where(kne.KindIn(selection.Kinds...))
	}
	if selection.AfterID != nil {
		query.Where(kne.IDGT(*selection.AfterID))
	}
	return query.Order(kne.ByID()).Limit(selection.Limit + extraAmount)
}

const (
	defaultRelationshipLimit = 200
	maxRelationshipLimit     = 1000
)

func (s *KnowledgeGraphQueryService) ExpandGraphRelationships(ctx context.Context, params rez.ExpandKnowledgeGraphRelationshipsParams) (*rez.KnowledgeGraphRelationshipsPage, error) {
	selection, selectionErr := s.decodeEntitySelectionRef(params.EntitySelectionRef)
	if selectionErr != nil {
		return nil, fmt.Errorf("decode entity selection reference: %w", selectionErr)
	}

	predicates := slices.Clone(params.Predicates)
	for _, relationshipPredicate := range predicates {
		if validationErr := knr.PredicateValidator(relationshipPredicate); validationErr != nil {
			return nil, fmt.Errorf("%w: invalid relationship predicate", errs.ErrInvalidInput)
		}
	}
	slices.Sort(predicates)
	predicates = slices.Compact(predicates)

	limit := defaultRelationshipLimit
	if params.Limit != nil {
		if *params.Limit < 1 || *params.Limit > maxRelationshipLimit {
			return nil, fmt.Errorf("%w: limit is outside the supported range", errs.ErrInvalidInput)
		}
		limit = *params.Limit
	}

	fingerprint, fingerprintErr := selection.makeRelationshipFingerprint(predicates, limit)
	if fingerprintErr != nil {
		return nil, fmt.Errorf("fingerprint relationship expansion: %w", fingerprintErr)
	}

	var afterID *uuid.UUID
	if params.Cursor != nil {
		cursor, cursorErr := s.decodeGraphQueryCursor(string(*params.Cursor), fingerprint, false)
		if cursorErr != nil {
			return nil, fmt.Errorf("decode relationship expansion cursor: %w", cursorErr)
		}
		afterID = cursor.AfterID
	}

	entityQuery := s.selectedEntitiesQuery(ctx, selection, 0)
	entityIDs, entityQueryErr := entityQuery.IDs(ctx)
	if entityQueryErr != nil {
		return nil, fmt.Errorf("replay selected knowledge graph entities: %w", entityQueryErr)
	}

	page := &rez.KnowledgeGraphRelationshipsPage{
		Relationships: make([]*ent.KnowledgeRelationship, 0),
	}
	if len(entityIDs) == 0 {
		return page, nil
	}

	relationshipQuery := s.db.Client(ctx).KnowledgeRelationship.Query().
		Where(knr.Or(knr.SourceEntityIDIn(entityIDs...), knr.TargetEntityIDIn(entityIDs...))).
		Order(knr.ByID()).
		Limit(limit + 1)
	if len(predicates) > 0 {
		relationshipQuery.Where(knr.PredicateIn(predicates...))
	}
	if afterID != nil {
		relationshipQuery.Where(knr.IDGT(*afterID))
	}
	selectSummaryFields := relationshipQuery.Select(
		knr.FieldID,
		knr.FieldSourceEntityID,
		knr.FieldTargetEntityID,
		knr.FieldPredicate,
	)
	relationships, queryErr := selectSummaryFields.All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("query knowledge graph relationships: %w", queryErr)
	}
	if len(relationships) > limit {
		page.Relationships = relationships[:limit]
		descriptor := knowledgeGraphQueryCursor{
			Version:       knowledgeGraphQueryVersion,
			IsEntityQuery: false,
			AfterID:       new(page.Relationships[len(page.Relationships)-1].ID),
			Fingerprint:   fingerprint,
		}
		encodedCursor, cursorErr := descriptor.encode()
		if cursorErr != nil {
			return nil, fmt.Errorf("encode relationship expansion cursor: %w", cursorErr)
		}
		page.NextCursor = new(rez.RelationshipExpansionCursor(encodedCursor))
	} else if relationships != nil {
		page.Relationships = relationships
	}
	return page, nil
}

func (s *KnowledgeGraphQueryService) validateResolveStructureParams(p rez.ResolveStructureParams) error {
	if len(p.TargetCategories) == 0 {
		return fmt.Errorf("%w: target categories are required", errs.ErrInvalidInput)
	}
	for _, category := range p.TargetCategories {
		if validationErr := kne.CategoryValidator(category); validationErr != nil {
			return fmt.Errorf("%w: invalid target category", errs.ErrInvalidInput)
		}
	}
	if p.MaxDepth < 0 {
		return fmt.Errorf("%w: max depth must not be negative", errs.ErrInvalidInput)
	}
	return nil
}

func (s *KnowledgeGraphQueryService) ResolveStructure(ctx context.Context, params rez.ResolveStructureParams) (map[uuid.UUID][]uuid.UUID, error) {
	if paramsErr := s.validateResolveStructureParams(params); paramsErr != nil {
		return nil, paramsErr
	}
	if len(params.EntityIDs) == 0 {
		return make(map[uuid.UUID][]uuid.UUID), nil
	}

	client := s.db.Client(ctx)
	entitiesQuery := client.KnowledgeEntity.Query().
		Where(kne.IDIn(params.EntityIDs...)).
		Select(kne.FieldID, kne.FieldCategory)
	entities, entitiesErr := entitiesQuery.All(ctx)
	if entitiesErr != nil {
		return nil, fmt.Errorf("load entity categories: %w", entitiesErr)
	}

	graph := knowledgegraph.StructureGraph{
		Categories: make(map[uuid.UUID]kne.Category),
	}
	targets := mapset.NewSet(params.TargetCategories...)
	frontier := mapset.NewSet[uuid.UUID]()
	for _, entity := range entities {
		graph.Categories[entity.ID] = entity.Category
		if !targets.Contains(entity.Category) {
			frontier.Add(entity.ID)
		}
	}

	// Load one step of the hierarchy above the frontier at a time, with the parents' categories. Entities
	// already loaded are not climbed again, so a cycle ends the loading.
	for depth := 0; depth < params.MaxDepth && !frontier.IsEmpty(); depth++ {
		linksQuery := client.KnowledgeRelationship.Query().
			Where(
				knr.PredicateIn(knowledgegraph.StructuralPredicates...),
				knr.TargetEntityIDIn(frontier.ToSlice()...),
			).
			Select(knr.FieldSourceEntityID, knr.FieldTargetEntityID).
			WithSourceEntity(func(pq *ent.KnowledgeEntityQuery) {
				pq.Select(kne.FieldID, kne.FieldCategory)
			})
		links, linksErr := linksQuery.All(ctx)
		if linksErr != nil {
			return nil, fmt.Errorf("load structure links: %w", linksErr)
		}
		next := mapset.NewSet[uuid.UUID]()
		for _, link := range links {
			structureLink := knowledgegraph.StructureLink{
				Parent: link.SourceEntityID,
				Child:  link.TargetEntityID,
			}
			graph.Links = append(graph.Links, structureLink)
			if parent := link.Edges.SourceEntity; parent != nil {
				if graph.IsLoaded(parent.ID) {
					continue
				}
				graph.Categories[parent.ID] = parent.Category
				if !targets.Contains(parent.Category) {
					next.Add(parent.ID)
				}
			}
		}
		frontier = next
	}

	return knowledgegraph.Resolve(graph, params.EntityIDs, params.TargetCategories, params.MaxDepth), nil
}

// relatedEventPropertyKeyPattern bounds property filter keys, which the JSON path writes into the query text.
var relatedEventPropertyKeyPattern = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

func (s *KnowledgeGraphQueryService) validateListRelatedEventsParams(p rez.ListRelatedEventsParams) error {
	if p.Limit < 1 {
		return fmt.Errorf("%w: limit must be positive", errs.ErrInvalidInput)
	}
	if p.Kind == "" {
		return fmt.Errorf("%w: event kind is required", errs.ErrInvalidInput)
	}
	if validationErr := knr.PredicateValidator(p.Predicate); validationErr != nil {
		return fmt.Errorf("%w: invalid relationship predicate", errs.ErrInvalidInput)
	}
	for key := range p.PropertyEquals {
		if !relatedEventPropertyKeyPattern.MatchString(key) {
			return fmt.Errorf("%w: invalid property key", errs.ErrInvalidInput)
		}
	}
	for _, key := range p.PropertyAbsent {
		if !relatedEventPropertyKeyPattern.MatchString(key) {
			return fmt.Errorf("%w: invalid property key", errs.ErrInvalidInput)
		}
	}
	return nil
}

func (s *KnowledgeGraphQueryService) ListRelatedEvents(ctx context.Context, params rez.ListRelatedEventsParams) (*rez.RelatedEvents, error) {
	if paramsErr := s.validateListRelatedEventsParams(params); paramsErr != nil {
		return nil, paramsErr
	}

	query := s.db.Client(ctx).KnowledgeEntity.Query().
		Where(
			kne.CategoryEQ(kne.CategoryEvent),
			kne.Kind(params.Kind),
			kne.StateEffectiveAtNotNil(),
			kne.HasSourceRelationshipsWith(
				knr.PredicateEQ(params.Predicate),
				knr.TargetEntityID(params.EntityID),
			),
		)
	if !params.From.IsZero() {
		if params.FromInclusive {
			query.Where(kne.StateEffectiveAtGTE(params.From))
		} else {
			query.Where(kne.StateEffectiveAtGT(params.From))
		}
	}
	if !params.To.IsZero() {
		if params.ToInclusive {
			query.Where(kne.StateEffectiveAtLTE(params.To))
		} else {
			query.Where(kne.StateEffectiveAtLT(params.To))
		}
	}
	for key, value := range params.PropertyEquals {
		// state->'properties'->>key = value
		query.Where(func(sel *sql.Selector) {
			sel.Where(sqljson.ValueEQ(sel.C(kne.FieldState), value, sqljson.Path("properties", key)))
		})
	}
	for _, key := range params.PropertyAbsent {
		// state->'properties'->>key IS NULL
		query.Where(func(sel *sql.Selector) {
			sel.Where(sql.Not(sqljson.HasKey(sel.C(kne.FieldState), sqljson.Path("properties", key), sqljson.Unquote(true))))
		})
	}
	query.
		Order(kne.ByStateEffectiveAt(sql.OrderDesc()), kne.ByID()).
		Limit(params.Limit + 1)

	events, queryErr := query.All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("query related events: %w", queryErr)
	}
	result := &rez.RelatedEvents{
		Events: events,
	}
	if len(events) > params.Limit {
		result.Events = events[:params.Limit]
		result.Cut = true
	}
	return result, nil
}

const knowledgeGraphQueryVersion = 1

type knowledgeEntitySelection struct {
	Version    int            `json:"version"`
	Categories []kne.Category `json:"categories,omitempty"`
	Kinds      []string       `json:"kinds,omitempty"`
	AfterID    *uuid.UUID     `json:"afterId,omitempty"`
	Limit      int            `json:"limit"`
}

func (es *knowledgeEntitySelection) normalize() error {
	if es.Limit < 1 || es.Limit > maxEntityQueryLimit {
		return fmt.Errorf("%w: entity limit is outside the supported range", errs.ErrInvalidInput)
	}

	for _, category := range es.Categories {
		if validationErr := kne.CategoryValidator(category); validationErr != nil {
			return fmt.Errorf("%w: invalid entity category", errs.ErrInvalidInput)
		}
	}
	categories := slices.Clone(es.Categories)
	slices.Sort(categories)

	kinds := slices.Clone(es.Kinds)
	if slices.Contains(kinds, "") {
		return fmt.Errorf("%w: entity kinds must be non-empty", errs.ErrInvalidInput)
	}
	slices.Sort(kinds)

	es.Version = knowledgeGraphQueryVersion
	es.Categories = slices.Compact(categories)
	es.Kinds = slices.Compact(kinds)
	return nil
}

func (es *knowledgeEntitySelection) encodeFingerprint[F any](payload F) (string, error) {
	jsonPayload, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return "", fmt.Errorf("marshal fingerprint: %w", marshalErr)
	}
	hash := sha256.Sum256(jsonPayload)
	return hex.EncodeToString(hash[:]), nil
}

type selectKnowledgeEntitiesFingerprint struct {
	Categories []kne.Category `json:"categories,omitempty"`
	Kinds      []string       `json:"kinds,omitempty"`
	Limit      int            `json:"limit"`
}

func (es *knowledgeEntitySelection) makeEntityFingerprint() (string, error) {
	return es.encodeFingerprint(selectKnowledgeEntitiesFingerprint{
		Categories: es.Categories,
		Kinds:      es.Kinds,
		Limit:      es.Limit,
	})
}

type expandKnowledgeRelationshipsFingerprint struct {
	Selection  knowledgeEntitySelection `json:"selection"`
	Predicates []knr.Predicate          `json:"predicates,omitempty"`
	Limit      int                      `json:"limit"`
}

func (es *knowledgeEntitySelection) makeRelationshipFingerprint(predicates []knr.Predicate, limit int) (string, error) {
	return es.encodeFingerprint(expandKnowledgeRelationshipsFingerprint{
		Selection:  *es,
		Predicates: predicates,
		Limit:      limit,
	})
}

const (
	maxKnowledgeEntitySelectionRefSize        = 4 * 1024
	knowledgeEntitySelectionBoundaryAllowance = len(`,"afterId":""`) + 36
)

func (s *KnowledgeGraphQueryService) encodeEntitySelectionRef(es *knowledgeEntitySelection) (rez.EntitySelectionRef, error) {
	jsonBytes, marshalErr := json.Marshal(es)
	if marshalErr != nil {
		return "", fmt.Errorf("marshal entity selection reference: %w", marshalErr)
	}

	// check that accepted filters also fit on continuation pages with the afterId field set
	reservedSize := len(jsonBytes)
	if es.AfterID == nil {
		reservedSize += knowledgeEntitySelectionBoundaryAllowance
	}
	if base64.RawURLEncoding.EncodedLen(reservedSize) > maxKnowledgeEntitySelectionRefSize {
		return "", fmt.Errorf("%w: entity selection reference exceeds the maximum size", errs.ErrInvalidInput)
	}

	return rez.EntitySelectionRef(base64.RawURLEncoding.EncodeToString(jsonBytes)), nil
}

func (s *KnowledgeGraphQueryService) decodeRef[T any](ref string) (*T, error) {
	refJson, decodeErr := base64.RawURLEncoding.Strict().DecodeString(ref)
	if decodeErr != nil {
		return nil, fmt.Errorf("%w: invalid encoding", errs.ErrInvalidInput)
	}
	decoder := json.NewDecoder(bytes.NewReader(refJson))
	decoder.DisallowUnknownFields()
	var t T
	if jsonErr := decoder.Decode(&t); jsonErr != nil {
		return nil, fmt.Errorf("%w: invalid reference JSON", errs.ErrInvalidInput)
	}
	if _, trailingErr := decoder.Token(); trailingErr != io.EOF {
		return nil, fmt.Errorf("%w: trailing data in reference", errs.ErrInvalidInput)
	}
	return &t, nil
}

func (s *KnowledgeGraphQueryService) decodeEntitySelectionRef(ref rez.EntitySelectionRef) (*knowledgeEntitySelection, error) {
	if ref == "" || len(ref) > maxKnowledgeEntitySelectionRefSize {
		return nil, fmt.Errorf("%w: invalid entity selection reference size", errs.ErrInvalidInput)
	}
	selection, decodeErr := s.decodeRef[knowledgeEntitySelection](string(ref))
	if decodeErr != nil {
		return nil, fmt.Errorf("%w: invalid entity selection reference: %w", errs.ErrInvalidInput, decodeErr)
	}
	if selection.Version != knowledgeGraphQueryVersion {
		return nil, fmt.Errorf("%w: unsupported entity selection reference", errs.ErrInvalidInput)
	}
	if normalizeErr := selection.normalize(); normalizeErr != nil {
		return nil, fmt.Errorf("%w: invalid entity selection reference", errs.ErrInvalidInput)
	}
	return selection, nil
}

type knowledgeGraphQueryCursor struct {
	Version       int        `json:"version"`
	IsEntityQuery bool       `json:"is_entity_query"`
	AfterID       *uuid.UUID `json:"afterId"`
	Fingerprint   string     `json:"fingerprint"`
}

const maxQueryCursorSize = 1 * 1024

func (cursor *knowledgeGraphQueryCursor) encode() (string, error) {
	jsonBytes, marshalErr := json.Marshal(cursor)
	if marshalErr != nil {
		return "", fmt.Errorf("marshal query cursor: %w", marshalErr)
	}
	encoded := base64.RawURLEncoding.EncodeToString(jsonBytes)
	if len(encoded) > maxQueryCursorSize {
		return "", fmt.Errorf("%w: query cursor exceeds the maximum size", errs.ErrInvalidInput)
	}
	return encoded, nil
}

func (s *KnowledgeGraphQueryService) decodeGraphQueryCursor(encoded string, fingerprint string, isEntityQuery bool) (*knowledgeGraphQueryCursor, error) {
	if encoded == "" || len(encoded) > maxQueryCursorSize {
		return nil, fmt.Errorf("%w: invalid query cursor size", errs.ErrInvalidInput)
	}
	cursor, decodeErr := s.decodeRef[knowledgeGraphQueryCursor](encoded)
	if decodeErr != nil {
		return nil, fmt.Errorf("%w: invalid cursor: %w", errs.ErrInvalidInput, decodeErr)
	}
	if cursor.Version != knowledgeGraphQueryVersion || cursor.IsEntityQuery != isEntityQuery {
		return nil, fmt.Errorf("%w: unsupported query cursor", errs.ErrInvalidInput)
	}
	if cursor.Fingerprint != fingerprint {
		return nil, fmt.Errorf("%w: query cursor does not match the query", errs.ErrInvalidInput)
	}
	if cursor.AfterID == nil {
		return nil, fmt.Errorf("%w: missing query cursor boundary", errs.ErrInvalidInput)
	}
	return cursor, nil
}
