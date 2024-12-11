// Code generated by ent, DO NOT EDIT.

package ent

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/twohundreds/rezible/ent/meetingschedule"
)

// MeetingSchedule is the model entity for the MeetingSchedule schema.
type MeetingSchedule struct {
	config `json:"-"`
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// Name holds the value of the "name" field.
	Name string `json:"name,omitempty"`
	// Description holds the value of the "description" field.
	Description string `json:"description,omitempty"`
	// BeginMinute holds the value of the "begin_minute" field.
	BeginMinute int `json:"begin_minute,omitempty"`
	// DurationMinutes holds the value of the "duration_minutes" field.
	DurationMinutes int `json:"duration_minutes,omitempty"`
	// StartDate holds the value of the "start_date" field.
	StartDate time.Time `json:"start_date,omitempty"`
	// Repeats holds the value of the "repeats" field.
	Repeats meetingschedule.Repeats `json:"repeats,omitempty"`
	// RepetitionStep holds the value of the "repetition_step" field.
	RepetitionStep int `json:"repetition_step,omitempty"`
	// WeekDays holds the value of the "week_days" field.
	WeekDays []string `json:"week_days,omitempty"`
	// MonthlyOn holds the value of the "monthly_on" field.
	MonthlyOn meetingschedule.MonthlyOn `json:"monthly_on,omitempty"`
	// UntilDate holds the value of the "until_date" field.
	UntilDate time.Time `json:"until_date,omitempty"`
	// NumRepetitions holds the value of the "num_repetitions" field.
	NumRepetitions int `json:"num_repetitions,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the MeetingScheduleQuery when eager-loading is set.
	Edges        MeetingScheduleEdges `json:"edges"`
	selectValues sql.SelectValues
}

// MeetingScheduleEdges holds the relations/edges for other nodes in the graph.
type MeetingScheduleEdges struct {
	// Sessions holds the value of the sessions edge.
	Sessions []*MeetingSession `json:"sessions,omitempty"`
	// OwningTeam holds the value of the owning_team edge.
	OwningTeam []*Team `json:"owning_team,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [2]bool
}

// SessionsOrErr returns the Sessions value or an error if the edge
// was not loaded in eager-loading.
func (e MeetingScheduleEdges) SessionsOrErr() ([]*MeetingSession, error) {
	if e.loadedTypes[0] {
		return e.Sessions, nil
	}
	return nil, &NotLoadedError{edge: "sessions"}
}

// OwningTeamOrErr returns the OwningTeam value or an error if the edge
// was not loaded in eager-loading.
func (e MeetingScheduleEdges) OwningTeamOrErr() ([]*Team, error) {
	if e.loadedTypes[1] {
		return e.OwningTeam, nil
	}
	return nil, &NotLoadedError{edge: "owning_team"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*MeetingSchedule) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case meetingschedule.FieldWeekDays:
			values[i] = new([]byte)
		case meetingschedule.FieldBeginMinute, meetingschedule.FieldDurationMinutes, meetingschedule.FieldRepetitionStep, meetingschedule.FieldNumRepetitions:
			values[i] = new(sql.NullInt64)
		case meetingschedule.FieldName, meetingschedule.FieldDescription, meetingschedule.FieldRepeats, meetingschedule.FieldMonthlyOn:
			values[i] = new(sql.NullString)
		case meetingschedule.FieldStartDate, meetingschedule.FieldUntilDate:
			values[i] = new(sql.NullTime)
		case meetingschedule.FieldID:
			values[i] = new(uuid.UUID)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the MeetingSchedule fields.
func (ms *MeetingSchedule) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case meetingschedule.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				ms.ID = *value
			}
		case meetingschedule.FieldName:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field name", values[i])
			} else if value.Valid {
				ms.Name = value.String
			}
		case meetingschedule.FieldDescription:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field description", values[i])
			} else if value.Valid {
				ms.Description = value.String
			}
		case meetingschedule.FieldBeginMinute:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field begin_minute", values[i])
			} else if value.Valid {
				ms.BeginMinute = int(value.Int64)
			}
		case meetingschedule.FieldDurationMinutes:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field duration_minutes", values[i])
			} else if value.Valid {
				ms.DurationMinutes = int(value.Int64)
			}
		case meetingschedule.FieldStartDate:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field start_date", values[i])
			} else if value.Valid {
				ms.StartDate = value.Time
			}
		case meetingschedule.FieldRepeats:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field repeats", values[i])
			} else if value.Valid {
				ms.Repeats = meetingschedule.Repeats(value.String)
			}
		case meetingschedule.FieldRepetitionStep:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field repetition_step", values[i])
			} else if value.Valid {
				ms.RepetitionStep = int(value.Int64)
			}
		case meetingschedule.FieldWeekDays:
			if value, ok := values[i].(*[]byte); !ok {
				return fmt.Errorf("unexpected type %T for field week_days", values[i])
			} else if value != nil && len(*value) > 0 {
				if err := json.Unmarshal(*value, &ms.WeekDays); err != nil {
					return fmt.Errorf("unmarshal field week_days: %w", err)
				}
			}
		case meetingschedule.FieldMonthlyOn:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field monthly_on", values[i])
			} else if value.Valid {
				ms.MonthlyOn = meetingschedule.MonthlyOn(value.String)
			}
		case meetingschedule.FieldUntilDate:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field until_date", values[i])
			} else if value.Valid {
				ms.UntilDate = value.Time
			}
		case meetingschedule.FieldNumRepetitions:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field num_repetitions", values[i])
			} else if value.Valid {
				ms.NumRepetitions = int(value.Int64)
			}
		default:
			ms.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the MeetingSchedule.
