// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/twohundreds/rezible/ent/incidentdebriefquestion"
	"github.com/twohundreds/rezible/ent/incidentrole"
	"github.com/twohundreds/rezible/ent/incidentroleassignment"
)

// IncidentRoleCreate is the builder for creating a IncidentRole entity.
type IncidentRoleCreate struct {
	config
	mutation *IncidentRoleMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetArchiveTime sets the "archive_time" field.
func (irc *IncidentRoleCreate) SetArchiveTime(t time.Time) *IncidentRoleCreate {
	irc.mutation.SetArchiveTime(t)
	return irc
}

// SetNillableArchiveTime sets the "archive_time" field if the given value is not nil.
func (irc *IncidentRoleCreate) SetNillableArchiveTime(t *time.Time) *IncidentRoleCreate {
	if t != nil {
		irc.SetArchiveTime(*t)
	}
	return irc
}

// SetName sets the "name" field.
func (irc *IncidentRoleCreate) SetName(s string) *IncidentRoleCreate {
	irc.mutation.SetName(s)
	return irc
}

// SetProviderID sets the "provider_id" field.
func (irc *IncidentRoleCreate) SetProviderID(s string) *IncidentRoleCreate {
	irc.mutation.SetProviderID(s)
	return irc
}

// SetRequired sets the "required" field.
func (irc *IncidentRoleCreate) SetRequired(b bool) *IncidentRoleCreate {
	irc.mutation.SetRequired(b)
	return irc
}

// SetNillableRequired sets the "required" field if the given value is not nil.
func (irc *IncidentRoleCreate) SetNillableRequired(b *bool) *IncidentRoleCreate {
	if b != nil {
		irc.SetRequired(*b)
	}
	return irc
}

// SetID sets the "id" field.
func (irc *IncidentRoleCreate) SetID(u uuid.UUID) *IncidentRoleCreate {
	irc.mutation.SetID(u)
	return irc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (irc *IncidentRoleCreate) SetNillableID(u *uuid.UUID) *IncidentRoleCreate {
	if u != nil {
		irc.SetID(*u)
	}
	return irc
}

// AddAssignmentIDs adds the "assignments" edge to the IncidentRoleAssignment entity by IDs.
func (irc *IncidentRoleCreate) AddAssignmentIDs(ids ...uuid.UUID) *IncidentRoleCreate {
	irc.mutation.AddAssignmentIDs(ids...)
	return irc
}

// AddAssignments adds the "assignments" edges to the IncidentRoleAssignment entity.
func (irc *IncidentRoleCreate) AddAssignments(i ...*IncidentRoleAssignment) *IncidentRoleCreate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return irc.AddAssignmentIDs(ids...)
}

// AddDebriefQuestionIDs adds the "debrief_questions" edge to the IncidentDebriefQuestion entity by IDs.
func (irc *IncidentRoleCreate) AddDebriefQuestionIDs(ids ...uuid.UUID) *IncidentRoleCreate {
	irc.mutation.AddDebriefQuestionIDs(ids...)
	return irc
}

// AddDebriefQuestions adds the "debrief_questions" edges to the IncidentDebriefQuestion entity.
func (irc *IncidentRoleCreate) AddDebriefQuestions(i ...*IncidentDebriefQuestion) *IncidentRoleCreate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return irc.AddDebriefQuestionIDs(ids...)
}

// Mutation returns the IncidentRoleMutation object of the builder.
func (irc *IncidentRoleCreate) Mutation() *IncidentRoleMutation {
	return irc.mutation
}

