// Code generated by ent, DO NOT EDIT.

package incidenteventevidence

import (
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
)

const (
	// Label holds the string label denoting the incidenteventevidence type in the database.
	Label = "incident_event_evidence"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldEvidenceType holds the string denoting the evidence_type field in the database.
	FieldEvidenceType = "evidence_type"
	// FieldURL holds the string denoting the url field in the database.
	FieldURL = "url"
	// FieldTitle holds the string denoting the title field in the database.
	FieldTitle = "title"
	// FieldDescription holds the string denoting the description field in the database.
	FieldDescription = "description"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// EdgeEvent holds the string denoting the event edge name in mutations.
	EdgeEvent = "event"
	// Table holds the table name of the incidenteventevidence in the database.
	Table = "incident_event_evidences"
	// EventTable is the table that holds the event relation/edge.
	EventTable = "incident_event_evidences"
	// EventInverseTable is the table name for the IncidentEvent entity.
	// It exists in this package in order to avoid circular dependency with the "incidentevent" package.
	EventInverseTable = "incident_events"
	// EventColumn is the table column denoting the event relation/edge.
	EventColumn = "incident_event_evidence"
)

// Columns holds all SQL columns for incidenteventevidence fields.
var Columns = []string{
	FieldID,
	FieldEvidenceType,
	FieldURL,
	FieldTitle,
	FieldDescription,
	FieldCreatedAt,
}

// ForeignKeys holds the SQL foreign-keys that are owned by the "incident_event_evidences"
// table and are not defined as standalone fields in the schema.
var ForeignKeys = []string{
	"incident_event_evidence",
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
	// URLValidator is a validator for the "url" field. It is called by the builders before save.
	URLValidator func(string) error
	// TitleValidator is a validator for the "title" field. It is called by the builders before save.
	TitleValidator func(string) error
	// DefaultCreatedAt holds the default value on creation for the "created_at" field.
	DefaultCreatedAt func() time.Time
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() uuid.UUID
)

// EvidenceType defines the type for the "evidence_type" enum field.
type EvidenceType string

// EvidenceType values.
const (
	EvidenceTypeLog    EvidenceType = "log"
	EvidenceTypeMetric EvidenceType = "metric"
	EvidenceTypeChat   EvidenceType = "chat"
	EvidenceTypeTicket EvidenceType = "ticket"
	EvidenceTypeOther  EvidenceType = "other"
)

func (et EvidenceType) String() string {
	return string(et)
}

// EvidenceTypeValidator is a validator for the "evidence_type" field enum values. It is called by the builders before save.
func EvidenceTypeValidator(et EvidenceType) error {
	switch et {
	case EvidenceTypeLog, EvidenceTypeMetric, EvidenceTypeChat, EvidenceTypeTicket, EvidenceTypeOther:
		return nil
	default:
		return fmt.Errorf("incidenteventevidence: invalid enum value for evidence_type field: %q", et)
	}
}

// OrderOption defines the ordering options for the IncidentEventEvidence queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByEvidenceType orders the results by the evidence_type field.
func ByEvidenceType(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldEvidenceType, opts...).ToFunc()
}

// ByURL orders the results by the url field.
func ByURL(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldURL, opts...).ToFunc()
}

// ByTitle orders the results by the title field.
func ByTitle(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldTitle, opts...).ToFunc()
}

// ByDescription orders the results by the description field.
func ByDescription(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDescription, opts...).ToFunc()
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
		sqlgraph.Edge(sqlgraph.M2O, true, EventTable, EventColumn),
	)
}
