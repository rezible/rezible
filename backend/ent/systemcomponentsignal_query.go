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
	"github.com/rezible/rezible/ent/systemcomponentsignal"
	"github.com/rezible/rezible/ent/systemrelationship"
	"github.com/rezible/rezible/ent/systemrelationshipfeedbacksignal"
)

// SystemComponentSignalQuery is the builder for querying SystemComponentSignal entities.
type SystemComponentSignalQuery struct {
	config
	ctx                 *QueryContext
	order               []systemcomponentsignal.OrderOption
	inters              []Interceptor
	predicates          []predicate.SystemComponentSignal
	withComponent       *SystemComponentQuery
	withRelationships   *SystemRelationshipQuery
	withFeedbackSignals *SystemRelationshipFeedbackSignalQuery
	modifiers           []func(*sql.Selector)
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the SystemComponentSignalQuery builder.
func (scsq *SystemComponentSignalQuery) Where(ps ...predicate.SystemComponentSignal) *SystemComponentSignalQuery {
	scsq.predicates = append(scsq.predicates, ps...)
	return scsq
}

// Limit the number of records to be returned by this query.
func (scsq *SystemComponentSignalQuery) Limit(limit int) *SystemComponentSignalQuery {
	scsq.ctx.Limit = &limit
	return scsq
}

// Offset to start from.
func (scsq *SystemComponentSignalQuery) Offset(offset int) *SystemComponentSignalQuery {
	scsq.ctx.Offset = &offset
	return scsq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (scsq *SystemComponentSignalQuery) Unique(unique bool) *SystemComponentSignalQuery {
	scsq.ctx.Unique = &unique
	return scsq
}

// Order specifies how the records should be ordered.
func (scsq *SystemComponentSignalQuery) Order(o ...systemcomponentsignal.OrderOption) *SystemComponentSignalQuery {
	scsq.order = append(scsq.order, o...)
	return scsq
}

// QueryComponent chains the current query on the "component" edge.
func (scsq *SystemComponentSignalQuery) QueryComponent() *SystemComponentQuery {
	query := (&SystemComponentClient{config: scsq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := scsq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := scsq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(systemcomponentsignal.Table, systemcomponentsignal.FieldID, selector),
			sqlgraph.To(systemcomponent.Table, systemcomponent.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, systemcomponentsignal.ComponentTable, systemcomponentsignal.ComponentColumn),
		)
		fromU = sqlgraph.SetNeighbors(scsq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryRelationships chains the current query on the "relationships" edge.
func (scsq *SystemComponentSignalQuery) QueryRelationships() *SystemRelationshipQuery {
	query := (&SystemRelationshipClient{config: scsq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := scsq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := scsq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(systemcomponentsignal.Table, systemcomponentsignal.FieldID, selector),
			sqlgraph.To(systemrelationship.Table, systemrelationship.FieldID),
			sqlgraph.Edge(sqlgraph.M2M, true, systemcomponentsignal.RelationshipsTable, systemcomponentsignal.RelationshipsPrimaryKey...),
		)
		fromU = sqlgraph.SetNeighbors(scsq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryFeedbackSignals chains the current query on the "feedback_signals" edge.
func (scsq *SystemComponentSignalQuery) QueryFeedbackSignals() *SystemRelationshipFeedbackSignalQuery {
	query := (&SystemRelationshipFeedbackSignalClient{config: scsq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := scsq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := scsq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(systemcomponentsignal.Table, systemcomponentsignal.FieldID, selector),
			sqlgraph.To(systemrelationshipfeedbacksignal.Table, systemrelationshipfeedbacksignal.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, true, systemcomponentsignal.FeedbackSignalsTable, systemcomponentsignal.FeedbackSignalsColumn),
		)
		fromU = sqlgraph.SetNeighbors(scsq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first SystemComponentSignal entity from the query.
// Returns a *NotFoundError when no SystemComponentSignal was found.
func (scsq *SystemComponentSignalQuery) First(ctx context.Context) (*SystemComponentSignal, error) {
	nodes, err := scsq.Limit(1).All(setContextOp(ctx, scsq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{systemcomponentsignal.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (scsq *SystemComponentSignalQuery) FirstX(ctx context.Context) *SystemComponentSignal {
	node, err := scsq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first SystemComponentSignal ID from the query.
// Returns a *NotFoundError when no SystemComponentSignal ID was found.
func (scsq *SystemComponentSignalQuery) FirstID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = scsq.Limit(1).IDs(setContextOp(ctx, scsq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{systemcomponentsignal.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (scsq *SystemComponentSignalQuery) FirstIDX(ctx context.Context) uuid.UUID {
	id, err := scsq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single SystemComponentSignal entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one SystemComponentSignal entity is found.
// Returns a *NotFoundError when no SystemComponentSignal entities are found.
func (scsq *SystemComponentSignalQuery) Only(ctx context.Context) (*SystemComponentSignal, error) {
	nodes, err := scsq.Limit(2).All(setContextOp(ctx, scsq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{systemcomponentsignal.Label}
	default:
		return nil, &NotSingularError{systemcomponentsignal.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (scsq *SystemComponentSignalQuery) OnlyX(ctx context.Context) *SystemComponentSignal {
	node, err := scsq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only SystemComponentSignal ID in the query.
// Returns a *NotSingularError when more than one SystemComponentSignal ID is found.
// Returns a *NotFoundError when no entities are found.
func (scsq *SystemComponentSignalQuery) OnlyID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = scsq.Limit(2).IDs(setContextOp(ctx, scsq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{systemcomponentsignal.Label}
	default:
		err = &NotSingularError{systemcomponentsignal.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (scsq *SystemComponentSignalQuery) OnlyIDX(ctx context.Context) uuid.UUID {
	id, err := scsq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of SystemComponentSignals.
func (scsq *SystemComponentSignalQuery) All(ctx context.Context) ([]*SystemComponentSignal, error) {
	ctx = setContextOp(ctx, scsq.ctx, ent.OpQueryAll)
	if err := scsq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*SystemComponentSignal, *SystemComponentSignalQuery]()
	return withInterceptors[[]*SystemComponentSignal](ctx, scsq, qr, scsq.inters)
}

// AllX is like All, but panics if an error occurs.
func (scsq *SystemComponentSignalQuery) AllX(ctx context.Context) []*SystemComponentSignal {
	nodes, err := scsq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of SystemComponentSignal IDs.
func (scsq *SystemComponentSignalQuery) IDs(ctx context.Context) (ids []uuid.UUID, err error) {
	if scsq.ctx.Unique == nil && scsq.path != nil {
		scsq.Unique(true)
	}
	ctx = setContextOp(ctx, scsq.ctx, ent.OpQueryIDs)
	if err = scsq.Select(systemcomponentsignal.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (scsq *SystemComponentSignalQuery) IDsX(ctx context.Context) []uuid.UUID {
	ids, err := scsq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (scsq *SystemComponentSignalQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, scsq.ctx, ent.OpQueryCount)
	if err := scsq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, scsq, querierCount[*SystemComponentSignalQuery](), scsq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (scsq *SystemComponentSignalQuery) CountX(ctx context.Context) int {
	count, err := scsq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (scsq *SystemComponentSignalQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, scsq.ctx, ent.OpQueryExist)
	switch _, err := scsq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (scsq *SystemComponentSignalQuery) ExistX(ctx context.Context) bool {
	exist, err := scsq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the SystemComponentSignalQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (scsq *SystemComponentSignalQuery) Clone() *SystemComponentSignalQuery {
	if scsq == nil {
		return nil
	}
	return &SystemComponentSignalQuery{
		config:              scsq.config,
		ctx:                 scsq.ctx.Clone(),
		order:               append([]systemcomponentsignal.OrderOption{}, scsq.order...),
		inters:              append([]Interceptor{}, scsq.inters...),
		predicates:          append([]predicate.SystemComponentSignal{}, scsq.predicates...),
		withComponent:       scsq.withComponent.Clone(),
		withRelationships:   scsq.withRelationships.Clone(),
		withFeedbackSignals: scsq.withFeedbackSignals.Clone(),
		// clone intermediate query.
		sql:       scsq.sql.Clone(),
		path:      scsq.path,
		modifiers: append([]func(*sql.Selector){}, scsq.modifiers...),
	}
}

// WithComponent tells the query-builder to eager-load the nodes that are connected to
// the "component" edge. The optional arguments are used to configure the query builder of the edge.
func (scsq *SystemComponentSignalQuery) WithComponent(opts ...func(*SystemComponentQuery)) *SystemComponentSignalQuery {
	query := (&SystemComponentClient{config: scsq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	scsq.withComponent = query
	return scsq
}

// WithRelationships tells the query-builder to eager-load the nodes that are connected to
// the "relationships" edge. The optional arguments are used to configure the query builder of the edge.
func (scsq *SystemComponentSignalQuery) WithRelationships(opts ...func(*SystemRelationshipQuery)) *SystemComponentSignalQuery {
	query := (&SystemRelationshipClient{config: scsq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	scsq.withRelationships = query
	return scsq
}

// WithFeedbackSignals tells the query-builder to eager-load the nodes that are connected to
// the "feedback_signals" edge. The optional arguments are used to configure the query builder of the edge.
func (scsq *SystemComponentSignalQuery) WithFeedbackSignals(opts ...func(*SystemRelationshipFeedbackSignalQuery)) *SystemComponentSignalQuery {
	query := (&SystemRelationshipFeedbackSignalClient{config: scsq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	scsq.withFeedbackSignals = query
	return scsq
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
//	client.SystemComponentSignal.Query().
//		GroupBy(systemcomponentsignal.FieldComponentID).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (scsq *SystemComponentSignalQuery) GroupBy(field string, fields ...string) *SystemComponentSignalGroupBy {
	scsq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &SystemComponentSignalGroupBy{build: scsq}
	grbuild.flds = &scsq.ctx.Fields
	grbuild.label = systemcomponentsignal.Label
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
//	client.SystemComponentSignal.Query().
//		Select(systemcomponentsignal.FieldComponentID).
//		Scan(ctx, &v)
func (scsq *SystemComponentSignalQuery) Select(fields ...string) *SystemComponentSignalSelect {
	scsq.ctx.Fields = append(scsq.ctx.Fields, fields...)
	sbuild := &SystemComponentSignalSelect{SystemComponentSignalQuery: scsq}
	sbuild.label = systemcomponentsignal.Label
	sbuild.flds, sbuild.scan = &scsq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a SystemComponentSignalSelect configured with the given aggregations.
func (scsq *SystemComponentSignalQuery) Aggregate(fns ...AggregateFunc) *SystemComponentSignalSelect {
	return scsq.Select().Aggregate(fns...)
}

func (scsq *SystemComponentSignalQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range scsq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, scsq); err != nil {
				return err
			}
		}
	}
	for _, f := range scsq.ctx.Fields {
		if !systemcomponentsignal.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if scsq.path != nil {
		prev, err := scsq.path(ctx)
		if err != nil {
			return err
		}
		scsq.sql = prev
	}
	return nil
}

func (scsq *SystemComponentSignalQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*SystemComponentSignal, error) {
	var (
		nodes       = []*SystemComponentSignal{}
		_spec       = scsq.querySpec()
		loadedTypes = [3]bool{
			scsq.withComponent != nil,
			scsq.withRelationships != nil,
			scsq.withFeedbackSignals != nil,
		}
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*SystemComponentSignal).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &SystemComponentSignal{config: scsq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	if len(scsq.modifiers) > 0 {
		_spec.Modifiers = scsq.modifiers
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, scsq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := scsq.withComponent; query != nil {
		if err := scsq.loadComponent(ctx, query, nodes, nil,
			func(n *SystemComponentSignal, e *SystemComponent) { n.Edges.Component = e }); err != nil {
			return nil, err
		}
	}
	if query := scsq.withRelationships; query != nil {
		if err := scsq.loadRelationships(ctx, query, nodes,
			func(n *SystemComponentSignal) { n.Edges.Relationships = []*SystemRelationship{} },
			func(n *SystemComponentSignal, e *SystemRelationship) {
				n.Edges.Relationships = append(n.Edges.Relationships, e)
			}); err != nil {
			return nil, err
		}
	}
	if query := scsq.withFeedbackSignals; query != nil {
		if err := scsq.loadFeedbackSignals(ctx, query, nodes,
			func(n *SystemComponentSignal) { n.Edges.FeedbackSignals = []*SystemRelationshipFeedbackSignal{} },
			func(n *SystemComponentSignal, e *SystemRelationshipFeedbackSignal) {
				n.Edges.FeedbackSignals = append(n.Edges.FeedbackSignals, e)
			}); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (scsq *SystemComponentSignalQuery) loadComponent(ctx context.Context, query *SystemComponentQuery, nodes []*SystemComponentSignal, init func(*SystemComponentSignal), assign func(*SystemComponentSignal, *SystemComponent)) error {
	ids := make([]uuid.UUID, 0, len(nodes))
	nodeids := make(map[uuid.UUID][]*SystemComponentSignal)
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
func (scsq *SystemComponentSignalQuery) loadRelationships(ctx context.Context, query *SystemRelationshipQuery, nodes []*SystemComponentSignal, init func(*SystemComponentSignal), assign func(*SystemComponentSignal, *SystemRelationship)) error {
	edgeIDs := make([]driver.Value, len(nodes))
	byID := make(map[uuid.UUID]*SystemComponentSignal)
	nids := make(map[uuid.UUID]map[*SystemComponentSignal]struct{})
	for i, node := range nodes {
		edgeIDs[i] = node.ID
		byID[node.ID] = node
		if init != nil {
			init(node)
		}
	}
	query.Where(func(s *sql.Selector) {
		joinT := sql.Table(systemcomponentsignal.RelationshipsTable)
		s.Join(joinT).On(s.C(systemrelationship.FieldID), joinT.C(systemcomponentsignal.RelationshipsPrimaryKey[0]))
		s.Where(sql.InValues(joinT.C(systemcomponentsignal.RelationshipsPrimaryKey[1]), edgeIDs...))
		columns := s.SelectedColumns()
		s.Select(joinT.C(systemcomponentsignal.RelationshipsPrimaryKey[1]))
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
					nids[inValue] = map[*SystemComponentSignal]struct{}{byID[outValue]: {}}
					return assign(columns[1:], values[1:])
				}
				nids[inValue][byID[outValue]] = struct{}{}
				return nil
			}
		})
	})
	neighbors, err := withInterceptors[[]*SystemRelationship](ctx, query, qr, query.inters)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		nodes, ok := nids[n.ID]
		if !ok {
			return fmt.Errorf(`unexpected "relationships" node returned %v`, n.ID)
		}
		for kn := range nodes {
			assign(kn, n)
		}
	}
	return nil
}
func (scsq *SystemComponentSignalQuery) loadFeedbackSignals(ctx context.Context, query *SystemRelationshipFeedbackSignalQuery, nodes []*SystemComponentSignal, init func(*SystemComponentSignal), assign func(*SystemComponentSignal, *SystemRelationshipFeedbackSignal)) error {
	fks := make([]driver.Value, 0, len(nodes))
	nodeids := make(map[uuid.UUID]*SystemComponentSignal)
	for i := range nodes {
		fks = append(fks, nodes[i].ID)
		nodeids[nodes[i].ID] = nodes[i]
		if init != nil {
			init(nodes[i])
		}
	}
	if len(query.ctx.Fields) > 0 {
		query.ctx.AppendFieldOnce(systemrelationshipfeedbacksignal.FieldSignalID)
	}
	query.Where(predicate.SystemRelationshipFeedbackSignal(func(s *sql.Selector) {
		s.Where(sql.InValues(s.C(systemcomponentsignal.FeedbackSignalsColumn), fks...))
	}))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		fk := n.SignalID
		node, ok := nodeids[fk]
		if !ok {
			return fmt.Errorf(`unexpected referenced foreign-key "signal_id" returned %v for node %v`, fk, n.ID)
		}
		assign(node, n)
	}
	return nil
}

func (scsq *SystemComponentSignalQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := scsq.querySpec()
	if len(scsq.modifiers) > 0 {
		_spec.Modifiers = scsq.modifiers
	}
	_spec.Node.Columns = scsq.ctx.Fields
	if len(scsq.ctx.Fields) > 0 {
		_spec.Unique = scsq.ctx.Unique != nil && *scsq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, scsq.driver, _spec)
}

func (scsq *SystemComponentSignalQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(systemcomponentsignal.Table, systemcomponentsignal.Columns, sqlgraph.NewFieldSpec(systemcomponentsignal.FieldID, field.TypeUUID))
	_spec.From = scsq.sql
	if unique := scsq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if scsq.path != nil {
		_spec.Unique = true
	}
	if fields := scsq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, systemcomponentsignal.FieldID)
		for i := range fields {
			if fields[i] != systemcomponentsignal.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
		if scsq.withComponent != nil {
			_spec.Node.AddColumnOnce(systemcomponentsignal.FieldComponentID)
		}
	}
	if ps := scsq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := scsq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := scsq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := scsq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (scsq *SystemComponentSignalQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(scsq.driver.Dialect())
	t1 := builder.Table(systemcomponentsignal.Table)
	columns := scsq.ctx.Fields
	if len(columns) == 0 {
		columns = systemcomponentsignal.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if scsq.sql != nil {
		selector = scsq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if scsq.ctx.Unique != nil && *scsq.ctx.Unique {
		selector.Distinct()
	}
	for _, m := range scsq.modifiers {
		m(selector)
	}
	for _, p := range scsq.predicates {
		p(selector)
	}
	for _, p := range scsq.order {
		p(selector)
	}
	if offset := scsq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := scsq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// Modify adds a query modifier for attaching custom logic to queries.
func (scsq *SystemComponentSignalQuery) Modify(modifiers ...func(s *sql.Selector)) *SystemComponentSignalSelect {
	scsq.modifiers = append(scsq.modifiers, modifiers...)
	return scsq.Select()
}

// SystemComponentSignalGroupBy is the group-by builder for SystemComponentSignal entities.
type SystemComponentSignalGroupBy struct {
	selector
	build *SystemComponentSignalQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (scsgb *SystemComponentSignalGroupBy) Aggregate(fns ...AggregateFunc) *SystemComponentSignalGroupBy {
	scsgb.fns = append(scsgb.fns, fns...)
	return scsgb
}

// Scan applies the selector query and scans the result into the given value.
func (scsgb *SystemComponentSignalGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, scsgb.build.ctx, ent.OpQueryGroupBy)
	if err := scsgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*SystemComponentSignalQuery, *SystemComponentSignalGroupBy](ctx, scsgb.build, scsgb, scsgb.build.inters, v)
}

func (scsgb *SystemComponentSignalGroupBy) sqlScan(ctx context.Context, root *SystemComponentSignalQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(scsgb.fns))
	for _, fn := range scsgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*scsgb.flds)+len(scsgb.fns))
		for _, f := range *scsgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*scsgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := scsgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// SystemComponentSignalSelect is the builder for selecting fields of SystemComponentSignal entities.
type SystemComponentSignalSelect struct {
	*SystemComponentSignalQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (scss *SystemComponentSignalSelect) Aggregate(fns ...AggregateFunc) *SystemComponentSignalSelect {
	scss.fns = append(scss.fns, fns...)
	return scss
}

// Scan applies the selector query and scans the result into the given value.
func (scss *SystemComponentSignalSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, scss.ctx, ent.OpQuerySelect)
	if err := scss.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*SystemComponentSignalQuery, *SystemComponentSignalSelect](ctx, scss.SystemComponentSignalQuery, scss, scss.inters, v)
}

func (scss *SystemComponentSignalSelect) sqlScan(ctx context.Context, root *SystemComponentSignalQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(scss.fns))
	for _, fn := range scss.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*scss.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := scss.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// Modify adds a query modifier for attaching custom logic to queries.
func (scss *SystemComponentSignalSelect) Modify(modifiers ...func(s *sql.Selector)) *SystemComponentSignalSelect {
	scss.modifiers = append(scss.modifiers, modifiers...)
	return scss
}
