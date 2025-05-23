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
	"github.com/rezible/rezible/ent/systemcomponentrelationship"
	"github.com/rezible/rezible/ent/systemhazard"
)

// SystemHazardUpdate is the builder for updating SystemHazard entities.
type SystemHazardUpdate struct {
	config
	hooks     []Hook
	mutation  *SystemHazardMutation
	modifiers []func(*sql.UpdateBuilder)
}

// Where appends a list predicates to the SystemHazardUpdate builder.
func (shu *SystemHazardUpdate) Where(ps ...predicate.SystemHazard) *SystemHazardUpdate {
	shu.mutation.Where(ps...)
	return shu
}

// SetName sets the "name" field.
func (shu *SystemHazardUpdate) SetName(s string) *SystemHazardUpdate {
	shu.mutation.SetName(s)
	return shu
}

// SetNillableName sets the "name" field if the given value is not nil.
func (shu *SystemHazardUpdate) SetNillableName(s *string) *SystemHazardUpdate {
	if s != nil {
		shu.SetName(*s)
	}
	return shu
}

// SetDescription sets the "description" field.
func (shu *SystemHazardUpdate) SetDescription(s string) *SystemHazardUpdate {
	shu.mutation.SetDescription(s)
	return shu
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (shu *SystemHazardUpdate) SetNillableDescription(s *string) *SystemHazardUpdate {
	if s != nil {
		shu.SetDescription(*s)
	}
	return shu
}

// SetCreatedAt sets the "created_at" field.
func (shu *SystemHazardUpdate) SetCreatedAt(t time.Time) *SystemHazardUpdate {
	shu.mutation.SetCreatedAt(t)
	return shu
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (shu *SystemHazardUpdate) SetNillableCreatedAt(t *time.Time) *SystemHazardUpdate {
	if t != nil {
		shu.SetCreatedAt(*t)
	}
	return shu
}

// SetUpdatedAt sets the "updated_at" field.
func (shu *SystemHazardUpdate) SetUpdatedAt(t time.Time) *SystemHazardUpdate {
	shu.mutation.SetUpdatedAt(t)
	return shu
}

// AddComponentIDs adds the "components" edge to the SystemComponent entity by IDs.
func (shu *SystemHazardUpdate) AddComponentIDs(ids ...uuid.UUID) *SystemHazardUpdate {
	shu.mutation.AddComponentIDs(ids...)
	return shu
}

// AddComponents adds the "components" edges to the SystemComponent entity.
func (shu *SystemHazardUpdate) AddComponents(s ...*SystemComponent) *SystemHazardUpdate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return shu.AddComponentIDs(ids...)
}

// AddConstraintIDs adds the "constraints" edge to the SystemComponentConstraint entity by IDs.
func (shu *SystemHazardUpdate) AddConstraintIDs(ids ...uuid.UUID) *SystemHazardUpdate {
	shu.mutation.AddConstraintIDs(ids...)
	return shu
}

// AddConstraints adds the "constraints" edges to the SystemComponentConstraint entity.
func (shu *SystemHazardUpdate) AddConstraints(s ...*SystemComponentConstraint) *SystemHazardUpdate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return shu.AddConstraintIDs(ids...)
}

// AddRelationshipIDs adds the "relationships" edge to the SystemComponentRelationship entity by IDs.
func (shu *SystemHazardUpdate) AddRelationshipIDs(ids ...uuid.UUID) *SystemHazardUpdate {
	shu.mutation.AddRelationshipIDs(ids...)
	return shu
}

// AddRelationships adds the "relationships" edges to the SystemComponentRelationship entity.
func (shu *SystemHazardUpdate) AddRelationships(s ...*SystemComponentRelationship) *SystemHazardUpdate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return shu.AddRelationshipIDs(ids...)
}

// Mutation returns the SystemHazardMutation object of the builder.
func (shu *SystemHazardUpdate) Mutation() *SystemHazardMutation {
	return shu.mutation
}

// ClearComponents clears all "components" edges to the SystemComponent entity.
func (shu *SystemHazardUpdate) ClearComponents() *SystemHazardUpdate {
	shu.mutation.ClearComponents()
	return shu
}

