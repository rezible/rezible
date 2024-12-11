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
	"github.com/twohundreds/rezible/ent/functionality"
	"github.com/twohundreds/rezible/ent/incident"
	"github.com/twohundreds/rezible/ent/incidentlink"
	"github.com/twohundreds/rezible/ent/incidentresourceimpact"
	"github.com/twohundreds/rezible/ent/predicate"
	"github.com/twohundreds/rezible/ent/service"
)

// IncidentResourceImpactUpdate is the builder for updating IncidentResourceImpact entities.
type IncidentResourceImpactUpdate struct {
	config
	hooks     []Hook
	mutation  *IncidentResourceImpactMutation
	modifiers []func(*sql.UpdateBuilder)
}

// Where appends a list predicates to the IncidentResourceImpactUpdate builder.
func (iriu *IncidentResourceImpactUpdate) Where(ps ...predicate.IncidentResourceImpact) *IncidentResourceImpactUpdate {
	iriu.mutation.Where(ps...)
	return iriu
}

// SetIncidentID sets the "incident_id" field.
func (iriu *IncidentResourceImpactUpdate) SetIncidentID(u uuid.UUID) *IncidentResourceImpactUpdate {
	iriu.mutation.SetIncidentID(u)
	return iriu
}

// SetNillableIncidentID sets the "incident_id" field if the given value is not nil.
func (iriu *IncidentResourceImpactUpdate) SetNillableIncidentID(u *uuid.UUID) *IncidentResourceImpactUpdate {
	if u != nil {
		iriu.SetIncidentID(*u)
	}
	return iriu
}

// SetServiceID sets the "service_id" field.
func (iriu *IncidentResourceImpactUpdate) SetServiceID(u uuid.UUID) *IncidentResourceImpactUpdate {
	iriu.mutation.SetServiceID(u)
	return iriu
}

// SetNillableServiceID sets the "service_id" field if the given value is not nil.
func (iriu *IncidentResourceImpactUpdate) SetNillableServiceID(u *uuid.UUID) *IncidentResourceImpactUpdate {
	if u != nil {
		iriu.SetServiceID(*u)
	}
	return iriu
}

// ClearServiceID clears the value of the "service_id" field.
func (iriu *IncidentResourceImpactUpdate) ClearServiceID() *IncidentResourceImpactUpdate {
	iriu.mutation.ClearServiceID()
	return iriu
}

// SetFunctionalityID sets the "functionality_id" field.
func (iriu *IncidentResourceImpactUpdate) SetFunctionalityID(u uuid.UUID) *IncidentResourceImpactUpdate {
	iriu.mutation.SetFunctionalityID(u)
	return iriu
}

// SetNillableFunctionalityID sets the "functionality_id" field if the given value is not nil.
func (iriu *IncidentResourceImpactUpdate) SetNillableFunctionalityID(u *uuid.UUID) *IncidentResourceImpactUpdate {
	if u != nil {
		iriu.SetFunctionalityID(*u)
	}
	return iriu
}

// ClearFunctionalityID clears the value of the "functionality_id" field.
func (iriu *IncidentResourceImpactUpdate) ClearFunctionalityID() *IncidentResourceImpactUpdate {
	iriu.mutation.ClearFunctionalityID()
	return iriu
}

// SetIncident sets the "incident" edge to the Incident entity.
func (iriu *IncidentResourceImpactUpdate) SetIncident(i *Incident) *IncidentResourceImpactUpdate {
	return iriu.SetIncidentID(i.ID)
}

// SetService sets the "service" edge to the Service entity.
func (iriu *IncidentResourceImpactUpdate) SetService(s *Service) *IncidentResourceImpactUpdate {
	return iriu.SetServiceID(s.ID)
}

// SetFunctionality sets the "functionality" edge to the Functionality entity.
func (iriu *IncidentResourceImpactUpdate) SetFunctionality(f *Functionality) *IncidentResourceImpactUpdate {
	return iriu.SetFunctionalityID(f.ID)
}

