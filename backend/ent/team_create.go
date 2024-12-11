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
	"github.com/twohundreds/rezible/ent/incidentteamassignment"
	"github.com/twohundreds/rezible/ent/meetingschedule"
	"github.com/twohundreds/rezible/ent/oncallroster"
	"github.com/twohundreds/rezible/ent/service"
	"github.com/twohundreds/rezible/ent/subscription"
	"github.com/twohundreds/rezible/ent/team"
	"github.com/twohundreds/rezible/ent/user"
)

// TeamCreate is the builder for creating a Team entity.
type TeamCreate struct {
	config
	mutation *TeamMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetSlug sets the "slug" field.
func (tc *TeamCreate) SetSlug(s string) *TeamCreate {
	tc.mutation.SetSlug(s)
	return tc
}

// SetName sets the "name" field.
func (tc *TeamCreate) SetName(s string) *TeamCreate {
	tc.mutation.SetName(s)
	return tc
}

// SetChatChannelID sets the "chat_channel_id" field.
func (tc *TeamCreate) SetChatChannelID(s string) *TeamCreate {
	tc.mutation.SetChatChannelID(s)
	return tc
}

// SetNillableChatChannelID sets the "chat_channel_id" field if the given value is not nil.
func (tc *TeamCreate) SetNillableChatChannelID(s *string) *TeamCreate {
	if s != nil {
		tc.SetChatChannelID(*s)
	}
	return tc
}

// SetTimezone sets the "timezone" field.
func (tc *TeamCreate) SetTimezone(s string) *TeamCreate {
	tc.mutation.SetTimezone(s)
	return tc
}

// SetNillableTimezone sets the "timezone" field if the given value is not nil.
func (tc *TeamCreate) SetNillableTimezone(s *string) *TeamCreate {
	if s != nil {
		tc.SetTimezone(*s)
	}
	return tc
}

// SetID sets the "id" field.
func (tc *TeamCreate) SetID(u uuid.UUID) *TeamCreate {
	tc.mutation.SetID(u)
	return tc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (tc *TeamCreate) SetNillableID(u *uuid.UUID) *TeamCreate {
	if u != nil {
		tc.SetID(*u)
	}
	return tc
}

// AddUserIDs adds the "users" edge to the User entity by IDs.
func (tc *TeamCreate) AddUserIDs(ids ...uuid.UUID) *TeamCreate {
	tc.mutation.AddUserIDs(ids...)
	return tc
}

// AddUsers adds the "users" edges to the User entity.
func (tc *TeamCreate) AddUsers(u ...*User) *TeamCreate {
	ids := make([]uuid.UUID, len(u))
	for i := range u {
		ids[i] = u[i].ID
	}
	return tc.AddUserIDs(ids...)
}

// AddServiceIDs adds the "services" edge to the Service entity by IDs.
func (tc *TeamCreate) AddServiceIDs(ids ...uuid.UUID) *TeamCreate {
	tc.mutation.AddServiceIDs(ids...)
	return tc
}

// AddServices adds the "services" edges to the Service entity.
func (tc *TeamCreate) AddServices(s ...*Service) *TeamCreate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return tc.AddServiceIDs(ids...)
}

// AddOncallRosterIDs adds the "oncall_rosters" edge to the OncallRoster entity by IDs.
func (tc *TeamCreate) AddOncallRosterIDs(ids ...uuid.UUID) *TeamCreate {
	tc.mutation.AddOncallRosterIDs(ids...)
	return tc
}

// AddOncallRosters adds the "oncall_rosters" edges to the OncallRoster entity.
func (tc *TeamCreate) AddOncallRosters(o ...*OncallRoster) *TeamCreate {
	ids := make([]uuid.UUID, len(o))
	for i := range o {
		ids[i] = o[i].ID
	}
	return tc.AddOncallRosterIDs(ids...)
}

// AddSubscriptionIDs adds the "subscriptions" edge to the Subscription entity by IDs.
func (tc *TeamCreate) AddSubscriptionIDs(ids ...uuid.UUID) *TeamCreate {
	tc.mutation.AddSubscriptionIDs(ids...)
	return tc
}

// AddSubscriptions adds the "subscriptions" edges to the Subscription entity.
func (tc *TeamCreate) AddSubscriptions(s ...*Subscription) *TeamCreate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return tc.AddSubscriptionIDs(ids...)
}

