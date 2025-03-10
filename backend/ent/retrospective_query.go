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
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/retrospective"
	"github.com/rezible/rezible/ent/retrospectivediscussion"
	"github.com/rezible/rezible/ent/systemanalysis"
)

// RetrospectiveQuery is the builder for querying Retrospective entities.
type RetrospectiveQuery struct {
	config
	ctx                *QueryContext
	order              []retrospective.OrderOption
	inters             []Interceptor
	predicates         []predicate.Retrospective
	withIncident       *IncidentQuery
	withDiscussions    *RetrospectiveDiscussionQuery
	withSystemAnalysis *SystemAnalysisQuery
	modifiers          []func(*sql.Selector)
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the RetrospectiveQuery builder.
func (rq *RetrospectiveQuery) Where(ps ...predicate.Retrospective) *RetrospectiveQuery {
	rq.predicates = append(rq.predicates, ps...)
	return rq
}

// Limit the number of records to be returned by this query.
func (rq *RetrospectiveQuery) Limit(limit int) *RetrospectiveQuery {
	rq.ctx.Limit = &limit
	return rq
}

// Offset to start from.
func (rq *RetrospectiveQuery) Offset(offset int) *RetrospectiveQuery {
	rq.ctx.Offset = &offset
	return rq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (rq *RetrospectiveQuery) Unique(unique bool) *RetrospectiveQuery {
	rq.ctx.Unique = &unique
	return rq
}

// Order specifies how the records should be ordered.
func (rq *RetrospectiveQuery) Order(o ...retrospective.OrderOption) *RetrospectiveQuery {
	rq.order = append(rq.order, o...)
	return rq
}

// QueryIncident chains the current query on the "incident" edge.
func (rq *RetrospectiveQuery) QueryIncident() *IncidentQuery {
	query := (&IncidentClient{config: rq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := rq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := rq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(retrospective.Table, retrospective.FieldID, selector),
			sqlgraph.To(incident.Table, incident.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, retrospective.IncidentTable, retrospective.IncidentColumn),
		)
		fromU = sqlgraph.SetNeighbors(rq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryDiscussions chains the current query on the "discussions" edge.
func (rq *RetrospectiveQuery) QueryDiscussions() *RetrospectiveDiscussionQuery {
	query := (&RetrospectiveDiscussionClient{config: rq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := rq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := rq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(retrospective.Table, retrospective.FieldID, selector),
			sqlgraph.To(retrospectivediscussion.Table, retrospectivediscussion.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, true, retrospective.DiscussionsTable, retrospective.DiscussionsColumn),
		)
		fromU = sqlgraph.SetNeighbors(rq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QuerySystemAnalysis chains the current query on the "system_analysis" edge.
func (rq *RetrospectiveQuery) QuerySystemAnalysis() *SystemAnalysisQuery {
	query := (&SystemAnalysisClient{config: rq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := rq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := rq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(retrospective.Table, retrospective.FieldID, selector),
			sqlgraph.To(systemanalysis.Table, systemanalysis.FieldID),
			sqlgraph.Edge(sqlgraph.O2O, true, retrospective.SystemAnalysisTable, retrospective.SystemAnalysisColumn),
		)
		fromU = sqlgraph.SetNeighbors(rq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first Retrospective entity from the query.
// Returns a *NotFoundError when no Retrospective was found.
func (rq *RetrospectiveQuery) First(ctx context.Context) (*Retrospective, error) {
	nodes, err := rq.Limit(1).All(setContextOp(ctx, rq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{retrospective.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (rq *RetrospectiveQuery) FirstX(ctx context.Context) *Retrospective {
	node, err := rq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first Retrospective ID from the query.
// Returns a *NotFoundError when no Retrospective ID was found.
func (rq *RetrospectiveQuery) FirstID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = rq.Limit(1).IDs(setContextOp(ctx, rq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{retrospective.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (rq *RetrospectiveQuery) FirstIDX(ctx context.Context) uuid.UUID {
	id, err := rq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single Retrospective entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one Retrospective entity is found.
// Returns a *NotFoundError when no Retrospective entities are found.
func (rq *RetrospectiveQuery) Only(ctx context.Context) (*Retrospective, error) {
	nodes, err := rq.Limit(2).All(setContextOp(ctx, rq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{retrospective.Label}
	default:
		return nil, &NotSingularError{retrospective.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (rq *RetrospectiveQuery) OnlyX(ctx context.Context) *Retrospective {
	node, err := rq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only Retrospective ID in the query.
// Returns a *NotSingularError when more than one Retrospective ID is found.
// Returns a *NotFoundError when no entities are found.
func (rq *RetrospectiveQuery) OnlyID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = rq.Limit(2).IDs(setContextOp(ctx, rq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{retrospective.Label}
	default:
		err = &NotSingularError{retrospective.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (rq *RetrospectiveQuery) OnlyIDX(ctx context.Context) uuid.UUID {
	id, err := rq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of Retrospectives.
func (rq *RetrospectiveQuery) All(ctx context.Context) ([]*Retrospective, error) {
	ctx = setContextOp(ctx, rq.ctx, ent.OpQueryAll)
	if err := rq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*Retrospective, *RetrospectiveQuery]()
	return withInterceptors[[]*Retrospective](ctx, rq, qr, rq.inters)
}

// AllX is like All, but panics if an error occurs.
func (rq *RetrospectiveQuery) AllX(ctx context.Context) []*Retrospective {
	nodes, err := rq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of Retrospective IDs.
func (rq *RetrospectiveQuery) IDs(ctx context.Context) (ids []uuid.UUID, err error) {
	if rq.ctx.Unique == nil && rq.path != nil {
		rq.Unique(true)
	}
	ctx = setContextOp(ctx, rq.ctx, ent.OpQueryIDs)
	if err = rq.Select(retrospective.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (rq *RetrospectiveQuery) IDsX(ctx context.Context) []uuid.UUID {
	ids, err := rq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (rq *RetrospectiveQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, rq.ctx, ent.OpQueryCount)
	if err := rq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, rq, querierCount[*RetrospectiveQuery](), rq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (rq *RetrospectiveQuery) CountX(ctx context.Context) int {
	count, err := rq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (rq *RetrospectiveQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, rq.ctx, ent.OpQueryExist)
	switch _, err := rq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (rq *RetrospectiveQuery) ExistX(ctx context.Context) bool {
	exist, err := rq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the RetrospectiveQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (rq *RetrospectiveQuery) Clone() *RetrospectiveQuery {
	if rq == nil {
		return nil
	}
	return &RetrospectiveQuery{
		config:             rq.config,
		ctx:                rq.ctx.Clone(),
		order:              append([]retrospective.OrderOption{}, rq.order...),
		inters:             append([]Interceptor{}, rq.inters...),
		predicates:         append([]predicate.Retrospective{}, rq.predicates...),
		withIncident:       rq.withIncident.Clone(),
		withDiscussions:    rq.withDiscussions.Clone(),
		withSystemAnalysis: rq.withSystemAnalysis.Clone(),
		// clone intermediate query.
		sql:       rq.sql.Clone(),
		path:      rq.path,
		modifiers: append([]func(*sql.Selector){}, rq.modifiers...),
	}
}

// WithIncident tells the query-builder to eager-load the nodes that are connected to
// the "incident" edge. The optional arguments are used to configure the query builder of the edge.
func (rq *RetrospectiveQuery) WithIncident(opts ...func(*IncidentQuery)) *RetrospectiveQuery {
	query := (&IncidentClient{config: rq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	rq.withIncident = query
	return rq
}

// WithDiscussions tells the query-builder to eager-load the nodes that are connected to
// the "discussions" edge. The optional arguments are used to configure the query builder of the edge.
func (rq *RetrospectiveQuery) WithDiscussions(opts ...func(*RetrospectiveDiscussionQuery)) *RetrospectiveQuery {
	query := (&RetrospectiveDiscussionClient{config: rq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	rq.withDiscussions = query
	return rq
}

// WithSystemAnalysis tells the query-builder to eager-load the nodes that are connected to
// the "system_analysis" edge. The optional arguments are used to configure the query builder of the edge.
func (rq *RetrospectiveQuery) WithSystemAnalysis(opts ...func(*SystemAnalysisQuery)) *RetrospectiveQuery {
	query := (&SystemAnalysisClient{config: rq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	rq.withSystemAnalysis = query
	return rq
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		IncidentID uuid.UUID `json:"incident_id,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.Retrospective.Query().
//		GroupBy(retrospective.FieldIncidentID).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (rq *RetrospectiveQuery) GroupBy(field string, fields ...string) *RetrospectiveGroupBy {
	rq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &RetrospectiveGroupBy{build: rq}
	grbuild.flds = &rq.ctx.Fields
	grbuild.label = retrospective.Label
	grbuild.scan = grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		IncidentID uuid.UUID `json:"incident_id,omitempty"`
//	}
//
//	client.Retrospective.Query().
//		Select(retrospective.FieldIncidentID).
//		Scan(ctx, &v)
func (rq *RetrospectiveQuery) Select(fields ...string) *RetrospectiveSelect {
	rq.ctx.Fields = append(rq.ctx.Fields, fields...)
	sbuild := &RetrospectiveSelect{RetrospectiveQuery: rq}
	sbuild.label = retrospective.Label
	sbuild.flds, sbuild.scan = &rq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a RetrospectiveSelect configured with the given aggregations.
func (rq *RetrospectiveQuery) Aggregate(fns ...AggregateFunc) *RetrospectiveSelect {
	return rq.Select().Aggregate(fns...)
}

func (rq *RetrospectiveQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range rq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, rq); err != nil {
				return err
			}
		}
	}
	for _, f := range rq.ctx.Fields {
		if !retrospective.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if rq.path != nil {
		prev, err := rq.path(ctx)
		if err != nil {
			return err
		}
		rq.sql = prev
	}
	return nil
}

func (rq *RetrospectiveQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*Retrospective, error) {
	var (
		nodes       = []*Retrospective{}
		_spec       = rq.querySpec()
		loadedTypes = [3]bool{
			rq.withIncident != nil,
			rq.withDiscussions != nil,
			rq.withSystemAnalysis != nil,
		}
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*Retrospective).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &Retrospective{config: rq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	if len(rq.modifiers) > 0 {
		_spec.Modifiers = rq.modifiers
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, rq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := rq.withIncident; query != nil {
		if err := rq.loadIncident(ctx, query, nodes, nil,
			func(n *Retrospective, e *Incident) { n.Edges.Incident = e }); err != nil {
			return nil, err
		}
	}
	if query := rq.withDiscussions; query != nil {
		if err := rq.loadDiscussions(ctx, query, nodes,
			func(n *Retrospective) { n.Edges.Discussions = []*RetrospectiveDiscussion{} },
			func(n *Retrospective, e *RetrospectiveDiscussion) {
				n.Edges.Discussions = append(n.Edges.Discussions, e)
			}); err != nil {
			return nil, err
		}
	}
	if query := rq.withSystemAnalysis; query != nil {
		if err := rq.loadSystemAnalysis(ctx, query, nodes, nil,
			func(n *Retrospective, e *SystemAnalysis) { n.Edges.SystemAnalysis = e }); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (rq *RetrospectiveQuery) loadIncident(ctx context.Context, query *IncidentQuery, nodes []*Retrospective, init func(*Retrospective), assign func(*Retrospective, *Incident)) error {
	ids := make([]uuid.UUID, 0, len(nodes))
	nodeids := make(map[uuid.UUID][]*Retrospective)
	for i := range nodes {
		fk := nodes[i].IncidentID
		if _, ok := nodeids[fk]; !ok {
			ids = append(ids, fk)
		}
		nodeids[fk] = append(nodeids[fk], nodes[i])
	}
	if len(ids) == 0 {
		return nil
	}
	query.Where(incident.IDIn(ids...))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		nodes, ok := nodeids[n.ID]
		if !ok {
			return fmt.Errorf(`unexpected foreign-key "incident_id" returned %v`, n.ID)
		}
		for i := range nodes {
			assign(nodes[i], n)
		}
	}
	return nil
}
func (rq *RetrospectiveQuery) loadDiscussions(ctx context.Context, query *RetrospectiveDiscussionQuery, nodes []*Retrospective, init func(*Retrospective), assign func(*Retrospective, *RetrospectiveDiscussion)) error {
	fks := make([]driver.Value, 0, len(nodes))
	nodeids := make(map[uuid.UUID]*Retrospective)
	for i := range nodes {
		fks = append(fks, nodes[i].ID)
		nodeids[nodes[i].ID] = nodes[i]
		if init != nil {
			init(nodes[i])
		}
	}
	if len(query.ctx.Fields) > 0 {
		query.ctx.AppendFieldOnce(retrospectivediscussion.FieldRetrospectiveID)
	}
	query.Where(predicate.RetrospectiveDiscussion(func(s *sql.Selector) {
		s.Where(sql.InValues(s.C(retrospective.DiscussionsColumn), fks...))
	}))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		fk := n.RetrospectiveID
		node, ok := nodeids[fk]
		if !ok {
			return fmt.Errorf(`unexpected referenced foreign-key "retrospective_id" returned %v for node %v`, fk, n.ID)
		}
		assign(node, n)
	}
	return nil
}
func (rq *RetrospectiveQuery) loadSystemAnalysis(ctx context.Context, query *SystemAnalysisQuery, nodes []*Retrospective, init func(*Retrospective), assign func(*Retrospective, *SystemAnalysis)) error {
	ids := make([]uuid.UUID, 0, len(nodes))
	nodeids := make(map[uuid.UUID][]*Retrospective)
	for i := range nodes {
		fk := nodes[i].SystemAnalysisID
		if _, ok := nodeids[fk]; !ok {
			ids = append(ids, fk)
		}
		nodeids[fk] = append(nodeids[fk], nodes[i])
	}
	if len(ids) == 0 {
		return nil
	}
	query.Where(systemanalysis.IDIn(ids...))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		nodes, ok := nodeids[n.ID]
		if !ok {
			return fmt.Errorf(`unexpected foreign-key "system_analysis_id" returned %v`, n.ID)
		}
		for i := range nodes {
			assign(nodes[i], n)
		}
	}
	return nil
}

func (rq *RetrospectiveQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := rq.querySpec()
	if len(rq.modifiers) > 0 {
		_spec.Modifiers = rq.modifiers
	}
	_spec.Node.Columns = rq.ctx.Fields
	if len(rq.ctx.Fields) > 0 {
		_spec.Unique = rq.ctx.Unique != nil && *rq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, rq.driver, _spec)
}

func (rq *RetrospectiveQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(retrospective.Table, retrospective.Columns, sqlgraph.NewFieldSpec(retrospective.FieldID, field.TypeUUID))
	_spec.From = rq.sql
	if unique := rq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if rq.path != nil {
		_spec.Unique = true
	}
	if fields := rq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, retrospective.FieldID)
		for i := range fields {
			if fields[i] != retrospective.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
		if rq.withIncident != nil {
			_spec.Node.AddColumnOnce(retrospective.FieldIncidentID)
		}
		if rq.withSystemAnalysis != nil {
			_spec.Node.AddColumnOnce(retrospective.FieldSystemAnalysisID)
		}
	}
	if ps := rq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := rq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := rq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := rq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (rq *RetrospectiveQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(rq.driver.Dialect())
	t1 := builder.Table(retrospective.Table)
	columns := rq.ctx.Fields
	if len(columns) == 0 {
		columns = retrospective.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if rq.sql != nil {
		selector = rq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if rq.ctx.Unique != nil && *rq.ctx.Unique {
		selector.Distinct()
	}
	for _, m := range rq.modifiers {
		m(selector)
	}
	for _, p := range rq.predicates {
		p(selector)
	}
	for _, p := range rq.order {
		p(selector)
	}
	if offset := rq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := rq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// Modify adds a query modifier for attaching custom logic to queries.
func (rq *RetrospectiveQuery) Modify(modifiers ...func(s *sql.Selector)) *RetrospectiveSelect {
	rq.modifiers = append(rq.modifiers, modifiers...)
	return rq.Select()
}

// RetrospectiveGroupBy is the group-by builder for Retrospective entities.
type RetrospectiveGroupBy struct {
	selector
	build *RetrospectiveQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (rgb *RetrospectiveGroupBy) Aggregate(fns ...AggregateFunc) *RetrospectiveGroupBy {
	rgb.fns = append(rgb.fns, fns...)
	return rgb
}

// Scan applies the selector query and scans the result into the given value.
func (rgb *RetrospectiveGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, rgb.build.ctx, ent.OpQueryGroupBy)
	if err := rgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*RetrospectiveQuery, *RetrospectiveGroupBy](ctx, rgb.build, rgb, rgb.build.inters, v)
}

func (rgb *RetrospectiveGroupBy) sqlScan(ctx context.Context, root *RetrospectiveQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(rgb.fns))
	for _, fn := range rgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*rgb.flds)+len(rgb.fns))
		for _, f := range *rgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*rgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := rgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// RetrospectiveSelect is the builder for selecting fields of Retrospective entities.
type RetrospectiveSelect struct {
	*RetrospectiveQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (rs *RetrospectiveSelect) Aggregate(fns ...AggregateFunc) *RetrospectiveSelect {
	rs.fns = append(rs.fns, fns...)
	return rs
}

// Scan applies the selector query and scans the result into the given value.
func (rs *RetrospectiveSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, rs.ctx, ent.OpQuerySelect)
	if err := rs.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*RetrospectiveQuery, *RetrospectiveSelect](ctx, rs.RetrospectiveQuery, rs, rs.inters, v)
}

func (rs *RetrospectiveSelect) sqlScan(ctx context.Context, root *RetrospectiveQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(rs.fns))
	for _, fn := range rs.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*rs.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := rs.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// Modify adds a query modifier for attaching custom logic to queries.
func (rs *RetrospectiveSelect) Modify(modifiers ...func(s *sql.Selector)) *RetrospectiveSelect {
	rs.modifiers = append(rs.modifiers, modifiers...)
	return rs
}
