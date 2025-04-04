// Code generated by ent, DO NOT EDIT.

package systemrelationshipfeedbacksignal

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
)

const (
	// Label holds the string label denoting the systemrelationshipfeedbacksignal type in the database.
	Label = "system_relationship_feedback_signal"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldRelationshipID holds the string denoting the relationship_id field in the database.
	FieldRelationshipID = "relationship_id"
	// FieldSignalID holds the string denoting the signal_id field in the database.
	FieldSignalID = "signal_id"
	// FieldType holds the string denoting the type field in the database.
	FieldType = "type"
	// FieldDescription holds the string denoting the description field in the database.
	FieldDescription = "description"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// EdgeRelationship holds the string denoting the relationship edge name in mutations.
	EdgeRelationship = "relationship"
	// EdgeSignal holds the string denoting the signal edge name in mutations.
	EdgeSignal = "signal"
	// Table holds the table name of the systemrelationshipfeedbacksignal in the database.
	Table = "system_relationship_feedback_signals"
	// RelationshipTable is the table that holds the relationship relation/edge.
	RelationshipTable = "system_relationship_feedback_signals"
	// RelationshipInverseTable is the table name for the SystemAnalysisRelationship entity.
	// It exists in this package in order to avoid circular dependency with the "systemanalysisrelationship" package.
	RelationshipInverseTable = "system_analysis_relationships"
	// RelationshipColumn is the table column denoting the relationship relation/edge.
	RelationshipColumn = "relationship_id"
	// SignalTable is the table that holds the signal relation/edge.
	SignalTable = "system_relationship_feedback_signals"
	// SignalInverseTable is the table name for the SystemComponentSignal entity.
	// It exists in this package in order to avoid circular dependency with the "systemcomponentsignal" package.
	SignalInverseTable = "system_component_signals"
	// SignalColumn is the table column denoting the signal relation/edge.
	SignalColumn = "signal_id"
)

// Columns holds all SQL columns for systemrelationshipfeedbacksignal fields.
var Columns = []string{
	FieldID,
	FieldRelationshipID,
	FieldSignalID,
	FieldType,
	FieldDescription,
	FieldCreatedAt,
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
	// TypeValidator is a validator for the "type" field. It is called by the builders before save.
	TypeValidator func(string) error
	// DefaultCreatedAt holds the default value on creation for the "created_at" field.
	DefaultCreatedAt func() time.Time
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() uuid.UUID
)

// OrderOption defines the ordering options for the SystemRelationshipFeedbackSignal queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByRelationshipID orders the results by the relationship_id field.
func ByRelationshipID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldRelationshipID, opts...).ToFunc()
}

// BySignalID orders the results by the signal_id field.
func BySignalID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldSignalID, opts...).ToFunc()
}

// ByType orders the results by the type field.
func ByType(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldType, opts...).ToFunc()
}

// ByDescription orders the results by the description field.
func ByDescription(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDescription, opts...).ToFunc()
}

// ByCreatedAt orders the results by the created_at field.
func ByCreatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCreatedAt, opts...).ToFunc()
}

// ByRelationshipField orders the results by relationship field.
func ByRelationshipField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newRelationshipStep(), sql.OrderByField(field, opts...))
	}
}

// BySignalField orders the results by signal field.
func BySignalField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newSignalStep(), sql.OrderByField(field, opts...))
	}
}
func newRelationshipStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(RelationshipInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, false, RelationshipTable, RelationshipColumn),
	)
}
func newSignalStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(SignalInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, false, SignalTable, SignalColumn),
	)
}
