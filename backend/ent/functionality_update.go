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
	"github.com/rezible/rezible/ent/functionality"
	"github.com/rezible/rezible/ent/incidentresourceimpact"
	"github.com/rezible/rezible/ent/predicate"
)

// FunctionalityUpdate is the builder for updating Functionality entities.
type FunctionalityUpdate struct {
	config
	hooks     []Hook
	mutation  *FunctionalityMutation
	modifiers []func(*sql.UpdateBuilder)
}

// Where appends a list predicates to the FunctionalityUpdate builder.
func (fu *FunctionalityUpdate) Where(ps ...predicate.Functionality) *FunctionalityUpdate {
	fu.mutation.Where(ps...)
	return fu
}

// SetName sets the "name" field.
func (fu *FunctionalityUpdate) SetName(s string) *FunctionalityUpdate {
	fu.mutation.SetName(s)
	return fu
}

// SetNillableName sets the "name" field if the given value is not nil.
func (fu *FunctionalityUpdate) SetNillableName(s *string) *FunctionalityUpdate {
	if s != nil {
		fu.SetName(*s)
	}
	return fu
}

// AddIncidentIDs adds the "incidents" edge to the IncidentResourceImpact entity by IDs.
func (fu *FunctionalityUpdate) AddIncidentIDs(ids ...uuid.UUID) *FunctionalityUpdate {
	fu.mutation.AddIncidentIDs(ids...)
	return fu
}

// AddIncidents adds the "incidents" edges to the IncidentResourceImpact entity.
func (fu *FunctionalityUpdate) AddIncidents(i ...*IncidentResourceImpact) *FunctionalityUpdate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return fu.AddIncidentIDs(ids...)
}

// Mutation returns the FunctionalityMutation object of the builder.
func (fu *FunctionalityUpdate) Mutation() *FunctionalityMutation {
	return fu.mutation
}

// ClearIncidents clears all "incidents" edges to the IncidentResourceImpact entity.
func (fu *FunctionalityUpdate) ClearIncidents() *FunctionalityUpdate {
	fu.mutation.ClearIncidents()
	return fu
}

// RemoveIncidentIDs removes the "incidents" edge to IncidentResourceImpact entities by IDs.
func (fu *FunctionalityUpdate) RemoveIncidentIDs(ids ...uuid.UUID) *FunctionalityUpdate {
	fu.mutation.RemoveIncidentIDs(ids...)
	return fu
}

