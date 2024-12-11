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
	"github.com/twohundreds/rezible/ent/environment"
	"github.com/twohundreds/rezible/ent/incident"
	"github.com/twohundreds/rezible/ent/predicate"
)

// EnvironmentUpdate is the builder for updating Environment entities.
type EnvironmentUpdate struct {
	config
	hooks     []Hook
	mutation  *EnvironmentMutation
	modifiers []func(*sql.UpdateBuilder)
}

// Where appends a list predicates to the EnvironmentUpdate builder.
func (eu *EnvironmentUpdate) Where(ps ...predicate.Environment) *EnvironmentUpdate {
	eu.mutation.Where(ps...)
	return eu
}

// SetArchiveTime sets the "archive_time" field.
func (eu *EnvironmentUpdate) SetArchiveTime(t time.Time) *EnvironmentUpdate {
	eu.mutation.SetArchiveTime(t)
	return eu
}

// SetNillableArchiveTime sets the "archive_time" field if the given value is not nil.
func (eu *EnvironmentUpdate) SetNillableArchiveTime(t *time.Time) *EnvironmentUpdate {
	if t != nil {
		eu.SetArchiveTime(*t)
	}
	return eu
}

// ClearArchiveTime clears the value of the "archive_time" field.
func (eu *EnvironmentUpdate) ClearArchiveTime() *EnvironmentUpdate {
	eu.mutation.ClearArchiveTime()
	return eu
}

// SetName sets the "name" field.
func (eu *EnvironmentUpdate) SetName(s string) *EnvironmentUpdate {
	eu.mutation.SetName(s)
	return eu
}

// SetNillableName sets the "name" field if the given value is not nil.
func (eu *EnvironmentUpdate) SetNillableName(s *string) *EnvironmentUpdate {
	if s != nil {
		eu.SetName(*s)
	}
	return eu
}

// AddIncidentIDs adds the "incidents" edge to the Incident entity by IDs.
func (eu *EnvironmentUpdate) AddIncidentIDs(ids ...uuid.UUID) *EnvironmentUpdate {
	eu.mutation.AddIncidentIDs(ids...)
	return eu
}

// AddIncidents adds the "incidents" edges to the Incident entity.
func (eu *EnvironmentUpdate) AddIncidents(i ...*Incident) *EnvironmentUpdate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return eu.AddIncidentIDs(ids...)
}

// Mutation returns the EnvironmentMutation object of the builder.
func (eu *EnvironmentUpdate) Mutation() *EnvironmentMutation {
	return eu.mutation
}

// ClearIncidents clears all "incidents" edges to the Incident entity.
func (eu *EnvironmentUpdate) ClearIncidents() *EnvironmentUpdate {
	eu.mutation.ClearIncidents()
	return eu
}

// RemoveIncidentIDs removes the "incidents" edge to Incident entities by IDs.
func (eu *EnvironmentUpdate) RemoveIncidentIDs(ids ...uuid.UUID) *EnvironmentUpdate {
	eu.mutation.RemoveIncidentIDs(ids...)
	return eu
}

