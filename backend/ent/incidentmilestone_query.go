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
	"github.com/rezible/rezible/ent/incidentmilestone"
	"github.com/rezible/rezible/ent/predicate"
)

// IncidentMilestoneQuery is the builder for querying IncidentMilestone entities.
type IncidentMilestoneQuery struct {
	config
	ctx          *QueryContext
	order        []incidentmilestone.OrderOption
	inters       []Interceptor
	predicates   []predicate.IncidentMilestone
	withIncident *IncidentQuery
	modifiers    []func(*sql.Selector)
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the IncidentMilestoneQuery builder.
func (imq *IncidentMilestoneQuery) Where(ps ...predicate.IncidentMilestone) *IncidentMilestoneQuery {
	imq.predicates = append(imq.predicates, ps...)
	return imq
}

// Limit the number of records to be returned by this query.
func (imq *IncidentMilestoneQuery) Limit(limit int) *IncidentMilestoneQuery {
	imq.ctx.Limit = &limit
	return imq
}

// Offset to start from.
func (imq *IncidentMilestoneQuery) Offset(offset int) *IncidentMilestoneQuery {
	imq.ctx.Offset = &offset
	return imq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (imq *IncidentMilestoneQuery) Unique(unique bool) *IncidentMilestoneQuery {
	imq.ctx.Unique = &unique
	return imq
}

// Order specifies how the records should be ordered.
func (imq *IncidentMilestoneQuery) Order(o ...incidentmilestone.OrderOption) *IncidentMilestoneQuery {
	imq.order = append(imq.order, o...)
	return imq
}

// QueryIncident chains the current query on the "incident" edge.
func (imq *IncidentMilestoneQuery) QueryIncident() *IncidentQuery {
	query := (&IncidentClient{config: imq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := imq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := imq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(incidentmilestone.Table, incidentmilestone.FieldID, selector),
			sqlgraph.To(incident.Table, incident.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, incidentmilestone.IncidentTable, incidentmilestone.IncidentColumn),
		)
		fromU = sqlgraph.SetNeighbors(imq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first IncidentMilestone entity from the query.
// Returns a *NotFoundError when no IncidentMilestone was found.
func (imq *IncidentMilestoneQuery) First(ctx context.Context) (*IncidentMilestone, error) {
	nodes, err := imq.Limit(1).All(setContextOp(ctx, imq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{incidentmilestone.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (imq *IncidentMilestoneQuery) FirstX(ctx context.Context) *IncidentMilestone {
	node, err := imq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first IncidentMilestone ID from the query.
// Returns a *NotFoundError when no IncidentMilestone ID was found.
func (imq *IncidentMilestoneQuery) FirstID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = imq.Limit(1).IDs(setContextOp(ctx, imq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{incidentmilestone.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (imq *IncidentMilestoneQuery) FirstIDX(ctx context.Context) uuid.UUID {
	id, err := imq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single IncidentMilestone entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one IncidentMilestone entity is found.
// Returns a *NotFoundError when no IncidentMilestone entities are found.
func (imq *IncidentMilestoneQuery) Only(ctx context.Context) (*IncidentMilestone, error) {
	nodes, err := imq.Limit(2).All(setContextOp(ctx, imq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{incidentmilestone.Label}
	default:
		return nil, &NotSingularError{incidentmilestone.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (imq *IncidentMilestoneQuery) OnlyX(ctx context.Context) *IncidentMilestone {
	node, err := imq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only IncidentMilestone ID in the query.
// Returns a *NotSingularError when more than one IncidentMilestone ID is found.
// Returns a *NotFoundError when no entities are found.
func (imq *IncidentMilestoneQuery) OnlyID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = imq.Limit(2).IDs(setContextOp(ctx, imq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{incidentmilestone.Label}
	default:
		err = &NotSingularError{incidentmilestone.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (imq *IncidentMilestoneQuery) OnlyIDX(ctx context.Context) uuid.UUID {
	id, err := imq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of IncidentMilestones.
func (imq *IncidentMilestoneQuery) All(ctx context.Context) ([]*IncidentMilestone, error) {
	ctx = setContextOp(ctx, imq.ctx, ent.OpQueryAll)
	if err := imq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*IncidentMilestone, *IncidentMilestoneQuery]()
	return withInterceptors[[]*IncidentMilestone](ctx, imq, qr, imq.inters)
}

// AllX is like All, but panics if an error occurs.
func (imq *IncidentMilestoneQuery) AllX(ctx context.Context) []*IncidentMilestone {
	nodes, err := imq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of IncidentMilestone IDs.
func (imq *IncidentMilestoneQuery) IDs(ctx context.Context) (ids []uuid.UUID, err error) {
	if imq.ctx.Unique == nil && imq.path != nil {
		imq.Unique(true)
	}
	ctx = setContextOp(ctx, imq.ctx, ent.OpQueryIDs)
	if err = imq.Select(incidentmilestone.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (imq *IncidentMilestoneQuery) IDsX(ctx context.Context) []uuid.UUID {
	ids, err := imq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (imq *IncidentMilestoneQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, imq.ctx, ent.OpQueryCount)
	if err := imq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, imq, querierCount[*IncidentMilestoneQuery](), imq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (imq *IncidentMilestoneQuery) CountX(ctx context.Context) int {
	count, err := imq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (imq *IncidentMilestoneQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, imq.ctx, ent.OpQueryExist)
	switch _, err := imq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (imq *IncidentMilestoneQuery) ExistX(ctx context.Context) bool {
	exist, err := imq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the IncidentMilestoneQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (imq *IncidentMilestoneQuery) Clone() *IncidentMilestoneQuery {
	if imq == nil {
		return nil
	}
	return &IncidentMilestoneQuery{
		config:       imq.config,
		ctx:          imq.ctx.Clone(),
		order:        append([]incidentmilestone.OrderOption{}, imq.order...),
		inters:       append([]Interceptor{}, imq.inters...),
		predicates:   append([]predicate.IncidentMilestone{}, imq.predicates...),
		withIncident: imq.withIncident.Clone(),
		// clone intermediate query.
		sql:       imq.sql.Clone(),
		path:      imq.path,
		modifiers: append([]func(*sql.Selector){}, imq.modifiers...),
	}
}

// WithIncident tells the query-builder to eager-load the nodes that are connected to
// the "incident" edge. The optional arguments are used to configure the query builder of the edge.
func (imq *IncidentMilestoneQuery) WithIncident(opts ...func(*IncidentQuery)) *IncidentMilestoneQuery {
	query := (&IncidentClient{config: imq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	imq.withIncident = query
	return imq
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
//	client.IncidentMilestone.Query().
//		GroupBy(incidentmilestone.FieldIncidentID).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (imq *IncidentMilestoneQuery) GroupBy(field string, fields ...string) *IncidentMilestoneGroupBy {
	imq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &IncidentMilestoneGroupBy{build: imq}
	grbuild.flds = &imq.ctx.Fields
	grbuild.label = incidentmilestone.Label
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
//	client.IncidentMilestone.Query().
//		Select(incidentmilestone.FieldIncidentID).
//		Scan(ctx, &v)
func (imq *IncidentMilestoneQuery) Select(fields ...string) *IncidentMilestoneSelect {
	imq.ctx.Fields = append(imq.ctx.Fields, fields...)
	sbuild := &IncidentMilestoneSelect{IncidentMilestoneQuery: imq}
	sbuild.label = incidentmilestone.Label
	sbuild.flds, sbuild.scan = &imq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a IncidentMilestoneSelect configured with the given aggregations.
func (imq *IncidentMilestoneQuery) Aggregate(fns ...AggregateFunc) *IncidentMilestoneSelect {
	return imq.Select().Aggregate(fns...)
}

func (imq *IncidentMilestoneQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range imq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, imq); err != nil {
				return err
			}
		}
	}
	for _, f := range imq.ctx.Fields {
		if !incidentmilestone.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if imq.path != nil {
		prev, err := imq.path(ctx)
		if err != nil {
			return err
		}
		imq.sql = prev
	}
	return nil
}

func (imq *IncidentMilestoneQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*IncidentMilestone, error) {
	var (
		nodes       = []*IncidentMilestone{}
		_spec       = imq.querySpec()
		loadedTypes = [1]bool{
			imq.withIncident != nil,
		}
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*IncidentMilestone).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &IncidentMilestone{config: imq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	if len(imq.modifiers) > 0 {
		_spec.Modifiers = imq.modifiers
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, imq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := imq.withIncident; query != nil {
		if err := imq.loadIncident(ctx, query, nodes, nil,
			func(n *IncidentMilestone, e *Incident) { n.Edges.Incident = e }); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (imq *IncidentMilestoneQuery) loadIncident(ctx context.Context, query *IncidentQuery, nodes []*IncidentMilestone, init func(*IncidentMilestone), assign func(*IncidentMilestone, *Incident)) error {
	ids := make([]uuid.UUID, 0, len(nodes))
	nodeids := make(map[uuid.UUID][]*IncidentMilestone)
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

func (imq *IncidentMilestoneQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := imq.querySpec()
	if len(imq.modifiers) > 0 {
		_spec.Modifiers = imq.modifiers
	}
	_spec.Node.Columns = imq.ctx.Fields
	if len(imq.ctx.Fields) > 0 {
		_spec.Unique = imq.ctx.Unique != nil && *imq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, imq.driver, _spec)
}

func (imq *IncidentMilestoneQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(incidentmilestone.Table, incidentmilestone.Columns, sqlgraph.NewFieldSpec(incidentmilestone.FieldID, field.TypeUUID))
	_spec.From = imq.sql
	if unique := imq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if imq.path != nil {
		_spec.Unique = true
	}
	if fields := imq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, incidentmilestone.FieldID)
		for i := range fields {
			if fields[i] != incidentmilestone.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
		if imq.withIncident != nil {
			_spec.Node.AddColumnOnce(incidentmilestone.FieldIncidentID)
		}
	}
	if ps := imq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := imq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := imq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := imq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (imq *IncidentMilestoneQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(imq.driver.Dialect())
	t1 := builder.Table(incidentmilestone.Table)
	columns := imq.ctx.Fields
	if len(columns) == 0 {
		columns = incidentmilestone.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if imq.sql != nil {
		selector = imq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if imq.ctx.Unique != nil && *imq.ctx.Unique {
		selector.Distinct()
	}
	for _, m := range imq.modifiers {
		m(selector)
	}
	for _, p := range imq.predicates {
		p(selector)
	}
	for _, p := range imq.order {
		p(selector)
	}
	if offset := imq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := imq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// Modify adds a query modifier for attaching custom logic to queries.
func (imq *IncidentMilestoneQuery) Modify(modifiers ...func(s *sql.Selector)) *IncidentMilestoneSelect {
	imq.modifiers = append(imq.modifiers, modifiers...)
	return imq.Select()
}

// IncidentMilestoneGroupBy is the group-by builder for IncidentMilestone entities.
type IncidentMilestoneGroupBy struct {
	selector
	build *IncidentMilestoneQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (imgb *IncidentMilestoneGroupBy) Aggregate(fns ...AggregateFunc) *IncidentMilestoneGroupBy {
	imgb.fns = append(imgb.fns, fns...)
	return imgb
}

// Scan applies the selector query and scans the result into the given value.
func (imgb *IncidentMilestoneGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, imgb.build.ctx, ent.OpQueryGroupBy)
	if err := imgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*IncidentMilestoneQuery, *IncidentMilestoneGroupBy](ctx, imgb.build, imgb, imgb.build.inters, v)
}

func (imgb *IncidentMilestoneGroupBy) sqlScan(ctx context.Context, root *IncidentMilestoneQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(imgb.fns))
	for _, fn := range imgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*imgb.flds)+len(imgb.fns))
		for _, f := range *imgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*imgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := imgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// IncidentMilestoneSelect is the builder for selecting fields of IncidentMilestone entities.
type IncidentMilestoneSelect struct {
	*IncidentMilestoneQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (ims *IncidentMilestoneSelect) Aggregate(fns ...AggregateFunc) *IncidentMilestoneSelect {
	ims.fns = append(ims.fns, fns...)
	return ims
}

// Scan applies the selector query and scans the result into the given value.
func (ims *IncidentMilestoneSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, ims.ctx, ent.OpQuerySelect)
	if err := ims.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*IncidentMilestoneQuery, *IncidentMilestoneSelect](ctx, ims.IncidentMilestoneQuery, ims, ims.inters, v)
}

func (ims *IncidentMilestoneSelect) sqlScan(ctx context.Context, root *IncidentMilestoneQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(ims.fns))
	for _, fn := range ims.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*ims.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := ims.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// Modify adds a query modifier for attaching custom logic to queries.
func (ims *IncidentMilestoneSelect) Modify(modifiers ...func(s *sql.Selector)) *IncidentMilestoneSelect {
	ims.modifiers = append(ims.modifiers, modifiers...)
	return ims
}