// AddIncidentAssignmentIDs adds the "incident_assignments" edge to the IncidentTeamAssignment entity by IDs.
func (tc *TeamCreate) AddIncidentAssignmentIDs(ids ...int) *TeamCreate {
	tc.mutation.AddIncidentAssignmentIDs(ids...)
	return tc
}

// AddIncidentAssignments adds the "incident_assignments" edges to the IncidentTeamAssignment entity.
func (tc *TeamCreate) AddIncidentAssignments(i ...*IncidentTeamAssignment) *TeamCreate {
	ids := make([]int, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return tc.AddIncidentAssignmentIDs(ids...)
}

// AddScheduledMeetingIDs adds the "scheduled_meetings" edge to the MeetingSchedule entity by IDs.
func (tc *TeamCreate) AddScheduledMeetingIDs(ids ...uuid.UUID) *TeamCreate {
	tc.mutation.AddScheduledMeetingIDs(ids...)
	return tc
}

// AddScheduledMeetings adds the "scheduled_meetings" edges to the MeetingSchedule entity.
func (tc *TeamCreate) AddScheduledMeetings(m ...*MeetingSchedule) *TeamCreate {
	ids := make([]uuid.UUID, len(m))
	for i := range m {
		ids[i] = m[i].ID
	}
	return tc.AddScheduledMeetingIDs(ids...)
}

// Mutation returns the TeamMutation object of the builder.
func (tc *TeamCreate) Mutation() *TeamMutation {
	return tc.mutation
}

// Save creates the Team in the database.
func (tc *TeamCreate) Save(ctx context.Context) (*Team, error) {
	tc.defaults()
	return withHooks(ctx, tc.sqlSave, tc.mutation, tc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (tc *TeamCreate) SaveX(ctx context.Context) *Team {
	v, err := tc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (tc *TeamCreate) Exec(ctx context.Context) error {
	_, err := tc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (tc *TeamCreate) ExecX(ctx context.Context) {
	if err := tc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (tc *TeamCreate) defaults() {
	if _, ok := tc.mutation.ID(); !ok {
		v := team.DefaultID()
		tc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (tc *TeamCreate) check() error {
	if _, ok := tc.mutation.Slug(); !ok {
		return &ValidationError{Name: "slug", err: errors.New(`ent: missing required field "Team.slug"`)}
	}
	if _, ok := tc.mutation.Name(); !ok {
		return &ValidationError{Name: "name", err: errors.New(`ent: missing required field "Team.name"`)}
	}
	return nil
}

func (tc *TeamCreate) sqlSave(ctx context.Context) (*Team, error) {
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

func (tc *TeamCreate) createSpec() (*Team, *sqlgraph.CreateSpec) {
	var (
		_node = &Team{config: tc.config}
		_spec = sqlgraph.NewCreateSpec(team.Table, sqlgraph.NewFieldSpec(team.FieldID, field.TypeUUID))
	)
	_spec.OnConflict = tc.conflict
	if id, ok := tc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := tc.mutation.Slug(); ok {
		_spec.SetField(team.FieldSlug, field.TypeString, value)
		_node.Slug = value
	}
	if value, ok := tc.mutation.Name(); ok {
		_spec.SetField(team.FieldName, field.TypeString, value)
		_node.Name = value
	}
	if value, ok := tc.mutation.ChatChannelID(); ok {
		_spec.SetField(team.FieldChatChannelID, field.TypeString, value)
		_node.ChatChannelID = value
	}
	if value, ok := tc.mutation.Timezone(); ok {
		_spec.SetField(team.FieldTimezone, field.TypeString, value)
		_node.Timezone = value
	}
	if nodes := tc.mutation.UsersIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   team.UsersTable,
			Columns: team.UsersPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(user.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := tc.mutation.ServicesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   team.ServicesTable,
			Columns: []string{team.ServicesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(service.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := tc.mutation.OncallRostersIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   team.OncallRostersTable,
			Columns: team.OncallRostersPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(oncallroster.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := tc.mutation.SubscriptionsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   team.SubscriptionsTable,
			Columns: []string{team.SubscriptionsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(subscription.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := tc.mutation.IncidentAssignmentsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   team.IncidentAssignmentsTable,
			Columns: []string{team.IncidentAssignmentsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentteamassignment.FieldID, field.TypeInt),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := tc.mutation.ScheduledMeetingsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   team.ScheduledMeetingsTable,
			Columns: team.ScheduledMeetingsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(meetingschedule.FieldID, field.TypeUUID),
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
//	client.Team.Create().
//		SetSlug(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.TeamUpsert) {
//			SetSlug(v+v).
//		}).
//		Exec(ctx)
func (tc *TeamCreate) OnConflict(opts ...sql.ConflictOption) *TeamUpsertOne {
	tc.conflict = opts
	return &TeamUpsertOne{
		create: tc,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.Team.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (tc *TeamCreate) OnConflictColumns(columns ...string) *TeamUpsertOne {
	tc.conflict = append(tc.conflict, sql.ConflictColumns(columns...))
	return &TeamUpsertOne{
		create: tc,
	}
}

type (
	// TeamUpsertOne is the builder for "upsert"-ing
	//  one Team node.
	TeamUpsertOne struct {
		create *TeamCreate
	}

	// TeamUpsert is the "OnConflict" setter.
	TeamUpsert struct {
		*sql.UpdateSet
	}
)

// SetSlug sets the "slug" field.
func (u *TeamUpsert) SetSlug(v string) *TeamUpsert {
	u.Set(team.FieldSlug, v)
	return u
}

// UpdateSlug sets the "slug" field to the value that was provided on create.
func (u *TeamUpsert) UpdateSlug() *TeamUpsert {
	u.SetExcluded(team.FieldSlug)
	return u
}

// SetName sets the "name" field.
func (u *TeamUpsert) SetName(v string) *TeamUpsert {
	u.Set(team.FieldName, v)
	return u
}

// UpdateName sets the "name" field to the value that was provided on create.
func (u *TeamUpsert) UpdateName() *TeamUpsert {
	u.SetExcluded(team.FieldName)
	return u
}

// SetChatChannelID sets the "chat_channel_id" field.
func (u *TeamUpsert) SetChatChannelID(v string) *TeamUpsert {
	u.Set(team.FieldChatChannelID, v)
	return u
}

// UpdateChatChannelID sets the "chat_channel_id" field to the value that was provided on create.
func (u *TeamUpsert) UpdateChatChannelID() *TeamUpsert {
	u.SetExcluded(team.FieldChatChannelID)
	return u
}

// ClearChatChannelID clears the value of the "chat_channel_id" field.
func (u *TeamUpsert) ClearChatChannelID() *TeamUpsert {
	u.SetNull(team.FieldChatChannelID)
	return u
}

// SetTimezone sets the "timezone" field.
func (u *TeamUpsert) SetTimezone(v string) *TeamUpsert {
	u.Set(team.FieldTimezone, v)
	return u
}

// UpdateTimezone sets the "timezone" field to the value that was provided on create.
func (u *TeamUpsert) UpdateTimezone() *TeamUpsert {
	u.SetExcluded(team.FieldTimezone)
	return u
}

// ClearTimezone clears the value of the "timezone" field.
func (u *TeamUpsert) ClearTimezone() *TeamUpsert {
	u.SetNull(team.FieldTimezone)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create except the ID field.
// Using this option is equivalent to using:
//
//	client.Team.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(team.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *TeamUpsertOne) UpdateNewValues() *TeamUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.ID(); exists {
			s.SetIgnore(team.FieldID)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.Team.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *TeamUpsertOne) Ignore() *TeamUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *TeamUpsertOne) DoNothing() *TeamUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the TeamCreate.OnConflict
// documentation for more info.
func (u *TeamUpsertOne) Update(set func(*TeamUpsert)) *TeamUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&TeamUpsert{UpdateSet: update})
	}))
	return u
}

// SetSlug sets the "slug" field.
func (u *TeamUpsertOne) SetSlug(v string) *TeamUpsertOne {
	return u.Update(func(s *TeamUpsert) {
		s.SetSlug(v)
	})
}

// UpdateSlug sets the "slug" field to the value that was provided on create.
func (u *TeamUpsertOne) UpdateSlug() *TeamUpsertOne {
	return u.Update(func(s *TeamUpsert) {
		s.UpdateSlug()
	})
}

// SetName sets the "name" field.
func (u *TeamUpsertOne) SetName(v string) *TeamUpsertOne {
	return u.Update(func(s *TeamUpsert) {
		s.SetName(v)
	})
}

// UpdateName sets the "name" field to the value that was provided on create.
func (u *TeamUpsertOne) UpdateName() *TeamUpsertOne {
	return u.Update(func(s *TeamUpsert) {
		s.UpdateName()
	})
}

// SetChatChannelID sets the "chat_channel_id" field.
func (u *TeamUpsertOne) SetChatChannelID(v string) *TeamUpsertOne {
	return u.Update(func(s *TeamUpsert) {
		s.SetChatChannelID(v)
	})
}

// UpdateChatChannelID sets the "chat_channel_id" field to the value that was provided on create.
func (u *TeamUpsertOne) UpdateChatChannelID() *TeamUpsertOne {
	return u.Update(func(s *TeamUpsert) {
		s.UpdateChatChannelID()
	})
}

// ClearChatChannelID clears the value of the "chat_channel_id" field.
func (u *TeamUpsertOne) ClearChatChannelID() *TeamUpsertOne {
	return u.Update(func(s *TeamUpsert) {
		s.ClearChatChannelID()
	})
}

// SetTimezone sets the "timezone" field.
func (u *TeamUpsertOne) SetTimezone(v string) *TeamUpsertOne {
	return u.Update(func(s *TeamUpsert) {
		s.SetTimezone(v)
	})
}

// UpdateTimezone sets the "timezone" field to the value that was provided on create.
func (u *TeamUpsertOne) UpdateTimezone() *TeamUpsertOne {
	return u.Update(func(s *TeamUpsert) {
		s.UpdateTimezone()
	})
}

// ClearTimezone clears the value of the "timezone" field.
func (u *TeamUpsertOne) ClearTimezone() *TeamUpsertOne {
	return u.Update(func(s *TeamUpsert) {
		s.ClearTimezone()
	})
}

// Exec executes the query.
func (u *TeamUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for TeamCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *TeamUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *TeamUpsertOne) ID(ctx context.Context) (id uuid.UUID, err error) {
	if u.create.driver.Dialect() == dialect.MySQL {
		// In case of "ON CONFLICT", there is no way to get back non-numeric ID
		// fields from the database since MySQL does not support the RETURNING clause.
		return id, errors.New("ent: TeamUpsertOne.ID is not supported by MySQL driver. Use TeamUpsertOne.Exec instead")
	}
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *TeamUpsertOne) IDX(ctx context.Context) uuid.UUID {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// TeamCreateBulk is the builder for creating many Team entities in bulk.
type TeamCreateBulk struct {
	config
	err      error
	builders []*TeamCreate
	conflict []sql.ConflictOption
}

// Save creates the Team entities in the database.
func (tcb *TeamCreateBulk) Save(ctx context.Context) ([]*Team, error) {
	if tcb.err != nil {
		return nil, tcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(tcb.builders))
	nodes := make([]*Team, len(tcb.builders))
	mutators := make([]Mutator, len(tcb.builders))
	for i := range tcb.builders {
		func(i int, root context.Context) {
			builder := tcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*TeamMutation)
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
func (tcb *TeamCreateBulk) SaveX(ctx context.Context) []*Team {
	v, err := tcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (tcb *TeamCreateBulk) Exec(ctx context.Context) error {
	_, err := tcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (tcb *TeamCreateBulk) ExecX(ctx context.Context) {
	if err := tcb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.Team.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.TeamUpsert) {
//			SetSlug(v+v).
//		}).
//		Exec(ctx)
func (tcb *TeamCreateBulk) OnConflict(opts ...sql.ConflictOption) *TeamUpsertBulk {
	tcb.conflict = opts
	return &TeamUpsertBulk{
		create: tcb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.Team.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (tcb *TeamCreateBulk) OnConflictColumns(columns ...string) *TeamUpsertBulk {
	tcb.conflict = append(tcb.conflict, sql.ConflictColumns(columns...))
	return &TeamUpsertBulk{
		create: tcb,
	}
}

// TeamUpsertBulk is the builder for "upsert"-ing
// a bulk of Team nodes.
type TeamUpsertBulk struct {
	create *TeamCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.Team.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(team.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *TeamUpsertBulk) UpdateNewValues() *TeamUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.ID(); exists {
				s.SetIgnore(team.FieldID)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.Team.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *TeamUpsertBulk) Ignore() *TeamUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *TeamUpsertBulk) DoNothing() *TeamUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the TeamCreateBulk.OnConflict
// documentation for more info.
func (u *TeamUpsertBulk) Update(set func(*TeamUpsert)) *TeamUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&TeamUpsert{UpdateSet: update})
	}))
	return u
}

// SetSlug sets the "slug" field.
func (u *TeamUpsertBulk) SetSlug(v string) *TeamUpsertBulk {
	return u.Update(func(s *TeamUpsert) {
		s.SetSlug(v)
	})
}

// UpdateSlug sets the "slug" field to the value that was provided on create.
func (u *TeamUpsertBulk) UpdateSlug() *TeamUpsertBulk {
	return u.Update(func(s *TeamUpsert) {
		s.UpdateSlug()
	})
}

// SetName sets the "name" field.
func (u *TeamUpsertBulk) SetName(v string) *TeamUpsertBulk {
	return u.Update(func(s *TeamUpsert) {
		s.SetName(v)
	})
}

// UpdateName sets the "name" field to the value that was provided on create.
func (u *TeamUpsertBulk) UpdateName() *TeamUpsertBulk {
	return u.Update(func(s *TeamUpsert) {
		s.UpdateName()
	})
}

// SetChatChannelID sets the "chat_channel_id" field.
func (u *TeamUpsertBulk) SetChatChannelID(v string) *TeamUpsertBulk {
	return u.Update(func(s *TeamUpsert) {
		s.SetChatChannelID(v)
	})
}

// UpdateChatChannelID sets the "chat_channel_id" field to the value that was provided on create.
func (u *TeamUpsertBulk) UpdateChatChannelID() *TeamUpsertBulk {
	return u.Update(func(s *TeamUpsert) {
		s.UpdateChatChannelID()
	})
}

// ClearChatChannelID clears the value of the "chat_channel_id" field.
func (u *TeamUpsertBulk) ClearChatChannelID() *TeamUpsertBulk {
	return u.Update(func(s *TeamUpsert) {
		s.ClearChatChannelID()
	})
}

// SetTimezone sets the "timezone" field.
func (u *TeamUpsertBulk) SetTimezone(v string) *TeamUpsertBulk {
	return u.Update(func(s *TeamUpsert) {
		s.SetTimezone(v)
	})
}

// UpdateTimezone sets the "timezone" field to the value that was provided on create.
func (u *TeamUpsertBulk) UpdateTimezone() *TeamUpsertBulk {
	return u.Update(func(s *TeamUpsert) {
		s.UpdateTimezone()
	})
}

// ClearTimezone clears the value of the "timezone" field.
func (u *TeamUpsertBulk) ClearTimezone() *TeamUpsertBulk {
	return u.Update(func(s *TeamUpsert) {
		s.ClearTimezone()
	})
}

// Exec executes the query.
func (u *TeamUpsertBulk) Exec(ctx context.Context) error {
	if u.create.err != nil {
		return u.create.err
	}
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the TeamCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for TeamCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *TeamUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}
