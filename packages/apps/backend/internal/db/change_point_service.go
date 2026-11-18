package db

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	cp "github.com/rezible/rezible/ent/changepoint"
	cpi "github.com/rezible/rezible/ent/changepointinvolvement"
	ke "github.com/rezible/rezible/ent/knowledgeevidence"
	ksa "github.com/rezible/rezible/ent/knowledgesubjectalias"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/projections"
)

type ChangePointService struct {
	db rez.Database
}

var _ rez.ChangePointService = (*ChangePointService)(nil)

func NewChangePointService(db rez.Database) (*ChangePointService, error) {
	return &ChangePointService{db: db}, nil
}

var changePointIdentityColumns = sql.ConflictColumns(
	cp.FieldTenantID,
	cp.FieldKind,
	cp.FieldProvider,
	cp.FieldProviderNamespace,
	cp.FieldProviderResourceRef,
)

func (s *ChangePointService) RecordChangePoint(ctx context.Context, event *ent.NormalizedEvent, occurredAt time.Time) error {
	return s.db.WithTx(ctx, func(ctx context.Context, tx *ent.Client) error {
		createPoint := tx.ChangePoint.Create().
			SetKind(event.Kind).
			SetProvider(event.Provider).
			SetProviderNamespace(event.ProviderNamespace).
			SetProviderResourceRef(event.ProviderResourceRef).
			SetOccurredAt(occurredAt).
			SetLatestEventID(event.ID).
			OnConflict(changePointIdentityColumns).
			Ignore()
		if createErr := createPoint.Exec(ctx); createErr != nil {
			return fmt.Errorf("create change point: %w", createErr)
		}

		// Two events of one change point projected concurrently would each compare against the stored latest
		// event and could leave the earlier one stored, or the other's involvement. Locking the change point's
		// row serializes them, so each compares against the other's result. It is the only lock taken here,
		// after the identity locks evidence ingestion took earlier in the transaction.
		queryPoint := tx.ChangePoint.Query().
			Where(
				cp.Kind(event.Kind),
				cp.Provider(event.Provider),
				cp.ProviderNamespace(event.ProviderNamespace),
				cp.ProviderResourceRef(event.ProviderResourceRef),
			).
			WithLatestEvent().
			ForUpdate()
		point, queryErr := queryPoint.Only(ctx)
		if queryErr != nil {
			return fmt.Errorf("load change point: %w", queryErr)
		}
		if !projections.ChangePointEventAfter(event, occurredAt, point.Edges.LatestEvent, point.OccurredAt) {
			if point.LatestEventID != event.ID {
				return nil
			}
			// The event is already the latest: the row was just created for it, or the event is reprocessed.
			queryInvolvement := tx.ChangePointInvolvement.Query().Where(cpi.ChangePointID(point.ID))
			hasInvolvement, existsErr := queryInvolvement.Exist(ctx)
			if existsErr != nil {
				return fmt.Errorf("query change point involvement: %w", existsErr)
			}
			if hasInvolvement {
				return nil
			}
			return s.replaceInvolvement(ctx, point.ID, event)
		}
		updatePoint := tx.ChangePoint.UpdateOneID(point.ID).
			SetOccurredAt(occurredAt).
			SetLatestEventID(event.ID)
		if updateErr := updatePoint.Exec(ctx); updateErr != nil {
			return fmt.Errorf("update change point: %w", updateErr)
		}
		return s.replaceInvolvement(ctx, point.ID, event)
	})
}

// replaceInvolvement replaces the change point's involvement rows with one per involvement evidence row of the
// event, its new latest event.
func (s *ChangePointService) replaceInvolvement(ctx context.Context, pointID uuid.UUID, event *ent.NormalizedEvent) error {
	client := s.db.Client(ctx)
	deleteInvolvement := client.ChangePointInvolvement.Delete().Where(cpi.ChangePointID(pointID))
	if _, deleteErr := deleteInvolvement.Exec(ctx); deleteErr != nil {
		return fmt.Errorf("delete change point involvement: %w", deleteErr)
	}
	queryEvidence := client.KnowledgeEvidence.Query().
		Where(ke.EventID(event.ID), ke.AssertionIn(projections.InvolvementAssertions(event.Kind)...))
	evidence, evidenceErr := queryEvidence.All(ctx)
	if evidenceErr != nil {
		return fmt.Errorf("query involvement evidence: %w", evidenceErr)
	}
	if len(evidence) == 0 {
		return nil
	}
	return client.ChangePointInvolvement.MapCreateBulk(evidence, func(create *ent.ChangePointInvolvementCreate, index int) {
		create.SetChangePointID(pointID).
			SetRole(evidence[index].Assertion).
			SetSubjectAliasID(evidence[index].SubjectAliasID).
			SetEvidenceID(evidence[index].ID)
	}).Exec(ctx)
}

