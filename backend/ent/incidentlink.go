// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/twohundreds/rezible/ent/incident"
	"github.com/twohundreds/rezible/ent/incidentlink"
	"github.com/twohundreds/rezible/ent/incidentresourceimpact"
)

// IncidentLink is the model entity for the IncidentLink schema.
type IncidentLink struct {
	config `json:"-"`
	// ID of the ent.
	ID int `json:"id,omitempty"`
	// IncidentID holds the value of the "incident_id" field.
	IncidentID uuid.UUID `json:"incident_id,omitempty"`
	// LinkedIncidentID holds the value of the "linked_incident_id" field.
	LinkedIncidentID uuid.UUID `json:"linked_incident_id,omitempty"`
	// Description holds the value of the "description" field.
	Description string `json:"description,omitempty"`
	// LinkType holds the value of the "link_type" field.
	LinkType incidentlink.LinkType `json:"link_type,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the IncidentLinkQuery when eager-loading is set.
	Edges                         IncidentLinkEdges `json:"edges"`
	incident_link_resource_impact *uuid.UUID
	selectValues                  sql.SelectValues
}

// IncidentLinkEdges holds the relations/edges for other nodes in the graph.
type IncidentLinkEdges struct {
	// Incident holds the value of the incident edge.
	Incident *Incident `json:"incident,omitempty"`
	// LinkedIncident holds the value of the linked_incident edge.
	LinkedIncident *Incident `json:"linked_incident,omitempty"`
	// ResourceImpact holds the value of the resource_impact edge.
	ResourceImpact *IncidentResourceImpact `json:"resource_impact,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [3]bool
}

// IncidentOrErr returns the Incident value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e IncidentLinkEdges) IncidentOrErr() (*Incident, error) {
	if e.Incident != nil {
		return e.Incident, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: incident.Label}
	}
	return nil, &NotLoadedError{edge: "incident"}
}

// LinkedIncidentOrErr returns the LinkedIncident value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e IncidentLinkEdges) LinkedIncidentOrErr() (*Incident, error) {
	if e.LinkedIncident != nil {
		return e.LinkedIncident, nil
	} else if e.loadedTypes[1] {
		return nil, &NotFoundError{label: incident.Label}
	}
	return nil, &NotLoadedError{edge: "linked_incident"}
}

// ResourceImpactOrErr returns the ResourceImpact value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e IncidentLinkEdges) ResourceImpactOrErr() (*IncidentResourceImpact, error) {
	if e.ResourceImpact != nil {
		return e.ResourceImpact, nil
	} else if e.loadedTypes[2] {
		return nil, &NotFoundError{label: incidentresourceimpact.Label}
	}
	return nil, &NotLoadedError{edge: "resource_impact"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*IncidentLink) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case incidentlink.FieldID:
			values[i] = new(sql.NullInt64)
		case incidentlink.FieldDescription, incidentlink.FieldLinkType:
			values[i] = new(sql.NullString)
		case incidentlink.FieldIncidentID, incidentlink.FieldLinkedIncidentID:
			values[i] = new(uuid.UUID)
		case incidentlink.ForeignKeys[0]: // incident_link_resource_impact
			values[i] = &sql.NullScanner{S: new(uuid.UUID)}
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the IncidentLink fields.
func (il *IncidentLink) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case incidentlink.FieldID:
			value, ok := values[i].(*sql.NullInt64)
			if !ok {
				return fmt.Errorf("unexpected type %T for field id", value)
			}
			il.ID = int(value.Int64)
		case incidentlink.FieldIncidentID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field incident_id", values[i])
			} else if value != nil {
				il.IncidentID = *value
			}
		case incidentlink.FieldLinkedIncidentID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field linked_incident_id", values[i])
			} else if value != nil {
				il.LinkedIncidentID = *value
			}
		case incidentlink.FieldDescription:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field description", values[i])
			} else if value.Valid {
				il.Description = value.String
			}
		case incidentlink.FieldLinkType:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field link_type", values[i])
			} else if value.Valid {
				il.LinkType = incidentlink.LinkType(value.String)
			}
		case incidentlink.ForeignKeys[0]:
			if value, ok := values[i].(*sql.NullScanner); !ok {
				return fmt.Errorf("unexpected type %T for field incident_link_resource_impact", values[i])
			} else if value.Valid {
				il.incident_link_resource_impact = new(uuid.UUID)
				*il.incident_link_resource_impact = *value.S.(*uuid.UUID)
			}
		default:
			il.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the IncidentLink.
// This includes values selected through modifiers, order, etc.
func (il *IncidentLink) Value(name string) (ent.Value, error) {
	return il.selectValues.Get(name)
}

// QueryIncident queries the "incident" edge of the IncidentLink entity.
func (il *IncidentLink) QueryIncident() *IncidentQuery {
	return NewIncidentLinkClient(il.config).QueryIncident(il)
}

// QueryLinkedIncident queries the "linked_incident" edge of the IncidentLink entity.
func (il *IncidentLink) QueryLinkedIncident() *IncidentQuery {
	return NewIncidentLinkClient(il.config).QueryLinkedIncident(il)
}

// QueryResourceImpact queries the "resource_impact" edge of the IncidentLink entity.
func (il *IncidentLink) QueryResourceImpact() *IncidentResourceImpactQuery {
	return NewIncidentLinkClient(il.config).QueryResourceImpact(il)
}

// Update returns a builder for updating this IncidentLink.
// Note that you need to call IncidentLink.Unwrap() before calling this method if this IncidentLink
// was returned from a transaction, and the transaction was committed or rolled back.
func (il *IncidentLink) Update() *IncidentLinkUpdateOne {
	return NewIncidentLinkClient(il.config).UpdateOne(il)
}

// Unwrap unwraps the IncidentLink entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (il *IncidentLink) Unwrap() *IncidentLink {
	_tx, ok := il.config.driver.(*txDriver)
	if !ok {
		panic("ent: IncidentLink is not a transactional entity")
	}
	il.config.driver = _tx.drv
	return il
}

// String implements the fmt.Stringer.
func (il *IncidentLink) String() string {
	var builder strings.Builder
	builder.WriteString("IncidentLink(")
	builder.WriteString(fmt.Sprintf("id=%v, ", il.ID))
	builder.WriteString("incident_id=")
	builder.WriteString(fmt.Sprintf("%v", il.IncidentID))
	builder.WriteString(", ")
	builder.WriteString("linked_incident_id=")
	builder.WriteString(fmt.Sprintf("%v", il.LinkedIncidentID))
	builder.WriteString(", ")
	builder.WriteString("description=")
	builder.WriteString(il.Description)
	builder.WriteString(", ")
	builder.WriteString("link_type=")
	builder.WriteString(fmt.Sprintf("%v", il.LinkType))
	builder.WriteByte(')')
	return builder.String()
}

// IncidentLinks is a parsable slice of IncidentLink.
type IncidentLinks []*IncidentLink
