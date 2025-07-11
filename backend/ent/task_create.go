// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/task"
	"github.com/rezible/rezible/ent/ticket"
	"github.com/rezible/rezible/ent/user"
)

// TaskCreate is the builder for creating a Task entity.
type TaskCreate struct {
	config
	mutation *TaskMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetType sets the "type" field.
func (tc *TaskCreate) SetType(t task.Type) *TaskCreate {
	tc.mutation.SetType(t)
	return tc
}

// SetTitle sets the "title" field.
func (tc *TaskCreate) SetTitle(s string) *TaskCreate {
	tc.mutation.SetTitle(s)
	return tc
}

// SetIncidentID sets the "incident_id" field.
func (tc *TaskCreate) SetIncidentID(u uuid.UUID) *TaskCreate {
	tc.mutation.SetIncidentID(u)
	return tc
}

// SetNillableIncidentID sets the "incident_id" field if the given value is not nil.
func (tc *TaskCreate) SetNillableIncidentID(u *uuid.UUID) *TaskCreate {
	if u != nil {
		tc.SetIncidentID(*u)
	}
	return tc
}

// SetAssigneeID sets the "assignee_id" field.
func (tc *TaskCreate) SetAssigneeID(u uuid.UUID) *TaskCreate {
	tc.mutation.SetAssigneeID(u)
	return tc
}

// SetNillableAssigneeID sets the "assignee_id" field if the given value is not nil.
func (tc *TaskCreate) SetNillableAssigneeID(u *uuid.UUID) *TaskCreate {
	if u != nil {
		tc.SetAssigneeID(*u)
	}
	return tc
}

// SetCreatorID sets the "creator_id" field.
func (tc *TaskCreate) SetCreatorID(u uuid.UUID) *TaskCreate {
	tc.mutation.SetCreatorID(u)
	return tc
}

// SetNillableCreatorID sets the "creator_id" field if the given value is not nil.
func (tc *TaskCreate) SetNillableCreatorID(u *uuid.UUID) *TaskCreate {
	if u != nil {
		tc.SetCreatorID(*u)
	}
	return tc
}

// SetID sets the "id" field.
func (tc *TaskCreate) SetID(u uuid.UUID) *TaskCreate {
	tc.mutation.SetID(u)
	return tc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (tc *TaskCreate) SetNillableID(u *uuid.UUID) *TaskCreate {
	if u != nil {
		tc.SetID(*u)
	}
	return tc
}

// AddTicketIDs adds the "tickets" edge to the Ticket entity by IDs.
func (tc *TaskCreate) AddTicketIDs(ids ...uuid.UUID) *TaskCreate {
	tc.mutation.AddTicketIDs(ids...)
	return tc
}

// AddTickets adds the "tickets" edges to the Ticket entity.
func (tc *TaskCreate) AddTickets(t ...*Ticket) *TaskCreate {
	ids := make([]uuid.UUID, len(t))
	for i := range t {
		ids[i] = t[i].ID
	}
	return tc.AddTicketIDs(ids...)
}

// SetIncident sets the "incident" edge to the Incident entity.
func (tc *TaskCreate) SetIncident(i *Incident) *TaskCreate {
	return tc.SetIncidentID(i.ID)
}

// SetAssignee sets the "assignee" edge to the User entity.
func (tc *TaskCreate) SetAssignee(u *User) *TaskCreate {
	return tc.SetAssigneeID(u.ID)
}

// SetCreator sets the "creator" edge to the User entity.
func (tc *TaskCreate) SetCreator(u *User) *TaskCreate {
	return tc.SetCreatorID(u.ID)
}

// Mutation returns the TaskMutation object of the builder.
func (tc *TaskCreate) Mutation() *TaskMutation {
	return tc.mutation
}

// Save creates the Task in the database.
func (tc *TaskCreate) Save(ctx context.Context) (*Task, error) {
	tc.defaults()
	return withHooks(ctx, tc.sqlSave, tc.mutation, tc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (tc *TaskCreate) SaveX(ctx context.Context) *Task {
	v, err := tc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (tc *TaskCreate) Exec(ctx context.Context) error {
	_, err := tc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (tc *TaskCreate) ExecX(ctx context.Context) {
	if err := tc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (tc *TaskCreate) defaults() {
	if _, ok := tc.mutation.ID(); !ok {
		v := task.DefaultID()
		tc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (tc *TaskCreate) check() error {
	if _, ok := tc.mutation.GetType(); !ok {
		return &ValidationError{Name: "type", err: errors.New(`ent: missing required field "Task.type"`)}
	}
	if v, ok := tc.mutation.GetType(); ok {
		if err := task.TypeValidator(v); err != nil {
			return &ValidationError{Name: "type", err: fmt.Errorf(`ent: validator failed for field "Task.type": %w`, err)}
		}
	}
	if _, ok := tc.mutation.Title(); !ok {
		return &ValidationError{Name: "title", err: errors.New(`ent: missing required field "Task.title"`)}
	}
	return nil
}

func (tc *TaskCreate) sqlSave(ctx context.Context) (*Task, error) {
	if err := tc.check(); err != nil {
		return nil, err
	}
	_node, _spec := tc.createSpec()
	if err := sqlgraph.CreateNode(ctx, tc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != nil {
		if id, ok := _spec.ID.Value.(*uuid.UUID); ok {
			_node.ID = *id
		} else if err := _node.ID.Scan(_spec.ID.Value); err != nil {
			return nil, err
		}
	}
	tc.mutation.id = &_node.ID
	tc.mutation.done = true
	return _node, nil
}

func (tc *TaskCreate) createSpec() (*Task, *sqlgraph.CreateSpec) {
	var (
		_node = &Task{config: tc.config}
		_spec = sqlgraph.NewCreateSpec(task.Table, sqlgraph.NewFieldSpec(task.FieldID, field.TypeUUID))
	)
	_spec.OnConflict = tc.conflict
	if id, ok := tc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := tc.mutation.GetType(); ok {
		_spec.SetField(task.FieldType, field.TypeEnum, value)
		_node.Type = value
	}
	if value, ok := tc.mutation.Title(); ok {
		_spec.SetField(task.FieldTitle, field.TypeString, value)
		_node.Title = value
	}
	if nodes := tc.mutation.TicketsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   task.TicketsTable,
			Columns: task.TicketsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(ticket.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := tc.mutation.IncidentIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   task.IncidentTable,
			Columns: []string{task.IncidentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incident.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.IncidentID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := tc.mutation.AssigneeIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   task.AssigneeTable,
			Columns: []string{task.AssigneeColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(user.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.AssigneeID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := tc.mutation.CreatorIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   task.CreatorTable,
			Columns: []string{task.CreatorColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(user.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.CreatorID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.Task.Create().
//		SetType(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.TaskUpsert) {
//			SetType(v+v).
//		}).
//		Exec(ctx)
func (tc *TaskCreate) OnConflict(opts ...sql.ConflictOption) *TaskUpsertOne {
	tc.conflict = opts
	return &TaskUpsertOne{
		create: tc,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.Task.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (tc *TaskCreate) OnConflictColumns(columns ...string) *TaskUpsertOne {
	tc.conflict = append(tc.conflict, sql.ConflictColumns(columns...))
	return &TaskUpsertOne{
		create: tc,
	}
}

type (
	// TaskUpsertOne is the builder for "upsert"-ing
	//  one Task node.
	TaskUpsertOne struct {
		create *TaskCreate
	}

	// TaskUpsert is the "OnConflict" setter.
	TaskUpsert struct {
		*sql.UpdateSet
	}
)

// SetType sets the "type" field.
func (u *TaskUpsert) SetType(v task.Type) *TaskUpsert {
	u.Set(task.FieldType, v)
	return u
}

// UpdateType sets the "type" field to the value that was provided on create.
func (u *TaskUpsert) UpdateType() *TaskUpsert {
	u.SetExcluded(task.FieldType)
	return u
}

// SetTitle sets the "title" field.
func (u *TaskUpsert) SetTitle(v string) *TaskUpsert {
	u.Set(task.FieldTitle, v)
	return u
}

// UpdateTitle sets the "title" field to the value that was provided on create.
func (u *TaskUpsert) UpdateTitle() *TaskUpsert {
	u.SetExcluded(task.FieldTitle)
	return u
}

// SetIncidentID sets the "incident_id" field.
func (u *TaskUpsert) SetIncidentID(v uuid.UUID) *TaskUpsert {
	u.Set(task.FieldIncidentID, v)
	return u
}

// UpdateIncidentID sets the "incident_id" field to the value that was provided on create.
func (u *TaskUpsert) UpdateIncidentID() *TaskUpsert {
	u.SetExcluded(task.FieldIncidentID)
	return u
}

// ClearIncidentID clears the value of the "incident_id" field.
func (u *TaskUpsert) ClearIncidentID() *TaskUpsert {
	u.SetNull(task.FieldIncidentID)
	return u
}

// SetAssigneeID sets the "assignee_id" field.
func (u *TaskUpsert) SetAssigneeID(v uuid.UUID) *TaskUpsert {
	u.Set(task.FieldAssigneeID, v)
	return u
}

// UpdateAssigneeID sets the "assignee_id" field to the value that was provided on create.
func (u *TaskUpsert) UpdateAssigneeID() *TaskUpsert {
	u.SetExcluded(task.FieldAssigneeID)
	return u
}

// ClearAssigneeID clears the value of the "assignee_id" field.
func (u *TaskUpsert) ClearAssigneeID() *TaskUpsert {
	u.SetNull(task.FieldAssigneeID)
	return u
}

// SetCreatorID sets the "creator_id" field.
func (u *TaskUpsert) SetCreatorID(v uuid.UUID) *TaskUpsert {
	u.Set(task.FieldCreatorID, v)
	return u
}

// UpdateCreatorID sets the "creator_id" field to the value that was provided on create.
func (u *TaskUpsert) UpdateCreatorID() *TaskUpsert {
	u.SetExcluded(task.FieldCreatorID)
	return u
}

// ClearCreatorID clears the value of the "creator_id" field.
func (u *TaskUpsert) ClearCreatorID() *TaskUpsert {
	u.SetNull(task.FieldCreatorID)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create except the ID field.
// Using this option is equivalent to using:
//
//	client.Task.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(task.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *TaskUpsertOne) UpdateNewValues() *TaskUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.ID(); exists {
			s.SetIgnore(task.FieldID)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.Task.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *TaskUpsertOne) Ignore() *TaskUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *TaskUpsertOne) DoNothing() *TaskUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the TaskCreate.OnConflict
// documentation for more info.
func (u *TaskUpsertOne) Update(set func(*TaskUpsert)) *TaskUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&TaskUpsert{UpdateSet: update})
	}))
	return u
}

// SetType sets the "type" field.
func (u *TaskUpsertOne) SetType(v task.Type) *TaskUpsertOne {
	return u.Update(func(s *TaskUpsert) {
		s.SetType(v)
	})
}

// UpdateType sets the "type" field to the value that was provided on create.
func (u *TaskUpsertOne) UpdateType() *TaskUpsertOne {
	return u.Update(func(s *TaskUpsert) {
		s.UpdateType()
	})
}

// SetTitle sets the "title" field.
func (u *TaskUpsertOne) SetTitle(v string) *TaskUpsertOne {
	return u.Update(func(s *TaskUpsert) {
		s.SetTitle(v)
	})
}

// UpdateTitle sets the "title" field to the value that was provided on create.
func (u *TaskUpsertOne) UpdateTitle() *TaskUpsertOne {
	return u.Update(func(s *TaskUpsert) {
		s.UpdateTitle()
	})
}

// SetIncidentID sets the "incident_id" field.
func (u *TaskUpsertOne) SetIncidentID(v uuid.UUID) *TaskUpsertOne {
	return u.Update(func(s *TaskUpsert) {
		s.SetIncidentID(v)
	})
}

// UpdateIncidentID sets the "incident_id" field to the value that was provided on create.
func (u *TaskUpsertOne) UpdateIncidentID() *TaskUpsertOne {
	return u.Update(func(s *TaskUpsert) {
		s.UpdateIncidentID()
	})
}

// ClearIncidentID clears the value of the "incident_id" field.
func (u *TaskUpsertOne) ClearIncidentID() *TaskUpsertOne {
	return u.Update(func(s *TaskUpsert) {
		s.ClearIncidentID()
	})
}

// SetAssigneeID sets the "assignee_id" field.
func (u *TaskUpsertOne) SetAssigneeID(v uuid.UUID) *TaskUpsertOne {
	return u.Update(func(s *TaskUpsert) {
		s.SetAssigneeID(v)
	})
}

// UpdateAssigneeID sets the "assignee_id" field to the value that was provided on create.
func (u *TaskUpsertOne) UpdateAssigneeID() *TaskUpsertOne {
	return u.Update(func(s *TaskUpsert) {
		s.UpdateAssigneeID()
	})
}

// ClearAssigneeID clears the value of the "assignee_id" field.
func (u *TaskUpsertOne) ClearAssigneeID() *TaskUpsertOne {
	return u.Update(func(s *TaskUpsert) {
		s.ClearAssigneeID()
	})
}

// SetCreatorID sets the "creator_id" field.
func (u *TaskUpsertOne) SetCreatorID(v uuid.UUID) *TaskUpsertOne {
	return u.Update(func(s *TaskUpsert) {
		s.SetCreatorID(v)
	})
}

// UpdateCreatorID sets the "creator_id" field to the value that was provided on create.
func (u *TaskUpsertOne) UpdateCreatorID() *TaskUpsertOne {
	return u.Update(func(s *TaskUpsert) {
		s.UpdateCreatorID()
	})
}

// ClearCreatorID clears the value of the "creator_id" field.
func (u *TaskUpsertOne) ClearCreatorID() *TaskUpsertOne {
	return u.Update(func(s *TaskUpsert) {
		s.ClearCreatorID()
	})
}

// Exec executes the query.
func (u *TaskUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for TaskCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *TaskUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *TaskUpsertOne) ID(ctx context.Context) (id uuid.UUID, err error) {
	if u.create.driver.Dialect() == dialect.MySQL {
		// In case of "ON CONFLICT", there is no way to get back non-numeric ID
		// fields from the database since MySQL does not support the RETURNING clause.
		return id, errors.New("ent: TaskUpsertOne.ID is not supported by MySQL driver. Use TaskUpsertOne.Exec instead")
	}
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *TaskUpsertOne) IDX(ctx context.Context) uuid.UUID {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// TaskCreateBulk is the builder for creating many Task entities in bulk.
type TaskCreateBulk struct {
	config
	err      error
	builders []*TaskCreate
	conflict []sql.ConflictOption
}

// Save creates the Task entities in the database.
func (tcb *TaskCreateBulk) Save(ctx context.Context) ([]*Task, error) {
	if tcb.err != nil {
		return nil, tcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(tcb.builders))
	nodes := make([]*Task, len(tcb.builders))
	mutators := make([]Mutator, len(tcb.builders))
	for i := range tcb.builders {
		func(i int, root context.Context) {
			builder := tcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*TaskMutation)
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}
				if err := builder.check(); err != nil {
					return nil, err
				}
				builder.mutation = mutation
				var err error
				nodes[i], specs[i] = builder.createSpec()
				if i < len(mutators)-1 {
					_, err = mutators[i+1].Mutate(root, tcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = tcb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, tcb.driver, spec); err != nil {
						if sqlgraph.IsConstraintError(err) {
							err = &ConstraintError{msg: err.Error(), wrap: err}
						}
					}
				}
				if err != nil {
					return nil, err
				}
				mutation.id = &nodes[i].ID
				mutation.done = true
				return nodes[i], nil
			})
			for i := len(builder.hooks) - 1; i >= 0; i-- {
				mut = builder.hooks[i](mut)
			}
			mutators[i] = mut
		}(i, ctx)
	}
	if len(mutators) > 0 {
		if _, err := mutators[0].Mutate(ctx, tcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (tcb *TaskCreateBulk) SaveX(ctx context.Context) []*Task {
	v, err := tcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (tcb *TaskCreateBulk) Exec(ctx context.Context) error {
	_, err := tcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (tcb *TaskCreateBulk) ExecX(ctx context.Context) {
	if err := tcb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.Task.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.TaskUpsert) {
//			SetType(v+v).
//		}).
//		Exec(ctx)
func (tcb *TaskCreateBulk) OnConflict(opts ...sql.ConflictOption) *TaskUpsertBulk {
	tcb.conflict = opts
	return &TaskUpsertBulk{
		create: tcb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.Task.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (tcb *TaskCreateBulk) OnConflictColumns(columns ...string) *TaskUpsertBulk {
	tcb.conflict = append(tcb.conflict, sql.ConflictColumns(columns...))
	return &TaskUpsertBulk{
		create: tcb,
	}
}

// TaskUpsertBulk is the builder for "upsert"-ing
// a bulk of Task nodes.
type TaskUpsertBulk struct {
	create *TaskCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.Task.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(task.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *TaskUpsertBulk) UpdateNewValues() *TaskUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.ID(); exists {
				s.SetIgnore(task.FieldID)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.Task.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *TaskUpsertBulk) Ignore() *TaskUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *TaskUpsertBulk) DoNothing() *TaskUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the TaskCreateBulk.OnConflict
// documentation for more info.
func (u *TaskUpsertBulk) Update(set func(*TaskUpsert)) *TaskUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&TaskUpsert{UpdateSet: update})
	}))
	return u
}

// SetType sets the "type" field.
func (u *TaskUpsertBulk) SetType(v task.Type) *TaskUpsertBulk {
	return u.Update(func(s *TaskUpsert) {
		s.SetType(v)
	})
}

// UpdateType sets the "type" field to the value that was provided on create.
func (u *TaskUpsertBulk) UpdateType() *TaskUpsertBulk {
	return u.Update(func(s *TaskUpsert) {
		s.UpdateType()
	})
}

// SetTitle sets the "title" field.
func (u *TaskUpsertBulk) SetTitle(v string) *TaskUpsertBulk {
	return u.Update(func(s *TaskUpsert) {
		s.SetTitle(v)
	})
}

// UpdateTitle sets the "title" field to the value that was provided on create.
func (u *TaskUpsertBulk) UpdateTitle() *TaskUpsertBulk {
	return u.Update(func(s *TaskUpsert) {
		s.UpdateTitle()
	})
}

// SetIncidentID sets the "incident_id" field.
func (u *TaskUpsertBulk) SetIncidentID(v uuid.UUID) *TaskUpsertBulk {
	return u.Update(func(s *TaskUpsert) {
		s.SetIncidentID(v)
	})
}

// UpdateIncidentID sets the "incident_id" field to the value that was provided on create.
func (u *TaskUpsertBulk) UpdateIncidentID() *TaskUpsertBulk {
	return u.Update(func(s *TaskUpsert) {
		s.UpdateIncidentID()
	})
}

// ClearIncidentID clears the value of the "incident_id" field.
func (u *TaskUpsertBulk) ClearIncidentID() *TaskUpsertBulk {
	return u.Update(func(s *TaskUpsert) {
		s.ClearIncidentID()
	})
}

// SetAssigneeID sets the "assignee_id" field.
func (u *TaskUpsertBulk) SetAssigneeID(v uuid.UUID) *TaskUpsertBulk {
	return u.Update(func(s *TaskUpsert) {
		s.SetAssigneeID(v)
	})
}

// UpdateAssigneeID sets the "assignee_id" field to the value that was provided on create.
func (u *TaskUpsertBulk) UpdateAssigneeID() *TaskUpsertBulk {
	return u.Update(func(s *TaskUpsert) {
		s.UpdateAssigneeID()
	})
}

// ClearAssigneeID clears the value of the "assignee_id" field.
func (u *TaskUpsertBulk) ClearAssigneeID() *TaskUpsertBulk {
	return u.Update(func(s *TaskUpsert) {
		s.ClearAssigneeID()
	})
}

// SetCreatorID sets the "creator_id" field.
func (u *TaskUpsertBulk) SetCreatorID(v uuid.UUID) *TaskUpsertBulk {
	return u.Update(func(s *TaskUpsert) {
		s.SetCreatorID(v)
	})
}

// UpdateCreatorID sets the "creator_id" field to the value that was provided on create.
func (u *TaskUpsertBulk) UpdateCreatorID() *TaskUpsertBulk {
	return u.Update(func(s *TaskUpsert) {
		s.UpdateCreatorID()
	})
}

// ClearCreatorID clears the value of the "creator_id" field.
func (u *TaskUpsertBulk) ClearCreatorID() *TaskUpsertBulk {
	return u.Update(func(s *TaskUpsert) {
		s.ClearCreatorID()
	})
}

// Exec executes the query.
func (u *TaskUpsertBulk) Exec(ctx context.Context) error {
	if u.create.err != nil {
		return u.create.err
	}
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the TaskCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for TaskCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *TaskUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}
