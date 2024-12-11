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
	"github.com/twohundreds/rezible/ent/oncallroster"
	"github.com/twohundreds/rezible/ent/oncallschedule"
	"github.com/twohundreds/rezible/ent/oncallscheduleparticipant"
	"github.com/twohundreds/rezible/ent/predicate"
)

// OncallScheduleQuery is the builder for querying OncallSchedule entities.
type OncallScheduleQuery struct {
	config
	ctx              *QueryContext
	order            []oncallschedule.OrderOption
	inters           []Interceptor
	predicates       []predicate.OncallSchedule
	withParticipants *OncallScheduleParticipantQuery
	withRoster       *OncallRosterQuery
	modifiers        []func(*sql.Selector)
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the OncallScheduleQuery builder.
func (osq *OncallScheduleQuery) Where(ps ...predicate.OncallSchedule) *OncallScheduleQuery {
	osq.predicates = append(osq.predicates, ps...)
	return osq
}

// Limit the number of records to be returned by this query.
func (osq *OncallScheduleQuery) Limit(limit int) *OncallScheduleQuery {
	osq.ctx.Limit = &limit
	return osq
}

// Offset to start from.
func (osq *OncallScheduleQuery) Offset(offset int) *OncallScheduleQuery {
	osq.ctx.Offset = &offset
	return osq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (osq *OncallScheduleQuery) Unique(unique bool) *OncallScheduleQuery {
	osq.ctx.Unique = &unique
	return osq
}

// Order specifies how the records should be ordered.
func (osq *OncallScheduleQuery) Order(o ...oncallschedule.OrderOption) *OncallScheduleQuery {
	osq.order = append(osq.order, o...)
	return osq
}

// QueryParticipants chains the current query on the "participants" edge.
func (osq *OncallScheduleQuery) QueryParticipants() *OncallScheduleParticipantQuery {
	query := (&OncallScheduleParticipantClient{config: osq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := osq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := osq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(oncallschedule.Table, oncallschedule.FieldID, selector),
			sqlgraph.To(oncallscheduleparticipant.Table, oncallscheduleparticipant.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, oncallschedule.ParticipantsTable, oncallschedule.ParticipantsColumn),
		)
		fromU = sqlgraph.SetNeighbors(osq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryRoster chains the current query on the "roster" edge.
func (osq *OncallScheduleQuery) QueryRoster() *OncallRosterQuery {
	query := (&OncallRosterClient{config: osq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := osq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := osq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(oncallschedule.Table, oncallschedule.FieldID, selector),
			sqlgraph.To(oncallroster.Table, oncallroster.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, oncallschedule.RosterTable, oncallschedule.RosterColumn),
		)
		fromU = sqlgraph.SetNeighbors(osq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first OncallSchedule entity from the query.
// Returns a *NotFoundError when no OncallSchedule was found.
func (osq *OncallScheduleQuery) First(ctx context.Context) (*OncallSchedule, error) {
	nodes, err := osq.Limit(1).All(setContextOp(ctx, osq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{oncallschedule.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (osq *OncallScheduleQuery) FirstX(ctx context.Context) *OncallSchedule {
	node, err := osq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first OncallSchedule ID from the query.
// Returns a *NotFoundError when no OncallSchedule ID was found.
func (osq *OncallScheduleQuery) FirstID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = osq.Limit(1).IDs(setContextOp(ctx, osq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{oncallschedule.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (osq *OncallScheduleQuery) FirstIDX(ctx context.Context) uuid.UUID {
	id, err := osq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single OncallSchedule entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one OncallSchedule entity is found.
// Returns a *NotFoundError when no OncallSchedule entities are found.
func (osq *OncallScheduleQuery) Only(ctx context.Context) (*OncallSchedule, error) {
	nodes, err := osq.Limit(2).All(setContextOp(ctx, osq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{oncallschedule.Label}
	default:
		return nil, &NotSingularError{oncallschedule.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (osq *OncallScheduleQuery) OnlyX(ctx context.Context) *OncallSchedule {
	node, err := osq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only OncallSchedule ID in the query.
// Returns a *NotSingularError when more than one OncallSchedule ID is found.
// Returns a *NotFoundError when no entities are found.
func (osq *OncallScheduleQuery) OnlyID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = osq.Limit(2).IDs(setContextOp(ctx, osq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{oncallschedule.Label}
	default:
		err = &NotSingularError{oncallschedule.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (osq *OncallScheduleQuery) OnlyIDX(ctx context.Context) uuid.UUID {
	id, err := osq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of OncallSchedules.
func (osq *OncallScheduleQuery) All(ctx context.Context) ([]*OncallSchedule, error) {
	ctx = setContextOp(ctx, osq.ctx, ent.OpQueryAll)
	if err := osq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*OncallSchedule, *OncallScheduleQuery]()
	return withInterceptors[[]*OncallSchedule](ctx, osq, qr, osq.inters)
}

// AllX is like All, but panics if an error occurs.
func (osq *OncallScheduleQuery) AllX(ctx context.Context) []*OncallSchedule {
	nodes, err := osq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of OncallSchedule IDs.
func (osq *OncallScheduleQuery) IDs(ctx context.Context) (ids []uuid.UUID, err error) {
	if osq.ctx.Unique == nil && osq.path != nil {
		osq.Unique(true)
	}
	ctx = setContextOp(ctx, osq.ctx, ent.OpQueryIDs)
	if err = osq.Select(oncallschedule.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (osq *OncallScheduleQuery) IDsX(ctx context.Context) []uuid.UUID {
	ids, err := osq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (osq *OncallScheduleQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, osq.ctx, ent.OpQueryCount)
	if err := osq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, osq, querierCount[*OncallScheduleQuery](), osq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (osq *OncallScheduleQuery) CountX(ctx context.Context) int {
	count, err := osq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (osq *OncallScheduleQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, osq.ctx, ent.OpQueryExist)
	switch _, err := osq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (osq *OncallScheduleQuery) ExistX(ctx context.Context) bool {
	exist, err := osq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the OncallScheduleQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (osq *OncallScheduleQuery) Clone() *OncallScheduleQuery {
	if osq == nil {
		return nil
	}
	return &OncallScheduleQuery{
		config:           osq.config,
		ctx:              osq.ctx.Clone(),
		order:            append([]oncallschedule.OrderOption{}, osq.order...),
		inters:           append([]Interceptor{}, osq.inters...),
		predicates:       append([]predicate.OncallSchedule{}, osq.predicates...),
		withParticipants: osq.withParticipants.Clone(),
		withRoster:       osq.withRoster.Clone(),
		// clone intermediate query.
		sql:       osq.sql.Clone(),
		path:      osq.path,
		modifiers: append([]func(*sql.Selector){}, osq.modifiers...),
	}
}

// WithParticipants tells the query-builder to eager-load the nodes that are connected to
// the "participants" edge. The optional arguments are used to configure the query builder of the edge.
func (osq *OncallScheduleQuery) WithParticipants(opts ...func(*OncallScheduleParticipantQuery)) *OncallScheduleQuery {
	query := (&OncallScheduleParticipantClient{config: osq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	osq.withParticipants = query
	return osq
}

// WithRoster tells the query-builder to eager-load the nodes that are connected to
// the "roster" edge. The optional arguments are used to configure the query builder of the edge.
func (osq *OncallScheduleQuery) WithRoster(opts ...func(*OncallRosterQuery)) *OncallScheduleQuery {
	query := (&OncallRosterClient{config: osq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	osq.withRoster = query
	return osq
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
//	client.OncallSchedule.Query().
//		GroupBy(oncallschedule.FieldArchiveTime).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (osq *OncallScheduleQuery) GroupBy(field string, fields ...string) *OncallScheduleGroupBy {
	osq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &OncallScheduleGroupBy{build: osq}
	grbuild.flds = &osq.ctx.Fields
	grbuild.label = oncallschedule.Label
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
//	client.OncallSchedule.Query().
//		Select(oncallschedule.FieldArchiveTime).
//		Scan(ctx, &v)
func (osq *OncallScheduleQuery) Select(fields ...string) *OncallScheduleSelect {
	osq.ctx.Fields = append(osq.ctx.Fields, fields...)
	sbuild := &OncallScheduleSelect{OncallScheduleQuery: osq}
	sbuild.label = oncallschedule.Label
	sbuild.flds, sbuild.scan = &osq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a OncallScheduleSelect configured with the given aggregations.
func (osq *OncallScheduleQuery) Aggregate(fns ...AggregateFunc) *OncallScheduleSelect {
	return osq.Select().Aggregate(fns...)
}

func (osq *OncallScheduleQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range osq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, osq); err != nil {
				return err
			}
		}
	}
	for _, f := range osq.ctx.Fields {
		if !oncallschedule.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if osq.path != nil {
		prev, err := osq.path(ctx)
		if err != nil {
			return err
		}
		osq.sql = prev
	}
	return nil
}

func (osq *OncallScheduleQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*OncallSchedule, error) {
	var (
		nodes       = []*OncallSchedule{}
		_spec       = osq.querySpec()
		loadedTypes = [2]bool{
			osq.withParticipants != nil,
			osq.withRoster != nil,
		}
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*OncallSchedule).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &OncallSchedule{config: osq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	if len(osq.modifiers) > 0 {
		_spec.Modifiers = osq.modifiers
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, osq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := osq.withParticipants; query != nil {
		if err := osq.loadParticipants(ctx, query, nodes,
			func(n *OncallSchedule) { n.Edges.Participants = []*OncallScheduleParticipant{} },
			func(n *OncallSchedule, e *OncallScheduleParticipant) {
				n.Edges.Participants = append(n.Edges.Participants, e)
			}); err != nil {
			return nil, err
		}
	}
	if query := osq.withRoster; query != nil {
		if err := osq.loadRoster(ctx, query, nodes, nil,
			func(n *OncallSchedule, e *OncallRoster) { n.Edges.Roster = e }); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (osq *OncallScheduleQuery) loadParticipants(ctx context.Context, query *OncallScheduleParticipantQuery, nodes []*OncallSchedule, init func(*OncallSchedule), assign func(*OncallSchedule, *OncallScheduleParticipant)) error {
	fks := make([]driver.Value, 0, len(nodes))
	nodeids := make(map[uuid.UUID]*OncallSchedule)
	for i := range nodes {
		fks = append(fks, nodes[i].ID)
		nodeids[nodes[i].ID] = nodes[i]
		if init != nil {
			init(nodes[i])
		}
	}
	if len(query.ctx.Fields) > 0 {
		query.ctx.AppendFieldOnce(oncallscheduleparticipant.FieldScheduleID)
	}
	query.Where(predicate.OncallScheduleParticipant(func(s *sql.Selector) {
		s.Where(sql.InValues(s.C(oncallschedule.ParticipantsColumn), fks...))
	}))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		fk := n.ScheduleID
		node, ok := nodeids[fk]
		if !ok {
			return fmt.Errorf(`unexpected referenced foreign-key "schedule_id" returned %v for node %v`, fk, n.ID)
		}
		assign(node, n)
	}
	return nil
}
func (osq *OncallScheduleQuery) loadRoster(ctx context.Context, query *OncallRosterQuery, nodes []*OncallSchedule, init func(*OncallSchedule), assign func(*OncallSchedule, *OncallRoster)) error {
	ids := make([]uuid.UUID, 0, len(nodes))
	nodeids := make(map[uuid.UUID][]*OncallSchedule)
	for i := range nodes {
		fk := nodes[i].RosterID
		if _, ok := nodeids[fk]; !ok {
			ids = append(ids, fk)
		}
		nodeids[fk] = append(nodeids[fk], nodes[i])
	}
	if len(ids) == 0 {
		return nil
	}
	query.Where(oncallroster.IDIn(ids...))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		nodes, ok := nodeids[n.ID]
		if !ok {
			return fmt.Errorf(`unexpected foreign-key "roster_id" returned %v`, n.ID)
		}
		for i := range nodes {
			assign(nodes[i], n)
		}
	}
	return nil
}

func (osq *OncallScheduleQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := osq.querySpec()
	if len(osq.modifiers) > 0 {
		_spec.Modifiers = osq.modifiers
	}
	_spec.Node.Columns = osq.ctx.Fields
	if len(osq.ctx.Fields) > 0 {
		_spec.Unique = osq.ctx.Unique != nil && *osq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, osq.driver, _spec)
}

func (osq *OncallScheduleQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(oncallschedule.Table, oncallschedule.Columns, sqlgraph.NewFieldSpec(oncallschedule.FieldID, field.TypeUUID))
	_spec.From = osq.sql
	if unique := osq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if osq.path != nil {
		_spec.Unique = true
	}
	if fields := osq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, oncallschedule.FieldID)
		for i := range fields {
			if fields[i] != oncallschedule.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
		if osq.withRoster != nil {
			_spec.Node.AddColumnOnce(oncallschedule.FieldRosterID)
		}
	}
	if ps := osq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := osq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := osq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := osq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (osq *OncallScheduleQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(osq.driver.Dialect())
	t1 := builder.Table(oncallschedule.Table)
	columns := osq.ctx.Fields
	if len(columns) == 0 {
		columns = oncallschedule.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if osq.sql != nil {
		selector = osq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if osq.ctx.Unique != nil && *osq.ctx.Unique {
		selector.Distinct()
	}
	for _, m := range osq.modifiers {
		m(selector)
	}
	for _, p := range osq.predicates {
		p(selector)
	}
	for _, p := range osq.order {
		p(selector)
	}
	if offset := osq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := osq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// Modify adds a query modifier for attaching custom logic to queries.
func (osq *OncallScheduleQuery) Modify(modifiers ...func(s *sql.Selector)) *OncallScheduleSelect {
	osq.modifiers = append(osq.modifiers, modifiers...)
	return osq.Select()
}

// OncallScheduleGroupBy is the group-by builder for OncallSchedule entities.
type OncallScheduleGroupBy struct {
	selector
	build *OncallScheduleQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (osgb *OncallScheduleGroupBy) Aggregate(fns ...AggregateFunc) *OncallScheduleGroupBy {
	osgb.fns = append(osgb.fns, fns...)
	return osgb
}

// Scan applies the selector query and scans the result into the given value.
func (osgb *OncallScheduleGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, osgb.build.ctx, ent.OpQueryGroupBy)
	if err := osgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*OncallScheduleQuery, *OncallScheduleGroupBy](ctx, osgb.build, osgb, osgb.build.inters, v)
}

func (osgb *OncallScheduleGroupBy) sqlScan(ctx context.Context, root *OncallScheduleQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(osgb.fns))
	for _, fn := range osgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*osgb.flds)+len(osgb.fns))
		for _, f := range *osgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*osgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := osgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// OncallScheduleSelect is the builder for selecting fields of OncallSchedule entities.
type OncallScheduleSelect struct {
	*OncallScheduleQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (oss *OncallScheduleSelect) Aggregate(fns ...AggregateFunc) *OncallScheduleSelect {
	oss.fns = append(oss.fns, fns...)
	return oss
}

// Scan applies the selector query and scans the result into the given value.
func (oss *OncallScheduleSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, oss.ctx, ent.OpQuerySelect)
	if err := oss.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*OncallScheduleQuery, *OncallScheduleSelect](ctx, oss.OncallScheduleQuery, oss, oss.inters, v)
}

func (oss *OncallScheduleSelect) sqlScan(ctx context.Context, root *OncallScheduleQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(oss.fns))
	for _, fn := range oss.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*oss.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := oss.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// Modify adds a query modifier for attaching custom logic to queries.
func (oss *OncallScheduleSelect) Modify(modifiers ...func(s *sql.Selector)) *OncallScheduleSelect {
	oss.modifiers = append(oss.modifiers, modifiers...)
	return oss
}
