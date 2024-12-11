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
	"github.com/twohundreds/rezible/ent/oncallalert"
	"github.com/twohundreds/rezible/ent/oncallalertinstance"
	"github.com/twohundreds/rezible/ent/oncallroster"
	"github.com/twohundreds/rezible/ent/predicate"
)

// OncallAlertUpdate is the builder for updating OncallAlert entities.
type OncallAlertUpdate struct {
	config
	hooks     []Hook
	mutation  *OncallAlertMutation
	modifiers []func(*sql.UpdateBuilder)
}

// Where appends a list predicates to the OncallAlertUpdate builder.
func (oau *OncallAlertUpdate) Where(ps ...predicate.OncallAlert) *OncallAlertUpdate {
	oau.mutation.Where(ps...)
	return oau
}

// SetRosterID sets the "roster_id" field.
func (oau *OncallAlertUpdate) SetRosterID(u uuid.UUID) *OncallAlertUpdate {
	oau.mutation.SetRosterID(u)
	return oau
}

// SetNillableRosterID sets the "roster_id" field if the given value is not nil.
func (oau *OncallAlertUpdate) SetNillableRosterID(u *uuid.UUID) *OncallAlertUpdate {
	if u != nil {
		oau.SetRosterID(*u)
	}
	return oau
}

// SetName sets the "name" field.
func (oau *OncallAlertUpdate) SetName(s string) *OncallAlertUpdate {
	oau.mutation.SetName(s)
	return oau
}

// SetNillableName sets the "name" field if the given value is not nil.
func (oau *OncallAlertUpdate) SetNillableName(s *string) *OncallAlertUpdate {
	if s != nil {
		oau.SetName(*s)
	}
	return oau
}

// SetTimestamp sets the "timestamp" field.
func (oau *OncallAlertUpdate) SetTimestamp(t time.Time) *OncallAlertUpdate {
	oau.mutation.SetTimestamp(t)
	return oau
}

// SetNillableTimestamp sets the "timestamp" field if the given value is not nil.
func (oau *OncallAlertUpdate) SetNillableTimestamp(t *time.Time) *OncallAlertUpdate {
	if t != nil {
		oau.SetTimestamp(*t)
	}
	return oau
}

// AddInstanceIDs adds the "instances" edge to the OncallAlertInstance entity by IDs.
func (oau *OncallAlertUpdate) AddInstanceIDs(ids ...uuid.UUID) *OncallAlertUpdate {
	oau.mutation.AddInstanceIDs(ids...)
	return oau
}

// AddInstances adds the "instances" edges to the OncallAlertInstance entity.
func (oau *OncallAlertUpdate) AddInstances(o ...*OncallAlertInstance) *OncallAlertUpdate {
	ids := make([]uuid.UUID, len(o))
	for i := range o {
		ids[i] = o[i].ID
	}
	return oau.AddInstanceIDs(ids...)
}

// SetRoster sets the "roster" edge to the OncallRoster entity.
func (oau *OncallAlertUpdate) SetRoster(o *OncallRoster) *OncallAlertUpdate {
	return oau.SetRosterID(o.ID)
}

// Mutation returns the OncallAlertMutation object of the builder.
func (oau *OncallAlertUpdate) Mutation() *OncallAlertMutation {
	return oau.mutation
}

// ClearInstances clears all "instances" edges to the OncallAlertInstance entity.
func (oau *OncallAlertUpdate) ClearInstances() *OncallAlertUpdate {
	oau.mutation.ClearInstances()
	return oau
}

// RemoveInstanceIDs removes the "instances" edge to OncallAlertInstance entities by IDs.
func (oau *OncallAlertUpdate) RemoveInstanceIDs(ids ...uuid.UUID) *OncallAlertUpdate {
	oau.mutation.RemoveInstanceIDs(ids...)
	return oau
}

// RemoveInstances removes "instances" edges to OncallAlertInstance entities.
func (oau *OncallAlertUpdate) RemoveInstances(o ...*OncallAlertInstance) *OncallAlertUpdate {
	ids := make([]uuid.UUID, len(o))
	for i := range o {
		ids[i] = o[i].ID
	}
	return oau.RemoveInstanceIDs(ids...)
}

