// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/systemcomponent"
	"github.com/rezible/rezible/ent/systemcomponentconstraint"
)

// SystemComponentConstraintUpdate is the builder for updating SystemComponentConstraint entities.
type SystemComponentConstraintUpdate struct {
	config
	hooks     []Hook
	mutation  *SystemComponentConstraintMutation
	modifiers []func(*sql.UpdateBuilder)
}

// Where appends a list predicates to the SystemComponentConstraintUpdate builder.
func (sccu *SystemComponentConstraintUpdate) Where(ps ...predicate.SystemComponentConstraint) *SystemComponentConstraintUpdate {
	sccu.mutation.Where(ps...)
	return sccu
}

// SetComponentID sets the "component_id" field.
func (sccu *SystemComponentConstraintUpdate) SetComponentID(u uuid.UUID) *SystemComponentConstraintUpdate {
	sccu.mutation.SetComponentID(u)
	return sccu
}

// SetNillableComponentID sets the "component_id" field if the given value is not nil.
func (sccu *SystemComponentConstraintUpdate) SetNillableComponentID(u *uuid.UUID) *SystemComponentConstraintUpdate {
	if u != nil {
		sccu.SetComponentID(*u)
	}
	return sccu
}

// SetLabel sets the "label" field.
func (sccu *SystemComponentConstraintUpdate) SetLabel(s string) *SystemComponentConstraintUpdate {
	sccu.mutation.SetLabel(s)
	return sccu
}

// SetNillableLabel sets the "label" field if the given value is not nil.
func (sccu *SystemComponentConstraintUpdate) SetNillableLabel(s *string) *SystemComponentConstraintUpdate {
	if s != nil {
		sccu.SetLabel(*s)
	}
	return sccu
}

