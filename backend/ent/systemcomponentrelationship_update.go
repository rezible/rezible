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
	"github.com/rezible/rezible/ent/systemanalysisrelationship"
	"github.com/rezible/rezible/ent/systemcomponent"
	"github.com/rezible/rezible/ent/systemcomponentrelationship"
)

// SystemComponentRelationshipUpdate is the builder for updating SystemComponentRelationship entities.
type SystemComponentRelationshipUpdate struct {
	config
	hooks     []Hook
	mutation  *SystemComponentRelationshipMutation
	modifiers []func(*sql.UpdateBuilder)
}

// Where appends a list predicates to the SystemComponentRelationshipUpdate builder.
func (scru *SystemComponentRelationshipUpdate) Where(ps ...predicate.SystemComponentRelationship) *SystemComponentRelationshipUpdate {
	scru.mutation.Where(ps...)
	return scru
}

// SetProviderID sets the "provider_id" field.
func (scru *SystemComponentRelationshipUpdate) SetProviderID(s string) *SystemComponentRelationshipUpdate {
	scru.mutation.SetProviderID(s)
	return scru
}

// SetNillableProviderID sets the "provider_id" field if the given value is not nil.
func (scru *SystemComponentRelationshipUpdate) SetNillableProviderID(s *string) *SystemComponentRelationshipUpdate {
	if s != nil {
		scru.SetProviderID(*s)
	}
	return scru
}

// ClearProviderID clears the value of the "provider_id" field.
func (scru *SystemComponentRelationshipUpdate) ClearProviderID() *SystemComponentRelationshipUpdate {
	scru.mutation.ClearProviderID()
	return scru
}

// SetSourceID sets the "source_id" field.
func (scru *SystemComponentRelationshipUpdate) SetSourceID(u uuid.UUID) *SystemComponentRelationshipUpdate {
	scru.mutation.SetSourceID(u)
	return scru
}

// SetNillableSourceID sets the "source_id" field if the given value is not nil.
func (scru *SystemComponentRelationshipUpdate) SetNillableSourceID(u *uuid.UUID) *SystemComponentRelationshipUpdate {
	if u != nil {
		scru.SetSourceID(*u)
	}
	return scru
}

// SetTargetID sets the "target_id" field.
func (scru *SystemComponentRelationshipUpdate) SetTargetID(u uuid.UUID) *SystemComponentRelationshipUpdate {
	scru.mutation.SetTargetID(u)
	return scru
}

// SetNillableTargetID sets the "target_id" field if the given value is not nil.
func (scru *SystemComponentRelationshipUpdate) SetNillableTargetID(u *uuid.UUID) *SystemComponentRelationshipUpdate {
	if u != nil {
		scru.SetTargetID(*u)
	}
	return scru
}

