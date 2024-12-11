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
	"github.com/twohundreds/rezible/ent/incident"
	"github.com/twohundreds/rezible/ent/incidentrole"
	"github.com/twohundreds/rezible/ent/incidentroleassignment"
	"github.com/twohundreds/rezible/ent/predicate"
	"github.com/twohundreds/rezible/ent/user"
)

// IncidentRoleAssignmentUpdate is the builder for updating IncidentRoleAssignment entities.
type IncidentRoleAssignmentUpdate struct {
	config
	hooks     []Hook
	mutation  *IncidentRoleAssignmentMutation
	modifiers []func(*sql.UpdateBuilder)
}

// Where appends a list predicates to the IncidentRoleAssignmentUpdate builder.
func (irau *IncidentRoleAssignmentUpdate) Where(ps ...predicate.IncidentRoleAssignment) *IncidentRoleAssignmentUpdate {
	irau.mutation.Where(ps...)
	return irau
}

// SetRoleID sets the "role_id" field.
func (irau *IncidentRoleAssignmentUpdate) SetRoleID(u uuid.UUID) *IncidentRoleAssignmentUpdate {
	irau.mutation.SetRoleID(u)
	return irau
}

// SetNillableRoleID sets the "role_id" field if the given value is not nil.
func (irau *IncidentRoleAssignmentUpdate) SetNillableRoleID(u *uuid.UUID) *IncidentRoleAssignmentUpdate {
	if u != nil {
		irau.SetRoleID(*u)
	}
	return irau
}

// SetIncidentID sets the "incident_id" field.
func (irau *IncidentRoleAssignmentUpdate) SetIncidentID(u uuid.UUID) *IncidentRoleAssignmentUpdate {
	irau.mutation.SetIncidentID(u)
	return irau
}

// SetNillableIncidentID sets the "incident_id" field if the given value is not nil.
func (irau *IncidentRoleAssignmentUpdate) SetNillableIncidentID(u *uuid.UUID) *IncidentRoleAssignmentUpdate {
	if u != nil {
		irau.SetIncidentID(*u)
	}
	return irau
}

// SetUserID sets the "user_id" field.
func (irau *IncidentRoleAssignmentUpdate) SetUserID(u uuid.UUID) *IncidentRoleAssignmentUpdate {
	irau.mutation.SetUserID(u)
	return irau
}

// SetNillableUserID sets the "user_id" field if the given value is not nil.
func (irau *IncidentRoleAssignmentUpdate) SetNillableUserID(u *uuid.UUID) *IncidentRoleAssignmentUpdate {
	if u != nil {
		irau.SetUserID(*u)
	}
	return irau
}

// SetRole sets the "role" edge to the IncidentRole entity.
func (irau *IncidentRoleAssignmentUpdate) SetRole(i *IncidentRole) *IncidentRoleAssignmentUpdate {
	return irau.SetRoleID(i.ID)
}

// SetIncident sets the "incident" edge to the Incident entity.
func (irau *IncidentRoleAssignmentUpdate) SetIncident(i *Incident) *IncidentRoleAssignmentUpdate {
	return irau.SetIncidentID(i.ID)
}

// SetUser sets the "user" edge to the User entity.
func (irau *IncidentRoleAssignmentUpdate) SetUser(u *User) *IncidentRoleAssignmentUpdate {
	return irau.SetUserID(u.ID)
}

// Mutation returns the IncidentRoleAssignmentMutation object of the builder.
func (irau *IncidentRoleAssignmentUpdate) Mutation() *IncidentRoleAssignmentMutation {
	return irau.mutation
}

// ClearRole clears the "role" edge to the IncidentRole entity.
func (irau *IncidentRoleAssignmentUpdate) ClearRole() *IncidentRoleAssignmentUpdate {
	irau.mutation.ClearRole()
	return irau
}

// ClearIncident clears the "incident" edge to the Incident entity.
func (irau *IncidentRoleAssignmentUpdate) ClearIncident() *IncidentRoleAssignmentUpdate {
	irau.mutation.ClearIncident()
	return irau
}

