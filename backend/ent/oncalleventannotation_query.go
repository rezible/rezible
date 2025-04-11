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
	"github.com/rezible/rezible/ent/oncalleventannotation"
	"github.com/rezible/rezible/ent/oncallusershift"
	"github.com/rezible/rezible/ent/predicate"
)

// OncallEventAnnotationQuery is the builder for querying OncallEventAnnotation entities.
type OncallEventAnnotationQuery struct {
	config
	ctx        *QueryContext
	order      []oncalleventannotation.OrderOption
	inters     []Interceptor
	predicates []predicate.OncallEventAnnotation
	withShifts *OncallUserShiftQuery
	modifiers  []func(*sql.Selector)
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the OncallEventAnnotationQuery builder.
func (oeaq *OncallEventAnnotationQuery) Where(ps ...predicate.OncallEventAnnotation) *OncallEventAnnotationQuery {
	oeaq.predicates = append(oeaq.predicates, ps...)
	return oeaq
}

// Limit the number of records to be returned by this query.
func (oeaq *OncallEventAnnotationQuery) Limit(limit int) *OncallEventAnnotationQuery {
	oeaq.ctx.Limit = &limit
	return oeaq
}

// Offset to start from.
func (oeaq *OncallEventAnnotationQuery) Offset(offset int) *OncallEventAnnotationQuery {
	oeaq.ctx.Offset = &offset
	return oeaq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (oeaq *OncallEventAnnotationQuery) Unique(unique bool) *OncallEventAnnotationQuery {
	oeaq.ctx.Unique = &unique
	return oeaq
}

// Order specifies how the records should be ordered.
func (oeaq *OncallEventAnnotationQuery) Order(o ...oncalleventannotation.OrderOption) *OncallEventAnnotationQuery {
	oeaq.order = append(oeaq.order, o...)
	return oeaq
}

// QueryShifts chains the current query on the "shifts" edge.
func (oeaq *OncallEventAnnotationQuery) QueryShifts() *OncallUserShiftQuery {
	query := (&OncallUserShiftClient{config: oeaq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := oeaq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := oeaq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(oncalleventannotation.Table, oncalleventannotation.FieldID, selector),
			sqlgraph.To(oncallusershift.Table, oncallusershift.FieldID),
			sqlgraph.Edge(sqlgraph.M2M, true, oncalleventannotation.ShiftsTable, oncalleventannotation.ShiftsPrimaryKey...),
		)
		fromU = sqlgraph.SetNeighbors(oeaq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first OncallEventAnnotation entity from the query.
// Returns a *NotFoundError when no OncallEventAnnotation was found.
func (oeaq *OncallEventAnnotationQuery) First(ctx context.Context) (*OncallEventAnnotation, error) {
	nodes, err := oeaq.Limit(1).All(setContextOp(ctx, oeaq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{oncalleventannotation.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (oeaq *OncallEventAnnotationQuery) FirstX(ctx context.Context) *OncallEventAnnotation {
	node, err := oeaq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first OncallEventAnnotation ID from the query.
// Returns a *NotFoundError when no OncallEventAnnotation ID was found.
func (oeaq *OncallEventAnnotationQuery) FirstID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = oeaq.Limit(1).IDs(setContextOp(ctx, oeaq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{oncalleventannotation.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (oeaq *OncallEventAnnotationQuery) FirstIDX(ctx context.Context) uuid.UUID {
	id, err := oeaq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single OncallEventAnnotation entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one OncallEventAnnotation entity is found.
// Returns a *NotFoundError when no OncallEventAnnotation entities are found.
func (oeaq *OncallEventAnnotationQuery) Only(ctx context.Context) (*OncallEventAnnotation, error) {
	nodes, err := oeaq.Limit(2).All(setContextOp(ctx, oeaq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{oncalleventannotation.Label}
	default:
		return nil, &NotSingularError{oncalleventannotation.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (oeaq *OncallEventAnnotationQuery) OnlyX(ctx context.Context) *OncallEventAnnotation {
	node, err := oeaq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only OncallEventAnnotation ID in the query.
// Returns a *NotSingularError when more than one OncallEventAnnotation ID is found.
// Returns a *NotFoundError when no entities are found.
func (oeaq *OncallEventAnnotationQuery) OnlyID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = oeaq.Limit(2).IDs(setContextOp(ctx, oeaq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{oncalleventannotation.Label}
	default:
		err = &NotSingularError{oncalleventannotation.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (oeaq *OncallEventAnnotationQuery) OnlyIDX(ctx context.Context) uuid.UUID {
	id, err := oeaq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of OncallEventAnnotations.
func (oeaq *OncallEventAnnotationQuery) All(ctx context.Context) ([]*OncallEventAnnotation, error) {
	ctx = setContextOp(ctx, oeaq.ctx, ent.OpQueryAll)
	if err := oeaq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*OncallEventAnnotation, *OncallEventAnnotationQuery]()
	return withInterceptors[[]*OncallEventAnnotation](ctx, oeaq, qr, oeaq.inters)
}

// AllX is like All, but panics if an error occurs.
func (oeaq *OncallEventAnnotationQuery) AllX(ctx context.Context) []*OncallEventAnnotation {
	nodes, err := oeaq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of OncallEventAnnotation IDs.
func (oeaq *OncallEventAnnotationQuery) IDs(ctx context.Context) (ids []uuid.UUID, err error) {
	if oeaq.ctx.Unique == nil && oeaq.path != nil {
		oeaq.Unique(true)
	}
	ctx = setContextOp(ctx, oeaq.ctx, ent.OpQueryIDs)
	if err = oeaq.Select(oncalleventannotation.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (oeaq *OncallEventAnnotationQuery) IDsX(ctx context.Context) []uuid.UUID {
	ids, err := oeaq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (oeaq *OncallEventAnnotationQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, oeaq.ctx, ent.OpQueryCount)
	if err := oeaq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, oeaq, querierCount[*OncallEventAnnotationQuery](), oeaq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (oeaq *OncallEventAnnotationQuery) CountX(ctx context.Context) int {
	count, err := oeaq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (oeaq *OncallEventAnnotationQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, oeaq.ctx, ent.OpQueryExist)
	switch _, err := oeaq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (oeaq *OncallEventAnnotationQuery) ExistX(ctx context.Context) bool {
	exist, err := oeaq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the OncallEventAnnotationQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (oeaq *OncallEventAnnotationQuery) Clone() *OncallEventAnnotationQuery {
	if oeaq == nil {
		return nil
	}
	return &OncallEventAnnotationQuery{
		config:     oeaq.config,
		ctx:        oeaq.ctx.Clone(),
		order:      append([]oncalleventannotation.OrderOption{}, oeaq.order...),
		inters:     append([]Interceptor{}, oeaq.inters...),
		predicates: append([]predicate.OncallEventAnnotation{}, oeaq.predicates...),
		withShifts: oeaq.withShifts.Clone(),
		// clone intermediate query.
		sql:       oeaq.sql.Clone(),
		path:      oeaq.path,
		modifiers: append([]func(*sql.Selector){}, oeaq.modifiers...),
	}
}

// WithShifts tells the query-builder to eager-load the nodes that are connected to
// the "shifts" edge. The optional arguments are used to configure the query builder of the edge.
func (oeaq *OncallEventAnnotationQuery) WithShifts(opts ...func(*OncallUserShiftQuery)) *OncallEventAnnotationQuery {
	query := (&OncallUserShiftClient{config: oeaq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	oeaq.withShifts = query
	return oeaq
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		EventID string `json:"event_id,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.OncallEventAnnotation.Query().
//		GroupBy(oncalleventannotation.FieldEventID).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (oeaq *OncallEventAnnotationQuery) GroupBy(field string, fields ...string) *OncallEventAnnotationGroupBy {
	oeaq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &OncallEventAnnotationGroupBy{build: oeaq}
	grbuild.flds = &oeaq.ctx.Fields
	grbuild.label = oncalleventannotation.Label
	grbuild.scan = grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		EventID string `json:"event_id,omitempty"`
//	}
//
//	client.OncallEventAnnotation.Query().
//		Select(oncalleventannotation.FieldEventID).
//		Scan(ctx, &v)
func (oeaq *OncallEventAnnotationQuery) Select(fields ...string) *OncallEventAnnotationSelect {
	oeaq.ctx.Fields = append(oeaq.ctx.Fields, fields...)
	sbuild := &OncallEventAnnotationSelect{OncallEventAnnotationQuery: oeaq}
	sbuild.label = oncalleventannotation.Label
	sbuild.flds, sbuild.scan = &oeaq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a OncallEventAnnotationSelect configured with the given aggregations.
func (oeaq *OncallEventAnnotationQuery) Aggregate(fns ...AggregateFunc) *OncallEventAnnotationSelect {
	return oeaq.Select().Aggregate(fns...)
}

func (oeaq *OncallEventAnnotationQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range oeaq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, oeaq); err != nil {
				return err
			}
		}
	}
	for _, f := range oeaq.ctx.Fields {
		if !oncalleventannotation.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if oeaq.path != nil {
		prev, err := oeaq.path(ctx)
		if err != nil {
			return err
		}
		oeaq.sql = prev
	}
	return nil
}

func (oeaq *OncallEventAnnotationQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*OncallEventAnnotation, error) {
	var (
		nodes       = []*OncallEventAnnotation{}
		_spec       = oeaq.querySpec()
		loadedTypes = [1]bool{
			oeaq.withShifts != nil,
		}
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*OncallEventAnnotation).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &OncallEventAnnotation{config: oeaq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	if len(oeaq.modifiers) > 0 {
		_spec.Modifiers = oeaq.modifiers
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, oeaq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := oeaq.withShifts; query != nil {
		if err := oeaq.loadShifts(ctx, query, nodes,
			func(n *OncallEventAnnotation) { n.Edges.Shifts = []*OncallUserShift{} },
			func(n *OncallEventAnnotation, e *OncallUserShift) { n.Edges.Shifts = append(n.Edges.Shifts, e) }); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (oeaq *OncallEventAnnotationQuery) loadShifts(ctx context.Context, query *OncallUserShiftQuery, nodes []*OncallEventAnnotation, init func(*OncallEventAnnotation), assign func(*OncallEventAnnotation, *OncallUserShift)) error {
	edgeIDs := make([]driver.Value, len(nodes))
	byID := make(map[uuid.UUID]*OncallEventAnnotation)
	nids := make(map[uuid.UUID]map[*OncallEventAnnotation]struct{})
	for i, node := range nodes {
		edgeIDs[i] = node.ID
		byID[node.ID] = node
		if init != nil {
			init(node)
		}
	}
	query.Where(func(s *sql.Selector) {
		joinT := sql.Table(oncalleventannotation.ShiftsTable)
		s.Join(joinT).On(s.C(oncallusershift.FieldID), joinT.C(oncalleventannotation.ShiftsPrimaryKey[0]))
		s.Where(sql.InValues(joinT.C(oncalleventannotation.ShiftsPrimaryKey[1]), edgeIDs...))
		columns := s.SelectedColumns()
		s.Select(joinT.C(oncalleventannotation.ShiftsPrimaryKey[1]))
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
					nids[inValue] = map[*OncallEventAnnotation]struct{}{byID[outValue]: {}}
					return assign(columns[1:], values[1:])
				}
				nids[inValue][byID[outValue]] = struct{}{}
				return nil
			}
		})
	})
	neighbors, err := withInterceptors[[]*OncallUserShift](ctx, query, qr, query.inters)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		nodes, ok := nids[n.ID]
		if !ok {
			return fmt.Errorf(`unexpected "shifts" node returned %v`, n.ID)
		}
		for kn := range nodes {
			assign(kn, n)
		}
	}
	return nil
}

func (oeaq *OncallEventAnnotationQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := oeaq.querySpec()
	if len(oeaq.modifiers) > 0 {
		_spec.Modifiers = oeaq.modifiers
	}
	_spec.Node.Columns = oeaq.ctx.Fields
	if len(oeaq.ctx.Fields) > 0 {
		_spec.Unique = oeaq.ctx.Unique != nil && *oeaq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, oeaq.driver, _spec)
}

func (oeaq *OncallEventAnnotationQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(oncalleventannotation.Table, oncalleventannotation.Columns, sqlgraph.NewFieldSpec(oncalleventannotation.FieldID, field.TypeUUID))
	_spec.From = oeaq.sql
	if unique := oeaq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if oeaq.path != nil {
		_spec.Unique = true
	}
	if fields := oeaq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, oncalleventannotation.FieldID)
		for i := range fields {
			if fields[i] != oncalleventannotation.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := oeaq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := oeaq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := oeaq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := oeaq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (oeaq *OncallEventAnnotationQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(oeaq.driver.Dialect())
	t1 := builder.Table(oncalleventannotation.Table)
	columns := oeaq.ctx.Fields
	if len(columns) == 0 {
		columns = oncalleventannotation.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if oeaq.sql != nil {
		selector = oeaq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if oeaq.ctx.Unique != nil && *oeaq.ctx.Unique {
		selector.Distinct()
	}
	for _, m := range oeaq.modifiers {
		m(selector)
	}
	for _, p := range oeaq.predicates {
		p(selector)
	}
	for _, p := range oeaq.order {
		p(selector)
	}
	if offset := oeaq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := oeaq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// Modify adds a query modifier for attaching custom logic to queries.
func (oeaq *OncallEventAnnotationQuery) Modify(modifiers ...func(s *sql.Selector)) *OncallEventAnnotationSelect {
	oeaq.modifiers = append(oeaq.modifiers, modifiers...)
	return oeaq.Select()
}

// OncallEventAnnotationGroupBy is the group-by builder for OncallEventAnnotation entities.
type OncallEventAnnotationGroupBy struct {
	selector
	build *OncallEventAnnotationQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (oeagb *OncallEventAnnotationGroupBy) Aggregate(fns ...AggregateFunc) *OncallEventAnnotationGroupBy {
	oeagb.fns = append(oeagb.fns, fns...)
	return oeagb
}

// Scan applies the selector query and scans the result into the given value.
func (oeagb *OncallEventAnnotationGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, oeagb.build.ctx, ent.OpQueryGroupBy)
	if err := oeagb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*OncallEventAnnotationQuery, *OncallEventAnnotationGroupBy](ctx, oeagb.build, oeagb, oeagb.build.inters, v)
}

func (oeagb *OncallEventAnnotationGroupBy) sqlScan(ctx context.Context, root *OncallEventAnnotationQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(oeagb.fns))
	for _, fn := range oeagb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*oeagb.flds)+len(oeagb.fns))
		for _, f := range *oeagb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*oeagb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := oeagb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// OncallEventAnnotationSelect is the builder for selecting fields of OncallEventAnnotation entities.
type OncallEventAnnotationSelect struct {
	*OncallEventAnnotationQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (oeas *OncallEventAnnotationSelect) Aggregate(fns ...AggregateFunc) *OncallEventAnnotationSelect {
	oeas.fns = append(oeas.fns, fns...)
	return oeas
}

// Scan applies the selector query and scans the result into the given value.
func (oeas *OncallEventAnnotationSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, oeas.ctx, ent.OpQuerySelect)
	if err := oeas.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*OncallEventAnnotationQuery, *OncallEventAnnotationSelect](ctx, oeas.OncallEventAnnotationQuery, oeas, oeas.inters, v)
}

func (oeas *OncallEventAnnotationSelect) sqlScan(ctx context.Context, root *OncallEventAnnotationQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(oeas.fns))
	for _, fn := range oeas.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*oeas.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := oeas.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// Modify adds a query modifier for attaching custom logic to queries.
func (oeas *OncallEventAnnotationSelect) Modify(modifiers ...func(s *sql.Selector)) *OncallEventAnnotationSelect {
	oeas.modifiers = append(oeas.modifiers, modifiers...)
	return oeas
}
