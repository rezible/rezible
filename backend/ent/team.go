// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/twohundreds/rezible/ent/team"
)

// Team is the model entity for the Team schema.
type Team struct {
	config `json:"-"`
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// Slug holds the value of the "slug" field.
	Slug string `json:"slug,omitempty"`
	// Name holds the value of the "name" field.
	Name string `json:"name,omitempty"`
	// ChatChannelID holds the value of the "chat_channel_id" field.
	ChatChannelID string `json:"chat_channel_id,omitempty"`
	// Timezone holds the value of the "timezone" field.
	Timezone string `json:"timezone,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the TeamQuery when eager-loading is set.
	Edges        TeamEdges `json:"edges"`
	selectValues sql.SelectValues
}

// TeamEdges holds the relations/edges for other nodes in the graph.
type TeamEdges struct {
	// Users holds the value of the users edge.
	Users []*User `json:"users,omitempty"`
	// Services holds the value of the services edge.
	Services []*Service `json:"services,omitempty"`
	// OncallRosters holds the value of the oncall_rosters edge.
	OncallRosters []*OncallRoster `json:"oncall_rosters,omitempty"`
	// Subscriptions holds the value of the subscriptions edge.
	Subscriptions []*Subscription `json:"subscriptions,omitempty"`
	// IncidentAssignments holds the value of the incident_assignments edge.
	IncidentAssignments []*IncidentTeamAssignment `json:"incident_assignments,omitempty"`
	// ScheduledMeetings holds the value of the scheduled_meetings edge.
	ScheduledMeetings []*MeetingSchedule `json:"scheduled_meetings,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [6]bool
}

// UsersOrErr returns the Users value or an error if the edge
// was not loaded in eager-loading.
func (e TeamEdges) UsersOrErr() ([]*User, error) {
	if e.loadedTypes[0] {
		return e.Users, nil
	}
	return nil, &NotLoadedError{edge: "users"}
}

// ServicesOrErr returns the Services value or an error if the edge
// was not loaded in eager-loading.
func (e TeamEdges) ServicesOrErr() ([]*Service, error) {
	if e.loadedTypes[1] {
		return e.Services, nil
	}
	return nil, &NotLoadedError{edge: "services"}
}

// OncallRostersOrErr returns the OncallRosters value or an error if the edge
// was not loaded in eager-loading.
func (e TeamEdges) OncallRostersOrErr() ([]*OncallRoster, error) {
	if e.loadedTypes[2] {
		return e.OncallRosters, nil
	}
	return nil, &NotLoadedError{edge: "oncall_rosters"}
}

// SubscriptionsOrErr returns the Subscriptions value or an error if the edge
// was not loaded in eager-loading.
func (e TeamEdges) SubscriptionsOrErr() ([]*Subscription, error) {
	if e.loadedTypes[3] {
		return e.Subscriptions, nil
	}
	return nil, &NotLoadedError{edge: "subscriptions"}
}

// IncidentAssignmentsOrErr returns the IncidentAssignments value or an error if the edge
// was not loaded in eager-loading.
func (e TeamEdges) IncidentAssignmentsOrErr() ([]*IncidentTeamAssignment, error) {
	if e.loadedTypes[4] {
		return e.IncidentAssignments, nil
	}
	return nil, &NotLoadedError{edge: "incident_assignments"}
}

// ScheduledMeetingsOrErr returns the ScheduledMeetings value or an error if the edge
// was not loaded in eager-loading.
func (e TeamEdges) ScheduledMeetingsOrErr() ([]*MeetingSchedule, error) {
	if e.loadedTypes[5] {
		return e.ScheduledMeetings, nil
	}
	return nil, &NotLoadedError{edge: "scheduled_meetings"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Team) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case team.FieldSlug, team.FieldName, team.FieldChatChannelID, team.FieldTimezone:
			values[i] = new(sql.NullString)
		case team.FieldID:
			values[i] = new(uuid.UUID)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Team fields.
func (t *Team) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case team.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				t.ID = *value
			}
		case team.FieldSlug:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field slug", values[i])
			} else if value.Valid {
				t.Slug = value.String
			}
		case team.FieldName:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field name", values[i])
			} else if value.Valid {
				t.Name = value.String
			}
		case team.FieldChatChannelID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field chat_channel_id", values[i])
			} else if value.Valid {
				t.ChatChannelID = value.String
			}
		case team.FieldTimezone:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field timezone", values[i])
			} else if value.Valid {
				t.Timezone = value.String
			}
		default:
			t.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the Team.
// This includes values selected through modifiers, order, etc.
func (t *Team) Value(name string) (ent.Value, error) {
	return t.selectValues.Get(name)
}

// QueryUsers queries the "users" edge of the Team entity.
func (t *Team) QueryUsers() *UserQuery {
	return NewTeamClient(t.config).QueryUsers(t)
}

// QueryServices queries the "services" edge of the Team entity.
func (t *Team) QueryServices() *ServiceQuery {
	return NewTeamClient(t.config).QueryServices(t)
}

// QueryOncallRosters queries the "oncall_rosters" edge of the Team entity.
func (t *Team) QueryOncallRosters() *OncallRosterQuery {
	return NewTeamClient(t.config).QueryOncallRosters(t)
}

// QuerySubscriptions queries the "subscriptions" edge of the Team entity.
func (t *Team) QuerySubscriptions() *SubscriptionQuery {
	return NewTeamClient(t.config).QuerySubscriptions(t)
}

// QueryIncidentAssignments queries the "incident_assignments" edge of the Team entity.
func (t *Team) QueryIncidentAssignments() *IncidentTeamAssignmentQuery {
	return NewTeamClient(t.config).QueryIncidentAssignments(t)
}

// QueryScheduledMeetings queries the "scheduled_meetings" edge of the Team entity.
func (t *Team) QueryScheduledMeetings() *MeetingScheduleQuery {
	return NewTeamClient(t.config).QueryScheduledMeetings(t)
}

// Update returns a builder for updating this Team.
// Note that you need to call Team.Unwrap() before calling this method if this Team
// was returned from a transaction, and the transaction was committed or rolled back.
func (t *Team) Update() *TeamUpdateOne {
	return NewTeamClient(t.config).UpdateOne(t)
}

// Unwrap unwraps the Team entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (t *Team) Unwrap() *Team {
	_tx, ok := t.config.driver.(*txDriver)
	if !ok {
		panic("ent: Team is not a transactional entity")
	}
	t.config.driver = _tx.drv
	return t
}

// String implements the fmt.Stringer.
func (t *Team) String() string {
	var builder strings.Builder
	builder.WriteString("Team(")
	builder.WriteString(fmt.Sprintf("id=%v, ", t.ID))
	builder.WriteString("slug=")
	builder.WriteString(t.Slug)
	builder.WriteString(", ")
	builder.WriteString("name=")
	builder.WriteString(t.Name)
	builder.WriteString(", ")
	builder.WriteString("chat_channel_id=")
	builder.WriteString(t.ChatChannelID)
	builder.WriteString(", ")
	builder.WriteString("timezone=")
	builder.WriteString(t.Timezone)
	builder.WriteByte(')')
	return builder.String()
}

// Teams is a parsable slice of Team.
type Teams []*Team
