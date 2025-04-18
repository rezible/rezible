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
	"github.com/rezible/rezible/ent/systemcomponentcontrol"
)

// SystemComponentControl is the model entity for the SystemComponentControl schema.
type SystemComponentControl struct {
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
	// The values are being populated by the SystemComponentControlQuery when eager-loading is set.
	Edges        SystemComponentControlEdges `json:"edges"`
	selectValues sql.SelectValues
}

// SystemComponentControlEdges holds the relations/edges for other nodes in the graph.
type SystemComponentControlEdges struct {
	// Component holds the value of the component edge.
	Component *SystemComponent `json:"component,omitempty"`
	// Relationships holds the value of the relationships edge.
	Relationships []*SystemAnalysisRelationship `json:"relationships,omitempty"`
	// ControlActions holds the value of the control_actions edge.
	ControlActions []*SystemRelationshipControlAction `json:"control_actions,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [3]bool
}

// ComponentOrErr returns the Component value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e SystemComponentControlEdges) ComponentOrErr() (*SystemComponent, error) {
	if e.Component != nil {
		return e.Component, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: systemcomponent.Label}
	}
	return nil, &NotLoadedError{edge: "component"}
}

// RelationshipsOrErr returns the Relationships value or an error if the edge
// was not loaded in eager-loading.
func (e SystemComponentControlEdges) RelationshipsOrErr() ([]*SystemAnalysisRelationship, error) {
	if e.loadedTypes[1] {
		return e.Relationships, nil
	}
	return nil, &NotLoadedError{edge: "relationships"}
}

// ControlActionsOrErr returns the ControlActions value or an error if the edge
// was not loaded in eager-loading.
func (e SystemComponentControlEdges) ControlActionsOrErr() ([]*SystemRelationshipControlAction, error) {
	if e.loadedTypes[2] {
		return e.ControlActions, nil
	}
	return nil, &NotLoadedError{edge: "control_actions"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*SystemComponentControl) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case systemcomponentcontrol.FieldLabel, systemcomponentcontrol.FieldDescription:
			values[i] = new(sql.NullString)
		case systemcomponentcontrol.FieldCreatedAt:
			values[i] = new(sql.NullTime)
		case systemcomponentcontrol.FieldID, systemcomponentcontrol.FieldComponentID:
			values[i] = new(uuid.UUID)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the SystemComponentControl fields.
func (scc *SystemComponentControl) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case systemcomponentcontrol.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				scc.ID = *value
			}
		case systemcomponentcontrol.FieldComponentID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field component_id", values[i])
			} else if value != nil {
				scc.ComponentID = *value
			}
		case systemcomponentcontrol.FieldLabel:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field label", values[i])
			} else if value.Valid {
				scc.Label = value.String
			}
		case systemcomponentcontrol.FieldDescription:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field description", values[i])
			} else if value.Valid {
				scc.Description = value.String
			}
		case systemcomponentcontrol.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				scc.CreatedAt = value.Time
			}
		default:
			scc.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the SystemComponentControl.
// This includes values selected through modifiers, order, etc.
func (scc *SystemComponentControl) Value(name string) (ent.Value, error) {
	return scc.selectValues.Get(name)
}

// QueryComponent queries the "component" edge of the SystemComponentControl entity.
func (scc *SystemComponentControl) QueryComponent() *SystemComponentQuery {
	return NewSystemComponentControlClient(scc.config).QueryComponent(scc)
}

// QueryRelationships queries the "relationships" edge of the SystemComponentControl entity.
func (scc *SystemComponentControl) QueryRelationships() *SystemAnalysisRelationshipQuery {
	return NewSystemComponentControlClient(scc.config).QueryRelationships(scc)
}

// QueryControlActions queries the "control_actions" edge of the SystemComponentControl entity.
func (scc *SystemComponentControl) QueryControlActions() *SystemRelationshipControlActionQuery {
	return NewSystemComponentControlClient(scc.config).QueryControlActions(scc)
}

// Update returns a builder for updating this SystemComponentControl.
// Note that you need to call SystemComponentControl.Unwrap() before calling this method if this SystemComponentControl
// was returned from a transaction, and the transaction was committed or rolled back.
func (scc *SystemComponentControl) Update() *SystemComponentControlUpdateOne {
	return NewSystemComponentControlClient(scc.config).UpdateOne(scc)
}

// Unwrap unwraps the SystemComponentControl entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (scc *SystemComponentControl) Unwrap() *SystemComponentControl {
	_tx, ok := scc.config.driver.(*txDriver)
	if !ok {
		panic("ent: SystemComponentControl is not a transactional entity")
	}
	scc.config.driver = _tx.drv
	return scc
}

// String implements the fmt.Stringer.
func (scc *SystemComponentControl) String() string {
	var builder strings.Builder
	builder.WriteString("SystemComponentControl(")
	builder.WriteString(fmt.Sprintf("id=%v, ", scc.ID))
	builder.WriteString("component_id=")
	builder.WriteString(fmt.Sprintf("%v", scc.ComponentID))
	builder.WriteString(", ")
	builder.WriteString("label=")
	builder.WriteString(scc.Label)
	builder.WriteString(", ")
	builder.WriteString("description=")
	builder.WriteString(scc.Description)
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(scc.CreatedAt.Format(time.ANSIC))
	builder.WriteByte(')')
	return builder.String()
}

// SystemComponentControls is a parsable slice of SystemComponentControl.
type SystemComponentControls []*SystemComponentControl
