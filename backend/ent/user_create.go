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
	"github.com/rezible/rezible/ent/incidentdebrief"
	"github.com/rezible/rezible/ent/incidentroleassignment"
	"github.com/rezible/rezible/ent/oncallannotation"
	"github.com/rezible/rezible/ent/oncallroster"
	"github.com/rezible/rezible/ent/oncallscheduleparticipant"
	"github.com/rezible/rezible/ent/oncallusershift"
	"github.com/rezible/rezible/ent/retrospectivereview"
	"github.com/rezible/rezible/ent/task"
	"github.com/rezible/rezible/ent/team"
	"github.com/rezible/rezible/ent/user"
)

// UserCreate is the builder for creating a User entity.
type UserCreate struct {
	config
	mutation *UserMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetName sets the "name" field.
func (uc *UserCreate) SetName(s string) *UserCreate {
	uc.mutation.SetName(s)
	return uc
}

// SetEmail sets the "email" field.
func (uc *UserCreate) SetEmail(s string) *UserCreate {
	uc.mutation.SetEmail(s)
	return uc
}

// SetChatID sets the "chat_id" field.
func (uc *UserCreate) SetChatID(s string) *UserCreate {
	uc.mutation.SetChatID(s)
	return uc
}

// SetNillableChatID sets the "chat_id" field if the given value is not nil.
func (uc *UserCreate) SetNillableChatID(s *string) *UserCreate {
	if s != nil {
		uc.SetChatID(*s)
	}
	return uc
}

// SetTimezone sets the "timezone" field.
func (uc *UserCreate) SetTimezone(s string) *UserCreate {
	uc.mutation.SetTimezone(s)
	return uc
}

// SetNillableTimezone sets the "timezone" field if the given value is not nil.
func (uc *UserCreate) SetNillableTimezone(s *string) *UserCreate {
	if s != nil {
		uc.SetTimezone(*s)
	}
	return uc
}

// SetID sets the "id" field.
func (uc *UserCreate) SetID(u uuid.UUID) *UserCreate {
	uc.mutation.SetID(u)
	return uc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (uc *UserCreate) SetNillableID(u *uuid.UUID) *UserCreate {
	if u != nil {
		uc.SetID(*u)
	}
	return uc
}

// AddTeamIDs adds the "teams" edge to the Team entity by IDs.
func (uc *UserCreate) AddTeamIDs(ids ...uuid.UUID) *UserCreate {
	uc.mutation.AddTeamIDs(ids...)
	return uc
}

// AddTeams adds the "teams" edges to the Team entity.
func (uc *UserCreate) AddTeams(t ...*Team) *UserCreate {
	ids := make([]uuid.UUID, len(t))
	for i := range t {
		ids[i] = t[i].ID
	}
	return uc.AddTeamIDs(ids...)
}

// AddWatchedOncallRosterIDs adds the "watched_oncall_rosters" edge to the OncallRoster entity by IDs.
func (uc *UserCreate) AddWatchedOncallRosterIDs(ids ...uuid.UUID) *UserCreate {
	uc.mutation.AddWatchedOncallRosterIDs(ids...)
	return uc
}

// AddWatchedOncallRosters adds the "watched_oncall_rosters" edges to the OncallRoster entity.
func (uc *UserCreate) AddWatchedOncallRosters(o ...*OncallRoster) *UserCreate {
	ids := make([]uuid.UUID, len(o))
	for i := range o {
		ids[i] = o[i].ID
	}
	return uc.AddWatchedOncallRosterIDs(ids...)
}

// AddOncallScheduleIDs adds the "oncall_schedules" edge to the OncallScheduleParticipant entity by IDs.
func (uc *UserCreate) AddOncallScheduleIDs(ids ...uuid.UUID) *UserCreate {
	uc.mutation.AddOncallScheduleIDs(ids...)
	return uc
}

// AddOncallSchedules adds the "oncall_schedules" edges to the OncallScheduleParticipant entity.
func (uc *UserCreate) AddOncallSchedules(o ...*OncallScheduleParticipant) *UserCreate {
	ids := make([]uuid.UUID, len(o))
	for i := range o {
		ids[i] = o[i].ID
	}
	return uc.AddOncallScheduleIDs(ids...)
}

// AddOncallShiftIDs adds the "oncall_shifts" edge to the OncallUserShift entity by IDs.
func (uc *UserCreate) AddOncallShiftIDs(ids ...uuid.UUID) *UserCreate {
	uc.mutation.AddOncallShiftIDs(ids...)
	return uc
}

// AddOncallShifts adds the "oncall_shifts" edges to the OncallUserShift entity.
func (uc *UserCreate) AddOncallShifts(o ...*OncallUserShift) *UserCreate {
	ids := make([]uuid.UUID, len(o))
	for i := range o {
		ids[i] = o[i].ID
	}
	return uc.AddOncallShiftIDs(ids...)
}

// AddOncallAnnotationIDs adds the "oncall_annotations" edge to the OncallAnnotation entity by IDs.
func (uc *UserCreate) AddOncallAnnotationIDs(ids ...uuid.UUID) *UserCreate {
	uc.mutation.AddOncallAnnotationIDs(ids...)
	return uc
}

// AddOncallAnnotations adds the "oncall_annotations" edges to the OncallAnnotation entity.
func (uc *UserCreate) AddOncallAnnotations(o ...*OncallAnnotation) *UserCreate {
	ids := make([]uuid.UUID, len(o))
	for i := range o {
		ids[i] = o[i].ID
	}
	return uc.AddOncallAnnotationIDs(ids...)
}

// AddIncidentRoleAssignmentIDs adds the "incident_role_assignments" edge to the IncidentRoleAssignment entity by IDs.
func (uc *UserCreate) AddIncidentRoleAssignmentIDs(ids ...uuid.UUID) *UserCreate {
	uc.mutation.AddIncidentRoleAssignmentIDs(ids...)
	return uc
}

// AddIncidentRoleAssignments adds the "incident_role_assignments" edges to the IncidentRoleAssignment entity.
func (uc *UserCreate) AddIncidentRoleAssignments(i ...*IncidentRoleAssignment) *UserCreate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return uc.AddIncidentRoleAssignmentIDs(ids...)
}

// AddIncidentDebriefIDs adds the "incident_debriefs" edge to the IncidentDebrief entity by IDs.
func (uc *UserCreate) AddIncidentDebriefIDs(ids ...uuid.UUID) *UserCreate {
	uc.mutation.AddIncidentDebriefIDs(ids...)
	return uc
}

// AddIncidentDebriefs adds the "incident_debriefs" edges to the IncidentDebrief entity.
func (uc *UserCreate) AddIncidentDebriefs(i ...*IncidentDebrief) *UserCreate {
	ids := make([]uuid.UUID, len(i))
	for j := range i {
		ids[j] = i[j].ID
	}
	return uc.AddIncidentDebriefIDs(ids...)
}

// AddAssignedTaskIDs adds the "assigned_tasks" edge to the Task entity by IDs.
func (uc *UserCreate) AddAssignedTaskIDs(ids ...uuid.UUID) *UserCreate {
	uc.mutation.AddAssignedTaskIDs(ids...)
	return uc
}

// AddAssignedTasks adds the "assigned_tasks" edges to the Task entity.
func (uc *UserCreate) AddAssignedTasks(t ...*Task) *UserCreate {
	ids := make([]uuid.UUID, len(t))
	for i := range t {
		ids[i] = t[i].ID
	}
	return uc.AddAssignedTaskIDs(ids...)
}

// AddCreatedTaskIDs adds the "created_tasks" edge to the Task entity by IDs.
func (uc *UserCreate) AddCreatedTaskIDs(ids ...uuid.UUID) *UserCreate {
	uc.mutation.AddCreatedTaskIDs(ids...)
	return uc
}

// AddCreatedTasks adds the "created_tasks" edges to the Task entity.
func (uc *UserCreate) AddCreatedTasks(t ...*Task) *UserCreate {
	ids := make([]uuid.UUID, len(t))
	for i := range t {
		ids[i] = t[i].ID
	}
	return uc.AddCreatedTaskIDs(ids...)
}

// AddRetrospectiveReviewRequestIDs adds the "retrospective_review_requests" edge to the RetrospectiveReview entity by IDs.
func (uc *UserCreate) AddRetrospectiveReviewRequestIDs(ids ...uuid.UUID) *UserCreate {
	uc.mutation.AddRetrospectiveReviewRequestIDs(ids...)
	return uc
}

// AddRetrospectiveReviewRequests adds the "retrospective_review_requests" edges to the RetrospectiveReview entity.
func (uc *UserCreate) AddRetrospectiveReviewRequests(r ...*RetrospectiveReview) *UserCreate {
	ids := make([]uuid.UUID, len(r))
	for i := range r {
		ids[i] = r[i].ID
	}
	return uc.AddRetrospectiveReviewRequestIDs(ids...)
}

// AddRetrospectiveReviewResponseIDs adds the "retrospective_review_responses" edge to the RetrospectiveReview entity by IDs.
func (uc *UserCreate) AddRetrospectiveReviewResponseIDs(ids ...uuid.UUID) *UserCreate {
	uc.mutation.AddRetrospectiveReviewResponseIDs(ids...)
	return uc
}

// AddRetrospectiveReviewResponses adds the "retrospective_review_responses" edges to the RetrospectiveReview entity.
func (uc *UserCreate) AddRetrospectiveReviewResponses(r ...*RetrospectiveReview) *UserCreate {
	ids := make([]uuid.UUID, len(r))
	for i := range r {
		ids[i] = r[i].ID
	}
	return uc.AddRetrospectiveReviewResponseIDs(ids...)
}

// Mutation returns the UserMutation object of the builder.
func (uc *UserCreate) Mutation() *UserMutation {
	return uc.mutation
}

// Save creates the User in the database.
func (uc *UserCreate) Save(ctx context.Context) (*User, error) {
	uc.defaults()
	return withHooks(ctx, uc.sqlSave, uc.mutation, uc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (uc *UserCreate) SaveX(ctx context.Context) *User {
	v, err := uc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (uc *UserCreate) Exec(ctx context.Context) error {
	_, err := uc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (uc *UserCreate) ExecX(ctx context.Context) {
	if err := uc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (uc *UserCreate) defaults() {
	if _, ok := uc.mutation.ID(); !ok {
		v := user.DefaultID()
		uc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (uc *UserCreate) check() error {
	if _, ok := uc.mutation.Name(); !ok {
		return &ValidationError{Name: "name", err: errors.New(`ent: missing required field "User.name"`)}
	}
	if _, ok := uc.mutation.Email(); !ok {
		return &ValidationError{Name: "email", err: errors.New(`ent: missing required field "User.email"`)}
	}
	return nil
}

func (uc *UserCreate) sqlSave(ctx context.Context) (*User, error) {
	if err := uc.check(); err != nil {
		return nil, err
	}
	_node, _spec := uc.createSpec()
	if err := sqlgraph.CreateNode(ctx, uc.driver, _spec); err != nil {
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
	uc.mutation.id = &_node.ID
	uc.mutation.done = true
	return _node, nil
}

func (uc *UserCreate) createSpec() (*User, *sqlgraph.CreateSpec) {
	var (
		_node = &User{config: uc.config}
		_spec = sqlgraph.NewCreateSpec(user.Table, sqlgraph.NewFieldSpec(user.FieldID, field.TypeUUID))
	)
	_spec.OnConflict = uc.conflict
	if id, ok := uc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := uc.mutation.Name(); ok {
		_spec.SetField(user.FieldName, field.TypeString, value)
		_node.Name = value
	}
	if value, ok := uc.mutation.Email(); ok {
		_spec.SetField(user.FieldEmail, field.TypeString, value)
		_node.Email = value
	}
	if value, ok := uc.mutation.ChatID(); ok {
		_spec.SetField(user.FieldChatID, field.TypeString, value)
		_node.ChatID = value
	}
	if value, ok := uc.mutation.Timezone(); ok {
		_spec.SetField(user.FieldTimezone, field.TypeString, value)
		_node.Timezone = value
	}
	if nodes := uc.mutation.TeamsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: true,
			Table:   user.TeamsTable,
			Columns: user.TeamsPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(team.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := uc.mutation.WatchedOncallRostersIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   user.WatchedOncallRostersTable,
			Columns: user.WatchedOncallRostersPrimaryKey,
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
	if nodes := uc.mutation.OncallSchedulesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   user.OncallSchedulesTable,
			Columns: []string{user.OncallSchedulesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(oncallscheduleparticipant.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := uc.mutation.OncallShiftsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   user.OncallShiftsTable,
			Columns: []string{user.OncallShiftsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(oncallusershift.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := uc.mutation.OncallAnnotationsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   user.OncallAnnotationsTable,
			Columns: []string{user.OncallAnnotationsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(oncallannotation.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := uc.mutation.IncidentRoleAssignmentsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   user.IncidentRoleAssignmentsTable,
			Columns: []string{user.IncidentRoleAssignmentsColumn},
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
	if nodes := uc.mutation.IncidentDebriefsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   user.IncidentDebriefsTable,
			Columns: []string{user.IncidentDebriefsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(incidentdebrief.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := uc.mutation.AssignedTasksIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   user.AssignedTasksTable,
			Columns: []string{user.AssignedTasksColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(task.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := uc.mutation.CreatedTasksIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   user.CreatedTasksTable,
			Columns: []string{user.CreatedTasksColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(task.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := uc.mutation.RetrospectiveReviewRequestsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   user.RetrospectiveReviewRequestsTable,
			Columns: []string{user.RetrospectiveReviewRequestsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(retrospectivereview.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := uc.mutation.RetrospectiveReviewResponsesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: true,
			Table:   user.RetrospectiveReviewResponsesTable,
			Columns: []string{user.RetrospectiveReviewResponsesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(retrospectivereview.FieldID, field.TypeUUID),
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
//	client.User.Create().
//		SetName(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.UserUpsert) {
//			SetName(v+v).
//		}).
//		Exec(ctx)
func (uc *UserCreate) OnConflict(opts ...sql.ConflictOption) *UserUpsertOne {
	uc.conflict = opts
	return &UserUpsertOne{
		create: uc,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.User.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (uc *UserCreate) OnConflictColumns(columns ...string) *UserUpsertOne {
	uc.conflict = append(uc.conflict, sql.ConflictColumns(columns...))
	return &UserUpsertOne{
		create: uc,
	}
}

type (
	// UserUpsertOne is the builder for "upsert"-ing
	//  one User node.
	UserUpsertOne struct {
		create *UserCreate
	}

	// UserUpsert is the "OnConflict" setter.
	UserUpsert struct {
		*sql.UpdateSet
	}
)

// SetName sets the "name" field.
func (u *UserUpsert) SetName(v string) *UserUpsert {
	u.Set(user.FieldName, v)
	return u
}

// UpdateName sets the "name" field to the value that was provided on create.
func (u *UserUpsert) UpdateName() *UserUpsert {
	u.SetExcluded(user.FieldName)
	return u
}

// SetEmail sets the "email" field.
func (u *UserUpsert) SetEmail(v string) *UserUpsert {
	u.Set(user.FieldEmail, v)
	return u
}

// UpdateEmail sets the "email" field to the value that was provided on create.
func (u *UserUpsert) UpdateEmail() *UserUpsert {
	u.SetExcluded(user.FieldEmail)
	return u
}

// SetChatID sets the "chat_id" field.
func (u *UserUpsert) SetChatID(v string) *UserUpsert {
	u.Set(user.FieldChatID, v)
	return u
}

// UpdateChatID sets the "chat_id" field to the value that was provided on create.
func (u *UserUpsert) UpdateChatID() *UserUpsert {
	u.SetExcluded(user.FieldChatID)
	return u
}

// ClearChatID clears the value of the "chat_id" field.
func (u *UserUpsert) ClearChatID() *UserUpsert {
	u.SetNull(user.FieldChatID)
	return u
}

// SetTimezone sets the "timezone" field.
func (u *UserUpsert) SetTimezone(v string) *UserUpsert {
	u.Set(user.FieldTimezone, v)
	return u
}

// UpdateTimezone sets the "timezone" field to the value that was provided on create.
func (u *UserUpsert) UpdateTimezone() *UserUpsert {
	u.SetExcluded(user.FieldTimezone)
	return u
}

// ClearTimezone clears the value of the "timezone" field.
func (u *UserUpsert) ClearTimezone() *UserUpsert {
	u.SetNull(user.FieldTimezone)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create except the ID field.
// Using this option is equivalent to using:
//
//	client.User.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(user.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *UserUpsertOne) UpdateNewValues() *UserUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.ID(); exists {
			s.SetIgnore(user.FieldID)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.User.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *UserUpsertOne) Ignore() *UserUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *UserUpsertOne) DoNothing() *UserUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the UserCreate.OnConflict
// documentation for more info.
func (u *UserUpsertOne) Update(set func(*UserUpsert)) *UserUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&UserUpsert{UpdateSet: update})
	}))
	return u
}

// SetName sets the "name" field.
func (u *UserUpsertOne) SetName(v string) *UserUpsertOne {
	return u.Update(func(s *UserUpsert) {
		s.SetName(v)
	})
}

// UpdateName sets the "name" field to the value that was provided on create.
func (u *UserUpsertOne) UpdateName() *UserUpsertOne {
	return u.Update(func(s *UserUpsert) {
		s.UpdateName()
	})
}

// SetEmail sets the "email" field.
func (u *UserUpsertOne) SetEmail(v string) *UserUpsertOne {
	return u.Update(func(s *UserUpsert) {
		s.SetEmail(v)
	})
}

// UpdateEmail sets the "email" field to the value that was provided on create.
func (u *UserUpsertOne) UpdateEmail() *UserUpsertOne {
	return u.Update(func(s *UserUpsert) {
		s.UpdateEmail()
	})
}

// SetChatID sets the "chat_id" field.
func (u *UserUpsertOne) SetChatID(v string) *UserUpsertOne {
	return u.Update(func(s *UserUpsert) {
		s.SetChatID(v)
	})
}

// UpdateChatID sets the "chat_id" field to the value that was provided on create.
func (u *UserUpsertOne) UpdateChatID() *UserUpsertOne {
	return u.Update(func(s *UserUpsert) {
		s.UpdateChatID()
	})
}

// ClearChatID clears the value of the "chat_id" field.
func (u *UserUpsertOne) ClearChatID() *UserUpsertOne {
	return u.Update(func(s *UserUpsert) {
		s.ClearChatID()
	})
}

// SetTimezone sets the "timezone" field.
func (u *UserUpsertOne) SetTimezone(v string) *UserUpsertOne {
	return u.Update(func(s *UserUpsert) {
		s.SetTimezone(v)
	})
}

// UpdateTimezone sets the "timezone" field to the value that was provided on create.
func (u *UserUpsertOne) UpdateTimezone() *UserUpsertOne {
	return u.Update(func(s *UserUpsert) {
		s.UpdateTimezone()
	})
}

// ClearTimezone clears the value of the "timezone" field.
func (u *UserUpsertOne) ClearTimezone() *UserUpsertOne {
	return u.Update(func(s *UserUpsert) {
		s.ClearTimezone()
	})
}

// Exec executes the query.
func (u *UserUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for UserCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *UserUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *UserUpsertOne) ID(ctx context.Context) (id uuid.UUID, err error) {
	if u.create.driver.Dialect() == dialect.MySQL {
		// In case of "ON CONFLICT", there is no way to get back non-numeric ID
		// fields from the database since MySQL does not support the RETURNING clause.
		return id, errors.New("ent: UserUpsertOne.ID is not supported by MySQL driver. Use UserUpsertOne.Exec instead")
	}
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *UserUpsertOne) IDX(ctx context.Context) uuid.UUID {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// UserCreateBulk is the builder for creating many User entities in bulk.
type UserCreateBulk struct {
	config
	err      error
	builders []*UserCreate
	conflict []sql.ConflictOption
}

// Save creates the User entities in the database.
func (ucb *UserCreateBulk) Save(ctx context.Context) ([]*User, error) {
	if ucb.err != nil {
		return nil, ucb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(ucb.builders))
	nodes := make([]*User, len(ucb.builders))
	mutators := make([]Mutator, len(ucb.builders))
	for i := range ucb.builders {
		func(i int, root context.Context) {
			builder := ucb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*UserMutation)
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
					_, err = mutators[i+1].Mutate(root, ucb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = ucb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, ucb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, ucb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (ucb *UserCreateBulk) SaveX(ctx context.Context) []*User {
	v, err := ucb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (ucb *UserCreateBulk) Exec(ctx context.Context) error {
	_, err := ucb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ucb *UserCreateBulk) ExecX(ctx context.Context) {
	if err := ucb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.User.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.UserUpsert) {
//			SetName(v+v).
//		}).
//		Exec(ctx)
func (ucb *UserCreateBulk) OnConflict(opts ...sql.ConflictOption) *UserUpsertBulk {
	ucb.conflict = opts
	return &UserUpsertBulk{
		create: ucb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.User.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (ucb *UserCreateBulk) OnConflictColumns(columns ...string) *UserUpsertBulk {
	ucb.conflict = append(ucb.conflict, sql.ConflictColumns(columns...))
	return &UserUpsertBulk{
		create: ucb,
	}
}

// UserUpsertBulk is the builder for "upsert"-ing
// a bulk of User nodes.
type UserUpsertBulk struct {
	create *UserCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.User.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(user.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *UserUpsertBulk) UpdateNewValues() *UserUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.ID(); exists {
				s.SetIgnore(user.FieldID)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.User.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *UserUpsertBulk) Ignore() *UserUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *UserUpsertBulk) DoNothing() *UserUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the UserCreateBulk.OnConflict
// documentation for more info.
func (u *UserUpsertBulk) Update(set func(*UserUpsert)) *UserUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&UserUpsert{UpdateSet: update})
	}))
	return u
}

// SetName sets the "name" field.
func (u *UserUpsertBulk) SetName(v string) *UserUpsertBulk {
	return u.Update(func(s *UserUpsert) {
		s.SetName(v)
	})
}

// UpdateName sets the "name" field to the value that was provided on create.
func (u *UserUpsertBulk) UpdateName() *UserUpsertBulk {
	return u.Update(func(s *UserUpsert) {
		s.UpdateName()
	})
}

// SetEmail sets the "email" field.
func (u *UserUpsertBulk) SetEmail(v string) *UserUpsertBulk {
	return u.Update(func(s *UserUpsert) {
		s.SetEmail(v)
	})
}

// UpdateEmail sets the "email" field to the value that was provided on create.
func (u *UserUpsertBulk) UpdateEmail() *UserUpsertBulk {
	return u.Update(func(s *UserUpsert) {
		s.UpdateEmail()
	})
}

// SetChatID sets the "chat_id" field.
func (u *UserUpsertBulk) SetChatID(v string) *UserUpsertBulk {
	return u.Update(func(s *UserUpsert) {
		s.SetChatID(v)
	})
}

// UpdateChatID sets the "chat_id" field to the value that was provided on create.
func (u *UserUpsertBulk) UpdateChatID() *UserUpsertBulk {
	return u.Update(func(s *UserUpsert) {
		s.UpdateChatID()
	})
}

// ClearChatID clears the value of the "chat_id" field.
func (u *UserUpsertBulk) ClearChatID() *UserUpsertBulk {
	return u.Update(func(s *UserUpsert) {
		s.ClearChatID()
	})
}

// SetTimezone sets the "timezone" field.
func (u *UserUpsertBulk) SetTimezone(v string) *UserUpsertBulk {
	return u.Update(func(s *UserUpsert) {
		s.SetTimezone(v)
	})
}

// UpdateTimezone sets the "timezone" field to the value that was provided on create.
func (u *UserUpsertBulk) UpdateTimezone() *UserUpsertBulk {
	return u.Update(func(s *UserUpsert) {
		s.UpdateTimezone()
	})
}

// ClearTimezone clears the value of the "timezone" field.
func (u *UserUpsertBulk) ClearTimezone() *UserUpsertBulk {
	return u.Update(func(s *UserUpsert) {
		s.ClearTimezone()
	})
}

// Exec executes the query.
func (u *UserUpsertBulk) Exec(ctx context.Context) error {
	if u.create.err != nil {
		return u.create.err
	}
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the UserCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for UserCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *UserUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}
