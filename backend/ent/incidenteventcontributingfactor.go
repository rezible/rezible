// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/incidentevent"
	"github.com/rezible/rezible/ent/incidenteventcontributingfactor"
)

// IncidentEventContributingFactor is the model entity for the IncidentEventContributingFactor schema.
type IncidentEventContributingFactor struct {
	config `json:"-"`
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// FactorType holds the value of the "factor_type" field.
	FactorType string `json:"factor_type,omitempty"`
	// Description holds the value of the "description" field.
	Description string `json:"description,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the IncidentEventContributingFactorQuery when eager-loading is set.
	Edges                  IncidentEventContributingFactorEdges `json:"edges"`
	incident_event_factors *uuid.UUID
	selectValues           sql.SelectValues
}

// IncidentEventContributingFactorEdges holds the relations/edges for other nodes in the graph.
type IncidentEventContributingFactorEdges struct {
	// Event holds the value of the event edge.
	Event *IncidentEvent `json:"event,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [1]bool
}

// EventOrErr returns the Event value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e IncidentEventContributingFactorEdges) EventOrErr() (*IncidentEvent, error) {
	if e.Event != nil {
		return e.Event, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: incidentevent.Label}
	}
	return nil, &NotLoadedError{edge: "event"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*IncidentEventContributingFactor) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case incidenteventcontributingfactor.FieldFactorType, incidenteventcontributingfactor.FieldDescription:
			values[i] = new(sql.NullString)
		case incidenteventcontributingfactor.FieldCreatedAt:
			values[i] = new(sql.NullTime)
		case incidenteventcontributingfactor.FieldID:
			values[i] = new(uuid.UUID)
		case incidenteventcontributingfactor.ForeignKeys[0]: // incident_event_factors
			values[i] = &sql.NullScanner{S: new(uuid.UUID)}
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the IncidentEventContributingFactor fields.
func (iecf *IncidentEventContributingFactor) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case incidenteventcontributingfactor.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				iecf.ID = *value
			}
		case incidenteventcontributingfactor.FieldFactorType:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field factor_type", values[i])
			} else if value.Valid {
				iecf.FactorType = value.String
			}
		case incidenteventcontributingfactor.FieldDescription:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field description", values[i])
			} else if value.Valid {
				iecf.Description = value.String
			}
		case incidenteventcontributingfactor.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				iecf.CreatedAt = value.Time
			}
		case incidenteventcontributingfactor.ForeignKeys[0]:
			if value, ok := values[i].(*sql.NullScanner); !ok {
				return fmt.Errorf("unexpected type %T for field incident_event_factors", values[i])
			} else if value.Valid {
				iecf.incident_event_factors = new(uuid.UUID)
				*iecf.incident_event_factors = *value.S.(*uuid.UUID)
			}
		default:
			iecf.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the IncidentEventContributingFactor.
// This includes values selected through modifiers, order, etc.
func (iecf *IncidentEventContributingFactor) Value(name string) (ent.Value, error) {
	return iecf.selectValues.Get(name)
}

// QueryEvent queries the "event" edge of the IncidentEventContributingFactor entity.
func (iecf *IncidentEventContributingFactor) QueryEvent() *IncidentEventQuery {
	return NewIncidentEventContributingFactorClient(iecf.config).QueryEvent(iecf)
}

// Update returns a builder for updating this IncidentEventContributingFactor.
// Note that you need to call IncidentEventContributingFactor.Unwrap() before calling this method if this IncidentEventContributingFactor
// was returned from a transaction, and the transaction was committed or rolled back.
func (iecf *IncidentEventContributingFactor) Update() *IncidentEventContributingFactorUpdateOne {
	return NewIncidentEventContributingFactorClient(iecf.config).UpdateOne(iecf)
}

// Unwrap unwraps the IncidentEventContributingFactor entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (iecf *IncidentEventContributingFactor) Unwrap() *IncidentEventContributingFactor {
	_tx, ok := iecf.config.driver.(*txDriver)
	if !ok {
		panic("ent: IncidentEventContributingFactor is not a transactional entity")
	}
	iecf.config.driver = _tx.drv
	return iecf
}

// String implements the fmt.Stringer.
func (iecf *IncidentEventContributingFactor) String() string {
	var builder strings.Builder
	builder.WriteString("IncidentEventContributingFactor(")
	builder.WriteString(fmt.Sprintf("id=%v, ", iecf.ID))
	builder.WriteString("factor_type=")
	builder.WriteString(iecf.FactorType)
	builder.WriteString(", ")
	builder.WriteString("description=")
	builder.WriteString(iecf.Description)
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(iecf.CreatedAt.Format(time.ANSIC))
	builder.WriteByte(')')
	return builder.String()
}

// IncidentEventContributingFactors is a parsable slice of IncidentEventContributingFactor.
type IncidentEventContributingFactors []*IncidentEventContributingFactor
