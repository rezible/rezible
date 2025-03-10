// Code generated by ent, DO NOT EDIT.

package systemanalysisrelationship

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
)

const (
	// Label holds the string label denoting the systemanalysisrelationship type in the database.
	Label = "system_analysis_relationship"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldAnalysisID holds the string denoting the analysis_id field in the database.
	FieldAnalysisID = "analysis_id"
	// FieldComponentRelationshipID holds the string denoting the component_relationship_id field in the database.
	FieldComponentRelationshipID = "component_relationship_id"
	// FieldDescription holds the string denoting the description field in the database.
	FieldDescription = "description"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// EdgeSystemAnalysis holds the string denoting the system_analysis edge name in mutations.
	EdgeSystemAnalysis = "system_analysis"
	// EdgeComponentRelationship holds the string denoting the component_relationship edge name in mutations.
	EdgeComponentRelationship = "component_relationship"
	// EdgeControls holds the string denoting the controls edge name in mutations.
	EdgeControls = "controls"
	// EdgeSignals holds the string denoting the signals edge name in mutations.
	EdgeSignals = "signals"
	// EdgeControlActions holds the string denoting the control_actions edge name in mutations.
	EdgeControlActions = "control_actions"
	// EdgeFeedbackSignals holds the string denoting the feedback_signals edge name in mutations.
	EdgeFeedbackSignals = "feedback_signals"
	// Table holds the table name of the systemanalysisrelationship in the database.
	Table = "system_analysis_relationships"
	// SystemAnalysisTable is the table that holds the system_analysis relation/edge.
	SystemAnalysisTable = "system_analysis_relationships"
	// SystemAnalysisInverseTable is the table name for the SystemAnalysis entity.
	// It exists in this package in order to avoid circular dependency with the "systemanalysis" package.
	SystemAnalysisInverseTable = "system_analyses"
	// SystemAnalysisColumn is the table column denoting the system_analysis relation/edge.
	SystemAnalysisColumn = "analysis_id"
	// ComponentRelationshipTable is the table that holds the component_relationship relation/edge.
	ComponentRelationshipTable = "system_analysis_relationships"
	// ComponentRelationshipInverseTable is the table name for the SystemComponentRelationship entity.
	// It exists in this package in order to avoid circular dependency with the "systemcomponentrelationship" package.
	ComponentRelationshipInverseTable = "system_component_relationships"
	// ComponentRelationshipColumn is the table column denoting the component_relationship relation/edge.
	ComponentRelationshipColumn = "component_relationship_id"
	// ControlsTable is the table that holds the controls relation/edge. The primary key declared below.
	ControlsTable = "system_relationship_control_actions"
	// ControlsInverseTable is the table name for the SystemComponentControl entity.
	// It exists in this package in order to avoid circular dependency with the "systemcomponentcontrol" package.
	ControlsInverseTable = "system_component_controls"
	// SignalsTable is the table that holds the signals relation/edge. The primary key declared below.
	SignalsTable = "system_relationship_feedback_signals"
	// SignalsInverseTable is the table name for the SystemComponentSignal entity.
	// It exists in this package in order to avoid circular dependency with the "systemcomponentsignal" package.
	SignalsInverseTable = "system_component_signals"
	// ControlActionsTable is the table that holds the control_actions relation/edge.
	ControlActionsTable = "system_relationship_control_actions"
	// ControlActionsInverseTable is the table name for the SystemRelationshipControlAction entity.
	// It exists in this package in order to avoid circular dependency with the "systemrelationshipcontrolaction" package.
	ControlActionsInverseTable = "system_relationship_control_actions"
	// ControlActionsColumn is the table column denoting the control_actions relation/edge.
	ControlActionsColumn = "relationship_id"
	// FeedbackSignalsTable is the table that holds the feedback_signals relation/edge.
	FeedbackSignalsTable = "system_relationship_feedback_signals"
	// FeedbackSignalsInverseTable is the table name for the SystemRelationshipFeedbackSignal entity.
	// It exists in this package in order to avoid circular dependency with the "systemrelationshipfeedbacksignal" package.
	FeedbackSignalsInverseTable = "system_relationship_feedback_signals"
	// FeedbackSignalsColumn is the table column denoting the feedback_signals relation/edge.
	FeedbackSignalsColumn = "relationship_id"
)

// Columns holds all SQL columns for systemanalysisrelationship fields.
var Columns = []string{
	FieldID,
	FieldAnalysisID,
	FieldComponentRelationshipID,
	FieldDescription,
	FieldCreatedAt,
}

var (
	// ControlsPrimaryKey and ControlsColumn2 are the table columns denoting the
	// primary key for the controls relation (M2M).
	ControlsPrimaryKey = []string{"relationship_id", "control_id"}
	// SignalsPrimaryKey and SignalsColumn2 are the table columns denoting the
	// primary key for the signals relation (M2M).
	SignalsPrimaryKey = []string{"relationship_id", "signal_id"}
)

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
	// DefaultCreatedAt holds the default value on creation for the "created_at" field.
	DefaultCreatedAt func() time.Time
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() uuid.UUID
)

// OrderOption defines the ordering options for the SystemAnalysisRelationship queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByAnalysisID orders the results by the analysis_id field.
func ByAnalysisID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldAnalysisID, opts...).ToFunc()
}

// ByComponentRelationshipID orders the results by the component_relationship_id field.
func ByComponentRelationshipID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldComponentRelationshipID, opts...).ToFunc()
}

// ByDescription orders the results by the description field.
func ByDescription(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDescription, opts...).ToFunc()
}

// ByCreatedAt orders the results by the created_at field.
func ByCreatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCreatedAt, opts...).ToFunc()
}

// BySystemAnalysisField orders the results by system_analysis field.
func BySystemAnalysisField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newSystemAnalysisStep(), sql.OrderByField(field, opts...))
	}
}

// ByComponentRelationshipField orders the results by component_relationship field.
func ByComponentRelationshipField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newComponentRelationshipStep(), sql.OrderByField(field, opts...))
	}
}

// ByControlsCount orders the results by controls count.
func ByControlsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newControlsStep(), opts...)
	}
}

// ByControls orders the results by controls terms.
func ByControls(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newControlsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// BySignalsCount orders the results by signals count.
func BySignalsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newSignalsStep(), opts...)
	}
}

// BySignals orders the results by signals terms.
func BySignals(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newSignalsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByControlActionsCount orders the results by control_actions count.
func ByControlActionsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newControlActionsStep(), opts...)
	}
}

// ByControlActions orders the results by control_actions terms.
func ByControlActions(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newControlActionsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByFeedbackSignalsCount orders the results by feedback_signals count.
func ByFeedbackSignalsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newFeedbackSignalsStep(), opts...)
	}
}

// ByFeedbackSignals orders the results by feedback_signals terms.
func ByFeedbackSignals(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newFeedbackSignalsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}
func newSystemAnalysisStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(SystemAnalysisInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, false, SystemAnalysisTable, SystemAnalysisColumn),
	)
}
func newComponentRelationshipStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(ComponentRelationshipInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, false, ComponentRelationshipTable, ComponentRelationshipColumn),
	)
}
func newControlsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(ControlsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, false, ControlsTable, ControlsPrimaryKey...),
	)
}
func newSignalsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(SignalsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, false, SignalsTable, SignalsPrimaryKey...),
	)
}
func newControlActionsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(ControlActionsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, true, ControlActionsTable, ControlActionsColumn),
	)
}
func newFeedbackSignalsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(FeedbackSignalsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, true, FeedbackSignalsTable, FeedbackSignalsColumn),
	)
}
