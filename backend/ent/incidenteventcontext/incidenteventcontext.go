// Code generated by ent, DO NOT EDIT.

package incidenteventcontext

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
)

const (
	// Label holds the string label denoting the incidenteventcontext type in the database.
	Label = "incident_event_context"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldSystemState holds the string denoting the system_state field in the database.
	FieldSystemState = "system_state"
	// FieldDecisionOptions holds the string denoting the decision_options field in the database.
	FieldDecisionOptions = "decision_options"
	// FieldDecisionRationale holds the string denoting the decision_rationale field in the database.
	FieldDecisionRationale = "decision_rationale"
	// FieldInvolvedPersonnel holds the string denoting the involved_personnel field in the database.
	FieldInvolvedPersonnel = "involved_personnel"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// EdgeEvent holds the string denoting the event edge name in mutations.
	EdgeEvent = "event"
	// Table holds the table name of the incidenteventcontext in the database.
	Table = "incident_event_contexts"
	// EventTable is the table that holds the event relation/edge.
	EventTable = "incident_event_contexts"
	// EventInverseTable is the table name for the IncidentEvent entity.
	// It exists in this package in order to avoid circular dependency with the "incidentevent" package.
	EventInverseTable = "incident_events"
	// EventColumn is the table column denoting the event relation/edge.
	EventColumn = "incident_event_context"
)

// Columns holds all SQL columns for incidenteventcontext fields.
var Columns = []string{
	FieldID,
	FieldSystemState,
	FieldDecisionOptions,
	FieldDecisionRationale,
	FieldInvolvedPersonnel,
	FieldCreatedAt,
}

// ForeignKeys holds the SQL foreign-keys that are owned by the "incident_event_contexts"
// table and are not defined as standalone fields in the schema.
var ForeignKeys = []string{
	"incident_event_context",
}

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	for i := range ForeignKeys {
		if column == ForeignKeys[i] {
			return true
		}
	}
	return false
}

var (
	// DefaultCreatedAt holds the default value on creation for the "created_at" field.
	DefaultCreatedAt func() time.Time
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() uuid.UUID
)

// OrderOption defines the ordering options for the IncidentEventContext queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// BySystemState orders the results by the system_state field.
func BySystemState(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldSystemState, opts...).ToFunc()
}

// ByDecisionRationale orders the results by the decision_rationale field.
func ByDecisionRationale(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDecisionRationale, opts...).ToFunc()
}

// ByCreatedAt orders the results by the created_at field.
func ByCreatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCreatedAt, opts...).ToFunc()
}

// ByEventField orders the results by event field.
func ByEventField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newEventStep(), sql.OrderByField(field, opts...))
	}
}
func newEventStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(EventInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2O, true, EventTable, EventColumn),
	)
}