// AddResultingIncidentIDs adds the "resulting_incidents" edge to the IncidentLink entity by IDs.
func (iriu *IncidentResourceImpactUpdate) AddResultingIncidentIDs(ids ...int) *IncidentResourceImpactUpdate {
	iriu.mutation.AddResultingIncidentIDs(ids...)
	return iriu
}

// AddResultingIncidents adds the "resulting_incidents" edges to the IncidentLink entity.
func (iriu *IncidentResourceImpactUpdate) AddResultingIncidents(i ...*IncidentLink) *IncidentResourceImpactUpdate {
	ids := make([]int, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return iriu.AddResultingIncidentIDs(ids...)
}

// Mutation returns the IncidentResourceImpactMutation object of the builder.
func (iriu *IncidentResourceImpactUpdate) Mutation() *IncidentResourceImpactMutation {
	return iriu.mutation
}

// ClearIncident clears the "incident" edge to the Incident entity.
func (iriu *IncidentResourceImpactUpdate) ClearIncident() *IncidentResourceImpactUpdate {
	iriu.mutation.ClearIncident()
	return iriu
}

// ClearService clears the "service" edge to the Service entity.
func (iriu *IncidentResourceImpactUpdate) ClearService() *IncidentResourceImpactUpdate {
	iriu.mutation.ClearService()
	return iriu
}

// ClearFunctionality clears the "functionality" edge to the Functionality entity.
func (iriu *IncidentResourceImpactUpdate) ClearFunctionality() *IncidentResourceImpactUpdate {
	iriu.mutation.ClearFunctionality()
	return iriu
}

// ClearResultingIncidents clears all "resulting_incidents" edges to the IncidentLink entity.
func (iriu *IncidentResourceImpactUpdate) ClearResultingIncidents() *IncidentResourceImpactUpdate {
	iriu.mutation.ClearResultingIncidents()
	return iriu
}

// RemoveResultingIncidentIDs removes the "resulting_incidents" edge to IncidentLink entities by IDs.
func (iriu *IncidentResourceImpactUpdate) RemoveResultingIncidentIDs(ids ...int) *IncidentResourceImpactUpdate {
	iriu.mutation.RemoveResultingIncidentIDs(ids...)
	return iriu
}

// RemoveResultingIncidents removes "resulting_incidents" edges to IncidentLink entities.
func (iriu *IncidentResourceImpactUpdate) RemoveResultingIncidents(i ...*IncidentLink) *IncidentResourceImpactUpdate {
	ids := make([]int, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return iriu.RemoveResultingIncidentIDs(ids...)
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (iriu *IncidentResourceImpactUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, iriu.sqlSave, iriu.mutation, iriu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (iriu *IncidentResourceImpactUpdate) SaveX(ctx context.Context) int {
	affected, err := iriu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (iriu *IncidentResourceImpactUpdate) Exec(ctx context.Context) error {
	_, err := iriu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (iriu *IncidentResourceImpactUpdate) ExecX(ctx context.Context) {
	if err := iriu.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (iriu *IncidentResourceImpactUpdate) check() error {
	if iriu.mutation.IncidentCleared() && len(iriu.mutation.IncidentIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "IncidentResourceImpact.incident"`)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (iriu *IncidentResourceImpactUpdate) Modify(modifiers ...func(u *sql.UpdateBuilder)) *IncidentResourceImpactUpdate {
	iriu.modifiers = append(iriu.modifiers, modifiers...)
	return iriu
}

func (iriu *IncidentResourceImpactUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := iriu.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(incidentresourceimpact.Table, incidentresourceimpact.Columns, sqlgraph.NewFieldSpec(incidentresourceimpact.FieldID, field.TypeUUID))
	if ps := iriu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if iriu.mutation.IncidentCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   incidentresourceimpact.IncidentTable,
			Columns: []string{incidentresourceimpact.IncidentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incident.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := iriu.mutation.IncidentIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   incidentresourceimpact.IncidentTable,
			Columns: []string{incidentresourceimpact.IncidentColumn},
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
	if iriu.mutation.ServiceCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   incidentresourceimpact.ServiceTable,
			Columns: []string{incidentresourceimpact.ServiceColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(service.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := iriu.mutation.ServiceIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   incidentresourceimpact.ServiceTable,
			Columns: []string{incidentresourceimpact.ServiceColumn},
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
	if iriu.mutation.FunctionalityCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   incidentresourceimpact.FunctionalityTable,
			Columns: []string{incidentresourceimpact.FunctionalityColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(functionality.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := iriu.mutation.FunctionalityIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   incidentresourceimpact.FunctionalityTable,
			Columns: []string{incidentresourceimpact.FunctionalityColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(functionality.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if iriu.mutation.ResultingIncidentsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   incidentresourceimpact.ResultingIncidentsTable,
			Columns: []string{incidentresourceimpact.ResultingIncidentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentlink.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := iriu.mutation.RemovedResultingIncidentsIDs(); len(nodes) > 0 && !iriu.mutation.ResultingIncidentsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   incidentresourceimpact.ResultingIncidentsTable,
			Columns: []string{incidentresourceimpact.ResultingIncidentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentlink.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := iriu.mutation.ResultingIncidentsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   incidentresourceimpact.ResultingIncidentsTable,
			Columns: []string{incidentresourceimpact.ResultingIncidentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentlink.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(iriu.modifiers...)
	if n, err = sqlgraph.UpdateNodes(ctx, iriu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{incidentresourceimpact.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	iriu.mutation.done = true
	return n, nil
}

// IncidentResourceImpactUpdateOne is the builder for updating a single IncidentResourceImpact entity.
type IncidentResourceImpactUpdateOne struct {
	config
	fields    []string
	hooks     []Hook
	mutation  *IncidentResourceImpactMutation
	modifiers []func(*sql.UpdateBuilder)
}

// SetIncidentID sets the "incident_id" field.
func (iriuo *IncidentResourceImpactUpdateOne) SetIncidentID(u uuid.UUID) *IncidentResourceImpactUpdateOne {
	iriuo.mutation.SetIncidentID(u)
	return iriuo
}

// SetNillableIncidentID sets the "incident_id" field if the given value is not nil.
func (iriuo *IncidentResourceImpactUpdateOne) SetNillableIncidentID(u *uuid.UUID) *IncidentResourceImpactUpdateOne {
	if u != nil {
		iriuo.SetIncidentID(*u)
	}
	return iriuo
}

// SetServiceID sets the "service_id" field.
func (iriuo *IncidentResourceImpactUpdateOne) SetServiceID(u uuid.UUID) *IncidentResourceImpactUpdateOne {
	iriuo.mutation.SetServiceID(u)
	return iriuo
}

// SetNillableServiceID sets the "service_id" field if the given value is not nil.
func (iriuo *IncidentResourceImpactUpdateOne) SetNillableServiceID(u *uuid.UUID) *IncidentResourceImpactUpdateOne {
	if u != nil {
		iriuo.SetServiceID(*u)
	}
	return iriuo
}

// ClearServiceID clears the value of the "service_id" field.
func (iriuo *IncidentResourceImpactUpdateOne) ClearServiceID() *IncidentResourceImpactUpdateOne {
	iriuo.mutation.ClearServiceID()
	return iriuo
}

// SetFunctionalityID sets the "functionality_id" field.
func (iriuo *IncidentResourceImpactUpdateOne) SetFunctionalityID(u uuid.UUID) *IncidentResourceImpactUpdateOne {
	iriuo.mutation.SetFunctionalityID(u)
	return iriuo
}

// SetNillableFunctionalityID sets the "functionality_id" field if the given value is not nil.
func (iriuo *IncidentResourceImpactUpdateOne) SetNillableFunctionalityID(u *uuid.UUID) *IncidentResourceImpactUpdateOne {
	if u != nil {
		iriuo.SetFunctionalityID(*u)
	}
	return iriuo
}

// ClearFunctionalityID clears the value of the "functionality_id" field.
func (iriuo *IncidentResourceImpactUpdateOne) ClearFunctionalityID() *IncidentResourceImpactUpdateOne {
	iriuo.mutation.ClearFunctionalityID()
	return iriuo
}

// SetIncident sets the "incident" edge to the Incident entity.
func (iriuo *IncidentResourceImpactUpdateOne) SetIncident(i *Incident) *IncidentResourceImpactUpdateOne {
	return iriuo.SetIncidentID(i.ID)
}

// SetService sets the "service" edge to the Service entity.
func (iriuo *IncidentResourceImpactUpdateOne) SetService(s *Service) *IncidentResourceImpactUpdateOne {
	return iriuo.SetServiceID(s.ID)
}

// SetFunctionality sets the "functionality" edge to the Functionality entity.
func (iriuo *IncidentResourceImpactUpdateOne) SetFunctionality(f *Functionality) *IncidentResourceImpactUpdateOne {
	return iriuo.SetFunctionalityID(f.ID)
}

// AddResultingIncidentIDs adds the "resulting_incidents" edge to the IncidentLink entity by IDs.
func (iriuo *IncidentResourceImpactUpdateOne) AddResultingIncidentIDs(ids ...int) *IncidentResourceImpactUpdateOne {
	iriuo.mutation.AddResultingIncidentIDs(ids...)
	return iriuo
}

// AddResultingIncidents adds the "resulting_incidents" edges to the IncidentLink entity.
func (iriuo *IncidentResourceImpactUpdateOne) AddResultingIncidents(i ...*IncidentLink) *IncidentResourceImpactUpdateOne {
	ids := make([]int, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return iriuo.AddResultingIncidentIDs(ids...)
}

// Mutation returns the IncidentResourceImpactMutation object of the builder.
func (iriuo *IncidentResourceImpactUpdateOne) Mutation() *IncidentResourceImpactMutation {
	return iriuo.mutation
}

// ClearIncident clears the "incident" edge to the Incident entity.
func (iriuo *IncidentResourceImpactUpdateOne) ClearIncident() *IncidentResourceImpactUpdateOne {
	iriuo.mutation.ClearIncident()
	return iriuo
}

// ClearService clears the "service" edge to the Service entity.
func (iriuo *IncidentResourceImpactUpdateOne) ClearService() *IncidentResourceImpactUpdateOne {
	iriuo.mutation.ClearService()
	return iriuo
}

// ClearFunctionality clears the "functionality" edge to the Functionality entity.
func (iriuo *IncidentResourceImpactUpdateOne) ClearFunctionality() *IncidentResourceImpactUpdateOne {
	iriuo.mutation.ClearFunctionality()
	return iriuo
}

// ClearResultingIncidents clears all "resulting_incidents" edges to the IncidentLink entity.
func (iriuo *IncidentResourceImpactUpdateOne) ClearResultingIncidents() *IncidentResourceImpactUpdateOne {
	iriuo.mutation.ClearResultingIncidents()
	return iriuo
}

// RemoveResultingIncidentIDs removes the "resulting_incidents" edge to IncidentLink entities by IDs.
func (iriuo *IncidentResourceImpactUpdateOne) RemoveResultingIncidentIDs(ids ...int) *IncidentResourceImpactUpdateOne {
	iriuo.mutation.RemoveResultingIncidentIDs(ids...)
	return iriuo
}

// RemoveResultingIncidents removes "resulting_incidents" edges to IncidentLink entities.
func (iriuo *IncidentResourceImpactUpdateOne) RemoveResultingIncidents(i ...*IncidentLink) *IncidentResourceImpactUpdateOne {
	ids := make([]int, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return iriuo.RemoveResultingIncidentIDs(ids...)
}

// Where appends a list predicates to the IncidentResourceImpactUpdate builder.
func (iriuo *IncidentResourceImpactUpdateOne) Where(ps ...predicate.IncidentResourceImpact) *IncidentResourceImpactUpdateOne {
	iriuo.mutation.Where(ps...)
	return iriuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (iriuo *IncidentResourceImpactUpdateOne) Select(field string, fields ...string) *IncidentResourceImpactUpdateOne {
	iriuo.fields = append([]string{field}, fields...)
	return iriuo
}

// Save executes the query and returns the updated IncidentResourceImpact entity.
func (iriuo *IncidentResourceImpactUpdateOne) Save(ctx context.Context) (*IncidentResourceImpact, error) {
	return withHooks(ctx, iriuo.sqlSave, iriuo.mutation, iriuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (iriuo *IncidentResourceImpactUpdateOne) SaveX(ctx context.Context) *IncidentResourceImpact {
	node, err := iriuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (iriuo *IncidentResourceImpactUpdateOne) Exec(ctx context.Context) error {
	_, err := iriuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (iriuo *IncidentResourceImpactUpdateOne) ExecX(ctx context.Context) {
	if err := iriuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (iriuo *IncidentResourceImpactUpdateOne) check() error {
	if iriuo.mutation.IncidentCleared() && len(iriuo.mutation.IncidentIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "IncidentResourceImpact.incident"`)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (iriuo *IncidentResourceImpactUpdateOne) Modify(modifiers ...func(u *sql.UpdateBuilder)) *IncidentResourceImpactUpdateOne {
	iriuo.modifiers = append(iriuo.modifiers, modifiers...)
	return iriuo
}

func (iriuo *IncidentResourceImpactUpdateOne) sqlSave(ctx context.Context) (_node *IncidentResourceImpact, err error) {
	if err := iriuo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(incidentresourceimpact.Table, incidentresourceimpact.Columns, sqlgraph.NewFieldSpec(incidentresourceimpact.FieldID, field.TypeUUID))
	id, ok := iriuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "IncidentResourceImpact.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := iriuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, incidentresourceimpact.FieldID)
		for _, f := range fields {
			if !incidentresourceimpact.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != incidentresourceimpact.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := iriuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if iriuo.mutation.IncidentCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   incidentresourceimpact.IncidentTable,
			Columns: []string{incidentresourceimpact.IncidentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incident.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := iriuo.mutation.IncidentIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   incidentresourceimpact.IncidentTable,
			Columns: []string{incidentresourceimpact.IncidentColumn},
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
	if iriuo.mutation.ServiceCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   incidentresourceimpact.ServiceTable,
			Columns: []string{incidentresourceimpact.ServiceColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(service.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := iriuo.mutation.ServiceIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   incidentresourceimpact.ServiceTable,
			Columns: []string{incidentresourceimpact.ServiceColumn},
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
	if iriuo.mutation.FunctionalityCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   incidentresourceimpact.FunctionalityTable,
			Columns: []string{incidentresourceimpact.FunctionalityColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(functionality.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := iriuo.mutation.FunctionalityIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   incidentresourceimpact.FunctionalityTable,
			Columns: []string{incidentresourceimpact.FunctionalityColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(functionality.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if iriuo.mutation.ResultingIncidentsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   incidentresourceimpact.ResultingIncidentsTable,
			Columns: []string{incidentresourceimpact.ResultingIncidentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentlink.FieldID, field.TypeInt),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := iriuo.mutation.RemovedResultingIncidentsIDs(); len(nodes) > 0 && !iriuo.mutation.ResultingIncidentsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   incidentresourceimpact.ResultingIncidentsTable,
			Columns: []string{incidentresourceimpact.ResultingIncidentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentlink.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := iriuo.mutation.ResultingIncidentsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   incidentresourceimpact.ResultingIncidentsTable,
			Columns: []string{incidentresourceimpact.ResultingIncidentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentlink.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(iriuo.modifiers...)
	_node = &IncidentResourceImpact{config: iriuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, iriuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{incidentresourceimpact.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	iriuo.mutation.done = true
	return _node, nil
}