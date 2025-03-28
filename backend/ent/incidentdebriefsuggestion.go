// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/incidentdebrief"
	"github.com/rezible/rezible/ent/incidentdebriefsuggestion"
)

// IncidentDebriefSuggestion is the model entity for the IncidentDebriefSuggestion schema.
type IncidentDebriefSuggestion struct {
	config `json:"-"`
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// Content holds the value of the "content" field.
	Content string `json:"content,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the IncidentDebriefSuggestionQuery when eager-loading is set.
	Edges                        IncidentDebriefSuggestionEdges `json:"edges"`
	incident_debrief_suggestions *uuid.UUID
	selectValues                 sql.SelectValues
}

// IncidentDebriefSuggestionEdges holds the relations/edges for other nodes in the graph.
type IncidentDebriefSuggestionEdges struct {
	// Debrief holds the value of the debrief edge.
	Debrief *IncidentDebrief `json:"debrief,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [1]bool
}

// DebriefOrErr returns the Debrief value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e IncidentDebriefSuggestionEdges) DebriefOrErr() (*IncidentDebrief, error) {
	if e.Debrief != nil {
		return e.Debrief, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: incidentdebrief.Label}
	}
	return nil, &NotLoadedError{edge: "debrief"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*IncidentDebriefSuggestion) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case incidentdebriefsuggestion.FieldContent:
			values[i] = new(sql.NullString)
		case incidentdebriefsuggestion.FieldID:
			values[i] = new(uuid.UUID)
		case incidentdebriefsuggestion.ForeignKeys[0]: // incident_debrief_suggestions
			values[i] = &sql.NullScanner{S: new(uuid.UUID)}
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the IncidentDebriefSuggestion fields.
func (ids *IncidentDebriefSuggestion) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case incidentdebriefsuggestion.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				ids.ID = *value
			}
		case incidentdebriefsuggestion.FieldContent:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field content", values[i])
			} else if value.Valid {
				ids.Content = value.String
			}
		case incidentdebriefsuggestion.ForeignKeys[0]:
			if value, ok := values[i].(*sql.NullScanner); !ok {
				return fmt.Errorf("unexpected type %T for field incident_debrief_suggestions", values[i])
			} else if value.Valid {
				ids.incident_debrief_suggestions = new(uuid.UUID)
				*ids.incident_debrief_suggestions = *value.S.(*uuid.UUID)
			}
		default:
			ids.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the IncidentDebriefSuggestion.
// This includes values selected through modifiers, order, etc.
func (ids *IncidentDebriefSuggestion) Value(name string) (ent.Value, error) {
	return ids.selectValues.Get(name)
}

// QueryDebrief queries the "debrief" edge of the IncidentDebriefSuggestion entity.
func (ids *IncidentDebriefSuggestion) QueryDebrief() *IncidentDebriefQuery {
	return NewIncidentDebriefSuggestionClient(ids.config).QueryDebrief(ids)
}

// Update returns a builder for updating this IncidentDebriefSuggestion.
// Note that you need to call IncidentDebriefSuggestion.Unwrap() before calling this method if this IncidentDebriefSuggestion
// was returned from a transaction, and the transaction was committed or rolled back.
func (ids *IncidentDebriefSuggestion) Update() *IncidentDebriefSuggestionUpdateOne {
	return NewIncidentDebriefSuggestionClient(ids.config).UpdateOne(ids)
}

// Unwrap unwraps the IncidentDebriefSuggestion entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (ids *IncidentDebriefSuggestion) Unwrap() *IncidentDebriefSuggestion {
	_tx, ok := ids.config.driver.(*txDriver)
	if !ok {
		panic("ent: IncidentDebriefSuggestion is not a transactional entity")
	}
	ids.config.driver = _tx.drv
	return ids
}

// String implements the fmt.Stringer.
func (ids *IncidentDebriefSuggestion) String() string {
	var builder strings.Builder
	builder.WriteString("IncidentDebriefSuggestion(")
	builder.WriteString(fmt.Sprintf("id=%v, ", ids.ID))
	builder.WriteString("content=")
	builder.WriteString(ids.Content)
	builder.WriteByte(')')
	return builder.String()
}

// IncidentDebriefSuggestions is a parsable slice of IncidentDebriefSuggestion.
type IncidentDebriefSuggestions []*IncidentDebriefSuggestion
