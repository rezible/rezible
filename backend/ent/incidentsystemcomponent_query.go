// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"fmt"
	"math"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/incidentsystemcomponent"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/systemcomponent"
)

// IncidentSystemComponentQuery is the builder for querying IncidentSystemComponent entities.
type IncidentSystemComponentQuery struct {
	config
	ctx                 *QueryContext
	order               []incidentsystemcomponent.OrderOption
	inters              []Interceptor
	predicates          []predicate.IncidentSystemComponent
	withIncident        *IncidentQuery
	withSystemComponent *SystemComponentQuery
	modifiers           []func(*sql.Selector)
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the IncidentSystemComponentQuery builder.
func (iscq *IncidentSystemComponentQuery) Where(ps ...predicate.IncidentSystemComponent) *IncidentSystemComponentQuery {
	iscq.predicates = append(iscq.predicates, ps...)
	return iscq
}

// Limit the number of records to be returned by this query.
func (iscq *IncidentSystemComponentQuery) Limit(limit int) *IncidentSystemComponentQuery {
	iscq.ctx.Limit = &limit
	return iscq
}

// Offset to start from.
func (iscq *IncidentSystemComponentQuery) Offset(offset int) *IncidentSystemComponentQuery {
	iscq.ctx.Offset = &offset
	return iscq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (iscq *IncidentSystemComponentQuery) Unique(unique bool) *IncidentSystemComponentQuery {
	iscq.ctx.Unique = &unique
	return iscq
}

// Order specifies how the records should be ordered.
func (iscq *IncidentSystemComponentQuery) Order(o ...incidentsystemcomponent.OrderOption) *IncidentSystemComponentQuery {
	iscq.order = append(iscq.order, o...)
	return iscq
}

// QueryIncident chains the current query on the "incident" edge.
func (iscq *IncidentSystemComponentQuery) QueryIncident() *IncidentQuery {
	query := (&IncidentClient{config: iscq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := iscq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := iscq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(incidentsystemcomponent.Table, incidentsystemcomponent.FieldID, selector),
			sqlgraph.To(incident.Table, incident.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, incidentsystemcomponent.IncidentTable, incidentsystemcomponent.IncidentColumn),
		)
		fromU = sqlgraph.SetNeighbors(iscq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QuerySystemComponent chains the current query on the "system_component" edge.
func (iscq *IncidentSystemComponentQuery) QuerySystemComponent() *SystemComponentQuery {
	query := (&SystemComponentClient{config: iscq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := iscq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := iscq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(incidentsystemcomponent.Table, incidentsystemcomponent.FieldID, selector),
			sqlgraph.To(systemcomponent.Table, systemcomponent.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, incidentsystemcomponent.SystemComponentTable, incidentsystemcomponent.SystemComponentColumn),
		)
		fromU = sqlgraph.SetNeighbors(iscq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first IncidentSystemComponent entity from the query.
// Returns a *NotFoundError when no IncidentSystemComponent was found.
func (iscq *IncidentSystemComponentQuery) First(ctx context.Context) (*IncidentSystemComponent, error) {
	nodes, err := iscq.Limit(1).All(setContextOp(ctx, iscq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{incidentsystemcomponent.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (iscq *IncidentSystemComponentQuery) FirstX(ctx context.Context) *IncidentSystemComponent {
	node, err := iscq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first IncidentSystemComponent ID from the query.
// Returns a *NotFoundError when no IncidentSystemComponent ID was found.
func (iscq *IncidentSystemComponentQuery) FirstID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = iscq.Limit(1).IDs(setContextOp(ctx, iscq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{incidentsystemcomponent.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (iscq *IncidentSystemComponentQuery) FirstIDX(ctx context.Context) uuid.UUID {
	id, err := iscq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single IncidentSystemComponent entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one IncidentSystemComponent entity is found.
// Returns a *NotFoundError when no IncidentSystemComponent entities are found.
func (iscq *IncidentSystemComponentQuery) Only(ctx context.Context) (*IncidentSystemComponent, error) {
	nodes, err := iscq.Limit(2).All(setContextOp(ctx, iscq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{incidentsystemcomponent.Label}
	default:
		return nil, &NotSingularError{incidentsystemcomponent.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (iscq *IncidentSystemComponentQuery) OnlyX(ctx context.Context) *IncidentSystemComponent {
	node, err := iscq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only IncidentSystemComponent ID in the query.
// Returns a *NotSingularError when more than one IncidentSystemComponent ID is found.
// Returns a *NotFoundError when no entities are found.
func (iscq *IncidentSystemComponentQuery) OnlyID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = iscq.Limit(2).IDs(setContextOp(ctx, iscq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{incidentsystemcomponent.Label}
	default:
		err = &NotSingularError{incidentsystemcomponent.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (iscq *IncidentSystemComponentQuery) OnlyIDX(ctx context.Context) uuid.UUID {
	id, err := iscq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of IncidentSystemComponents.
func (iscq *IncidentSystemComponentQuery) All(ctx context.Context) ([]*IncidentSystemComponent, error) {
	ctx = setContextOp(ctx, iscq.ctx, ent.OpQueryAll)
	if err := iscq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*IncidentSystemComponent, *IncidentSystemComponentQuery]()
	return withInterceptors[[]*IncidentSystemComponent](ctx, iscq, qr, iscq.inters)
}

// AllX is like All, but panics if an error occurs.
func (iscq *IncidentSystemComponentQuery) AllX(ctx context.Context) []*IncidentSystemComponent {
	nodes, err := iscq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of IncidentSystemComponent IDs.
func (iscq *IncidentSystemComponentQuery) IDs(ctx context.Context) (ids []uuid.UUID, err error) {
	if iscq.ctx.Unique == nil && iscq.path != nil {
		iscq.Unique(true)
	}
	ctx = setContextOp(ctx, iscq.ctx, ent.OpQueryIDs)
	if err = iscq.Select(incidentsystemcomponent.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (iscq *IncidentSystemComponentQuery) IDsX(ctx context.Context) []uuid.UUID {
	ids, err := iscq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (iscq *IncidentSystemComponentQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, iscq.ctx, ent.OpQueryCount)
	if err := iscq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, iscq, querierCount[*IncidentSystemComponentQuery](), iscq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (iscq *IncidentSystemComponentQuery) CountX(ctx context.Context) int {
	count, err := iscq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (iscq *IncidentSystemComponentQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, iscq.ctx, ent.OpQueryExist)
	switch _, err := iscq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (iscq *IncidentSystemComponentQuery) ExistX(ctx context.Context) bool {
	exist, err := iscq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the IncidentSystemComponentQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (iscq *IncidentSystemComponentQuery) Clone() *IncidentSystemComponentQuery {
	if iscq == nil {
		return nil
	}
	return &IncidentSystemComponentQuery{
		config:              iscq.config,
		ctx:                 iscq.ctx.Clone(),
		order:               append([]incidentsystemcomponent.OrderOption{}, iscq.order...),
		inters:              append([]Interceptor{}, iscq.inters...),
		predicates:          append([]predicate.IncidentSystemComponent{}, iscq.predicates...),
		withIncident:        iscq.withIncident.Clone(),
		withSystemComponent: iscq.withSystemComponent.Clone(),
		// clone intermediate query.
		sql:       iscq.sql.Clone(),
		path:      iscq.path,
		modifiers: append([]func(*sql.Selector){}, iscq.modifiers...),
	}
}

// WithIncident tells the query-builder to eager-load the nodes that are connected to
// the "incident" edge. The optional arguments are used to configure the query builder of the edge.
func (iscq *IncidentSystemComponentQuery) WithIncident(opts ...func(*IncidentQuery)) *IncidentSystemComponentQuery {
	query := (&IncidentClient{config: iscq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	iscq.withIncident = query
	return iscq
}

// WithSystemComponent tells the query-builder to eager-load the nodes that are connected to
// the "system_component" edge. The optional arguments are used to configure the query builder of the edge.
func (iscq *IncidentSystemComponentQuery) WithSystemComponent(opts ...func(*SystemComponentQuery)) *IncidentSystemComponentQuery {
	query := (&SystemComponentClient{config: iscq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	iscq.withSystemComponent = query
	return iscq
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
//	client.IncidentSystemComponent.Query().
//		GroupBy(incidentsystemcomponent.FieldIncidentID).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (iscq *IncidentSystemComponentQuery) GroupBy(field string, fields ...string) *IncidentSystemComponentGroupBy {
	iscq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &IncidentSystemComponentGroupBy{build: iscq}
	grbuild.flds = &iscq.ctx.Fields
	grbuild.label = incidentsystemcomponent.Label
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
//	client.IncidentSystemComponent.Query().
//		Select(incidentsystemcomponent.FieldIncidentID).
//		Scan(ctx, &v)
func (iscq *IncidentSystemComponentQuery) Select(fields ...string) *IncidentSystemComponentSelect {
	iscq.ctx.Fields = append(iscq.ctx.Fields, fields...)
	sbuild := &IncidentSystemComponentSelect{IncidentSystemComponentQuery: iscq}
	sbuild.label = incidentsystemcomponent.Label
	sbuild.flds, sbuild.scan = &iscq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a IncidentSystemComponentSelect configured with the given aggregations.
func (iscq *IncidentSystemComponentQuery) Aggregate(fns ...AggregateFunc) *IncidentSystemComponentSelect {
	return iscq.Select().Aggregate(fns...)
}

func (iscq *IncidentSystemComponentQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range iscq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, iscq); err != nil {
				return err
			}
		}
	}
	for _, f := range iscq.ctx.Fields {
		if !incidentsystemcomponent.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if iscq.path != nil {
		prev, err := iscq.path(ctx)
		if err != nil {
			return err
		}
		iscq.sql = prev
	}
	return nil
}

func (iscq *IncidentSystemComponentQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*IncidentSystemComponent, error) {
	var (
		nodes       = []*IncidentSystemComponent{}
		_spec       = iscq.querySpec()
		loadedTypes = [2]bool{
			iscq.withIncident != nil,
			iscq.withSystemComponent != nil,
		}
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*IncidentSystemComponent).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &IncidentSystemComponent{config: iscq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	if len(iscq.modifiers) > 0 {
		_spec.Modifiers = iscq.modifiers
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, iscq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := iscq.withIncident; query != nil {
		if err := iscq.loadIncident(ctx, query, nodes, nil,
			func(n *IncidentSystemComponent, e *Incident) { n.Edges.Incident = e }); err != nil {
			return nil, err
		}
	}
	if query := iscq.withSystemComponent; query != nil {
		if err := iscq.loadSystemComponent(ctx, query, nodes, nil,
			func(n *IncidentSystemComponent, e *SystemComponent) { n.Edges.SystemComponent = e }); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (iscq *IncidentSystemComponentQuery) loadIncident(ctx context.Context, query *IncidentQuery, nodes []*IncidentSystemComponent, init func(*IncidentSystemComponent), assign func(*IncidentSystemComponent, *Incident)) error {
	ids := make([]uuid.UUID, 0, len(nodes))
	nodeids := make(map[uuid.UUID][]*IncidentSystemComponent)
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
func (iscq *IncidentSystemComponentQuery) loadSystemComponent(ctx context.Context, query *SystemComponentQuery, nodes []*IncidentSystemComponent, init func(*IncidentSystemComponent), assign func(*IncidentSystemComponent, *SystemComponent)) error {
	ids := make([]uuid.UUID, 0, len(nodes))
	nodeids := make(map[uuid.UUID][]*IncidentSystemComponent)
	for i := range nodes {
		fk := nodes[i].SystemComponentID
		if _, ok := nodeids[fk]; !ok {
			ids = append(ids, fk)
		}
		nodeids[fk] = append(nodeids[fk], nodes[i])
	}
	if len(ids) == 0 {
		return nil
	}
	query.Where(systemcomponent.IDIn(ids...))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		nodes, ok := nodeids[n.ID]
		if !ok {
			return fmt.Errorf(`unexpected foreign-key "system_component_id" returned %v`, n.ID)
		}
		for i := range nodes {
			assign(nodes[i], n)
		}
	}
	return nil
}

func (iscq *IncidentSystemComponentQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := iscq.querySpec()
	if len(iscq.modifiers) > 0 {
		_spec.Modifiers = iscq.modifiers
	}
	_spec.Node.Columns = iscq.ctx.Fields
	if len(iscq.ctx.Fields) > 0 {
		_spec.Unique = iscq.ctx.Unique != nil && *iscq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, iscq.driver, _spec)
}

func (iscq *IncidentSystemComponentQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(incidentsystemcomponent.Table, incidentsystemcomponent.Columns, sqlgraph.NewFieldSpec(incidentsystemcomponent.FieldID, field.TypeUUID))
	_spec.From = iscq.sql
	if unique := iscq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if iscq.path != nil {
		_spec.Unique = true
	}
	if fields := iscq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, incidentsystemcomponent.FieldID)
		for i := range fields {
			if fields[i] != incidentsystemcomponent.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
		if iscq.withIncident != nil {
			_spec.Node.AddColumnOnce(incidentsystemcomponent.FieldIncidentID)
		}
		if iscq.withSystemComponent != nil {
			_spec.Node.AddColumnOnce(incidentsystemcomponent.FieldSystemComponentID)
		}
	}
	if ps := iscq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := iscq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := iscq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := iscq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (iscq *IncidentSystemComponentQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(iscq.driver.Dialect())
	t1 := builder.Table(incidentsystemcomponent.Table)
	columns := iscq.ctx.Fields
	if len(columns) == 0 {
		columns = incidentsystemcomponent.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if iscq.sql != nil {
		selector = iscq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if iscq.ctx.Unique != nil && *iscq.ctx.Unique {
		selector.Distinct()
	}
	for _, m := range iscq.modifiers {
		m(selector)
	}
	for _, p := range iscq.predicates {
		p(selector)
	}
	for _, p := range iscq.order {
		p(selector)
	}
	if offset := iscq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := iscq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// Modify adds a query modifier for attaching custom logic to queries.
func (iscq *IncidentSystemComponentQuery) Modify(modifiers ...func(s *sql.Selector)) *IncidentSystemComponentSelect {
	iscq.modifiers = append(iscq.modifiers, modifiers...)
	return iscq.Select()
}

// IncidentSystemComponentGroupBy is the group-by builder for IncidentSystemComponent entities.
type IncidentSystemComponentGroupBy struct {
	selector
	build *IncidentSystemComponentQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (iscgb *IncidentSystemComponentGroupBy) Aggregate(fns ...AggregateFunc) *IncidentSystemComponentGroupBy {
	iscgb.fns = append(iscgb.fns, fns...)
	return iscgb
}

// Scan applies the selector query and scans the result into the given value.
func (iscgb *IncidentSystemComponentGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, iscgb.build.ctx, ent.OpQueryGroupBy)
	if err := iscgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*IncidentSystemComponentQuery, *IncidentSystemComponentGroupBy](ctx, iscgb.build, iscgb, iscgb.build.inters, v)
}

func (iscgb *IncidentSystemComponentGroupBy) sqlScan(ctx context.Context, root *IncidentSystemComponentQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(iscgb.fns))
	for _, fn := range iscgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*iscgb.flds)+len(iscgb.fns))
		for _, f := range *iscgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*iscgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := iscgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// IncidentSystemComponentSelect is the builder for selecting fields of IncidentSystemComponent entities.
type IncidentSystemComponentSelect struct {
	*IncidentSystemComponentQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (iscs *IncidentSystemComponentSelect) Aggregate(fns ...AggregateFunc) *IncidentSystemComponentSelect {
	iscs.fns = append(iscs.fns, fns...)
	return iscs
}

// Scan applies the selector query and scans the result into the given value.
func (iscs *IncidentSystemComponentSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, iscs.ctx, ent.OpQuerySelect)
	if err := iscs.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*IncidentSystemComponentQuery, *IncidentSystemComponentSelect](ctx, iscs.IncidentSystemComponentQuery, iscs, iscs.inters, v)
}

func (iscs *IncidentSystemComponentSelect) sqlScan(ctx context.Context, root *IncidentSystemComponentQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(iscs.fns))
	for _, fn := range iscs.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*iscs.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := iscs.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// Modify adds a query modifier for attaching custom logic to queries.
func (iscs *IncidentSystemComponentSelect) Modify(modifiers ...func(s *sql.Selector)) *IncidentSystemComponentSelect {
	iscs.modifiers = append(iscs.modifiers, modifiers...)
	return iscs
}
