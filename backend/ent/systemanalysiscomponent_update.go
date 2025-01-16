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
	"github.com/rezible/rezible/ent/systemanalysis"
	"github.com/rezible/rezible/ent/systemanalysiscomponent"
	"github.com/rezible/rezible/ent/systemcomponent"
)

// SystemAnalysisComponentUpdate is the builder for updating SystemAnalysisComponent entities.
type SystemAnalysisComponentUpdate struct {
	config
	hooks     []Hook
	mutation  *SystemAnalysisComponentMutation
	modifiers []func(*sql.UpdateBuilder)
}

// Where appends a list predicates to the SystemAnalysisComponentUpdate builder.
func (sacu *SystemAnalysisComponentUpdate) Where(ps ...predicate.SystemAnalysisComponent) *SystemAnalysisComponentUpdate {
	sacu.mutation.Where(ps...)
	return sacu
}

// SetAnalysisID sets the "analysis_id" field.
func (sacu *SystemAnalysisComponentUpdate) SetAnalysisID(u uuid.UUID) *SystemAnalysisComponentUpdate {
	sacu.mutation.SetAnalysisID(u)
	return sacu
}

// SetNillableAnalysisID sets the "analysis_id" field if the given value is not nil.
func (sacu *SystemAnalysisComponentUpdate) SetNillableAnalysisID(u *uuid.UUID) *SystemAnalysisComponentUpdate {
	if u != nil {
		sacu.SetAnalysisID(*u)
	}
	return sacu
}

// SetComponentID sets the "component_id" field.
func (sacu *SystemAnalysisComponentUpdate) SetComponentID(u uuid.UUID) *SystemAnalysisComponentUpdate {
	sacu.mutation.SetComponentID(u)
	return sacu
}

// SetNillableComponentID sets the "component_id" field if the given value is not nil.
func (sacu *SystemAnalysisComponentUpdate) SetNillableComponentID(u *uuid.UUID) *SystemAnalysisComponentUpdate {
	if u != nil {
		sacu.SetComponentID(*u)
	}
	return sacu
}