// RemoveIncidents removes "incidents" edges to IncidentResourceImpact entities.
func (fu *FunctionalityUpdate) RemoveIncidents(i ...*IncidentResourceImpact) *FunctionalityUpdate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return fu.RemoveIncidentIDs(ids...)
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (fu *FunctionalityUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, fu.sqlSave, fu.mutation, fu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (fu *FunctionalityUpdate) SaveX(ctx context.Context) int {
	affected, err := fu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (fu *FunctionalityUpdate) Exec(ctx context.Context) error {
	_, err := fu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (fu *FunctionalityUpdate) ExecX(ctx context.Context) {
	if err := fu.Exec(ctx); err != nil {
		panic(err)
	}
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (fu *FunctionalityUpdate) Modify(modifiers ...func(u *sql.UpdateBuilder)) *FunctionalityUpdate {
	fu.modifiers = append(fu.modifiers, modifiers...)
	return fu
}

func (fu *FunctionalityUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := sqlgraph.NewUpdateSpec(functionality.Table, functionality.Columns, sqlgraph.NewFieldSpec(functionality.FieldID, field.TypeUUID))
	if ps := fu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := fu.mutation.Name(); ok {
		_spec.SetField(functionality.FieldName, field.TypeString, value)
	}
	if fu.mutation.IncidentsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   functionality.IncidentsTable,
			Columns: []string{functionality.IncidentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentresourceimpact.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := fu.mutation.RemovedIncidentsIDs(); len(nodes) > 0 && !fu.mutation.IncidentsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   functionality.IncidentsTable,
			Columns: []string{functionality.IncidentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentresourceimpact.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := fu.mutation.IncidentsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   functionality.IncidentsTable,
			Columns: []string{functionality.IncidentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentresourceimpact.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(fu.modifiers...)
	if n, err = sqlgraph.UpdateNodes(ctx, fu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{functionality.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	fu.mutation.done = true
	return n, nil
}

// FunctionalityUpdateOne is the builder for updating a single Functionality entity.
type FunctionalityUpdateOne struct {
	config
	fields    []string
	hooks     []Hook
	mutation  *FunctionalityMutation
	modifiers []func(*sql.UpdateBuilder)
}

// SetName sets the "name" field.
func (fuo *FunctionalityUpdateOne) SetName(s string) *FunctionalityUpdateOne {
	fuo.mutation.SetName(s)
	return fuo
}

// SetNillableName sets the "name" field if the given value is not nil.
func (fuo *FunctionalityUpdateOne) SetNillableName(s *string) *FunctionalityUpdateOne {
	if s != nil {
		fuo.SetName(*s)
	}
	return fuo
}

// AddIncidentIDs adds the "incidents" edge to the IncidentResourceImpact entity by IDs.
func (fuo *FunctionalityUpdateOne) AddIncidentIDs(ids ...uuid.UUID) *FunctionalityUpdateOne {
	fuo.mutation.AddIncidentIDs(ids...)
	return fuo
}

// AddIncidents adds the "incidents" edges to the IncidentResourceImpact entity.
func (fuo *FunctionalityUpdateOne) AddIncidents(i ...*IncidentResourceImpact) *FunctionalityUpdateOne {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return fuo.AddIncidentIDs(ids...)
}

// Mutation returns the FunctionalityMutation object of the builder.
func (fuo *FunctionalityUpdateOne) Mutation() *FunctionalityMutation {
	return fuo.mutation
}

// ClearIncidents clears all "incidents" edges to the IncidentResourceImpact entity.
func (fuo *FunctionalityUpdateOne) ClearIncidents() *FunctionalityUpdateOne {
	fuo.mutation.ClearIncidents()
	return fuo
}

// RemoveIncidentIDs removes the "incidents" edge to IncidentResourceImpact entities by IDs.
func (fuo *FunctionalityUpdateOne) RemoveIncidentIDs(ids ...uuid.UUID) *FunctionalityUpdateOne {
	fuo.mutation.RemoveIncidentIDs(ids...)
	return fuo
}

// RemoveIncidents removes "incidents" edges to IncidentResourceImpact entities.
func (fuo *FunctionalityUpdateOne) RemoveIncidents(i ...*IncidentResourceImpact) *FunctionalityUpdateOne {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return fuo.RemoveIncidentIDs(ids...)
}

// Where appends a list predicates to the FunctionalityUpdate builder.
func (fuo *FunctionalityUpdateOne) Where(ps ...predicate.Functionality) *FunctionalityUpdateOne {
	fuo.mutation.Where(ps...)
	return fuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (fuo *FunctionalityUpdateOne) Select(field string, fields ...string) *FunctionalityUpdateOne {
	fuo.fields = append([]string{field}, fields...)
	return fuo
}

// Save executes the query and returns the updated Functionality entity.
func (fuo *FunctionalityUpdateOne) Save(ctx context.Context) (*Functionality, error) {
	return withHooks(ctx, fuo.sqlSave, fuo.mutation, fuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (fuo *FunctionalityUpdateOne) SaveX(ctx context.Context) *Functionality {
	node, err := fuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (fuo *FunctionalityUpdateOne) Exec(ctx context.Context) error {
	_, err := fuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (fuo *FunctionalityUpdateOne) ExecX(ctx context.Context) {
	if err := fuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (fuo *FunctionalityUpdateOne) Modify(modifiers ...func(u *sql.UpdateBuilder)) *FunctionalityUpdateOne {
	fuo.modifiers = append(fuo.modifiers, modifiers...)
	return fuo
}

func (fuo *FunctionalityUpdateOne) sqlSave(ctx context.Context) (_node *Functionality, err error) {
	_spec := sqlgraph.NewUpdateSpec(functionality.Table, functionality.Columns, sqlgraph.NewFieldSpec(functionality.FieldID, field.TypeUUID))
	id, ok := fuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "Functionality.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := fuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, functionality.FieldID)
		for _, f := range fields {
			if !functionality.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != functionality.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := fuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := fuo.mutation.Name(); ok {
		_spec.SetField(functionality.FieldName, field.TypeString, value)
	}
	if fuo.mutation.IncidentsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   functionality.IncidentsTable,
			Columns: []string{functionality.IncidentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentresourceimpact.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := fuo.mutation.RemovedIncidentsIDs(); len(nodes) > 0 && !fuo.mutation.IncidentsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   functionality.IncidentsTable,
			Columns: []string{functionality.IncidentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentresourceimpact.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := fuo.mutation.IncidentsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   functionality.IncidentsTable,
			Columns: []string{functionality.IncidentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentresourceimpact.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(fuo.modifiers...)
	_node = &Functionality{config: fuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, fuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{functionality.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	fuo.mutation.done = true
	return _node, nil
}
