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
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/incidentmilestone"
	"github.com/rezible/rezible/ent/predicate"
)

// IncidentMilestoneUpdate is the builder for updating IncidentMilestone entities.
type IncidentMilestoneUpdate struct {
	config
	hooks     []Hook
	mutation  *IncidentMilestoneMutation
	modifiers []func(*sql.UpdateBuilder)
}

// Where appends a list predicates to the IncidentMilestoneUpdate builder.
func (imu *IncidentMilestoneUpdate) Where(ps ...predicate.IncidentMilestone) *IncidentMilestoneUpdate {
	imu.mutation.Where(ps...)
	return imu
}

// SetIncidentID sets the "incident_id" field.
func (imu *IncidentMilestoneUpdate) SetIncidentID(u uuid.UUID) *IncidentMilestoneUpdate {
	imu.mutation.SetIncidentID(u)
	return imu
}

// SetNillableIncidentID sets the "incident_id" field if the given value is not nil.
func (imu *IncidentMilestoneUpdate) SetNillableIncidentID(u *uuid.UUID) *IncidentMilestoneUpdate {
	if u != nil {
		imu.SetIncidentID(*u)
	}
	return imu
}

// SetType sets the "type" field.
func (imu *IncidentMilestoneUpdate) SetType(i incidentmilestone.Type) *IncidentMilestoneUpdate {
	imu.mutation.SetType(i)
	return imu
}

// SetNillableType sets the "type" field if the given value is not nil.
func (imu *IncidentMilestoneUpdate) SetNillableType(i *incidentmilestone.Type) *IncidentMilestoneUpdate {
	if i != nil {
		imu.SetType(*i)
	}
	return imu
}

// SetTime sets the "time" field.
func (imu *IncidentMilestoneUpdate) SetTime(t time.Time) *IncidentMilestoneUpdate {
	imu.mutation.SetTime(t)
	return imu
}

// SetNillableTime sets the "time" field if the given value is not nil.
func (imu *IncidentMilestoneUpdate) SetNillableTime(t *time.Time) *IncidentMilestoneUpdate {
	if t != nil {
		imu.SetTime(*t)
	}
	return imu
}

// SetIncident sets the "incident" edge to the Incident entity.
func (imu *IncidentMilestoneUpdate) SetIncident(i *Incident) *IncidentMilestoneUpdate {
	return imu.SetIncidentID(i.ID)
}

// Mutation returns the IncidentMilestoneMutation object of the builder.
func (imu *IncidentMilestoneUpdate) Mutation() *IncidentMilestoneMutation {
	return imu.mutation
}

