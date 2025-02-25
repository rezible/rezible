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
	"github.com/rezible/rezible/ent/systemrelationshipfeedbacksignal"
)

// SystemRelationshipFeedbackSignalUpdate is the builder for updating SystemRelationshipFeedbackSignal entities.
type SystemRelationshipFeedbackSignalUpdate struct {
	config
	hooks     []Hook
	mutation  *SystemRelationshipFeedbackSignalMutation
	modifiers []func(*sql.UpdateBuilder)
}

// Where appends a list predicates to the SystemRelationshipFeedbackSignalUpdate builder.
func (srfsu *SystemRelationshipFeedbackSignalUpdate) Where(ps ...predicate.SystemRelationshipFeedbackSignal) *SystemRelationshipFeedbackSignalUpdate {
	srfsu.mutation.Where(ps...)
	return srfsu
}

// SetRelationshipID sets the "relationship_id" field.
func (srfsu *SystemRelationshipFeedbackSignalUpdate) SetRelationshipID(u uuid.UUID) *SystemRelationshipFeedbackSignalUpdate {
	srfsu.mutation.SetRelationshipID(u)
	return srfsu
}

// SetNillableRelationshipID sets the "relationship_id" field if the given value is not nil.
func (srfsu *SystemRelationshipFeedbackSignalUpdate) SetNillableRelationshipID(u *uuid.UUID) *SystemRelationshipFeedbackSignalUpdate {
	if u != nil {
		srfsu.SetRelationshipID(*u)
	}
	return srfsu
}

// SetSignalID sets the "signal_id" field.
func (srfsu *SystemRelationshipFeedbackSignalUpdate) SetSignalID(u uuid.UUID) *SystemRelationshipFeedbackSignalUpdate {
	srfsu.mutation.SetSignalID(u)
	return srfsu
}

// SetNillableSignalID sets the "signal_id" field if the given value is not nil.
func (srfsu *SystemRelationshipFeedbackSignalUpdate) SetNillableSignalID(u *uuid.UUID) *SystemRelationshipFeedbackSignalUpdate {
	if u != nil {
		srfsu.SetSignalID(*u)
	}
	return srfsu
}

// SetType sets the "type" field.
func (srfsu *SystemRelationshipFeedbackSignalUpdate) SetType(s string) *SystemRelationshipFeedbackSignalUpdate {
	srfsu.mutation.SetType(s)
	return srfsu
}

// SetNillableType sets the "type" field if the given value is not nil.
func (srfsu *SystemRelationshipFeedbackSignalUpdate) SetNillableType(s *string) *SystemRelationshipFeedbackSignalUpdate {
	if s != nil {
		srfsu.SetType(*s)
	}
	return srfsu
}

