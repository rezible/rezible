// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/systemcomponent"
	"github.com/rezible/rezible/ent/systemcomponentsignal"
)

// SystemComponentSignal is the model entity for the SystemComponentSignal schema.
type SystemComponentSignal struct {
	config `json:"-"`
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// ComponentID holds the value of the "component_id" field.
	ComponentID uuid.UUID `json:"component_id,omitempty"`
	// Label holds the value of the "label" field.
	Label string `json:"label,omitempty"`
	// Description holds the value of the "description" field.
	Description string `json:"description,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the SystemComponentSignalQuery when eager-loading is set.
	Edges        SystemComponentSignalEdges `json:"edges"`
	selectValues sql.SelectValues
}

// SystemComponentSignalEdges holds the relations/edges for other nodes in the graph.
type SystemComponentSignalEdges struct {
	// Component holds the value of the component edge.
	Component *SystemComponent `json:"component,omitempty"`
	// Relationships holds the value of the relationships edge.
	Relationships []*SystemAnalysisRelationship `json:"relationships,omitempty"`
	// FeedbackSignals holds the value of the feedback_signals edge.
	FeedbackSignals []*SystemRelationshipFeedbackSignal `json:"feedback_signals,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [3]bool
}

// ComponentOrErr returns the Component value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e SystemComponentSignalEdges) ComponentOrErr() (*SystemComponent, error) {
	if e.Component != nil {
		return e.Component, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: systemcomponent.Label}
	}
	return nil, &NotLoadedError{edge: "component"}
}

// RelationshipsOrErr returns the Relationships value or an error if the edge
// was not loaded in eager-loading.
func (e SystemComponentSignalEdges) RelationshipsOrErr() ([]*SystemAnalysisRelationship, error) {
	if e.loadedTypes[1] {
		return e.Relationships, nil
	}
	return nil, &NotLoadedError{edge: "relationships"}
}

// FeedbackSignalsOrErr returns the FeedbackSignals value or an error if the edge
// was not loaded in eager-loading.
func (e SystemComponentSignalEdges) FeedbackSignalsOrErr() ([]*SystemRelationshipFeedbackSignal, error) {
	if e.loadedTypes[2] {
		return e.FeedbackSignals, nil
	}
	return nil, &NotLoadedError{edge: "feedback_signals"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*SystemComponentSignal) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case systemcomponentsignal.FieldLabel, systemcomponentsignal.FieldDescription:
			values[i] = new(sql.NullString)
		case systemcomponentsignal.FieldCreatedAt:
			values[i] = new(sql.NullTime)
		case systemcomponentsignal.FieldID, systemcomponentsignal.FieldComponentID:
			values[i] = new(uuid.UUID)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the SystemComponentSignal fields.
func (scs *SystemComponentSignal) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case systemcomponentsignal.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				scs.ID = *value
			}
		case systemcomponentsignal.FieldComponentID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field component_id", values[i])
			} else if value != nil {
				scs.ComponentID = *value
			}
		case systemcomponentsignal.FieldLabel:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field label", values[i])
			} else if value.Valid {
				scs.Label = value.String
			}
		case systemcomponentsignal.FieldDescription:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field description", values[i])
			} else if value.Valid {
				scs.Description = value.String
			}
		case systemcomponentsignal.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				scs.CreatedAt = value.Time
			}
		default:
			scs.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the SystemComponentSignal.
// This includes values selected through modifiers, order, etc.
func (scs *SystemComponentSignal) Value(name string) (ent.Value, error) {
	return scs.selectValues.Get(name)
}

// QueryComponent queries the "component" edge of the SystemComponentSignal entity.
func (scs *SystemComponentSignal) QueryComponent() *SystemComponentQuery {
	return NewSystemComponentSignalClient(scs.config).QueryComponent(scs)
}

// QueryRelationships queries the "relationships" edge of the SystemComponentSignal entity.
func (scs *SystemComponentSignal) QueryRelationships() *SystemAnalysisRelationshipQuery {
	return NewSystemComponentSignalClient(scs.config).QueryRelationships(scs)
}

// QueryFeedbackSignals queries the "feedback_signals" edge of the SystemComponentSignal entity.
func (scs *SystemComponentSignal) QueryFeedbackSignals() *SystemRelationshipFeedbackSignalQuery {
	return NewSystemComponentSignalClient(scs.config).QueryFeedbackSignals(scs)
}

// Update returns a builder for updating this SystemComponentSignal.
// Note that you need to call SystemComponentSignal.Unwrap() before calling this method if this SystemComponentSignal
// was returned from a transaction, and the transaction was committed or rolled back.
func (scs *SystemComponentSignal) Update() *SystemComponentSignalUpdateOne {
	return NewSystemComponentSignalClient(scs.config).UpdateOne(scs)
}

// Unwrap unwraps the SystemComponentSignal entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (scs *SystemComponentSignal) Unwrap() *SystemComponentSignal {
	_tx, ok := scs.config.driver.(*txDriver)
	if !ok {
		panic("ent: SystemComponentSignal is not a transactional entity")
	}
	scs.config.driver = _tx.drv
	return scs
}

// String implements the fmt.Stringer.
func (scs *SystemComponentSignal) String() string {
	var builder strings.Builder
	builder.WriteString("SystemComponentSignal(")
	builder.WriteString(fmt.Sprintf("id=%v, ", scs.ID))
	builder.WriteString("component_id=")
	builder.WriteString(fmt.Sprintf("%v", scs.ComponentID))
	builder.WriteString(", ")
	builder.WriteString("label=")
	builder.WriteString(scs.Label)
	builder.WriteString(", ")
	builder.WriteString("description=")
	builder.WriteString(scs.Description)
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(scs.CreatedAt.Format(time.ANSIC))
	builder.WriteByte(')')
	return builder.String()
}

// SystemComponentSignals is a parsable slice of SystemComponentSignal.
type SystemComponentSignals []*SystemComponentSignal
