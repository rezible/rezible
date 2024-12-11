// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"database/sql/driver"
	"fmt"
	"math"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/twohundreds/rezible/ent/incident"
	"github.com/twohundreds/rezible/ent/incidentdebriefquestion"
	"github.com/twohundreds/rezible/ent/incidenttag"
	"github.com/twohundreds/rezible/ent/predicate"
)

// IncidentTagQuery is the builder for querying IncidentTag entities.
type IncidentTagQuery struct {
	config
	ctx                  *QueryContext
	order                []incidenttag.OrderOption
	inters               []Interceptor
	predicates           []predicate.IncidentTag
	withIncidents        *IncidentQuery
	withDebriefQuestions *IncidentDebriefQuestionQuery
	modifiers            []func(*sql.Selector)
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the IncidentTagQuery builder.
func (itq *IncidentTagQuery) Where(ps ...predicate.IncidentTag) *IncidentTagQuery {
	itq.predicates = append(itq.predicates, ps...)
	return itq
}

// Limit the number of records to be returned by this query.
func (itq *IncidentTagQuery) Limit(limit int) *IncidentTagQuery {
	itq.ctx.Limit = &limit
	return itq
}

// Offset to start from.
func (itq *IncidentTagQuery) Offset(offset int) *IncidentTagQuery {
	itq.ctx.Offset = &offset
	return itq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (itq *IncidentTagQuery) Unique(unique bool) *IncidentTagQuery {
	itq.ctx.Unique = &unique
	return itq
}

// Order specifies how the records should be ordered.
func (itq *IncidentTagQuery) Order(o ...incidenttag.OrderOption) *IncidentTagQuery {
	itq.order = append(itq.order, o...)
	return itq
}

// QueryIncidents chains the current query on the "incidents" edge.
func (itq *IncidentTagQuery) QueryIncidents() *IncidentQuery {
	query := (&IncidentClient{config: itq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := itq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := itq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(incidenttag.Table, incidenttag.FieldID, selector),
			sqlgraph.To(incident.Table, incident.FieldID),
			sqlgraph.Edge(sqlgraph.M2M, true, incidenttag.IncidentsTable, incidenttag.IncidentsPrimaryKey...),
		)
		fromU = sqlgraph.SetNeighbors(itq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryDebriefQuestions chains the current query on the "debrief_questions" edge.
func (itq *IncidentTagQuery) QueryDebriefQuestions() *IncidentDebriefQuestionQuery {
	query := (&IncidentDebriefQuestionClient{config: itq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := itq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := itq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(incidenttag.Table, incidenttag.FieldID, selector),
			sqlgraph.To(incidentdebriefquestion.Table, incidentdebriefquestion.FieldID),
			sqlgraph.Edge(sqlgraph.M2M, true, incidenttag.DebriefQuestionsTable, incidenttag.DebriefQuestionsPrimaryKey...),
		)
		fromU = sqlgraph.SetNeighbors(itq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first IncidentTag entity from the query.
// Returns a *NotFoundError when no IncidentTag was found.
func (itq *IncidentTagQuery) First(ctx context.Context) (*IncidentTag, error) {
	nodes, err := itq.Limit(1).All(setContextOp(ctx, itq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{incidenttag.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (itq *IncidentTagQuery) FirstX(ctx context.Context) *IncidentTag {
	node, err := itq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first IncidentTag ID from the query.
// Returns a *NotFoundError when no IncidentTag ID was found.
func (itq *IncidentTagQuery) FirstID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = itq.Limit(1).IDs(setContextOp(ctx, itq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{incidenttag.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (itq *IncidentTagQuery) FirstIDX(ctx context.Context) uuid.UUID {
	id, err := itq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single IncidentTag entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one IncidentTag entity is found.
// Returns a *NotFoundError when no IncidentTag entities are found.
func (itq *IncidentTagQuery) Only(ctx context.Context) (*IncidentTag, error) {
	nodes, err := itq.Limit(2).All(setContextOp(ctx, itq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{incidenttag.Label}
	default:
		return nil, &NotSingularError{incidenttag.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (itq *IncidentTagQuery) OnlyX(ctx context.Context) *IncidentTag {
	node, err := itq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only IncidentTag ID in the query.
// Returns a *NotSingularError when more than one IncidentTag ID is found.
// Returns a *NotFoundError when no entities are found.
func (itq *IncidentTagQuery) OnlyID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = itq.Limit(2).IDs(setContextOp(ctx, itq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{incidenttag.Label}
	default:
		err = &NotSingularError{incidenttag.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (itq *IncidentTagQuery) OnlyIDX(ctx context.Context) uuid.UUID {
	id, err := itq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of IncidentTags.
func (itq *IncidentTagQuery) All(ctx context.Context) ([]*IncidentTag, error) {
	ctx = setContextOp(ctx, itq.ctx, ent.OpQueryAll)
	if err := itq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*IncidentTag, *IncidentTagQuery]()
	return withInterceptors[[]*IncidentTag](ctx, itq, qr, itq.inters)
}

// AllX is like All, but panics if an error occurs.
func (itq *IncidentTagQuery) AllX(ctx context.Context) []*IncidentTag {
	nodes, err := itq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of IncidentTag IDs.
func (itq *IncidentTagQuery) IDs(ctx context.Context) (ids []uuid.UUID, err error) {
	if itq.ctx.Unique == nil && itq.path != nil {
		itq.Unique(true)
	}
	ctx = setContextOp(ctx, itq.ctx, ent.OpQueryIDs)
	if err = itq.Select(incidenttag.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (itq *IncidentTagQuery) IDsX(ctx context.Context) []uuid.UUID {
	ids, err := itq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (itq *IncidentTagQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, itq.ctx, ent.OpQueryCount)
	if err := itq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, itq, querierCount[*IncidentTagQuery](), itq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (itq *IncidentTagQuery) CountX(ctx context.Context) int {
	count, err := itq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (itq *IncidentTagQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, itq.ctx, ent.OpQueryExist)
	switch _, err := itq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (itq *IncidentTagQuery) ExistX(ctx context.Context) bool {
	exist, err := itq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the IncidentTagQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (itq *IncidentTagQuery) Clone() *IncidentTagQuery {
	if itq == nil {
		return nil
	}
	return &IncidentTagQuery{
		config:               itq.config,
		ctx:                  itq.ctx.Clone(),
		order:                append([]incidenttag.OrderOption{}, itq.order...),
		inters:               append([]Interceptor{}, itq.inters...),
		predicates:           append([]predicate.IncidentTag{}, itq.predicates...),
		withIncidents:        itq.withIncidents.Clone(),
		withDebriefQuestions: itq.withDebriefQuestions.Clone(),
		// clone intermediate query.
		sql:       itq.sql.Clone(),
		path:      itq.path,
		modifiers: append([]func(*sql.Selector){}, itq.modifiers...),
	}
}

// WithIncidents tells the query-builder to eager-load the nodes that are connected to
// the "incidents" edge. The optional arguments are used to configure the query builder of the edge.
func (itq *IncidentTagQuery) WithIncidents(opts ...func(*IncidentQuery)) *IncidentTagQuery {
	query := (&IncidentClient{config: itq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	itq.withIncidents = query
	return itq
}

// WithDebriefQuestions tells the query-builder to eager-load the nodes that are connected to
// the "debrief_questions" edge. The optional arguments are used to configure the query builder of the edge.
func (itq *IncidentTagQuery) WithDebriefQuestions(opts ...func(*IncidentDebriefQuestionQuery)) *IncidentTagQuery {
	query := (&IncidentDebriefQuestionClient{config: itq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	itq.withDebriefQuestions = query
	return itq
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		ArchiveTime time.Time `json:"archive_time,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.IncidentTag.Query().
//		GroupBy(incidenttag.FieldArchiveTime).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (itq *IncidentTagQuery) GroupBy(field string, fields ...string) *IncidentTagGroupBy {
	itq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &IncidentTagGroupBy{build: itq}
	grbuild.flds = &itq.ctx.Fields
	grbuild.label = incidenttag.Label
	grbuild.scan = grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		ArchiveTime time.Time `json:"archive_time,omitempty"`
//	}
//
//	client.IncidentTag.Query().
//		Select(incidenttag.FieldArchiveTime).
//		Scan(ctx, &v)
func (itq *IncidentTagQuery) Select(fields ...string) *IncidentTagSelect {
	itq.ctx.Fields = append(itq.ctx.Fields, fields...)
	sbuild := &IncidentTagSelect{IncidentTagQuery: itq}
	sbuild.label = incidenttag.Label
	sbuild.flds, sbuild.scan = &itq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a IncidentTagSelect configured with the given aggregations.
func (itq *IncidentTagQuery) Aggregate(fns ...AggregateFunc) *IncidentTagSelect {
	return itq.Select().Aggregate(fns...)
}

func (itq *IncidentTagQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range itq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, itq); err != nil {
				return err
			}
		}
	}
	for _, f := range itq.ctx.Fields {
		if !incidenttag.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if itq.path != nil {
		prev, err := itq.path(ctx)
		if err != nil {
			return err
		}
		itq.sql = prev
	}
	return nil
}

func (itq *IncidentTagQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*IncidentTag, error) {
	var (
		nodes       = []*IncidentTag{}
		_spec       = itq.querySpec()
		loadedTypes = [2]bool{
			itq.withIncidents != nil,
			itq.withDebriefQuestions != nil,
		}
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*IncidentTag).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &IncidentTag{config: itq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	if len(itq.modifiers) > 0 {
		_spec.Modifiers = itq.modifiers
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, itq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := itq.withIncidents; query != nil {
		if err := itq.loadIncidents(ctx, query, nodes,
			func(n *IncidentTag) { n.Edges.Incidents = []*Incident{} },
			func(n *IncidentTag, e *Incident) { n.Edges.Incidents = append(n.Edges.Incidents, e) }); err != nil {
			return nil, err
		}
	}
	if query := itq.withDebriefQuestions; query != nil {
		if err := itq.loadDebriefQuestions(ctx, query, nodes,
			func(n *IncidentTag) { n.Edges.DebriefQuestions = []*IncidentDebriefQuestion{} },
			func(n *IncidentTag, e *IncidentDebriefQuestion) {
				n.Edges.DebriefQuestions = append(n.Edges.DebriefQuestions, e)
			}); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (itq *IncidentTagQuery) loadIncidents(ctx context.Context, query *IncidentQuery, nodes []*IncidentTag, init func(*IncidentTag), assign func(*IncidentTag, *Incident)) error {
	edgeIDs := make([]driver.Value, len(nodes))
	byID := make(map[uuid.UUID]*IncidentTag)
	nids := make(map[uuid.UUID]map[*IncidentTag]struct{})
	for i, node := range nodes {
		edgeIDs[i] = node.ID
		byID[node.ID] = node
		if init != nil {
			init(node)
		}
	}
	query.Where(func(s *sql.Selector) {
		joinT := sql.Table(incidenttag.IncidentsTable)
		s.Join(joinT).On(s.C(incident.FieldID), joinT.C(incidenttag.IncidentsPrimaryKey[0]))
		s.Where(sql.InValues(joinT.C(incidenttag.IncidentsPrimaryKey[1]), edgeIDs...))
		columns := s.SelectedColumns()
		s.Select(joinT.C(incidenttag.IncidentsPrimaryKey[1]))
		s.AppendSelect(columns...)
		s.SetDistinct(false)
	})
	if err := query.prepareQuery(ctx); err != nil {
		return err
	}
	qr := QuerierFunc(func(ctx context.Context, q Query) (Value, error) {
		return query.sqlAll(ctx, func(_ context.Context, spec *sqlgraph.QuerySpec) {
			assign := spec.Assign
			values := spec.ScanValues
			spec.ScanValues = func(columns []string) ([]any, error) {
				values, err := values(columns[1:])
				if err != nil {
					return nil, err
				}
				return append([]any{new(uuid.UUID)}, values...), nil
			}
			spec.Assign = func(columns []string, values []any) error {
				outValue := *values[0].(*uuid.UUID)
				inValue := *values[1].(*uuid.UUID)
				if nids[inValue] == nil {
					nids[inValue] = map[*IncidentTag]struct{}{byID[outValue]: {}}
					return assign(columns[1:], values[1:])
				}
				nids[inValue][byID[outValue]] = struct{}{}
				return nil
			}
		})
	})
	neighbors, err := withInterceptors[[]*Incident](ctx, query, qr, query.inters)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		nodes, ok := nids[n.ID]
		if !ok {
			return fmt.Errorf(`unexpected "incidents" node returned %v`, n.ID)
		}
		for kn := range nodes {
			assign(kn, n)
		}
	}
	return nil
}
func (itq *IncidentTagQuery) loadDebriefQuestions(ctx context.Context, query *IncidentDebriefQuestionQuery, nodes []*IncidentTag, init func(*IncidentTag), assign func(*IncidentTag, *IncidentDebriefQuestion)) error {
	edgeIDs := make([]driver.Value, len(nodes))
	byID := make(map[uuid.UUID]*IncidentTag)
	nids := make(map[uuid.UUID]map[*IncidentTag]struct{})
	for i, node := range nodes {
		edgeIDs[i] = node.ID
		byID[node.ID] = node
		if init != nil {
			init(node)
		}
	}
	query.Where(func(s *sql.Selector) {
		joinT := sql.Table(incidenttag.DebriefQuestionsTable)
		s.Join(joinT).On(s.C(incidentdebriefquestion.FieldID), joinT.C(incidenttag.DebriefQuestionsPrimaryKey[0]))
		s.Where(sql.InValues(joinT.C(incidenttag.DebriefQuestionsPrimaryKey[1]), edgeIDs...))
		columns := s.SelectedColumns()
		s.Select(joinT.C(incidenttag.DebriefQuestionsPrimaryKey[1]))
		s.AppendSelect(columns...)
		s.SetDistinct(false)
	})
	if err := query.prepareQuery(ctx); err != nil {
		return err
	}
	qr := QuerierFunc(func(ctx context.Context, q Query) (Value, error) {
		return query.sqlAll(ctx, func(_ context.Context, spec *sqlgraph.QuerySpec) {
			assign := spec.Assign
			values := spec.ScanValues
			spec.ScanValues = func(columns []string) ([]any, error) {
				values, err := values(columns[1:])
				if err != nil {
					return nil, err
				}
				return append([]any{new(uuid.UUID)}, values...), nil
			}
			spec.Assign = func(columns []string, values []any) error {
				outValue := *values[0].(*uuid.UUID)
				inValue := *values[1].(*uuid.UUID)
				if nids[inValue] == nil {
					nids[inValue] = map[*IncidentTag]struct{}{byID[outValue]: {}}
					return assign(columns[1:], values[1:])
				}
				nids[inValue][byID[outValue]] = struct{}{}
				return nil
			}
		})
	})
	neighbors, err := withInterceptors[[]*IncidentDebriefQuestion](ctx, query, qr, query.inters)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		nodes, ok := nids[n.ID]
		if !ok {
			return fmt.Errorf(`unexpected "debrief_questions" node returned %v`, n.ID)
		}
		for kn := range nodes {
			assign(kn, n)
		}
	}
	return nil
}

func (itq *IncidentTagQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := itq.querySpec()
	if len(itq.modifiers) > 0 {
		_spec.Modifiers = itq.modifiers
	}
	_spec.Node.Columns = itq.ctx.Fields
	if len(itq.ctx.Fields) > 0 {
		_spec.Unique = itq.ctx.Unique != nil && *itq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, itq.driver, _spec)
}

func (itq *IncidentTagQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(incidenttag.Table, incidenttag.Columns, sqlgraph.NewFieldSpec(incidenttag.FieldID, field.TypeUUID))
	_spec.From = itq.sql
	if unique := itq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if itq.path != nil {
		_spec.Unique = true
	}
	if fields := itq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, incidenttag.FieldID)
		for i := range fields {
			if fields[i] != incidenttag.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := itq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := itq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := itq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := itq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (itq *IncidentTagQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(itq.driver.Dialect())
	t1 := builder.Table(incidenttag.Table)
	columns := itq.ctx.Fields
	if len(columns) == 0 {
		columns = incidenttag.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if itq.sql != nil {
		selector = itq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if itq.ctx.Unique != nil && *itq.ctx.Unique {
		selector.Distinct()
	}
	for _, m := range itq.modifiers {
		m(selector)
	}
	for _, p := range itq.predicates {
		p(selector)
	}
	for _, p := range itq.order {
		p(selector)
	}
	if offset := itq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := itq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// Modify adds a query modifier for attaching custom logic to queries.
func (itq *IncidentTagQuery) Modify(modifiers ...func(s *sql.Selector)) *IncidentTagSelect {
	itq.modifiers = append(itq.modifiers, modifiers...)
	return itq.Select()
}

// IncidentTagGroupBy is the group-by builder for IncidentTag entities.
type IncidentTagGroupBy struct {
	selector
	build *IncidentTagQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (itgb *IncidentTagGroupBy) Aggregate(fns ...AggregateFunc) *IncidentTagGroupBy {
	itgb.fns = append(itgb.fns, fns...)
	return itgb
}

// Scan applies the selector query and scans the result into the given value.
func (itgb *IncidentTagGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, itgb.build.ctx, ent.OpQueryGroupBy)
	if err := itgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*IncidentTagQuery, *IncidentTagGroupBy](ctx, itgb.build, itgb, itgb.build.inters, v)
}

func (itgb *IncidentTagGroupBy) sqlScan(ctx context.Context, root *IncidentTagQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(itgb.fns))
	for _, fn := range itgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*itgb.flds)+len(itgb.fns))
		for _, f := range *itgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*itgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := itgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// IncidentTagSelect is the builder for selecting fields of IncidentTag entities.
type IncidentTagSelect struct {
	*IncidentTagQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (its *IncidentTagSelect) Aggregate(fns ...AggregateFunc) *IncidentTagSelect {
	its.fns = append(its.fns, fns...)
	return its
}

// Scan applies the selector query and scans the result into the given value.
func (its *IncidentTagSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, its.ctx, ent.OpQuerySelect)
	if err := its.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*IncidentTagQuery, *IncidentTagSelect](ctx, its.IncidentTagQuery, its, its.inters, v)
}

func (its *IncidentTagSelect) sqlScan(ctx context.Context, root *IncidentTagQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(its.fns))
	for _, fn := range its.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*its.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := its.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// Modify adds a query modifier for attaching custom logic to queries.
func (its *IncidentTagSelect) Modify(modifiers ...func(s *sql.Selector)) *IncidentTagSelect {
	its.modifiers = append(its.modifiers, modifiers...)
	return its
}