// SetDescription sets the "description" field.
func (srfsu *SystemRelationshipFeedbackSignalUpdate) SetDescription(s string) *SystemRelationshipFeedbackSignalUpdate {
	srfsu.mutation.SetDescription(s)
	return srfsu
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (srfsu *SystemRelationshipFeedbackSignalUpdate) SetNillableDescription(s *string) *SystemRelationshipFeedbackSignalUpdate {
	if s != nil {
		srfsu.SetDescription(*s)
	}
	return srfsu
}

// ClearDescription clears the value of the "description" field.
func (srfsu *SystemRelationshipFeedbackSignalUpdate) ClearDescription() *SystemRelationshipFeedbackSignalUpdate {
	srfsu.mutation.ClearDescription()
	return srfsu
}

// SetCreatedAt sets the "created_at" field.
func (srfsu *SystemRelationshipFeedbackSignalUpdate) SetCreatedAt(t time.Time) *SystemRelationshipFeedbackSignalUpdate {
	srfsu.mutation.SetCreatedAt(t)
	return srfsu
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (srfsu *SystemRelationshipFeedbackSignalUpdate) SetNillableCreatedAt(t *time.Time) *SystemRelationshipFeedbackSignalUpdate {
	if t != nil {
		srfsu.SetCreatedAt(*t)
	}
	return srfsu
}

// SetRelationship sets the "relationship" edge to the SystemRelationship entity.
func (srfsu *SystemRelationshipFeedbackSignalUpdate) SetRelationship(s *SystemRelationship) *SystemRelationshipFeedbackSignalUpdate {
	return srfsu.SetRelationshipID(s.ID)
}

// SetSignal sets the "signal" edge to the SystemComponentSignal entity.
func (srfsu *SystemRelationshipFeedbackSignalUpdate) SetSignal(s *SystemComponentSignal) *SystemRelationshipFeedbackSignalUpdate {
	return srfsu.SetSignalID(s.ID)
}

// Mutation returns the SystemRelationshipFeedbackSignalMutation object of the builder.
func (srfsu *SystemRelationshipFeedbackSignalUpdate) Mutation() *SystemRelationshipFeedbackSignalMutation {
	return srfsu.mutation
}

// ClearRelationship clears the "relationship" edge to the SystemRelationship entity.
func (srfsu *SystemRelationshipFeedbackSignalUpdate) ClearRelationship() *SystemRelationshipFeedbackSignalUpdate {
	srfsu.mutation.ClearRelationship()
	return srfsu
}

// ClearSignal clears the "signal" edge to the SystemComponentSignal entity.
func (srfsu *SystemRelationshipFeedbackSignalUpdate) ClearSignal() *SystemRelationshipFeedbackSignalUpdate {
	srfsu.mutation.ClearSignal()
	return srfsu
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (srfsu *SystemRelationshipFeedbackSignalUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, srfsu.sqlSave, srfsu.mutation, srfsu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (srfsu *SystemRelationshipFeedbackSignalUpdate) SaveX(ctx context.Context) int {
	affected, err := srfsu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (srfsu *SystemRelationshipFeedbackSignalUpdate) Exec(ctx context.Context) error {
	_, err := srfsu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (srfsu *SystemRelationshipFeedbackSignalUpdate) ExecX(ctx context.Context) {
	if err := srfsu.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (srfsu *SystemRelationshipFeedbackSignalUpdate) check() error {
	if v, ok := srfsu.mutation.GetType(); ok {
		if err := systemrelationshipfeedbacksignal.TypeValidator(v); err != nil {
			return &ValidationError{Name: "type", err: fmt.Errorf(`ent: validator failed for field "SystemRelationshipFeedbackSignal.type": %w`, err)}
		}
	}
	if srfsu.mutation.RelationshipCleared() && len(srfsu.mutation.RelationshipIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "SystemRelationshipFeedbackSignal.relationship"`)
	}
	if srfsu.mutation.SignalCleared() && len(srfsu.mutation.SignalIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "SystemRelationshipFeedbackSignal.signal"`)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (srfsu *SystemRelationshipFeedbackSignalUpdate) Modify(modifiers ...func(u *sql.UpdateBuilder)) *SystemRelationshipFeedbackSignalUpdate {
	srfsu.modifiers = append(srfsu.modifiers, modifiers...)
	return srfsu
}

func (srfsu *SystemRelationshipFeedbackSignalUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := srfsu.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(systemrelationshipfeedbacksignal.Table, systemrelationshipfeedbacksignal.Columns, sqlgraph.NewFieldSpec(systemrelationshipfeedbacksignal.FieldID, field.TypeUUID))
	if ps := srfsu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := srfsu.mutation.GetType(); ok {
		_spec.SetField(systemrelationshipfeedbacksignal.FieldType, field.TypeString, value)
	}
	if value, ok := srfsu.mutation.Description(); ok {
		_spec.SetField(systemrelationshipfeedbacksignal.FieldDescription, field.TypeString, value)
	}
	if srfsu.mutation.DescriptionCleared() {
		_spec.ClearField(systemrelationshipfeedbacksignal.FieldDescription, field.TypeString)
	}
	if value, ok := srfsu.mutation.CreatedAt(); ok {
		_spec.SetField(systemrelationshipfeedbacksignal.FieldCreatedAt, field.TypeTime, value)
	}
	if srfsu.mutation.RelationshipCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemrelationshipfeedbacksignal.RelationshipTable,
			Columns: []string{systemrelationshipfeedbacksignal.RelationshipColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemrelationship.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := srfsu.mutation.RelationshipIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemrelationshipfeedbacksignal.RelationshipTable,
			Columns: []string{systemrelationshipfeedbacksignal.RelationshipColumn},
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
	if srfsu.mutation.SignalCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemrelationshipfeedbacksignal.SignalTable,
			Columns: []string{systemrelationshipfeedbacksignal.SignalColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponentsignal.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := srfsu.mutation.SignalIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemrelationshipfeedbacksignal.SignalTable,
			Columns: []string{systemrelationshipfeedbacksignal.SignalColumn},
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
	_spec.AddModifiers(srfsu.modifiers...)
	if n, err = sqlgraph.UpdateNodes(ctx, srfsu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{systemrelationshipfeedbacksignal.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	srfsu.mutation.done = true
	return n, nil
}

// SystemRelationshipFeedbackSignalUpdateOne is the builder for updating a single SystemRelationshipFeedbackSignal entity.
type SystemRelationshipFeedbackSignalUpdateOne struct {
	config
	fields    []string
	hooks     []Hook
	mutation  *SystemRelationshipFeedbackSignalMutation
	modifiers []func(*sql.UpdateBuilder)
}

// SetRelationshipID sets the "relationship_id" field.
func (srfsuo *SystemRelationshipFeedbackSignalUpdateOne) SetRelationshipID(u uuid.UUID) *SystemRelationshipFeedbackSignalUpdateOne {
	srfsuo.mutation.SetRelationshipID(u)
	return srfsuo
}

// SetNillableRelationshipID sets the "relationship_id" field if the given value is not nil.
func (srfsuo *SystemRelationshipFeedbackSignalUpdateOne) SetNillableRelationshipID(u *uuid.UUID) *SystemRelationshipFeedbackSignalUpdateOne {
	if u != nil {
		srfsuo.SetRelationshipID(*u)
	}
	return srfsuo
}

// SetSignalID sets the "signal_id" field.
func (srfsuo *SystemRelationshipFeedbackSignalUpdateOne) SetSignalID(u uuid.UUID) *SystemRelationshipFeedbackSignalUpdateOne {
	srfsuo.mutation.SetSignalID(u)
	return srfsuo
}

// SetNillableSignalID sets the "signal_id" field if the given value is not nil.
func (srfsuo *SystemRelationshipFeedbackSignalUpdateOne) SetNillableSignalID(u *uuid.UUID) *SystemRelationshipFeedbackSignalUpdateOne {
	if u != nil {
		srfsuo.SetSignalID(*u)
	}
	return srfsuo
}

// SetType sets the "type" field.
func (srfsuo *SystemRelationshipFeedbackSignalUpdateOne) SetType(s string) *SystemRelationshipFeedbackSignalUpdateOne {
	srfsuo.mutation.SetType(s)
	return srfsuo
}

// SetNillableType sets the "type" field if the given value is not nil.
func (srfsuo *SystemRelationshipFeedbackSignalUpdateOne) SetNillableType(s *string) *SystemRelationshipFeedbackSignalUpdateOne {
	if s != nil {
		srfsuo.SetType(*s)
	}
	return srfsuo
}

// SetDescription sets the "description" field.
func (srfsuo *SystemRelationshipFeedbackSignalUpdateOne) SetDescription(s string) *SystemRelationshipFeedbackSignalUpdateOne {
	srfsuo.mutation.SetDescription(s)
	return srfsuo
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (srfsuo *SystemRelationshipFeedbackSignalUpdateOne) SetNillableDescription(s *string) *SystemRelationshipFeedbackSignalUpdateOne {
	if s != nil {
		srfsuo.SetDescription(*s)
	}
	return srfsuo
}

// ClearDescription clears the value of the "description" field.
func (srfsuo *SystemRelationshipFeedbackSignalUpdateOne) ClearDescription() *SystemRelationshipFeedbackSignalUpdateOne {
	srfsuo.mutation.ClearDescription()
	return srfsuo
}

// SetCreatedAt sets the "created_at" field.
func (srfsuo *SystemRelationshipFeedbackSignalUpdateOne) SetCreatedAt(t time.Time) *SystemRelationshipFeedbackSignalUpdateOne {
	srfsuo.mutation.SetCreatedAt(t)
	return srfsuo
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (srfsuo *SystemRelationshipFeedbackSignalUpdateOne) SetNillableCreatedAt(t *time.Time) *SystemRelationshipFeedbackSignalUpdateOne {
	if t != nil {
		srfsuo.SetCreatedAt(*t)
	}
	return srfsuo
}

// SetRelationship sets the "relationship" edge to the SystemRelationship entity.
func (srfsuo *SystemRelationshipFeedbackSignalUpdateOne) SetRelationship(s *SystemRelationship) *SystemRelationshipFeedbackSignalUpdateOne {
	return srfsuo.SetRelationshipID(s.ID)
}

// SetSignal sets the "signal" edge to the SystemComponentSignal entity.
func (srfsuo *SystemRelationshipFeedbackSignalUpdateOne) SetSignal(s *SystemComponentSignal) *SystemRelationshipFeedbackSignalUpdateOne {
	return srfsuo.SetSignalID(s.ID)
}

// Mutation returns the SystemRelationshipFeedbackSignalMutation object of the builder.
func (srfsuo *SystemRelationshipFeedbackSignalUpdateOne) Mutation() *SystemRelationshipFeedbackSignalMutation {
	return srfsuo.mutation
}

// ClearRelationship clears the "relationship" edge to the SystemRelationship entity.
func (srfsuo *SystemRelationshipFeedbackSignalUpdateOne) ClearRelationship() *SystemRelationshipFeedbackSignalUpdateOne {
	srfsuo.mutation.ClearRelationship()
	return srfsuo
}

// ClearSignal clears the "signal" edge to the SystemComponentSignal entity.
func (srfsuo *SystemRelationshipFeedbackSignalUpdateOne) ClearSignal() *SystemRelationshipFeedbackSignalUpdateOne {
	srfsuo.mutation.ClearSignal()
	return srfsuo
}

// Where appends a list predicates to the SystemRelationshipFeedbackSignalUpdate builder.
func (srfsuo *SystemRelationshipFeedbackSignalUpdateOne) Where(ps ...predicate.SystemRelationshipFeedbackSignal) *SystemRelationshipFeedbackSignalUpdateOne {
	srfsuo.mutation.Where(ps...)
	return srfsuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (srfsuo *SystemRelationshipFeedbackSignalUpdateOne) Select(field string, fields ...string) *SystemRelationshipFeedbackSignalUpdateOne {
	srfsuo.fields = append([]string{field}, fields...)
	return srfsuo
}

// Save executes the query and returns the updated SystemRelationshipFeedbackSignal entity.
func (srfsuo *SystemRelationshipFeedbackSignalUpdateOne) Save(ctx context.Context) (*SystemRelationshipFeedbackSignal, error) {
	return withHooks(ctx, srfsuo.sqlSave, srfsuo.mutation, srfsuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (srfsuo *SystemRelationshipFeedbackSignalUpdateOne) SaveX(ctx context.Context) *SystemRelationshipFeedbackSignal {
	node, err := srfsuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (srfsuo *SystemRelationshipFeedbackSignalUpdateOne) Exec(ctx context.Context) error {
	_, err := srfsuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (srfsuo *SystemRelationshipFeedbackSignalUpdateOne) ExecX(ctx context.Context) {
	if err := srfsuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (srfsuo *SystemRelationshipFeedbackSignalUpdateOne) check() error {
	if v, ok := srfsuo.mutation.GetType(); ok {
		if err := systemrelationshipfeedbacksignal.TypeValidator(v); err != nil {
			return &ValidationError{Name: "type", err: fmt.Errorf(`ent: validator failed for field "SystemRelationshipFeedbackSignal.type": %w`, err)}
		}
	}
	if srfsuo.mutation.RelationshipCleared() && len(srfsuo.mutation.RelationshipIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "SystemRelationshipFeedbackSignal.relationship"`)
	}
	if srfsuo.mutation.SignalCleared() && len(srfsuo.mutation.SignalIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "SystemRelationshipFeedbackSignal.signal"`)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (srfsuo *SystemRelationshipFeedbackSignalUpdateOne) Modify(modifiers ...func(u *sql.UpdateBuilder)) *SystemRelationshipFeedbackSignalUpdateOne {
	srfsuo.modifiers = append(srfsuo.modifiers, modifiers...)
	return srfsuo
}

func (srfsuo *SystemRelationshipFeedbackSignalUpdateOne) sqlSave(ctx context.Context) (_node *SystemRelationshipFeedbackSignal, err error) {
	if err := srfsuo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(systemrelationshipfeedbacksignal.Table, systemrelationshipfeedbacksignal.Columns, sqlgraph.NewFieldSpec(systemrelationshipfeedbacksignal.FieldID, field.TypeUUID))
	id, ok := srfsuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "SystemRelationshipFeedbackSignal.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := srfsuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, systemrelationshipfeedbacksignal.FieldID)
		for _, f := range fields {
			if !systemrelationshipfeedbacksignal.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != systemrelationshipfeedbacksignal.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := srfsuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := srfsuo.mutation.GetType(); ok {
		_spec.SetField(systemrelationshipfeedbacksignal.FieldType, field.TypeString, value)
	}
	if value, ok := srfsuo.mutation.Description(); ok {
		_spec.SetField(systemrelationshipfeedbacksignal.FieldDescription, field.TypeString, value)
	}
	if srfsuo.mutation.DescriptionCleared() {
		_spec.ClearField(systemrelationshipfeedbacksignal.FieldDescription, field.TypeString)
	}
	if value, ok := srfsuo.mutation.CreatedAt(); ok {
		_spec.SetField(systemrelationshipfeedbacksignal.FieldCreatedAt, field.TypeTime, value)
	}
	if srfsuo.mutation.RelationshipCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemrelationshipfeedbacksignal.RelationshipTable,
			Columns: []string{systemrelationshipfeedbacksignal.RelationshipColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemrelationship.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := srfsuo.mutation.RelationshipIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemrelationshipfeedbacksignal.RelationshipTable,
			Columns: []string{systemrelationshipfeedbacksignal.RelationshipColumn},
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
	if srfsuo.mutation.SignalCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemrelationshipfeedbacksignal.SignalTable,
			Columns: []string{systemrelationshipfeedbacksignal.SignalColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponentsignal.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := srfsuo.mutation.SignalIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemrelationshipfeedbacksignal.SignalTable,
			Columns: []string{systemrelationshipfeedbacksignal.SignalColumn},
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
	_spec.AddModifiers(srfsuo.modifiers...)
	_node = &SystemRelationshipFeedbackSignal{config: srfsuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, srfsuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{systemrelationshipfeedbacksignal.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	srfsuo.mutation.done = true
	return _node, nil
}
