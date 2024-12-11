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
	"github.com/twohundreds/rezible/ent/incident"
	"github.com/twohundreds/rezible/ent/incidentevent"
	"github.com/twohundreds/rezible/ent/predicate"
	"github.com/twohundreds/rezible/ent/service"
)

// IncidentEventUpdate is the builder for updating IncidentEvent entities.
type IncidentEventUpdate struct {
	config
	hooks     []Hook
	mutation  *IncidentEventMutation
	modifiers []func(*sql.UpdateBuilder)
}

// Where appends a list predicates to the IncidentEventUpdate builder.
func (ieu *IncidentEventUpdate) Where(ps ...predicate.IncidentEvent) *IncidentEventUpdate {
	ieu.mutation.Where(ps...)
	return ieu
}

// SetType sets the "type" field.
func (ieu *IncidentEventUpdate) SetType(i incidentevent.Type) *IncidentEventUpdate {
	ieu.mutation.SetType(i)
	return ieu
}

// SetNillableType sets the "type" field if the given value is not nil.
func (ieu *IncidentEventUpdate) SetNillableType(i *incidentevent.Type) *IncidentEventUpdate {
	if i != nil {
		ieu.SetType(*i)
	}
	return ieu
}

// SetTime sets the "time" field.
func (ieu *IncidentEventUpdate) SetTime(t time.Time) *IncidentEventUpdate {
	ieu.mutation.SetTime(t)
	return ieu
}

// SetNillableTime sets the "time" field if the given value is not nil.
func (ieu *IncidentEventUpdate) SetNillableTime(t *time.Time) *IncidentEventUpdate {
	if t != nil {
		ieu.SetTime(*t)
	}
	return ieu
}

// SetIncidentID sets the "incident_id" field.
func (ieu *IncidentEventUpdate) SetIncidentID(u uuid.UUID) *IncidentEventUpdate {
	ieu.mutation.SetIncidentID(u)
	return ieu
}

// SetNillableIncidentID sets the "incident_id" field if the given value is not nil.
func (ieu *IncidentEventUpdate) SetNillableIncidentID(u *uuid.UUID) *IncidentEventUpdate {
	if u != nil {
		ieu.SetIncidentID(*u)
	}
	return ieu
}

// SetIncident sets the "incident" edge to the Incident entity.
func (ieu *IncidentEventUpdate) SetIncident(i *Incident) *IncidentEventUpdate {
	return ieu.SetIncidentID(i.ID)
}

// AddServiceIDs adds the "services" edge to the Service entity by IDs.
func (ieu *IncidentEventUpdate) AddServiceIDs(ids ...uuid.UUID) *IncidentEventUpdate {
	ieu.mutation.AddServiceIDs(ids...)
	return ieu
}

// AddServices adds the "services" edges to the Service entity.
func (ieu *IncidentEventUpdate) AddServices(s ...*Service) *IncidentEventUpdate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return ieu.AddServiceIDs(ids...)
}

// Mutation returns the IncidentEventMutation object of the builder.
func (ieu *IncidentEventUpdate) Mutation() *IncidentEventMutation {
	return ieu.mutation
}

// ClearIncident clears the "incident" edge to the Incident entity.
func (ieu *IncidentEventUpdate) ClearIncident() *IncidentEventUpdate {
	ieu.mutation.ClearIncident()
	return ieu
}

// ClearServices clears all "services" edges to the Service entity.
func (ieu *IncidentEventUpdate) ClearServices() *IncidentEventUpdate {
	ieu.mutation.ClearServices()
	return ieu
}

// RemoveServiceIDs removes the "services" edge to Service entities by IDs.
func (ieu *IncidentEventUpdate) RemoveServiceIDs(ids ...uuid.UUID) *IncidentEventUpdate {
	ieu.mutation.RemoveServiceIDs(ids...)
	return ieu
}

