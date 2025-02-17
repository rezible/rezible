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
	"github.com/rezible/rezible/ent/systemcomponentsignal"
	"github.com/rezible/rezible/ent/systemrelationship"
	"github.com/rezible/rezible/ent/systemrelationshipfeedback"
)

// SystemRelationshipFeedbackUpdate is the builder for updating SystemRelationshipFeedback entities.
type SystemRelationshipFeedbackUpdate struct {
	config
	hooks     []Hook
	mutation  *SystemRelationshipFeedbackMutation
	modifiers []func(*sql.UpdateBuilder)
}

// Where appends a list predicates to the SystemRelationshipFeedbackUpdate builder.
func (srfu *SystemRelationshipFeedbackUpdate) Where(ps ...predicate.SystemRelationshipFeedback) *SystemRelationshipFeedbackUpdate {
	srfu.mutation.Where(ps...)
	return srfu
}

// SetRelationshipID sets the "relationship_id" field.
func (srfu *SystemRelationshipFeedbackUpdate) SetRelationshipID(u uuid.UUID) *SystemRelationshipFeedbackUpdate {
	srfu.mutation.SetRelationshipID(u)
	return srfu
}

// SetNillableRelationshipID sets the "relationship_id" field if the given value is not nil.
func (srfu *SystemRelationshipFeedbackUpdate) SetNillableRelationshipID(u *uuid.UUID) *SystemRelationshipFeedbackUpdate {
	if u != nil {
		srfu.SetRelationshipID(*u)
	}
	return srfu
}

// SetSignalID sets the "signal_id" field.
func (srfu *SystemRelationshipFeedbackUpdate) SetSignalID(u uuid.UUID) *SystemRelationshipFeedbackUpdate {
	srfu.mutation.SetSignalID(u)
	return srfu
}

// SetNillableSignalID sets the "signal_id" field if the given value is not nil.
func (srfu *SystemRelationshipFeedbackUpdate) SetNillableSignalID(u *uuid.UUID) *SystemRelationshipFeedbackUpdate {
	if u != nil {
		srfu.SetSignalID(*u)
	}
	return srfu
}

// SetType sets the "type" field.
func (srfu *SystemRelationshipFeedbackUpdate) SetType(s string) *SystemRelationshipFeedbackUpdate {
	srfu.mutation.SetType(s)
	return srfu
}

// SetNillableType sets the "type" field if the given value is not nil.
func (srfu *SystemRelationshipFeedbackUpdate) SetNillableType(s *string) *SystemRelationshipFeedbackUpdate {
	if s != nil {
		srfu.SetType(*s)
	}
	return srfu
}