// Save creates the IncidentRole in the database.
func (irc *IncidentRoleCreate) Save(ctx context.Context) (*IncidentRole, error) {
	if err := irc.defaults(); err != nil {
		return nil, err
	}
	return withHooks(ctx, irc.sqlSave, irc.mutation, irc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (irc *IncidentRoleCreate) SaveX(ctx context.Context) *IncidentRole {
	v, err := irc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (irc *IncidentRoleCreate) Exec(ctx context.Context) error {
	_, err := irc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (irc *IncidentRoleCreate) ExecX(ctx context.Context) {
	if err := irc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (irc *IncidentRoleCreate) defaults() error {
	if _, ok := irc.mutation.Required(); !ok {
		v := incidentrole.DefaultRequired
		irc.mutation.SetRequired(v)
	}
	if _, ok := irc.mutation.ID(); !ok {
		if incidentrole.DefaultID == nil {
			return fmt.Errorf("ent: uninitialized incidentrole.DefaultID (forgotten import ent/runtime?)")
		}
		v := incidentrole.DefaultID()
		irc.mutation.SetID(v)
	}
	return nil
}

// check runs all checks and user-defined validators on the builder.
func (irc *IncidentRoleCreate) check() error {
	if _, ok := irc.mutation.Name(); !ok {
		return &ValidationError{Name: "name", err: errors.New(`ent: missing required field "IncidentRole.name"`)}
	}
	if _, ok := irc.mutation.ProviderID(); !ok {
		return &ValidationError{Name: "provider_id", err: errors.New(`ent: missing required field "IncidentRole.provider_id"`)}
	}
	if _, ok := irc.mutation.Required(); !ok {
		return &ValidationError{Name: "required", err: errors.New(`ent: missing required field "IncidentRole.required"`)}
	}
	return nil
}

func (irc *IncidentRoleCreate) sqlSave(ctx context.Context) (*IncidentRole, error) {
	if err := irc.check(); err != nil {
		return nil, err
	}
	_node, _spec := irc.createSpec()
	if err := sqlgraph.CreateNode(ctx, irc.driver, _spec); err != nil {
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
	irc.mutation.id = &_node.ID
	irc.mutation.done = true
	return _node, nil
}

func (irc *IncidentRoleCreate) createSpec() (*IncidentRole, *sqlgraph.CreateSpec) {
	var (
		_node = &IncidentRole{config: irc.config}
		_spec = sqlgraph.NewCreateSpec(incidentrole.Table, sqlgraph.NewFieldSpec(incidentrole.FieldID, field.TypeUUID))
	)
	_spec.OnConflict = irc.conflict
	if id, ok := irc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := irc.mutation.ArchiveTime(); ok {
		_spec.SetField(incidentrole.FieldArchiveTime, field.TypeTime, value)
		_node.ArchiveTime = value
	}
	if value, ok := irc.mutation.Name(); ok {
		_spec.SetField(incidentrole.FieldName, field.TypeString, value)
		_node.Name = value
	}
	if value, ok := irc.mutation.ProviderID(); ok {
		_spec.SetField(incidentrole.FieldProviderID, field.TypeString, value)
		_node.ProviderID = value
	}
	if value, ok := irc.mutation.Required(); ok {
		_spec.SetField(incidentrole.FieldRequired, field.TypeBool, value)
		_node.Required = value
	}
	if nodes := irc.mutation.AssignmentsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   incidentrole.AssignmentsTable,
			Columns: []string{incidentrole.AssignmentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentroleassignment.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := irc.mutation.DebriefQuestionsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   incidentrole.DebriefQuestionsTable,
			Columns: incidentrole.DebriefQuestionsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentdebriefquestion.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.IncidentRole.Create().
//		SetArchiveTime(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.IncidentRoleUpsert) {
//			SetArchiveTime(v+v).
//		}).
//		Exec(ctx)
func (irc *IncidentRoleCreate) OnConflict(opts ...sql.ConflictOption) *IncidentRoleUpsertOne {
	irc.conflict = opts
	return &IncidentRoleUpsertOne{
		create: irc,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.IncidentRole.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (irc *IncidentRoleCreate) OnConflictColumns(columns ...string) *IncidentRoleUpsertOne {
	irc.conflict = append(irc.conflict, sql.ConflictColumns(columns...))
	return &IncidentRoleUpsertOne{
		create: irc,
	}
}

type (
	// IncidentRoleUpsertOne is the builder for "upsert"-ing
	//  one IncidentRole node.
	IncidentRoleUpsertOne struct {
		create *IncidentRoleCreate
	}

	// IncidentRoleUpsert is the "OnConflict" setter.
	IncidentRoleUpsert struct {
		*sql.UpdateSet
	}
)

// SetArchiveTime sets the "archive_time" field.
func (u *IncidentRoleUpsert) SetArchiveTime(v time.Time) *IncidentRoleUpsert {
	u.Set(incidentrole.FieldArchiveTime, v)
	return u
}

// UpdateArchiveTime sets the "archive_time" field to the value that was provided on create.
func (u *IncidentRoleUpsert) UpdateArchiveTime() *IncidentRoleUpsert {
	u.SetExcluded(incidentrole.FieldArchiveTime)
	return u
}

// ClearArchiveTime clears the value of the "archive_time" field.
func (u *IncidentRoleUpsert) ClearArchiveTime() *IncidentRoleUpsert {
	u.SetNull(incidentrole.FieldArchiveTime)
	return u
}

// SetName sets the "name" field.
func (u *IncidentRoleUpsert) SetName(v string) *IncidentRoleUpsert {
	u.Set(incidentrole.FieldName, v)
	return u
}

// UpdateName sets the "name" field to the value that was provided on create.
func (u *IncidentRoleUpsert) UpdateName() *IncidentRoleUpsert {
	u.SetExcluded(incidentrole.FieldName)
	return u
}

// SetProviderID sets the "provider_id" field.
func (u *IncidentRoleUpsert) SetProviderID(v string) *IncidentRoleUpsert {
	u.Set(incidentrole.FieldProviderID, v)
	return u
}

// UpdateProviderID sets the "provider_id" field to the value that was provided on create.
func (u *IncidentRoleUpsert) UpdateProviderID() *IncidentRoleUpsert {
	u.SetExcluded(incidentrole.FieldProviderID)
	return u
}

// SetRequired sets the "required" field.
func (u *IncidentRoleUpsert) SetRequired(v bool) *IncidentRoleUpsert {
	u.Set(incidentrole.FieldRequired, v)
	return u
}

// UpdateRequired sets the "required" field to the value that was provided on create.
func (u *IncidentRoleUpsert) UpdateRequired() *IncidentRoleUpsert {
	u.SetExcluded(incidentrole.FieldRequired)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create except the ID field.
// Using this option is equivalent to using:
//
//	client.IncidentRole.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(incidentrole.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *IncidentRoleUpsertOne) UpdateNewValues() *IncidentRoleUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.ID(); exists {
			s.SetIgnore(incidentrole.FieldID)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.IncidentRole.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *IncidentRoleUpsertOne) Ignore() *IncidentRoleUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *IncidentRoleUpsertOne) DoNothing() *IncidentRoleUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the IncidentRoleCreate.OnConflict
// documentation for more info.
func (u *IncidentRoleUpsertOne) Update(set func(*IncidentRoleUpsert)) *IncidentRoleUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&IncidentRoleUpsert{UpdateSet: update})
	}))
	return u
}

// SetArchiveTime sets the "archive_time" field.
func (u *IncidentRoleUpsertOne) SetArchiveTime(v time.Time) *IncidentRoleUpsertOne {
	return u.Update(func(s *IncidentRoleUpsert) {
		s.SetArchiveTime(v)
	})
}

// UpdateArchiveTime sets the "archive_time" field to the value that was provided on create.
func (u *IncidentRoleUpsertOne) UpdateArchiveTime() *IncidentRoleUpsertOne {
	return u.Update(func(s *IncidentRoleUpsert) {
		s.UpdateArchiveTime()
	})
}

// ClearArchiveTime clears the value of the "archive_time" field.
func (u *IncidentRoleUpsertOne) ClearArchiveTime() *IncidentRoleUpsertOne {
	return u.Update(func(s *IncidentRoleUpsert) {
		s.ClearArchiveTime()
	})
}

// SetName sets the "name" field.
func (u *IncidentRoleUpsertOne) SetName(v string) *IncidentRoleUpsertOne {
	return u.Update(func(s *IncidentRoleUpsert) {
		s.SetName(v)
	})
}

// UpdateName sets the "name" field to the value that was provided on create.
func (u *IncidentRoleUpsertOne) UpdateName() *IncidentRoleUpsertOne {
	return u.Update(func(s *IncidentRoleUpsert) {
		s.UpdateName()
	})
}

// SetProviderID sets the "provider_id" field.
func (u *IncidentRoleUpsertOne) SetProviderID(v string) *IncidentRoleUpsertOne {
	return u.Update(func(s *IncidentRoleUpsert) {
		s.SetProviderID(v)
	})
}

// UpdateProviderID sets the "provider_id" field to the value that was provided on create.
func (u *IncidentRoleUpsertOne) UpdateProviderID() *IncidentRoleUpsertOne {
	return u.Update(func(s *IncidentRoleUpsert) {
		s.UpdateProviderID()
	})
}

// SetRequired sets the "required" field.
func (u *IncidentRoleUpsertOne) SetRequired(v bool) *IncidentRoleUpsertOne {
	return u.Update(func(s *IncidentRoleUpsert) {
		s.SetRequired(v)
	})
}

// UpdateRequired sets the "required" field to the value that was provided on create.
func (u *IncidentRoleUpsertOne) UpdateRequired() *IncidentRoleUpsertOne {
	return u.Update(func(s *IncidentRoleUpsert) {
		s.UpdateRequired()
	})
}

// Exec executes the query.
func (u *IncidentRoleUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for IncidentRoleCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *IncidentRoleUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *IncidentRoleUpsertOne) ID(ctx context.Context) (id uuid.UUID, err error) {
	if u.create.driver.Dialect() == dialect.MySQL {
		// In case of "ON CONFLICT", there is no way to get back non-numeric ID
		// fields from the database since MySQL does not support the RETURNING clause.
		return id, errors.New("ent: IncidentRoleUpsertOne.ID is not supported by MySQL driver. Use IncidentRoleUpsertOne.Exec instead")
	}
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *IncidentRoleUpsertOne) IDX(ctx context.Context) uuid.UUID {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// IncidentRoleCreateBulk is the builder for creating many IncidentRole entities in bulk.
type IncidentRoleCreateBulk struct {
	config
	err      error
	builders []*IncidentRoleCreate
	conflict []sql.ConflictOption
}

// Save creates the IncidentRole entities in the database.
func (ircb *IncidentRoleCreateBulk) Save(ctx context.Context) ([]*IncidentRole, error) {
	if ircb.err != nil {
		return nil, ircb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(ircb.builders))
	nodes := make([]*IncidentRole, len(ircb.builders))
	mutators := make([]Mutator, len(ircb.builders))
	for i := range ircb.builders {
		func(i int, root context.Context) {
			builder := ircb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*IncidentRoleMutation)
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
					_, err = mutators[i+1].Mutate(root, ircb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = ircb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, ircb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, ircb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (ircb *IncidentRoleCreateBulk) SaveX(ctx context.Context) []*IncidentRole {
	v, err := ircb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (ircb *IncidentRoleCreateBulk) Exec(ctx context.Context) error {
	_, err := ircb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ircb *IncidentRoleCreateBulk) ExecX(ctx context.Context) {
	if err := ircb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.IncidentRole.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.IncidentRoleUpsert) {
//			SetArchiveTime(v+v).
//		}).
//		Exec(ctx)
func (ircb *IncidentRoleCreateBulk) OnConflict(opts ...sql.ConflictOption) *IncidentRoleUpsertBulk {
	ircb.conflict = opts
	return &IncidentRoleUpsertBulk{
		create: ircb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.IncidentRole.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (ircb *IncidentRoleCreateBulk) OnConflictColumns(columns ...string) *IncidentRoleUpsertBulk {
	ircb.conflict = append(ircb.conflict, sql.ConflictColumns(columns...))
	return &IncidentRoleUpsertBulk{
		create: ircb,
	}
}

// IncidentRoleUpsertBulk is the builder for "upsert"-ing
// a bulk of IncidentRole nodes.
type IncidentRoleUpsertBulk struct {
	create *IncidentRoleCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.IncidentRole.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(incidentrole.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *IncidentRoleUpsertBulk) UpdateNewValues() *IncidentRoleUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.ID(); exists {
				s.SetIgnore(incidentrole.FieldID)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.IncidentRole.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *IncidentRoleUpsertBulk) Ignore() *IncidentRoleUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *IncidentRoleUpsertBulk) DoNothing() *IncidentRoleUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the IncidentRoleCreateBulk.OnConflict
// documentation for more info.
func (u *IncidentRoleUpsertBulk) Update(set func(*IncidentRoleUpsert)) *IncidentRoleUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&IncidentRoleUpsert{UpdateSet: update})
	}))
	return u
}

// SetArchiveTime sets the "archive_time" field.
func (u *IncidentRoleUpsertBulk) SetArchiveTime(v time.Time) *IncidentRoleUpsertBulk {
	return u.Update(func(s *IncidentRoleUpsert) {
		s.SetArchiveTime(v)
	})
}

// UpdateArchiveTime sets the "archive_time" field to the value that was provided on create.
func (u *IncidentRoleUpsertBulk) UpdateArchiveTime() *IncidentRoleUpsertBulk {
	return u.Update(func(s *IncidentRoleUpsert) {
		s.UpdateArchiveTime()
	})
}

// ClearArchiveTime clears the value of the "archive_time" field.
func (u *IncidentRoleUpsertBulk) ClearArchiveTime() *IncidentRoleUpsertBulk {
	return u.Update(func(s *IncidentRoleUpsert) {
		s.ClearArchiveTime()
	})
}

// SetName sets the "name" field.
func (u *IncidentRoleUpsertBulk) SetName(v string) *IncidentRoleUpsertBulk {
	return u.Update(func(s *IncidentRoleUpsert) {
		s.SetName(v)
	})
}

// UpdateName sets the "name" field to the value that was provided on create.
func (u *IncidentRoleUpsertBulk) UpdateName() *IncidentRoleUpsertBulk {
	return u.Update(func(s *IncidentRoleUpsert) {
		s.UpdateName()
	})
}

// SetProviderID sets the "provider_id" field.
func (u *IncidentRoleUpsertBulk) SetProviderID(v string) *IncidentRoleUpsertBulk {
	return u.Update(func(s *IncidentRoleUpsert) {
		s.SetProviderID(v)
	})
}

// UpdateProviderID sets the "provider_id" field to the value that was provided on create.
func (u *IncidentRoleUpsertBulk) UpdateProviderID() *IncidentRoleUpsertBulk {
	return u.Update(func(s *IncidentRoleUpsert) {
		s.UpdateProviderID()
	})
}

// SetRequired sets the "required" field.
func (u *IncidentRoleUpsertBulk) SetRequired(v bool) *IncidentRoleUpsertBulk {
	return u.Update(func(s *IncidentRoleUpsert) {
		s.SetRequired(v)
	})
}

// UpdateRequired sets the "required" field to the value that was provided on create.
func (u *IncidentRoleUpsertBulk) UpdateRequired() *IncidentRoleUpsertBulk {
	return u.Update(func(s *IncidentRoleUpsert) {
		s.UpdateRequired()
	})
}

// Exec executes the query.
func (u *IncidentRoleUpsertBulk) Exec(ctx context.Context) error {
	if u.create.err != nil {
		return u.create.err
	}
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the IncidentRoleCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for IncidentRoleCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *IncidentRoleUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}
