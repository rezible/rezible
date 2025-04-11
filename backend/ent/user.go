// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/user"
)

// User is the model entity for the User schema.
type User struct {
	config `json:"-"`
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// Name holds the value of the "name" field.
	Name string `json:"name,omitempty"`
	// Email holds the value of the "email" field.
	Email string `json:"email,omitempty"`
	// ChatID holds the value of the "chat_id" field.
	ChatID string `json:"chat_id,omitempty"`
	// Timezone holds the value of the "timezone" field.
	Timezone string `json:"timezone,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the UserQuery when eager-loading is set.
	Edges        UserEdges `json:"edges"`
	selectValues sql.SelectValues
}

// UserEdges holds the relations/edges for other nodes in the graph.
type UserEdges struct {
	// Teams holds the value of the teams edge.
	Teams []*Team `json:"teams,omitempty"`
	// OncallSchedules holds the value of the oncall_schedules edge.
	OncallSchedules []*OncallScheduleParticipant `json:"oncall_schedules,omitempty"`
	// OncallShifts holds the value of the oncall_shifts edge.
	OncallShifts []*OncallUserShift `json:"oncall_shifts,omitempty"`
	// OncallShiftCovers holds the value of the oncall_shift_covers edge.
	OncallShiftCovers []*OncallUserShiftCover `json:"oncall_shift_covers,omitempty"`
	// OncallEventAnnotations holds the value of the oncall_event_annotations edge.
	OncallEventAnnotations []*OncallEventAnnotation `json:"oncall_event_annotations,omitempty"`
	// IncidentRoleAssignments holds the value of the incident_role_assignments edge.
	IncidentRoleAssignments []*IncidentRoleAssignment `json:"incident_role_assignments,omitempty"`
	// IncidentDebriefs holds the value of the incident_debriefs edge.
	IncidentDebriefs []*IncidentDebrief `json:"incident_debriefs,omitempty"`
	// AssignedTasks holds the value of the assigned_tasks edge.
	AssignedTasks []*Task `json:"assigned_tasks,omitempty"`
	// CreatedTasks holds the value of the created_tasks edge.
	CreatedTasks []*Task `json:"created_tasks,omitempty"`
	// RetrospectiveReviewRequests holds the value of the retrospective_review_requests edge.
	RetrospectiveReviewRequests []*RetrospectiveReview `json:"retrospective_review_requests,omitempty"`
	// RetrospectiveReviewResponses holds the value of the retrospective_review_responses edge.
	RetrospectiveReviewResponses []*RetrospectiveReview `json:"retrospective_review_responses,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [11]bool
}

// TeamsOrErr returns the Teams value or an error if the edge
// was not loaded in eager-loading.
func (e UserEdges) TeamsOrErr() ([]*Team, error) {
	if e.loadedTypes[0] {
		return e.Teams, nil
	}
	return nil, &NotLoadedError{edge: "teams"}
}

// OncallSchedulesOrErr returns the OncallSchedules value or an error if the edge
// was not loaded in eager-loading.
func (e UserEdges) OncallSchedulesOrErr() ([]*OncallScheduleParticipant, error) {
	if e.loadedTypes[1] {
		return e.OncallSchedules, nil
	}
	return nil, &NotLoadedError{edge: "oncall_schedules"}
}

// OncallShiftsOrErr returns the OncallShifts value or an error if the edge
// was not loaded in eager-loading.
func (e UserEdges) OncallShiftsOrErr() ([]*OncallUserShift, error) {
	if e.loadedTypes[2] {
		return e.OncallShifts, nil
	}
	return nil, &NotLoadedError{edge: "oncall_shifts"}
}

// OncallShiftCoversOrErr returns the OncallShiftCovers value or an error if the edge
// was not loaded in eager-loading.
func (e UserEdges) OncallShiftCoversOrErr() ([]*OncallUserShiftCover, error) {
	if e.loadedTypes[3] {
		return e.OncallShiftCovers, nil
	}
	return nil, &NotLoadedError{edge: "oncall_shift_covers"}
}

