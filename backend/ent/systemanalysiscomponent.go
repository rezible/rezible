// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/systemanalysis"
	"github.com/rezible/rezible/ent/systemanalysiscomponent"
	"github.com/rezible/rezible/ent/systemcomponent"
)

// SystemAnalysisComponent is the model entity for the SystemAnalysisComponent schema.
type SystemAnalysisComponent struct {
	config `json:"-"`
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// AnalysisID holds the value of the "analysis_id" field.
	AnalysisID uuid.UUID `json:"analysis_id,omitempty"`
	// ComponentID holds the value of the "component_id" field.
	ComponentID uuid.UUID `json:"component_id,omitempty"`
	// Description holds the value of the "description" field.
	Description string `json:"description,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the SystemAnalysisComponentQuery when eager-loading is set.
	Edges        SystemAnalysisComponentEdges `json:"edges"`
	selectValues sql.SelectValues
}

// SystemAnalysisComponentEdges holds the relations/edges for other nodes in the graph.
type SystemAnalysisComponentEdges struct {
	// Analysis holds the value of the analysis edge.
	Analysis *SystemAnalysis `json:"analysis,omitempty"`
	// Component holds the value of the component edge.
	Component *SystemComponent `json:"component,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [2]bool
}

// AnalysisOrErr returns the Analysis value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e SystemAnalysisComponentEdges) AnalysisOrErr() (*SystemAnalysis, error) {
	if e.Analysis != nil {
		return e.Analysis, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: systemanalysis.Label}
	}
	return nil, &NotLoadedError{edge: "analysis"}
}

// ComponentOrErr returns the Component value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e SystemAnalysisComponentEdges) ComponentOrErr() (*SystemComponent, error) {
	if e.Component != nil {
		return e.Component, nil
	} else if e.loadedTypes[1] {
		return nil, &NotFoundError{label: systemcomponent.Label}
	}
	return nil, &NotLoadedError{edge: "component"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*SystemAnalysisComponent) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case systemanalysiscomponent.FieldDescription:
			values[i] = new(sql.NullString)
		case systemanalysiscomponent.FieldCreatedAt:
			values[i] = new(sql.NullTime)
		case systemanalysiscomponent.FieldID, systemanalysiscomponent.FieldAnalysisID, systemanalysiscomponent.FieldComponentID:
			values[i] = new(uuid.UUID)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the SystemAnalysisComponent fields.
func (sac *SystemAnalysisComponent) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case systemanalysiscomponent.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				sac.ID = *value
			}
		case systemanalysiscomponent.FieldAnalysisID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field analysis_id", values[i])
			} else if value != nil {
				sac.AnalysisID = *value
			}
		case systemanalysiscomponent.FieldComponentID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field component_id", values[i])
			} else if value != nil {
				sac.ComponentID = *value
			}
		case systemanalysiscomponent.FieldDescription:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field description", values[i])
			} else if value.Valid {
				sac.Description = value.String
			}
		case systemanalysiscomponent.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				sac.CreatedAt = value.Time
			}
		default:
			sac.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the SystemAnalysisComponent.
// This includes values selected through modifiers, order, etc.
func (sac *SystemAnalysisComponent) Value(name string) (ent.Value, error) {
	return sac.selectValues.Get(name)
}

// QueryAnalysis queries the "analysis" edge of the SystemAnalysisComponent entity.
func (sac *SystemAnalysisComponent) QueryAnalysis() *SystemAnalysisQuery {
	return NewSystemAnalysisComponentClient(sac.config).QueryAnalysis(sac)
}

// QueryComponent queries the "component" edge of the SystemAnalysisComponent entity.
func (sac *SystemAnalysisComponent) QueryComponent() *SystemComponentQuery {
	return NewSystemAnalysisComponentClient(sac.config).QueryComponent(sac)
}

// Update returns a builder for updating this SystemAnalysisComponent.
// Note that you need to call SystemAnalysisComponent.Unwrap() before calling this method if this SystemAnalysisComponent
// was returned from a transaction, and the transaction was committed or rolled back.
func (sac *SystemAnalysisComponent) Update() *SystemAnalysisComponentUpdateOne {
	return NewSystemAnalysisComponentClient(sac.config).UpdateOne(sac)
}

// Unwrap unwraps the SystemAnalysisComponent entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (sac *SystemAnalysisComponent) Unwrap() *SystemAnalysisComponent {
	_tx, ok := sac.config.driver.(*txDriver)
	if !ok {
		panic("ent: SystemAnalysisComponent is not a transactional entity")
	}
	sac.config.driver = _tx.drv
	return sac
}

// String implements the fmt.Stringer.
func (sac *SystemAnalysisComponent) String() string {
	var builder strings.Builder
	builder.WriteString("SystemAnalysisComponent(")
	builder.WriteString(fmt.Sprintf("id=%v, ", sac.ID))
	builder.WriteString("analysis_id=")
	builder.WriteString(fmt.Sprintf("%v", sac.AnalysisID))
	builder.WriteString(", ")
	builder.WriteString("component_id=")
	builder.WriteString(fmt.Sprintf("%v", sac.ComponentID))
	builder.WriteString(", ")
	builder.WriteString("description=")
	builder.WriteString(sac.Description)
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(sac.CreatedAt.Format(time.ANSIC))
	builder.WriteByte(')')
	return builder.String()
}

// SystemAnalysisComponents is a parsable slice of SystemAnalysisComponent.
type SystemAnalysisComponents []*SystemAnalysisComponent