// RemoveComponentIDs removes the "components" edge to SystemComponent entities by IDs.
func (shu *SystemHazardUpdate) RemoveComponentIDs(ids ...uuid.UUID) *SystemHazardUpdate {
	shu.mutation.RemoveComponentIDs(ids...)
	return shu
}

// RemoveComponents removes "components" edges to SystemComponent entities.
func (shu *SystemHazardUpdate) RemoveComponents(s ...*SystemComponent) *SystemHazardUpdate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return shu.RemoveComponentIDs(ids...)
}

// ClearConstraints clears all "constraints" edges to the SystemComponentConstraint entity.
func (shu *SystemHazardUpdate) ClearConstraints() *SystemHazardUpdate {
	shu.mutation.ClearConstraints()
	return shu
}

// RemoveConstraintIDs removes the "constraints" edge to SystemComponentConstraint entities by IDs.
func (shu *SystemHazardUpdate) RemoveConstraintIDs(ids ...uuid.UUID) *SystemHazardUpdate {
	shu.mutation.RemoveConstraintIDs(ids...)
	return shu
}

// RemoveConstraints removes "constraints" edges to SystemComponentConstraint entities.
func (shu *SystemHazardUpdate) RemoveConstraints(s ...*SystemComponentConstraint) *SystemHazardUpdate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return shu.RemoveConstraintIDs(ids...)
}

// ClearRelationships clears all "relationships" edges to the SystemComponentRelationship entity.
func (shu *SystemHazardUpdate) ClearRelationships() *SystemHazardUpdate {
	shu.mutation.ClearRelationships()
	return shu
}

// RemoveRelationshipIDs removes the "relationships" edge to SystemComponentRelationship entities by IDs.
func (shu *SystemHazardUpdate) RemoveRelationshipIDs(ids ...uuid.UUID) *SystemHazardUpdate {
	shu.mutation.RemoveRelationshipIDs(ids...)
	return shu
}

