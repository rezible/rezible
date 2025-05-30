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
	"github.com/rezible/rezible/ent/oncallannotation"
	"github.com/rezible/rezible/ent/oncallannotationalertfeedback"
	"github.com/rezible/rezible/ent/oncallevent"
	"github.com/rezible/rezible/ent/oncallroster"
	"github.com/rezible/rezible/ent/oncallusershifthandover"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/user"
)

// OncallAnnotationQuery is the builder for querying OncallAnnotation entities.
type OncallAnnotationQuery struct {
	config
	ctx               *QueryContext
	order             []oncallannotation.OrderOption
	inters            []Interceptor
	predicates        []predicate.OncallAnnotation
	withEvent         *OncallEventQuery
	withRoster        *OncallRosterQuery
	withCreator       *UserQuery
	withAlertFeedback *OncallAnnotationAlertFeedbackQuery
	withHandovers     *OncallUserShiftHandoverQuery
	modifiers         []func(*sql.Selector)
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the OncallAnnotationQuery builder.
func (oaq *OncallAnnotationQuery) Where(ps ...predicate.OncallAnnotation) *OncallAnnotationQuery {
	oaq.predicates = append(oaq.predicates, ps...)
	return oaq
}

// Limit the number of records to be returned by this query.
func (oaq *OncallAnnotationQuery) Limit(limit int) *OncallAnnotationQuery {
	oaq.ctx.Limit = &limit
	return oaq
}

// Offset to start from.
func (oaq *OncallAnnotationQuery) Offset(offset int) *OncallAnnotationQuery {
	oaq.ctx.Offset = &offset
	return oaq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (oaq *OncallAnnotationQuery) Unique(unique bool) *OncallAnnotationQuery {
	oaq.ctx.Unique = &unique
	return oaq
}

// Order specifies how the records should be ordered.
func (oaq *OncallAnnotationQuery) Order(o ...oncallannotation.OrderOption) *OncallAnnotationQuery {
	oaq.order = append(oaq.order, o...)
	return oaq
}

// QueryEvent chains the current query on the "event" edge.
func (oaq *OncallAnnotationQuery) QueryEvent() *OncallEventQuery {
	query := (&OncallEventClient{config: oaq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := oaq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := oaq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(oncallannotation.Table, oncallannotation.FieldID, selector),
			sqlgraph.To(oncallevent.Table, oncallevent.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, oncallannotation.EventTable, oncallannotation.EventColumn),
		)
		fromU = sqlgraph.SetNeighbors(oaq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryRoster chains the current query on the "roster" edge.
func (oaq *OncallAnnotationQuery) QueryRoster() *OncallRosterQuery {
	query := (&OncallRosterClient{config: oaq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := oaq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := oaq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(oncallannotation.Table, oncallannotation.FieldID, selector),
			sqlgraph.To(oncallroster.Table, oncallroster.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, oncallannotation.RosterTable, oncallannotation.RosterColumn),
		)
		fromU = sqlgraph.SetNeighbors(oaq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryCreator chains the current query on the "creator" edge.
func (oaq *OncallAnnotationQuery) QueryCreator() *UserQuery {
	query := (&UserClient{config: oaq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := oaq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := oaq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(oncallannotation.Table, oncallannotation.FieldID, selector),
			sqlgraph.To(user.Table, user.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, oncallannotation.CreatorTable, oncallannotation.CreatorColumn),
		)
		fromU = sqlgraph.SetNeighbors(oaq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryAlertFeedback chains the current query on the "alert_feedback" edge.
func (oaq *OncallAnnotationQuery) QueryAlertFeedback() *OncallAnnotationAlertFeedbackQuery {
	query := (&OncallAnnotationAlertFeedbackClient{config: oaq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := oaq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := oaq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(oncallannotation.Table, oncallannotation.FieldID, selector),
			sqlgraph.To(oncallannotationalertfeedback.Table, oncallannotationalertfeedback.FieldID),
			sqlgraph.Edge(sqlgraph.O2O, false, oncallannotation.AlertFeedbackTable, oncallannotation.AlertFeedbackColumn),
		)
		fromU = sqlgraph.SetNeighbors(oaq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryHandovers chains the current query on the "handovers" edge.
func (oaq *OncallAnnotationQuery) QueryHandovers() *OncallUserShiftHandoverQuery {
	query := (&OncallUserShiftHandoverClient{config: oaq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := oaq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := oaq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(oncallannotation.Table, oncallannotation.FieldID, selector),
			sqlgraph.To(oncallusershifthandover.Table, oncallusershifthandover.FieldID),
			sqlgraph.Edge(sqlgraph.M2M, true, oncallannotation.HandoversTable, oncallannotation.HandoversPrimaryKey...),
		)
		fromU = sqlgraph.SetNeighbors(oaq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first OncallAnnotation entity from the query.
// Returns a *NotFoundError when no OncallAnnotation was found.
func (oaq *OncallAnnotationQuery) First(ctx context.Context) (*OncallAnnotation, error) {
	nodes, err := oaq.Limit(1).All(setContextOp(ctx, oaq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{oncallannotation.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (oaq *OncallAnnotationQuery) FirstX(ctx context.Context) *OncallAnnotation {
	node, err := oaq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first OncallAnnotation ID from the query.
// Returns a *NotFoundError when no OncallAnnotation ID was found.
func (oaq *OncallAnnotationQuery) FirstID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = oaq.Limit(1).IDs(setContextOp(ctx, oaq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{oncallannotation.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (oaq *OncallAnnotationQuery) FirstIDX(ctx context.Context) uuid.UUID {
	id, err := oaq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single OncallAnnotation entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one OncallAnnotation entity is found.
// Returns a *NotFoundError when no OncallAnnotation entities are found.
func (oaq *OncallAnnotationQuery) Only(ctx context.Context) (*OncallAnnotation, error) {
	nodes, err := oaq.Limit(2).All(setContextOp(ctx, oaq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{oncallannotation.Label}
	default:
		return nil, &NotSingularError{oncallannotation.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (oaq *OncallAnnotationQuery) OnlyX(ctx context.Context) *OncallAnnotation {
	node, err := oaq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only OncallAnnotation ID in the query.
// Returns a *NotSingularError when more than one OncallAnnotation ID is found.
// Returns a *NotFoundError when no entities are found.
func (oaq *OncallAnnotationQuery) OnlyID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = oaq.Limit(2).IDs(setContextOp(ctx, oaq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{oncallannotation.Label}
	default:
		err = &NotSingularError{oncallannotation.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (oaq *OncallAnnotationQuery) OnlyIDX(ctx context.Context) uuid.UUID {
	id, err := oaq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of OncallAnnotations.
func (oaq *OncallAnnotationQuery) All(ctx context.Context) ([]*OncallAnnotation, error) {
	ctx = setContextOp(ctx, oaq.ctx, ent.OpQueryAll)
	if err := oaq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*OncallAnnotation, *OncallAnnotationQuery]()
	return withInterceptors[[]*OncallAnnotation](ctx, oaq, qr, oaq.inters)
}

// AllX is like All, but panics if an error occurs.
func (oaq *OncallAnnotationQuery) AllX(ctx context.Context) []*OncallAnnotation {
	nodes, err := oaq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of OncallAnnotation IDs.
func (oaq *OncallAnnotationQuery) IDs(ctx context.Context) (ids []uuid.UUID, err error) {
	if oaq.ctx.Unique == nil && oaq.path != nil {
		oaq.Unique(true)
	}
	ctx = setContextOp(ctx, oaq.ctx, ent.OpQueryIDs)
	if err = oaq.Select(oncallannotation.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (oaq *OncallAnnotationQuery) IDsX(ctx context.Context) []uuid.UUID {
	ids, err := oaq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (oaq *OncallAnnotationQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, oaq.ctx, ent.OpQueryCount)
	if err := oaq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, oaq, querierCount[*OncallAnnotationQuery](), oaq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (oaq *OncallAnnotationQuery) CountX(ctx context.Context) int {
	count, err := oaq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (oaq *OncallAnnotationQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, oaq.ctx, ent.OpQueryExist)
	switch _, err := oaq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (oaq *OncallAnnotationQuery) ExistX(ctx context.Context) bool {
	exist, err := oaq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the OncallAnnotationQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (oaq *OncallAnnotationQuery) Clone() *OncallAnnotationQuery {
	if oaq == nil {
		return nil
	}
	return &OncallAnnotationQuery{
		config:            oaq.config,
		ctx:               oaq.ctx.Clone(),
		order:             append([]oncallannotation.OrderOption{}, oaq.order...),
		inters:            append([]Interceptor{}, oaq.inters...),
		predicates:        append([]predicate.OncallAnnotation{}, oaq.predicates...),
		withEvent:         oaq.withEvent.Clone(),
		withRoster:        oaq.withRoster.Clone(),
		withCreator:       oaq.withCreator.Clone(),
		withAlertFeedback: oaq.withAlertFeedback.Clone(),
		withHandovers:     oaq.withHandovers.Clone(),
		// clone intermediate query.
		sql:       oaq.sql.Clone(),
		path:      oaq.path,
		modifiers: append([]func(*sql.Selector){}, oaq.modifiers...),
	}
}

// WithEvent tells the query-builder to eager-load the nodes that are connected to
// the "event" edge. The optional arguments are used to configure the query builder of the edge.
func (oaq *OncallAnnotationQuery) WithEvent(opts ...func(*OncallEventQuery)) *OncallAnnotationQuery {
	query := (&OncallEventClient{config: oaq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	oaq.withEvent = query
	return oaq
}

// WithRoster tells the query-builder to eager-load the nodes that are connected to
// the "roster" edge. The optional arguments are used to configure the query builder of the edge.
func (oaq *OncallAnnotationQuery) WithRoster(opts ...func(*OncallRosterQuery)) *OncallAnnotationQuery {
	query := (&OncallRosterClient{config: oaq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	oaq.withRoster = query
	return oaq
}

// WithCreator tells the query-builder to eager-load the nodes that are connected to
// the "creator" edge. The optional arguments are used to configure the query builder of the edge.
func (oaq *OncallAnnotationQuery) WithCreator(opts ...func(*UserQuery)) *OncallAnnotationQuery {
	query := (&UserClient{config: oaq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	oaq.withCreator = query
	return oaq
}

// WithAlertFeedback tells the query-builder to eager-load the nodes that are connected to
// the "alert_feedback" edge. The optional arguments are used to configure the query builder of the edge.
func (oaq *OncallAnnotationQuery) WithAlertFeedback(opts ...func(*OncallAnnotationAlertFeedbackQuery)) *OncallAnnotationQuery {
	query := (&OncallAnnotationAlertFeedbackClient{config: oaq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	oaq.withAlertFeedback = query
	return oaq
}

// WithHandovers tells the query-builder to eager-load the nodes that are connected to
// the "handovers" edge. The optional arguments are used to configure the query builder of the edge.
func (oaq *OncallAnnotationQuery) WithHandovers(opts ...func(*OncallUserShiftHandoverQuery)) *OncallAnnotationQuery {
	query := (&OncallUserShiftHandoverClient{config: oaq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	oaq.withHandovers = query
	return oaq
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		EventID uuid.UUID `json:"event_id,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.OncallAnnotation.Query().
//		GroupBy(oncallannotation.FieldEventID).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (oaq *OncallAnnotationQuery) GroupBy(field string, fields ...string) *OncallAnnotationGroupBy {
	oaq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &OncallAnnotationGroupBy{build: oaq}
	grbuild.flds = &oaq.ctx.Fields
	grbuild.label = oncallannotation.Label
	grbuild.scan = grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		EventID uuid.UUID `json:"event_id,omitempty"`
//	}
//
//	client.OncallAnnotation.Query().
//		Select(oncallannotation.FieldEventID).
//		Scan(ctx, &v)
func (oaq *OncallAnnotationQuery) Select(fields ...string) *OncallAnnotationSelect {
	oaq.ctx.Fields = append(oaq.ctx.Fields, fields...)
	sbuild := &OncallAnnotationSelect{OncallAnnotationQuery: oaq}
	sbuild.label = oncallannotation.Label
	sbuild.flds, sbuild.scan = &oaq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a OncallAnnotationSelect configured with the given aggregations.
func (oaq *OncallAnnotationQuery) Aggregate(fns ...AggregateFunc) *OncallAnnotationSelect {
	return oaq.Select().Aggregate(fns...)
}

func (oaq *OncallAnnotationQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range oaq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, oaq); err != nil {
				return err
			}
		}
	}
	for _, f := range oaq.ctx.Fields {
		if !oncallannotation.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if oaq.path != nil {
		prev, err := oaq.path(ctx)
		if err != nil {
			return err
		}
		oaq.sql = prev
	}
	return nil
}

func (oaq *OncallAnnotationQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*OncallAnnotation, error) {
	var (
		nodes       = []*OncallAnnotation{}
		_spec       = oaq.querySpec()
		loadedTypes = [5]bool{
			oaq.withEvent != nil,
			oaq.withRoster != nil,
			oaq.withCreator != nil,
			oaq.withAlertFeedback != nil,
			oaq.withHandovers != nil,
		}
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*OncallAnnotation).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &OncallAnnotation{config: oaq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	if len(oaq.modifiers) > 0 {
		_spec.Modifiers = oaq.modifiers
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, oaq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := oaq.withEvent; query != nil {
		if err := oaq.loadEvent(ctx, query, nodes, nil,
			func(n *OncallAnnotation, e *OncallEvent) { n.Edges.Event = e }); err != nil {
			return nil, err
		}
	}
	if query := oaq.withRoster; query != nil {
		if err := oaq.loadRoster(ctx, query, nodes, nil,
			func(n *OncallAnnotation, e *OncallRoster) { n.Edges.Roster = e }); err != nil {
			return nil, err
		}
	}
	if query := oaq.withCreator; query != nil {
		if err := oaq.loadCreator(ctx, query, nodes, nil,
			func(n *OncallAnnotation, e *User) { n.Edges.Creator = e }); err != nil {
			return nil, err
		}
	}
	if query := oaq.withAlertFeedback; query != nil {
		if err := oaq.loadAlertFeedback(ctx, query, nodes, nil,
			func(n *OncallAnnotation, e *OncallAnnotationAlertFeedback) { n.Edges.AlertFeedback = e }); err != nil {
			return nil, err
		}
	}
	if query := oaq.withHandovers; query != nil {
		if err := oaq.loadHandovers(ctx, query, nodes,
			func(n *OncallAnnotation) { n.Edges.Handovers = []*OncallUserShiftHandover{} },
			func(n *OncallAnnotation, e *OncallUserShiftHandover) {
				n.Edges.Handovers = append(n.Edges.Handovers, e)
			}); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (oaq *OncallAnnotationQuery) loadEvent(ctx context.Context, query *OncallEventQuery, nodes []*OncallAnnotation, init func(*OncallAnnotation), assign func(*OncallAnnotation, *OncallEvent)) error {
	ids := make([]uuid.UUID, 0, len(nodes))
	nodeids := make(map[uuid.UUID][]*OncallAnnotation)
	for i := range nodes {
		fk := nodes[i].EventID
		if _, ok := nodeids[fk]; !ok {
			ids = append(ids, fk)
		}
		nodeids[fk] = append(nodeids[fk], nodes[i])
	}
	if len(ids) == 0 {
		return nil
	}
	query.Where(oncallevent.IDIn(ids...))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		nodes, ok := nodeids[n.ID]
		if !ok {
			return fmt.Errorf(`unexpected foreign-key "event_id" returned %v`, n.ID)
		}
		for i := range nodes {
			assign(nodes[i], n)
		}
	}
	return nil
}
func (oaq *OncallAnnotationQuery) loadRoster(ctx context.Context, query *OncallRosterQuery, nodes []*OncallAnnotation, init func(*OncallAnnotation), assign func(*OncallAnnotation, *OncallRoster)) error {
	ids := make([]uuid.UUID, 0, len(nodes))
	nodeids := make(map[uuid.UUID][]*OncallAnnotation)
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
func (oaq *OncallAnnotationQuery) loadCreator(ctx context.Context, query *UserQuery, nodes []*OncallAnnotation, init func(*OncallAnnotation), assign func(*OncallAnnotation, *User)) error {
	ids := make([]uuid.UUID, 0, len(nodes))
	nodeids := make(map[uuid.UUID][]*OncallAnnotation)
	for i := range nodes {
		fk := nodes[i].CreatorID
		if _, ok := nodeids[fk]; !ok {
			ids = append(ids, fk)
		}
		nodeids[fk] = append(nodeids[fk], nodes[i])
	}
	if len(ids) == 0 {
		return nil
	}
	query.Where(user.IDIn(ids...))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		nodes, ok := nodeids[n.ID]
		if !ok {
			return fmt.Errorf(`unexpected foreign-key "creator_id" returned %v`, n.ID)
		}
		for i := range nodes {
			assign(nodes[i], n)
		}
	}
	return nil
}
func (oaq *OncallAnnotationQuery) loadAlertFeedback(ctx context.Context, query *OncallAnnotationAlertFeedbackQuery, nodes []*OncallAnnotation, init func(*OncallAnnotation), assign func(*OncallAnnotation, *OncallAnnotationAlertFeedback)) error {
	fks := make([]driver.Value, 0, len(nodes))
	nodeids := make(map[uuid.UUID]*OncallAnnotation)
	for i := range nodes {
		fks = append(fks, nodes[i].ID)
		nodeids[nodes[i].ID] = nodes[i]
	}
	if len(query.ctx.Fields) > 0 {
		query.ctx.AppendFieldOnce(oncallannotationalertfeedback.FieldAnnotationID)
	}
	query.Where(predicate.OncallAnnotationAlertFeedback(func(s *sql.Selector) {
		s.Where(sql.InValues(s.C(oncallannotation.AlertFeedbackColumn), fks...))
	}))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		fk := n.AnnotationID
		node, ok := nodeids[fk]
		if !ok {
			return fmt.Errorf(`unexpected referenced foreign-key "annotation_id" returned %v for node %v`, fk, n.ID)
		}
		assign(node, n)
	}
	return nil
}
func (oaq *OncallAnnotationQuery) loadHandovers(ctx context.Context, query *OncallUserShiftHandoverQuery, nodes []*OncallAnnotation, init func(*OncallAnnotation), assign func(*OncallAnnotation, *OncallUserShiftHandover)) error {
	edgeIDs := make([]driver.Value, len(nodes))
	byID := make(map[uuid.UUID]*OncallAnnotation)
	nids := make(map[uuid.UUID]map[*OncallAnnotation]struct{})
	for i, node := range nodes {
		edgeIDs[i] = node.ID
		byID[node.ID] = node
		if init != nil {
			init(node)
		}
	}
	query.Where(func(s *sql.Selector) {
		joinT := sql.Table(oncallannotation.HandoversTable)
		s.Join(joinT).On(s.C(oncallusershifthandover.FieldID), joinT.C(oncallannotation.HandoversPrimaryKey[0]))
		s.Where(sql.InValues(joinT.C(oncallannotation.HandoversPrimaryKey[1]), edgeIDs...))
		columns := s.SelectedColumns()
		s.Select(joinT.C(oncallannotation.HandoversPrimaryKey[1]))
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
					nids[inValue] = map[*OncallAnnotation]struct{}{byID[outValue]: {}}
					return assign(columns[1:], values[1:])
				}
				nids[inValue][byID[outValue]] = struct{}{}
				return nil
			}
		})
	})
	neighbors, err := withInterceptors[[]*OncallUserShiftHandover](ctx, query, qr, query.inters)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		nodes, ok := nids[n.ID]
		if !ok {
			return fmt.Errorf(`unexpected "handovers" node returned %v`, n.ID)
		}
		for kn := range nodes {
			assign(kn, n)
		}
	}
	return nil
}

func (oaq *OncallAnnotationQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := oaq.querySpec()
	if len(oaq.modifiers) > 0 {
		_spec.Modifiers = oaq.modifiers
	}
	_spec.Node.Columns = oaq.ctx.Fields
	if len(oaq.ctx.Fields) > 0 {
		_spec.Unique = oaq.ctx.Unique != nil && *oaq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, oaq.driver, _spec)
}

func (oaq *OncallAnnotationQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(oncallannotation.Table, oncallannotation.Columns, sqlgraph.NewFieldSpec(oncallannotation.FieldID, field.TypeUUID))
	_spec.From = oaq.sql
	if unique := oaq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if oaq.path != nil {
		_spec.Unique = true
	}
	if fields := oaq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, oncallannotation.FieldID)
		for i := range fields {
			if fields[i] != oncallannotation.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
		if oaq.withEvent != nil {
			_spec.Node.AddColumnOnce(oncallannotation.FieldEventID)
		}
		if oaq.withRoster != nil {
			_spec.Node.AddColumnOnce(oncallannotation.FieldRosterID)
		}
		if oaq.withCreator != nil {
			_spec.Node.AddColumnOnce(oncallannotation.FieldCreatorID)
		}
	}
	if ps := oaq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := oaq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := oaq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := oaq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (oaq *OncallAnnotationQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(oaq.driver.Dialect())
	t1 := builder.Table(oncallannotation.Table)
	columns := oaq.ctx.Fields
	if len(columns) == 0 {
		columns = oncallannotation.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if oaq.sql != nil {
		selector = oaq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if oaq.ctx.Unique != nil && *oaq.ctx.Unique {
		selector.Distinct()
	}
	for _, m := range oaq.modifiers {
		m(selector)
	}
	for _, p := range oaq.predicates {
		p(selector)
	}
	for _, p := range oaq.order {
		p(selector)
	}
	if offset := oaq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := oaq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// Modify adds a query modifier for attaching custom logic to queries.
func (oaq *OncallAnnotationQuery) Modify(modifiers ...func(s *sql.Selector)) *OncallAnnotationSelect {
	oaq.modifiers = append(oaq.modifiers, modifiers...)
	return oaq.Select()
}

// OncallAnnotationGroupBy is the group-by builder for OncallAnnotation entities.
type OncallAnnotationGroupBy struct {
	selector
	build *OncallAnnotationQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (oagb *OncallAnnotationGroupBy) Aggregate(fns ...AggregateFunc) *OncallAnnotationGroupBy {
	oagb.fns = append(oagb.fns, fns...)
	return oagb
}

// Scan applies the selector query and scans the result into the given value.
func (oagb *OncallAnnotationGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, oagb.build.ctx, ent.OpQueryGroupBy)
	if err := oagb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*OncallAnnotationQuery, *OncallAnnotationGroupBy](ctx, oagb.build, oagb, oagb.build.inters, v)
}

func (oagb *OncallAnnotationGroupBy) sqlScan(ctx context.Context, root *OncallAnnotationQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(oagb.fns))
	for _, fn := range oagb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*oagb.flds)+len(oagb.fns))
		for _, f := range *oagb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*oagb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := oagb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// OncallAnnotationSelect is the builder for selecting fields of OncallAnnotation entities.
type OncallAnnotationSelect struct {
	*OncallAnnotationQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (oas *OncallAnnotationSelect) Aggregate(fns ...AggregateFunc) *OncallAnnotationSelect {
	oas.fns = append(oas.fns, fns...)
	return oas
}

// Scan applies the selector query and scans the result into the given value.
func (oas *OncallAnnotationSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, oas.ctx, ent.OpQuerySelect)
	if err := oas.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*OncallAnnotationQuery, *OncallAnnotationSelect](ctx, oas.OncallAnnotationQuery, oas, oas.inters, v)
}

func (oas *OncallAnnotationSelect) sqlScan(ctx context.Context, root *OncallAnnotationQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(oas.fns))
	for _, fn := range oas.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*oas.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := oas.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// Modify adds a query modifier for attaching custom logic to queries.
func (oas *OncallAnnotationSelect) Modify(modifiers ...func(s *sql.Selector)) *OncallAnnotationSelect {
	oas.modifiers = append(oas.modifiers, modifiers...)
	return oas
}
