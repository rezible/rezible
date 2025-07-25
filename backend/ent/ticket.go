// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/ticket"
)

// Ticket is the model entity for the Ticket schema.
type Ticket struct {
	config `json:"-"`
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// ProviderID holds the value of the "provider_id" field.
	ProviderID string `json:"provider_id,omitempty"`
	// Title holds the value of the "title" field.
	Title string `json:"title,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the TicketQuery when eager-loading is set.
	Edges        TicketEdges `json:"edges"`
	selectValues sql.SelectValues
}

// TicketEdges holds the relations/edges for other nodes in the graph.
type TicketEdges struct {
	// Tasks holds the value of the tasks edge.
	Tasks []*Task `json:"tasks,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [1]bool
}

// TasksOrErr returns the Tasks value or an error if the edge
// was not loaded in eager-loading.
func (e TicketEdges) TasksOrErr() ([]*Task, error) {
	if e.loadedTypes[0] {
		return e.Tasks, nil
	}
	return nil, &NotLoadedError{edge: "tasks"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Ticket) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case ticket.FieldProviderID, ticket.FieldTitle:
			values[i] = new(sql.NullString)
		case ticket.FieldID:
			values[i] = new(uuid.UUID)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Ticket fields.
func (t *Ticket) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case ticket.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				t.ID = *value
			}
		case ticket.FieldProviderID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field provider_id", values[i])
			} else if value.Valid {
				t.ProviderID = value.String
			}
		case ticket.FieldTitle:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field title", values[i])
			} else if value.Valid {
				t.Title = value.String
			}
		default:
			t.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the Ticket.
// This includes values selected through modifiers, order, etc.
func (t *Ticket) Value(name string) (ent.Value, error) {
	return t.selectValues.Get(name)
}

// QueryTasks queries the "tasks" edge of the Ticket entity.
func (t *Ticket) QueryTasks() *TaskQuery {
	return NewTicketClient(t.config).QueryTasks(t)
}

// Update returns a builder for updating this Ticket.
// Note that you need to call Ticket.Unwrap() before calling this method if this Ticket
// was returned from a transaction, and the transaction was committed or rolled back.
func (t *Ticket) Update() *TicketUpdateOne {
	return NewTicketClient(t.config).UpdateOne(t)
}

// Unwrap unwraps the Ticket entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (t *Ticket) Unwrap() *Ticket {
	_tx, ok := t.config.driver.(*txDriver)
	if !ok {
		panic("ent: Ticket is not a transactional entity")
	}
	t.config.driver = _tx.drv
	return t
}

// String implements the fmt.Stringer.
func (t *Ticket) String() string {
	var builder strings.Builder
	builder.WriteString("Ticket(")
	builder.WriteString(fmt.Sprintf("id=%v, ", t.ID))
	builder.WriteString("provider_id=")
	builder.WriteString(t.ProviderID)
	builder.WriteString(", ")
	builder.WriteString("title=")
	builder.WriteString(t.Title)
	builder.WriteByte(')')
	return builder.String()
}

// Tickets is a parsable slice of Ticket.
type Tickets []*Ticket
