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
	"log/slog"
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

// currentKnowledgeEntity keeps entities that are not withdrawn. Queries leave merged entities out already.
func currentKnowledgeEntity() predicate.KnowledgeEntity {
	return kne.WithdrawnAtIsNil()
}

// currentKnowledgeRelationship keeps relationships that are not withdrawn and whose endpoints are both current.
// Queries leave folded relationships out already, and a relationship that is not folded has no merged endpoint.
func currentKnowledgeRelationship() predicate.KnowledgeRelationship {
	return knr.And(
		knr.WithdrawnAtIsNil(),
		knr.HasSourceEntityWith(currentKnowledgeEntity()),
		knr.HasTargetEntityWith(currentKnowledgeEntity()),
	)
}

// withParentClaim loads, as an entity's target relationships, only the contains relationship chosen as its
// parent, for ent.KnowledgeEntity.ParentClaim.
func withParentClaim(rq *ent.KnowledgeRelationshipQuery) {
	rq.Where(knr.IsParent(true))
}

// otherParentClaim keeps the current contains claims on the entities that were not chosen as their parent,
// leaving out an entity's claim to contain itself.
func otherParentClaim(entityIDs []uuid.UUID) predicate.KnowledgeRelationship {
	return knr.And(
		knr.PredicateEQ(knr.PredicateContains),
		knr.TargetEntityIDIn(entityIDs...),
		knr.IsParent(false),
		func(s *sql.Selector) {
			s.Where(sql.ColumnsNEQ(s.C(knr.FieldSourceEntityID), s.C(knr.FieldTargetEntityID)))
		},
		currentKnowledgeRelationship(),
	)
}

// knowledgeAncestor is an entity's ancestor depth steps above it, read by ancestryQuery.
type knowledgeAncestor struct {
	EntityID   uuid.UUID `sql:"entity_id"`
	AncestorID uuid.UUID `sql:"source_entity_id"`
	Depth      int       `sql:"depth"`
	IsCycle    bool      `sql:"is_cycle"`
}

// ancestryQuery walks the flagged parent claims up from the entities in one recursive query, at most maxDepth
// steps, or all of them when maxDepth is 0. The generated query's WHERE, tenant filter included, applies at every
// level. A cycle is corrupt data: the step that closes one has IsCycle set, and the walk ends there.
func ancestryQuery(ctx context.Context, client *ent.Client, entityIDs []uuid.UUID, maxDepth int) ([]knowledgeAncestor, error) {
	var ancestors []knowledgeAncestor
	if len(entityIDs) == 0 {
		return ancestors, nil
	}
	starts := make([]any, len(entityIDs))
	for i, id := range entityIDs {
		starts[i] = id
	}
	walk := func(s *sql.Selector) {
		// The selector's predicates are added after the modifiers run, and read only when the query is built. They
		// carry bound arguments (is_parent and the tenant filter), which every ancestry read numbers correctly.
		generated := sql.P(func(b *sql.Builder) {
			b.Join(s.P())
		})
		builder := sql.Dialect(s.Dialect())
		relationships := s.Table()
		ancestry := sql.WithRecursive("knowledge_ancestry", "entity_id", "relationship_id", "ancestor_id", "depth")
		steps := builder.Table(ancestry.Name()).As("ancestry")
		start := builder.Select(
			relationships.C(knr.FieldTargetEntityID),
			relationships.C(knr.FieldID),
			relationships.C(knr.FieldSourceEntityID),
		).
			AppendSelectExpr(sql.Expr("1")).
			From(relationships).
			Where(sql.And(generated, sql.In(relationships.C(knr.FieldTargetEntityID), starts...)))
		climb := builder.Select(
			steps.C("entity_id"),
			relationships.C(knr.FieldID),
			relationships.C(knr.FieldSourceEntityID),
		).
			AppendSelectExpr(sql.Expr(steps.C("depth")+" + 1")).
			From(relationships).
			Join(steps).
			On(relationships.C(knr.FieldTargetEntityID), steps.C("ancestor_id")).
			Where(generated)
		if maxDepth > 0 {
			climb.Where(sql.LT(steps.C("depth"), maxDepth))
		}
		ancestry.As(start.UnionAll(climb))
		s.Prefix(cycleClause{ancestry}).
			Join(steps).
			On(s.C(knr.FieldID), steps.C("relationship_id")).
			AppendSelect(steps.C("entity_id"), steps.C("depth"), steps.C("is_cycle"))
	}
	queryAncestry := client.KnowledgeRelationship.Query().
		Where(knr.IsParent(true)).
		Select(knr.FieldSourceEntityID).
		Modify(walk)
	if queryErr := queryAncestry.Scan(ctx, &ancestors); queryErr != nil {
		return nil, fmt.Errorf("query ancestry: %w", queryErr)
	}
	return ancestors, nil
}