// RemoveIncidents removes "incidents" edges to Incident entities.
func (eu *EnvironmentUpdate) RemoveIncidents(i ...*Incident) *EnvironmentUpdate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return eu.RemoveIncidentIDs(ids...)
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (eu *EnvironmentUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, eu.sqlSave, eu.mutation, eu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (eu *EnvironmentUpdate) SaveX(ctx context.Context) int {
	affected, err := eu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (eu *EnvironmentUpdate) Exec(ctx context.Context) error {
	_, err := eu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (eu *EnvironmentUpdate) ExecX(ctx context.Context) {
	if err := eu.Exec(ctx); err != nil {
		panic(err)
	}
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (eu *EnvironmentUpdate) Modify(modifiers ...func(u *sql.UpdateBuilder)) *EnvironmentUpdate {
	eu.modifiers = append(eu.modifiers, modifiers...)
	return eu
}

func (eu *EnvironmentUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := sqlgraph.NewUpdateSpec(environment.Table, environment.Columns, sqlgraph.NewFieldSpec(environment.FieldID, field.TypeUUID))
	if ps := eu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := eu.mutation.ArchiveTime(); ok {
		_spec.SetField(environment.FieldArchiveTime, field.TypeTime, value)
	}
	if eu.mutation.ArchiveTimeCleared() {
		_spec.ClearField(environment.FieldArchiveTime, field.TypeTime)
	}
	if value, ok := eu.mutation.Name(); ok {
		_spec.SetField(environment.FieldName, field.TypeString, value)
	}
	if eu.mutation.IncidentsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   environment.IncidentsTable,
			Columns: environment.IncidentsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incident.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := eu.mutation.RemovedIncidentsIDs(); len(nodes) > 0 && !eu.mutation.IncidentsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   environment.IncidentsTable,
			Columns: environment.IncidentsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incident.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := eu.mutation.IncidentsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   environment.IncidentsTable,
			Columns: environment.IncidentsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incident.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(eu.modifiers...)
	if n, err = sqlgraph.UpdateNodes(ctx, eu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{environment.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	eu.mutation.done = true
	return n, nil
}

// EnvironmentUpdateOne is the builder for updating a single Environment entity.
type EnvironmentUpdateOne struct {
	config
	fields    []string
	hooks     []Hook
	mutation  *EnvironmentMutation
	modifiers []func(*sql.UpdateBuilder)
}

// SetArchiveTime sets the "archive_time" field.
func (euo *EnvironmentUpdateOne) SetArchiveTime(t time.Time) *EnvironmentUpdateOne {
	euo.mutation.SetArchiveTime(t)
	return euo
}

// SetNillableArchiveTime sets the "archive_time" field if the given value is not nil.
func (euo *EnvironmentUpdateOne) SetNillableArchiveTime(t *time.Time) *EnvironmentUpdateOne {
	if t != nil {
		euo.SetArchiveTime(*t)
	}
	return euo
}

// ClearArchiveTime clears the value of the "archive_time" field.
func (euo *EnvironmentUpdateOne) ClearArchiveTime() *EnvironmentUpdateOne {
	euo.mutation.ClearArchiveTime()
	return euo
}

// SetName sets the "name" field.
func (euo *EnvironmentUpdateOne) SetName(s string) *EnvironmentUpdateOne {
	euo.mutation.SetName(s)
	return euo
}

// SetNillableName sets the "name" field if the given value is not nil.
func (euo *EnvironmentUpdateOne) SetNillableName(s *string) *EnvironmentUpdateOne {
	if s != nil {
		euo.SetName(*s)
	}
	return euo
}

// AddIncidentIDs adds the "incidents" edge to the Incident entity by IDs.
func (euo *EnvironmentUpdateOne) AddIncidentIDs(ids ...uuid.UUID) *EnvironmentUpdateOne {
	euo.mutation.AddIncidentIDs(ids...)
	return euo
}

// AddIncidents adds the "incidents" edges to the Incident entity.
func (euo *EnvironmentUpdateOne) AddIncidents(i ...*Incident) *EnvironmentUpdateOne {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return euo.AddIncidentIDs(ids...)
}

// Mutation returns the EnvironmentMutation object of the builder.
func (euo *EnvironmentUpdateOne) Mutation() *EnvironmentMutation {
	return euo.mutation
}

// ClearIncidents clears all "incidents" edges to the Incident entity.
func (euo *EnvironmentUpdateOne) ClearIncidents() *EnvironmentUpdateOne {
	euo.mutation.ClearIncidents()
	return euo
}

// RemoveIncidentIDs removes the "incidents" edge to Incident entities by IDs.
func (euo *EnvironmentUpdateOne) RemoveIncidentIDs(ids ...uuid.UUID) *EnvironmentUpdateOne {
	euo.mutation.RemoveIncidentIDs(ids...)
	return euo
}

// RemoveIncidents removes "incidents" edges to Incident entities.
func (euo *EnvironmentUpdateOne) RemoveIncidents(i ...*Incident) *EnvironmentUpdateOne {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return euo.RemoveIncidentIDs(ids...)
}

// Where appends a list predicates to the EnvironmentUpdate builder.
func (euo *EnvironmentUpdateOne) Where(ps ...predicate.Environment) *EnvironmentUpdateOne {
	euo.mutation.Where(ps...)
	return euo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (euo *EnvironmentUpdateOne) Select(field string, fields ...string) *EnvironmentUpdateOne {
	euo.fields = append([]string{field}, fields...)
	return euo
}

// Save executes the query and returns the updated Environment entity.
func (euo *EnvironmentUpdateOne) Save(ctx context.Context) (*Environment, error) {
	return withHooks(ctx, euo.sqlSave, euo.mutation, euo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (euo *EnvironmentUpdateOne) SaveX(ctx context.Context) *Environment {
	node, err := euo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (euo *EnvironmentUpdateOne) Exec(ctx context.Context) error {
	_, err := euo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (euo *EnvironmentUpdateOne) ExecX(ctx context.Context) {
	if err := euo.Exec(ctx); err != nil {
		panic(err)
	}
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (euo *EnvironmentUpdateOne) Modify(modifiers ...func(u *sql.UpdateBuilder)) *EnvironmentUpdateOne {
	euo.modifiers = append(euo.modifiers, modifiers...)
	return euo
}

func (euo *EnvironmentUpdateOne) sqlSave(ctx context.Context) (_node *Environment, err error) {
	_spec := sqlgraph.NewUpdateSpec(environment.Table, environment.Columns, sqlgraph.NewFieldSpec(environment.FieldID, field.TypeUUID))
	id, ok := euo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "Environment.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := euo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, environment.FieldID)
		for _, f := range fields {
			if !environment.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != environment.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := euo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := euo.mutation.ArchiveTime(); ok {
		_spec.SetField(environment.FieldArchiveTime, field.TypeTime, value)
	}
	if euo.mutation.ArchiveTimeCleared() {
		_spec.ClearField(environment.FieldArchiveTime, field.TypeTime)
	}
	if value, ok := euo.mutation.Name(); ok {
		_spec.SetField(environment.FieldName, field.TypeString, value)
	}
	if euo.mutation.IncidentsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   environment.IncidentsTable,
			Columns: environment.IncidentsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incident.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := euo.mutation.RemovedIncidentsIDs(); len(nodes) > 0 && !euo.mutation.IncidentsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   environment.IncidentsTable,
			Columns: environment.IncidentsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incident.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := euo.mutation.IncidentsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   environment.IncidentsTable,
			Columns: environment.IncidentsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incident.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(euo.modifiers...)
	_node = &Environment{config: euo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, euo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{environment.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	euo.mutation.done = true
	return _node, nil
}
