// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/systemanalysis"
)

// SystemAnalysis is the model entity for the SystemAnalysis schema.
type SystemAnalysis struct {
	config `json:"-"`
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// IncidentID holds the value of the "incident_id" field.
	IncidentID uuid.UUID `json:"incident_id,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// UpdatedAt holds the value of the "updated_at" field.
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the SystemAnalysisQuery when eager-loading is set.
	Edges        SystemAnalysisEdges `json:"edges"`
	selectValues sql.SelectValues
}

// SystemAnalysisEdges holds the relations/edges for other nodes in the graph.
type SystemAnalysisEdges struct {
	// Incident holds the value of the incident edge.
	Incident *Incident `json:"incident,omitempty"`
	// Components holds the value of the components edge.
	Components []*SystemComponent `json:"components,omitempty"`
	// AnalysisComponents holds the value of the analysis_components edge.
	AnalysisComponents []*SystemAnalysisComponent `json:"analysis_components,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [3]bool
}

// IncidentOrErr returns the Incident value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e SystemAnalysisEdges) IncidentOrErr() (*Incident, error) {
	if e.Incident != nil {
		return e.Incident, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: incident.Label}
	}
	return nil, &NotLoadedError{edge: "incident"}
}

// ComponentsOrErr returns the Components value or an error if the edge
// was not loaded in eager-loading.
func (e SystemAnalysisEdges) ComponentsOrErr() ([]*SystemComponent, error) {
	if e.loadedTypes[1] {
		return e.Components, nil
	}
	return nil, &NotLoadedError{edge: "components"}
}

// AnalysisComponentsOrErr returns the AnalysisComponents value or an error if the edge
// was not loaded in eager-loading.
func (e SystemAnalysisEdges) AnalysisComponentsOrErr() ([]*SystemAnalysisComponent, error) {
	if e.loadedTypes[2] {
		return e.AnalysisComponents, nil
	}
	return nil, &NotLoadedError{edge: "analysis_components"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*SystemAnalysis) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case systemanalysis.FieldCreatedAt, systemanalysis.FieldUpdatedAt:
			values[i] = new(sql.NullTime)
		case systemanalysis.FieldID, systemanalysis.FieldIncidentID:
			values[i] = new(uuid.UUID)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the SystemAnalysis fields.
func (sa *SystemAnalysis) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case systemanalysis.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				sa.ID = *value
			}
		case systemanalysis.FieldIncidentID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field incident_id", values[i])
			} else if value != nil {
				sa.IncidentID = *value
			}
		case systemanalysis.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				sa.CreatedAt = value.Time
			}
		case systemanalysis.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				sa.UpdatedAt = value.Time
			}
		default:
			sa.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the SystemAnalysis.
// This includes values selected through modifiers, order, etc.
func (sa *SystemAnalysis) Value(name string) (ent.Value, error) {
	return sa.selectValues.Get(name)
}

// QueryIncident queries the "incident" edge of the SystemAnalysis entity.
func (sa *SystemAnalysis) QueryIncident() *IncidentQuery {
	return NewSystemAnalysisClient(sa.config).QueryIncident(sa)
}

// QueryComponents queries the "components" edge of the SystemAnalysis entity.
func (sa *SystemAnalysis) QueryComponents() *SystemComponentQuery {
	return NewSystemAnalysisClient(sa.config).QueryComponents(sa)
}

// QueryAnalysisComponents queries the "analysis_components" edge of the SystemAnalysis entity.
func (sa *SystemAnalysis) QueryAnalysisComponents() *SystemAnalysisComponentQuery {
	return NewSystemAnalysisClient(sa.config).QueryAnalysisComponents(sa)
}

// Update returns a builder for updating this SystemAnalysis.
// Note that you need to call SystemAnalysis.Unwrap() before calling this method if this SystemAnalysis
// was returned from a transaction, and the transaction was committed or rolled back.
func (sa *SystemAnalysis) Update() *SystemAnalysisUpdateOne {
	return NewSystemAnalysisClient(sa.config).UpdateOne(sa)
}

// Unwrap unwraps the SystemAnalysis entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (sa *SystemAnalysis) Unwrap() *SystemAnalysis {
	_tx, ok := sa.config.driver.(*txDriver)
	if !ok {
		panic("ent: SystemAnalysis is not a transactional entity")
	}
	sa.config.driver = _tx.drv
	return sa
}

// String implements the fmt.Stringer.
func (sa *SystemAnalysis) String() string {
	var builder strings.Builder
	builder.WriteString("SystemAnalysis(")
	builder.WriteString(fmt.Sprintf("id=%v, ", sa.ID))
	builder.WriteString("incident_id=")
	builder.WriteString(fmt.Sprintf("%v", sa.IncidentID))
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(sa.CreatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(sa.UpdatedAt.Format(time.ANSIC))
	builder.WriteByte(')')
	return builder.String()
}

// SystemAnalyses is a parsable slice of SystemAnalysis.
type SystemAnalyses []*SystemAnalysis
