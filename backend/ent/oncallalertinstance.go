// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/oncallalert"
	"github.com/rezible/rezible/ent/oncallalertinstance"
	"github.com/rezible/rezible/ent/user"
)

// OncallAlertInstance is the model entity for the OncallAlertInstance schema.
type OncallAlertInstance struct {
	config `json:"-"`
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// AlertID holds the value of the "alert_id" field.
	AlertID uuid.UUID `json:"alert_id,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// AckedAt holds the value of the "acked_at" field.
	AckedAt time.Time `json:"acked_at,omitempty"`
	// ReceiverUserID holds the value of the "receiver_user_id" field.
	ReceiverUserID uuid.UUID `json:"receiver_user_id,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the OncallAlertInstanceQuery when eager-loading is set.
	Edges        OncallAlertInstanceEdges `json:"edges"`
	selectValues sql.SelectValues
}

// OncallAlertInstanceEdges holds the relations/edges for other nodes in the graph.
type OncallAlertInstanceEdges struct {
	// Alert holds the value of the alert edge.
	Alert *OncallAlert `json:"alert,omitempty"`
	// Receiver holds the value of the receiver edge.
	Receiver *User `json:"receiver,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [2]bool
}

// AlertOrErr returns the Alert value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e OncallAlertInstanceEdges) AlertOrErr() (*OncallAlert, error) {
	if e.Alert != nil {
		return e.Alert, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: oncallalert.Label}
	}
	return nil, &NotLoadedError{edge: "alert"}
}

// ReceiverOrErr returns the Receiver value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e OncallAlertInstanceEdges) ReceiverOrErr() (*User, error) {
	if e.Receiver != nil {
		return e.Receiver, nil
	} else if e.loadedTypes[1] {
		return nil, &NotFoundError{label: user.Label}
	}
	return nil, &NotLoadedError{edge: "receiver"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*OncallAlertInstance) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case oncallalertinstance.FieldCreatedAt, oncallalertinstance.FieldAckedAt:
			values[i] = new(sql.NullTime)
		case oncallalertinstance.FieldID, oncallalertinstance.FieldAlertID, oncallalertinstance.FieldReceiverUserID:
			values[i] = new(uuid.UUID)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the OncallAlertInstance fields.
func (oai *OncallAlertInstance) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case oncallalertinstance.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				oai.ID = *value
			}
		case oncallalertinstance.FieldAlertID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field alert_id", values[i])
			} else if value != nil {
				oai.AlertID = *value
			}
		case oncallalertinstance.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				oai.CreatedAt = value.Time
			}
		case oncallalertinstance.FieldAckedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field acked_at", values[i])
			} else if value.Valid {
				oai.AckedAt = value.Time
			}
		case oncallalertinstance.FieldReceiverUserID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field receiver_user_id", values[i])
			} else if value != nil {
				oai.ReceiverUserID = *value
			}
		default:
			oai.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the OncallAlertInstance.
// This includes values selected through modifiers, order, etc.
func (oai *OncallAlertInstance) Value(name string) (ent.Value, error) {
	return oai.selectValues.Get(name)
}

// QueryAlert queries the "alert" edge of the OncallAlertInstance entity.
func (oai *OncallAlertInstance) QueryAlert() *OncallAlertQuery {
	return NewOncallAlertInstanceClient(oai.config).QueryAlert(oai)
}

// QueryReceiver queries the "receiver" edge of the OncallAlertInstance entity.
func (oai *OncallAlertInstance) QueryReceiver() *UserQuery {
	return NewOncallAlertInstanceClient(oai.config).QueryReceiver(oai)
}

// Update returns a builder for updating this OncallAlertInstance.
// Note that you need to call OncallAlertInstance.Unwrap() before calling this method if this OncallAlertInstance
// was returned from a transaction, and the transaction was committed or rolled back.
func (oai *OncallAlertInstance) Update() *OncallAlertInstanceUpdateOne {
	return NewOncallAlertInstanceClient(oai.config).UpdateOne(oai)
}

// Unwrap unwraps the OncallAlertInstance entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (oai *OncallAlertInstance) Unwrap() *OncallAlertInstance {
	_tx, ok := oai.config.driver.(*txDriver)
	if !ok {
		panic("ent: OncallAlertInstance is not a transactional entity")
	}
	oai.config.driver = _tx.drv
	return oai
}

// String implements the fmt.Stringer.
func (oai *OncallAlertInstance) String() string {
	var builder strings.Builder
	builder.WriteString("OncallAlertInstance(")
	builder.WriteString(fmt.Sprintf("id=%v, ", oai.ID))
	builder.WriteString("alert_id=")
	builder.WriteString(fmt.Sprintf("%v", oai.AlertID))
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(oai.CreatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("acked_at=")
	builder.WriteString(oai.AckedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("receiver_user_id=")
	builder.WriteString(fmt.Sprintf("%v", oai.ReceiverUserID))
	builder.WriteByte(')')
	return builder.String()
}

// OncallAlertInstances is a parsable slice of OncallAlertInstance.
type OncallAlertInstances []*OncallAlertInstance
