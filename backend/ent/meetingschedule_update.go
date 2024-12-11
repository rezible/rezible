// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/dialect/sql/sqljson"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/meetingschedule"
	"github.com/rezible/rezible/ent/meetingsession"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/ent/team"
)

// MeetingScheduleUpdate is the builder for updating MeetingSchedule entities.
type MeetingScheduleUpdate struct {
	config
	hooks     []Hook
	mutation  *MeetingScheduleMutation
	modifiers []func(*sql.UpdateBuilder)
}

// Where appends a list predicates to the MeetingScheduleUpdate builder.
func (msu *MeetingScheduleUpdate) Where(ps ...predicate.MeetingSchedule) *MeetingScheduleUpdate {
	msu.mutation.Where(ps...)
	return msu
}

// SetName sets the "name" field.
func (msu *MeetingScheduleUpdate) SetName(s string) *MeetingScheduleUpdate {
	msu.mutation.SetName(s)
	return msu
}

// SetNillableName sets the "name" field if the given value is not nil.
func (msu *MeetingScheduleUpdate) SetNillableName(s *string) *MeetingScheduleUpdate {
	if s != nil {
		msu.SetName(*s)
	}
	return msu
}

// SetDescription sets the "description" field.
func (msu *MeetingScheduleUpdate) SetDescription(s string) *MeetingScheduleUpdate {
	msu.mutation.SetDescription(s)
	return msu
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (msu *MeetingScheduleUpdate) SetNillableDescription(s *string) *MeetingScheduleUpdate {
	if s != nil {
		msu.SetDescription(*s)
	}
	return msu
}

// ClearDescription clears the value of the "description" field.
func (msu *MeetingScheduleUpdate) ClearDescription() *MeetingScheduleUpdate {
	msu.mutation.ClearDescription()
	return msu
}

// SetBeginMinute sets the "begin_minute" field.
func (msu *MeetingScheduleUpdate) SetBeginMinute(i int) *MeetingScheduleUpdate {
	msu.mutation.ResetBeginMinute()
	msu.mutation.SetBeginMinute(i)
	return msu
}

// SetNillableBeginMinute sets the "begin_minute" field if the given value is not nil.
func (msu *MeetingScheduleUpdate) SetNillableBeginMinute(i *int) *MeetingScheduleUpdate {
	if i != nil {
		msu.SetBeginMinute(*i)
	}
	return msu
}

// AddBeginMinute adds i to the "begin_minute" field.
func (msu *MeetingScheduleUpdate) AddBeginMinute(i int) *MeetingScheduleUpdate {
	msu.mutation.AddBeginMinute(i)
	return msu
}

// SetDurationMinutes sets the "duration_minutes" field.
func (msu *MeetingScheduleUpdate) SetDurationMinutes(i int) *MeetingScheduleUpdate {
	msu.mutation.ResetDurationMinutes()
	msu.mutation.SetDurationMinutes(i)
	return msu
}

// SetNillableDurationMinutes sets the "duration_minutes" field if the given value is not nil.
func (msu *MeetingScheduleUpdate) SetNillableDurationMinutes(i *int) *MeetingScheduleUpdate {
	if i != nil {
		msu.SetDurationMinutes(*i)
	}
	return msu
}

// AddDurationMinutes adds i to the "duration_minutes" field.
func (msu *MeetingScheduleUpdate) AddDurationMinutes(i int) *MeetingScheduleUpdate {
	msu.mutation.AddDurationMinutes(i)
	return msu
}

// SetStartDate sets the "start_date" field.
func (msu *MeetingScheduleUpdate) SetStartDate(t time.Time) *MeetingScheduleUpdate {
	msu.mutation.SetStartDate(t)
	return msu
}

// SetNillableStartDate sets the "start_date" field if the given value is not nil.
func (msu *MeetingScheduleUpdate) SetNillableStartDate(t *time.Time) *MeetingScheduleUpdate {
	if t != nil {
		msu.SetStartDate(*t)
	}
	return msu
}

// SetRepeats sets the "repeats" field.
func (msu *MeetingScheduleUpdate) SetRepeats(m meetingschedule.Repeats) *MeetingScheduleUpdate {
	msu.mutation.SetRepeats(m)
	return msu
}

// SetNillableRepeats sets the "repeats" field if the given value is not nil.
func (msu *MeetingScheduleUpdate) SetNillableRepeats(m *meetingschedule.Repeats) *MeetingScheduleUpdate {
	if m != nil {
		msu.SetRepeats(*m)
	}
	return msu
}

// SetRepetitionStep sets the "repetition_step" field.
func (msu *MeetingScheduleUpdate) SetRepetitionStep(i int) *MeetingScheduleUpdate {
	msu.mutation.ResetRepetitionStep()
	msu.mutation.SetRepetitionStep(i)
	return msu
}

// SetNillableRepetitionStep sets the "repetition_step" field if the given value is not nil.
func (msu *MeetingScheduleUpdate) SetNillableRepetitionStep(i *int) *MeetingScheduleUpdate {
	if i != nil {
		msu.SetRepetitionStep(*i)
	}
	return msu
}

// AddRepetitionStep adds i to the "repetition_step" field.
func (msu *MeetingScheduleUpdate) AddRepetitionStep(i int) *MeetingScheduleUpdate {
	msu.mutation.AddRepetitionStep(i)
	return msu
}

// SetWeekDays sets the "week_days" field.
func (msu *MeetingScheduleUpdate) SetWeekDays(s []string) *MeetingScheduleUpdate {
	msu.mutation.SetWeekDays(s)
	return msu
}

// AppendWeekDays appends s to the "week_days" field.
func (msu *MeetingScheduleUpdate) AppendWeekDays(s []string) *MeetingScheduleUpdate {
	msu.mutation.AppendWeekDays(s)
	return msu
}

// ClearWeekDays clears the value of the "week_days" field.
func (msu *MeetingScheduleUpdate) ClearWeekDays() *MeetingScheduleUpdate {
	msu.mutation.ClearWeekDays()
	return msu
}

// SetMonthlyOn sets the "monthly_on" field.
func (msu *MeetingScheduleUpdate) SetMonthlyOn(mo meetingschedule.MonthlyOn) *MeetingScheduleUpdate {
	msu.mutation.SetMonthlyOn(mo)
	return msu
}

// SetNillableMonthlyOn sets the "monthly_on" field if the given value is not nil.
func (msu *MeetingScheduleUpdate) SetNillableMonthlyOn(mo *meetingschedule.MonthlyOn) *MeetingScheduleUpdate {
	if mo != nil {
		msu.SetMonthlyOn(*mo)
	}
	return msu
}

// ClearMonthlyOn clears the value of the "monthly_on" field.
func (msu *MeetingScheduleUpdate) ClearMonthlyOn() *MeetingScheduleUpdate {
	msu.mutation.ClearMonthlyOn()
	return msu
}

// SetUntilDate sets the "until_date" field.
func (msu *MeetingScheduleUpdate) SetUntilDate(t time.Time) *MeetingScheduleUpdate {
	msu.mutation.SetUntilDate(t)
	return msu
}

// SetNillableUntilDate sets the "until_date" field if the given value is not nil.
func (msu *MeetingScheduleUpdate) SetNillableUntilDate(t *time.Time) *MeetingScheduleUpdate {
	if t != nil {
		msu.SetUntilDate(*t)
	}
	return msu
}

// ClearUntilDate clears the value of the "until_date" field.
func (msu *MeetingScheduleUpdate) ClearUntilDate() *MeetingScheduleUpdate {
	msu.mutation.ClearUntilDate()
	return msu
}

// SetNumRepetitions sets the "num_repetitions" field.
func (msu *MeetingScheduleUpdate) SetNumRepetitions(i int) *MeetingScheduleUpdate {
	msu.mutation.ResetNumRepetitions()
	msu.mutation.SetNumRepetitions(i)
	return msu
}

// SetNillableNumRepetitions sets the "num_repetitions" field if the given value is not nil.
func (msu *MeetingScheduleUpdate) SetNillableNumRepetitions(i *int) *MeetingScheduleUpdate {
	if i != nil {
		msu.SetNumRepetitions(*i)
	}
	return msu
}

// AddNumRepetitions adds i to the "num_repetitions" field.
func (msu *MeetingScheduleUpdate) AddNumRepetitions(i int) *MeetingScheduleUpdate {
	msu.mutation.AddNumRepetitions(i)
	return msu
}

// ClearNumRepetitions clears the value of the "num_repetitions" field.
func (msu *MeetingScheduleUpdate) ClearNumRepetitions() *MeetingScheduleUpdate {
	msu.mutation.ClearNumRepetitions()
	return msu
}

// AddSessionIDs adds the "sessions" edge to the MeetingSession entity by IDs.
func (msu *MeetingScheduleUpdate) AddSessionIDs(ids ...uuid.UUID) *MeetingScheduleUpdate {
	msu.mutation.AddSessionIDs(ids...)
	return msu
}

// AddSessions adds the "sessions" edges to the MeetingSession entity.
func (msu *MeetingScheduleUpdate) AddSessions(m ...*MeetingSession) *MeetingScheduleUpdate {
	ids := make([]uuid.UUID, len(m))
	for i := range m {
		ids[i] = m[i].ID
	}
	return msu.AddSessionIDs(ids...)
}

// AddOwningTeamIDs adds the "owning_team" edge to the Team entity by IDs.
func (msu *MeetingScheduleUpdate) AddOwningTeamIDs(ids ...uuid.UUID) *MeetingScheduleUpdate {
	msu.mutation.AddOwningTeamIDs(ids...)
	return msu
}

// AddOwningTeam adds the "owning_team" edges to the Team entity.
func (msu *MeetingScheduleUpdate) AddOwningTeam(t ...*Team) *MeetingScheduleUpdate {
	ids := make([]uuid.UUID, len(t))
	for i := range t {
		ids[i] = t[i].ID
	}
	return msu.AddOwningTeamIDs(ids...)
}

// Mutation returns the MeetingScheduleMutation object of the builder.
func (msu *MeetingScheduleUpdate) Mutation() *MeetingScheduleMutation {
	return msu.mutation
}

// ClearSessions clears all "sessions" edges to the MeetingSession entity.
func (msu *MeetingScheduleUpdate) ClearSessions() *MeetingScheduleUpdate {
	msu.mutation.ClearSessions()
	return msu
}

// RemoveSessionIDs removes the "sessions" edge to MeetingSession entities by IDs.
func (msu *MeetingScheduleUpdate) RemoveSessionIDs(ids ...uuid.UUID) *MeetingScheduleUpdate {
	msu.mutation.RemoveSessionIDs(ids...)
	return msu
}

// RemoveSessions removes "sessions" edges to MeetingSession entities.
func (msu *MeetingScheduleUpdate) RemoveSessions(m ...*MeetingSession) *MeetingScheduleUpdate {
	ids := make([]uuid.UUID, len(m))
	for i := range m {
		ids[i] = m[i].ID
	}
	return msu.RemoveSessionIDs(ids...)
}

// ClearOwningTeam clears all "owning_team" edges to the Team entity.
func (msu *MeetingScheduleUpdate) ClearOwningTeam() *MeetingScheduleUpdate {
	msu.mutation.ClearOwningTeam()
	return msu
}

// RemoveOwningTeamIDs removes the "owning_team" edge to Team entities by IDs.
func (msu *MeetingScheduleUpdate) RemoveOwningTeamIDs(ids ...uuid.UUID) *MeetingScheduleUpdate {
	msu.mutation.RemoveOwningTeamIDs(ids...)
	return msu
}

// RemoveOwningTeam removes "owning_team" edges to Team entities.
func (msu *MeetingScheduleUpdate) RemoveOwningTeam(t ...*Team) *MeetingScheduleUpdate {
	ids := make([]uuid.UUID, len(t))
	for i := range t {
		ids[i] = t[i].ID
	}
	return msu.RemoveOwningTeamIDs(ids...)
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (msu *MeetingScheduleUpdate) Save(ctx context.Context) (int, error) {
	return withHooks(ctx, msu.sqlSave, msu.mutation, msu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (msu *MeetingScheduleUpdate) SaveX(ctx context.Context) int {
	affected, err := msu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (msu *MeetingScheduleUpdate) Exec(ctx context.Context) error {
	_, err := msu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (msu *MeetingScheduleUpdate) ExecX(ctx context.Context) {
	if err := msu.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (msu *MeetingScheduleUpdate) check() error {
	if v, ok := msu.mutation.Repeats(); ok {
		if err := meetingschedule.RepeatsValidator(v); err != nil {
			return &ValidationError{Name: "repeats", err: fmt.Errorf(`ent: validator failed for field "MeetingSchedule.repeats": %w`, err)}
		}
	}
	if v, ok := msu.mutation.WeekDays(); ok {
		if err := meetingschedule.WeekDaysValidator(v); err != nil {
			return &ValidationError{Name: "week_days", err: fmt.Errorf(`ent: validator failed for field "MeetingSchedule.week_days": %w`, err)}
		}
	}
	if v, ok := msu.mutation.MonthlyOn(); ok {
		if err := meetingschedule.MonthlyOnValidator(v); err != nil {
			return &ValidationError{Name: "monthly_on", err: fmt.Errorf(`ent: validator failed for field "MeetingSchedule.monthly_on": %w`, err)}
		}
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (msu *MeetingScheduleUpdate) Modify(modifiers ...func(u *sql.UpdateBuilder)) *MeetingScheduleUpdate {
	msu.modifiers = append(msu.modifiers, modifiers...)
	return msu
}

func (msu *MeetingScheduleUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := msu.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(meetingschedule.Table, meetingschedule.Columns, sqlgraph.NewFieldSpec(meetingschedule.FieldID, field.TypeUUID))
	if ps := msu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := msu.mutation.Name(); ok {
		_spec.SetField(meetingschedule.FieldName, field.TypeString, value)
	}
	if value, ok := msu.mutation.Description(); ok {
		_spec.SetField(meetingschedule.FieldDescription, field.TypeString, value)
	}
	if msu.mutation.DescriptionCleared() {
		_spec.ClearField(meetingschedule.FieldDescription, field.TypeString)
	}
	if value, ok := msu.mutation.BeginMinute(); ok {
		_spec.SetField(meetingschedule.FieldBeginMinute, field.TypeInt, value)
	}
	if value, ok := msu.mutation.AddedBeginMinute(); ok {
		_spec.AddField(meetingschedule.FieldBeginMinute, field.TypeInt, value)
	}
	if value, ok := msu.mutation.DurationMinutes(); ok {
		_spec.SetField(meetingschedule.FieldDurationMinutes, field.TypeInt, value)
	}
	if value, ok := msu.mutation.AddedDurationMinutes(); ok {
		_spec.AddField(meetingschedule.FieldDurationMinutes, field.TypeInt, value)
	}
	if value, ok := msu.mutation.StartDate(); ok {
		_spec.SetField(meetingschedule.FieldStartDate, field.TypeTime, value)
	}
	if value, ok := msu.mutation.Repeats(); ok {
		_spec.SetField(meetingschedule.FieldRepeats, field.TypeEnum, value)
	}
	if value, ok := msu.mutation.RepetitionStep(); ok {
		_spec.SetField(meetingschedule.FieldRepetitionStep, field.TypeInt, value)
	}
	if value, ok := msu.mutation.AddedRepetitionStep(); ok {
		_spec.AddField(meetingschedule.FieldRepetitionStep, field.TypeInt, value)
	}
	if value, ok := msu.mutation.WeekDays(); ok {
		_spec.SetField(meetingschedule.FieldWeekDays, field.TypeJSON, value)
	}
	if value, ok := msu.mutation.AppendedWeekDays(); ok {
		_spec.AddModifier(func(u *sql.UpdateBuilder) {
			sqljson.Append(u, meetingschedule.FieldWeekDays, value)
		})
	}
	if msu.mutation.WeekDaysCleared() {
		_spec.ClearField(meetingschedule.FieldWeekDays, field.TypeJSON)
	}
	if value, ok := msu.mutation.MonthlyOn(); ok {
		_spec.SetField(meetingschedule.FieldMonthlyOn, field.TypeEnum, value)
	}
	if msu.mutation.MonthlyOnCleared() {
		_spec.ClearField(meetingschedule.FieldMonthlyOn, field.TypeEnum)
	}
	if value, ok := msu.mutation.UntilDate(); ok {
		_spec.SetField(meetingschedule.FieldUntilDate, field.TypeTime, value)
	}
	if msu.mutation.UntilDateCleared() {
		_spec.ClearField(meetingschedule.FieldUntilDate, field.TypeTime)
	}
	if value, ok := msu.mutation.NumRepetitions(); ok {
		_spec.SetField(meetingschedule.FieldNumRepetitions, field.TypeInt, value)
	}
	if value, ok := msu.mutation.AddedNumRepetitions(); ok {
		_spec.AddField(meetingschedule.FieldNumRepetitions, field.TypeInt, value)
	}
	if msu.mutation.NumRepetitionsCleared() {
		_spec.ClearField(meetingschedule.FieldNumRepetitions, field.TypeInt)
	}
	if msu.mutation.SessionsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   meetingschedule.SessionsTable,
			Columns: []string{meetingschedule.SessionsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(meetingsession.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := msu.mutation.RemovedSessionsIDs(); len(nodes) > 0 && !msu.mutation.SessionsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   meetingschedule.SessionsTable,
			Columns: []string{meetingschedule.SessionsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(meetingsession.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := msu.mutation.SessionsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   meetingschedule.SessionsTable,
			Columns: []string{meetingschedule.SessionsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(meetingsession.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if msu.mutation.OwningTeamCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   meetingschedule.OwningTeamTable,
			Columns: meetingschedule.OwningTeamPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(team.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := msu.mutation.RemovedOwningTeamIDs(); len(nodes) > 0 && !msu.mutation.OwningTeamCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   meetingschedule.OwningTeamTable,
			Columns: meetingschedule.OwningTeamPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(team.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := msu.mutation.OwningTeamIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   meetingschedule.OwningTeamTable,
			Columns: meetingschedule.OwningTeamPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(team.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(msu.modifiers...)
	if n, err = sqlgraph.UpdateNodes(ctx, msu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{meetingschedule.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	msu.mutation.done = true
	return n, nil
}

// MeetingScheduleUpdateOne is the builder for updating a single MeetingSchedule entity.
type MeetingScheduleUpdateOne struct {
	config
	fields    []string
	hooks     []Hook
	mutation  *MeetingScheduleMutation
	modifiers []func(*sql.UpdateBuilder)
}

// SetName sets the "name" field.
func (msuo *MeetingScheduleUpdateOne) SetName(s string) *MeetingScheduleUpdateOne {
	msuo.mutation.SetName(s)
	return msuo
}

// SetNillableName sets the "name" field if the given value is not nil.
func (msuo *MeetingScheduleUpdateOne) SetNillableName(s *string) *MeetingScheduleUpdateOne {
	if s != nil {
		msuo.SetName(*s)
	}
	return msuo
}

// SetDescription sets the "description" field.
func (msuo *MeetingScheduleUpdateOne) SetDescription(s string) *MeetingScheduleUpdateOne {
	msuo.mutation.SetDescription(s)
	return msuo
}

// SetNillableDescription sets the "description" field if the given value is not nil.
func (msuo *MeetingScheduleUpdateOne) SetNillableDescription(s *string) *MeetingScheduleUpdateOne {
	if s != nil {
		msuo.SetDescription(*s)
	}
	return msuo
}

// ClearDescription clears the value of the "description" field.
func (msuo *MeetingScheduleUpdateOne) ClearDescription() *MeetingScheduleUpdateOne {
	msuo.mutation.ClearDescription()
	return msuo
}

// SetBeginMinute sets the "begin_minute" field.
func (msuo *MeetingScheduleUpdateOne) SetBeginMinute(i int) *MeetingScheduleUpdateOne {
	msuo.mutation.ResetBeginMinute()
	msuo.mutation.SetBeginMinute(i)
	return msuo
}

// SetNillableBeginMinute sets the "begin_minute" field if the given value is not nil.
func (msuo *MeetingScheduleUpdateOne) SetNillableBeginMinute(i *int) *MeetingScheduleUpdateOne {
	if i != nil {
		msuo.SetBeginMinute(*i)
	}
	return msuo
}

// AddBeginMinute adds i to the "begin_minute" field.
func (msuo *MeetingScheduleUpdateOne) AddBeginMinute(i int) *MeetingScheduleUpdateOne {
	msuo.mutation.AddBeginMinute(i)
	return msuo
}

// SetDurationMinutes sets the "duration_minutes" field.
func (msuo *MeetingScheduleUpdateOne) SetDurationMinutes(i int) *MeetingScheduleUpdateOne {
	msuo.mutation.ResetDurationMinutes()
	msuo.mutation.SetDurationMinutes(i)
	return msuo
}

// SetNillableDurationMinutes sets the "duration_minutes" field if the given value is not nil.
func (msuo *MeetingScheduleUpdateOne) SetNillableDurationMinutes(i *int) *MeetingScheduleUpdateOne {
	if i != nil {
		msuo.SetDurationMinutes(*i)
	}
	return msuo
}

// AddDurationMinutes adds i to the "duration_minutes" field.
func (msuo *MeetingScheduleUpdateOne) AddDurationMinutes(i int) *MeetingScheduleUpdateOne {
	msuo.mutation.AddDurationMinutes(i)
	return msuo
}

// SetStartDate sets the "start_date" field.
func (msuo *MeetingScheduleUpdateOne) SetStartDate(t time.Time) *MeetingScheduleUpdateOne {
	msuo.mutation.SetStartDate(t)
	return msuo
}

// SetNillableStartDate sets the "start_date" field if the given value is not nil.
func (msuo *MeetingScheduleUpdateOne) SetNillableStartDate(t *time.Time) *MeetingScheduleUpdateOne {
	if t != nil {
		msuo.SetStartDate(*t)
	}
	return msuo
}

// SetRepeats sets the "repeats" field.
func (msuo *MeetingScheduleUpdateOne) SetRepeats(m meetingschedule.Repeats) *MeetingScheduleUpdateOne {
	msuo.mutation.SetRepeats(m)
	return msuo
}

// SetNillableRepeats sets the "repeats" field if the given value is not nil.
func (msuo *MeetingScheduleUpdateOne) SetNillableRepeats(m *meetingschedule.Repeats) *MeetingScheduleUpdateOne {
	if m != nil {
		msuo.SetRepeats(*m)
	}
	return msuo
}

// SetRepetitionStep sets the "repetition_step" field.
func (msuo *MeetingScheduleUpdateOne) SetRepetitionStep(i int) *MeetingScheduleUpdateOne {
	msuo.mutation.ResetRepetitionStep()
	msuo.mutation.SetRepetitionStep(i)
	return msuo
}

// SetNillableRepetitionStep sets the "repetition_step" field if the given value is not nil.
func (msuo *MeetingScheduleUpdateOne) SetNillableRepetitionStep(i *int) *MeetingScheduleUpdateOne {
	if i != nil {
		msuo.SetRepetitionStep(*i)
	}
	return msuo
}

// AddRepetitionStep adds i to the "repetition_step" field.
func (msuo *MeetingScheduleUpdateOne) AddRepetitionStep(i int) *MeetingScheduleUpdateOne {
	msuo.mutation.AddRepetitionStep(i)
	return msuo
}

// SetWeekDays sets the "week_days" field.
func (msuo *MeetingScheduleUpdateOne) SetWeekDays(s []string) *MeetingScheduleUpdateOne {
	msuo.mutation.SetWeekDays(s)
	return msuo
}

// AppendWeekDays appends s to the "week_days" field.
func (msuo *MeetingScheduleUpdateOne) AppendWeekDays(s []string) *MeetingScheduleUpdateOne {
	msuo.mutation.AppendWeekDays(s)
	return msuo
}

// ClearWeekDays clears the value of the "week_days" field.
func (msuo *MeetingScheduleUpdateOne) ClearWeekDays() *MeetingScheduleUpdateOne {
	msuo.mutation.ClearWeekDays()
	return msuo
}

// SetMonthlyOn sets the "monthly_on" field.
func (msuo *MeetingScheduleUpdateOne) SetMonthlyOn(mo meetingschedule.MonthlyOn) *MeetingScheduleUpdateOne {
	msuo.mutation.SetMonthlyOn(mo)
	return msuo
}

// SetNillableMonthlyOn sets the "monthly_on" field if the given value is not nil.
func (msuo *MeetingScheduleUpdateOne) SetNillableMonthlyOn(mo *meetingschedule.MonthlyOn) *MeetingScheduleUpdateOne {
	if mo != nil {
		msuo.SetMonthlyOn(*mo)
	}
	return msuo
}

// ClearMonthlyOn clears the value of the "monthly_on" field.
func (msuo *MeetingScheduleUpdateOne) ClearMonthlyOn() *MeetingScheduleUpdateOne {
	msuo.mutation.ClearMonthlyOn()
	return msuo
}

// SetUntilDate sets the "until_date" field.
func (msuo *MeetingScheduleUpdateOne) SetUntilDate(t time.Time) *MeetingScheduleUpdateOne {
	msuo.mutation.SetUntilDate(t)
	return msuo
}

// SetNillableUntilDate sets the "until_date" field if the given value is not nil.
func (msuo *MeetingScheduleUpdateOne) SetNillableUntilDate(t *time.Time) *MeetingScheduleUpdateOne {
	if t != nil {
		msuo.SetUntilDate(*t)
	}
	return msuo
}

// ClearUntilDate clears the value of the "until_date" field.
func (msuo *MeetingScheduleUpdateOne) ClearUntilDate() *MeetingScheduleUpdateOne {
	msuo.mutation.ClearUntilDate()
	return msuo
}

// SetNumRepetitions sets the "num_repetitions" field.
func (msuo *MeetingScheduleUpdateOne) SetNumRepetitions(i int) *MeetingScheduleUpdateOne {
	msuo.mutation.ResetNumRepetitions()
	msuo.mutation.SetNumRepetitions(i)
	return msuo
}

// SetNillableNumRepetitions sets the "num_repetitions" field if the given value is not nil.
func (msuo *MeetingScheduleUpdateOne) SetNillableNumRepetitions(i *int) *MeetingScheduleUpdateOne {
	if i != nil {
		msuo.SetNumRepetitions(*i)
	}
	return msuo
}

// AddNumRepetitions adds i to the "num_repetitions" field.
func (msuo *MeetingScheduleUpdateOne) AddNumRepetitions(i int) *MeetingScheduleUpdateOne {
	msuo.mutation.AddNumRepetitions(i)
	return msuo
}

// ClearNumRepetitions clears the value of the "num_repetitions" field.
func (msuo *MeetingScheduleUpdateOne) ClearNumRepetitions() *MeetingScheduleUpdateOne {
	msuo.mutation.ClearNumRepetitions()
	return msuo
}

// AddSessionIDs adds the "sessions" edge to the MeetingSession entity by IDs.
func (msuo *MeetingScheduleUpdateOne) AddSessionIDs(ids ...uuid.UUID) *MeetingScheduleUpdateOne {
	msuo.mutation.AddSessionIDs(ids...)
	return msuo
}

// AddSessions adds the "sessions" edges to the MeetingSession entity.
func (msuo *MeetingScheduleUpdateOne) AddSessions(m ...*MeetingSession) *MeetingScheduleUpdateOne {
	ids := make([]uuid.UUID, len(m))
	for i := range m {
		ids[i] = m[i].ID
	}
	return msuo.AddSessionIDs(ids...)
}

// AddOwningTeamIDs adds the "owning_team" edge to the Team entity by IDs.
func (msuo *MeetingScheduleUpdateOne) AddOwningTeamIDs(ids ...uuid.UUID) *MeetingScheduleUpdateOne {
	msuo.mutation.AddOwningTeamIDs(ids...)
	return msuo
}

// AddOwningTeam adds the "owning_team" edges to the Team entity.
func (msuo *MeetingScheduleUpdateOne) AddOwningTeam(t ...*Team) *MeetingScheduleUpdateOne {
	ids := make([]uuid.UUID, len(t))
	for i := range t {
		ids[i] = t[i].ID
	}
	return msuo.AddOwningTeamIDs(ids...)
}

// Mutation returns the MeetingScheduleMutation object of the builder.
func (msuo *MeetingScheduleUpdateOne) Mutation() *MeetingScheduleMutation {
	return msuo.mutation
}

// ClearSessions clears all "sessions" edges to the MeetingSession entity.
func (msuo *MeetingScheduleUpdateOne) ClearSessions() *MeetingScheduleUpdateOne {
	msuo.mutation.ClearSessions()
	return msuo
}

// RemoveSessionIDs removes the "sessions" edge to MeetingSession entities by IDs.
func (msuo *MeetingScheduleUpdateOne) RemoveSessionIDs(ids ...uuid.UUID) *MeetingScheduleUpdateOne {
	msuo.mutation.RemoveSessionIDs(ids...)
	return msuo
}

// RemoveSessions removes "sessions" edges to MeetingSession entities.
func (msuo *MeetingScheduleUpdateOne) RemoveSessions(m ...*MeetingSession) *MeetingScheduleUpdateOne {
	ids := make([]uuid.UUID, len(m))
	for i := range m {
		ids[i] = m[i].ID
	}
	return msuo.RemoveSessionIDs(ids...)
}

// ClearOwningTeam clears all "owning_team" edges to the Team entity.
func (msuo *MeetingScheduleUpdateOne) ClearOwningTeam() *MeetingScheduleUpdateOne {
	msuo.mutation.ClearOwningTeam()
	return msuo
}

// RemoveOwningTeamIDs removes the "owning_team" edge to Team entities by IDs.
func (msuo *MeetingScheduleUpdateOne) RemoveOwningTeamIDs(ids ...uuid.UUID) *MeetingScheduleUpdateOne {
	msuo.mutation.RemoveOwningTeamIDs(ids...)
	return msuo
}

// RemoveOwningTeam removes "owning_team" edges to Team entities.
func (msuo *MeetingScheduleUpdateOne) RemoveOwningTeam(t ...*Team) *MeetingScheduleUpdateOne {
	ids := make([]uuid.UUID, len(t))
	for i := range t {
		ids[i] = t[i].ID
	}
	return msuo.RemoveOwningTeamIDs(ids...)
}

// Where appends a list predicates to the MeetingScheduleUpdate builder.
func (msuo *MeetingScheduleUpdateOne) Where(ps ...predicate.MeetingSchedule) *MeetingScheduleUpdateOne {
	msuo.mutation.Where(ps...)
	return msuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (msuo *MeetingScheduleUpdateOne) Select(field string, fields ...string) *MeetingScheduleUpdateOne {
	msuo.fields = append([]string{field}, fields...)
	return msuo
}

// Save executes the query and returns the updated MeetingSchedule entity.
func (msuo *MeetingScheduleUpdateOne) Save(ctx context.Context) (*MeetingSchedule, error) {
	return withHooks(ctx, msuo.sqlSave, msuo.mutation, msuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (msuo *MeetingScheduleUpdateOne) SaveX(ctx context.Context) *MeetingSchedule {
	node, err := msuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (msuo *MeetingScheduleUpdateOne) Exec(ctx context.Context) error {
	_, err := msuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (msuo *MeetingScheduleUpdateOne) ExecX(ctx context.Context) {
	if err := msuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (msuo *MeetingScheduleUpdateOne) check() error {
	if v, ok := msuo.mutation.Repeats(); ok {
		if err := meetingschedule.RepeatsValidator(v); err != nil {
			return &ValidationError{Name: "repeats", err: fmt.Errorf(`ent: validator failed for field "MeetingSchedule.repeats": %w`, err)}
		}
	}
	if v, ok := msuo.mutation.WeekDays(); ok {
		if err := meetingschedule.WeekDaysValidator(v); err != nil {
			return &ValidationError{Name: "week_days", err: fmt.Errorf(`ent: validator failed for field "MeetingSchedule.week_days": %w`, err)}
		}
	}
	if v, ok := msuo.mutation.MonthlyOn(); ok {
		if err := meetingschedule.MonthlyOnValidator(v); err != nil {
			return &ValidationError{Name: "monthly_on", err: fmt.Errorf(`ent: validator failed for field "MeetingSchedule.monthly_on": %w`, err)}
		}
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (msuo *MeetingScheduleUpdateOne) Modify(modifiers ...func(u *sql.UpdateBuilder)) *MeetingScheduleUpdateOne {
	msuo.modifiers = append(msuo.modifiers, modifiers...)
	return msuo
}

func (msuo *MeetingScheduleUpdateOne) sqlSave(ctx context.Context) (_node *MeetingSchedule, err error) {
	if err := msuo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(meetingschedule.Table, meetingschedule.Columns, sqlgraph.NewFieldSpec(meetingschedule.FieldID, field.TypeUUID))
	id, ok := msuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "MeetingSchedule.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := msuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, meetingschedule.FieldID)
		for _, f := range fields {
			if !meetingschedule.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != meetingschedule.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := msuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := msuo.mutation.Name(); ok {
		_spec.SetField(meetingschedule.FieldName, field.TypeString, value)
	}
	if value, ok := msuo.mutation.Description(); ok {
		_spec.SetField(meetingschedule.FieldDescription, field.TypeString, value)
	}
	if msuo.mutation.DescriptionCleared() {
		_spec.ClearField(meetingschedule.FieldDescription, field.TypeString)
	}
	if value, ok := msuo.mutation.BeginMinute(); ok {
		_spec.SetField(meetingschedule.FieldBeginMinute, field.TypeInt, value)
	}
	if value, ok := msuo.mutation.AddedBeginMinute(); ok {
		_spec.AddField(meetingschedule.FieldBeginMinute, field.TypeInt, value)
	}
	if value, ok := msuo.mutation.DurationMinutes(); ok {
		_spec.SetField(meetingschedule.FieldDurationMinutes, field.TypeInt, value)
	}
	if value, ok := msuo.mutation.AddedDurationMinutes(); ok {
		_spec.AddField(meetingschedule.FieldDurationMinutes, field.TypeInt, value)
	}
	if value, ok := msuo.mutation.StartDate(); ok {
		_spec.SetField(meetingschedule.FieldStartDate, field.TypeTime, value)
	}
	if value, ok := msuo.mutation.Repeats(); ok {
		_spec.SetField(meetingschedule.FieldRepeats, field.TypeEnum, value)
	}
	if value, ok := msuo.mutation.RepetitionStep(); ok {
		_spec.SetField(meetingschedule.FieldRepetitionStep, field.TypeInt, value)
	}
	if value, ok := msuo.mutation.AddedRepetitionStep(); ok {
		_spec.AddField(meetingschedule.FieldRepetitionStep, field.TypeInt, value)
	}
	if value, ok := msuo.mutation.WeekDays(); ok {
		_spec.SetField(meetingschedule.FieldWeekDays, field.TypeJSON, value)
	}
	if value, ok := msuo.mutation.AppendedWeekDays(); ok {
		_spec.AddModifier(func(u *sql.UpdateBuilder) {
			sqljson.Append(u, meetingschedule.FieldWeekDays, value)
		})
	}
	if msuo.mutation.WeekDaysCleared() {
		_spec.ClearField(meetingschedule.FieldWeekDays, field.TypeJSON)
	}
	if value, ok := msuo.mutation.MonthlyOn(); ok {
		_spec.SetField(meetingschedule.FieldMonthlyOn, field.TypeEnum, value)
	}
	if msuo.mutation.MonthlyOnCleared() {
		_spec.ClearField(meetingschedule.FieldMonthlyOn, field.TypeEnum)
	}
	if value, ok := msuo.mutation.UntilDate(); ok {
		_spec.SetField(meetingschedule.FieldUntilDate, field.TypeTime, value)
	}
	if msuo.mutation.UntilDateCleared() {
		_spec.ClearField(meetingschedule.FieldUntilDate, field.TypeTime)
	}
	if value, ok := msuo.mutation.NumRepetitions(); ok {
		_spec.SetField(meetingschedule.FieldNumRepetitions, field.TypeInt, value)
	}
	if value, ok := msuo.mutation.AddedNumRepetitions(); ok {
		_spec.AddField(meetingschedule.FieldNumRepetitions, field.TypeInt, value)
	}
	if msuo.mutation.NumRepetitionsCleared() {
		_spec.ClearField(meetingschedule.FieldNumRepetitions, field.TypeInt)
	}
	if msuo.mutation.SessionsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   meetingschedule.SessionsTable,
			Columns: []string{meetingschedule.SessionsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(meetingsession.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := msuo.mutation.RemovedSessionsIDs(); len(nodes) > 0 && !msuo.mutation.SessionsCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   meetingschedule.SessionsTable,
			Columns: []string{meetingschedule.SessionsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(meetingsession.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := msuo.mutation.SessionsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   meetingschedule.SessionsTable,
			Columns: []string{meetingschedule.SessionsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(meetingsession.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if msuo.mutation.OwningTeamCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   meetingschedule.OwningTeamTable,
			Columns: meetingschedule.OwningTeamPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(team.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := msuo.mutation.RemovedOwningTeamIDs(); len(nodes) > 0 && !msuo.mutation.OwningTeamCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   meetingschedule.OwningTeamTable,
			Columns: meetingschedule.OwningTeamPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(team.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := msuo.mutation.OwningTeamIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   meetingschedule.OwningTeamTable,
			Columns: meetingschedule.OwningTeamPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(team.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(msuo.modifiers...)
	_node = &MeetingSchedule{config: msuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, msuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{meetingschedule.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	msuo.mutation.done = true
	return _node, nil
}