// ClearIncident clears the "incident" edge to the Incident entity.
func (imu *IncidentMilestoneUpdate) ClearIncident() *IncidentMilestoneUpdate {
	imu.mutation.ClearIncident()
	return imu
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (imu *IncidentMilestoneUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, imu.sqlSave, imu.mutation, imu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (imu *IncidentMilestoneUpdate) SaveX(ctx context.Context) int {
	affected, err := imu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (imu *IncidentMilestoneUpdate) Exec(ctx context.Context) error {
	_, err := imu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (imu *IncidentMilestoneUpdate) ExecX(ctx context.Context) {
	if err := imu.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (imu *IncidentMilestoneUpdate) check() error {
	if v, ok := imu.mutation.GetType(); ok {
		if err := incidentmilestone.TypeValidator(v); err != nil {
			return &ValidationError{Name: "type", err: fmt.Errorf(`ent: validator failed for field "IncidentMilestone.type": %w`, err)}
		}
	}
	if imu.mutation.IncidentCleared() && len(imu.mutation.IncidentIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "IncidentMilestone.incident"`)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (imu *IncidentMilestoneUpdate) Modify(modifiers ...func(u *sql.UpdateBuilder)) *IncidentMilestoneUpdate {
	imu.modifiers = append(imu.modifiers, modifiers...)
	return imu
}

func (imu *IncidentMilestoneUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := imu.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(incidentmilestone.Table, incidentmilestone.Columns, sqlgraph.NewFieldSpec(incidentmilestone.FieldID, field.TypeUUID))
	if ps := imu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := imu.mutation.GetType(); ok {
		_spec.SetField(incidentmilestone.FieldType, field.TypeEnum, value)
	}
	if value, ok := imu.mutation.Time(); ok {
		_spec.SetField(incidentmilestone.FieldTime, field.TypeTime, value)
	}
	if imu.mutation.IncidentCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   incidentmilestone.IncidentTable,
			Columns: []string{incidentmilestone.IncidentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incident.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := imu.mutation.IncidentIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   incidentmilestone.IncidentTable,
			Columns: []string{incidentmilestone.IncidentColumn},
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
	_spec.AddModifiers(imu.modifiers...)
	if n, err = sqlgraph.UpdateNodes(ctx, imu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{incidentmilestone.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	imu.mutation.done = true
	return n, nil
}

// IncidentMilestoneUpdateOne is the builder for updating a single IncidentMilestone entity.
type IncidentMilestoneUpdateOne struct {
	config
	fields    []string
	hooks     []Hook
	mutation  *IncidentMilestoneMutation
	modifiers []func(*sql.UpdateBuilder)
}

// SetIncidentID sets the "incident_id" field.
func (imuo *IncidentMilestoneUpdateOne) SetIncidentID(u uuid.UUID) *IncidentMilestoneUpdateOne {
	imuo.mutation.SetIncidentID(u)
	return imuo
}

// SetNillableIncidentID sets the "incident_id" field if the given value is not nil.
func (imuo *IncidentMilestoneUpdateOne) SetNillableIncidentID(u *uuid.UUID) *IncidentMilestoneUpdateOne {
	if u != nil {
		imuo.SetIncidentID(*u)
	}
	return imuo
}

// SetType sets the "type" field.
func (imuo *IncidentMilestoneUpdateOne) SetType(i incidentmilestone.Type) *IncidentMilestoneUpdateOne {
	imuo.mutation.SetType(i)
	return imuo
}

// SetNillableType sets the "type" field if the given value is not nil.
func (imuo *IncidentMilestoneUpdateOne) SetNillableType(i *incidentmilestone.Type) *IncidentMilestoneUpdateOne {
	if i != nil {
		imuo.SetType(*i)
	}
	return imuo
}

// SetTime sets the "time" field.
func (imuo *IncidentMilestoneUpdateOne) SetTime(t time.Time) *IncidentMilestoneUpdateOne {
	imuo.mutation.SetTime(t)
	return imuo
}

// SetNillableTime sets the "time" field if the given value is not nil.
func (imuo *IncidentMilestoneUpdateOne) SetNillableTime(t *time.Time) *IncidentMilestoneUpdateOne {
	if t != nil {
		imuo.SetTime(*t)
	}
	return imuo
}

// SetIncident sets the "incident" edge to the Incident entity.
func (imuo *IncidentMilestoneUpdateOne) SetIncident(i *Incident) *IncidentMilestoneUpdateOne {
	return imuo.SetIncidentID(i.ID)
}

// Mutation returns the IncidentMilestoneMutation object of the builder.
func (imuo *IncidentMilestoneUpdateOne) Mutation() *IncidentMilestoneMutation {
	return imuo.mutation
}

// ClearIncident clears the "incident" edge to the Incident entity.
func (imuo *IncidentMilestoneUpdateOne) ClearIncident() *IncidentMilestoneUpdateOne {
	imuo.mutation.ClearIncident()
	return imuo
}

// Where appends a list predicates to the IncidentMilestoneUpdate builder.
func (imuo *IncidentMilestoneUpdateOne) Where(ps ...predicate.IncidentMilestone) *IncidentMilestoneUpdateOne {
	imuo.mutation.Where(ps...)
	return imuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (imuo *IncidentMilestoneUpdateOne) Select(field string, fields ...string) *IncidentMilestoneUpdateOne {
	imuo.fields = append([]string{field}, fields...)
	return imuo
}

// Save executes the query and returns the updated IncidentMilestone entity.
func (imuo *IncidentMilestoneUpdateOne) Save(ctx context.Context) (*IncidentMilestone, error) {
	return withHooks(ctx, imuo.sqlSave, imuo.mutation, imuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (imuo *IncidentMilestoneUpdateOne) SaveX(ctx context.Context) *IncidentMilestone {
	node, err := imuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (imuo *IncidentMilestoneUpdateOne) Exec(ctx context.Context) error {
	_, err := imuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (imuo *IncidentMilestoneUpdateOne) ExecX(ctx context.Context) {
	if err := imuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (imuo *IncidentMilestoneUpdateOne) check() error {
	if v, ok := imuo.mutation.GetType(); ok {
		if err := incidentmilestone.TypeValidator(v); err != nil {
			return &ValidationError{Name: "type", err: fmt.Errorf(`ent: validator failed for field "IncidentMilestone.type": %w`, err)}
		}
	}
	if imuo.mutation.IncidentCleared() && len(imuo.mutation.IncidentIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "IncidentMilestone.incident"`)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (imuo *IncidentMilestoneUpdateOne) Modify(modifiers ...func(u *sql.UpdateBuilder)) *IncidentMilestoneUpdateOne {
	imuo.modifiers = append(imuo.modifiers, modifiers...)
	return imuo
}

func (imuo *IncidentMilestoneUpdateOne) sqlSave(ctx context.Context) (_node *IncidentMilestone, err error) {
	if err := imuo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(incidentmilestone.Table, incidentmilestone.Columns, sqlgraph.NewFieldSpec(incidentmilestone.FieldID, field.TypeUUID))
	id, ok := imuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "IncidentMilestone.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := imuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, incidentmilestone.FieldID)
		for _, f := range fields {
			if !incidentmilestone.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != incidentmilestone.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := imuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := imuo.mutation.GetType(); ok {
		_spec.SetField(incidentmilestone.FieldType, field.TypeEnum, value)
	}
	if value, ok := imuo.mutation.Time(); ok {
		_spec.SetField(incidentmilestone.FieldTime, field.TypeTime, value)
	}
	if imuo.mutation.IncidentCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   incidentmilestone.IncidentTable,
			Columns: []string{incidentmilestone.IncidentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incident.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := imuo.mutation.IncidentIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   incidentmilestone.IncidentTable,
			Columns: []string{incidentmilestone.IncidentColumn},
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
	_spec.AddModifiers(imuo.modifiers...)
	_node = &IncidentMilestone{config: imuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, imuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{incidentmilestone.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	imuo.mutation.done = true
	return _node, nil
}
