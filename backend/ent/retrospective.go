// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/retrospective"
)

// Retrospective is the model entity for the Retrospective schema.
type Retrospective struct {
	config `json:"-"`
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// DocumentName holds the value of the "document_name" field.
	DocumentName string `json:"document_name,omitempty"`
	// Type holds the value of the "type" field.
	Type retrospective.Type `json:"type,omitempty"`
	// State holds the value of the "state" field.
	State retrospective.State `json:"state,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the RetrospectiveQuery when eager-loading is set.
	Edges                  RetrospectiveEdges `json:"edges"`
	incident_retrospective *uuid.UUID
	selectValues           sql.SelectValues
}

// RetrospectiveEdges holds the relations/edges for other nodes in the graph.
type RetrospectiveEdges struct {
	// Incident holds the value of the incident edge.
	Incident *Incident `json:"incident,omitempty"`
	// Discussions holds the value of the discussions edge.
	Discussions []*RetrospectiveDiscussion `json:"discussions,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [2]bool
}

// IncidentOrErr returns the Incident value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e RetrospectiveEdges) IncidentOrErr() (*Incident, error) {
	if e.Incident != nil {
		return e.Incident, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: incident.Label}
	}
	return nil, &NotLoadedError{edge: "incident"}
}

// DiscussionsOrErr returns the Discussions value or an error if the edge
// was not loaded in eager-loading.
func (e RetrospectiveEdges) DiscussionsOrErr() ([]*RetrospectiveDiscussion, error) {
	if e.loadedTypes[1] {
		return e.Discussions, nil
	}
	return nil, &NotLoadedError{edge: "discussions"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Retrospective) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case retrospective.FieldDocumentName, retrospective.FieldType, retrospective.FieldState:
			values[i] = new(sql.NullString)
		case retrospective.FieldID:
			values[i] = new(uuid.UUID)
		case retrospective.ForeignKeys[0]: // incident_retrospective
			values[i] = &sql.NullScanner{S: new(uuid.UUID)}
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Retrospective fields.
func (r *Retrospective) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case retrospective.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				r.ID = *value
			}
		case retrospective.FieldDocumentName:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field document_name", values[i])
			} else if value.Valid {
				r.DocumentName = value.String
			}
		case retrospective.FieldType:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field type", values[i])
			} else if value.Valid {
				r.Type = retrospective.Type(value.String)
			}
		case retrospective.FieldState:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field state", values[i])
			} else if value.Valid {
				r.State = retrospective.State(value.String)
			}
		case retrospective.ForeignKeys[0]:
			if value, ok := values[i].(*sql.NullScanner); !ok {
				return fmt.Errorf("unexpected type %T for field incident_retrospective", values[i])
			} else if value.Valid {
				r.incident_retrospective = new(uuid.UUID)
				*r.incident_retrospective = *value.S.(*uuid.UUID)
			}
		default:
			r.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the Retrospective.
// This includes values selected through modifiers, order, etc.
func (r *Retrospective) Value(name string) (ent.Value, error) {
	return r.selectValues.Get(name)
}

// QueryIncident queries the "incident" edge of the Retrospective entity.
func (r *Retrospective) QueryIncident() *IncidentQuery {
	return NewRetrospectiveClient(r.config).QueryIncident(r)
}

// QueryDiscussions queries the "discussions" edge of the Retrospective entity.
func (r *Retrospective) QueryDiscussions() *RetrospectiveDiscussionQuery {
	return NewRetrospectiveClient(r.config).QueryDiscussions(r)
}

// Update returns a builder for updating this Retrospective.
// Note that you need to call Retrospective.Unwrap() before calling this method if this Retrospective
// was returned from a transaction, and the transaction was committed or rolled back.
func (r *Retrospective) Update() *RetrospectiveUpdateOne {
	return NewRetrospectiveClient(r.config).UpdateOne(r)
}

// Unwrap unwraps the Retrospective entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (r *Retrospective) Unwrap() *Retrospective {
	_tx, ok := r.config.driver.(*txDriver)
	if !ok {
		panic("ent: Retrospective is not a transactional entity")
	}
	r.config.driver = _tx.drv
	return r
}

// String implements the fmt.Stringer.
func (r *Retrospective) String() string {
	var builder strings.Builder
	builder.WriteString("Retrospective(")
	builder.WriteString(fmt.Sprintf("id=%v, ", r.ID))
	builder.WriteString("document_name=")
	builder.WriteString(r.DocumentName)
	builder.WriteString(", ")
	builder.WriteString("type=")
	builder.WriteString(fmt.Sprintf("%v", r.Type))
	builder.WriteString(", ")
	builder.WriteString("state=")
	builder.WriteString(fmt.Sprintf("%v", r.State))
	builder.WriteByte(')')
	return builder.String()
}

// Retrospectives is a parsable slice of Retrospective.
type Retrospectives []*Retrospective
