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
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/systemcomponent"
	"github.com/rezible/rezible/ent/systemcomponentcontrol"
	"github.com/rezible/rezible/ent/systemcomponentrelationshipcontrolaction"
)

// SystemComponentControlQuery is the builder for querying SystemComponentControl entities.
type SystemComponentControlQuery struct {
	config
	ctx                *QueryContext
	order              []systemcomponentcontrol.OrderOption
	inters             []Interceptor
	predicates         []predicate.SystemComponentControl
	withComponent      *SystemComponentQuery
	withControlActions *SystemComponentRelationshipControlActionQuery
	modifiers          []func(*sql.Selector)
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the SystemComponentControlQuery builder.
func (sccq *SystemComponentControlQuery) Where(ps ...predicate.SystemComponentControl) *SystemComponentControlQuery {
	sccq.predicates = append(sccq.predicates, ps...)
	return sccq
}

// Limit the number of records to be returned by this query.
func (sccq *SystemComponentControlQuery) Limit(limit int) *SystemComponentControlQuery {
	sccq.ctx.Limit = &limit
	return sccq
}

// Offset to start from.
func (sccq *SystemComponentControlQuery) Offset(offset int) *SystemComponentControlQuery {
	sccq.ctx.Offset = &offset
	return sccq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (sccq *SystemComponentControlQuery) Unique(unique bool) *SystemComponentControlQuery {
	sccq.ctx.Unique = &unique
	return sccq
}

// Order specifies how the records should be ordered.
func (sccq *SystemComponentControlQuery) Order(o ...systemcomponentcontrol.OrderOption) *SystemComponentControlQuery {
	sccq.order = append(sccq.order, o...)
	return sccq
}

// QueryComponent chains the current query on the "component" edge.
func (sccq *SystemComponentControlQuery) QueryComponent() *SystemComponentQuery {
	query := (&SystemComponentClient{config: sccq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := sccq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := sccq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(systemcomponentcontrol.Table, systemcomponentcontrol.FieldID, selector),
			sqlgraph.To(systemcomponent.Table, systemcomponent.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, systemcomponentcontrol.ComponentTable, systemcomponentcontrol.ComponentColumn),
		)
		fromU = sqlgraph.SetNeighbors(sccq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryControlActions chains the current query on the "control_actions" edge.
func (sccq *SystemComponentControlQuery) QueryControlActions() *SystemComponentRelationshipControlActionQuery {
	query := (&SystemComponentRelationshipControlActionClient{config: sccq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := sccq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := sccq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(systemcomponentcontrol.Table, systemcomponentcontrol.FieldID, selector),
			sqlgraph.To(systemcomponentrelationshipcontrolaction.Table, systemcomponentrelationshipcontrolaction.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, true, systemcomponentcontrol.ControlActionsTable, systemcomponentcontrol.ControlActionsColumn),
		)
		fromU = sqlgraph.SetNeighbors(sccq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first SystemComponentControl entity from the query.
// Returns a *NotFoundError when no SystemComponentControl was found.
func (sccq *SystemComponentControlQuery) First(ctx context.Context) (*SystemComponentControl, error) {
	nodes, err := sccq.Limit(1).All(setContextOp(ctx, sccq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{systemcomponentcontrol.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (sccq *SystemComponentControlQuery) FirstX(ctx context.Context) *SystemComponentControl {
	node, err := sccq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first SystemComponentControl ID from the query.
// Returns a *NotFoundError when no SystemComponentControl ID was found.
func (sccq *SystemComponentControlQuery) FirstID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = sccq.Limit(1).IDs(setContextOp(ctx, sccq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{systemcomponentcontrol.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (sccq *SystemComponentControlQuery) FirstIDX(ctx context.Context) uuid.UUID {
	id, err := sccq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single SystemComponentControl entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one SystemComponentControl entity is found.
// Returns a *NotFoundError when no SystemComponentControl entities are found.
func (sccq *SystemComponentControlQuery) Only(ctx context.Context) (*SystemComponentControl, error) {
	nodes, err := sccq.Limit(2).All(setContextOp(ctx, sccq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{systemcomponentcontrol.Label}
	default:
		return nil, &NotSingularError{systemcomponentcontrol.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (sccq *SystemComponentControlQuery) OnlyX(ctx context.Context) *SystemComponentControl {
	node, err := sccq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only SystemComponentControl ID in the query.
// Returns a *NotSingularError when more than one SystemComponentControl ID is found.
// Returns a *NotFoundError when no entities are found.
func (sccq *SystemComponentControlQuery) OnlyID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = sccq.Limit(2).IDs(setContextOp(ctx, sccq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{systemcomponentcontrol.Label}
	default:
		err = &NotSingularError{systemcomponentcontrol.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (sccq *SystemComponentControlQuery) OnlyIDX(ctx context.Context) uuid.UUID {
	id, err := sccq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of SystemComponentControls.
func (sccq *SystemComponentControlQuery) All(ctx context.Context) ([]*SystemComponentControl, error) {
	ctx = setContextOp(ctx, sccq.ctx, ent.OpQueryAll)
	if err := sccq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*SystemComponentControl, *SystemComponentControlQuery]()
	return withInterceptors[[]*SystemComponentControl](ctx, sccq, qr, sccq.inters)
}

// AllX is like All, but panics if an error occurs.
func (sccq *SystemComponentControlQuery) AllX(ctx context.Context) []*SystemComponentControl {
	nodes, err := sccq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of SystemComponentControl IDs.
func (sccq *SystemComponentControlQuery) IDs(ctx context.Context) (ids []uuid.UUID, err error) {
	if sccq.ctx.Unique == nil && sccq.path != nil {
		sccq.Unique(true)
	}
	ctx = setContextOp(ctx, sccq.ctx, ent.OpQueryIDs)
	if err = sccq.Select(systemcomponentcontrol.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (sccq *SystemComponentControlQuery) IDsX(ctx context.Context) []uuid.UUID {
	ids, err := sccq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (sccq *SystemComponentControlQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, sccq.ctx, ent.OpQueryCount)
	if err := sccq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, sccq, querierCount[*SystemComponentControlQuery](), sccq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (sccq *SystemComponentControlQuery) CountX(ctx context.Context) int {
	count, err := sccq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (sccq *SystemComponentControlQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, sccq.ctx, ent.OpQueryExist)
	switch _, err := sccq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (sccq *SystemComponentControlQuery) ExistX(ctx context.Context) bool {
	exist, err := sccq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the SystemComponentControlQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (sccq *SystemComponentControlQuery) Clone() *SystemComponentControlQuery {
	if sccq == nil {
		return nil
	}
	return &SystemComponentControlQuery{
		config:             sccq.config,
		ctx:                sccq.ctx.Clone(),
		order:              append([]systemcomponentcontrol.OrderOption{}, sccq.order...),
		inters:             append([]Interceptor{}, sccq.inters...),
		predicates:         append([]predicate.SystemComponentControl{}, sccq.predicates...),
		withComponent:      sccq.withComponent.Clone(),
		withControlActions: sccq.withControlActions.Clone(),
		// clone intermediate query.
		sql:       sccq.sql.Clone(),
		path:      sccq.path,
		modifiers: append([]func(*sql.Selector){}, sccq.modifiers...),
	}
}

// WithComponent tells the query-builder to eager-load the nodes that are connected to
// the "component" edge. The optional arguments are used to configure the query builder of the edge.
func (sccq *SystemComponentControlQuery) WithComponent(opts ...func(*SystemComponentQuery)) *SystemComponentControlQuery {
	query := (&SystemComponentClient{config: sccq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	sccq.withComponent = query
	return sccq
}

// WithControlActions tells the query-builder to eager-load the nodes that are connected to
// the "control_actions" edge. The optional arguments are used to configure the query builder of the edge.
func (sccq *SystemComponentControlQuery) WithControlActions(opts ...func(*SystemComponentRelationshipControlActionQuery)) *SystemComponentControlQuery {
	query := (&SystemComponentRelationshipControlActionClient{config: sccq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	sccq.withControlActions = query
	return sccq
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		ComponentID uuid.UUID `json:"component_id,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.SystemComponentControl.Query().
//		GroupBy(systemcomponentcontrol.FieldComponentID).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (sccq *SystemComponentControlQuery) GroupBy(field string, fields ...string) *SystemComponentControlGroupBy {
	sccq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &SystemComponentControlGroupBy{build: sccq}
	grbuild.flds = &sccq.ctx.Fields
	grbuild.label = systemcomponentcontrol.Label
	grbuild.scan = grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		ComponentID uuid.UUID `json:"component_id,omitempty"`
//	}
//
//	client.SystemComponentControl.Query().
//		Select(systemcomponentcontrol.FieldComponentID).
//		Scan(ctx, &v)
func (sccq *SystemComponentControlQuery) Select(fields ...string) *SystemComponentControlSelect {
	sccq.ctx.Fields = append(sccq.ctx.Fields, fields...)
	sbuild := &SystemComponentControlSelect{SystemComponentControlQuery: sccq}
	sbuild.label = systemcomponentcontrol.Label
	sbuild.flds, sbuild.scan = &sccq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a SystemComponentControlSelect configured with the given aggregations.
func (sccq *SystemComponentControlQuery) Aggregate(fns ...AggregateFunc) *SystemComponentControlSelect {
	return sccq.Select().Aggregate(fns...)
}

func (sccq *SystemComponentControlQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range sccq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, sccq); err != nil {
				return err
			}
		}
	}
	for _, f := range sccq.ctx.Fields {
		if !systemcomponentcontrol.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if sccq.path != nil {
		prev, err := sccq.path(ctx)
		if err != nil {
			return err
		}
		sccq.sql = prev
	}
	return nil
}

func (sccq *SystemComponentControlQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*SystemComponentControl, error) {
	var (
		nodes       = []*SystemComponentControl{}
		_spec       = sccq.querySpec()
		loadedTypes = [2]bool{
			sccq.withComponent != nil,
			sccq.withControlActions != nil,
		}
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*SystemComponentControl).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &SystemComponentControl{config: sccq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	if len(sccq.modifiers) > 0 {
		_spec.Modifiers = sccq.modifiers
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, sccq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := sccq.withComponent; query != nil {
		if err := sccq.loadComponent(ctx, query, nodes, nil,
			func(n *SystemComponentControl, e *SystemComponent) { n.Edges.Component = e }); err != nil {
			return nil, err
		}
	}
	if query := sccq.withControlActions; query != nil {
		if err := sccq.loadControlActions(ctx, query, nodes,
			func(n *SystemComponentControl) {
				n.Edges.ControlActions = []*SystemComponentRelationshipControlAction{}
			},
			func(n *SystemComponentControl, e *SystemComponentRelationshipControlAction) {
				n.Edges.ControlActions = append(n.Edges.ControlActions, e)
			}); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (sccq *SystemComponentControlQuery) loadComponent(ctx context.Context, query *SystemComponentQuery, nodes []*SystemComponentControl, init func(*SystemComponentControl), assign func(*SystemComponentControl, *SystemComponent)) error {
	ids := make([]uuid.UUID, 0, len(nodes))
	nodeids := make(map[uuid.UUID][]*SystemComponentControl)
	for i := range nodes {
		fk := nodes[i].ComponentID
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
			return fmt.Errorf(`unexpected foreign-key "component_id" returned %v`, n.ID)
		}
		for i := range nodes {
			assign(nodes[i], n)
		}
	}
	return nil
}
func (sccq *SystemComponentControlQuery) loadControlActions(ctx context.Context, query *SystemComponentRelationshipControlActionQuery, nodes []*SystemComponentControl, init func(*SystemComponentControl), assign func(*SystemComponentControl, *SystemComponentRelationshipControlAction)) error {
	fks := make([]driver.Value, 0, len(nodes))
	nodeids := make(map[uuid.UUID]*SystemComponentControl)
	for i := range nodes {
		fks = append(fks, nodes[i].ID)
		nodeids[nodes[i].ID] = nodes[i]
		if init != nil {
			init(nodes[i])
		}
	}
	if len(query.ctx.Fields) > 0 {
		query.ctx.AppendFieldOnce(systemcomponentrelationshipcontrolaction.FieldControlID)
	}
	query.Where(predicate.SystemComponentRelationshipControlAction(func(s *sql.Selector) {
		s.Where(sql.InValues(s.C(systemcomponentcontrol.ControlActionsColumn), fks...))
	}))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		fk := n.ControlID
		node, ok := nodeids[fk]
		if !ok {
			return fmt.Errorf(`unexpected referenced foreign-key "control_id" returned %v for node %v`, fk, n.ID)
		}
		assign(node, n)
	}
	return nil
}

func (sccq *SystemComponentControlQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := sccq.querySpec()
	if len(sccq.modifiers) > 0 {
		_spec.Modifiers = sccq.modifiers
	}
	_spec.Node.Columns = sccq.ctx.Fields
	if len(sccq.ctx.Fields) > 0 {
		_spec.Unique = sccq.ctx.Unique != nil && *sccq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, sccq.driver, _spec)
}

func (sccq *SystemComponentControlQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(systemcomponentcontrol.Table, systemcomponentcontrol.Columns, sqlgraph.NewFieldSpec(systemcomponentcontrol.FieldID, field.TypeUUID))
	_spec.From = sccq.sql
	if unique := sccq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if sccq.path != nil {
		_spec.Unique = true
	}
	if fields := sccq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, systemcomponentcontrol.FieldID)
		for i := range fields {
			if fields[i] != systemcomponentcontrol.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
		if sccq.withComponent != nil {
			_spec.Node.AddColumnOnce(systemcomponentcontrol.FieldComponentID)
		}
	}
	if ps := sccq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := sccq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := sccq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := sccq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (sccq *SystemComponentControlQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(sccq.driver.Dialect())
	t1 := builder.Table(systemcomponentcontrol.Table)
	columns := sccq.ctx.Fields
	if len(columns) == 0 {
		columns = systemcomponentcontrol.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if sccq.sql != nil {
		selector = sccq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if sccq.ctx.Unique != nil && *sccq.ctx.Unique {
		selector.Distinct()
	}
	for _, m := range sccq.modifiers {
		m(selector)
	}
	for _, p := range sccq.predicates {
		p(selector)
	}
	for _, p := range sccq.order {
		p(selector)
	}
	if offset := sccq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := sccq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// Modify adds a query modifier for attaching custom logic to queries.
func (sccq *SystemComponentControlQuery) Modify(modifiers ...func(s *sql.Selector)) *SystemComponentControlSelect {
	sccq.modifiers = append(sccq.modifiers, modifiers...)
	return sccq.Select()
}

// SystemComponentControlGroupBy is the group-by builder for SystemComponentControl entities.
type SystemComponentControlGroupBy struct {
	selector
	build *SystemComponentControlQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (sccgb *SystemComponentControlGroupBy) Aggregate(fns ...AggregateFunc) *SystemComponentControlGroupBy {
	sccgb.fns = append(sccgb.fns, fns...)
	return sccgb
}

// Scan applies the selector query and scans the result into the given value.
func (sccgb *SystemComponentControlGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, sccgb.build.ctx, ent.OpQueryGroupBy)
	if err := sccgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*SystemComponentControlQuery, *SystemComponentControlGroupBy](ctx, sccgb.build, sccgb, sccgb.build.inters, v)
}

func (sccgb *SystemComponentControlGroupBy) sqlScan(ctx context.Context, root *SystemComponentControlQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(sccgb.fns))
	for _, fn := range sccgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*sccgb.flds)+len(sccgb.fns))
		for _, f := range *sccgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*sccgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := sccgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// SystemComponentControlSelect is the builder for selecting fields of SystemComponentControl entities.
type SystemComponentControlSelect struct {
	*SystemComponentControlQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (sccs *SystemComponentControlSelect) Aggregate(fns ...AggregateFunc) *SystemComponentControlSelect {
	sccs.fns = append(sccs.fns, fns...)
	return sccs
}

// Scan applies the selector query and scans the result into the given value.
func (sccs *SystemComponentControlSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, sccs.ctx, ent.OpQuerySelect)
	if err := sccs.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*SystemComponentControlQuery, *SystemComponentControlSelect](ctx, sccs.SystemComponentControlQuery, sccs, sccs.inters, v)
}

func (sccs *SystemComponentControlSelect) sqlScan(ctx context.Context, root *SystemComponentControlQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(sccs.fns))
	for _, fn := range sccs.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*sccs.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := sccs.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// Modify adds a query modifier for attaching custom logic to queries.
func (sccs *SystemComponentControlSelect) Modify(modifiers ...func(s *sql.Selector)) *SystemComponentControlSelect {
	sccs.modifiers = append(sccs.modifiers, modifiers...)
	return sccs
}
