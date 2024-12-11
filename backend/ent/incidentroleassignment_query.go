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
	"github.com/rezible/rezible/ent/incidentrole"
	"github.com/rezible/rezible/ent/incidentroleassignment"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/user"
)

// IncidentRoleAssignmentQuery is the builder for querying IncidentRoleAssignment entities.
type IncidentRoleAssignmentQuery struct {
	config
	ctx          *QueryContext
	order        []incidentroleassignment.OrderOption
	inters       []Interceptor
	predicates   []predicate.IncidentRoleAssignment
	withRole     *IncidentRoleQuery
	withIncident *IncidentQuery
	withUser     *UserQuery
	modifiers    []func(*sql.Selector)
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the IncidentRoleAssignmentQuery builder.
func (iraq *IncidentRoleAssignmentQuery) Where(ps ...predicate.IncidentRoleAssignment) *IncidentRoleAssignmentQuery {
	iraq.predicates = append(iraq.predicates, ps...)
	return iraq
}

// Limit the number of records to be returned by this query.
func (iraq *IncidentRoleAssignmentQuery) Limit(limit int) *IncidentRoleAssignmentQuery {
	iraq.ctx.Limit = &limit
	return iraq
}

// Offset to start from.
func (iraq *IncidentRoleAssignmentQuery) Offset(offset int) *IncidentRoleAssignmentQuery {
	iraq.ctx.Offset = &offset
	return iraq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (iraq *IncidentRoleAssignmentQuery) Unique(unique bool) *IncidentRoleAssignmentQuery {
	iraq.ctx.Unique = &unique
	return iraq
}

// Order specifies how the records should be ordered.
func (iraq *IncidentRoleAssignmentQuery) Order(o ...incidentroleassignment.OrderOption) *IncidentRoleAssignmentQuery {
	iraq.order = append(iraq.order, o...)
	return iraq
}

// QueryRole chains the current query on the "role" edge.
func (iraq *IncidentRoleAssignmentQuery) QueryRole() *IncidentRoleQuery {
	query := (&IncidentRoleClient{config: iraq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := iraq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := iraq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(incidentroleassignment.Table, incidentroleassignment.FieldID, selector),
			sqlgraph.To(incidentrole.Table, incidentrole.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, incidentroleassignment.RoleTable, incidentroleassignment.RoleColumn),
		)
		fromU = sqlgraph.SetNeighbors(iraq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryIncident chains the current query on the "incident" edge.
func (iraq *IncidentRoleAssignmentQuery) QueryIncident() *IncidentQuery {
	query := (&IncidentClient{config: iraq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := iraq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := iraq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(incidentroleassignment.Table, incidentroleassignment.FieldID, selector),
			sqlgraph.To(incident.Table, incident.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, incidentroleassignment.IncidentTable, incidentroleassignment.IncidentColumn),
		)
		fromU = sqlgraph.SetNeighbors(iraq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryUser chains the current query on the "user" edge.
func (iraq *IncidentRoleAssignmentQuery) QueryUser() *UserQuery {
	query := (&UserClient{config: iraq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := iraq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := iraq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(incidentroleassignment.Table, incidentroleassignment.FieldID, selector),
			sqlgraph.To(user.Table, user.FieldID),
			sqlgraph.Edge(sqlgraph.M2O, false, incidentroleassignment.UserTable, incidentroleassignment.UserColumn),
		)
		fromU = sqlgraph.SetNeighbors(iraq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first IncidentRoleAssignment entity from the query.
// Returns a *NotFoundError when no IncidentRoleAssignment was found.
func (iraq *IncidentRoleAssignmentQuery) First(ctx context.Context) (*IncidentRoleAssignment, error) {
	nodes, err := iraq.Limit(1).All(setContextOp(ctx, iraq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{incidentroleassignment.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (iraq *IncidentRoleAssignmentQuery) FirstX(ctx context.Context) *IncidentRoleAssignment {
	node, err := iraq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first IncidentRoleAssignment ID from the query.
// Returns a *NotFoundError when no IncidentRoleAssignment ID was found.
func (iraq *IncidentRoleAssignmentQuery) FirstID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = iraq.Limit(1).IDs(setContextOp(ctx, iraq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{incidentroleassignment.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (iraq *IncidentRoleAssignmentQuery) FirstIDX(ctx context.Context) uuid.UUID {
	id, err := iraq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single IncidentRoleAssignment entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one IncidentRoleAssignment entity is found.
// Returns a *NotFoundError when no IncidentRoleAssignment entities are found.
func (iraq *IncidentRoleAssignmentQuery) Only(ctx context.Context) (*IncidentRoleAssignment, error) {
	nodes, err := iraq.Limit(2).All(setContextOp(ctx, iraq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{incidentroleassignment.Label}
	default:
		return nil, &NotSingularError{incidentroleassignment.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (iraq *IncidentRoleAssignmentQuery) OnlyX(ctx context.Context) *IncidentRoleAssignment {
	node, err := iraq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only IncidentRoleAssignment ID in the query.
// Returns a *NotSingularError when more than one IncidentRoleAssignment ID is found.
// Returns a *NotFoundError when no entities are found.
func (iraq *IncidentRoleAssignmentQuery) OnlyID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = iraq.Limit(2).IDs(setContextOp(ctx, iraq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{incidentroleassignment.Label}
	default:
		err = &NotSingularError{incidentroleassignment.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (iraq *IncidentRoleAssignmentQuery) OnlyIDX(ctx context.Context) uuid.UUID {
	id, err := iraq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of IncidentRoleAssignments.
func (iraq *IncidentRoleAssignmentQuery) All(ctx context.Context) ([]*IncidentRoleAssignment, error) {
	ctx = setContextOp(ctx, iraq.ctx, ent.OpQueryAll)
	if err := iraq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*IncidentRoleAssignment, *IncidentRoleAssignmentQuery]()
	return withInterceptors[[]*IncidentRoleAssignment](ctx, iraq, qr, iraq.inters)
}

// AllX is like All, but panics if an error occurs.
func (iraq *IncidentRoleAssignmentQuery) AllX(ctx context.Context) []*IncidentRoleAssignment {
	nodes, err := iraq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of IncidentRoleAssignment IDs.
func (iraq *IncidentRoleAssignmentQuery) IDs(ctx context.Context) (ids []uuid.UUID, err error) {
	if iraq.ctx.Unique == nil && iraq.path != nil {
		iraq.Unique(true)
	}
	ctx = setContextOp(ctx, iraq.ctx, ent.OpQueryIDs)
	if err = iraq.Select(incidentroleassignment.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (iraq *IncidentRoleAssignmentQuery) IDsX(ctx context.Context) []uuid.UUID {
	ids, err := iraq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (iraq *IncidentRoleAssignmentQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, iraq.ctx, ent.OpQueryCount)
	if err := iraq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, iraq, querierCount[*IncidentRoleAssignmentQuery](), iraq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (iraq *IncidentRoleAssignmentQuery) CountX(ctx context.Context) int {
	count, err := iraq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (iraq *IncidentRoleAssignmentQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, iraq.ctx, ent.OpQueryExist)
	switch _, err := iraq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (iraq *IncidentRoleAssignmentQuery) ExistX(ctx context.Context) bool {
	exist, err := iraq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the IncidentRoleAssignmentQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (iraq *IncidentRoleAssignmentQuery) Clone() *IncidentRoleAssignmentQuery {
	if iraq == nil {
		return nil
	}
	return &IncidentRoleAssignmentQuery{
		config:       iraq.config,
		ctx:          iraq.ctx.Clone(),
		order:        append([]incidentroleassignment.OrderOption{}, iraq.order...),
		inters:       append([]Interceptor{}, iraq.inters...),
		predicates:   append([]predicate.IncidentRoleAssignment{}, iraq.predicates...),
		withRole:     iraq.withRole.Clone(),
		withIncident: iraq.withIncident.Clone(),
		withUser:     iraq.withUser.Clone(),
		// clone intermediate query.
		sql:       iraq.sql.Clone(),
		path:      iraq.path,
		modifiers: append([]func(*sql.Selector){}, iraq.modifiers...),
	}
}

// WithRole tells the query-builder to eager-load the nodes that are connected to
// the "role" edge. The optional arguments are used to configure the query builder of the edge.
func (iraq *IncidentRoleAssignmentQuery) WithRole(opts ...func(*IncidentRoleQuery)) *IncidentRoleAssignmentQuery {
	query := (&IncidentRoleClient{config: iraq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	iraq.withRole = query
	return iraq
}

// WithIncident tells the query-builder to eager-load the nodes that are connected to
// the "incident" edge. The optional arguments are used to configure the query builder of the edge.
func (iraq *IncidentRoleAssignmentQuery) WithIncident(opts ...func(*IncidentQuery)) *IncidentRoleAssignmentQuery {
	query := (&IncidentClient{config: iraq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	iraq.withIncident = query
	return iraq
}

// WithUser tells the query-builder to eager-load the nodes that are connected to
// the "user" edge. The optional arguments are used to configure the query builder of the edge.
func (iraq *IncidentRoleAssignmentQuery) WithUser(opts ...func(*UserQuery)) *IncidentRoleAssignmentQuery {
	query := (&UserClient{config: iraq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	iraq.withUser = query
	return iraq
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		RoleID uuid.UUID `json:"role_id,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.IncidentRoleAssignment.Query().
//		GroupBy(incidentroleassignment.FieldRoleID).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (iraq *IncidentRoleAssignmentQuery) GroupBy(field string, fields ...string) *IncidentRoleAssignmentGroupBy {
	iraq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &IncidentRoleAssignmentGroupBy{build: iraq}
	grbuild.flds = &iraq.ctx.Fields
	grbuild.label = incidentroleassignment.Label
	grbuild.scan = grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		RoleID uuid.UUID `json:"role_id,omitempty"`
//	}
//
//	client.IncidentRoleAssignment.Query().
//		Select(incidentroleassignment.FieldRoleID).
//		Scan(ctx, &v)
func (iraq *IncidentRoleAssignmentQuery) Select(fields ...string) *IncidentRoleAssignmentSelect {
	iraq.ctx.Fields = append(iraq.ctx.Fields, fields...)
	sbuild := &IncidentRoleAssignmentSelect{IncidentRoleAssignmentQuery: iraq}
	sbuild.label = incidentroleassignment.Label
	sbuild.flds, sbuild.scan = &iraq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a IncidentRoleAssignmentSelect configured with the given aggregations.
func (iraq *IncidentRoleAssignmentQuery) Aggregate(fns ...AggregateFunc) *IncidentRoleAssignmentSelect {
	return iraq.Select().Aggregate(fns...)
}

func (iraq *IncidentRoleAssignmentQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range iraq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, iraq); err != nil {
				return err
			}
		}
	}
	for _, f := range iraq.ctx.Fields {
		if !incidentroleassignment.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if iraq.path != nil {
		prev, err := iraq.path(ctx)
		if err != nil {
			return err
		}
		iraq.sql = prev
	}
	return nil
}

func (iraq *IncidentRoleAssignmentQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*IncidentRoleAssignment, error) {
	var (
		nodes       = []*IncidentRoleAssignment{}
		_spec       = iraq.querySpec()
		loadedTypes = [3]bool{
			iraq.withRole != nil,
			iraq.withIncident != nil,
			iraq.withUser != nil,
		}
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*IncidentRoleAssignment).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &IncidentRoleAssignment{config: iraq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	if len(iraq.modifiers) > 0 {
		_spec.Modifiers = iraq.modifiers
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, iraq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := iraq.withRole; query != nil {
		if err := iraq.loadRole(ctx, query, nodes, nil,
			func(n *IncidentRoleAssignment, e *IncidentRole) { n.Edges.Role = e }); err != nil {
			return nil, err
		}
	}
	if query := iraq.withIncident; query != nil {
		if err := iraq.loadIncident(ctx, query, nodes, nil,
			func(n *IncidentRoleAssignment, e *Incident) { n.Edges.Incident = e }); err != nil {
			return nil, err
		}
	}
	if query := iraq.withUser; query != nil {
		if err := iraq.loadUser(ctx, query, nodes, nil,
			func(n *IncidentRoleAssignment, e *User) { n.Edges.User = e }); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (iraq *IncidentRoleAssignmentQuery) loadRole(ctx context.Context, query *IncidentRoleQuery, nodes []*IncidentRoleAssignment, init func(*IncidentRoleAssignment), assign func(*IncidentRoleAssignment, *IncidentRole)) error {
	ids := make([]uuid.UUID, 0, len(nodes))
	nodeids := make(map[uuid.UUID][]*IncidentRoleAssignment)
	for i := range nodes {
		fk := nodes[i].RoleID
		if _, ok := nodeids[fk]; !ok {
			ids = append(ids, fk)
		}
		nodeids[fk] = append(nodeids[fk], nodes[i])
	}
	if len(ids) == 0 {
		return nil
	}
	query.Where(incidentrole.IDIn(ids...))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		nodes, ok := nodeids[n.ID]
		if !ok {
			return fmt.Errorf(`unexpected foreign-key "role_id" returned %v`, n.ID)
		}
		for i := range nodes {
			assign(nodes[i], n)
		}
	}
	return nil
}
func (iraq *IncidentRoleAssignmentQuery) loadIncident(ctx context.Context, query *IncidentQuery, nodes []*IncidentRoleAssignment, init func(*IncidentRoleAssignment), assign func(*IncidentRoleAssignment, *Incident)) error {
	ids := make([]uuid.UUID, 0, len(nodes))
	nodeids := make(map[uuid.UUID][]*IncidentRoleAssignment)
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
func (iraq *IncidentRoleAssignmentQuery) loadUser(ctx context.Context, query *UserQuery, nodes []*IncidentRoleAssignment, init func(*IncidentRoleAssignment), assign func(*IncidentRoleAssignment, *User)) error {
	ids := make([]uuid.UUID, 0, len(nodes))
	nodeids := make(map[uuid.UUID][]*IncidentRoleAssignment)
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

func (iraq *IncidentRoleAssignmentQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := iraq.querySpec()
	if len(iraq.modifiers) > 0 {
		_spec.Modifiers = iraq.modifiers
	}
	_spec.Node.Columns = iraq.ctx.Fields
	if len(iraq.ctx.Fields) > 0 {
		_spec.Unique = iraq.ctx.Unique != nil && *iraq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, iraq.driver, _spec)
}

func (iraq *IncidentRoleAssignmentQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(incidentroleassignment.Table, incidentroleassignment.Columns, sqlgraph.NewFieldSpec(incidentroleassignment.FieldID, field.TypeUUID))
	_spec.From = iraq.sql
	if unique := iraq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if iraq.path != nil {
		_spec.Unique = true
	}
	if fields := iraq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, incidentroleassignment.FieldID)
		for i := range fields {
			if fields[i] != incidentroleassignment.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
		if iraq.withRole != nil {
			_spec.Node.AddColumnOnce(incidentroleassignment.FieldRoleID)
		}
		if iraq.withIncident != nil {
			_spec.Node.AddColumnOnce(incidentroleassignment.FieldIncidentID)
		}
		if iraq.withUser != nil {
			_spec.Node.AddColumnOnce(incidentroleassignment.FieldUserID)
		}
	}
	if ps := iraq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := iraq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := iraq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := iraq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (iraq *IncidentRoleAssignmentQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(iraq.driver.Dialect())
	t1 := builder.Table(incidentroleassignment.Table)
	columns := iraq.ctx.Fields
	if len(columns) == 0 {
		columns = incidentroleassignment.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if iraq.sql != nil {
		selector = iraq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if iraq.ctx.Unique != nil && *iraq.ctx.Unique {
		selector.Distinct()
	}
	for _, m := range iraq.modifiers {
		m(selector)
	}
	for _, p := range iraq.predicates {
		p(selector)
	}
	for _, p := range iraq.order {
		p(selector)
	}
	if offset := iraq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := iraq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// Modify adds a query modifier for attaching custom logic to queries.
func (iraq *IncidentRoleAssignmentQuery) Modify(modifiers ...func(s *sql.Selector)) *IncidentRoleAssignmentSelect {
	iraq.modifiers = append(iraq.modifiers, modifiers...)
	return iraq.Select()
}

// IncidentRoleAssignmentGroupBy is the group-by builder for IncidentRoleAssignment entities.
type IncidentRoleAssignmentGroupBy struct {
	selector
	build *IncidentRoleAssignmentQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (iragb *IncidentRoleAssignmentGroupBy) Aggregate(fns ...AggregateFunc) *IncidentRoleAssignmentGroupBy {
	iragb.fns = append(iragb.fns, fns...)
	return iragb
}

// Scan applies the selector query and scans the result into the given value.
func (iragb *IncidentRoleAssignmentGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, iragb.build.ctx, ent.OpQueryGroupBy)
	if err := iragb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*IncidentRoleAssignmentQuery, *IncidentRoleAssignmentGroupBy](ctx, iragb.build, iragb, iragb.build.inters, v)
}

func (iragb *IncidentRoleAssignmentGroupBy) sqlScan(ctx context.Context, root *IncidentRoleAssignmentQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(iragb.fns))
	for _, fn := range iragb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*iragb.flds)+len(iragb.fns))
		for _, f := range *iragb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*iragb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := iragb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// IncidentRoleAssignmentSelect is the builder for selecting fields of IncidentRoleAssignment entities.
type IncidentRoleAssignmentSelect struct {
	*IncidentRoleAssignmentQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (iras *IncidentRoleAssignmentSelect) Aggregate(fns ...AggregateFunc) *IncidentRoleAssignmentSelect {
	iras.fns = append(iras.fns, fns...)
	return iras
}

// Scan applies the selector query and scans the result into the given value.
func (iras *IncidentRoleAssignmentSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, iras.ctx, ent.OpQuerySelect)
	if err := iras.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*IncidentRoleAssignmentQuery, *IncidentRoleAssignmentSelect](ctx, iras.IncidentRoleAssignmentQuery, iras, iras.inters, v)
}

func (iras *IncidentRoleAssignmentSelect) sqlScan(ctx context.Context, root *IncidentRoleAssignmentQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(iras.fns))
	for _, fn := range iras.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*iras.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := iras.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// Modify adds a query modifier for attaching custom logic to queries.
func (iras *IncidentRoleAssignmentSelect) Modify(modifiers ...func(s *sql.Selector)) *IncidentRoleAssignmentSelect {
	iras.modifiers = append(iras.modifiers, modifiers...)
	return iras
}