// SetDescription sets the "description" field.
func (srfu *SystemRelationshipFeedbackUpdate) SetDescription(s string) *SystemRelationshipFeedbackUpdate {
	srfu.mutation.SetDescription(s)
	return srfu
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (srfu *SystemRelationshipFeedbackUpdate) SetNillableDescription(s *string) *SystemRelationshipFeedbackUpdate {
	if s != nil {
		srfu.SetDescription(*s)
	}
	return srfu
}

// ClearDescription clears the value of the "description" field.
func (srfu *SystemRelationshipFeedbackUpdate) ClearDescription() *SystemRelationshipFeedbackUpdate {
	srfu.mutation.ClearDescription()
	return srfu
}

// SetCreatedAt sets the "created_at" field.
func (srfu *SystemRelationshipFeedbackUpdate) SetCreatedAt(t time.Time) *SystemRelationshipFeedbackUpdate {
	srfu.mutation.SetCreatedAt(t)
	return srfu
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (srfu *SystemRelationshipFeedbackUpdate) SetNillableCreatedAt(t *time.Time) *SystemRelationshipFeedbackUpdate {
	if t != nil {
		srfu.SetCreatedAt(*t)
	}
	return srfu
}

// SetSignal sets the "signal" edge to the SystemComponentSignal entity.
func (srfu *SystemRelationshipFeedbackUpdate) SetSignal(s *SystemComponentSignal) *SystemRelationshipFeedbackUpdate {
	return srfu.SetSignalID(s.ID)
}

// SetRelationship sets the "relationship" edge to the SystemRelationship entity.
func (srfu *SystemRelationshipFeedbackUpdate) SetRelationship(s *SystemRelationship) *SystemRelationshipFeedbackUpdate {
	return srfu.SetRelationshipID(s.ID)
}

// Mutation returns the SystemRelationshipFeedbackMutation object of the builder.
func (srfu *SystemRelationshipFeedbackUpdate) Mutation() *SystemRelationshipFeedbackMutation {
	return srfu.mutation
}

// ClearSignal clears the "signal" edge to the SystemComponentSignal entity.
func (srfu *SystemRelationshipFeedbackUpdate) ClearSignal() *SystemRelationshipFeedbackUpdate {
	srfu.mutation.ClearSignal()
	return srfu
}

// ClearRelationship clears the "relationship" edge to the SystemRelationship entity.
func (srfu *SystemRelationshipFeedbackUpdate) ClearRelationship() *SystemRelationshipFeedbackUpdate {
	srfu.mutation.ClearRelationship()
	return srfu
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (srfu *SystemRelationshipFeedbackUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, srfu.sqlSave, srfu.mutation, srfu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (srfu *SystemRelationshipFeedbackUpdate) SaveX(ctx context.Context) int {
	affected, err := srfu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (srfu *SystemRelationshipFeedbackUpdate) Exec(ctx context.Context) error {
	_, err := srfu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (srfu *SystemRelationshipFeedbackUpdate) ExecX(ctx context.Context) {
	if err := srfu.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (srfu *SystemRelationshipFeedbackUpdate) check() error {
	if v, ok := srfu.mutation.GetType(); ok {
		if err := systemrelationshipfeedback.TypeValidator(v); err != nil {
			return &ValidationError{Name: "type", err: fmt.Errorf(`ent: validator failed for field "SystemRelationshipFeedback.type": %w`, err)}
		}
	}
	if srfu.mutation.SignalCleared() && len(srfu.mutation.SignalIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "SystemRelationshipFeedback.signal"`)
	}
	if srfu.mutation.RelationshipCleared() && len(srfu.mutation.RelationshipIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "SystemRelationshipFeedback.relationship"`)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (srfu *SystemRelationshipFeedbackUpdate) Modify(modifiers ...func(u *sql.UpdateBuilder)) *SystemRelationshipFeedbackUpdate {
	srfu.modifiers = append(srfu.modifiers, modifiers...)
	return srfu
}

func (srfu *SystemRelationshipFeedbackUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := srfu.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(systemrelationshipfeedback.Table, systemrelationshipfeedback.Columns, sqlgraph.NewFieldSpec(systemrelationshipfeedback.FieldID, field.TypeUUID))
	if ps := srfu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := srfu.mutation.GetType(); ok {
		_spec.SetField(systemrelationshipfeedback.FieldType, field.TypeString, value)
	}
	if value, ok := srfu.mutation.Description(); ok {
		_spec.SetField(systemrelationshipfeedback.FieldDescription, field.TypeString, value)
	}
	if srfu.mutation.DescriptionCleared() {
		_spec.ClearField(systemrelationshipfeedback.FieldDescription, field.TypeString)
	}
	if value, ok := srfu.mutation.CreatedAt(); ok {
		_spec.SetField(systemrelationshipfeedback.FieldCreatedAt, field.TypeTime, value)
	}
	if srfu.mutation.SignalCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemrelationshipfeedback.SignalTable,
			Columns: []string{systemrelationshipfeedback.SignalColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponentsignal.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := srfu.mutation.SignalIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemrelationshipfeedback.SignalTable,
			Columns: []string{systemrelationshipfeedback.SignalColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponentsignal.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if srfu.mutation.RelationshipCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemrelationshipfeedback.RelationshipTable,
			Columns: []string{systemrelationshipfeedback.RelationshipColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemrelationship.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := srfu.mutation.RelationshipIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemrelationshipfeedback.RelationshipTable,
			Columns: []string{systemrelationshipfeedback.RelationshipColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemrelationship.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(srfu.modifiers...)
	if n, err = sqlgraph.UpdateNodes(ctx, srfu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{systemrelationshipfeedback.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	srfu.mutation.done = true
	return n, nil
}

// SystemRelationshipFeedbackUpdateOne is the builder for updating a single SystemRelationshipFeedback entity.
type SystemRelationshipFeedbackUpdateOne struct {
	config
	fields    []string
	hooks     []Hook
	mutation  *SystemRelationshipFeedbackMutation
	modifiers []func(*sql.UpdateBuilder)
}

// SetRelationshipID sets the "relationship_id" field.
func (srfuo *SystemRelationshipFeedbackUpdateOne) SetRelationshipID(u uuid.UUID) *SystemRelationshipFeedbackUpdateOne {
	srfuo.mutation.SetRelationshipID(u)
	return srfuo
}

// SetNillableRelationshipID sets the "relationship_id" field if the given value is not nil.
func (srfuo *SystemRelationshipFeedbackUpdateOne) SetNillableRelationshipID(u *uuid.UUID) *SystemRelationshipFeedbackUpdateOne {
	if u != nil {
		srfuo.SetRelationshipID(*u)
	}
	return srfuo
}

// SetSignalID sets the "signal_id" field.
func (srfuo *SystemRelationshipFeedbackUpdateOne) SetSignalID(u uuid.UUID) *SystemRelationshipFeedbackUpdateOne {
	srfuo.mutation.SetSignalID(u)
	return srfuo
}

// SetNillableSignalID sets the "signal_id" field if the given value is not nil.
func (srfuo *SystemRelationshipFeedbackUpdateOne) SetNillableSignalID(u *uuid.UUID) *SystemRelationshipFeedbackUpdateOne {
	if u != nil {
		srfuo.SetSignalID(*u)
	}
	return srfuo
}

// SetType sets the "type" field.
func (srfuo *SystemRelationshipFeedbackUpdateOne) SetType(s string) *SystemRelationshipFeedbackUpdateOne {
	srfuo.mutation.SetType(s)
	return srfuo
}

// SetNillableType sets the "type" field if the given value is not nil.
func (srfuo *SystemRelationshipFeedbackUpdateOne) SetNillableType(s *string) *SystemRelationshipFeedbackUpdateOne {
	if s != nil {
		srfuo.SetType(*s)
	}
	return srfuo
}

// SetDescription sets the "description" field.
func (srfuo *SystemRelationshipFeedbackUpdateOne) SetDescription(s string) *SystemRelationshipFeedbackUpdateOne {
	srfuo.mutation.SetDescription(s)
	return srfuo
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (srfuo *SystemRelationshipFeedbackUpdateOne) SetNillableDescription(s *string) *SystemRelationshipFeedbackUpdateOne {
	if s != nil {
		srfuo.SetDescription(*s)
	}
	return srfuo
}

// ClearDescription clears the value of the "description" field.
func (srfuo *SystemRelationshipFeedbackUpdateOne) ClearDescription() *SystemRelationshipFeedbackUpdateOne {
	srfuo.mutation.ClearDescription()
	return srfuo
}

// SetCreatedAt sets the "created_at" field.
func (srfuo *SystemRelationshipFeedbackUpdateOne) SetCreatedAt(t time.Time) *SystemRelationshipFeedbackUpdateOne {
	srfuo.mutation.SetCreatedAt(t)
	return srfuo
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (srfuo *SystemRelationshipFeedbackUpdateOne) SetNillableCreatedAt(t *time.Time) *SystemRelationshipFeedbackUpdateOne {
	if t != nil {
		srfuo.SetCreatedAt(*t)
	}
	return srfuo
}

// SetSignal sets the "signal" edge to the SystemComponentSignal entity.
func (srfuo *SystemRelationshipFeedbackUpdateOne) SetSignal(s *SystemComponentSignal) *SystemRelationshipFeedbackUpdateOne {
	return srfuo.SetSignalID(s.ID)
}

// SetRelationship sets the "relationship" edge to the SystemRelationship entity.
func (srfuo *SystemRelationshipFeedbackUpdateOne) SetRelationship(s *SystemRelationship) *SystemRelationshipFeedbackUpdateOne {
	return srfuo.SetRelationshipID(s.ID)
}

// Mutation returns the SystemRelationshipFeedbackMutation object of the builder.
func (srfuo *SystemRelationshipFeedbackUpdateOne) Mutation() *SystemRelationshipFeedbackMutation {
	return srfuo.mutation
}

// ClearSignal clears the "signal" edge to the SystemComponentSignal entity.
func (srfuo *SystemRelationshipFeedbackUpdateOne) ClearSignal() *SystemRelationshipFeedbackUpdateOne {
	srfuo.mutation.ClearSignal()
	return srfuo
}

// ClearRelationship clears the "relationship" edge to the SystemRelationship entity.
func (srfuo *SystemRelationshipFeedbackUpdateOne) ClearRelationship() *SystemRelationshipFeedbackUpdateOne {
	srfuo.mutation.ClearRelationship()
	return srfuo
}

// Where appends a list predicates to the SystemRelationshipFeedbackUpdate builder.
func (srfuo *SystemRelationshipFeedbackUpdateOne) Where(ps ...predicate.SystemRelationshipFeedback) *SystemRelationshipFeedbackUpdateOne {
	srfuo.mutation.Where(ps...)
	return srfuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (srfuo *SystemRelationshipFeedbackUpdateOne) Select(field string, fields ...string) *SystemRelationshipFeedbackUpdateOne {
	srfuo.fields = append([]string{field}, fields...)
	return srfuo
}

// Save executes the query and returns the updated SystemRelationshipFeedback entity.
func (srfuo *SystemRelationshipFeedbackUpdateOne) Save(ctx context.Context) (*SystemRelationshipFeedback, error) {
	return withHooks(ctx, srfuo.sqlSave, srfuo.mutation, srfuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (srfuo *SystemRelationshipFeedbackUpdateOne) SaveX(ctx context.Context) *SystemRelationshipFeedback {
	node, err := srfuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (srfuo *SystemRelationshipFeedbackUpdateOne) Exec(ctx context.Context) error {
	_, err := srfuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (srfuo *SystemRelationshipFeedbackUpdateOne) ExecX(ctx context.Context) {
	if err := srfuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (srfuo *SystemRelationshipFeedbackUpdateOne) check() error {
	if v, ok := srfuo.mutation.GetType(); ok {
		if err := systemrelationshipfeedback.TypeValidator(v); err != nil {
			return &ValidationError{Name: "type", err: fmt.Errorf(`ent: validator failed for field "SystemRelationshipFeedback.type": %w`, err)}
		}
	}
	if srfuo.mutation.SignalCleared() && len(srfuo.mutation.SignalIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "SystemRelationshipFeedback.signal"`)
	}
	if srfuo.mutation.RelationshipCleared() && len(srfuo.mutation.RelationshipIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "SystemRelationshipFeedback.relationship"`)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (srfuo *SystemRelationshipFeedbackUpdateOne) Modify(modifiers ...func(u *sql.UpdateBuilder)) *SystemRelationshipFeedbackUpdateOne {
	srfuo.modifiers = append(srfuo.modifiers, modifiers...)
	return srfuo
}

func (srfuo *SystemRelationshipFeedbackUpdateOne) sqlSave(ctx context.Context) (_node *SystemRelationshipFeedback, err error) {
	if err := srfuo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(systemrelationshipfeedback.Table, systemrelationshipfeedback.Columns, sqlgraph.NewFieldSpec(systemrelationshipfeedback.FieldID, field.TypeUUID))
	id, ok := srfuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "SystemRelationshipFeedback.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := srfuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, systemrelationshipfeedback.FieldID)
		for _, f := range fields {
			if !systemrelationshipfeedback.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != systemrelationshipfeedback.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := srfuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := srfuo.mutation.GetType(); ok {
		_spec.SetField(systemrelationshipfeedback.FieldType, field.TypeString, value)
	}
	if value, ok := srfuo.mutation.Description(); ok {
		_spec.SetField(systemrelationshipfeedback.FieldDescription, field.TypeString, value)
	}
	if srfuo.mutation.DescriptionCleared() {
		_spec.ClearField(systemrelationshipfeedback.FieldDescription, field.TypeString)
	}
	if value, ok := srfuo.mutation.CreatedAt(); ok {
		_spec.SetField(systemrelationshipfeedback.FieldCreatedAt, field.TypeTime, value)
	}
	if srfuo.mutation.SignalCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemrelationshipfeedback.SignalTable,
			Columns: []string{systemrelationshipfeedback.SignalColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponentsignal.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := srfuo.mutation.SignalIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemrelationshipfeedback.SignalTable,
			Columns: []string{systemrelationshipfeedback.SignalColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponentsignal.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if srfuo.mutation.RelationshipCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemrelationshipfeedback.RelationshipTable,
			Columns: []string{systemrelationshipfeedback.RelationshipColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemrelationship.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := srfuo.mutation.RelationshipIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemrelationshipfeedback.RelationshipTable,
			Columns: []string{systemrelationshipfeedback.RelationshipColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemrelationship.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(srfuo.modifiers...)
	_node = &SystemRelationshipFeedback{config: srfuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, srfuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{systemrelationshipfeedback.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	srfuo.mutation.done = true
	return _node, nil
}