// SetDescription sets the "description" field.
func (scru *SystemComponentRelationshipUpdate) SetDescription(s string) *SystemComponentRelationshipUpdate {
	scru.mutation.SetDescription(s)
	return scru
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (scru *SystemComponentRelationshipUpdate) SetNillableDescription(s *string) *SystemComponentRelationshipUpdate {
	if s != nil {
		scru.SetDescription(*s)
	}
	return scru
}

// ClearDescription clears the value of the "description" field.
func (scru *SystemComponentRelationshipUpdate) ClearDescription() *SystemComponentRelationshipUpdate {
	scru.mutation.ClearDescription()
	return scru
}

// SetCreatedAt sets the "created_at" field.
func (scru *SystemComponentRelationshipUpdate) SetCreatedAt(t time.Time) *SystemComponentRelationshipUpdate {
	scru.mutation.SetCreatedAt(t)
	return scru
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (scru *SystemComponentRelationshipUpdate) SetNillableCreatedAt(t *time.Time) *SystemComponentRelationshipUpdate {
	if t != nil {
		scru.SetCreatedAt(*t)
	}
	return scru
}

// SetSource sets the "source" edge to the SystemComponent entity.
func (scru *SystemComponentRelationshipUpdate) SetSource(s *SystemComponent) *SystemComponentRelationshipUpdate {
	return scru.SetSourceID(s.ID)
}

// SetTarget sets the "target" edge to the SystemComponent entity.
func (scru *SystemComponentRelationshipUpdate) SetTarget(s *SystemComponent) *SystemComponentRelationshipUpdate {
	return scru.SetTargetID(s.ID)
}

// AddSystemAnalysisIDs adds the "system_analyses" edge to the SystemAnalysisRelationship entity by IDs.
func (scru *SystemComponentRelationshipUpdate) AddSystemAnalysisIDs(ids ...uuid.UUID) *SystemComponentRelationshipUpdate {
	scru.mutation.AddSystemAnalysisIDs(ids...)
	return scru
}

// AddSystemAnalyses adds the "system_analyses" edges to the SystemAnalysisRelationship entity.
func (scru *SystemComponentRelationshipUpdate) AddSystemAnalyses(s ...*SystemAnalysisRelationship) *SystemComponentRelationshipUpdate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return scru.AddSystemAnalysisIDs(ids...)
}

// Mutation returns the SystemComponentRelationshipMutation object of the builder.
func (scru *SystemComponentRelationshipUpdate) Mutation() *SystemComponentRelationshipMutation {
	return scru.mutation
}

// ClearSource clears the "source" edge to the SystemComponent entity.
func (scru *SystemComponentRelationshipUpdate) ClearSource() *SystemComponentRelationshipUpdate {
	scru.mutation.ClearSource()
	return scru
}

// ClearTarget clears the "target" edge to the SystemComponent entity.
func (scru *SystemComponentRelationshipUpdate) ClearTarget() *SystemComponentRelationshipUpdate {
	scru.mutation.ClearTarget()
	return scru
}

// ClearSystemAnalyses clears all "system_analyses" edges to the SystemAnalysisRelationship entity.
func (scru *SystemComponentRelationshipUpdate) ClearSystemAnalyses() *SystemComponentRelationshipUpdate {
	scru.mutation.ClearSystemAnalyses()
	return scru
}

// RemoveSystemAnalysisIDs removes the "system_analyses" edge to SystemAnalysisRelationship entities by IDs.
func (scru *SystemComponentRelationshipUpdate) RemoveSystemAnalysisIDs(ids ...uuid.UUID) *SystemComponentRelationshipUpdate {
	scru.mutation.RemoveSystemAnalysisIDs(ids...)
	return scru
}

// RemoveSystemAnalyses removes "system_analyses" edges to SystemAnalysisRelationship entities.
func (scru *SystemComponentRelationshipUpdate) RemoveSystemAnalyses(s ...*SystemAnalysisRelationship) *SystemComponentRelationshipUpdate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return scru.RemoveSystemAnalysisIDs(ids...)
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (scru *SystemComponentRelationshipUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, scru.sqlSave, scru.mutation, scru.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (scru *SystemComponentRelationshipUpdate) SaveX(ctx context.Context) int {
	affected, err := scru.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (scru *SystemComponentRelationshipUpdate) Exec(ctx context.Context) error {
	_, err := scru.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (scru *SystemComponentRelationshipUpdate) ExecX(ctx context.Context) {
	if err := scru.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (scru *SystemComponentRelationshipUpdate) check() error {
	if scru.mutation.SourceCleared() && len(scru.mutation.SourceIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "SystemComponentRelationship.source"`)
	}
	if scru.mutation.TargetCleared() && len(scru.mutation.TargetIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "SystemComponentRelationship.target"`)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (scru *SystemComponentRelationshipUpdate) Modify(modifiers ...func(u *sql.UpdateBuilder)) *SystemComponentRelationshipUpdate {
	scru.modifiers = append(scru.modifiers, modifiers...)
	return scru
}

func (scru *SystemComponentRelationshipUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := scru.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(systemcomponentrelationship.Table, systemcomponentrelationship.Columns, sqlgraph.NewFieldSpec(systemcomponentrelationship.FieldID, field.TypeUUID))
	if ps := scru.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := scru.mutation.ProviderID(); ok {
		_spec.SetField(systemcomponentrelationship.FieldProviderID, field.TypeString, value)
	}
	if scru.mutation.ProviderIDCleared() {
		_spec.ClearField(systemcomponentrelationship.FieldProviderID, field.TypeString)
	}
	if value, ok := scru.mutation.Description(); ok {
		_spec.SetField(systemcomponentrelationship.FieldDescription, field.TypeString, value)
	}
	if scru.mutation.DescriptionCleared() {
		_spec.ClearField(systemcomponentrelationship.FieldDescription, field.TypeString)
	}
	if value, ok := scru.mutation.CreatedAt(); ok {
		_spec.SetField(systemcomponentrelationship.FieldCreatedAt, field.TypeTime, value)
	}
	if scru.mutation.SourceCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemcomponentrelationship.SourceTable,
			Columns: []string{systemcomponentrelationship.SourceColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponent.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := scru.mutation.SourceIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemcomponentrelationship.SourceTable,
			Columns: []string{systemcomponentrelationship.SourceColumn},
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
	if scru.mutation.TargetCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemcomponentrelationship.TargetTable,
			Columns: []string{systemcomponentrelationship.TargetColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponent.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := scru.mutation.TargetIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemcomponentrelationship.TargetTable,
			Columns: []string{systemcomponentrelationship.TargetColumn},
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
	if scru.mutation.SystemAnalysesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   systemcomponentrelationship.SystemAnalysesTable,
			Columns: []string{systemcomponentrelationship.SystemAnalysesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemanalysisrelationship.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := scru.mutation.RemovedSystemAnalysesIDs(); len(nodes) > 0 && !scru.mutation.SystemAnalysesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   systemcomponentrelationship.SystemAnalysesTable,
			Columns: []string{systemcomponentrelationship.SystemAnalysesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemanalysisrelationship.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := scru.mutation.SystemAnalysesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   systemcomponentrelationship.SystemAnalysesTable,
			Columns: []string{systemcomponentrelationship.SystemAnalysesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemanalysisrelationship.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(scru.modifiers...)
	if n, err = sqlgraph.UpdateNodes(ctx, scru.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{systemcomponentrelationship.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	scru.mutation.done = true
	return n, nil
}

// SystemComponentRelationshipUpdateOne is the builder for updating a single SystemComponentRelationship entity.
type SystemComponentRelationshipUpdateOne struct {
	config
	fields    []string
	hooks     []Hook
	mutation  *SystemComponentRelationshipMutation
	modifiers []func(*sql.UpdateBuilder)
}

// SetProviderID sets the "provider_id" field.
func (scruo *SystemComponentRelationshipUpdateOne) SetProviderID(s string) *SystemComponentRelationshipUpdateOne {
	scruo.mutation.SetProviderID(s)
	return scruo
}

// SetNillableProviderID sets the "provider_id" field if the given value is not nil.
func (scruo *SystemComponentRelationshipUpdateOne) SetNillableProviderID(s *string) *SystemComponentRelationshipUpdateOne {
	if s != nil {
		scruo.SetProviderID(*s)
	}
	return scruo
}

// ClearProviderID clears the value of the "provider_id" field.
func (scruo *SystemComponentRelationshipUpdateOne) ClearProviderID() *SystemComponentRelationshipUpdateOne {
	scruo.mutation.ClearProviderID()
	return scruo
}

// SetSourceID sets the "source_id" field.
func (scruo *SystemComponentRelationshipUpdateOne) SetSourceID(u uuid.UUID) *SystemComponentRelationshipUpdateOne {
	scruo.mutation.SetSourceID(u)
	return scruo
}

// SetNillableSourceID sets the "source_id" field if the given value is not nil.
func (scruo *SystemComponentRelationshipUpdateOne) SetNillableSourceID(u *uuid.UUID) *SystemComponentRelationshipUpdateOne {
	if u != nil {
		scruo.SetSourceID(*u)
	}
	return scruo
}

// SetTargetID sets the "target_id" field.
func (scruo *SystemComponentRelationshipUpdateOne) SetTargetID(u uuid.UUID) *SystemComponentRelationshipUpdateOne {
	scruo.mutation.SetTargetID(u)
	return scruo
}

// SetNillableTargetID sets the "target_id" field if the given value is not nil.
func (scruo *SystemComponentRelationshipUpdateOne) SetNillableTargetID(u *uuid.UUID) *SystemComponentRelationshipUpdateOne {
	if u != nil {
		scruo.SetTargetID(*u)
	}
	return scruo
}

// SetDescription sets the "description" field.
func (scruo *SystemComponentRelationshipUpdateOne) SetDescription(s string) *SystemComponentRelationshipUpdateOne {
	scruo.mutation.SetDescription(s)
	return scruo
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (scruo *SystemComponentRelationshipUpdateOne) SetNillableDescription(s *string) *SystemComponentRelationshipUpdateOne {
	if s != nil {
		scruo.SetDescription(*s)
	}
	return scruo
}

// ClearDescription clears the value of the "description" field.
func (scruo *SystemComponentRelationshipUpdateOne) ClearDescription() *SystemComponentRelationshipUpdateOne {
	scruo.mutation.ClearDescription()
	return scruo
}

// SetCreatedAt sets the "created_at" field.
func (scruo *SystemComponentRelationshipUpdateOne) SetCreatedAt(t time.Time) *SystemComponentRelationshipUpdateOne {
	scruo.mutation.SetCreatedAt(t)
	return scruo
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (scruo *SystemComponentRelationshipUpdateOne) SetNillableCreatedAt(t *time.Time) *SystemComponentRelationshipUpdateOne {
	if t != nil {
		scruo.SetCreatedAt(*t)
	}
	return scruo
}

// SetSource sets the "source" edge to the SystemComponent entity.
func (scruo *SystemComponentRelationshipUpdateOne) SetSource(s *SystemComponent) *SystemComponentRelationshipUpdateOne {
	return scruo.SetSourceID(s.ID)
}

// SetTarget sets the "target" edge to the SystemComponent entity.
func (scruo *SystemComponentRelationshipUpdateOne) SetTarget(s *SystemComponent) *SystemComponentRelationshipUpdateOne {
	return scruo.SetTargetID(s.ID)
}

// AddSystemAnalysisIDs adds the "system_analyses" edge to the SystemAnalysisRelationship entity by IDs.
func (scruo *SystemComponentRelationshipUpdateOne) AddSystemAnalysisIDs(ids ...uuid.UUID) *SystemComponentRelationshipUpdateOne {
	scruo.mutation.AddSystemAnalysisIDs(ids...)
	return scruo
}

// AddSystemAnalyses adds the "system_analyses" edges to the SystemAnalysisRelationship entity.
func (scruo *SystemComponentRelationshipUpdateOne) AddSystemAnalyses(s ...*SystemAnalysisRelationship) *SystemComponentRelationshipUpdateOne {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return scruo.AddSystemAnalysisIDs(ids...)
}

// Mutation returns the SystemComponentRelationshipMutation object of the builder.
func (scruo *SystemComponentRelationshipUpdateOne) Mutation() *SystemComponentRelationshipMutation {
	return scruo.mutation
}

// ClearSource clears the "source" edge to the SystemComponent entity.
func (scruo *SystemComponentRelationshipUpdateOne) ClearSource() *SystemComponentRelationshipUpdateOne {
	scruo.mutation.ClearSource()
	return scruo
}

// ClearTarget clears the "target" edge to the SystemComponent entity.
func (scruo *SystemComponentRelationshipUpdateOne) ClearTarget() *SystemComponentRelationshipUpdateOne {
	scruo.mutation.ClearTarget()
	return scruo
}

// ClearSystemAnalyses clears all "system_analyses" edges to the SystemAnalysisRelationship entity.
func (scruo *SystemComponentRelationshipUpdateOne) ClearSystemAnalyses() *SystemComponentRelationshipUpdateOne {
	scruo.mutation.ClearSystemAnalyses()
	return scruo
}

// RemoveSystemAnalysisIDs removes the "system_analyses" edge to SystemAnalysisRelationship entities by IDs.
func (scruo *SystemComponentRelationshipUpdateOne) RemoveSystemAnalysisIDs(ids ...uuid.UUID) *SystemComponentRelationshipUpdateOne {
	scruo.mutation.RemoveSystemAnalysisIDs(ids...)
	return scruo
}

// RemoveSystemAnalyses removes "system_analyses" edges to SystemAnalysisRelationship entities.
func (scruo *SystemComponentRelationshipUpdateOne) RemoveSystemAnalyses(s ...*SystemAnalysisRelationship) *SystemComponentRelationshipUpdateOne {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return scruo.RemoveSystemAnalysisIDs(ids...)
}

// Where appends a list predicates to the SystemComponentRelationshipUpdate builder.
func (scruo *SystemComponentRelationshipUpdateOne) Where(ps ...predicate.SystemComponentRelationship) *SystemComponentRelationshipUpdateOne {
	scruo.mutation.Where(ps...)
	return scruo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (scruo *SystemComponentRelationshipUpdateOne) Select(field string, fields ...string) *SystemComponentRelationshipUpdateOne {
	scruo.fields = append([]string{field}, fields...)
	return scruo
}

// Save executes the query and returns the updated SystemComponentRelationship entity.
func (scruo *SystemComponentRelationshipUpdateOne) Save(ctx context.Context) (*SystemComponentRelationship, error) {
	return withHooks(ctx, scruo.sqlSave, scruo.mutation, scruo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (scruo *SystemComponentRelationshipUpdateOne) SaveX(ctx context.Context) *SystemComponentRelationship {
	node, err := scruo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (scruo *SystemComponentRelationshipUpdateOne) Exec(ctx context.Context) error {
	_, err := scruo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (scruo *SystemComponentRelationshipUpdateOne) ExecX(ctx context.Context) {
	if err := scruo.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (scruo *SystemComponentRelationshipUpdateOne) check() error {
	if scruo.mutation.SourceCleared() && len(scruo.mutation.SourceIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "SystemComponentRelationship.source"`)
	}
	if scruo.mutation.TargetCleared() && len(scruo.mutation.TargetIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "SystemComponentRelationship.target"`)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (scruo *SystemComponentRelationshipUpdateOne) Modify(modifiers ...func(u *sql.UpdateBuilder)) *SystemComponentRelationshipUpdateOne {
	scruo.modifiers = append(scruo.modifiers, modifiers...)
	return scruo
}

func (scruo *SystemComponentRelationshipUpdateOne) sqlSave(ctx context.Context) (_node *SystemComponentRelationship, err error) {
	if err := scruo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(systemcomponentrelationship.Table, systemcomponentrelationship.Columns, sqlgraph.NewFieldSpec(systemcomponentrelationship.FieldID, field.TypeUUID))
	id, ok := scruo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "SystemComponentRelationship.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := scruo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, systemcomponentrelationship.FieldID)
		for _, f := range fields {
			if !systemcomponentrelationship.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != systemcomponentrelationship.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := scruo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := scruo.mutation.ProviderID(); ok {
		_spec.SetField(systemcomponentrelationship.FieldProviderID, field.TypeString, value)
	}
	if scruo.mutation.ProviderIDCleared() {
		_spec.ClearField(systemcomponentrelationship.FieldProviderID, field.TypeString)
	}
	if value, ok := scruo.mutation.Description(); ok {
		_spec.SetField(systemcomponentrelationship.FieldDescription, field.TypeString, value)
	}
	if scruo.mutation.DescriptionCleared() {
		_spec.ClearField(systemcomponentrelationship.FieldDescription, field.TypeString)
	}
	if value, ok := scruo.mutation.CreatedAt(); ok {
		_spec.SetField(systemcomponentrelationship.FieldCreatedAt, field.TypeTime, value)
	}
	if scruo.mutation.SourceCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemcomponentrelationship.SourceTable,
			Columns: []string{systemcomponentrelationship.SourceColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponent.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := scruo.mutation.SourceIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemcomponentrelationship.SourceTable,
			Columns: []string{systemcomponentrelationship.SourceColumn},
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
	if scruo.mutation.TargetCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemcomponentrelationship.TargetTable,
			Columns: []string{systemcomponentrelationship.TargetColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponent.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := scruo.mutation.TargetIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   systemcomponentrelationship.TargetTable,
			Columns: []string{systemcomponentrelationship.TargetColumn},
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
	if scruo.mutation.SystemAnalysesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   systemcomponentrelationship.SystemAnalysesTable,
			Columns: []string{systemcomponentrelationship.SystemAnalysesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemanalysisrelationship.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := scruo.mutation.RemovedSystemAnalysesIDs(); len(nodes) > 0 && !scruo.mutation.SystemAnalysesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   systemcomponentrelationship.SystemAnalysesTable,
			Columns: []string{systemcomponentrelationship.SystemAnalysesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemanalysisrelationship.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := scruo.mutation.SystemAnalysesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   systemcomponentrelationship.SystemAnalysesTable,
			Columns: []string{systemcomponentrelationship.SystemAnalysesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemanalysisrelationship.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(scruo.modifiers...)
	_node = &SystemComponentRelationship{config: scruo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, scruo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{systemcomponentrelationship.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	scruo.mutation.done = true
	return _node, nil
}