// ClearRoster clears the "roster" edge to the OncallRoster entity.
func (oau *OncallAlertUpdate) ClearRoster() *OncallAlertUpdate {
	oau.mutation.ClearRoster()
	return oau
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (oau *OncallAlertUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, oau.sqlSave, oau.mutation, oau.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (oau *OncallAlertUpdate) SaveX(ctx context.Context) int {
	affected, err := oau.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (oau *OncallAlertUpdate) Exec(ctx context.Context) error {
	_, err := oau.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (oau *OncallAlertUpdate) ExecX(ctx context.Context) {
	if err := oau.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (oau *OncallAlertUpdate) check() error {
	if oau.mutation.RosterCleared() && len(oau.mutation.RosterIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "OncallAlert.roster"`)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (oau *OncallAlertUpdate) Modify(modifiers ...func(u *sql.UpdateBuilder)) *OncallAlertUpdate {
	oau.modifiers = append(oau.modifiers, modifiers...)
	return oau
}

func (oau *OncallAlertUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := oau.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(oncallalert.Table, oncallalert.Columns, sqlgraph.NewFieldSpec(oncallalert.FieldID, field.TypeUUID))
	if ps := oau.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := oau.mutation.Name(); ok {
		_spec.SetField(oncallalert.FieldName, field.TypeString, value)
	}
	if value, ok := oau.mutation.Timestamp(); ok {
		_spec.SetField(oncallalert.FieldTimestamp, field.TypeTime, value)
	}
	if oau.mutation.InstancesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   oncallalert.InstancesTable,
			Columns: []string{oncallalert.InstancesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(oncallalertinstance.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := oau.mutation.RemovedInstancesIDs(); len(nodes) > 0 && !oau.mutation.InstancesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   oncallalert.InstancesTable,
			Columns: []string{oncallalert.InstancesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(oncallalertinstance.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := oau.mutation.InstancesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   oncallalert.InstancesTable,
			Columns: []string{oncallalert.InstancesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(oncallalertinstance.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if oau.mutation.RosterCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   oncallalert.RosterTable,
			Columns: []string{oncallalert.RosterColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(oncallroster.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := oau.mutation.RosterIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   oncallalert.RosterTable,
			Columns: []string{oncallalert.RosterColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(oncallroster.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(oau.modifiers...)
	if n, err = sqlgraph.UpdateNodes(ctx, oau.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{oncallalert.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	oau.mutation.done = true
	return n, nil
}

// OncallAlertUpdateOne is the builder for updating a single OncallAlert entity.
type OncallAlertUpdateOne struct {
	config
	fields    []string
	hooks     []Hook
	mutation  *OncallAlertMutation
	modifiers []func(*sql.UpdateBuilder)
}

// SetRosterID sets the "roster_id" field.
func (oauo *OncallAlertUpdateOne) SetRosterID(u uuid.UUID) *OncallAlertUpdateOne {
	oauo.mutation.SetRosterID(u)
	return oauo
}

// SetNillableRosterID sets the "roster_id" field if the given value is not nil.
func (oauo *OncallAlertUpdateOne) SetNillableRosterID(u *uuid.UUID) *OncallAlertUpdateOne {
	if u != nil {
		oauo.SetRosterID(*u)
	}
	return oauo
}

// SetName sets the "name" field.
func (oauo *OncallAlertUpdateOne) SetName(s string) *OncallAlertUpdateOne {
	oauo.mutation.SetName(s)
	return oauo
}

// SetNillableName sets the "name" field if the given value is not nil.
func (oauo *OncallAlertUpdateOne) SetNillableName(s *string) *OncallAlertUpdateOne {
	if s != nil {
		oauo.SetName(*s)
	}
	return oauo
}

// SetTimestamp sets the "timestamp" field.
func (oauo *OncallAlertUpdateOne) SetTimestamp(t time.Time) *OncallAlertUpdateOne {
	oauo.mutation.SetTimestamp(t)
	return oauo
}

// SetNillableTimestamp sets the "timestamp" field if the given value is not nil.
func (oauo *OncallAlertUpdateOne) SetNillableTimestamp(t *time.Time) *OncallAlertUpdateOne {
	if t != nil {
		oauo.SetTimestamp(*t)
	}
	return oauo
}

// AddInstanceIDs adds the "instances" edge to the OncallAlertInstance entity by IDs.
func (oauo *OncallAlertUpdateOne) AddInstanceIDs(ids ...uuid.UUID) *OncallAlertUpdateOne {
	oauo.mutation.AddInstanceIDs(ids...)
	return oauo
}

// AddInstances adds the "instances" edges to the OncallAlertInstance entity.
func (oauo *OncallAlertUpdateOne) AddInstances(o ...*OncallAlertInstance) *OncallAlertUpdateOne {
	ids := make([]uuid.UUID, len(o))
	for i := range o {
		ids[i] = o[i].ID
	}
	return oauo.AddInstanceIDs(ids...)
}

// SetRoster sets the "roster" edge to the OncallRoster entity.
func (oauo *OncallAlertUpdateOne) SetRoster(o *OncallRoster) *OncallAlertUpdateOne {
	return oauo.SetRosterID(o.ID)
}

// Mutation returns the OncallAlertMutation object of the builder.
func (oauo *OncallAlertUpdateOne) Mutation() *OncallAlertMutation {
	return oauo.mutation
}

// ClearInstances clears all "instances" edges to the OncallAlertInstance entity.
func (oauo *OncallAlertUpdateOne) ClearInstances() *OncallAlertUpdateOne {
	oauo.mutation.ClearInstances()
	return oauo
}

// RemoveInstanceIDs removes the "instances" edge to OncallAlertInstance entities by IDs.
func (oauo *OncallAlertUpdateOne) RemoveInstanceIDs(ids ...uuid.UUID) *OncallAlertUpdateOne {
	oauo.mutation.RemoveInstanceIDs(ids...)
	return oauo
}

// RemoveInstances removes "instances" edges to OncallAlertInstance entities.
func (oauo *OncallAlertUpdateOne) RemoveInstances(o ...*OncallAlertInstance) *OncallAlertUpdateOne {
	ids := make([]uuid.UUID, len(o))
	for i := range o {
		ids[i] = o[i].ID
	}
	return oauo.RemoveInstanceIDs(ids...)
}

// ClearRoster clears the "roster" edge to the OncallRoster entity.
func (oauo *OncallAlertUpdateOne) ClearRoster() *OncallAlertUpdateOne {
	oauo.mutation.ClearRoster()
	return oauo
}

// Where appends a list predicates to the OncallAlertUpdate builder.
func (oauo *OncallAlertUpdateOne) Where(ps ...predicate.OncallAlert) *OncallAlertUpdateOne {
	oauo.mutation.Where(ps...)
	return oauo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (oauo *OncallAlertUpdateOne) Select(field string, fields ...string) *OncallAlertUpdateOne {
	oauo.fields = append([]string{field}, fields...)
	return oauo
}

// Save executes the query and returns the updated OncallAlert entity.
func (oauo *OncallAlertUpdateOne) Save(ctx context.Context) (*OncallAlert, error) {
	return withHooks(ctx, oauo.sqlSave, oauo.mutation, oauo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (oauo *OncallAlertUpdateOne) SaveX(ctx context.Context) *OncallAlert {
	node, err := oauo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (oauo *OncallAlertUpdateOne) Exec(ctx context.Context) error {
	_, err := oauo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (oauo *OncallAlertUpdateOne) ExecX(ctx context.Context) {
	if err := oauo.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (oauo *OncallAlertUpdateOne) check() error {
	if oauo.mutation.RosterCleared() && len(oauo.mutation.RosterIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "OncallAlert.roster"`)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (oauo *OncallAlertUpdateOne) Modify(modifiers ...func(u *sql.UpdateBuilder)) *OncallAlertUpdateOne {
	oauo.modifiers = append(oauo.modifiers, modifiers...)
	return oauo
}

func (oauo *OncallAlertUpdateOne) sqlSave(ctx context.Context) (_node *OncallAlert, err error) {
	if err := oauo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(oncallalert.Table, oncallalert.Columns, sqlgraph.NewFieldSpec(oncallalert.FieldID, field.TypeUUID))
	id, ok := oauo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "OncallAlert.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := oauo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, oncallalert.FieldID)
		for _, f := range fields {
			if !oncallalert.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != oncallalert.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := oauo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := oauo.mutation.Name(); ok {
		_spec.SetField(oncallalert.FieldName, field.TypeString, value)
	}
	if value, ok := oauo.mutation.Timestamp(); ok {
		_spec.SetField(oncallalert.FieldTimestamp, field.TypeTime, value)
	}
	if oauo.mutation.InstancesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   oncallalert.InstancesTable,
			Columns: []string{oncallalert.InstancesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(oncallalertinstance.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := oauo.mutation.RemovedInstancesIDs(); len(nodes) > 0 && !oauo.mutation.InstancesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   oncallalert.InstancesTable,
			Columns: []string{oncallalert.InstancesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(oncallalertinstance.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := oauo.mutation.InstancesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   oncallalert.InstancesTable,
			Columns: []string{oncallalert.InstancesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(oncallalertinstance.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if oauo.mutation.RosterCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   oncallalert.RosterTable,
			Columns: []string{oncallalert.RosterColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(oncallroster.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := oauo.mutation.RosterIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   oncallalert.RosterTable,
			Columns: []string{oncallalert.RosterColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(oncallroster.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(oauo.modifiers...)
	_node = &OncallAlert{config: oauo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, oauo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{oncallalert.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	oauo.mutation.done = true
	return _node, nil
}