// OncallEventAnnotationsOrErr returns the OncallEventAnnotations value or an error if the edge
// was not loaded in eager-loading.
func (e UserEdges) OncallEventAnnotationsOrErr() ([]*OncallEventAnnotation, error) {
	if e.loadedTypes[4] {
		return e.OncallEventAnnotations, nil
	}
	return nil, &NotLoadedError{edge: "oncall_event_annotations"}
}

// IncidentRoleAssignmentsOrErr returns the IncidentRoleAssignments value or an error if the edge
// was not loaded in eager-loading.
func (e UserEdges) IncidentRoleAssignmentsOrErr() ([]*IncidentRoleAssignment, error) {
	if e.loadedTypes[5] {
		return e.IncidentRoleAssignments, nil
	}
	return nil, &NotLoadedError{edge: "incident_role_assignments"}
}

// IncidentDebriefsOrErr returns the IncidentDebriefs value or an error if the edge
// was not loaded in eager-loading.
func (e UserEdges) IncidentDebriefsOrErr() ([]*IncidentDebrief, error) {
	if e.loadedTypes[6] {
		return e.IncidentDebriefs, nil
	}
	return nil, &NotLoadedError{edge: "incident_debriefs"}
}

// AssignedTasksOrErr returns the AssignedTasks value or an error if the edge
// was not loaded in eager-loading.
func (e UserEdges) AssignedTasksOrErr() ([]*Task, error) {
	if e.loadedTypes[7] {
		return e.AssignedTasks, nil
	}
	return nil, &NotLoadedError{edge: "assigned_tasks"}
}

// CreatedTasksOrErr returns the CreatedTasks value or an error if the edge
// was not loaded in eager-loading.
func (e UserEdges) CreatedTasksOrErr() ([]*Task, error) {
	if e.loadedTypes[8] {
		return e.CreatedTasks, nil
	}
	return nil, &NotLoadedError{edge: "created_tasks"}
}

// RetrospectiveReviewRequestsOrErr returns the RetrospectiveReviewRequests value or an error if the edge
// was not loaded in eager-loading.
func (e UserEdges) RetrospectiveReviewRequestsOrErr() ([]*RetrospectiveReview, error) {
	if e.loadedTypes[9] {
		return e.RetrospectiveReviewRequests, nil
	}
	return nil, &NotLoadedError{edge: "retrospective_review_requests"}
}

// RetrospectiveReviewResponsesOrErr returns the RetrospectiveReviewResponses value or an error if the edge
// was not loaded in eager-loading.
func (e UserEdges) RetrospectiveReviewResponsesOrErr() ([]*RetrospectiveReview, error) {
	if e.loadedTypes[10] {
		return e.RetrospectiveReviewResponses, nil
	}
	return nil, &NotLoadedError{edge: "retrospective_review_responses"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*User) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case user.FieldName, user.FieldEmail, user.FieldChatID, user.FieldTimezone:
			values[i] = new(sql.NullString)
		case user.FieldID:
			values[i] = new(uuid.UUID)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the User fields.
func (u *User) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case user.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				u.ID = *value
			}
		case user.FieldName:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field name", values[i])
			} else if value.Valid {
				u.Name = value.String
			}
		case user.FieldEmail:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field email", values[i])
			} else if value.Valid {
				u.Email = value.String
			}
		case user.FieldChatID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field chat_id", values[i])
			} else if value.Valid {
				u.ChatID = value.String
			}
		case user.FieldTimezone:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field timezone", values[i])
			} else if value.Valid {
				u.Timezone = value.String
			}
		default:
			u.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the User.
// This includes values selected through modifiers, order, etc.
func (u *User) Value(name string) (ent.Value, error) {
	return u.selectValues.Get(name)
}

