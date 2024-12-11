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
	"github.com/twohundreds/rezible/ent/incidentdebriefquestion"
	"github.com/twohundreds/rezible/ent/incidentrole"
	"github.com/twohundreds/rezible/ent/incidentroleassignment"
	"github.com/twohundreds/rezible/ent/predicate"
)

// IncidentRoleQuery is the builder for querying IncidentRole entities.
type IncidentRoleQuery struct {
	config
	ctx                  *QueryContext
	order                []incidentrole.OrderOption
	inters               []Interceptor
	predicates           []predicate.IncidentRole
	withAssignments      *IncidentRoleAssignmentQuery
	withDebriefQuestions *IncidentDebriefQuestionQuery
	modifiers            []func(*sql.Selector)
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the IncidentRoleQuery builder.
func (irq *IncidentRoleQuery) Where(ps ...predicate.IncidentRole) *IncidentRoleQuery {
	irq.predicates = append(irq.predicates, ps...)
	return irq
}

// Limit the number of records to be returned by this query.
func (irq *IncidentRoleQuery) Limit(limit int) *IncidentRoleQuery {
	irq.ctx.Limit = &limit
	return irq
}

// Offset to start from.
func (irq *IncidentRoleQuery) Offset(offset int) *IncidentRoleQuery {
	irq.ctx.Offset = &offset
	return irq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (irq *IncidentRoleQuery) Unique(unique bool) *IncidentRoleQuery {
	irq.ctx.Unique = &unique
	return irq
}

// Order specifies how the records should be ordered.
func (irq *IncidentRoleQuery) Order(o ...incidentrole.OrderOption) *IncidentRoleQuery {
	irq.order = append(irq.order, o...)
	return irq
}

// QueryAssignments chains the current query on the "assignments" edge.
func (irq *IncidentRoleQuery) QueryAssignments() *IncidentRoleAssignmentQuery {
	query := (&IncidentRoleAssignmentClient{config: irq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := irq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := irq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(incidentrole.Table, incidentrole.FieldID, selector),
			sqlgraph.To(incidentroleassignment.Table, incidentroleassignment.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, true, incidentrole.AssignmentsTable, incidentrole.AssignmentsColumn),
		)
		fromU = sqlgraph.SetNeighbors(irq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryDebriefQuestions chains the current query on the "debrief_questions" edge.
func (irq *IncidentRoleQuery) QueryDebriefQuestions() *IncidentDebriefQuestionQuery {
	query := (&IncidentDebriefQuestionClient{config: irq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := irq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := irq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(incidentrole.Table, incidentrole.FieldID, selector),
			sqlgraph.To(incidentdebriefquestion.Table, incidentdebriefquestion.FieldID),
			sqlgraph.Edge(sqlgraph.M2M, true, incidentrole.DebriefQuestionsTable, incidentrole.DebriefQuestionsPrimaryKey...),
		)
		fromU = sqlgraph.SetNeighbors(irq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first IncidentRole entity from the query.
// Returns a *NotFoundError when no IncidentRole was found.
func (irq *IncidentRoleQuery) First(ctx context.Context) (*IncidentRole, error) {
	nodes, err := irq.Limit(1).All(setContextOp(ctx, irq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{incidentrole.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (irq *IncidentRoleQuery) FirstX(ctx context.Context) *IncidentRole {
	node, err := irq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first IncidentRole ID from the query.
// Returns a *NotFoundError when no IncidentRole ID was found.
func (irq *IncidentRoleQuery) FirstID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = irq.Limit(1).IDs(setContextOp(ctx, irq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{incidentrole.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (irq *IncidentRoleQuery) FirstIDX(ctx context.Context) uuid.UUID {
	id, err := irq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single IncidentRole entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one IncidentRole entity is found.
// Returns a *NotFoundError when no IncidentRole entities are found.
func (irq *IncidentRoleQuery) Only(ctx context.Context) (*IncidentRole, error) {
	nodes, err := irq.Limit(2).All(setContextOp(ctx, irq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{incidentrole.Label}
	default:
		return nil, &NotSingularError{incidentrole.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (irq *IncidentRoleQuery) OnlyX(ctx context.Context) *IncidentRole {
	node, err := irq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only IncidentRole ID in the query.
// Returns a *NotSingularError when more than one IncidentRole ID is found.
// Returns a *NotFoundError when no entities are found.
func (irq *IncidentRoleQuery) OnlyID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = irq.Limit(2).IDs(setContextOp(ctx, irq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{incidentrole.Label}
	default:
		err = &NotSingularError{incidentrole.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (irq *IncidentRoleQuery) OnlyIDX(ctx context.Context) uuid.UUID {
	id, err := irq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of IncidentRoles.
func (irq *IncidentRoleQuery) All(ctx context.Context) ([]*IncidentRole, error) {
	ctx = setContextOp(ctx, irq.ctx, ent.OpQueryAll)
	if err := irq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*IncidentRole, *IncidentRoleQuery]()
	return withInterceptors[[]*IncidentRole](ctx, irq, qr, irq.inters)
}

// AllX is like All, but panics if an error occurs.
func (irq *IncidentRoleQuery) AllX(ctx context.Context) []*IncidentRole {
	nodes, err := irq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of IncidentRole IDs.
func (irq *IncidentRoleQuery) IDs(ctx context.Context) (ids []uuid.UUID, err error) {
	if irq.ctx.Unique == nil && irq.path != nil {
		irq.Unique(true)
	}
	ctx = setContextOp(ctx, irq.ctx, ent.OpQueryIDs)
	if err = irq.Select(incidentrole.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (irq *IncidentRoleQuery) IDsX(ctx context.Context) []uuid.UUID {
	ids, err := irq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (irq *IncidentRoleQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, irq.ctx, ent.OpQueryCount)
	if err := irq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, irq, querierCount[*IncidentRoleQuery](), irq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (irq *IncidentRoleQuery) CountX(ctx context.Context) int {
	count, err := irq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (irq *IncidentRoleQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, irq.ctx, ent.OpQueryExist)
	switch _, err := irq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (irq *IncidentRoleQuery) ExistX(ctx context.Context) bool {
	exist, err := irq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the IncidentRoleQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (irq *IncidentRoleQuery) Clone() *IncidentRoleQuery {
	if irq == nil {
		return nil
	}
	return &IncidentRoleQuery{
		config:               irq.config,
		ctx:                  irq.ctx.Clone(),
		order:                append([]incidentrole.OrderOption{}, irq.order...),
		inters:               append([]Interceptor{}, irq.inters...),
		predicates:           append([]predicate.IncidentRole{}, irq.predicates...),
		withAssignments:      irq.withAssignments.Clone(),
		withDebriefQuestions: irq.withDebriefQuestions.Clone(),
		// clone intermediate query.
		sql:       irq.sql.Clone(),
		path:      irq.path,
		modifiers: append([]func(*sql.Selector){}, irq.modifiers...),
	}
}

// WithAssignments tells the query-builder to eager-load the nodes that are connected to
// the "assignments" edge. The optional arguments are used to configure the query builder of the edge.
func (irq *IncidentRoleQuery) WithAssignments(opts ...func(*IncidentRoleAssignmentQuery)) *IncidentRoleQuery {
	query := (&IncidentRoleAssignmentClient{config: irq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	irq.withAssignments = query
	return irq
}

// WithDebriefQuestions tells the query-builder to eager-load the nodes that are connected to
// the "debrief_questions" edge. The optional arguments are used to configure the query builder of the edge.
func (irq *IncidentRoleQuery) WithDebriefQuestions(opts ...func(*IncidentDebriefQuestionQuery)) *IncidentRoleQuery {
	query := (&IncidentDebriefQuestionClient{config: irq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	irq.withDebriefQuestions = query
	return irq
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
//	client.IncidentRole.Query().
//		GroupBy(incidentrole.FieldArchiveTime).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (irq *IncidentRoleQuery) GroupBy(field string, fields ...string) *IncidentRoleGroupBy {
	irq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &IncidentRoleGroupBy{build: irq}
	grbuild.flds = &irq.ctx.Fields
	grbuild.label = incidentrole.Label
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
//	client.IncidentRole.Query().
//		Select(incidentrole.FieldArchiveTime).
//		Scan(ctx, &v)
func (irq *IncidentRoleQuery) Select(fields ...string) *IncidentRoleSelect {
	irq.ctx.Fields = append(irq.ctx.Fields, fields...)
	sbuild := &IncidentRoleSelect{IncidentRoleQuery: irq}
	sbuild.label = incidentrole.Label
	sbuild.flds, sbuild.scan = &irq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a IncidentRoleSelect configured with the given aggregations.
func (irq *IncidentRoleQuery) Aggregate(fns ...AggregateFunc) *IncidentRoleSelect {
	return irq.Select().Aggregate(fns...)
}

func (irq *IncidentRoleQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range irq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, irq); err != nil {
				return err
			}
		}
	}
	for _, f := range irq.ctx.Fields {
		if !incidentrole.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if irq.path != nil {
		prev, err := irq.path(ctx)
		if err != nil {
			return err
		}
		irq.sql = prev
	}
	return nil
}

func (irq *IncidentRoleQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*IncidentRole, error) {
	var (
		nodes       = []*IncidentRole{}
		_spec       = irq.querySpec()
		loadedTypes = [2]bool{
			irq.withAssignments != nil,
			irq.withDebriefQuestions != nil,
		}
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*IncidentRole).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &IncidentRole{config: irq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	if len(irq.modifiers) > 0 {
		_spec.Modifiers = irq.modifiers
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, irq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := irq.withAssignments; query != nil {
		if err := irq.loadAssignments(ctx, query, nodes,
			func(n *IncidentRole) { n.Edges.Assignments = []*IncidentRoleAssignment{} },
			func(n *IncidentRole, e *IncidentRoleAssignment) { n.Edges.Assignments = append(n.Edges.Assignments, e) }); err != nil {
			return nil, err
		}
	}
	if query := irq.withDebriefQuestions; query != nil {
		if err := irq.loadDebriefQuestions(ctx, query, nodes,
			func(n *IncidentRole) { n.Edges.DebriefQuestions = []*IncidentDebriefQuestion{} },
			func(n *IncidentRole, e *IncidentDebriefQuestion) {
				n.Edges.DebriefQuestions = append(n.Edges.DebriefQuestions, e)
			}); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (irq *IncidentRoleQuery) loadAssignments(ctx context.Context, query *IncidentRoleAssignmentQuery, nodes []*IncidentRole, init func(*IncidentRole), assign func(*IncidentRole, *IncidentRoleAssignment)) error {
	fks := make([]driver.Value, 0, len(nodes))
	nodeids := make(map[uuid.UUID]*IncidentRole)
	for i := range nodes {
		fks = append(fks, nodes[i].ID)
		nodeids[nodes[i].ID] = nodes[i]
		if init != nil {
			init(nodes[i])
		}
	}
	if len(query.ctx.Fields) > 0 {
		query.ctx.AppendFieldOnce(incidentroleassignment.FieldRoleID)
	}
	query.Where(predicate.IncidentRoleAssignment(func(s *sql.Selector) {
		s.Where(sql.InValues(s.C(incidentrole.AssignmentsColumn), fks...))
	}))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		fk := n.RoleID
		node, ok := nodeids[fk]
		if !ok {
			return fmt.Errorf(`unexpected referenced foreign-key "role_id" returned %v for node %v`, fk, n.ID)
		}
		assign(node, n)
	}
	return nil
}
func (irq *IncidentRoleQuery) loadDebriefQuestions(ctx context.Context, query *IncidentDebriefQuestionQuery, nodes []*IncidentRole, init func(*IncidentRole), assign func(*IncidentRole, *IncidentDebriefQuestion)) error {
	edgeIDs := make([]driver.Value, len(nodes))
	byID := make(map[uuid.UUID]*IncidentRole)
	nids := make(map[uuid.UUID]map[*IncidentRole]struct{})
	for i, node := range nodes {
		edgeIDs[i] = node.ID
		byID[node.ID] = node
		if init != nil {
			init(node)
		}
	}
	query.Where(func(s *sql.Selector) {
		joinT := sql.Table(incidentrole.DebriefQuestionsTable)
		s.Join(joinT).On(s.C(incidentdebriefquestion.FieldID), joinT.C(incidentrole.DebriefQuestionsPrimaryKey[0]))
		s.Where(sql.InValues(joinT.C(incidentrole.DebriefQuestionsPrimaryKey[1]), edgeIDs...))
		columns := s.SelectedColumns()
		s.Select(joinT.C(incidentrole.DebriefQuestionsPrimaryKey[1]))
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
					nids[inValue] = map[*IncidentRole]struct{}{byID[outValue]: {}}
					return assign(columns[1:], values[1:])
				}
				nids[inValue][byID[outValue]] = struct{}{}
				return nil
			}
		})
	})
	neighbors, err := withInterceptors[[]*IncidentDebriefQuestion](ctx, query, qr, query.inters)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		nodes, ok := nids[n.ID]
		if !ok {
			return fmt.Errorf(`unexpected "debrief_questions" node returned %v`, n.ID)
		}
		for kn := range nodes {
			assign(kn, n)
		}
	}
	return nil
}

func (irq *IncidentRoleQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := irq.querySpec()
	if len(irq.modifiers) > 0 {
		_spec.Modifiers = irq.modifiers
	}
	_spec.Node.Columns = irq.ctx.Fields
	if len(irq.ctx.Fields) > 0 {
		_spec.Unique = irq.ctx.Unique != nil && *irq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, irq.driver, _spec)
}

func (irq *IncidentRoleQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(incidentrole.Table, incidentrole.Columns, sqlgraph.NewFieldSpec(incidentrole.FieldID, field.TypeUUID))
	_spec.From = irq.sql
	if unique := irq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if irq.path != nil {
		_spec.Unique = true
	}
	if fields := irq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, incidentrole.FieldID)
		for i := range fields {
			if fields[i] != incidentrole.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := irq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := irq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := irq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := irq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (irq *IncidentRoleQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(irq.driver.Dialect())
	t1 := builder.Table(incidentrole.Table)
	columns := irq.ctx.Fields
	if len(columns) == 0 {
		columns = incidentrole.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if irq.sql != nil {
		selector = irq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if irq.ctx.Unique != nil && *irq.ctx.Unique {
		selector.Distinct()
	}
	for _, m := range irq.modifiers {
		m(selector)
	}
	for _, p := range irq.predicates {
		p(selector)
	}
	for _, p := range irq.order {
		p(selector)
	}
	if offset := irq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := irq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// Modify adds a query modifier for attaching custom logic to queries.
func (irq *IncidentRoleQuery) Modify(modifiers ...func(s *sql.Selector)) *IncidentRoleSelect {
	irq.modifiers = append(irq.modifiers, modifiers...)
	return irq.Select()
}

// IncidentRoleGroupBy is the group-by builder for IncidentRole entities.
type IncidentRoleGroupBy struct {
	selector
	build *IncidentRoleQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (irgb *IncidentRoleGroupBy) Aggregate(fns ...AggregateFunc) *IncidentRoleGroupBy {
	irgb.fns = append(irgb.fns, fns...)
	return irgb
}

// Scan applies the selector query and scans the result into the given value.
func (irgb *IncidentRoleGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, irgb.build.ctx, ent.OpQueryGroupBy)
	if err := irgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*IncidentRoleQuery, *IncidentRoleGroupBy](ctx, irgb.build, irgb, irgb.build.inters, v)
}

func (irgb *IncidentRoleGroupBy) sqlScan(ctx context.Context, root *IncidentRoleQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(irgb.fns))
	for _, fn := range irgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*irgb.flds)+len(irgb.fns))
		for _, f := range *irgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*irgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := irgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// IncidentRoleSelect is the builder for selecting fields of IncidentRole entities.
type IncidentRoleSelect struct {
	*IncidentRoleQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (irs *IncidentRoleSelect) Aggregate(fns ...AggregateFunc) *IncidentRoleSelect {
	irs.fns = append(irs.fns, fns...)
	return irs
}

// Scan applies the selector query and scans the result into the given value.
func (irs *IncidentRoleSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, irs.ctx, ent.OpQuerySelect)
	if err := irs.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*IncidentRoleQuery, *IncidentRoleSelect](ctx, irs.IncidentRoleQuery, irs, irs.inters, v)
}

func (irs *IncidentRoleSelect) sqlScan(ctx context.Context, root *IncidentRoleQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(irs.fns))
	for _, fn := range irs.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*irs.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := irs.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// Modify adds a query modifier for attaching custom logic to queries.
func (irs *IncidentRoleSelect) Modify(modifiers ...func(s *sql.Selector)) *IncidentRoleSelect {
	irs.modifiers = append(irs.modifiers, modifiers...)
	return irs
}