// SetDescription sets the "description" field.
func (sacu *SystemAnalysisComponentUpdate) SetDescription(s string) *SystemAnalysisComponentUpdate {
	sacu.mutation.SetDescription(s)
	return sacu
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (sacu *SystemAnalysisComponentUpdate) SetNillableDescription(s *string) *SystemAnalysisComponentUpdate {
	if s != nil {
		sacu.SetDescription(*s)
	}
	return sacu
}

// ClearDescription clears the value of the "description" field.
func (sacu *SystemAnalysisComponentUpdate) ClearDescription() *SystemAnalysisComponentUpdate {
	sacu.mutation.ClearDescription()
	return sacu
}

// SetCreatedAt sets the "created_at" field.
func (sacu *SystemAnalysisComponentUpdate) SetCreatedAt(t time.Time) *SystemAnalysisComponentUpdate {
	sacu.mutation.SetCreatedAt(t)
	return sacu
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (sacu *SystemAnalysisComponentUpdate) SetNillableCreatedAt(t *time.Time) *SystemAnalysisComponentUpdate {
	if t != nil {
		sacu.SetCreatedAt(*t)
	}
	return sacu
}

// SetAnalysis sets the "analysis" edge to the SystemAnalysis entity.
func (sacu *SystemAnalysisComponentUpdate) SetAnalysis(s *SystemAnalysis) *SystemAnalysisComponentUpdate {
	return sacu.SetAnalysisID(s.ID)
}

// SetComponent sets the "component" edge to the SystemComponent entity.
func (sacu *SystemAnalysisComponentUpdate) SetComponent(s *SystemComponent) *SystemAnalysisComponentUpdate {
	return sacu.SetComponentID(s.ID)
}

// Mutation returns the SystemAnalysisComponentMutation object of the builder.
func (sacu *SystemAnalysisComponentUpdate) Mutation() *SystemAnalysisComponentMutation {
	return sacu.mutation
}

// ClearAnalysis clears the "analysis" edge to the SystemAnalysis entity.
func (sacu *SystemAnalysisComponentUpdate) ClearAnalysis() *SystemAnalysisComponentUpdate {
	sacu.mutation.ClearAnalysis()
	return sacu
}

// ClearComponent clears the "component" edge to the SystemComponent entity.
func (sacu *SystemAnalysisComponentUpdate) ClearComponent() *SystemAnalysisComponentUpdate {
	sacu.mutation.ClearComponent()
	return sacu
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (sacu *SystemAnalysisComponentUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, sacu.sqlSave, sacu.mutation, sacu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (sacu *SystemAnalysisComponentUpdate) SaveX(ctx context.Context) int {
	affected, err := sacu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (sacu *SystemAnalysisComponentUpdate) Exec(ctx context.Context) error {
	_, err := sacu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (sacu *SystemAnalysisComponentUpdate) ExecX(ctx context.Context) {
	if err := sacu.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (sacu *SystemAnalysisComponentUpdate) check() error {
	if sacu.mutation.AnalysisCleared() && len(sacu.mutation.AnalysisIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "SystemAnalysisComponent.analysis"`)
	}
	if sacu.mutation.ComponentCleared() && len(sacu.mutation.ComponentIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "SystemAnalysisComponent.component"`)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (sacu *SystemAnalysisComponentUpdate) Modify(modifiers ...func(u *sql.UpdateBuilder)) *SystemAnalysisComponentUpdate {
	sacu.modifiers = append(sacu.modifiers, modifiers...)
	return sacu
}

func (sacu *SystemAnalysisComponentUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := sacu.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(systemanalysiscomponent.Table, systemanalysiscomponent.Columns, sqlgraph.NewFieldSpec(systemanalysiscomponent.FieldID, field.TypeUUID))
	if ps := sacu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := sacu.mutation.Description(); ok {
		_spec.SetField(systemanalysiscomponent.FieldDescription, field.TypeString, value)
	}
	if sacu.mutation.DescriptionCleared() {
		_spec.ClearField(systemanalysiscomponent.FieldDescription, field.TypeString)
	}
	if value, ok := sacu.mutation.CreatedAt(); ok {
		_spec.SetField(systemanalysiscomponent.FieldCreatedAt, field.TypeTime, value)
	}
	if sacu.mutation.AnalysisCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemanalysiscomponent.AnalysisTable,
			Columns: []string{systemanalysiscomponent.AnalysisColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemanalysis.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := sacu.mutation.AnalysisIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemanalysiscomponent.AnalysisTable,
			Columns: []string{systemanalysiscomponent.AnalysisColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemanalysis.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if sacu.mutation.ComponentCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemanalysiscomponent.ComponentTable,
			Columns: []string{systemanalysiscomponent.ComponentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponent.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := sacu.mutation.ComponentIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemanalysiscomponent.ComponentTable,
			Columns: []string{systemanalysiscomponent.ComponentColumn},
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
	_spec.AddModifiers(sacu.modifiers...)
	if n, err = sqlgraph.UpdateNodes(ctx, sacu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{systemanalysiscomponent.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	sacu.mutation.done = true
	return n, nil
}

// SystemAnalysisComponentUpdateOne is the builder for updating a single SystemAnalysisComponent entity.
type SystemAnalysisComponentUpdateOne struct {
	config
	fields    []string
	hooks     []Hook
	mutation  *SystemAnalysisComponentMutation
	modifiers []func(*sql.UpdateBuilder)
}

// SetAnalysisID sets the "analysis_id" field.
func (sacuo *SystemAnalysisComponentUpdateOne) SetAnalysisID(u uuid.UUID) *SystemAnalysisComponentUpdateOne {
	sacuo.mutation.SetAnalysisID(u)
	return sacuo
}

// SetNillableAnalysisID sets the "analysis_id" field if the given value is not nil.
func (sacuo *SystemAnalysisComponentUpdateOne) SetNillableAnalysisID(u *uuid.UUID) *SystemAnalysisComponentUpdateOne {
	if u != nil {
		sacuo.SetAnalysisID(*u)
	}
	return sacuo
}

// SetComponentID sets the "component_id" field.
func (sacuo *SystemAnalysisComponentUpdateOne) SetComponentID(u uuid.UUID) *SystemAnalysisComponentUpdateOne {
	sacuo.mutation.SetComponentID(u)
	return sacuo
}

// SetNillableComponentID sets the "component_id" field if the given value is not nil.
func (sacuo *SystemAnalysisComponentUpdateOne) SetNillableComponentID(u *uuid.UUID) *SystemAnalysisComponentUpdateOne {
	if u != nil {
		sacuo.SetComponentID(*u)
	}
	return sacuo
}

// SetDescription sets the "description" field.
func (sacuo *SystemAnalysisComponentUpdateOne) SetDescription(s string) *SystemAnalysisComponentUpdateOne {
	sacuo.mutation.SetDescription(s)
	return sacuo
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (sacuo *SystemAnalysisComponentUpdateOne) SetNillableDescription(s *string) *SystemAnalysisComponentUpdateOne {
	if s != nil {
		sacuo.SetDescription(*s)
	}
	return sacuo
}

// ClearDescription clears the value of the "description" field.
func (sacuo *SystemAnalysisComponentUpdateOne) ClearDescription() *SystemAnalysisComponentUpdateOne {
	sacuo.mutation.ClearDescription()
	return sacuo
}

// SetCreatedAt sets the "created_at" field.
func (sacuo *SystemAnalysisComponentUpdateOne) SetCreatedAt(t time.Time) *SystemAnalysisComponentUpdateOne {
	sacuo.mutation.SetCreatedAt(t)
	return sacuo
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (sacuo *SystemAnalysisComponentUpdateOne) SetNillableCreatedAt(t *time.Time) *SystemAnalysisComponentUpdateOne {
	if t != nil {
		sacuo.SetCreatedAt(*t)
	}
	return sacuo
}

// SetAnalysis sets the "analysis" edge to the SystemAnalysis entity.
func (sacuo *SystemAnalysisComponentUpdateOne) SetAnalysis(s *SystemAnalysis) *SystemAnalysisComponentUpdateOne {
	return sacuo.SetAnalysisID(s.ID)
}

// SetComponent sets the "component" edge to the SystemComponent entity.
func (sacuo *SystemAnalysisComponentUpdateOne) SetComponent(s *SystemComponent) *SystemAnalysisComponentUpdateOne {
	return sacuo.SetComponentID(s.ID)
}

// Mutation returns the SystemAnalysisComponentMutation object of the builder.
func (sacuo *SystemAnalysisComponentUpdateOne) Mutation() *SystemAnalysisComponentMutation {
	return sacuo.mutation
}

// ClearAnalysis clears the "analysis" edge to the SystemAnalysis entity.
func (sacuo *SystemAnalysisComponentUpdateOne) ClearAnalysis() *SystemAnalysisComponentUpdateOne {
	sacuo.mutation.ClearAnalysis()
	return sacuo
}

// ClearComponent clears the "component" edge to the SystemComponent entity.
func (sacuo *SystemAnalysisComponentUpdateOne) ClearComponent() *SystemAnalysisComponentUpdateOne {
	sacuo.mutation.ClearComponent()
	return sacuo
}

// Where appends a list predicates to the SystemAnalysisComponentUpdate builder.
func (sacuo *SystemAnalysisComponentUpdateOne) Where(ps ...predicate.SystemAnalysisComponent) *SystemAnalysisComponentUpdateOne {
	sacuo.mutation.Where(ps...)
	return sacuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (sacuo *SystemAnalysisComponentUpdateOne) Select(field string, fields ...string) *SystemAnalysisComponentUpdateOne {
	sacuo.fields = append([]string{field}, fields...)
	return sacuo
}

// Save executes the query and returns the updated SystemAnalysisComponent entity.
func (sacuo *SystemAnalysisComponentUpdateOne) Save(ctx context.Context) (*SystemAnalysisComponent, error) {
	return withHooks(ctx, sacuo.sqlSave, sacuo.mutation, sacuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (sacuo *SystemAnalysisComponentUpdateOne) SaveX(ctx context.Context) *SystemAnalysisComponent {
	node, err := sacuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (sacuo *SystemAnalysisComponentUpdateOne) Exec(ctx context.Context) error {
	_, err := sacuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (sacuo *SystemAnalysisComponentUpdateOne) ExecX(ctx context.Context) {
	if err := sacuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (sacuo *SystemAnalysisComponentUpdateOne) check() error {
	if sacuo.mutation.AnalysisCleared() && len(sacuo.mutation.AnalysisIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "SystemAnalysisComponent.analysis"`)
	}
	if sacuo.mutation.ComponentCleared() && len(sacuo.mutation.ComponentIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "SystemAnalysisComponent.component"`)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (sacuo *SystemAnalysisComponentUpdateOne) Modify(modifiers ...func(u *sql.UpdateBuilder)) *SystemAnalysisComponentUpdateOne {
	sacuo.modifiers = append(sacuo.modifiers, modifiers...)
	return sacuo
}

func (sacuo *SystemAnalysisComponentUpdateOne) sqlSave(ctx context.Context) (_node *SystemAnalysisComponent, err error) {
	if err := sacuo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(systemanalysiscomponent.Table, systemanalysiscomponent.Columns, sqlgraph.NewFieldSpec(systemanalysiscomponent.FieldID, field.TypeUUID))
	id, ok := sacuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "SystemAnalysisComponent.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := sacuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, systemanalysiscomponent.FieldID)
		for _, f := range fields {
			if !systemanalysiscomponent.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != systemanalysiscomponent.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := sacuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := sacuo.mutation.Description(); ok {
		_spec.SetField(systemanalysiscomponent.FieldDescription, field.TypeString, value)
	}
	if sacuo.mutation.DescriptionCleared() {
		_spec.ClearField(systemanalysiscomponent.FieldDescription, field.TypeString)
	}
	if value, ok := sacuo.mutation.CreatedAt(); ok {
		_spec.SetField(systemanalysiscomponent.FieldCreatedAt, field.TypeTime, value)
	}
	if sacuo.mutation.AnalysisCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemanalysiscomponent.AnalysisTable,
			Columns: []string{systemanalysiscomponent.AnalysisColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemanalysis.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := sacuo.mutation.AnalysisIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemanalysiscomponent.AnalysisTable,
			Columns: []string{systemanalysiscomponent.AnalysisColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemanalysis.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if sacuo.mutation.ComponentCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemanalysiscomponent.ComponentTable,
			Columns: []string{systemanalysiscomponent.ComponentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponent.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := sacuo.mutation.ComponentIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemanalysiscomponent.ComponentTable,
			Columns: []string{systemanalysiscomponent.ComponentColumn},
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
	_spec.AddModifiers(sacuo.modifiers...)
	_node = &SystemAnalysisComponent{config: sacuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, sacuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{systemanalysiscomponent.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	sacuo.mutation.done = true
	return _node, nil
}