// QueryTeams queries the "teams" edge of the User entity.
func (u *User) QueryTeams() *TeamQuery {
	return NewUserClient(u.config).QueryTeams(u)
}

// QueryOncallSchedules queries the "oncall_schedules" edge of the User entity.
func (u *User) QueryOncallSchedules() *OncallScheduleParticipantQuery {
	return NewUserClient(u.config).QueryOncallSchedules(u)
}

// QueryOncallShifts queries the "oncall_shifts" edge of the User entity.
func (u *User) QueryOncallShifts() *OncallUserShiftQuery {
	return NewUserClient(u.config).QueryOncallShifts(u)
}

// QueryOncallShiftCovers queries the "oncall_shift_covers" edge of the User entity.
func (u *User) QueryOncallShiftCovers() *OncallUserShiftCoverQuery {
	return NewUserClient(u.config).QueryOncallShiftCovers(u)
}

// QueryOncallEventAnnotations queries the "oncall_event_annotations" edge of the User entity.
func (u *User) QueryOncallEventAnnotations() *OncallEventAnnotationQuery {
	return NewUserClient(u.config).QueryOncallEventAnnotations(u)
}

// QueryIncidentRoleAssignments queries the "incident_role_assignments" edge of the User entity.
func (u *User) QueryIncidentRoleAssignments() *IncidentRoleAssignmentQuery {
	return NewUserClient(u.config).QueryIncidentRoleAssignments(u)
}

// QueryIncidentDebriefs queries the "incident_debriefs" edge of the User entity.
func (u *User) QueryIncidentDebriefs() *IncidentDebriefQuery {
	return NewUserClient(u.config).QueryIncidentDebriefs(u)
}

// QueryAssignedTasks queries the "assigned_tasks" edge of the User entity.
func (u *User) QueryAssignedTasks() *TaskQuery {
	return NewUserClient(u.config).QueryAssignedTasks(u)
}

// QueryCreatedTasks queries the "created_tasks" edge of the User entity.
func (u *User) QueryCreatedTasks() *TaskQuery {
	return NewUserClient(u.config).QueryCreatedTasks(u)
}

// QueryRetrospectiveReviewRequests queries the "retrospective_review_requests" edge of the User entity.
func (u *User) QueryRetrospectiveReviewRequests() *RetrospectiveReviewQuery {
	return NewUserClient(u.config).QueryRetrospectiveReviewRequests(u)
}

// QueryRetrospectiveReviewResponses queries the "retrospective_review_responses" edge of the User entity.
func (u *User) QueryRetrospectiveReviewResponses() *RetrospectiveReviewQuery {
	return NewUserClient(u.config).QueryRetrospectiveReviewResponses(u)
}

// Update returns a builder for updating this User.
// Note that you need to call User.Unwrap() before calling this method if this User
// was returned from a transaction, and the transaction was committed or rolled back.
func (u *User) Update() *UserUpdateOne {
	return NewUserClient(u.config).UpdateOne(u)
}

// Unwrap unwraps the User entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (u *User) Unwrap() *User {
	_tx, ok := u.config.driver.(*txDriver)
	if !ok {
		panic("ent: User is not a transactional entity")
	}
	u.config.driver = _tx.drv
	return u
}

// String implements the fmt.Stringer.
func (u *User) String() string {
	var builder strings.Builder
	builder.WriteString("User(")
	builder.WriteString(fmt.Sprintf("id=%v, ", u.ID))
	builder.WriteString("name=")
	builder.WriteString(u.Name)
	builder.WriteString(", ")
	builder.WriteString("email=")
	builder.WriteString(u.Email)
	builder.WriteString(", ")
	builder.WriteString("chat_id=")
	builder.WriteString(u.ChatID)
	builder.WriteString(", ")
	builder.WriteString("timezone=")
	builder.WriteString(u.Timezone)
	builder.WriteByte(')')
	return builder.String()
}

// Users is a parsable slice of User.
type Users []*User