// This includes values selected through modifiers, order, etc.
func (ms *MeetingSchedule) Value(name string) (ent.Value, error) {
	return ms.selectValues.Get(name)
}

// QuerySessions queries the "sessions" edge of the MeetingSchedule entity.
func (ms *MeetingSchedule) QuerySessions() *MeetingSessionQuery {
	return NewMeetingScheduleClient(ms.config).QuerySessions(ms)
}

// QueryOwningTeam queries the "owning_team" edge of the MeetingSchedule entity.
func (ms *MeetingSchedule) QueryOwningTeam() *TeamQuery {
	return NewMeetingScheduleClient(ms.config).QueryOwningTeam(ms)
}

// Update returns a builder for updating this MeetingSchedule.
// Note that you need to call MeetingSchedule.Unwrap() before calling this method if this MeetingSchedule
// was returned from a transaction, and the transaction was committed or rolled back.
func (ms *MeetingSchedule) Update() *MeetingScheduleUpdateOne {
	return NewMeetingScheduleClient(ms.config).UpdateOne(ms)
}

// Unwrap unwraps the MeetingSchedule entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (ms *MeetingSchedule) Unwrap() *MeetingSchedule {
	_tx, ok := ms.config.driver.(*txDriver)
	if !ok {
		panic("ent: MeetingSchedule is not a transactional entity")
	}
	ms.config.driver = _tx.drv
	return ms
}

// String implements the fmt.Stringer.
func (ms *MeetingSchedule) String() string {
	var builder strings.Builder
	builder.WriteString("MeetingSchedule(")
	builder.WriteString(fmt.Sprintf("id=%v, ", ms.ID))
	builder.WriteString("name=")
	builder.WriteString(ms.Name)
	builder.WriteString(", ")
	builder.WriteString("description=")
	builder.WriteString(ms.Description)
	builder.WriteString(", ")
	builder.WriteString("begin_minute=")
	builder.WriteString(fmt.Sprintf("%v", ms.BeginMinute))
	builder.WriteString(", ")
	builder.WriteString("duration_minutes=")
	builder.WriteString(fmt.Sprintf("%v", ms.DurationMinutes))
	builder.WriteString(", ")
	builder.WriteString("start_date=")
	builder.WriteString(ms.StartDate.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("repeats=")
	builder.WriteString(fmt.Sprintf("%v", ms.Repeats))
	builder.WriteString(", ")
	builder.WriteString("repetition_step=")
	builder.WriteString(fmt.Sprintf("%v", ms.RepetitionStep))
	builder.WriteString(", ")
	builder.WriteString("week_days=")
	builder.WriteString(fmt.Sprintf("%v", ms.WeekDays))
	builder.WriteString(", ")
	builder.WriteString("monthly_on=")
	builder.WriteString(fmt.Sprintf("%v", ms.MonthlyOn))
	builder.WriteString(", ")
	builder.WriteString("until_date=")
	builder.WriteString(ms.UntilDate.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("num_repetitions=")
	builder.WriteString(fmt.Sprintf("%v", ms.NumRepetitions))
	builder.WriteByte(')')
	return builder.String()
}

// MeetingSchedules is a parsable slice of MeetingSchedule.
type MeetingSchedules []*MeetingSchedule