// ClearUser clears the "user" edge to the User entity.
func (irau *IncidentRoleAssignmentUpdate) ClearUser() *IncidentRoleAssignmentUpdate {
	irau.mutation.ClearUser()
	return irau
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (irau *IncidentRoleAssignmentUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, irau.sqlSave, irau.mutation, irau.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (irau *IncidentRoleAssignmentUpdate) SaveX(ctx context.Context) int {
	affected, err := irau.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (irau *IncidentRoleAssignmentUpdate) Exec(ctx context.Context) error {
	_, err := irau.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (irau *IncidentRoleAssignmentUpdate) ExecX(ctx context.Context) {
	if err := irau.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (irau *IncidentRoleAssignmentUpdate) check() error {
	if irau.mutation.RoleCleared() && len(irau.mutation.RoleIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "IncidentRoleAssignment.role"`)
	}
	if irau.mutation.IncidentCleared() && len(irau.mutation.IncidentIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "IncidentRoleAssignment.incident"`)
	}
	if irau.mutation.UserCleared() && len(irau.mutation.UserIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "IncidentRoleAssignment.user"`)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (irau *IncidentRoleAssignmentUpdate) Modify(modifiers ...func(u *sql.UpdateBuilder)) *IncidentRoleAssignmentUpdate {
	irau.modifiers = append(irau.modifiers, modifiers...)
	return irau
}

func (irau *IncidentRoleAssignmentUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := irau.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(incidentroleassignment.Table, incidentroleassignment.Columns, sqlgraph.NewFieldSpec(incidentroleassignment.FieldID, field.TypeUUID))
	if ps := irau.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if irau.mutation.RoleCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   incidentroleassignment.RoleTable,
			Columns: []string{incidentroleassignment.RoleColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentrole.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := irau.mutation.RoleIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   incidentroleassignment.RoleTable,
			Columns: []string{incidentroleassignment.RoleColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentrole.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if irau.mutation.IncidentCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   incidentroleassignment.IncidentTable,
			Columns: []string{incidentroleassignment.IncidentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incident.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := irau.mutation.IncidentIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   incidentroleassignment.IncidentTable,
			Columns: []string{incidentroleassignment.IncidentColumn},
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
	if irau.mutation.UserCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   incidentroleassignment.UserTable,
			Columns: []string{incidentroleassignment.UserColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(user.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := irau.mutation.UserIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   incidentroleassignment.UserTable,
			Columns: []string{incidentroleassignment.UserColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(user.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(irau.modifiers...)
	if n, err = sqlgraph.UpdateNodes(ctx, irau.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{incidentroleassignment.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	irau.mutation.done = true
	return n, nil
}

// IncidentRoleAssignmentUpdateOne is the builder for updating a single IncidentRoleAssignment entity.
type IncidentRoleAssignmentUpdateOne struct {
	config
	fields    []string
	hooks     []Hook
	mutation  *IncidentRoleAssignmentMutation
	modifiers []func(*sql.UpdateBuilder)
}

// SetRoleID sets the "role_id" field.
func (irauo *IncidentRoleAssignmentUpdateOne) SetRoleID(u uuid.UUID) *IncidentRoleAssignmentUpdateOne {
	irauo.mutation.SetRoleID(u)
	return irauo
}

// SetNillableRoleID sets the "role_id" field if the given value is not nil.
func (irauo *IncidentRoleAssignmentUpdateOne) SetNillableRoleID(u *uuid.UUID) *IncidentRoleAssignmentUpdateOne {
	if u != nil {
		irauo.SetRoleID(*u)
	}
	return irauo
}

// SetIncidentID sets the "incident_id" field.
func (irauo *IncidentRoleAssignmentUpdateOne) SetIncidentID(u uuid.UUID) *IncidentRoleAssignmentUpdateOne {
	irauo.mutation.SetIncidentID(u)
	return irauo
}

// SetNillableIncidentID sets the "incident_id" field if the given value is not nil.
func (irauo *IncidentRoleAssignmentUpdateOne) SetNillableIncidentID(u *uuid.UUID) *IncidentRoleAssignmentUpdateOne {
	if u != nil {
		irauo.SetIncidentID(*u)
	}
	return irauo
}

// SetUserID sets the "user_id" field.
func (irauo *IncidentRoleAssignmentUpdateOne) SetUserID(u uuid.UUID) *IncidentRoleAssignmentUpdateOne {
	irauo.mutation.SetUserID(u)
	return irauo
}

// SetNillableUserID sets the "user_id" field if the given value is not nil.
func (irauo *IncidentRoleAssignmentUpdateOne) SetNillableUserID(u *uuid.UUID) *IncidentRoleAssignmentUpdateOne {
	if u != nil {
		irauo.SetUserID(*u)
	}
	return irauo
}

// SetRole sets the "role" edge to the IncidentRole entity.
func (irauo *IncidentRoleAssignmentUpdateOne) SetRole(i *IncidentRole) *IncidentRoleAssignmentUpdateOne {
	return irauo.SetRoleID(i.ID)
}

// SetIncident sets the "incident" edge to the Incident entity.
func (irauo *IncidentRoleAssignmentUpdateOne) SetIncident(i *Incident) *IncidentRoleAssignmentUpdateOne {
	return irauo.SetIncidentID(i.ID)
}

// SetUser sets the "user" edge to the User entity.
func (irauo *IncidentRoleAssignmentUpdateOne) SetUser(u *User) *IncidentRoleAssignmentUpdateOne {
	return irauo.SetUserID(u.ID)
}

// Mutation returns the IncidentRoleAssignmentMutation object of the builder.
func (irauo *IncidentRoleAssignmentUpdateOne) Mutation() *IncidentRoleAssignmentMutation {
	return irauo.mutation
}

// ClearRole clears the "role" edge to the IncidentRole entity.
func (irauo *IncidentRoleAssignmentUpdateOne) ClearRole() *IncidentRoleAssignmentUpdateOne {
	irauo.mutation.ClearRole()
	return irauo
}

// ClearIncident clears the "incident" edge to the Incident entity.
func (irauo *IncidentRoleAssignmentUpdateOne) ClearIncident() *IncidentRoleAssignmentUpdateOne {
	irauo.mutation.ClearIncident()
	return irauo
}

// ClearUser clears the "user" edge to the User entity.
func (irauo *IncidentRoleAssignmentUpdateOne) ClearUser() *IncidentRoleAssignmentUpdateOne {
	irauo.mutation.ClearUser()
	return irauo
}

// Where appends a list predicates to the IncidentRoleAssignmentUpdate builder.
func (irauo *IncidentRoleAssignmentUpdateOne) Where(ps ...predicate.IncidentRoleAssignment) *IncidentRoleAssignmentUpdateOne {
	irauo.mutation.Where(ps...)
	return irauo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (irauo *IncidentRoleAssignmentUpdateOne) Select(field string, fields ...string) *IncidentRoleAssignmentUpdateOne {
	irauo.fields = append([]string{field}, fields...)
	return irauo
}

// Save executes the query and returns the updated IncidentRoleAssignment entity.
func (irauo *IncidentRoleAssignmentUpdateOne) Save(ctx context.Context) (*IncidentRoleAssignment, error) {
	return withHooks(ctx, irauo.sqlSave, irauo.mutation, irauo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (irauo *IncidentRoleAssignmentUpdateOne) SaveX(ctx context.Context) *IncidentRoleAssignment {
	node, err := irauo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (irauo *IncidentRoleAssignmentUpdateOne) Exec(ctx context.Context) error {
	_, err := irauo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (irauo *IncidentRoleAssignmentUpdateOne) ExecX(ctx context.Context) {
	if err := irauo.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (irauo *IncidentRoleAssignmentUpdateOne) check() error {
	if irauo.mutation.RoleCleared() && len(irauo.mutation.RoleIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "IncidentRoleAssignment.role"`)
	}
	if irauo.mutation.IncidentCleared() && len(irauo.mutation.IncidentIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "IncidentRoleAssignment.incident"`)
	}
	if irauo.mutation.UserCleared() && len(irauo.mutation.UserIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "IncidentRoleAssignment.user"`)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (irauo *IncidentRoleAssignmentUpdateOne) Modify(modifiers ...func(u *sql.UpdateBuilder)) *IncidentRoleAssignmentUpdateOne {
	irauo.modifiers = append(irauo.modifiers, modifiers...)
	return irauo
}

func (irauo *IncidentRoleAssignmentUpdateOne) sqlSave(ctx context.Context) (_node *IncidentRoleAssignment, err error) {
	if err := irauo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(incidentroleassignment.Table, incidentroleassignment.Columns, sqlgraph.NewFieldSpec(incidentroleassignment.FieldID, field.TypeUUID))
	id, ok := irauo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "IncidentRoleAssignment.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := irauo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, incidentroleassignment.FieldID)
		for _, f := range fields {
			if !incidentroleassignment.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != incidentroleassignment.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := irauo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if irauo.mutation.RoleCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   incidentroleassignment.RoleTable,
			Columns: []string{incidentroleassignment.RoleColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentrole.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := irauo.mutation.RoleIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   incidentroleassignment.RoleTable,
			Columns: []string{incidentroleassignment.RoleColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentrole.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if irauo.mutation.IncidentCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   incidentroleassignment.IncidentTable,
			Columns: []string{incidentroleassignment.IncidentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incident.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := irauo.mutation.IncidentIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   incidentroleassignment.IncidentTable,
			Columns: []string{incidentroleassignment.IncidentColumn},
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
	if irauo.mutation.UserCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   incidentroleassignment.UserTable,
			Columns: []string{incidentroleassignment.UserColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(user.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := irauo.mutation.UserIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   incidentroleassignment.UserTable,
			Columns: []string{incidentroleassignment.UserColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(user.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(irauo.modifiers...)
	_node = &IncidentRoleAssignment{config: irauo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, irauo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{incidentroleassignment.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	irauo.mutation.done = true
	return _node, nil
}
