// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/dialect/sql/sqljson"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/incidentevent"
	"github.com/rezible/rezible/ent/incidenteventcontext"
	"github.com/rezible/rezible/ent/predicate"
)

// IncidentEventContextUpdate is the builder for updating IncidentEventContext entities.
type IncidentEventContextUpdate struct {
	config
	hooks     []Hook
	mutation  *IncidentEventContextMutation
	modifiers []func(*sql.UpdateBuilder)
}

// Where appends a list predicates to the IncidentEventContextUpdate builder.
func (iecu *IncidentEventContextUpdate) Where(ps ...predicate.IncidentEventContext) *IncidentEventContextUpdate {
	iecu.mutation.Where(ps...)
	return iecu
}

// SetSystemState sets the "system_state" field.
func (iecu *IncidentEventContextUpdate) SetSystemState(s string) *IncidentEventContextUpdate {
	iecu.mutation.SetSystemState(s)
	return iecu
}

// SetNillableSystemState sets the "system_state" field if the given value is not nil.
func (iecu *IncidentEventContextUpdate) SetNillableSystemState(s *string) *IncidentEventContextUpdate {
	if s != nil {
		iecu.SetSystemState(*s)
	}
	return iecu
}

// ClearSystemState clears the value of the "system_state" field.
func (iecu *IncidentEventContextUpdate) ClearSystemState() *IncidentEventContextUpdate {
	iecu.mutation.ClearSystemState()
	return iecu
}

// SetDecisionOptions sets the "decision_options" field.
func (iecu *IncidentEventContextUpdate) SetDecisionOptions(s []string) *IncidentEventContextUpdate {
	iecu.mutation.SetDecisionOptions(s)
	return iecu
}

// AppendDecisionOptions appends s to the "decision_options" field.
func (iecu *IncidentEventContextUpdate) AppendDecisionOptions(s []string) *IncidentEventContextUpdate {
	iecu.mutation.AppendDecisionOptions(s)
	return iecu
}

// ClearDecisionOptions clears the value of the "decision_options" field.
func (iecu *IncidentEventContextUpdate) ClearDecisionOptions() *IncidentEventContextUpdate {
	iecu.mutation.ClearDecisionOptions()
	return iecu
}

// SetDecisionRationale sets the "decision_rationale" field.
func (iecu *IncidentEventContextUpdate) SetDecisionRationale(s string) *IncidentEventContextUpdate {
	iecu.mutation.SetDecisionRationale(s)
	return iecu
}

// SetNillableDecisionRationale sets the "decision_rationale" field if the given value is not nil.
func (iecu *IncidentEventContextUpdate) SetNillableDecisionRationale(s *string) *IncidentEventContextUpdate {
	if s != nil {
		iecu.SetDecisionRationale(*s)
	}
	return iecu
}

// ClearDecisionRationale clears the value of the "decision_rationale" field.
func (iecu *IncidentEventContextUpdate) ClearDecisionRationale() *IncidentEventContextUpdate {
	iecu.mutation.ClearDecisionRationale()
	return iecu
}

// SetInvolvedPersonnel sets the "involved_personnel" field.
func (iecu *IncidentEventContextUpdate) SetInvolvedPersonnel(s []string) *IncidentEventContextUpdate {
	iecu.mutation.SetInvolvedPersonnel(s)
	return iecu
}

// AppendInvolvedPersonnel appends s to the "involved_personnel" field.
func (iecu *IncidentEventContextUpdate) AppendInvolvedPersonnel(s []string) *IncidentEventContextUpdate {
	iecu.mutation.AppendInvolvedPersonnel(s)
	return iecu
}

// ClearInvolvedPersonnel clears the value of the "involved_personnel" field.
func (iecu *IncidentEventContextUpdate) ClearInvolvedPersonnel() *IncidentEventContextUpdate {
	iecu.mutation.ClearInvolvedPersonnel()
	return iecu
}