// RemoveServices removes "services" edges to Service entities.
func (ieu *IncidentEventUpdate) RemoveServices(s ...*Service) *IncidentEventUpdate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return ieu.RemoveServiceIDs(ids...)
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (ieu *IncidentEventUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, ieu.sqlSave, ieu.mutation, ieu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (ieu *IncidentEventUpdate) SaveX(ctx context.Context) int {
	affected, err := ieu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (ieu *IncidentEventUpdate) Exec(ctx context.Context) error {
	_, err := ieu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ieu *IncidentEventUpdate) ExecX(ctx context.Context) {
	if err := ieu.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (ieu *IncidentEventUpdate) check() error {
	if v, ok := ieu.mutation.GetType(); ok {
		if err := incidentevent.TypeValidator(v); err != nil {
			return &ValidationError{Name: "type", err: fmt.Errorf(`ent: validator failed for field "IncidentEvent.type": %w`, err)}
		}
	}
	if ieu.mutation.IncidentCleared() && len(ieu.mutation.IncidentIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "IncidentEvent.incident"`)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (ieu *IncidentEventUpdate) Modify(modifiers ...func(u *sql.UpdateBuilder)) *IncidentEventUpdate {
	ieu.modifiers = append(ieu.modifiers, modifiers...)
	return ieu
}

func (ieu *IncidentEventUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := ieu.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(incidentevent.Table, incidentevent.Columns, sqlgraph.NewFieldSpec(incidentevent.FieldID, field.TypeUUID))
	if ps := ieu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := ieu.mutation.GetType(); ok {
		_spec.SetField(incidentevent.FieldType, field.TypeEnum, value)
	}
	if value, ok := ieu.mutation.Time(); ok {
		_spec.SetField(incidentevent.FieldTime, field.TypeTime, value)
	}
	if ieu.mutation.IncidentCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   incidentevent.IncidentTable,
			Columns: []string{incidentevent.IncidentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incident.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := ieu.mutation.IncidentIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   incidentevent.IncidentTable,
			Columns: []string{incidentevent.IncidentColumn},
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
	if ieu.mutation.ServicesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   incidentevent.ServicesTable,
			Columns: []string{incidentevent.ServicesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(service.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := ieu.mutation.RemovedServicesIDs(); len(nodes) > 0 && !ieu.mutation.ServicesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   incidentevent.ServicesTable,
			Columns: []string{incidentevent.ServicesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(service.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := ieu.mutation.ServicesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   incidentevent.ServicesTable,
			Columns: []string{incidentevent.ServicesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(service.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(ieu.modifiers...)
	if n, err = sqlgraph.UpdateNodes(ctx, ieu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{incidentevent.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	ieu.mutation.done = true
	return n, nil
}

// IncidentEventUpdateOne is the builder for updating a single IncidentEvent entity.
type IncidentEventUpdateOne struct {
	config
	fields    []string
	hooks     []Hook
	mutation  *IncidentEventMutation
	modifiers []func(*sql.UpdateBuilder)
}

// SetType sets the "type" field.
func (ieuo *IncidentEventUpdateOne) SetType(i incidentevent.Type) *IncidentEventUpdateOne {
	ieuo.mutation.SetType(i)
	return ieuo
}

// SetNillableType sets the "type" field if the given value is not nil.
func (ieuo *IncidentEventUpdateOne) SetNillableType(i *incidentevent.Type) *IncidentEventUpdateOne {
	if i != nil {
		ieuo.SetType(*i)
	}
	return ieuo
}

// SetTime sets the "time" field.
func (ieuo *IncidentEventUpdateOne) SetTime(t time.Time) *IncidentEventUpdateOne {
	ieuo.mutation.SetTime(t)
	return ieuo
}

// SetNillableTime sets the "time" field if the given value is not nil.
func (ieuo *IncidentEventUpdateOne) SetNillableTime(t *time.Time) *IncidentEventUpdateOne {
	if t != nil {
		ieuo.SetTime(*t)
	}
	return ieuo
}

// SetIncidentID sets the "incident_id" field.
func (ieuo *IncidentEventUpdateOne) SetIncidentID(u uuid.UUID) *IncidentEventUpdateOne {
	ieuo.mutation.SetIncidentID(u)
	return ieuo
}

// SetNillableIncidentID sets the "incident_id" field if the given value is not nil.
func (ieuo *IncidentEventUpdateOne) SetNillableIncidentID(u *uuid.UUID) *IncidentEventUpdateOne {
	if u != nil {
		ieuo.SetIncidentID(*u)
	}
	return ieuo
}

// SetIncident sets the "incident" edge to the Incident entity.
func (ieuo *IncidentEventUpdateOne) SetIncident(i *Incident) *IncidentEventUpdateOne {
	return ieuo.SetIncidentID(i.ID)
}

// AddServiceIDs adds the "services" edge to the Service entity by IDs.
func (ieuo *IncidentEventUpdateOne) AddServiceIDs(ids ...uuid.UUID) *IncidentEventUpdateOne {
	ieuo.mutation.AddServiceIDs(ids...)
	return ieuo
}

// AddServices adds the "services" edges to the Service entity.
func (ieuo *IncidentEventUpdateOne) AddServices(s ...*Service) *IncidentEventUpdateOne {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return ieuo.AddServiceIDs(ids...)
}

// Mutation returns the IncidentEventMutation object of the builder.
func (ieuo *IncidentEventUpdateOne) Mutation() *IncidentEventMutation {
	return ieuo.mutation
}

// ClearIncident clears the "incident" edge to the Incident entity.
func (ieuo *IncidentEventUpdateOne) ClearIncident() *IncidentEventUpdateOne {
	ieuo.mutation.ClearIncident()
	return ieuo
}

// ClearServices clears all "services" edges to the Service entity.
func (ieuo *IncidentEventUpdateOne) ClearServices() *IncidentEventUpdateOne {
	ieuo.mutation.ClearServices()
	return ieuo
}

// RemoveServiceIDs removes the "services" edge to Service entities by IDs.
func (ieuo *IncidentEventUpdateOne) RemoveServiceIDs(ids ...uuid.UUID) *IncidentEventUpdateOne {
	ieuo.mutation.RemoveServiceIDs(ids...)
	return ieuo
}

// RemoveServices removes "services" edges to Service entities.
func (ieuo *IncidentEventUpdateOne) RemoveServices(s ...*Service) *IncidentEventUpdateOne {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return ieuo.RemoveServiceIDs(ids...)
}

// Where appends a list predicates to the IncidentEventUpdate builder.
func (ieuo *IncidentEventUpdateOne) Where(ps ...predicate.IncidentEvent) *IncidentEventUpdateOne {
	ieuo.mutation.Where(ps...)
	return ieuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (ieuo *IncidentEventUpdateOne) Select(field string, fields ...string) *IncidentEventUpdateOne {
	ieuo.fields = append([]string{field}, fields...)
	return ieuo
}

// Save executes the query and returns the updated IncidentEvent entity.
func (ieuo *IncidentEventUpdateOne) Save(ctx context.Context) (*IncidentEvent, error) {
	return withHooks(ctx, ieuo.sqlSave, ieuo.mutation, ieuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (ieuo *IncidentEventUpdateOne) SaveX(ctx context.Context) *IncidentEvent {
	node, err := ieuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (ieuo *IncidentEventUpdateOne) Exec(ctx context.Context) error {
	_, err := ieuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ieuo *IncidentEventUpdateOne) ExecX(ctx context.Context) {
	if err := ieuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (ieuo *IncidentEventUpdateOne) check() error {
	if v, ok := ieuo.mutation.GetType(); ok {
		if err := incidentevent.TypeValidator(v); err != nil {
			return &ValidationError{Name: "type", err: fmt.Errorf(`ent: validator failed for field "IncidentEvent.type": %w`, err)}
		}
	}
	if ieuo.mutation.IncidentCleared() && len(ieuo.mutation.IncidentIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "IncidentEvent.incident"`)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (ieuo *IncidentEventUpdateOne) Modify(modifiers ...func(u *sql.UpdateBuilder)) *IncidentEventUpdateOne {
	ieuo.modifiers = append(ieuo.modifiers, modifiers...)
	return ieuo
}

func (ieuo *IncidentEventUpdateOne) sqlSave(ctx context.Context) (_node *IncidentEvent, err error) {
	if err := ieuo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(incidentevent.Table, incidentevent.Columns, sqlgraph.NewFieldSpec(incidentevent.FieldID, field.TypeUUID))
	id, ok := ieuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "IncidentEvent.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := ieuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, incidentevent.FieldID)
		for _, f := range fields {
			if !incidentevent.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != incidentevent.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := ieuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := ieuo.mutation.GetType(); ok {
		_spec.SetField(incidentevent.FieldType, field.TypeEnum, value)
	}
	if value, ok := ieuo.mutation.Time(); ok {
		_spec.SetField(incidentevent.FieldTime, field.TypeTime, value)
	}
	if ieuo.mutation.IncidentCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   incidentevent.IncidentTable,
			Columns: []string{incidentevent.IncidentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incident.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := ieuo.mutation.IncidentIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   incidentevent.IncidentTable,
			Columns: []string{incidentevent.IncidentColumn},
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
	if ieuo.mutation.ServicesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   incidentevent.ServicesTable,
			Columns: []string{incidentevent.ServicesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(service.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := ieuo.mutation.RemovedServicesIDs(); len(nodes) > 0 && !ieuo.mutation.ServicesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   incidentevent.ServicesTable,
			Columns: []string{incidentevent.ServicesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(service.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := ieuo.mutation.ServicesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   incidentevent.ServicesTable,
			Columns: []string{incidentevent.ServicesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(service.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(ieuo.modifiers...)
	_node = &IncidentEvent{config: ieuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, ieuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{incidentevent.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	ieuo.mutation.done = true
	return _node, nil
}