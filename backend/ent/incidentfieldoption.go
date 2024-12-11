// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/twohundreds/rezible/ent/incidentfield"
	"github.com/twohundreds/rezible/ent/incidentfieldoption"
)

// IncidentFieldOption is the model entity for the IncidentFieldOption schema.
type IncidentFieldOption struct {
	config `json:"-"`
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// ArchiveTime holds the value of the "archive_time" field.
	ArchiveTime time.Time `json:"archive_time,omitempty"`
	// IncidentFieldID holds the value of the "incident_field_id" field.
	IncidentFieldID uuid.UUID `json:"incident_field_id,omitempty"`
	// Type holds the value of the "type" field.
	Type incidentfieldoption.Type `json:"type,omitempty"`
	// Value holds the value of the "value" field.
	Value string `json:"value,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the IncidentFieldOptionQuery when eager-loading is set.
	Edges        IncidentFieldOptionEdges `json:"edges"`
	selectValues sql.SelectValues
}

// IncidentFieldOptionEdges holds the relations/edges for other nodes in the graph.
type IncidentFieldOptionEdges struct {
	// IncidentField holds the value of the incident_field edge.
	IncidentField *IncidentField `json:"incident_field,omitempty"`
	// Incidents holds the value of the incidents edge.
	Incidents []*Incident `json:"incidents,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [2]bool
}

// IncidentFieldOrErr returns the IncidentField value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e IncidentFieldOptionEdges) IncidentFieldOrErr() (*IncidentField, error) {
	if e.IncidentField != nil {
		return e.IncidentField, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: incidentfield.Label}
	}
	return nil, &NotLoadedError{edge: "incident_field"}
}

// IncidentsOrErr returns the Incidents value or an error if the edge
// was not loaded in eager-loading.
func (e IncidentFieldOptionEdges) IncidentsOrErr() ([]*Incident, error) {
	if e.loadedTypes[1] {
		return e.Incidents, nil
	}
	return nil, &NotLoadedError{edge: "incidents"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*IncidentFieldOption) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case incidentfieldoption.FieldType, incidentfieldoption.FieldValue:
			values[i] = new(sql.NullString)
		case incidentfieldoption.FieldArchiveTime:
			values[i] = new(sql.NullTime)
		case incidentfieldoption.FieldID, incidentfieldoption.FieldIncidentFieldID:
			values[i] = new(uuid.UUID)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the IncidentFieldOption fields.
func (ifo *IncidentFieldOption) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case incidentfieldoption.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				ifo.ID = *value
			}
		case incidentfieldoption.FieldArchiveTime:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field archive_time", values[i])
			} else if value.Valid {
				ifo.ArchiveTime = value.Time
			}
		case incidentfieldoption.FieldIncidentFieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field incident_field_id", values[i])
			} else if value != nil {
				ifo.IncidentFieldID = *value
			}
		case incidentfieldoption.FieldType:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field type", values[i])
			} else if value.Valid {
				ifo.Type = incidentfieldoption.Type(value.String)
			}
		case incidentfieldoption.FieldValue:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field value", values[i])
			} else if value.Valid {
				ifo.Value = value.String
			}
		default:
			ifo.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// GetValue returns the ent.Value that was dynamically selected and assigned to the IncidentFieldOption.
// This includes values selected through modifiers, order, etc.
func (ifo *IncidentFieldOption) GetValue(name string) (ent.Value, error) {
	return ifo.selectValues.Get(name)
}

// QueryIncidentField queries the "incident_field" edge of the IncidentFieldOption entity.
func (ifo *IncidentFieldOption) QueryIncidentField() *IncidentFieldQuery {
	return NewIncidentFieldOptionClient(ifo.config).QueryIncidentField(ifo)
}

// QueryIncidents queries the "incidents" edge of the IncidentFieldOption entity.
func (ifo *IncidentFieldOption) QueryIncidents() *IncidentQuery {
	return NewIncidentFieldOptionClient(ifo.config).QueryIncidents(ifo)
}

// Update returns a builder for updating this IncidentFieldOption.
// Note that you need to call IncidentFieldOption.Unwrap() before calling this method if this IncidentFieldOption
// was returned from a transaction, and the transaction was committed or rolled back.
func (ifo *IncidentFieldOption) Update() *IncidentFieldOptionUpdateOne {
	return NewIncidentFieldOptionClient(ifo.config).UpdateOne(ifo)
}

// Unwrap unwraps the IncidentFieldOption entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (ifo *IncidentFieldOption) Unwrap() *IncidentFieldOption {
	_tx, ok := ifo.config.driver.(*txDriver)
	if !ok {
		panic("ent: IncidentFieldOption is not a transactional entity")
	}
	ifo.config.driver = _tx.drv
	return ifo
}

// String implements the fmt.Stringer.
func (ifo *IncidentFieldOption) String() string {
	var builder strings.Builder
	builder.WriteString("IncidentFieldOption(")
	builder.WriteString(fmt.Sprintf("id=%v, ", ifo.ID))
	builder.WriteString("archive_time=")
	builder.WriteString(ifo.ArchiveTime.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("incident_field_id=")
	builder.WriteString(fmt.Sprintf("%v", ifo.IncidentFieldID))
	builder.WriteString(", ")
	builder.WriteString("type=")
	builder.WriteString(fmt.Sprintf("%v", ifo.Type))
	builder.WriteString(", ")
	builder.WriteString("value=")
	builder.WriteString(ifo.Value)
	builder.WriteByte(')')
	return builder.String()
}

// IncidentFieldOptions is a parsable slice of IncidentFieldOption.
type IncidentFieldOptions []*IncidentFieldOption