// SetCreatedAt sets the "created_at" field.
func (iecu *IncidentEventContextUpdate) SetCreatedAt(t time.Time) *IncidentEventContextUpdate {
	iecu.mutation.SetCreatedAt(t)
	return iecu
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (iecu *IncidentEventContextUpdate) SetNillableCreatedAt(t *time.Time) *IncidentEventContextUpdate {
	if t != nil {
		iecu.SetCreatedAt(*t)
	}
	return iecu
}

// SetEventID sets the "event" edge to the IncidentEvent entity by ID.
func (iecu *IncidentEventContextUpdate) SetEventID(id uuid.UUID) *IncidentEventContextUpdate {
	iecu.mutation.SetEventID(id)
	return iecu
}

// SetEvent sets the "event" edge to the IncidentEvent entity.
func (iecu *IncidentEventContextUpdate) SetEvent(i *IncidentEvent) *IncidentEventContextUpdate {
	return iecu.SetEventID(i.ID)
}

// Mutation returns the IncidentEventContextMutation object of the builder.
func (iecu *IncidentEventContextUpdate) Mutation() *IncidentEventContextMutation {
	return iecu.mutation
}

// ClearEvent clears the "event" edge to the IncidentEvent entity.
func (iecu *IncidentEventContextUpdate) ClearEvent() *IncidentEventContextUpdate {
	iecu.mutation.ClearEvent()
	return iecu
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (iecu *IncidentEventContextUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, iecu.sqlSave, iecu.mutation, iecu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (iecu *IncidentEventContextUpdate) SaveX(ctx context.Context) int {
	affected, err := iecu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (iecu *IncidentEventContextUpdate) Exec(ctx context.Context) error {
	_, err := iecu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (iecu *IncidentEventContextUpdate) ExecX(ctx context.Context) {
	if err := iecu.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (iecu *IncidentEventContextUpdate) check() error {
	if iecu.mutation.EventCleared() && len(iecu.mutation.EventIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "IncidentEventContext.event"`)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (iecu *IncidentEventContextUpdate) Modify(modifiers ...func(u *sql.UpdateBuilder)) *IncidentEventContextUpdate {
	iecu.modifiers = append(iecu.modifiers, modifiers...)
	return iecu
}

func (iecu *IncidentEventContextUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := iecu.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(incidenteventcontext.Table, incidenteventcontext.Columns, sqlgraph.NewFieldSpec(incidenteventcontext.FieldID, field.TypeUUID))
	if ps := iecu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := iecu.mutation.SystemState(); ok {
		_spec.SetField(incidenteventcontext.FieldSystemState, field.TypeString, value)
	}
	if iecu.mutation.SystemStateCleared() {
		_spec.ClearField(incidenteventcontext.FieldSystemState, field.TypeString)
	}
	if value, ok := iecu.mutation.DecisionOptions(); ok {
		_spec.SetField(incidenteventcontext.FieldDecisionOptions, field.TypeJSON, value)
	}
	if value, ok := iecu.mutation.AppendedDecisionOptions(); ok {
		_spec.AddModifier(func(u *sql.UpdateBuilder) {
			sqljson.Append(u, incidenteventcontext.FieldDecisionOptions, value)
		})
	}
	if iecu.mutation.DecisionOptionsCleared() {
		_spec.ClearField(incidenteventcontext.FieldDecisionOptions, field.TypeJSON)
	}
	if value, ok := iecu.mutation.DecisionRationale(); ok {
		_spec.SetField(incidenteventcontext.FieldDecisionRationale, field.TypeString, value)
	}
	if iecu.mutation.DecisionRationaleCleared() {
		_spec.ClearField(incidenteventcontext.FieldDecisionRationale, field.TypeString)
	}
	if value, ok := iecu.mutation.InvolvedPersonnel(); ok {
		_spec.SetField(incidenteventcontext.FieldInvolvedPersonnel, field.TypeJSON, value)
	}
	if value, ok := iecu.mutation.AppendedInvolvedPersonnel(); ok {
		_spec.AddModifier(func(u *sql.UpdateBuilder) {
			sqljson.Append(u, incidenteventcontext.FieldInvolvedPersonnel, value)
		})
	}
	if iecu.mutation.InvolvedPersonnelCleared() {
		_spec.ClearField(incidenteventcontext.FieldInvolvedPersonnel, field.TypeJSON)
	}
	if value, ok := iecu.mutation.CreatedAt(); ok {
		_spec.SetField(incidenteventcontext.FieldCreatedAt, field.TypeTime, value)
	}
	if iecu.mutation.EventCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: true,
			Table:   incidenteventcontext.EventTable,
			Columns: []string{incidenteventcontext.EventColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentevent.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := iecu.mutation.EventIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: true,
			Table:   incidenteventcontext.EventTable,
			Columns: []string{incidenteventcontext.EventColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentevent.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(iecu.modifiers...)
	if n, err = sqlgraph.UpdateNodes(ctx, iecu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{incidenteventcontext.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	iecu.mutation.done = true
	return n, nil
}

// IncidentEventContextUpdateOne is the builder for updating a single IncidentEventContext entity.
type IncidentEventContextUpdateOne struct {
	config
	fields    []string
	hooks     []Hook
	mutation  *IncidentEventContextMutation
	modifiers []func(*sql.UpdateBuilder)
}

// SetSystemState sets the "system_state" field.
func (iecuo *IncidentEventContextUpdateOne) SetSystemState(s string) *IncidentEventContextUpdateOne {
	iecuo.mutation.SetSystemState(s)
	return iecuo
}

// SetNillableSystemState sets the "system_state" field if the given value is not nil.
func (iecuo *IncidentEventContextUpdateOne) SetNillableSystemState(s *string) *IncidentEventContextUpdateOne {
	if s != nil {
		iecuo.SetSystemState(*s)
	}
	return iecuo
}

// ClearSystemState clears the value of the "system_state" field.
func (iecuo *IncidentEventContextUpdateOne) ClearSystemState() *IncidentEventContextUpdateOne {
	iecuo.mutation.ClearSystemState()
	return iecuo
}

// SetDecisionOptions sets the "decision_options" field.
func (iecuo *IncidentEventContextUpdateOne) SetDecisionOptions(s []string) *IncidentEventContextUpdateOne {
	iecuo.mutation.SetDecisionOptions(s)
	return iecuo
}

// AppendDecisionOptions appends s to the "decision_options" field.
func (iecuo *IncidentEventContextUpdateOne) AppendDecisionOptions(s []string) *IncidentEventContextUpdateOne {
	iecuo.mutation.AppendDecisionOptions(s)
	return iecuo
}

// ClearDecisionOptions clears the value of the "decision_options" field.
func (iecuo *IncidentEventContextUpdateOne) ClearDecisionOptions() *IncidentEventContextUpdateOne {
	iecuo.mutation.ClearDecisionOptions()
	return iecuo
}

// SetDecisionRationale sets the "decision_rationale" field.
func (iecuo *IncidentEventContextUpdateOne) SetDecisionRationale(s string) *IncidentEventContextUpdateOne {
	iecuo.mutation.SetDecisionRationale(s)
	return iecuo
}

// SetNillableDecisionRationale sets the "decision_rationale" field if the given value is not nil.
func (iecuo *IncidentEventContextUpdateOne) SetNillableDecisionRationale(s *string) *IncidentEventContextUpdateOne {
	if s != nil {
		iecuo.SetDecisionRationale(*s)
	}
	return iecuo
}

// ClearDecisionRationale clears the value of the "decision_rationale" field.
func (iecuo *IncidentEventContextUpdateOne) ClearDecisionRationale() *IncidentEventContextUpdateOne {
	iecuo.mutation.ClearDecisionRationale()
	return iecuo
}

// SetInvolvedPersonnel sets the "involved_personnel" field.
func (iecuo *IncidentEventContextUpdateOne) SetInvolvedPersonnel(s []string) *IncidentEventContextUpdateOne {
	iecuo.mutation.SetInvolvedPersonnel(s)
	return iecuo
}

// AppendInvolvedPersonnel appends s to the "involved_personnel" field.
func (iecuo *IncidentEventContextUpdateOne) AppendInvolvedPersonnel(s []string) *IncidentEventContextUpdateOne {
	iecuo.mutation.AppendInvolvedPersonnel(s)
	return iecuo
}

// ClearInvolvedPersonnel clears the value of the "involved_personnel" field.
func (iecuo *IncidentEventContextUpdateOne) ClearInvolvedPersonnel() *IncidentEventContextUpdateOne {
	iecuo.mutation.ClearInvolvedPersonnel()
	return iecuo
}

// SetCreatedAt sets the "created_at" field.
func (iecuo *IncidentEventContextUpdateOne) SetCreatedAt(t time.Time) *IncidentEventContextUpdateOne {
	iecuo.mutation.SetCreatedAt(t)
	return iecuo
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (iecuo *IncidentEventContextUpdateOne) SetNillableCreatedAt(t *time.Time) *IncidentEventContextUpdateOne {
	if t != nil {
		iecuo.SetCreatedAt(*t)
	}
	return iecuo
}

// SetEventID sets the "event" edge to the IncidentEvent entity by ID.
func (iecuo *IncidentEventContextUpdateOne) SetEventID(id uuid.UUID) *IncidentEventContextUpdateOne {
	iecuo.mutation.SetEventID(id)
	return iecuo
}

// SetEvent sets the "event" edge to the IncidentEvent entity.
func (iecuo *IncidentEventContextUpdateOne) SetEvent(i *IncidentEvent) *IncidentEventContextUpdateOne {
	return iecuo.SetEventID(i.ID)
}

// Mutation returns the IncidentEventContextMutation object of the builder.
func (iecuo *IncidentEventContextUpdateOne) Mutation() *IncidentEventContextMutation {
	return iecuo.mutation
}

// ClearEvent clears the "event" edge to the IncidentEvent entity.
func (iecuo *IncidentEventContextUpdateOne) ClearEvent() *IncidentEventContextUpdateOne {
	iecuo.mutation.ClearEvent()
	return iecuo
}

// Where appends a list predicates to the IncidentEventContextUpdate builder.
func (iecuo *IncidentEventContextUpdateOne) Where(ps ...predicate.IncidentEventContext) *IncidentEventContextUpdateOne {
	iecuo.mutation.Where(ps...)
	return iecuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (iecuo *IncidentEventContextUpdateOne) Select(field string, fields ...string) *IncidentEventContextUpdateOne {
	iecuo.fields = append([]string{field}, fields...)
	return iecuo
}

// Save executes the query and returns the updated IncidentEventContext entity.
func (iecuo *IncidentEventContextUpdateOne) Save(ctx context.Context) (*IncidentEventContext, error) {
	return withHooks(ctx, iecuo.sqlSave, iecuo.mutation, iecuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (iecuo *IncidentEventContextUpdateOne) SaveX(ctx context.Context) *IncidentEventContext {
	node, err := iecuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (iecuo *IncidentEventContextUpdateOne) Exec(ctx context.Context) error {
	_, err := iecuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (iecuo *IncidentEventContextUpdateOne) ExecX(ctx context.Context) {
	if err := iecuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (iecuo *IncidentEventContextUpdateOne) check() error {
	if iecuo.mutation.EventCleared() && len(iecuo.mutation.EventIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "IncidentEventContext.event"`)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (iecuo *IncidentEventContextUpdateOne) Modify(modifiers ...func(u *sql.UpdateBuilder)) *IncidentEventContextUpdateOne {
	iecuo.modifiers = append(iecuo.modifiers, modifiers...)
	return iecuo
}

func (iecuo *IncidentEventContextUpdateOne) sqlSave(ctx context.Context) (_node *IncidentEventContext, err error) {
	if err := iecuo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(incidenteventcontext.Table, incidenteventcontext.Columns, sqlgraph.NewFieldSpec(incidenteventcontext.FieldID, field.TypeUUID))
	id, ok := iecuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "IncidentEventContext.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := iecuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, incidenteventcontext.FieldID)
		for _, f := range fields {
			if !incidenteventcontext.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != incidenteventcontext.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := iecuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := iecuo.mutation.SystemState(); ok {
		_spec.SetField(incidenteventcontext.FieldSystemState, field.TypeString, value)
	}
	if iecuo.mutation.SystemStateCleared() {
		_spec.ClearField(incidenteventcontext.FieldSystemState, field.TypeString)
	}
	if value, ok := iecuo.mutation.DecisionOptions(); ok {
		_spec.SetField(incidenteventcontext.FieldDecisionOptions, field.TypeJSON, value)
	}
	if value, ok := iecuo.mutation.AppendedDecisionOptions(); ok {
		_spec.AddModifier(func(u *sql.UpdateBuilder) {
			sqljson.Append(u, incidenteventcontext.FieldDecisionOptions, value)
		})
	}
	if iecuo.mutation.DecisionOptionsCleared() {
		_spec.ClearField(incidenteventcontext.FieldDecisionOptions, field.TypeJSON)
	}
	if value, ok := iecuo.mutation.DecisionRationale(); ok {
		_spec.SetField(incidenteventcontext.FieldDecisionRationale, field.TypeString, value)
	}
	if iecuo.mutation.DecisionRationaleCleared() {
		_spec.ClearField(incidenteventcontext.FieldDecisionRationale, field.TypeString)
	}
	if value, ok := iecuo.mutation.InvolvedPersonnel(); ok {
		_spec.SetField(incidenteventcontext.FieldInvolvedPersonnel, field.TypeJSON, value)
	}
	if value, ok := iecuo.mutation.AppendedInvolvedPersonnel(); ok {
		_spec.AddModifier(func(u *sql.UpdateBuilder) {
			sqljson.Append(u, incidenteventcontext.FieldInvolvedPersonnel, value)
		})
	}
	if iecuo.mutation.InvolvedPersonnelCleared() {
		_spec.ClearField(incidenteventcontext.FieldInvolvedPersonnel, field.TypeJSON)
	}
	if value, ok := iecuo.mutation.CreatedAt(); ok {
		_spec.SetField(incidenteventcontext.FieldCreatedAt, field.TypeTime, value)
	}
	if iecuo.mutation.EventCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: true,
			Table:   incidenteventcontext.EventTable,
			Columns: []string{incidenteventcontext.EventColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentevent.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := iecuo.mutation.EventIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2O,
			Inverse: true,
			Table:   incidenteventcontext.EventTable,
			Columns: []string{incidenteventcontext.EventColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentevent.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(iecuo.modifiers...)
	_node = &IncidentEventContext{config: iecuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, iecuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{incidenteventcontext.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	iecuo.mutation.done = true
	return _node, nil
}
