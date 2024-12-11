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
	"github.com/rezible/rezible/ent/incidentdebrief"
	"github.com/rezible/rezible/ent/incidentdebriefmessage"
	"github.com/rezible/rezible/ent/incidentdebriefsuggestion"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/user"
)

// IncidentDebriefQuery is the builder for querying IncidentDebrief entities.
type IncidentDebriefQuery struct {
	config
	ctx             *QueryContext
	order           []incidentdebrief.OrderOption
	inters          []Interceptor
	predicates      []predicate.IncidentDebrief
	withIncident    *IncidentQuery
	withUser        *UserQuery
	withMessages    *IncidentDebriefMessageQuery
	withSuggestions *IncidentDebriefSuggestionQuery
	modifiers       []func(*sql.Selector)
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the IncidentDebriefQuery builder.
func (idq *IncidentDebriefQuery) Where(ps ...predicate.IncidentDebrief) *IncidentDebriefQuery {
	idq.predicates = append(idq.predicates, ps...)
	return idq
}

// Limit the number of records to be returned by this query.
func (idq *IncidentDebriefQuery) Limit(limit int) *IncidentDebriefQuery {
	idq.ctx.Limit = &limit
	return idq
}

// Offset to start from.
func (idq *IncidentDebriefQuery) Offset(offset int) *IncidentDebriefQuery {
	idq.ctx.Offset = &offset
	return idq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (idq *IncidentDebriefQuery) Unique(unique bool) *IncidentDebriefQuery {
	idq.ctx.Unique = &unique
	return idq
}

// Order specifies how the records should be ordered.
func (idq *IncidentDebriefQuery) Order(o ...incidentdebrief.OrderOption) *IncidentDebriefQuery {
	idq.order = append(idq.order, o...)
	return idq
}

// QueryIncident chains the current query on the "incident" edge.
func (idq *IncidentDebriefQuery) QueryIncident() *IncidentQuery {
	query := (&IncidentClient{config: idq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := idq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := idq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(incidentdebrief.Table, incidentdebrief.FieldID, selector),
			sqlgraph.To(incident.Table, incident.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, incidentdebrief.IncidentTable, incidentdebrief.IncidentColumn),
		)
		fromU = sqlgraph.SetNeighbors(idq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryUser chains the current query on the "user" edge.
func (idq *IncidentDebriefQuery) QueryUser() *UserQuery {
	query := (&UserClient{config: idq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := idq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := idq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(incidentdebrief.Table, incidentdebrief.FieldID, selector),
			sqlgraph.To(user.Table, user.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, incidentdebrief.UserTable, incidentdebrief.UserColumn),
		)
		fromU = sqlgraph.SetNeighbors(idq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryMessages chains the current query on the "messages" edge.
func (idq *IncidentDebriefQuery) QueryMessages() *IncidentDebriefMessageQuery {
	query := (&IncidentDebriefMessageClient{config: idq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := idq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := idq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(incidentdebrief.Table, incidentdebrief.FieldID, selector),
			sqlgraph.To(incidentdebriefmessage.Table, incidentdebriefmessage.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, incidentdebrief.MessagesTable, incidentdebrief.MessagesColumn),
		)
		fromU = sqlgraph.SetNeighbors(idq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QuerySuggestions chains the current query on the "suggestions" edge.
func (idq *IncidentDebriefQuery) QuerySuggestions() *IncidentDebriefSuggestionQuery {
	query := (&IncidentDebriefSuggestionClient{config: idq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := idq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := idq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(incidentdebrief.Table, incidentdebrief.FieldID, selector),
			sqlgraph.To(incidentdebriefsuggestion.Table, incidentdebriefsuggestion.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, incidentdebrief.SuggestionsTable, incidentdebrief.SuggestionsColumn),
		)
		fromU = sqlgraph.SetNeighbors(idq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first IncidentDebrief entity from the query.
// Returns a *NotFoundError when no IncidentDebrief was found.
func (idq *IncidentDebriefQuery) First(ctx context.Context) (*IncidentDebrief, error) {
	nodes, err := idq.Limit(1).All(setContextOp(ctx, idq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{incidentdebrief.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (idq *IncidentDebriefQuery) FirstX(ctx context.Context) *IncidentDebrief {
	node, err := idq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first IncidentDebrief ID from the query.
// Returns a *NotFoundError when no IncidentDebrief ID was found.
func (idq *IncidentDebriefQuery) FirstID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = idq.Limit(1).IDs(setContextOp(ctx, idq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{incidentdebrief.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (idq *IncidentDebriefQuery) FirstIDX(ctx context.Context) uuid.UUID {
	id, err := idq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single IncidentDebrief entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one IncidentDebrief entity is found.
// Returns a *NotFoundError when no IncidentDebrief entities are found.
func (idq *IncidentDebriefQuery) Only(ctx context.Context) (*IncidentDebrief, error) {
	nodes, err := idq.Limit(2).All(setContextOp(ctx, idq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{incidentdebrief.Label}
	default:
		return nil, &NotSingularError{incidentdebrief.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (idq *IncidentDebriefQuery) OnlyX(ctx context.Context) *IncidentDebrief {
	node, err := idq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only IncidentDebrief ID in the query.
// Returns a *NotSingularError when more than one IncidentDebrief ID is found.
// Returns a *NotFoundError when no entities are found.
func (idq *IncidentDebriefQuery) OnlyID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = idq.Limit(2).IDs(setContextOp(ctx, idq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{incidentdebrief.Label}
	default:
		err = &NotSingularError{incidentdebrief.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (idq *IncidentDebriefQuery) OnlyIDX(ctx context.Context) uuid.UUID {
	id, err := idq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of IncidentDebriefs.
func (idq *IncidentDebriefQuery) All(ctx context.Context) ([]*IncidentDebrief, error) {
	ctx = setContextOp(ctx, idq.ctx, ent.OpQueryAll)
	if err := idq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*IncidentDebrief, *IncidentDebriefQuery]()
	return withInterceptors[[]*IncidentDebrief](ctx, idq, qr, idq.inters)
}

// AllX is like All, but panics if an error occurs.
func (idq *IncidentDebriefQuery) AllX(ctx context.Context) []*IncidentDebrief {
	nodes, err := idq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of IncidentDebrief IDs.
func (idq *IncidentDebriefQuery) IDs(ctx context.Context) (ids []uuid.UUID, err error) {
	if idq.ctx.Unique == nil && idq.path != nil {
		idq.Unique(true)
	}
	ctx = setContextOp(ctx, idq.ctx, ent.OpQueryIDs)
	if err = idq.Select(incidentdebrief.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (idq *IncidentDebriefQuery) IDsX(ctx context.Context) []uuid.UUID {
	ids, err := idq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (idq *IncidentDebriefQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, idq.ctx, ent.OpQueryCount)
	if err := idq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, idq, querierCount[*IncidentDebriefQuery](), idq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (idq *IncidentDebriefQuery) CountX(ctx context.Context) int {
	count, err := idq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (idq *IncidentDebriefQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, idq.ctx, ent.OpQueryExist)
	switch _, err := idq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (idq *IncidentDebriefQuery) ExistX(ctx context.Context) bool {
	exist, err := idq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the IncidentDebriefQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (idq *IncidentDebriefQuery) Clone() *IncidentDebriefQuery {
	if idq == nil {
		return nil
	}
	return &IncidentDebriefQuery{
		config:          idq.config,
		ctx:             idq.ctx.Clone(),
		order:           append([]incidentdebrief.OrderOption{}, idq.order...),
		inters:          append([]Interceptor{}, idq.inters...),
		predicates:      append([]predicate.IncidentDebrief{}, idq.predicates...),
		withIncident:    idq.withIncident.Clone(),
		withUser:        idq.withUser.Clone(),
		withMessages:    idq.withMessages.Clone(),
		withSuggestions: idq.withSuggestions.Clone(),
		// clone intermediate query.
		sql:       idq.sql.Clone(),
		path:      idq.path,
		modifiers: append([]func(*sql.Selector){}, idq.modifiers...),
	}
}

// WithIncident tells the query-builder to eager-load the nodes that are connected to
// the "incident" edge. The optional arguments are used to configure the query builder of the edge.
func (idq *IncidentDebriefQuery) WithIncident(opts ...func(*IncidentQuery)) *IncidentDebriefQuery {
	query := (&IncidentClient{config: idq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	idq.withIncident = query
	return idq
}

// WithUser tells the query-builder to eager-load the nodes that are connected to
// the "user" edge. The optional arguments are used to configure the query builder of the edge.
func (idq *IncidentDebriefQuery) WithUser(opts ...func(*UserQuery)) *IncidentDebriefQuery {
	query := (&UserClient{config: idq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	idq.withUser = query
	return idq
}

// WithMessages tells the query-builder to eager-load the nodes that are connected to
// the "messages" edge. The optional arguments are used to configure the query builder of the edge.
func (idq *IncidentDebriefQuery) WithMessages(opts ...func(*IncidentDebriefMessageQuery)) *IncidentDebriefQuery {
	query := (&IncidentDebriefMessageClient{config: idq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	idq.withMessages = query
	return idq
}

// WithSuggestions tells the query-builder to eager-load the nodes that are connected to
// the "suggestions" edge. The optional arguments are used to configure the query builder of the edge.
func (idq *IncidentDebriefQuery) WithSuggestions(opts ...func(*IncidentDebriefSuggestionQuery)) *IncidentDebriefQuery {
	query := (&IncidentDebriefSuggestionClient{config: idq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	idq.withSuggestions = query
	return idq
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
//	client.IncidentDebrief.Query().
//		GroupBy(incidentdebrief.FieldIncidentID).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (idq *IncidentDebriefQuery) GroupBy(field string, fields ...string) *IncidentDebriefGroupBy {
	idq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &IncidentDebriefGroupBy{build: idq}
	grbuild.flds = &idq.ctx.Fields
	grbuild.label = incidentdebrief.Label
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
//	client.IncidentDebrief.Query().
//		Select(incidentdebrief.FieldIncidentID).
//		Scan(ctx, &v)
func (idq *IncidentDebriefQuery) Select(fields ...string) *IncidentDebriefSelect {
	idq.ctx.Fields = append(idq.ctx.Fields, fields...)
	sbuild := &IncidentDebriefSelect{IncidentDebriefQuery: idq}
	sbuild.label = incidentdebrief.Label
	sbuild.flds, sbuild.scan = &idq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a IncidentDebriefSelect configured with the given aggregations.
func (idq *IncidentDebriefQuery) Aggregate(fns ...AggregateFunc) *IncidentDebriefSelect {
	return idq.Select().Aggregate(fns...)
}

func (idq *IncidentDebriefQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range idq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, idq); err != nil {
				return err
			}
		}
	}
	for _, f := range idq.ctx.Fields {
		if !incidentdebrief.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if idq.path != nil {
		prev, err := idq.path(ctx)
		if err != nil {
			return err
		}
		idq.sql = prev
	}
	return nil
}

func (idq *IncidentDebriefQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*IncidentDebrief, error) {
	var (
		nodes       = []*IncidentDebrief{}
		_spec       = idq.querySpec()
		loadedTypes = [4]bool{
			idq.withIncident != nil,
			idq.withUser != nil,
			idq.withMessages != nil,
			idq.withSuggestions != nil,
		}
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*IncidentDebrief).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &IncidentDebrief{config: idq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	if len(idq.modifiers) > 0 {
		_spec.Modifiers = idq.modifiers
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, idq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := idq.withIncident; query != nil {
		if err := idq.loadIncident(ctx, query, nodes, nil,
			func(n *IncidentDebrief, e *Incident) { n.Edges.Incident = e }); err != nil {
			return nil, err
		}
	}
	if query := idq.withUser; query != nil {
		if err := idq.loadUser(ctx, query, nodes, nil,
			func(n *IncidentDebrief, e *User) { n.Edges.User = e }); err != nil {
			return nil, err
		}
	}
	if query := idq.withMessages; query != nil {
		if err := idq.loadMessages(ctx, query, nodes,
			func(n *IncidentDebrief) { n.Edges.Messages = []*IncidentDebriefMessage{} },
			func(n *IncidentDebrief, e *IncidentDebriefMessage) { n.Edges.Messages = append(n.Edges.Messages, e) }); err != nil {
			return nil, err
		}
	}
	if query := idq.withSuggestions; query != nil {
		if err := idq.loadSuggestions(ctx, query, nodes,
			func(n *IncidentDebrief) { n.Edges.Suggestions = []*IncidentDebriefSuggestion{} },
			func(n *IncidentDebrief, e *IncidentDebriefSuggestion) {
				n.Edges.Suggestions = append(n.Edges.Suggestions, e)
			}); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (idq *IncidentDebriefQuery) loadIncident(ctx context.Context, query *IncidentQuery, nodes []*IncidentDebrief, init func(*IncidentDebrief), assign func(*IncidentDebrief, *Incident)) error {
	ids := make([]uuid.UUID, 0, len(nodes))
	nodeids := make(map[uuid.UUID][]*IncidentDebrief)
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
func (idq *IncidentDebriefQuery) loadUser(ctx context.Context, query *UserQuery, nodes []*IncidentDebrief, init func(*IncidentDebrief), assign func(*IncidentDebrief, *User)) error {
	ids := make([]uuid.UUID, 0, len(nodes))
	nodeids := make(map[uuid.UUID][]*IncidentDebrief)
	for i := range nodes {
		fk := nodes[i].UserID
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
			return fmt.Errorf(`unexpected foreign-key "user_id" returned %v`, n.ID)
		}
		for i := range nodes {
			assign(nodes[i], n)
		}
	}
	return nil
}
func (idq *IncidentDebriefQuery) loadMessages(ctx context.Context, query *IncidentDebriefMessageQuery, nodes []*IncidentDebrief, init func(*IncidentDebrief), assign func(*IncidentDebrief, *IncidentDebriefMessage)) error {
	fks := make([]driver.Value, 0, len(nodes))
	nodeids := make(map[uuid.UUID]*IncidentDebrief)
	for i := range nodes {
		fks = append(fks, nodes[i].ID)
		nodeids[nodes[i].ID] = nodes[i]
		if init != nil {
			init(nodes[i])
		}
	}
	if len(query.ctx.Fields) > 0 {
		query.ctx.AppendFieldOnce(incidentdebriefmessage.FieldDebriefID)
	}
	query.Where(predicate.IncidentDebriefMessage(func(s *sql.Selector) {
		s.Where(sql.InValues(s.C(incidentdebrief.MessagesColumn), fks...))
	}))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		fk := n.DebriefID
		node, ok := nodeids[fk]
		if !ok {
			return fmt.Errorf(`unexpected referenced foreign-key "debrief_id" returned %v for node %v`, fk, n.ID)
		}
		assign(node, n)
	}
	return nil
}
func (idq *IncidentDebriefQuery) loadSuggestions(ctx context.Context, query *IncidentDebriefSuggestionQuery, nodes []*IncidentDebrief, init func(*IncidentDebrief), assign func(*IncidentDebrief, *IncidentDebriefSuggestion)) error {
	fks := make([]driver.Value, 0, len(nodes))
	nodeids := make(map[uuid.UUID]*IncidentDebrief)
	for i := range nodes {
		fks = append(fks, nodes[i].ID)
		nodeids[nodes[i].ID] = nodes[i]
		if init != nil {
			init(nodes[i])
		}
	}
	query.withFKs = true
	query.Where(predicate.IncidentDebriefSuggestion(func(s *sql.Selector) {
		s.Where(sql.InValues(s.C(incidentdebrief.SuggestionsColumn), fks...))
	}))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		fk := n.incident_debrief_suggestions
		if fk == nil {
			return fmt.Errorf(`foreign-key "incident_debrief_suggestions" is nil for node %v`, n.ID)
		}
		node, ok := nodeids[*fk]
		if !ok {
			return fmt.Errorf(`unexpected referenced foreign-key "incident_debrief_suggestions" returned %v for node %v`, *fk, n.ID)
		}
		assign(node, n)
	}
	return nil
}

func (idq *IncidentDebriefQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := idq.querySpec()
	if len(idq.modifiers) > 0 {
		_spec.Modifiers = idq.modifiers
	}
	_spec.Node.Columns = idq.ctx.Fields
	if len(idq.ctx.Fields) > 0 {
		_spec.Unique = idq.ctx.Unique != nil && *idq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, idq.driver, _spec)
}

func (idq *IncidentDebriefQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(incidentdebrief.Table, incidentdebrief.Columns, sqlgraph.NewFieldSpec(incidentdebrief.FieldID, field.TypeUUID))
	_spec.From = idq.sql
	if unique := idq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if idq.path != nil {
		_spec.Unique = true
	}
	if fields := idq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, incidentdebrief.FieldID)
		for i := range fields {
			if fields[i] != incidentdebrief.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
		if idq.withIncident != nil {
			_spec.Node.AddColumnOnce(incidentdebrief.FieldIncidentID)
		}
		if idq.withUser != nil {
			_spec.Node.AddColumnOnce(incidentdebrief.FieldUserID)
		}
	}
	if ps := idq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := idq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := idq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := idq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (idq *IncidentDebriefQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(idq.driver.Dialect())
	t1 := builder.Table(incidentdebrief.Table)
	columns := idq.ctx.Fields
	if len(columns) == 0 {
		columns = incidentdebrief.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if idq.sql != nil {
		selector = idq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if idq.ctx.Unique != nil && *idq.ctx.Unique {
		selector.Distinct()
	}
	for _, m := range idq.modifiers {
		m(selector)
	}
	for _, p := range idq.predicates {
		p(selector)
	}
	for _, p := range idq.order {
		p(selector)
	}
	if offset := idq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := idq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// Modify adds a query modifier for attaching custom logic to queries.
func (idq *IncidentDebriefQuery) Modify(modifiers ...func(s *sql.Selector)) *IncidentDebriefSelect {
	idq.modifiers = append(idq.modifiers, modifiers...)
	return idq.Select()
}

// IncidentDebriefGroupBy is the group-by builder for IncidentDebrief entities.
type IncidentDebriefGroupBy struct {
	selector
	build *IncidentDebriefQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (idgb *IncidentDebriefGroupBy) Aggregate(fns ...AggregateFunc) *IncidentDebriefGroupBy {
	idgb.fns = append(idgb.fns, fns...)
	return idgb
}

// Scan applies the selector query and scans the result into the given value.
func (idgb *IncidentDebriefGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, idgb.build.ctx, ent.OpQueryGroupBy)
	if err := idgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*IncidentDebriefQuery, *IncidentDebriefGroupBy](ctx, idgb.build, idgb, idgb.build.inters, v)
}

func (idgb *IncidentDebriefGroupBy) sqlScan(ctx context.Context, root *IncidentDebriefQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(idgb.fns))
	for _, fn := range idgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*idgb.flds)+len(idgb.fns))
		for _, f := range *idgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*idgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := idgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// IncidentDebriefSelect is the builder for selecting fields of IncidentDebrief entities.
type IncidentDebriefSelect struct {
	*IncidentDebriefQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (ids *IncidentDebriefSelect) Aggregate(fns ...AggregateFunc) *IncidentDebriefSelect {
	ids.fns = append(ids.fns, fns...)
	return ids
}

// Scan applies the selector query and scans the result into the given value.
func (ids *IncidentDebriefSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, ids.ctx, ent.OpQuerySelect)
	if err := ids.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*IncidentDebriefQuery, *IncidentDebriefSelect](ctx, ids.IncidentDebriefQuery, ids, ids.inters, v)
}

func (ids *IncidentDebriefSelect) sqlScan(ctx context.Context, root *IncidentDebriefQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(ids.fns))
	for _, fn := range ids.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*ids.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := ids.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// Modify adds a query modifier for attaching custom logic to queries.
func (ids *IncidentDebriefSelect) Modify(modifiers ...func(s *sql.Selector)) *IncidentDebriefSelect {
	ids.modifiers = append(ids.modifiers, modifiers...)
	return ids
}