// RemoveRelationships removes "relationships" edges to SystemComponentRelationship entities.
func (shu *SystemHazardUpdate) RemoveRelationships(s ...*SystemComponentRelationship) *SystemHazardUpdate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return shu.RemoveRelationshipIDs(ids...)
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (shu *SystemHazardUpdate) Save(ctx context.Context) (int, error) {
	shu.defaults()
	return withHooks(ctx, shu.sqlSave, shu.mutation, shu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (shu *SystemHazardUpdate) SaveX(ctx context.Context) int {
	affected, err := shu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (shu *SystemHazardUpdate) Exec(ctx context.Context) error {
	_, err := shu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (shu *SystemHazardUpdate) ExecX(ctx context.Context) {
	if err := shu.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (shu *SystemHazardUpdate) defaults() {
	if _, ok := shu.mutation.UpdatedAt(); !ok {
		v := systemhazard.UpdateDefaultUpdatedAt()
		shu.mutation.SetUpdatedAt(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (shu *SystemHazardUpdate) check() error {
	if v, ok := shu.mutation.Name(); ok {
		if err := systemhazard.NameValidator(v); err != nil {
			return &ValidationError{Name: "name", err: fmt.Errorf(`ent: validator failed for field "SystemHazard.name": %w`, err)}
		}
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (shu *SystemHazardUpdate) Modify(modifiers ...func(u *sql.UpdateBuilder)) *SystemHazardUpdate {
	shu.modifiers = append(shu.modifiers, modifiers...)
	return shu
}

func (shu *SystemHazardUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := shu.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(systemhazard.Table, systemhazard.Columns, sqlgraph.NewFieldSpec(systemhazard.FieldID, field.TypeUUID))
	if ps := shu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := shu.mutation.Name(); ok {
		_spec.SetField(systemhazard.FieldName, field.TypeString, value)
	}
	if value, ok := shu.mutation.Description(); ok {
		_spec.SetField(systemhazard.FieldDescription, field.TypeString, value)
	}
	if value, ok := shu.mutation.CreatedAt(); ok {
		_spec.SetField(systemhazard.FieldCreatedAt, field.TypeTime, value)
	}
	if value, ok := shu.mutation.UpdatedAt(); ok {
		_spec.SetField(systemhazard.FieldUpdatedAt, field.TypeTime, value)
	}
	if shu.mutation.ComponentsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   systemhazard.ComponentsTable,
			Columns: systemhazard.ComponentsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponent.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := shu.mutation.RemovedComponentsIDs(); len(nodes) > 0 && !shu.mutation.ComponentsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   systemhazard.ComponentsTable,
			Columns: systemhazard.ComponentsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponent.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := shu.mutation.ComponentsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   systemhazard.ComponentsTable,
			Columns: systemhazard.ComponentsPrimaryKey,
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
	if shu.mutation.ConstraintsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   systemhazard.ConstraintsTable,
			Columns: systemhazard.ConstraintsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponentconstraint.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := shu.mutation.RemovedConstraintsIDs(); len(nodes) > 0 && !shu.mutation.ConstraintsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   systemhazard.ConstraintsTable,
			Columns: systemhazard.ConstraintsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponentconstraint.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := shu.mutation.ConstraintsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   systemhazard.ConstraintsTable,
			Columns: systemhazard.ConstraintsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponentconstraint.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if shu.mutation.RelationshipsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   systemhazard.RelationshipsTable,
			Columns: systemhazard.RelationshipsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponentrelationship.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := shu.mutation.RemovedRelationshipsIDs(); len(nodes) > 0 && !shu.mutation.RelationshipsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   systemhazard.RelationshipsTable,
			Columns: systemhazard.RelationshipsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponentrelationship.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := shu.mutation.RelationshipsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   systemhazard.RelationshipsTable,
			Columns: systemhazard.RelationshipsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponentrelationship.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(shu.modifiers...)
	if n, err = sqlgraph.UpdateNodes(ctx, shu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{systemhazard.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	shu.mutation.done = true
	return n, nil
}

// SystemHazardUpdateOne is the builder for updating a single SystemHazard entity.
type SystemHazardUpdateOne struct {
	config
	fields    []string
	hooks     []Hook
	mutation  *SystemHazardMutation
	modifiers []func(*sql.UpdateBuilder)
}

// SetName sets the "name" field.
func (shuo *SystemHazardUpdateOne) SetName(s string) *SystemHazardUpdateOne {
	shuo.mutation.SetName(s)
	return shuo
}

// SetNillableName sets the "name" field if the given value is not nil.
func (shuo *SystemHazardUpdateOne) SetNillableName(s *string) *SystemHazardUpdateOne {
	if s != nil {
		shuo.SetName(*s)
	}
	return shuo
}

// SetDescription sets the "description" field.
func (shuo *SystemHazardUpdateOne) SetDescription(s string) *SystemHazardUpdateOne {
	shuo.mutation.SetDescription(s)
	return shuo
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (shuo *SystemHazardUpdateOne) SetNillableDescription(s *string) *SystemHazardUpdateOne {
	if s != nil {
		shuo.SetDescription(*s)
	}
	return shuo
}

// SetCreatedAt sets the "created_at" field.
func (shuo *SystemHazardUpdateOne) SetCreatedAt(t time.Time) *SystemHazardUpdateOne {
	shuo.mutation.SetCreatedAt(t)
	return shuo
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (shuo *SystemHazardUpdateOne) SetNillableCreatedAt(t *time.Time) *SystemHazardUpdateOne {
	if t != nil {
		shuo.SetCreatedAt(*t)
	}
	return shuo
}

// SetUpdatedAt sets the "updated_at" field.
func (shuo *SystemHazardUpdateOne) SetUpdatedAt(t time.Time) *SystemHazardUpdateOne {
	shuo.mutation.SetUpdatedAt(t)
	return shuo
}

// AddComponentIDs adds the "components" edge to the SystemComponent entity by IDs.
func (shuo *SystemHazardUpdateOne) AddComponentIDs(ids ...uuid.UUID) *SystemHazardUpdateOne {
	shuo.mutation.AddComponentIDs(ids...)
	return shuo
}

// AddComponents adds the "components" edges to the SystemComponent entity.
func (shuo *SystemHazardUpdateOne) AddComponents(s ...*SystemComponent) *SystemHazardUpdateOne {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return shuo.AddComponentIDs(ids...)
}

// AddConstraintIDs adds the "constraints" edge to the SystemComponentConstraint entity by IDs.
func (shuo *SystemHazardUpdateOne) AddConstraintIDs(ids ...uuid.UUID) *SystemHazardUpdateOne {
	shuo.mutation.AddConstraintIDs(ids...)
	return shuo
}

// AddConstraints adds the "constraints" edges to the SystemComponentConstraint entity.
func (shuo *SystemHazardUpdateOne) AddConstraints(s ...*SystemComponentConstraint) *SystemHazardUpdateOne {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return shuo.AddConstraintIDs(ids...)
}

// AddRelationshipIDs adds the "relationships" edge to the SystemComponentRelationship entity by IDs.
func (shuo *SystemHazardUpdateOne) AddRelationshipIDs(ids ...uuid.UUID) *SystemHazardUpdateOne {
	shuo.mutation.AddRelationshipIDs(ids...)
	return shuo
}

// AddRelationships adds the "relationships" edges to the SystemComponentRelationship entity.
func (shuo *SystemHazardUpdateOne) AddRelationships(s ...*SystemComponentRelationship) *SystemHazardUpdateOne {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return shuo.AddRelationshipIDs(ids...)
}

// Mutation returns the SystemHazardMutation object of the builder.
func (shuo *SystemHazardUpdateOne) Mutation() *SystemHazardMutation {
	return shuo.mutation
}

// ClearComponents clears all "components" edges to the SystemComponent entity.
func (shuo *SystemHazardUpdateOne) ClearComponents() *SystemHazardUpdateOne {
	shuo.mutation.ClearComponents()
	return shuo
}

// RemoveComponentIDs removes the "components" edge to SystemComponent entities by IDs.
func (shuo *SystemHazardUpdateOne) RemoveComponentIDs(ids ...uuid.UUID) *SystemHazardUpdateOne {
	shuo.mutation.RemoveComponentIDs(ids...)
	return shuo
}

// RemoveComponents removes "components" edges to SystemComponent entities.
func (shuo *SystemHazardUpdateOne) RemoveComponents(s ...*SystemComponent) *SystemHazardUpdateOne {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return shuo.RemoveComponentIDs(ids...)
}

// ClearConstraints clears all "constraints" edges to the SystemComponentConstraint entity.
func (shuo *SystemHazardUpdateOne) ClearConstraints() *SystemHazardUpdateOne {
	shuo.mutation.ClearConstraints()
	return shuo
}

// RemoveConstraintIDs removes the "constraints" edge to SystemComponentConstraint entities by IDs.
func (shuo *SystemHazardUpdateOne) RemoveConstraintIDs(ids ...uuid.UUID) *SystemHazardUpdateOne {
	shuo.mutation.RemoveConstraintIDs(ids...)
	return shuo
}

// RemoveConstraints removes "constraints" edges to SystemComponentConstraint entities.
func (shuo *SystemHazardUpdateOne) RemoveConstraints(s ...*SystemComponentConstraint) *SystemHazardUpdateOne {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return shuo.RemoveConstraintIDs(ids...)
}

// ClearRelationships clears all "relationships" edges to the SystemComponentRelationship entity.
func (shuo *SystemHazardUpdateOne) ClearRelationships() *SystemHazardUpdateOne {
	shuo.mutation.ClearRelationships()
	return shuo
}

// RemoveRelationshipIDs removes the "relationships" edge to SystemComponentRelationship entities by IDs.
func (shuo *SystemHazardUpdateOne) RemoveRelationshipIDs(ids ...uuid.UUID) *SystemHazardUpdateOne {
	shuo.mutation.RemoveRelationshipIDs(ids...)
	return shuo
}

// RemoveRelationships removes "relationships" edges to SystemComponentRelationship entities.
func (shuo *SystemHazardUpdateOne) RemoveRelationships(s ...*SystemComponentRelationship) *SystemHazardUpdateOne {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return shuo.RemoveRelationshipIDs(ids...)
}

// Where appends a list predicates to the SystemHazardUpdate builder.
func (shuo *SystemHazardUpdateOne) Where(ps ...predicate.SystemHazard) *SystemHazardUpdateOne {
	shuo.mutation.Where(ps...)
	return shuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (shuo *SystemHazardUpdateOne) Select(field string, fields ...string) *SystemHazardUpdateOne {
	shuo.fields = append([]string{field}, fields...)
	return shuo
}

// Save executes the query and returns the updated SystemHazard entity.
func (shuo *SystemHazardUpdateOne) Save(ctx context.Context) (*SystemHazard, error) {
	shuo.defaults()
	return withHooks(ctx, shuo.sqlSave, shuo.mutation, shuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (shuo *SystemHazardUpdateOne) SaveX(ctx context.Context) *SystemHazard {
	node, err := shuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (shuo *SystemHazardUpdateOne) Exec(ctx context.Context) error {
	_, err := shuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (shuo *SystemHazardUpdateOne) ExecX(ctx context.Context) {
	if err := shuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (shuo *SystemHazardUpdateOne) defaults() {
	if _, ok := shuo.mutation.UpdatedAt(); !ok {
		v := systemhazard.UpdateDefaultUpdatedAt()
		shuo.mutation.SetUpdatedAt(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (shuo *SystemHazardUpdateOne) check() error {
	if v, ok := shuo.mutation.Name(); ok {
		if err := systemhazard.NameValidator(v); err != nil {
			return &ValidationError{Name: "name", err: fmt.Errorf(`ent: validator failed for field "SystemHazard.name": %w`, err)}
		}
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (shuo *SystemHazardUpdateOne) Modify(modifiers ...func(u *sql.UpdateBuilder)) *SystemHazardUpdateOne {
	shuo.modifiers = append(shuo.modifiers, modifiers...)
	return shuo
}

func (shuo *SystemHazardUpdateOne) sqlSave(ctx context.Context) (_node *SystemHazard, err error) {
	if err := shuo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(systemhazard.Table, systemhazard.Columns, sqlgraph.NewFieldSpec(systemhazard.FieldID, field.TypeUUID))
	id, ok := shuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "SystemHazard.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := shuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, systemhazard.FieldID)
		for _, f := range fields {
			if !systemhazard.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != systemhazard.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := shuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := shuo.mutation.Name(); ok {
		_spec.SetField(systemhazard.FieldName, field.TypeString, value)
	}
	if value, ok := shuo.mutation.Description(); ok {
		_spec.SetField(systemhazard.FieldDescription, field.TypeString, value)
	}
	if value, ok := shuo.mutation.CreatedAt(); ok {
		_spec.SetField(systemhazard.FieldCreatedAt, field.TypeTime, value)
	}
	if value, ok := shuo.mutation.UpdatedAt(); ok {
		_spec.SetField(systemhazard.FieldUpdatedAt, field.TypeTime, value)
	}
	if shuo.mutation.ComponentsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   systemhazard.ComponentsTable,
			Columns: systemhazard.ComponentsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponent.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := shuo.mutation.RemovedComponentsIDs(); len(nodes) > 0 && !shuo.mutation.ComponentsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   systemhazard.ComponentsTable,
			Columns: systemhazard.ComponentsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponent.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := shuo.mutation.ComponentsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   systemhazard.ComponentsTable,
			Columns: systemhazard.ComponentsPrimaryKey,
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
	if shuo.mutation.ConstraintsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   systemhazard.ConstraintsTable,
			Columns: systemhazard.ConstraintsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponentconstraint.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := shuo.mutation.RemovedConstraintsIDs(); len(nodes) > 0 && !shuo.mutation.ConstraintsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   systemhazard.ConstraintsTable,
			Columns: systemhazard.ConstraintsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponentconstraint.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := shuo.mutation.ConstraintsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   systemhazard.ConstraintsTable,
			Columns: systemhazard.ConstraintsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponentconstraint.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if shuo.mutation.RelationshipsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   systemhazard.RelationshipsTable,
			Columns: systemhazard.RelationshipsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponentrelationship.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := shuo.mutation.RemovedRelationshipsIDs(); len(nodes) > 0 && !shuo.mutation.RelationshipsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   systemhazard.RelationshipsTable,
			Columns: systemhazard.RelationshipsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponentrelationship.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := shuo.mutation.RelationshipsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   systemhazard.RelationshipsTable,
			Columns: systemhazard.RelationshipsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(systemcomponentrelationship.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(shuo.modifiers...)
	_node = &SystemHazard{config: shuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, shuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{systemhazard.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	shuo.mutation.done = true
	return _node, nil
}
