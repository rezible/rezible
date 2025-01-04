// Code generated by ent, DO NOT EDIT.

package incidentevent

import (
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
)

const (
	// Label holds the string label denoting the incidentevent type in the database.
	Label = "incident_event"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldIncidentID holds the string denoting the incident_id field in the database.
	FieldIncidentID = "incident_id"
	// FieldTimestamp holds the string denoting the timestamp field in the database.
	FieldTimestamp = "timestamp"
	// FieldType holds the string denoting the type field in the database.
	FieldType = "type"
	// FieldTitle holds the string denoting the title field in the database.
	FieldTitle = "title"
	// FieldDescription holds the string denoting the description field in the database.
	FieldDescription = "description"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// FieldUpdatedAt holds the string denoting the updated_at field in the database.
	FieldUpdatedAt = "updated_at"
	// FieldCreatedBy holds the string denoting the created_by field in the database.
	FieldCreatedBy = "created_by"
	// FieldSequence holds the string denoting the sequence field in the database.
	FieldSequence = "sequence"
	// FieldIsDraft holds the string denoting the is_draft field in the database.
	FieldIsDraft = "is_draft"
	// EdgeIncident holds the string denoting the incident edge name in mutations.
	EdgeIncident = "incident"
	// EdgeContext holds the string denoting the context edge name in mutations.
	EdgeContext = "context"
	// EdgeFactors holds the string denoting the factors edge name in mutations.
	EdgeFactors = "factors"
	// EdgeEvidence holds the string denoting the evidence edge name in mutations.
	EdgeEvidence = "evidence"
	// Table holds the table name of the incidentevent in the database.
	Table = "incident_events"
	// IncidentTable is the table that holds the incident relation/edge.
	IncidentTable = "incident_events"
	// IncidentInverseTable is the table name for the Incident entity.
	// It exists in this package in order to avoid circular dependency with the "incident" package.
	IncidentInverseTable = "incidents"
	// IncidentColumn is the table column denoting the incident relation/edge.
	IncidentColumn = "incident_id"
	// ContextTable is the table that holds the context relation/edge.
	ContextTable = "incident_event_contexts"
	// ContextInverseTable is the table name for the IncidentEventContext entity.
	// It exists in this package in order to avoid circular dependency with the "incidenteventcontext" package.
	ContextInverseTable = "incident_event_contexts"
	// ContextColumn is the table column denoting the context relation/edge.
	ContextColumn = "incident_event_context"
	// FactorsTable is the table that holds the factors relation/edge.
	FactorsTable = "incident_event_contributing_factors"
	// FactorsInverseTable is the table name for the IncidentEventContributingFactor entity.
	// It exists in this package in order to avoid circular dependency with the "incidenteventcontributingfactor" package.
	FactorsInverseTable = "incident_event_contributing_factors"
	// FactorsColumn is the table column denoting the factors relation/edge.
	FactorsColumn = "incident_event_factors"
	// EvidenceTable is the table that holds the evidence relation/edge.
	EvidenceTable = "incident_event_evidences"
	// EvidenceInverseTable is the table name for the IncidentEventEvidence entity.
	// It exists in this package in order to avoid circular dependency with the "incidenteventevidence" package.
	EvidenceInverseTable = "incident_event_evidences"
	// EvidenceColumn is the table column denoting the evidence relation/edge.
	EvidenceColumn = "incident_event_evidence"
)

// Columns holds all SQL columns for incidentevent fields.
var Columns = []string{
	FieldID,
	FieldIncidentID,
	FieldTimestamp,
	FieldType,
	FieldTitle,
	FieldDescription,
	FieldCreatedAt,
	FieldUpdatedAt,
	FieldCreatedBy,
	FieldSequence,
	FieldIsDraft,
}

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	return false
}

var (
	// TitleValidator is a validator for the "title" field. It is called by the builders before save.
	TitleValidator func(string) error
	// DefaultCreatedAt holds the default value on creation for the "created_at" field.
	DefaultCreatedAt func() time.Time
	// DefaultUpdatedAt holds the default value on creation for the "updated_at" field.
	DefaultUpdatedAt func() time.Time
	// UpdateDefaultUpdatedAt holds the default value on update for the "updated_at" field.
	UpdateDefaultUpdatedAt func() time.Time
	// DefaultIsDraft holds the default value on creation for the "is_draft" field.
	DefaultIsDraft bool
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() uuid.UUID
)

// Type defines the type for the "type" enum field.
type Type string

// Type values.
const (
	TypeObservation Type = "observation"
	TypeAction      Type = "action"
	TypeDecision    Type = "decision"
	TypeContext     Type = "context"
)

func (_type Type) String() string {
	return string(_type)
}

// TypeValidator is a validator for the "type" field enum values. It is called by the builders before save.
func TypeValidator(_type Type) error {
	switch _type {
	case TypeObservation, TypeAction, TypeDecision, TypeContext:
		return nil
	default:
		return fmt.Errorf("incidentevent: invalid enum value for type field: %q", _type)
	}
}

// OrderOption defines the ordering options for the IncidentEvent queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByIncidentID orders the results by the incident_id field.
func ByIncidentID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldIncidentID, opts...).ToFunc()
}

// ByTimestamp orders the results by the timestamp field.
func ByTimestamp(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldTimestamp, opts...).ToFunc()
}

// ByType orders the results by the type field.
func ByType(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldType, opts...).ToFunc()
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

// ByUpdatedAt orders the results by the updated_at field.
func ByUpdatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldUpdatedAt, opts...).ToFunc()
}

// ByCreatedBy orders the results by the created_by field.
func ByCreatedBy(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCreatedBy, opts...).ToFunc()
}

// BySequence orders the results by the sequence field.
func BySequence(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldSequence, opts...).ToFunc()
}

// ByIsDraft orders the results by the is_draft field.
func ByIsDraft(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldIsDraft, opts...).ToFunc()
}

// ByIncidentField orders the results by incident field.
func ByIncidentField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newIncidentStep(), sql.OrderByField(field, opts...))
	}
}

// ByContextField orders the results by context field.
func ByContextField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newContextStep(), sql.OrderByField(field, opts...))
	}
}

// ByFactorsCount orders the results by factors count.
func ByFactorsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newFactorsStep(), opts...)
	}
}

// ByFactors orders the results by factors terms.
func ByFactors(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newFactorsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByEvidenceCount orders the results by evidence count.
func ByEvidenceCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newEvidenceStep(), opts...)
	}
}

// ByEvidence orders the results by evidence terms.
func ByEvidence(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newEvidenceStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}
func newIncidentStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(IncidentInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, IncidentTable, IncidentColumn),
	)
}
func newContextStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(ContextInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2O, false, ContextTable, ContextColumn),
	)
}
func newFactorsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(FactorsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, FactorsTable, FactorsColumn),
	)
}
func newEvidenceStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(EvidenceInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, EvidenceTable, EvidenceColumn),
	)
}
