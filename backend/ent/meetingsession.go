// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/twohundreds/rezible/ent/meetingsession"
)

// MeetingSession is the model entity for the MeetingSession schema.
type MeetingSession struct {
	config `json:"-"`
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// Title holds the value of the "title" field.
	Title string `json:"title,omitempty"`
	// StartedAt holds the value of the "started_at" field.
	StartedAt time.Time `json:"started_at,omitempty"`
	// EndedAt holds the value of the "ended_at" field.
	EndedAt time.Time `json:"ended_at,omitempty"`
	// DocumentName holds the value of the "document_name" field.
	DocumentName string `json:"document_name,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the MeetingSessionQuery when eager-loading is set.
	Edges                     MeetingSessionEdges `json:"edges"`
	meeting_schedule_sessions *uuid.UUID
	selectValues              sql.SelectValues
}

// MeetingSessionEdges holds the relations/edges for other nodes in the graph.
type MeetingSessionEdges struct {
	// Incidents holds the value of the incidents edge.
	Incidents []*Incident `json:"incidents,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [1]bool
}

// IncidentsOrErr returns the Incidents value or an error if the edge
// was not loaded in eager-loading.
func (e MeetingSessionEdges) IncidentsOrErr() ([]*Incident, error) {
	if e.loadedTypes[0] {
		return e.Incidents, nil
	}
	return nil, &NotLoadedError{edge: "incidents"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*MeetingSession) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case meetingsession.FieldTitle, meetingsession.FieldDocumentName:
			values[i] = new(sql.NullString)
		case meetingsession.FieldStartedAt, meetingsession.FieldEndedAt:
			values[i] = new(sql.NullTime)
		case meetingsession.FieldID:
			values[i] = new(uuid.UUID)
		case meetingsession.ForeignKeys[0]: // meeting_schedule_sessions
			values[i] = &sql.NullScanner{S: new(uuid.UUID)}
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the MeetingSession fields.
func (ms *MeetingSession) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case meetingsession.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				ms.ID = *value
			}
		case meetingsession.FieldTitle:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field title", values[i])
			} else if value.Valid {
				ms.Title = value.String
			}
		case meetingsession.FieldStartedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field started_at", values[i])
			} else if value.Valid {
				ms.StartedAt = value.Time
			}
		case meetingsession.FieldEndedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field ended_at", values[i])
			} else if value.Valid {
				ms.EndedAt = value.Time
			}
		case meetingsession.FieldDocumentName:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field document_name", values[i])
			} else if value.Valid {
				ms.DocumentName = value.String
			}
		case meetingsession.ForeignKeys[0]:
			if value, ok := values[i].(*sql.NullScanner); !ok {
				return fmt.Errorf("unexpected type %T for field meeting_schedule_sessions", values[i])
			} else if value.Valid {
				ms.meeting_schedule_sessions = new(uuid.UUID)
				*ms.meeting_schedule_sessions = *value.S.(*uuid.UUID)
			}
		default:
			ms.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the MeetingSession.
// This includes values selected through modifiers, order, etc.
func (ms *MeetingSession) Value(name string) (ent.Value, error) {
	return ms.selectValues.Get(name)
}

// QueryIncidents queries the "incidents" edge of the MeetingSession entity.
func (ms *MeetingSession) QueryIncidents() *IncidentQuery {
	return NewMeetingSessionClient(ms.config).QueryIncidents(ms)
}

// Update returns a builder for updating this MeetingSession.
// Note that you need to call MeetingSession.Unwrap() before calling this method if this MeetingSession
// was returned from a transaction, and the transaction was committed or rolled back.
func (ms *MeetingSession) Update() *MeetingSessionUpdateOne {
	return NewMeetingSessionClient(ms.config).UpdateOne(ms)
}

// Unwrap unwraps the MeetingSession entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (ms *MeetingSession) Unwrap() *MeetingSession {
	_tx, ok := ms.config.driver.(*txDriver)
	if !ok {
		panic("ent: MeetingSession is not a transactional entity")
	}
	ms.config.driver = _tx.drv
	return ms
}

// String implements the fmt.Stringer.
func (ms *MeetingSession) String() string {
	var builder strings.Builder
	builder.WriteString("MeetingSession(")
	builder.WriteString(fmt.Sprintf("id=%v, ", ms.ID))
	builder.WriteString("title=")
	builder.WriteString(ms.Title)
	builder.WriteString(", ")
	builder.WriteString("started_at=")
	builder.WriteString(ms.StartedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("ended_at=")
	builder.WriteString(ms.EndedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("document_name=")
	builder.WriteString(ms.DocumentName)
	builder.WriteByte(')')
	return builder.String()
}

// MeetingSessions is a parsable slice of MeetingSession.
type MeetingSessions []*MeetingSession