func (s *ChangePointService) ListChangePoints(ctx context.Context, params rez.ListChangePointsParams) (*rez.ChangePoints, error) {
	if params.Kind == "" || params.EntityID == uuid.Nil {
		return nil, fmt.Errorf("%w: change point kind and entity are required", errs.ErrInvalidInput)
	}
	if params.Limit < 1 || params.Limit > 100 {
		return nil, fmt.Errorf("%w: limit must be 1 to 100", errs.ErrInvalidInput)
	}
	kindRoles := projections.InvolvementAssertions(params.Kind)
	if len(kindRoles) == 0 {
		return nil, fmt.Errorf("%w: %q records no change points", errs.ErrInvalidInput, params.Kind)
	}
	for _, role := range params.Roles {
		if !slices.Contains(kindRoles, role) {
			return nil, fmt.Errorf("%w: %q is not an involvement role of %q", errs.ErrInvalidInput, role, params.Kind)
		}
	}
	roles := params.Roles
	if len(roles) == 0 {
		roles = kindRoles
	}
	attributeFilter, filtersErr := s.attributeFilter(params)
	if filtersErr != nil {
		return nil, filtersErr
	}
	client := s.db.Client(ctx)
	entityID, canonicalErr := ent.CanonicalKnowledgeEntityID(ctx, client, params.EntityID)
	if canonicalErr != nil {
		return nil, fmt.Errorf("resolve entity redirect: %w", canonicalErr)
	}

	query := client.ChangePoint.Query().
		Where(
			cp.Kind(params.Kind),
			cp.HasInvolvementWith(
				cpi.RoleIn(roles...),
				cpi.HasSubjectAliasWith(ksa.EntityID(entityID)),
			),
		).
		WithLatestEvent().
		WithInvolvement(func(q *ent.ChangePointInvolvementQuery) {
			q.Order(cpi.ByEvidenceID()).
				WithSubjectAlias()
		}).
		Order(cp.ByOccurredAt(sql.OrderDesc()), cp.ByID()).
		Limit(params.Limit + 1)
	if attributeFilter != nil {
		query.Where(attributeFilter)
	}
	if !params.From.IsZero() {
		if params.FromInclusive {
			query.Where(cp.OccurredAtGTE(params.From))
		} else {
			query.Where(cp.OccurredAtGT(params.From))
		}
	}
	if !params.To.IsZero() {
		if params.ToInclusive {
			query.Where(cp.OccurredAtLTE(params.To))
		} else {
			query.Where(cp.OccurredAtLT(params.To))
		}
	}
	points, queryErr := query.All(ctx)
	if queryErr != nil {
		return nil, fmt.Errorf("query change points: %w", queryErr)
	}

	result := &rez.ChangePoints{}
	if len(points) > params.Limit {
		result.Cut = true
		points = points[:params.Limit]
	}
	for _, point := range points {
		item := rez.ChangePoint{
			ChangePoint: point,
			Involved:    point.Edges.Involvement,
		}
		for _, involvement := range item.Involved {
			if involvedID := involvement.Edges.SubjectAlias.EntityID; involvedID != nil && *involvedID == entityID && slices.Contains(roles, involvement.Role) {
				item.Involvement = involvement
				break
			}
		}
		result.ChangePoints = append(result.ChangePoints, item)
	}
	return result, nil
}

// attributePathSegment bounds attribute path segments, which the JSON path writes into the query text. Dots are
// allowed for linking attribute keys.
var attributePathSegment = regexp.MustCompile(`^[A-Za-z0-9_.]+$`)

// attributeFilter returns the attribute filters as one predicate on the attributes of the change point's latest
// event, or nil when there are none. Ent generates no predicates for values inside a JSON column, so these are the
// one hand-written comparison on change points; they keep filtering, ordering and limits in the database.
func (s *ChangePointService) attributeFilter(params rez.ListChangePointsParams) (predicate.ChangePoint, error) {
	var filters []predicate.NormalizedEvent
	for _, attribute := range params.AttributeEquals {
		path, pathErr := s.attributePath(attribute.Path)
		if pathErr != nil {
			return nil, pathErr
		}
		value := attribute.Value
		// attributes->'a'->>'b' = value
		filters = append(filters, func(sel *sql.Selector) {
			sel.Where(sqljson.ValueEQ(sel.C(ne.FieldAttributes), value, path))
		})
	}
	for _, attribute := range params.AttributeAbsent {
		path, pathErr := s.attributePath(attribute)
		if pathErr != nil {
			return nil, pathErr
		}
		// attributes->'a'->>'b' IS NULL, for a missing key or a JSON null
		filters = append(filters, func(sel *sql.Selector) {
			sel.Where(sql.P(func(b *sql.Builder) {
				b.Join(sqljson.ValuePath(sel.C(ne.FieldAttributes), path, sqljson.Unquote(true))).
					WriteString(" IS NULL")
			}))
		})
	}
	if len(filters) == 0 {
		return nil, nil
	}
	return cp.HasLatestEventWith(filters...), nil
}

func (s *ChangePointService) attributePath(path rez.AttributePath) (sqljson.Option, error) {
	segments := strings.Split(string(path), "/")
	for _, segment := range segments {
		if !attributePathSegment.MatchString(segment) {
			return nil, fmt.Errorf("%w: invalid attribute path %q", errs.ErrInvalidInput, path)
		}
	}
	return sqljson.Path(segments...), nil
}
