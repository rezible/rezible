// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/alert"
	"github.com/rezible/rezible/ent/alertmetrics"
	"github.com/rezible/rezible/ent/predicate"
)

// AlertMetricsUpdate is the builder for updating AlertMetrics entities.
type AlertMetricsUpdate struct {
	config
	hooks     []Hook
	mutation  *AlertMetricsMutation
	modifiers []func(*sql.UpdateBuilder)
}

// Where appends a list predicates to the AlertMetricsUpdate builder.
func (amu *AlertMetricsUpdate) Where(ps ...predicate.AlertMetrics) *AlertMetricsUpdate {
	amu.mutation.Where(ps...)
	return amu
}

// SetAlertID sets the "alert_id" field.
func (amu *AlertMetricsUpdate) SetAlertID(u uuid.UUID) *AlertMetricsUpdate {
	amu.mutation.SetAlertID(u)
	return amu
}

// SetNillableAlertID sets the "alert_id" field if the given value is not nil.
func (amu *AlertMetricsUpdate) SetNillableAlertID(u *uuid.UUID) *AlertMetricsUpdate {
	if u != nil {
		amu.SetAlertID(*u)
	}
	return amu
}

// SetAlert sets the "alert" edge to the Alert entity.
func (amu *AlertMetricsUpdate) SetAlert(a *Alert) *AlertMetricsUpdate {
	return amu.SetAlertID(a.ID)
}

// Mutation returns the AlertMetricsMutation object of the builder.
func (amu *AlertMetricsUpdate) Mutation() *AlertMetricsMutation {
	return amu.mutation
}

// ClearAlert clears the "alert" edge to the Alert entity.
func (amu *AlertMetricsUpdate) ClearAlert() *AlertMetricsUpdate {
	amu.mutation.ClearAlert()
	return amu
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (amu *AlertMetricsUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, amu.sqlSave, amu.mutation, amu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (amu *AlertMetricsUpdate) SaveX(ctx context.Context) int {
	affected, err := amu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (amu *AlertMetricsUpdate) Exec(ctx context.Context) error {
	_, err := amu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (amu *AlertMetricsUpdate) ExecX(ctx context.Context) {
	if err := amu.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (amu *AlertMetricsUpdate) check() error {
	if amu.mutation.AlertCleared() && len(amu.mutation.AlertIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "AlertMetrics.alert"`)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (amu *AlertMetricsUpdate) Modify(modifiers ...func(u *sql.UpdateBuilder)) *AlertMetricsUpdate {
	amu.modifiers = append(amu.modifiers, modifiers...)
	return amu
}

func (amu *AlertMetricsUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := amu.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(alertmetrics.Table, alertmetrics.Columns, sqlgraph.NewFieldSpec(alertmetrics.FieldID, field.TypeUUID))
	if ps := amu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if amu.mutation.AlertCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   alertmetrics.AlertTable,
			Columns: []string{alertmetrics.AlertColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(alert.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := amu.mutation.AlertIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   alertmetrics.AlertTable,
			Columns: []string{alertmetrics.AlertColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(alert.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(amu.modifiers...)
	if n, err = sqlgraph.UpdateNodes(ctx, amu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{alertmetrics.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	amu.mutation.done = true
	return n, nil
}

// AlertMetricsUpdateOne is the builder for updating a single AlertMetrics entity.
type AlertMetricsUpdateOne struct {
	config
	fields    []string
	hooks     []Hook
	mutation  *AlertMetricsMutation
	modifiers []func(*sql.UpdateBuilder)
}

// SetAlertID sets the "alert_id" field.
func (amuo *AlertMetricsUpdateOne) SetAlertID(u uuid.UUID) *AlertMetricsUpdateOne {
	amuo.mutation.SetAlertID(u)
	return amuo
}

// SetNillableAlertID sets the "alert_id" field if the given value is not nil.
func (amuo *AlertMetricsUpdateOne) SetNillableAlertID(u *uuid.UUID) *AlertMetricsUpdateOne {
	if u != nil {
		amuo.SetAlertID(*u)
	}
	return amuo
}

// SetAlert sets the "alert" edge to the Alert entity.
func (amuo *AlertMetricsUpdateOne) SetAlert(a *Alert) *AlertMetricsUpdateOne {
	return amuo.SetAlertID(a.ID)
}

// Mutation returns the AlertMetricsMutation object of the builder.
func (amuo *AlertMetricsUpdateOne) Mutation() *AlertMetricsMutation {
	return amuo.mutation
}

// ClearAlert clears the "alert" edge to the Alert entity.
func (amuo *AlertMetricsUpdateOne) ClearAlert() *AlertMetricsUpdateOne {
	amuo.mutation.ClearAlert()
	return amuo
}

// Where appends a list predicates to the AlertMetricsUpdate builder.
func (amuo *AlertMetricsUpdateOne) Where(ps ...predicate.AlertMetrics) *AlertMetricsUpdateOne {
	amuo.mutation.Where(ps...)
	return amuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (amuo *AlertMetricsUpdateOne) Select(field string, fields ...string) *AlertMetricsUpdateOne {
	amuo.fields = append([]string{field}, fields...)
	return amuo
}

// Save executes the query and returns the updated AlertMetrics entity.
func (amuo *AlertMetricsUpdateOne) Save(ctx context.Context) (*AlertMetrics, error) {
	return withHooks(ctx, amuo.sqlSave, amuo.mutation, amuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (amuo *AlertMetricsUpdateOne) SaveX(ctx context.Context) *AlertMetrics {
	node, err := amuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (amuo *AlertMetricsUpdateOne) Exec(ctx context.Context) error {
	_, err := amuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (amuo *AlertMetricsUpdateOne) ExecX(ctx context.Context) {
	if err := amuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (amuo *AlertMetricsUpdateOne) check() error {
	if amuo.mutation.AlertCleared() && len(amuo.mutation.AlertIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "AlertMetrics.alert"`)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (amuo *AlertMetricsUpdateOne) Modify(modifiers ...func(u *sql.UpdateBuilder)) *AlertMetricsUpdateOne {
	amuo.modifiers = append(amuo.modifiers, modifiers...)
	return amuo
}

func (amuo *AlertMetricsUpdateOne) sqlSave(ctx context.Context) (_node *AlertMetrics, err error) {
	if err := amuo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(alertmetrics.Table, alertmetrics.Columns, sqlgraph.NewFieldSpec(alertmetrics.FieldID, field.TypeUUID))
	id, ok := amuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "AlertMetrics.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := amuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, alertmetrics.FieldID)
		for _, f := range fields {
			if !alertmetrics.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != alertmetrics.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := amuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if amuo.mutation.AlertCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   alertmetrics.AlertTable,
			Columns: []string{alertmetrics.AlertColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(alert.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := amuo.mutation.AlertIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   alertmetrics.AlertTable,
			Columns: []string{alertmetrics.AlertColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(alert.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(amuo.modifiers...)
	_node = &AlertMetrics{config: amuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, amuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{alertmetrics.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	amuo.mutation.done = true
	return _node, nil
}
