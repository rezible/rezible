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
	"github.com/twohundreds/rezible/ent/incidentfield"
	"github.com/twohundreds/rezible/ent/incidentfieldoption"
	"github.com/twohundreds/rezible/ent/predicate"
)

// IncidentFieldOptionQuery is the builder for querying IncidentFieldOption entities.
type IncidentFieldOptionQuery struct {
	config
	ctx               *QueryContext
	order             []incidentfieldoption.OrderOption
	inters            []Interceptor
	predicates        []predicate.IncidentFieldOption
	withIncidentField *IncidentFieldQuery
	withIncidents     *IncidentQuery
	modifiers         []func(*sql.Selector)
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the IncidentFieldOptionQuery builder.
func (ifoq *IncidentFieldOptionQuery) Where(ps ...predicate.IncidentFieldOption) *IncidentFieldOptionQuery {
	ifoq.predicates = append(ifoq.predicates, ps...)
	return ifoq
}

// Limit the number of records to be returned by this query.
func (ifoq *IncidentFieldOptionQuery) Limit(limit int) *IncidentFieldOptionQuery {
	ifoq.ctx.Limit = &limit
	return ifoq
}

// Offset to start from.
func (ifoq *IncidentFieldOptionQuery) Offset(offset int) *IncidentFieldOptionQuery {
	ifoq.ctx.Offset = &offset
	return ifoq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (ifoq *IncidentFieldOptionQuery) Unique(unique bool) *IncidentFieldOptionQuery {
	ifoq.ctx.Unique = &unique
	return ifoq
}

// Order specifies how the records should be ordered.
func (ifoq *IncidentFieldOptionQuery) Order(o ...incidentfieldoption.OrderOption) *IncidentFieldOptionQuery {
	ifoq.order = append(ifoq.order, o...)
	return ifoq
}

// QueryIncidentField chains the current query on the "incident_field" edge.
func (ifoq *IncidentFieldOptionQuery) QueryIncidentField() *IncidentFieldQuery {
	query := (&IncidentFieldClient{config: ifoq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := ifoq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := ifoq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(incidentfieldoption.Table, incidentfieldoption.FieldID, selector),
			sqlgraph.To(incidentfield.Table, incidentfield.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, incidentfieldoption.IncidentFieldTable, incidentfieldoption.IncidentFieldColumn),
		)
		fromU = sqlgraph.SetNeighbors(ifoq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryIncidents chains the current query on the "incidents" edge.
func (ifoq *IncidentFieldOptionQuery) QueryIncidents() *IncidentQuery {
	query := (&IncidentClient{config: ifoq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := ifoq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := ifoq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(incidentfieldoption.Table, incidentfieldoption.FieldID, selector),
			sqlgraph.To(incident.Table, incident.FieldID),
			sqlgraph.Edge(sqlgraph.M2M, true, incidentfieldoption.IncidentsTable, incidentfieldoption.IncidentsPrimaryKey...),
		)
		fromU = sqlgraph.SetNeighbors(ifoq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first IncidentFieldOption entity from the query.
// Returns a *NotFoundError when no IncidentFieldOption was found.
func (ifoq *IncidentFieldOptionQuery) First(ctx context.Context) (*IncidentFieldOption, error) {
	nodes, err := ifoq.Limit(1).All(setContextOp(ctx, ifoq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{incidentfieldoption.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (ifoq *IncidentFieldOptionQuery) FirstX(ctx context.Context) *IncidentFieldOption {
	node, err := ifoq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first IncidentFieldOption ID from the query.
// Returns a *NotFoundError when no IncidentFieldOption ID was found.
func (ifoq *IncidentFieldOptionQuery) FirstID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = ifoq.Limit(1).IDs(setContextOp(ctx, ifoq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{incidentfieldoption.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (ifoq *IncidentFieldOptionQuery) FirstIDX(ctx context.Context) uuid.UUID {
	id, err := ifoq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single IncidentFieldOption entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one IncidentFieldOption entity is found.
// Returns a *NotFoundError when no IncidentFieldOption entities are found.
func (ifoq *IncidentFieldOptionQuery) Only(ctx context.Context) (*IncidentFieldOption, error) {
	nodes, err := ifoq.Limit(2).All(setContextOp(ctx, ifoq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{incidentfieldoption.Label}
	default:
		return nil, &NotSingularError{incidentfieldoption.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (ifoq *IncidentFieldOptionQuery) OnlyX(ctx context.Context) *IncidentFieldOption {
	node, err := ifoq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only IncidentFieldOption ID in the query.
// Returns a *NotSingularError when more than one IncidentFieldOption ID is found.
// Returns a *NotFoundError when no entities are found.
func (ifoq *IncidentFieldOptionQuery) OnlyID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = ifoq.Limit(2).IDs(setContextOp(ctx, ifoq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{incidentfieldoption.Label}
	default:
		err = &NotSingularError{incidentfieldoption.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (ifoq *IncidentFieldOptionQuery) OnlyIDX(ctx context.Context) uuid.UUID {
	id, err := ifoq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of IncidentFieldOptions.
func (ifoq *IncidentFieldOptionQuery) All(ctx context.Context) ([]*IncidentFieldOption, error) {
	ctx = setContextOp(ctx, ifoq.ctx, ent.OpQueryAll)
	if err := ifoq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*IncidentFieldOption, *IncidentFieldOptionQuery]()
	return withInterceptors[[]*IncidentFieldOption](ctx, ifoq, qr, ifoq.inters)
}

// AllX is like All, but panics if an error occurs.
func (ifoq *IncidentFieldOptionQuery) AllX(ctx context.Context) []*IncidentFieldOption {
	nodes, err := ifoq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of IncidentFieldOption IDs.
func (ifoq *IncidentFieldOptionQuery) IDs(ctx context.Context) (ids []uuid.UUID, err error) {
	if ifoq.ctx.Unique == nil && ifoq.path != nil {
		ifoq.Unique(true)
	}
	ctx = setContextOp(ctx, ifoq.ctx, ent.OpQueryIDs)
	if err = ifoq.Select(incidentfieldoption.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (ifoq *IncidentFieldOptionQuery) IDsX(ctx context.Context) []uuid.UUID {
	ids, err := ifoq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (ifoq *IncidentFieldOptionQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, ifoq.ctx, ent.OpQueryCount)
	if err := ifoq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, ifoq, querierCount[*IncidentFieldOptionQuery](), ifoq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (ifoq *IncidentFieldOptionQuery) CountX(ctx context.Context) int {
	count, err := ifoq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (ifoq *IncidentFieldOptionQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, ifoq.ctx, ent.OpQueryExist)
	switch _, err := ifoq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (ifoq *IncidentFieldOptionQuery) ExistX(ctx context.Context) bool {
	exist, err := ifoq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the IncidentFieldOptionQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (ifoq *IncidentFieldOptionQuery) Clone() *IncidentFieldOptionQuery {
	if ifoq == nil {
		return nil
	}
	return &IncidentFieldOptionQuery{
		config:            ifoq.config,
		ctx:               ifoq.ctx.Clone(),
		order:             append([]incidentfieldoption.OrderOption{}, ifoq.order...),
		inters:            append([]Interceptor{}, ifoq.inters...),
		predicates:        append([]predicate.IncidentFieldOption{}, ifoq.predicates...),
		withIncidentField: ifoq.withIncidentField.Clone(),
		withIncidents:     ifoq.withIncidents.Clone(),
		// clone intermediate query.
		sql:       ifoq.sql.Clone(),
		path:      ifoq.path,
		modifiers: append([]func(*sql.Selector){}, ifoq.modifiers...),
	}
}

// WithIncidentField tells the query-builder to eager-load the nodes that are connected to
// the "incident_field" edge. The optional arguments are used to configure the query builder of the edge.
func (ifoq *IncidentFieldOptionQuery) WithIncidentField(opts ...func(*IncidentFieldQuery)) *IncidentFieldOptionQuery {
	query := (&IncidentFieldClient{config: ifoq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	ifoq.withIncidentField = query
	return ifoq
}

// WithIncidents tells the query-builder to eager-load the nodes that are connected to
// the "incidents" edge. The optional arguments are used to configure the query builder of the edge.
func (ifoq *IncidentFieldOptionQuery) WithIncidents(opts ...func(*IncidentQuery)) *IncidentFieldOptionQuery {
	query := (&IncidentClient{config: ifoq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	ifoq.withIncidents = query
	return ifoq
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
//	client.IncidentFieldOption.Query().
//		GroupBy(incidentfieldoption.FieldArchiveTime).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (ifoq *IncidentFieldOptionQuery) GroupBy(field string, fields ...string) *IncidentFieldOptionGroupBy {
	ifoq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &IncidentFieldOptionGroupBy{build: ifoq}
	grbuild.flds = &ifoq.ctx.Fields
	grbuild.label = incidentfieldoption.Label
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
//	client.IncidentFieldOption.Query().
//		Select(incidentfieldoption.FieldArchiveTime).
//		Scan(ctx, &v)
func (ifoq *IncidentFieldOptionQuery) Select(fields ...string) *IncidentFieldOptionSelect {
	ifoq.ctx.Fields = append(ifoq.ctx.Fields, fields...)
	sbuild := &IncidentFieldOptionSelect{IncidentFieldOptionQuery: ifoq}
	sbuild.label = incidentfieldoption.Label
	sbuild.flds, sbuild.scan = &ifoq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a IncidentFieldOptionSelect configured with the given aggregations.
func (ifoq *IncidentFieldOptionQuery) Aggregate(fns ...AggregateFunc) *IncidentFieldOptionSelect {
	return ifoq.Select().Aggregate(fns...)
}

func (ifoq *IncidentFieldOptionQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range ifoq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, ifoq); err != nil {
				return err
			}
		}
	}
	for _, f := range ifoq.ctx.Fields {
		if !incidentfieldoption.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if ifoq.path != nil {
		prev, err := ifoq.path(ctx)
		if err != nil {
			return err
		}
		ifoq.sql = prev
	}
	return nil
}

func (ifoq *IncidentFieldOptionQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*IncidentFieldOption, error) {
	var (
		nodes       = []*IncidentFieldOption{}
		_spec       = ifoq.querySpec()
		loadedTypes = [2]bool{
			ifoq.withIncidentField != nil,
			ifoq.withIncidents != nil,
		}
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*IncidentFieldOption).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &IncidentFieldOption{config: ifoq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	if len(ifoq.modifiers) > 0 {
		_spec.Modifiers = ifoq.modifiers
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, ifoq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := ifoq.withIncidentField; query != nil {
		if err := ifoq.loadIncidentField(ctx, query, nodes, nil,
			func(n *IncidentFieldOption, e *IncidentField) { n.Edges.IncidentField = e }); err != nil {
			return nil, err
		}
	}
	if query := ifoq.withIncidents; query != nil {
		if err := ifoq.loadIncidents(ctx, query, nodes,
			func(n *IncidentFieldOption) { n.Edges.Incidents = []*Incident{} },
			func(n *IncidentFieldOption, e *Incident) { n.Edges.Incidents = append(n.Edges.Incidents, e) }); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (ifoq *IncidentFieldOptionQuery) loadIncidentField(ctx context.Context, query *IncidentFieldQuery, nodes []*IncidentFieldOption, init func(*IncidentFieldOption), assign func(*IncidentFieldOption, *IncidentField)) error {
	ids := make([]uuid.UUID, 0, len(nodes))
	nodeids := make(map[uuid.UUID][]*IncidentFieldOption)
	for i := range nodes {
		fk := nodes[i].IncidentFieldID
		if _, ok := nodeids[fk]; !ok {
			ids = append(ids, fk)
		}
		nodeids[fk] = append(nodeids[fk], nodes[i])
	}
	if len(ids) == 0 {
		return nil
	}
	query.Where(incidentfield.IDIn(ids...))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		nodes, ok := nodeids[n.ID]
		if !ok {
			return fmt.Errorf(`unexpected foreign-key "incident_field_id" returned %v`, n.ID)
		}
		for i := range nodes {
			assign(nodes[i], n)
		}
	}
	return nil
}
func (ifoq *IncidentFieldOptionQuery) loadIncidents(ctx context.Context, query *IncidentQuery, nodes []*IncidentFieldOption, init func(*IncidentFieldOption), assign func(*IncidentFieldOption, *Incident)) error {
	edgeIDs := make([]driver.Value, len(nodes))
	byID := make(map[uuid.UUID]*IncidentFieldOption)
	nids := make(map[uuid.UUID]map[*IncidentFieldOption]struct{})
	for i, node := range nodes {
		edgeIDs[i] = node.ID
		byID[node.ID] = node
		if init != nil {
			init(node)
		}
	}
	query.Where(func(s *sql.Selector) {
		joinT := sql.Table(incidentfieldoption.IncidentsTable)
		s.Join(joinT).On(s.C(incident.FieldID), joinT.C(incidentfieldoption.IncidentsPrimaryKey[0]))
		s.Where(sql.InValues(joinT.C(incidentfieldoption.IncidentsPrimaryKey[1]), edgeIDs...))
		columns := s.SelectedColumns()
		s.Select(joinT.C(incidentfieldoption.IncidentsPrimaryKey[1]))
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
					nids[inValue] = map[*IncidentFieldOption]struct{}{byID[outValue]: {}}
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

func (ifoq *IncidentFieldOptionQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := ifoq.querySpec()
	if len(ifoq.modifiers) > 0 {
		_spec.Modifiers = ifoq.modifiers
	}
	_spec.Node.Columns = ifoq.ctx.Fields
	if len(ifoq.ctx.Fields) > 0 {
		_spec.Unique = ifoq.ctx.Unique != nil && *ifoq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, ifoq.driver, _spec)
}

func (ifoq *IncidentFieldOptionQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(incidentfieldoption.Table, incidentfieldoption.Columns, sqlgraph.NewFieldSpec(incidentfieldoption.FieldID, field.TypeUUID))
	_spec.From = ifoq.sql
	if unique := ifoq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if ifoq.path != nil {
		_spec.Unique = true
	}
	if fields := ifoq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, incidentfieldoption.FieldID)
		for i := range fields {
			if fields[i] != incidentfieldoption.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
		if ifoq.withIncidentField != nil {
			_spec.Node.AddColumnOnce(incidentfieldoption.FieldIncidentFieldID)
		}
	}
	if ps := ifoq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := ifoq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := ifoq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := ifoq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (ifoq *IncidentFieldOptionQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(ifoq.driver.Dialect())
	t1 := builder.Table(incidentfieldoption.Table)
	columns := ifoq.ctx.Fields
	if len(columns) == 0 {
		columns = incidentfieldoption.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if ifoq.sql != nil {
		selector = ifoq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if ifoq.ctx.Unique != nil && *ifoq.ctx.Unique {
		selector.Distinct()
	}
	for _, m := range ifoq.modifiers {
		m(selector)
	}
	for _, p := range ifoq.predicates {
		p(selector)
	}
	for _, p := range ifoq.order {
		p(selector)
	}
	if offset := ifoq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := ifoq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// Modify adds a query modifier for attaching custom logic to queries.
func (ifoq *IncidentFieldOptionQuery) Modify(modifiers ...func(s *sql.Selector)) *IncidentFieldOptionSelect {
	ifoq.modifiers = append(ifoq.modifiers, modifiers...)
	return ifoq.Select()
}

// IncidentFieldOptionGroupBy is the group-by builder for IncidentFieldOption entities.
type IncidentFieldOptionGroupBy struct {
	selector
	build *IncidentFieldOptionQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (ifogb *IncidentFieldOptionGroupBy) Aggregate(fns ...AggregateFunc) *IncidentFieldOptionGroupBy {
	ifogb.fns = append(ifogb.fns, fns...)
	return ifogb
}

// Scan applies the selector query and scans the result into the given value.
func (ifogb *IncidentFieldOptionGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, ifogb.build.ctx, ent.OpQueryGroupBy)
	if err := ifogb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*IncidentFieldOptionQuery, *IncidentFieldOptionGroupBy](ctx, ifogb.build, ifogb, ifogb.build.inters, v)
}

func (ifogb *IncidentFieldOptionGroupBy) sqlScan(ctx context.Context, root *IncidentFieldOptionQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(ifogb.fns))
	for _, fn := range ifogb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*ifogb.flds)+len(ifogb.fns))
		for _, f := range *ifogb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*ifogb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := ifogb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// IncidentFieldOptionSelect is the builder for selecting fields of IncidentFieldOption entities.
type IncidentFieldOptionSelect struct {
	*IncidentFieldOptionQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (ifos *IncidentFieldOptionSelect) Aggregate(fns ...AggregateFunc) *IncidentFieldOptionSelect {
	ifos.fns = append(ifos.fns, fns...)
	return ifos
}

// Scan applies the selector query and scans the result into the given value.
func (ifos *IncidentFieldOptionSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, ifos.ctx, ent.OpQuerySelect)
	if err := ifos.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*IncidentFieldOptionQuery, *IncidentFieldOptionSelect](ctx, ifos.IncidentFieldOptionQuery, ifos, ifos.inters, v)
}

func (ifos *IncidentFieldOptionSelect) sqlScan(ctx context.Context, root *IncidentFieldOptionQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(ifos.fns))
	for _, fn := range ifos.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*ifos.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := ifos.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// Modify adds a query modifier for attaching custom logic to queries.
func (ifos *IncidentFieldOptionSelect) Modify(modifiers ...func(s *sql.Selector)) *IncidentFieldOptionSelect {
	ifos.modifiers = append(ifos.modifiers, modifiers...)
	return ifos
}