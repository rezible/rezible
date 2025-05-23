// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/systemcomponentkind"
)

// SystemComponentKind is the model entity for the SystemComponentKind schema.
type SystemComponentKind struct {
	config `json:"-"`
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// ProviderID holds the value of the "provider_id" field.
	ProviderID string `json:"provider_id,omitempty"`
	// Label holds the value of the "label" field.
	Label string `json:"label,omitempty"`
	// Description holds the value of the "description" field.
	Description string `json:"description,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the SystemComponentKindQuery when eager-loading is set.
	Edges        SystemComponentKindEdges `json:"edges"`
	selectValues sql.SelectValues
}

// SystemComponentKindEdges holds the relations/edges for other nodes in the graph.
type SystemComponentKindEdges struct {
	// Components holds the value of the components edge.
	Components []*SystemComponent `json:"components,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [1]bool
}

// ComponentsOrErr returns the Components value or an error if the edge
// was not loaded in eager-loading.
func (e SystemComponentKindEdges) ComponentsOrErr() ([]*SystemComponent, error) {
	if e.loadedTypes[0] {
		return e.Components, nil
	}
	return nil, &NotLoadedError{edge: "components"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*SystemComponentKind) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case systemcomponentkind.FieldProviderID, systemcomponentkind.FieldLabel, systemcomponentkind.FieldDescription:
			values[i] = new(sql.NullString)
		case systemcomponentkind.FieldCreatedAt:
			values[i] = new(sql.NullTime)
		case systemcomponentkind.FieldID:
			values[i] = new(uuid.UUID)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the SystemComponentKind fields.
func (sck *SystemComponentKind) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case systemcomponentkind.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				sck.ID = *value
			}
		case systemcomponentkind.FieldProviderID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field provider_id", values[i])
			} else if value.Valid {
				sck.ProviderID = value.String
			}
		case systemcomponentkind.FieldLabel:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field label", values[i])
			} else if value.Valid {
				sck.Label = value.String
			}
		case systemcomponentkind.FieldDescription:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field description", values[i])
			} else if value.Valid {
				sck.Description = value.String
			}
		case systemcomponentkind.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				sck.CreatedAt = value.Time
			}
		default:
			sck.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the SystemComponentKind.
// This includes values selected through modifiers, order, etc.
func (sck *SystemComponentKind) Value(name string) (ent.Value, error) {
	return sck.selectValues.Get(name)
}

// QueryComponents queries the "components" edge of the SystemComponentKind entity.
func (sck *SystemComponentKind) QueryComponents() *SystemComponentQuery {
	return NewSystemComponentKindClient(sck.config).QueryComponents(sck)
}

// Update returns a builder for updating this SystemComponentKind.
// Note that you need to call SystemComponentKind.Unwrap() before calling this method if this SystemComponentKind
// was returned from a transaction, and the transaction was committed or rolled back.
func (sck *SystemComponentKind) Update() *SystemComponentKindUpdateOne {
	return NewSystemComponentKindClient(sck.config).UpdateOne(sck)
}

// Unwrap unwraps the SystemComponentKind entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (sck *SystemComponentKind) Unwrap() *SystemComponentKind {
	_tx, ok := sck.config.driver.(*txDriver)
	if !ok {
		panic("ent: SystemComponentKind is not a transactional entity")
	}
	sck.config.driver = _tx.drv
	return sck
}

// String implements the fmt.Stringer.
func (sck *SystemComponentKind) String() string {
	var builder strings.Builder
	builder.WriteString("SystemComponentKind(")
	builder.WriteString(fmt.Sprintf("id=%v, ", sck.ID))
	builder.WriteString("provider_id=")
	builder.WriteString(sck.ProviderID)
	builder.WriteString(", ")
	builder.WriteString("label=")
	builder.WriteString(sck.Label)
	builder.WriteString(", ")
	builder.WriteString("description=")
	builder.WriteString(sck.Description)
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(sck.CreatedAt.Format(time.ANSIC))
	builder.WriteByte(')')
	return builder.String()
}

// SystemComponentKinds is a parsable slice of SystemComponentKind.
type SystemComponentKinds []*SystemComponentKind