// cyclicAncestry returns the entities whose ancestry steps close a cycle.
func cyclicAncestry(ancestors []knowledgeAncestor) mapset.Set[uuid.UUID] {
	cyclic := mapset.NewSet[uuid.UUID]()
	for _, ancestor := range ancestors {
		if ancestor.IsCycle {
			cyclic.Add(ancestor.EntityID)
		}
	}
	return cyclic
}

// cycleClause adds Postgres's CYCLE clause to the single recursive CTE of ancestryQuery, marking a step that
// returns to an ancestor already on its path as is_cycle and ending the walk there.
type cycleClause struct {
	*sql.WithBuilder
}

func (c cycleClause) Query() (string, []any) {
	query, args := c.WithBuilder.Query()
	return query + " CYCLE ancestor_id SET is_cycle USING ancestry_path", args
}

func (s *KnowledgeGraphQueryService) ListEntities(ctx context.Context, params rez.ListKnowledgeEntitiesParams) (*ent.ListResult[ent.KnowledgeEntity], error) {
	query := s.db.Client(ctx).KnowledgeEntity.Query().
		WithAliases().
		WithTargetRelationships(withParentClaim).
		Order(kne.ByID(params.GetOrder()))
	if !params.IncludeWithdrawn {
		query.Where(currentKnowledgeEntity())
	}

	if search := strings.TrimSpace(params.Search); search != "" {
		query.Where(kne.Or(
			s.displayNameContainsFold(search),
			kne.HasLinkingAttributesWith(kela.ValueContainsFold(search)),
		))
	}
	if params.WithLinkingAttributes {
		query.WithLinkingAttributes()
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

// displayNameContainsFold matches entities whose state's display name contains the text, ignoring case. Ent
// has no case-folding predicate on a JSON value, so this joins its JSON path and LIKE predicates.
func (s *KnowledgeGraphQueryService) displayNameContainsFold(text string) predicate.KnowledgeEntity {
	return func(sel *sql.Selector) {
		sel.Where(sql.P(func(b *sql.Builder) {
			b.Join(sqljson.ValuePath(sel.C(kne.FieldState), sqljson.Path("display_name"), sqljson.Unquote(true)))
			b.Join(sql.ContainsFold("", text))
		}))
	}
}

// GetEntity returns a merged entity too, with its redirect.
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
	query.WithTargetRelationships(func(rq *ent.KnowledgeRelationshipQuery) {
		withParentClaim(rq)
		rq.WithAliases(liveClaimAliases)
	})
	return query.Only(ent.IncludeMerged(ctx))
}

func (s *KnowledgeGraphQueryService) ListRelationships(ctx context.Context, params rez.ListKnowledgeRelationshipsParams) (*ent.ListResult[ent.KnowledgeRelationship], error) {
	query := s.db.Client(ctx).KnowledgeRelationship.Query().
		WithAliases()
	if !params.IncludeWithdrawn {
		query.Where(currentKnowledgeRelationship())
	}
	if len(params.Predicates) > 0 {
		query.Where(params.Predicates...)
	}
	if params.EntityID != uuid.Nil {
		entityID, canonicalErr := ent.CanonicalKnowledgeEntityID(ctx, s.db.Client(ctx), params.EntityID)
		if canonicalErr != nil {
			return nil, fmt.Errorf("resolve entity redirect: %w", canonicalErr)
		}
		query.Where(knr.Or(knr.SourceEntityID(entityID), knr.TargetEntityID(entityID)))
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

// GetRelationship returns a folded relationship too, with its redirect and the endpoints it was left with, which
// may be merged entities.
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
		Only(ent.IncludeMerged(ctx))
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
		entityID, canonicalErr := ent.CanonicalKnowledgeEntityID(ctx, client, params.EntityID)
		if canonicalErr != nil {
			return nil, fmt.Errorf("resolve entity redirect: %w", canonicalErr)
		}
		queryEntity := client.KnowledgeEntity.Query().
			Where(kne.ID(entityID))
		entityExists, existsErr := queryEntity.Exist(ctx)
		if existsErr != nil {
			return nil, fmt.Errorf("check knowledge entity: %w", existsErr)
		}
		if !entityExists {
			return nil, fmt.Errorf("%w: no knowledge entity %s", errs.ErrNotFound, params.EntityID)
		}
		onEntityOrRelationships := ksa.Or(
			ksa.EntityID(entityID),
			ksa.HasRelationshipWith(knr.Or(knr.SourceEntityID(entityID), knr.TargetEntityID(entityID))),
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
	grouping := knowledgeEvidenceGrouping{
		predicates:  predicates,
		byAssertion: params.ByAssertion,
		expandAlias: params.Expand.Alias,
	}
	groupColumns := grouping.columns()

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
	summaries, summarizeErr := s.summarizeEvidenceGroups(ctx, grouping, (page-1)*pageSize, pageSize)
	if summarizeErr != nil {
		return nil, summarizeErr
	}
	result.Data = summaries
	return result, nil
}

// knowledgeEvidenceGrouping is the evidence a summary groups, and how.
type knowledgeEvidenceGrouping struct {
	predicates []predicate.KnowledgeEvidence
	// byAssertion groups each assertion on an alias separately.
	byAssertion bool
	// expandAlias loads each group's alias, and a relationship alias's relationship and its endpoints.
	expandAlias bool
}

func (g knowledgeEvidenceGrouping) columns() []string {
	columns := []string{ke.FieldSubjectAliasID}
	if g.byAssertion {
		columns = append(columns, ke.FieldAssertion)
	}
	return columns
}

// summarizeEvidenceGroups summarizes the groups of evidence, newest group first, skipping offset groups and
// reading at most limit, or every group when limit is zero.
func (s *KnowledgeGraphQueryService) summarizeEvidenceGroups(ctx context.Context, grouping knowledgeEvidenceGrouping, offset, limit int) ([]*rez.KnowledgeEvidenceSummary, error) {
	client := s.db.Client(ctx)
	groupColumns := grouping.columns()
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
	queryEvidence := client.KnowledgeEvidence.Query().
		Where(grouping.predicates...).
		Order(newestGroupFirst)
	if limit > 0 {
		queryEvidence.Offset(offset).Limit(limit)
	}
	groupEvidence := queryEvidence.
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
	if grouping.expandAlias {
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

	summaries := make([]*rez.KnowledgeEvidenceSummary, 0, len(groups))
	for _, group := range groups {
		latestEvidence := latestByID[group.LatestID]
		alias := aliasesByID[group.SubjectAliasID]
		if latestEvidence == nil || (grouping.expandAlias && alias == nil) {
			return nil, fmt.Errorf("evidence of alias %s changed while it was summarized", group.SubjectAliasID)
		}
		summaries = append(summaries, &rez.KnowledgeEvidenceSummary{
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
	return summaries, nil
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
		Select(kne.FieldID, kne.FieldCategory, kne.FieldKind, kne.FieldState, kne.FieldExpiresAt).
		WithTargetRelationships(withParentClaim)
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

	ids := make([]uuid.UUID, len(page.Entities))
	for i, entity := range page.Entities {
		ids[i] = entity.ID
	}
	otherParentIDs, otherParentsErr := s.listOtherParentIDs(ctx, ids)
	if otherParentsErr != nil {
		return nil, fmt.Errorf("list other parent ids: %w", otherParentsErr)
	}
	page.OtherParentIDs = otherParentIDs
	return page, nil
}

func (s *KnowledgeGraphQueryService) selectedEntitiesQuery(ctx context.Context, selection *knowledgeEntitySelection, extraAmount int) *ent.KnowledgeEntityQuery {
	query := s.db.Client(ctx).KnowledgeEntity.Query().
		Where(currentKnowledgeEntity())
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
		Where(currentKnowledgeRelationship()).
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
		knr.FieldExpiresAt,
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

func (s *KnowledgeGraphQueryService) ListAncestors(ctx context.Context, entityIDs []uuid.UUID) (map[uuid.UUID][]*ent.KnowledgeEntity, error) {
	return s.listAncestors(ctx, entityIDs, 0)
}

// listAncestors returns each entity's ancestors, nearest first, at most maxDepth of them, or all when maxDepth is
// 0. An ancestor that is not current ends an entity's chain. A cycle is logged and ends the chain before it
// repeats an entity, so corrupt parent claims do not fail every read of the chain.
func (s *KnowledgeGraphQueryService) listAncestors(ctx context.Context, entityIDs []uuid.UUID, maxDepth int) (map[uuid.UUID][]*ent.KnowledgeEntity, error) {
	ancestors := make(map[uuid.UUID][]*ent.KnowledgeEntity)
	client := s.db.Client(ctx)
	steps, ancestryErr := ancestryQuery(ctx, client, entityIDs, maxDepth)
	if ancestryErr != nil {
		return nil, ancestryErr
	}
	if cyclic := cyclicAncestry(steps); !cyclic.IsEmpty() {
		slog.ErrorContext(ctx, "parent claims form a cycle", "entity_ids", cyclic.ToSlice())
	}
	if len(steps) == 0 {
		return ancestors, nil
	}
	ancestorIDs := mapset.NewSet[uuid.UUID]()
	for _, step := range steps {
		ancestorIDs.Add(step.AncestorID)
	}
	queryEntities := client.KnowledgeEntity.Query().
		Where(kne.IDIn(ancestorIDs.ToSlice()...), currentKnowledgeEntity()).
		WithTargetRelationships(withParentClaim)
	entities, queryErr := queryEntities.All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("load ancestors: %w", queryErr)
	}
	loaded := make(map[uuid.UUID]*ent.KnowledgeEntity, len(entities))
	for _, entity := range entities {
		loaded[entity.ID] = entity
	}
	slices.SortFunc(steps, func(a, b knowledgeAncestor) int {
		return a.Depth - b.Depth
	})
	ended := mapset.NewSet[uuid.UUID]()
	for _, step := range steps {
		if ended.Contains(step.EntityID) {
			continue
		}
		ancestor, found := loaded[step.AncestorID]
		if !found || step.IsCycle || step.AncestorID == step.EntityID {
			ended.Add(step.EntityID)
			continue
		}
		ancestors[step.EntityID] = append(ancestors[step.EntityID], ancestor)
	}
	return ancestors, nil
}

func (s *KnowledgeGraphQueryService) ListParentClaims(ctx context.Context, entityIDs []uuid.UUID) (map[uuid.UUID][]*ent.KnowledgeRelationship, error) {
	claims := make(map[uuid.UUID][]*ent.KnowledgeRelationship)
	if len(entityIDs) == 0 {
		return claims, nil
	}
	queryClaims := s.db.Client(ctx).KnowledgeRelationship.Query().
		Where(otherParentClaim(entityIDs)).
		WithSourceEntity(func(eq *ent.KnowledgeEntityQuery) {
			eq.WithTargetRelationships(withParentClaim)
		}).
		WithAliases(liveClaimAliases)
	relationships, queryErr := queryClaims.All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("query parent claims: %w", queryErr)
	}
	slices.SortFunc(relationships, func(a, b *ent.KnowledgeRelationship) int {
		return knowledgegraph.RankParentClaims(parentClaimOf(a), parentClaimOf(b))
	})
	for _, relationship := range relationships {
		claims[relationship.TargetEntityID] = append(claims[relationship.TargetEntityID], relationship)
	}
	return claims, nil
}

// listOtherParentIDs reads only the source entity IDs of the claims ListParentClaims returns, in the same order.
func (s *KnowledgeGraphQueryService) listOtherParentIDs(ctx context.Context, entityIDs []uuid.UUID) (map[uuid.UUID][]uuid.UUID, error) {
	otherParentIDs := make(map[uuid.UUID][]uuid.UUID)
	if len(entityIDs) == 0 {
		return otherParentIDs, nil
	}
	queryClaims := s.db.Client(ctx).KnowledgeRelationship.Query().
		Where(otherParentClaim(entityIDs)).
		Select(knr.FieldTargetEntityID, knr.FieldSourceEntityID).
		WithAliases(liveParentClaimAliases)
	claims, queryErr := queryClaims.All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("query other parent claims: %w", queryErr)
	}
	slices.SortFunc(claims, func(a, b *ent.KnowledgeRelationship) int {
		return knowledgegraph.RankParentClaims(parentClaimOf(a), parentClaimOf(b))
	})
	for _, claim := range claims {
		otherParentIDs[claim.TargetEntityID] = append(otherParentIDs[claim.TargetEntityID], claim.SourceEntityID)
	}
	return otherParentIDs, nil
}

func (s *KnowledgeGraphQueryService) ResolveRepresentatives(ctx context.Context, entityIDs []uuid.UUID, categories []kne.Category, maxDepth int) (map[uuid.UUID]uuid.UUID, error) {
	representatives := make(map[uuid.UUID]uuid.UUID)
	if len(entityIDs) == 0 {
		return representatives, nil
	}
	entities, queryErr := s.db.Client(ctx).KnowledgeEntity.Query().
		Where(kne.IDIn(entityIDs...), currentKnowledgeEntity()).
		All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("load entities: %w", queryErr)
	}
	climbing := make([]uuid.UUID, 0, len(entities))
	for _, entity := range entities {
		if slices.Contains(categories, entity.Category) {
			representatives[entity.ID] = entity.ID
		} else {
			climbing = append(climbing, entity.ID)
		}
	}
	if maxDepth < 1 {
		return representatives, nil
	}
	ancestors, ancestorsErr := s.listAncestors(ctx, climbing, maxDepth)
	if ancestorsErr != nil {
		return nil, ancestorsErr
	}
	for id, chain := range ancestors {
		for _, ancestor := range chain {
			if slices.Contains(categories, ancestor.Category) {
				representatives[id] = ancestor.ID
				break
			}
		}
	}
	return representatives, nil
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
