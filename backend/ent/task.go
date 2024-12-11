// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/twohundreds/rezible/ent/incident"
	"github.com/twohundreds/rezible/ent/task"
	"github.com/twohundreds/rezible/ent/user"
)

// Task is the model entity for the Task schema.
type Task struct {
	config `json:"-"`
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// Type holds the value of the "type" field.
	Type task.Type `json:"type,omitempty"`
	// Title holds the value of the "title" field.
	Title string `json:"title,omitempty"`
	// IncidentID holds the value of the "incident_id" field.
	IncidentID uuid.UUID `json:"incident_id,omitempty"`
	// AssigneeID holds the value of the "assignee_id" field.
	AssigneeID uuid.UUID `json:"assignee_id,omitempty"`
	// CreatorID holds the value of the "creator_id" field.
	CreatorID uuid.UUID `json:"creator_id,omitempty"`
	// IssueTrackerID holds the value of the "issue_tracker_id" field.
	IssueTrackerID string `json:"issue_tracker_id,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the TaskQuery when eager-loading is set.
	Edges        TaskEdges `json:"edges"`
	selectValues sql.SelectValues
}

// TaskEdges holds the relations/edges for other nodes in the graph.
type TaskEdges struct {
	// Incident holds the value of the incident edge.
	Incident *Incident `json:"incident,omitempty"`
	// Assignee holds the value of the assignee edge.
	Assignee *User `json:"assignee,omitempty"`
	// Creator holds the value of the creator edge.
	Creator *User `json:"creator,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [3]bool
}

// IncidentOrErr returns the Incident value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e TaskEdges) IncidentOrErr() (*Incident, error) {
	if e.Incident != nil {
		return e.Incident, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: incident.Label}
	}
	return nil, &NotLoadedError{edge: "incident"}
}

// AssigneeOrErr returns the Assignee value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e TaskEdges) AssigneeOrErr() (*User, error) {
	if e.Assignee != nil {
		return e.Assignee, nil
	} else if e.loadedTypes[1] {
		return nil, &NotFoundError{label: user.Label}
	}
	return nil, &NotLoadedError{edge: "assignee"}
}

// CreatorOrErr returns the Creator value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e TaskEdges) CreatorOrErr() (*User, error) {
	if e.Creator != nil {
		return e.Creator, nil
	} else if e.loadedTypes[2] {
		return nil, &NotFoundError{label: user.Label}
	}
	return nil, &NotLoadedError{edge: "creator"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Task) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case task.FieldType, task.FieldTitle, task.FieldIssueTrackerID:
			values[i] = new(sql.NullString)
		case task.FieldID, task.FieldIncidentID, task.FieldAssigneeID, task.FieldCreatorID:
			values[i] = new(uuid.UUID)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Task fields.
func (t *Task) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case task.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				t.ID = *value
			}
		case task.FieldType:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field type", values[i])
			} else if value.Valid {
				t.Type = task.Type(value.String)
			}
		case task.FieldTitle:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field title", values[i])
			} else if value.Valid {
				t.Title = value.String
			}
		case task.FieldIncidentID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field incident_id", values[i])
			} else if value != nil {
				t.IncidentID = *value
			}
		case task.FieldAssigneeID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field assignee_id", values[i])
			} else if value != nil {
				t.AssigneeID = *value
			}
		case task.FieldCreatorID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field creator_id", values[i])
			} else if value != nil {
				t.CreatorID = *value
			}
		case task.FieldIssueTrackerID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field issue_tracker_id", values[i])
			} else if value.Valid {
				t.IssueTrackerID = value.String
			}
		default:
			t.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the Task.
// This includes values selected through modifiers, order, etc.
func (t *Task) Value(name string) (ent.Value, error) {
	return t.selectValues.Get(name)
}

// QueryIncident queries the "incident" edge of the Task entity.
func (t *Task) QueryIncident() *IncidentQuery {
	return NewTaskClient(t.config).QueryIncident(t)
}

// QueryAssignee queries the "assignee" edge of the Task entity.
func (t *Task) QueryAssignee() *UserQuery {
	return NewTaskClient(t.config).QueryAssignee(t)
}

// QueryCreator queries the "creator" edge of the Task entity.
func (t *Task) QueryCreator() *UserQuery {
	return NewTaskClient(t.config).QueryCreator(t)
}

// Update returns a builder for updating this Task.
// Note that you need to call Task.Unwrap() before calling this method if this Task
// was returned from a transaction, and the transaction was committed or rolled back.
func (t *Task) Update() *TaskUpdateOne {
	return NewTaskClient(t.config).UpdateOne(t)
}

// Unwrap unwraps the Task entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (t *Task) Unwrap() *Task {
	_tx, ok := t.config.driver.(*txDriver)
	if !ok {
		panic("ent: Task is not a transactional entity")
	}
	t.config.driver = _tx.drv
	return t
}

// String implements the fmt.Stringer.
func (t *Task) String() string {
	var builder strings.Builder
	builder.WriteString("Task(")
	builder.WriteString(fmt.Sprintf("id=%v, ", t.ID))
	builder.WriteString("type=")
	builder.WriteString(fmt.Sprintf("%v", t.Type))
	builder.WriteString(", ")
	builder.WriteString("title=")
	builder.WriteString(t.Title)
	builder.WriteString(", ")
	builder.WriteString("incident_id=")
	builder.WriteString(fmt.Sprintf("%v", t.IncidentID))
	builder.WriteString(", ")
	builder.WriteString("assignee_id=")
	builder.WriteString(fmt.Sprintf("%v", t.AssigneeID))
	builder.WriteString(", ")
	builder.WriteString("creator_id=")
	builder.WriteString(fmt.Sprintf("%v", t.CreatorID))
	builder.WriteString(", ")
	builder.WriteString("issue_tracker_id=")
	builder.WriteString(t.IssueTrackerID)
	builder.WriteByte(')')
	return builder.String()
}

// Tasks is a parsable slice of Task.
type Tasks []*Task