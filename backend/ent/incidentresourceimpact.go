// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/functionality"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/incidentresourceimpact"
	"github.com/rezible/rezible/ent/service"
)

// IncidentResourceImpact is the model entity for the IncidentResourceImpact schema.
type IncidentResourceImpact struct {
	config `json:"-"`
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// IncidentID holds the value of the "incident_id" field.
	IncidentID uuid.UUID `json:"incident_id,omitempty"`
	// ServiceID holds the value of the "service_id" field.
	ServiceID uuid.UUID `json:"service_id,omitempty"`
	// FunctionalityID holds the value of the "functionality_id" field.
	FunctionalityID uuid.UUID `json:"functionality_id,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the IncidentResourceImpactQuery when eager-loading is set.
	Edges        IncidentResourceImpactEdges `json:"edges"`
	selectValues sql.SelectValues
}

// IncidentResourceImpactEdges holds the relations/edges for other nodes in the graph.
type IncidentResourceImpactEdges struct {
	// Incident holds the value of the incident edge.
	Incident *Incident `json:"incident,omitempty"`
	// Service holds the value of the service edge.
	Service *Service `json:"service,omitempty"`
	// Functionality holds the value of the functionality edge.
	Functionality *Functionality `json:"functionality,omitempty"`
	// ResultingIncidents holds the value of the resulting_incidents edge.
	ResultingIncidents []*IncidentLink `json:"resulting_incidents,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [4]bool
}

// IncidentOrErr returns the Incident value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e IncidentResourceImpactEdges) IncidentOrErr() (*Incident, error) {
	if e.Incident != nil {
		return e.Incident, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: incident.Label}
	}
	return nil, &NotLoadedError{edge: "incident"}
}

// ServiceOrErr returns the Service value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e IncidentResourceImpactEdges) ServiceOrErr() (*Service, error) {
	if e.Service != nil {
		return e.Service, nil
	} else if e.loadedTypes[1] {
		return nil, &NotFoundError{label: service.Label}
	}
	return nil, &NotLoadedError{edge: "service"}
}

// FunctionalityOrErr returns the Functionality value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e IncidentResourceImpactEdges) FunctionalityOrErr() (*Functionality, error) {
	if e.Functionality != nil {
		return e.Functionality, nil
	} else if e.loadedTypes[2] {
		return nil, &NotFoundError{label: functionality.Label}
	}
	return nil, &NotLoadedError{edge: "functionality"}
}

// ResultingIncidentsOrErr returns the ResultingIncidents value or an error if the edge
// was not loaded in eager-loading.
func (e IncidentResourceImpactEdges) ResultingIncidentsOrErr() ([]*IncidentLink, error) {
	if e.loadedTypes[3] {
		return e.ResultingIncidents, nil
	}
	return nil, &NotLoadedError{edge: "resulting_incidents"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*IncidentResourceImpact) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case incidentresourceimpact.FieldID, incidentresourceimpact.FieldIncidentID, incidentresourceimpact.FieldServiceID, incidentresourceimpact.FieldFunctionalityID:
			values[i] = new(uuid.UUID)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the IncidentResourceImpact fields.
func (iri *IncidentResourceImpact) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case incidentresourceimpact.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				iri.ID = *value
			}
		case incidentresourceimpact.FieldIncidentID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field incident_id", values[i])
			} else if value != nil {
				iri.IncidentID = *value
			}
		case incidentresourceimpact.FieldServiceID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field service_id", values[i])
			} else if value != nil {
				iri.ServiceID = *value
			}
		case incidentresourceimpact.FieldFunctionalityID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field functionality_id", values[i])
			} else if value != nil {
				iri.FunctionalityID = *value
			}
		default:
			iri.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the IncidentResourceImpact.
// This includes values selected through modifiers, order, etc.
func (iri *IncidentResourceImpact) Value(name string) (ent.Value, error) {
	return iri.selectValues.Get(name)
}

// QueryIncident queries the "incident" edge of the IncidentResourceImpact entity.
func (iri *IncidentResourceImpact) QueryIncident() *IncidentQuery {
	return NewIncidentResourceImpactClient(iri.config).QueryIncident(iri)
}

// QueryService queries the "service" edge of the IncidentResourceImpact entity.
func (iri *IncidentResourceImpact) QueryService() *ServiceQuery {
	return NewIncidentResourceImpactClient(iri.config).QueryService(iri)
}

// QueryFunctionality queries the "functionality" edge of the IncidentResourceImpact entity.
func (iri *IncidentResourceImpact) QueryFunctionality() *FunctionalityQuery {
	return NewIncidentResourceImpactClient(iri.config).QueryFunctionality(iri)
}

// QueryResultingIncidents queries the "resulting_incidents" edge of the IncidentResourceImpact entity.
func (iri *IncidentResourceImpact) QueryResultingIncidents() *IncidentLinkQuery {
	return NewIncidentResourceImpactClient(iri.config).QueryResultingIncidents(iri)
}

// Update returns a builder for updating this IncidentResourceImpact.
// Note that you need to call IncidentResourceImpact.Unwrap() before calling this method if this IncidentResourceImpact
// was returned from a transaction, and the transaction was committed or rolled back.
func (iri *IncidentResourceImpact) Update() *IncidentResourceImpactUpdateOne {
	return NewIncidentResourceImpactClient(iri.config).UpdateOne(iri)
}

// Unwrap unwraps the IncidentResourceImpact entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (iri *IncidentResourceImpact) Unwrap() *IncidentResourceImpact {
	_tx, ok := iri.config.driver.(*txDriver)
	if !ok {
		panic("ent: IncidentResourceImpact is not a transactional entity")
	}
	iri.config.driver = _tx.drv
	return iri
}

// String implements the fmt.Stringer.
func (iri *IncidentResourceImpact) String() string {
	var builder strings.Builder
	builder.WriteString("IncidentResourceImpact(")
	builder.WriteString(fmt.Sprintf("id=%v, ", iri.ID))
	builder.WriteString("incident_id=")
	builder.WriteString(fmt.Sprintf("%v", iri.IncidentID))
	builder.WriteString(", ")
	builder.WriteString("service_id=")
	builder.WriteString(fmt.Sprintf("%v", iri.ServiceID))
	builder.WriteString(", ")
	builder.WriteString("functionality_id=")
	builder.WriteString(fmt.Sprintf("%v", iri.FunctionalityID))
	builder.WriteByte(')')
	return builder.String()
}

// IncidentResourceImpacts is a parsable slice of IncidentResourceImpact.
type IncidentResourceImpacts []*IncidentResourceImpact
