// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/twohundreds/rezible/ent/incidenttag"
)

// IncidentTag is the model entity for the IncidentTag schema.
type IncidentTag struct {
	config `json:"-"`
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// ArchiveTime holds the value of the "archive_time" field.
	ArchiveTime time.Time `json:"archive_time,omitempty"`
	// Key holds the value of the "key" field.
	Key string `json:"key,omitempty"`
	// Value holds the value of the "value" field.
	Value string `json:"value,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the IncidentTagQuery when eager-loading is set.
	Edges        IncidentTagEdges `json:"edges"`
	selectValues sql.SelectValues
}

// IncidentTagEdges holds the relations/edges for other nodes in the graph.
type IncidentTagEdges struct {
	// Incidents holds the value of the incidents edge.
	Incidents []*Incident `json:"incidents,omitempty"`
	// DebriefQuestions holds the value of the debrief_questions edge.
	DebriefQuestions []*IncidentDebriefQuestion `json:"debrief_questions,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [2]bool
}

// IncidentsOrErr returns the Incidents value or an error if the edge
// was not loaded in eager-loading.
func (e IncidentTagEdges) IncidentsOrErr() ([]*Incident, error) {
	if e.loadedTypes[0] {
		return e.Incidents, nil
	}
	return nil, &NotLoadedError{edge: "incidents"}
}

// DebriefQuestionsOrErr returns the DebriefQuestions value or an error if the edge
// was not loaded in eager-loading.
func (e IncidentTagEdges) DebriefQuestionsOrErr() ([]*IncidentDebriefQuestion, error) {
	if e.loadedTypes[1] {
		return e.DebriefQuestions, nil
	}
	return nil, &NotLoadedError{edge: "debrief_questions"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*IncidentTag) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case incidenttag.FieldKey, incidenttag.FieldValue:
			values[i] = new(sql.NullString)
		case incidenttag.FieldArchiveTime:
			values[i] = new(sql.NullTime)
		case incidenttag.FieldID:
			values[i] = new(uuid.UUID)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the IncidentTag fields.
func (it *IncidentTag) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case incidenttag.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				it.ID = *value
			}
		case incidenttag.FieldArchiveTime:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field archive_time", values[i])
			} else if value.Valid {
				it.ArchiveTime = value.Time
			}
		case incidenttag.FieldKey:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field key", values[i])
			} else if value.Valid {
				it.Key = value.String
			}
		case incidenttag.FieldValue:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field value", values[i])
			} else if value.Valid {
				it.Value = value.String
			}
		default:
			it.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// GetValue returns the ent.Value that was dynamically selected and assigned to the IncidentTag.
// This includes values selected through modifiers, order, etc.
func (it *IncidentTag) GetValue(name string) (ent.Value, error) {
	return it.selectValues.Get(name)
}

// QueryIncidents queries the "incidents" edge of the IncidentTag entity.
func (it *IncidentTag) QueryIncidents() *IncidentQuery {
	return NewIncidentTagClient(it.config).QueryIncidents(it)
}

// QueryDebriefQuestions queries the "debrief_questions" edge of the IncidentTag entity.
func (it *IncidentTag) QueryDebriefQuestions() *IncidentDebriefQuestionQuery {
	return NewIncidentTagClient(it.config).QueryDebriefQuestions(it)
}

// Update returns a builder for updating this IncidentTag.
// Note that you need to call IncidentTag.Unwrap() before calling this method if this IncidentTag
// was returned from a transaction, and the transaction was committed or rolled back.
func (it *IncidentTag) Update() *IncidentTagUpdateOne {
	return NewIncidentTagClient(it.config).UpdateOne(it)
}

// Unwrap unwraps the IncidentTag entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (it *IncidentTag) Unwrap() *IncidentTag {
	_tx, ok := it.config.driver.(*txDriver)
	if !ok {
		panic("ent: IncidentTag is not a transactional entity")
	}
	it.config.driver = _tx.drv
	return it
}

// String implements the fmt.Stringer.
func (it *IncidentTag) String() string {
	var builder strings.Builder
	builder.WriteString("IncidentTag(")
	builder.WriteString(fmt.Sprintf("id=%v, ", it.ID))
	builder.WriteString("archive_time=")
	builder.WriteString(it.ArchiveTime.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("key=")
	builder.WriteString(it.Key)
	builder.WriteString(", ")
	builder.WriteString("value=")
	builder.WriteString(it.Value)
	builder.WriteByte(')')
	return builder.String()
}

// IncidentTags is a parsable slice of IncidentTag.
type IncidentTags []*IncidentTag