// SetDescription sets the "description" field.
func (sccu *SystemComponentConstraintUpdate) SetDescription(s string) *SystemComponentConstraintUpdate {
	sccu.mutation.SetDescription(s)
	return sccu
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (sccu *SystemComponentConstraintUpdate) SetNillableDescription(s *string) *SystemComponentConstraintUpdate {
	if s != nil {
		sccu.SetDescription(*s)
	}
	return sccu
}

// ClearDescription clears the value of the "description" field.
func (sccu *SystemComponentConstraintUpdate) ClearDescription() *SystemComponentConstraintUpdate {
	sccu.mutation.ClearDescription()
	return sccu
}

// SetCreatedAt sets the "created_at" field.
func (sccu *SystemComponentConstraintUpdate) SetCreatedAt(t time.Time) *SystemComponentConstraintUpdate {
	sccu.mutation.SetCreatedAt(t)
	return sccu
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (sccu *SystemComponentConstraintUpdate) SetNillableCreatedAt(t *time.Time) *SystemComponentConstraintUpdate {
	if t != nil {
		sccu.SetCreatedAt(*t)
	}
	return sccu
}

// SetComponent sets the "component" edge to the SystemComponent entity.
func (sccu *SystemComponentConstraintUpdate) SetComponent(s *SystemComponent) *SystemComponentConstraintUpdate {
	return sccu.SetComponentID(s.ID)
}

// Mutation returns the SystemComponentConstraintMutation object of the builder.
func (sccu *SystemComponentConstraintUpdate) Mutation() *SystemComponentConstraintMutation {
	return sccu.mutation
}

// ClearComponent clears the "component" edge to the SystemComponent entity.
func (sccu *SystemComponentConstraintUpdate) ClearComponent() *SystemComponentConstraintUpdate {
	sccu.mutation.ClearComponent()
	return sccu
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (sccu *SystemComponentConstraintUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, sccu.sqlSave, sccu.mutation, sccu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (sccu *SystemComponentConstraintUpdate) SaveX(ctx context.Context) int {
	affected, err := sccu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (sccu *SystemComponentConstraintUpdate) Exec(ctx context.Context) error {
	_, err := sccu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (sccu *SystemComponentConstraintUpdate) ExecX(ctx context.Context) {
	if err := sccu.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (sccu *SystemComponentConstraintUpdate) check() error {
	if sccu.mutation.ComponentCleared() && len(sccu.mutation.ComponentIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "SystemComponentConstraint.component"`)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (sccu *SystemComponentConstraintUpdate) Modify(modifiers ...func(u *sql.UpdateBuilder)) *SystemComponentConstraintUpdate {
	sccu.modifiers = append(sccu.modifiers, modifiers...)
	return sccu
}

func (sccu *SystemComponentConstraintUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := sccu.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(systemcomponentconstraint.Table, systemcomponentconstraint.Columns, sqlgraph.NewFieldSpec(systemcomponentconstraint.FieldID, field.TypeUUID))
	if ps := sccu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := sccu.mutation.Label(); ok {
		_spec.SetField(systemcomponentconstraint.FieldLabel, field.TypeString, value)
	}
	if value, ok := sccu.mutation.Description(); ok {
		_spec.SetField(systemcomponentconstraint.FieldDescription, field.TypeString, value)
	}
	if sccu.mutation.DescriptionCleared() {
		_spec.ClearField(systemcomponentconstraint.FieldDescription, field.TypeString)
	}
	if value, ok := sccu.mutation.CreatedAt(); ok {
		_spec.SetField(systemcomponentconstraint.FieldCreatedAt, field.TypeTime, value)
	}
	if sccu.mutation.ComponentCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemcomponentconstraint.ComponentTable,
			Columns: []string{systemcomponentconstraint.ComponentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponent.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := sccu.mutation.ComponentIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemcomponentconstraint.ComponentTable,
			Columns: []string{systemcomponentconstraint.ComponentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponent.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(sccu.modifiers...)
	if n, err = sqlgraph.UpdateNodes(ctx, sccu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{systemcomponentconstraint.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	sccu.mutation.done = true
	return n, nil
}

// SystemComponentConstraintUpdateOne is the builder for updating a single SystemComponentConstraint entity.
type SystemComponentConstraintUpdateOne struct {
	config
	fields    []string
	hooks     []Hook
	mutation  *SystemComponentConstraintMutation
	modifiers []func(*sql.UpdateBuilder)
}

// SetComponentID sets the "component_id" field.
func (sccuo *SystemComponentConstraintUpdateOne) SetComponentID(u uuid.UUID) *SystemComponentConstraintUpdateOne {
	sccuo.mutation.SetComponentID(u)
	return sccuo
}

// SetNillableComponentID sets the "component_id" field if the given value is not nil.
func (sccuo *SystemComponentConstraintUpdateOne) SetNillableComponentID(u *uuid.UUID) *SystemComponentConstraintUpdateOne {
	if u != nil {
		sccuo.SetComponentID(*u)
	}
	return sccuo
}

// SetLabel sets the "label" field.
func (sccuo *SystemComponentConstraintUpdateOne) SetLabel(s string) *SystemComponentConstraintUpdateOne {
	sccuo.mutation.SetLabel(s)
	return sccuo
}

// SetNillableLabel sets the "label" field if the given value is not nil.
func (sccuo *SystemComponentConstraintUpdateOne) SetNillableLabel(s *string) *SystemComponentConstraintUpdateOne {
	if s != nil {
		sccuo.SetLabel(*s)
	}
	return sccuo
}

// SetDescription sets the "description" field.
func (sccuo *SystemComponentConstraintUpdateOne) SetDescription(s string) *SystemComponentConstraintUpdateOne {
	sccuo.mutation.SetDescription(s)
	return sccuo
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (sccuo *SystemComponentConstraintUpdateOne) SetNillableDescription(s *string) *SystemComponentConstraintUpdateOne {
	if s != nil {
		sccuo.SetDescription(*s)
	}
	return sccuo
}

// ClearDescription clears the value of the "description" field.
func (sccuo *SystemComponentConstraintUpdateOne) ClearDescription() *SystemComponentConstraintUpdateOne {
	sccuo.mutation.ClearDescription()
	return sccuo
}

// SetCreatedAt sets the "created_at" field.
func (sccuo *SystemComponentConstraintUpdateOne) SetCreatedAt(t time.Time) *SystemComponentConstraintUpdateOne {
	sccuo.mutation.SetCreatedAt(t)
	return sccuo
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (sccuo *SystemComponentConstraintUpdateOne) SetNillableCreatedAt(t *time.Time) *SystemComponentConstraintUpdateOne {
	if t != nil {
		sccuo.SetCreatedAt(*t)
	}
	return sccuo
}

// SetComponent sets the "component" edge to the SystemComponent entity.
func (sccuo *SystemComponentConstraintUpdateOne) SetComponent(s *SystemComponent) *SystemComponentConstraintUpdateOne {
	return sccuo.SetComponentID(s.ID)
}

// Mutation returns the SystemComponentConstraintMutation object of the builder.
func (sccuo *SystemComponentConstraintUpdateOne) Mutation() *SystemComponentConstraintMutation {
	return sccuo.mutation
}

// ClearComponent clears the "component" edge to the SystemComponent entity.
func (sccuo *SystemComponentConstraintUpdateOne) ClearComponent() *SystemComponentConstraintUpdateOne {
	sccuo.mutation.ClearComponent()
	return sccuo
}

// Where appends a list predicates to the SystemComponentConstraintUpdate builder.
func (sccuo *SystemComponentConstraintUpdateOne) Where(ps ...predicate.SystemComponentConstraint) *SystemComponentConstraintUpdateOne {
	sccuo.mutation.Where(ps...)
	return sccuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (sccuo *SystemComponentConstraintUpdateOne) Select(field string, fields ...string) *SystemComponentConstraintUpdateOne {
	sccuo.fields = append([]string{field}, fields...)
	return sccuo
}

// Save executes the query and returns the updated SystemComponentConstraint entity.
func (sccuo *SystemComponentConstraintUpdateOne) Save(ctx context.Context) (*SystemComponentConstraint, error) {
	return withHooks(ctx, sccuo.sqlSave, sccuo.mutation, sccuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (sccuo *SystemComponentConstraintUpdateOne) SaveX(ctx context.Context) *SystemComponentConstraint {
	node, err := sccuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (sccuo *SystemComponentConstraintUpdateOne) Exec(ctx context.Context) error {
	_, err := sccuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (sccuo *SystemComponentConstraintUpdateOne) ExecX(ctx context.Context) {
	if err := sccuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (sccuo *SystemComponentConstraintUpdateOne) check() error {
	if sccuo.mutation.ComponentCleared() && len(sccuo.mutation.ComponentIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "SystemComponentConstraint.component"`)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (sccuo *SystemComponentConstraintUpdateOne) Modify(modifiers ...func(u *sql.UpdateBuilder)) *SystemComponentConstraintUpdateOne {
	sccuo.modifiers = append(sccuo.modifiers, modifiers...)
	return sccuo
}

func (sccuo *SystemComponentConstraintUpdateOne) sqlSave(ctx context.Context) (_node *SystemComponentConstraint, err error) {
	if err := sccuo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(systemcomponentconstraint.Table, systemcomponentconstraint.Columns, sqlgraph.NewFieldSpec(systemcomponentconstraint.FieldID, field.TypeUUID))
	id, ok := sccuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "SystemComponentConstraint.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := sccuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, systemcomponentconstraint.FieldID)
		for _, f := range fields {
			if !systemcomponentconstraint.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != systemcomponentconstraint.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := sccuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := sccuo.mutation.Label(); ok {
		_spec.SetField(systemcomponentconstraint.FieldLabel, field.TypeString, value)
	}
	if value, ok := sccuo.mutation.Description(); ok {
		_spec.SetField(systemcomponentconstraint.FieldDescription, field.TypeString, value)
	}
	if sccuo.mutation.DescriptionCleared() {
		_spec.ClearField(systemcomponentconstraint.FieldDescription, field.TypeString)
	}
	if value, ok := sccuo.mutation.CreatedAt(); ok {
		_spec.SetField(systemcomponentconstraint.FieldCreatedAt, field.TypeTime, value)
	}
	if sccuo.mutation.ComponentCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemcomponentconstraint.ComponentTable,
			Columns: []string{systemcomponentconstraint.ComponentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponent.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := sccuo.mutation.ComponentIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemcomponentconstraint.ComponentTable,
			Columns: []string{systemcomponentconstraint.ComponentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponent.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(sccuo.modifiers...)
	_node = &SystemComponentConstraint{config: sccuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, sccuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{systemcomponentconstraint.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	sccuo.mutation.done = true
	return _node, nil
}
