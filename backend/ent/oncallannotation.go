// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/oncallannotation"
	"github.com/rezible/rezible/ent/oncallroster"
	"github.com/rezible/rezible/ent/user"
)

// OncallAnnotation is the model entity for the OncallAnnotation schema.
type OncallAnnotation struct {
	config `json:"-"`
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// RosterID holds the value of the "roster_id" field.
	RosterID uuid.UUID `json:"roster_id,omitempty"`
	// CreatorID holds the value of the "creator_id" field.
	CreatorID uuid.UUID `json:"creator_id,omitempty"`
	// EventID holds the value of the "event_id" field.
	EventID string `json:"event_id,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// MinutesOccupied holds the value of the "minutes_occupied" field.
	MinutesOccupied int `json:"minutes_occupied,omitempty"`
	// Notes holds the value of the "notes" field.
	Notes string `json:"notes,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the OncallAnnotationQuery when eager-loading is set.
	Edges        OncallAnnotationEdges `json:"edges"`
	selectValues sql.SelectValues
}

// OncallAnnotationEdges holds the relations/edges for other nodes in the graph.
type OncallAnnotationEdges struct {
	// Roster holds the value of the roster edge.
	Roster *OncallRoster `json:"roster,omitempty"`
	// Creator holds the value of the creator edge.
	Creator *User `json:"creator,omitempty"`
	// Handovers holds the value of the handovers edge.
	Handovers []*OncallUserShiftHandover `json:"handovers,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [3]bool
}

// RosterOrErr returns the Roster value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e OncallAnnotationEdges) RosterOrErr() (*OncallRoster, error) {
	if e.Roster != nil {
		return e.Roster, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: oncallroster.Label}
	}
	return nil, &NotLoadedError{edge: "roster"}
}

// CreatorOrErr returns the Creator value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e OncallAnnotationEdges) CreatorOrErr() (*User, error) {
	if e.Creator != nil {
		return e.Creator, nil
	} else if e.loadedTypes[1] {
		return nil, &NotFoundError{label: user.Label}
	}
	return nil, &NotLoadedError{edge: "creator"}
}

// HandoversOrErr returns the Handovers value or an error if the edge
// was not loaded in eager-loading.
func (e OncallAnnotationEdges) HandoversOrErr() ([]*OncallUserShiftHandover, error) {
	if e.loadedTypes[2] {
		return e.Handovers, nil
	}
	return nil, &NotLoadedError{edge: "handovers"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*OncallAnnotation) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case oncallannotation.FieldMinutesOccupied:
			values[i] = new(sql.NullInt64)
		case oncallannotation.FieldEventID, oncallannotation.FieldNotes:
			values[i] = new(sql.NullString)
		case oncallannotation.FieldCreatedAt:
			values[i] = new(sql.NullTime)
		case oncallannotation.FieldID, oncallannotation.FieldRosterID, oncallannotation.FieldCreatorID:
			values[i] = new(uuid.UUID)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the OncallAnnotation fields.
func (oa *OncallAnnotation) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case oncallannotation.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				oa.ID = *value
			}
		case oncallannotation.FieldRosterID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field roster_id", values[i])
			} else if value != nil {
				oa.RosterID = *value
			}
		case oncallannotation.FieldCreatorID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field creator_id", values[i])
			} else if value != nil {
				oa.CreatorID = *value
			}
		case oncallannotation.FieldEventID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field event_id", values[i])
			} else if value.Valid {
				oa.EventID = value.String
			}
		case oncallannotation.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				oa.CreatedAt = value.Time
			}
		case oncallannotation.FieldMinutesOccupied:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field minutes_occupied", values[i])
			} else if value.Valid {
				oa.MinutesOccupied = int(value.Int64)
			}
		case oncallannotation.FieldNotes:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field notes", values[i])
			} else if value.Valid {
				oa.Notes = value.String
			}
		default:
			oa.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the OncallAnnotation.
// This includes values selected through modifiers, order, etc.
func (oa *OncallAnnotation) Value(name string) (ent.Value, error) {
	return oa.selectValues.Get(name)
}

// QueryRoster queries the "roster" edge of the OncallAnnotation entity.
func (oa *OncallAnnotation) QueryRoster() *OncallRosterQuery {
	return NewOncallAnnotationClient(oa.config).QueryRoster(oa)
}

// QueryCreator queries the "creator" edge of the OncallAnnotation entity.
func (oa *OncallAnnotation) QueryCreator() *UserQuery {
	return NewOncallAnnotationClient(oa.config).QueryCreator(oa)
}

// QueryHandovers queries the "handovers" edge of the OncallAnnotation entity.
func (oa *OncallAnnotation) QueryHandovers() *OncallUserShiftHandoverQuery {
	return NewOncallAnnotationClient(oa.config).QueryHandovers(oa)
}

// Update returns a builder for updating this OncallAnnotation.
// Note that you need to call OncallAnnotation.Unwrap() before calling this method if this OncallAnnotation
// was returned from a transaction, and the transaction was committed or rolled back.
func (oa *OncallAnnotation) Update() *OncallAnnotationUpdateOne {
	return NewOncallAnnotationClient(oa.config).UpdateOne(oa)
}

// Unwrap unwraps the OncallAnnotation entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (oa *OncallAnnotation) Unwrap() *OncallAnnotation {
	_tx, ok := oa.config.driver.(*txDriver)
	if !ok {
		panic("ent: OncallAnnotation is not a transactional entity")
	}
	oa.config.driver = _tx.drv
	return oa
}

// String implements the fmt.Stringer.
func (oa *OncallAnnotation) String() string {
	var builder strings.Builder
	builder.WriteString("OncallAnnotation(")
	builder.WriteString(fmt.Sprintf("id=%v, ", oa.ID))
	builder.WriteString("roster_id=")
	builder.WriteString(fmt.Sprintf("%v", oa.RosterID))
	builder.WriteString(", ")
	builder.WriteString("creator_id=")
	builder.WriteString(fmt.Sprintf("%v", oa.CreatorID))
	builder.WriteString(", ")
	builder.WriteString("event_id=")
	builder.WriteString(oa.EventID)
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(oa.CreatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("minutes_occupied=")
	builder.WriteString(fmt.Sprintf("%v", oa.MinutesOccupied))
	builder.WriteString(", ")
	builder.WriteString("notes=")
	builder.WriteString(oa.Notes)
	builder.WriteByte(')')
	return builder.String()
}

// OncallAnnotations is a parsable slice of OncallAnnotation.
type OncallAnnotations []*OncallAnnotation
