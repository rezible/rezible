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
	"github.com/twohundreds/rezible/ent/incidentfield"
	"github.com/twohundreds/rezible/ent/incidentfieldoption"
	"github.com/twohundreds/rezible/ent/predicate"
)

// IncidentFieldQuery is the builder for querying IncidentField entities.
type IncidentFieldQuery struct {
	config
	ctx                  *QueryContext
	order                []incidentfield.OrderOption
	inters               []Interceptor
	predicates           []predicate.IncidentField
	withOptions          *IncidentFieldOptionQuery
	withDebriefQuestions *IncidentDebriefQuestionQuery
	modifiers            []func(*sql.Selector)
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the IncidentFieldQuery builder.
func (ifq *IncidentFieldQuery) Where(ps ...predicate.IncidentField) *IncidentFieldQuery {
	ifq.predicates = append(ifq.predicates, ps...)
	return ifq
}

// Limit the number of records to be returned by this query.
func (ifq *IncidentFieldQuery) Limit(limit int) *IncidentFieldQuery {
	ifq.ctx.Limit = &limit
	return ifq
}

// Offset to start from.
func (ifq *IncidentFieldQuery) Offset(offset int) *IncidentFieldQuery {
	ifq.ctx.Offset = &offset
	return ifq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (ifq *IncidentFieldQuery) Unique(unique bool) *IncidentFieldQuery {
	ifq.ctx.Unique = &unique
	return ifq
}

// Order specifies how the records should be ordered.
func (ifq *IncidentFieldQuery) Order(o ...incidentfield.OrderOption) *IncidentFieldQuery {
	ifq.order = append(ifq.order, o...)
	return ifq
}

// QueryOptions chains the current query on the "options" edge.
func (ifq *IncidentFieldQuery) QueryOptions() *IncidentFieldOptionQuery {
	query := (&IncidentFieldOptionClient{config: ifq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := ifq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := ifq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(incidentfield.Table, incidentfield.FieldID, selector),
			sqlgraph.To(incidentfieldoption.Table, incidentfieldoption.FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, incidentfield.OptionsTable, incidentfield.OptionsColumn),
		)
		fromU = sqlgraph.SetNeighbors(ifq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// QueryDebriefQuestions chains the current query on the "debrief_questions" edge.
func (ifq *IncidentFieldQuery) QueryDebriefQuestions() *IncidentDebriefQuestionQuery {
	query := (&IncidentDebriefQuestionClient{config: ifq.config}).Query()
	query.path = func(ctx context.Context) (fromU *sql.Selector, err error) {
		if err := ifq.prepareQuery(ctx); err != nil {
			return nil, err
		}
		selector := ifq.sqlQuery(ctx)
		if err := selector.Err(); err != nil {
			return nil, err
		}
		step := sqlgraph.NewStep(
			sqlgraph.From(incidentfield.Table, incidentfield.FieldID, selector),
			sqlgraph.To(incidentdebriefquestion.Table, incidentdebriefquestion.FieldID),
			sqlgraph.Edge(sqlgraph.M2M, true, incidentfield.DebriefQuestionsTable, incidentfield.DebriefQuestionsPrimaryKey...),
		)
		fromU = sqlgraph.SetNeighbors(ifq.driver.Dialect(), step)
		return fromU, nil
	}
	return query
}

// First returns the first IncidentField entity from the query.
// Returns a *NotFoundError when no IncidentField was found.
func (ifq *IncidentFieldQuery) First(ctx context.Context) (*IncidentField, error) {
	nodes, err := ifq.Limit(1).All(setContextOp(ctx, ifq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{incidentfield.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (ifq *IncidentFieldQuery) FirstX(ctx context.Context) *IncidentField {
	node, err := ifq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first IncidentField ID from the query.
// Returns a *NotFoundError when no IncidentField ID was found.
func (ifq *IncidentFieldQuery) FirstID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = ifq.Limit(1).IDs(setContextOp(ctx, ifq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{incidentfield.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (ifq *IncidentFieldQuery) FirstIDX(ctx context.Context) uuid.UUID {
	id, err := ifq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single IncidentField entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one IncidentField entity is found.
// Returns a *NotFoundError when no IncidentField entities are found.
func (ifq *IncidentFieldQuery) Only(ctx context.Context) (*IncidentField, error) {
	nodes, err := ifq.Limit(2).All(setContextOp(ctx, ifq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{incidentfield.Label}
	default:
		return nil, &NotSingularError{incidentfield.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (ifq *IncidentFieldQuery) OnlyX(ctx context.Context) *IncidentField {
	node, err := ifq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only IncidentField ID in the query.
// Returns a *NotSingularError when more than one IncidentField ID is found.
// Returns a *NotFoundError when no entities are found.
func (ifq *IncidentFieldQuery) OnlyID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = ifq.Limit(2).IDs(setContextOp(ctx, ifq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{incidentfield.Label}
	default:
		err = &NotSingularError{incidentfield.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (ifq *IncidentFieldQuery) OnlyIDX(ctx context.Context) uuid.UUID {
	id, err := ifq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of IncidentFields.
func (ifq *IncidentFieldQuery) All(ctx context.Context) ([]*IncidentField, error) {
	ctx = setContextOp(ctx, ifq.ctx, ent.OpQueryAll)
	if err := ifq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*IncidentField, *IncidentFieldQuery]()
	return withInterceptors[[]*IncidentField](ctx, ifq, qr, ifq.inters)
}

// AllX is like All, but panics if an error occurs.
func (ifq *IncidentFieldQuery) AllX(ctx context.Context) []*IncidentField {
	nodes, err := ifq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of IncidentField IDs.
func (ifq *IncidentFieldQuery) IDs(ctx context.Context) (ids []uuid.UUID, err error) {
	if ifq.ctx.Unique == nil && ifq.path != nil {
		ifq.Unique(true)
	}
	ctx = setContextOp(ctx, ifq.ctx, ent.OpQueryIDs)
	if err = ifq.Select(incidentfield.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (ifq *IncidentFieldQuery) IDsX(ctx context.Context) []uuid.UUID {
	ids, err := ifq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (ifq *IncidentFieldQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, ifq.ctx, ent.OpQueryCount)
	if err := ifq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, ifq, querierCount[*IncidentFieldQuery](), ifq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (ifq *IncidentFieldQuery) CountX(ctx context.Context) int {
	count, err := ifq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (ifq *IncidentFieldQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, ifq.ctx, ent.OpQueryExist)
	switch _, err := ifq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (ifq *IncidentFieldQuery) ExistX(ctx context.Context) bool {
	exist, err := ifq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the IncidentFieldQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (ifq *IncidentFieldQuery) Clone() *IncidentFieldQuery {
	if ifq == nil {
		return nil
	}
	return &IncidentFieldQuery{
		config:               ifq.config,
		ctx:                  ifq.ctx.Clone(),
		order:                append([]incidentfield.OrderOption{}, ifq.order...),
		inters:               append([]Interceptor{}, ifq.inters...),
		predicates:           append([]predicate.IncidentField{}, ifq.predicates...),
		withOptions:          ifq.withOptions.Clone(),
		withDebriefQuestions: ifq.withDebriefQuestions.Clone(),
		// clone intermediate query.
		sql:       ifq.sql.Clone(),
		path:      ifq.path,
		modifiers: append([]func(*sql.Selector){}, ifq.modifiers...),
	}
}

// WithOptions tells the query-builder to eager-load the nodes that are connected to
// the "options" edge. The optional arguments are used to configure the query builder of the edge.
func (ifq *IncidentFieldQuery) WithOptions(opts ...func(*IncidentFieldOptionQuery)) *IncidentFieldQuery {
	query := (&IncidentFieldOptionClient{config: ifq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	ifq.withOptions = query
	return ifq
}

// WithDebriefQuestions tells the query-builder to eager-load the nodes that are connected to
// the "debrief_questions" edge. The optional arguments are used to configure the query builder of the edge.
func (ifq *IncidentFieldQuery) WithDebriefQuestions(opts ...func(*IncidentDebriefQuestionQuery)) *IncidentFieldQuery {
	query := (&IncidentDebriefQuestionClient{config: ifq.config}).Query()
	for _, opt := range opts {
		opt(query)
	}
	ifq.withDebriefQuestions = query
	return ifq
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
//	client.IncidentField.Query().
//		GroupBy(incidentfield.FieldArchiveTime).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (ifq *IncidentFieldQuery) GroupBy(field string, fields ...string) *IncidentFieldGroupBy {
	ifq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &IncidentFieldGroupBy{build: ifq}
	grbuild.flds = &ifq.ctx.Fields
	grbuild.label = incidentfield.Label
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
//	client.IncidentField.Query().
//		Select(incidentfield.FieldArchiveTime).
//		Scan(ctx, &v)
func (ifq *IncidentFieldQuery) Select(fields ...string) *IncidentFieldSelect {
	ifq.ctx.Fields = append(ifq.ctx.Fields, fields...)
	sbuild := &IncidentFieldSelect{IncidentFieldQuery: ifq}
	sbuild.label = incidentfield.Label
	sbuild.flds, sbuild.scan = &ifq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a IncidentFieldSelect configured with the given aggregations.
func (ifq *IncidentFieldQuery) Aggregate(fns ...AggregateFunc) *IncidentFieldSelect {
	return ifq.Select().Aggregate(fns...)
}

func (ifq *IncidentFieldQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range ifq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, ifq); err != nil {
				return err
			}
		}
	}
	for _, f := range ifq.ctx.Fields {
		if !incidentfield.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if ifq.path != nil {
		prev, err := ifq.path(ctx)
		if err != nil {
			return err
		}
		ifq.sql = prev
	}
	return nil
}

func (ifq *IncidentFieldQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*IncidentField, error) {
	var (
		nodes       = []*IncidentField{}
		_spec       = ifq.querySpec()
		loadedTypes = [2]bool{
			ifq.withOptions != nil,
			ifq.withDebriefQuestions != nil,
		}
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*IncidentField).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &IncidentField{config: ifq.config}
		nodes = append(nodes, node)
		node.Edges.loadedTypes = loadedTypes
		return node.assignValues(columns, values)
	}
	if len(ifq.modifiers) > 0 {
		_spec.Modifiers = ifq.modifiers
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, ifq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	if query := ifq.withOptions; query != nil {
		if err := ifq.loadOptions(ctx, query, nodes,
			func(n *IncidentField) { n.Edges.Options = []*IncidentFieldOption{} },
			func(n *IncidentField, e *IncidentFieldOption) { n.Edges.Options = append(n.Edges.Options, e) }); err != nil {
			return nil, err
		}
	}
	if query := ifq.withDebriefQuestions; query != nil {
		if err := ifq.loadDebriefQuestions(ctx, query, nodes,
			func(n *IncidentField) { n.Edges.DebriefQuestions = []*IncidentDebriefQuestion{} },
			func(n *IncidentField, e *IncidentDebriefQuestion) {
				n.Edges.DebriefQuestions = append(n.Edges.DebriefQuestions, e)
			}); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

func (ifq *IncidentFieldQuery) loadOptions(ctx context.Context, query *IncidentFieldOptionQuery, nodes []*IncidentField, init func(*IncidentField), assign func(*IncidentField, *IncidentFieldOption)) error {
	fks := make([]driver.Value, 0, len(nodes))
	nodeids := make(map[uuid.UUID]*IncidentField)
	for i := range nodes {
		fks = append(fks, nodes[i].ID)
		nodeids[nodes[i].ID] = nodes[i]
		if init != nil {
			init(nodes[i])
		}
	}
	if len(query.ctx.Fields) > 0 {
		query.ctx.AppendFieldOnce(incidentfieldoption.FieldIncidentFieldID)
	}
	query.Where(predicate.IncidentFieldOption(func(s *sql.Selector) {
		s.Where(sql.InValues(s.C(incidentfield.OptionsColumn), fks...))
	}))
	neighbors, err := query.All(ctx)
	if err != nil {
		return err
	}
	for _, n := range neighbors {
		fk := n.IncidentFieldID
		node, ok := nodeids[fk]
		if !ok {
			return fmt.Errorf(`unexpected referenced foreign-key "incident_field_id" returned %v for node %v`, fk, n.ID)
		}
		assign(node, n)
	}
	return nil
}
func (ifq *IncidentFieldQuery) loadDebriefQuestions(ctx context.Context, query *IncidentDebriefQuestionQuery, nodes []*IncidentField, init func(*IncidentField), assign func(*IncidentField, *IncidentDebriefQuestion)) error {
	edgeIDs := make([]driver.Value, len(nodes))
	byID := make(map[uuid.UUID]*IncidentField)
	nids := make(map[uuid.UUID]map[*IncidentField]struct{})
	for i, node := range nodes {
		edgeIDs[i] = node.ID
		byID[node.ID] = node
		if init != nil {
			init(node)
		}
	}
	query.Where(func(s *sql.Selector) {
		joinT := sql.Table(incidentfield.DebriefQuestionsTable)
		s.Join(joinT).On(s.C(incidentdebriefquestion.FieldID), joinT.C(incidentfield.DebriefQuestionsPrimaryKey[0]))
		s.Where(sql.InValues(joinT.C(incidentfield.DebriefQuestionsPrimaryKey[1]), edgeIDs...))
		columns := s.SelectedColumns()
		s.Select(joinT.C(incidentfield.DebriefQuestionsPrimaryKey[1]))
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
					nids[inValue] = map[*IncidentField]struct{}{byID[outValue]: {}}
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

func (ifq *IncidentFieldQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := ifq.querySpec()
	if len(ifq.modifiers) > 0 {
		_spec.Modifiers = ifq.modifiers
	}
	_spec.Node.Columns = ifq.ctx.Fields
	if len(ifq.ctx.Fields) > 0 {
		_spec.Unique = ifq.ctx.Unique != nil && *ifq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, ifq.driver, _spec)
}

func (ifq *IncidentFieldQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(incidentfield.Table, incidentfield.Columns, sqlgraph.NewFieldSpec(incidentfield.FieldID, field.TypeUUID))
	_spec.From = ifq.sql
	if unique := ifq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if ifq.path != nil {
		_spec.Unique = true
	}
	if fields := ifq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, incidentfield.FieldID)
		for i := range fields {
			if fields[i] != incidentfield.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := ifq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := ifq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := ifq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := ifq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (ifq *IncidentFieldQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(ifq.driver.Dialect())
	t1 := builder.Table(incidentfield.Table)
	columns := ifq.ctx.Fields
	if len(columns) == 0 {
		columns = incidentfield.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if ifq.sql != nil {
		selector = ifq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if ifq.ctx.Unique != nil && *ifq.ctx.Unique {
		selector.Distinct()
	}
	for _, m := range ifq.modifiers {
		m(selector)
	}
	for _, p := range ifq.predicates {
		p(selector)
	}
	for _, p := range ifq.order {
		p(selector)
	}
	if offset := ifq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := ifq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// Modify adds a query modifier for attaching custom logic to queries.
func (ifq *IncidentFieldQuery) Modify(modifiers ...func(s *sql.Selector)) *IncidentFieldSelect {
	ifq.modifiers = append(ifq.modifiers, modifiers...)
	return ifq.Select()
}

// IncidentFieldGroupBy is the group-by builder for IncidentField entities.
type IncidentFieldGroupBy struct {
	selector
	build *IncidentFieldQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (ifgb *IncidentFieldGroupBy) Aggregate(fns ...AggregateFunc) *IncidentFieldGroupBy {
	ifgb.fns = append(ifgb.fns, fns...)
	return ifgb
}

// Scan applies the selector query and scans the result into the given value.
func (ifgb *IncidentFieldGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, ifgb.build.ctx, ent.OpQueryGroupBy)
	if err := ifgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*IncidentFieldQuery, *IncidentFieldGroupBy](ctx, ifgb.build, ifgb, ifgb.build.inters, v)
}

func (ifgb *IncidentFieldGroupBy) sqlScan(ctx context.Context, root *IncidentFieldQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(ifgb.fns))
	for _, fn := range ifgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*ifgb.flds)+len(ifgb.fns))
		for _, f := range *ifgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*ifgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := ifgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// IncidentFieldSelect is the builder for selecting fields of IncidentField entities.
type IncidentFieldSelect struct {
	*IncidentFieldQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (ifs *IncidentFieldSelect) Aggregate(fns ...AggregateFunc) *IncidentFieldSelect {
	ifs.fns = append(ifs.fns, fns...)
	return ifs
}

// Scan applies the selector query and scans the result into the given value.
func (ifs *IncidentFieldSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, ifs.ctx, ent.OpQuerySelect)
	if err := ifs.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*IncidentFieldQuery, *IncidentFieldSelect](ctx, ifs.IncidentFieldQuery, ifs, ifs.inters, v)
}

func (ifs *IncidentFieldSelect) sqlScan(ctx context.Context, root *IncidentFieldQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(ifs.fns))
	for _, fn := range ifs.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*ifs.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := ifs.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// Modify adds a query modifier for attaching custom logic to queries.
func (ifs *IncidentFieldSelect) Modify(modifiers ...func(s *sql.Selector)) *IncidentFieldSelect {
	ifs.modifiers = append(ifs.modifiers, modifiers...)
	return ifs
}
