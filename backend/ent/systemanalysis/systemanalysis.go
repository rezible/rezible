// Code generated by ent, DO NOT EDIT.

package systemanalysis

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
)

const (
	// Label holds the string label denoting the systemanalysis type in the database.
	Label = "system_analysis"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// FieldUpdatedAt holds the string denoting the updated_at field in the database.
	FieldUpdatedAt = "updated_at"
	// EdgeRetrospective holds the string denoting the retrospective edge name in mutations.
	EdgeRetrospective = "retrospective"
	// EdgeComponents holds the string denoting the components edge name in mutations.
	EdgeComponents = "components"
	// EdgeRelationships holds the string denoting the relationships edge name in mutations.
	EdgeRelationships = "relationships"
	// EdgeAnalysisComponents holds the string denoting the analysis_components edge name in mutations.
	EdgeAnalysisComponents = "analysis_components"
	// Table holds the table name of the systemanalysis in the database.
	Table = "system_analyses"
	// RetrospectiveTable is the table that holds the retrospective relation/edge.
	RetrospectiveTable = "retrospectives"
	// RetrospectiveInverseTable is the table name for the Retrospective entity.
	// It exists in this package in order to avoid circular dependency with the "retrospective" package.
	RetrospectiveInverseTable = "retrospectives"
	// RetrospectiveColumn is the table column denoting the retrospective relation/edge.
	RetrospectiveColumn = "system_analysis_id"
	// ComponentsTable is the table that holds the components relation/edge. The primary key declared below.
	ComponentsTable = "system_analysis_components"
	// ComponentsInverseTable is the table name for the SystemComponent entity.
	// It exists in this package in order to avoid circular dependency with the "systemcomponent" package.
	ComponentsInverseTable = "system_components"
	// RelationshipsTable is the table that holds the relationships relation/edge.
	RelationshipsTable = "system_analysis_relationships"
	// RelationshipsInverseTable is the table name for the SystemAnalysisRelationship entity.
	// It exists in this package in order to avoid circular dependency with the "systemanalysisrelationship" package.
	RelationshipsInverseTable = "system_analysis_relationships"
	// RelationshipsColumn is the table column denoting the relationships relation/edge.
	RelationshipsColumn = "analysis_id"
	// AnalysisComponentsTable is the table that holds the analysis_components relation/edge.
	AnalysisComponentsTable = "system_analysis_components"
	// AnalysisComponentsInverseTable is the table name for the SystemAnalysisComponent entity.
	// It exists in this package in order to avoid circular dependency with the "systemanalysiscomponent" package.
	AnalysisComponentsInverseTable = "system_analysis_components"
	// AnalysisComponentsColumn is the table column denoting the analysis_components relation/edge.
	AnalysisComponentsColumn = "analysis_id"
)

// Columns holds all SQL columns for systemanalysis fields.
var Columns = []string{
	FieldID,
	FieldCreatedAt,
	FieldUpdatedAt,
}

var (
	// ComponentsPrimaryKey and ComponentsColumn2 are the table columns denoting the
	// primary key for the components relation (M2M).
	ComponentsPrimaryKey = []string{"component_id", "analysis_id"}
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
	// DefaultUpdatedAt holds the default value on creation for the "updated_at" field.
	DefaultUpdatedAt func() time.Time
	// UpdateDefaultUpdatedAt holds the default value on update for the "updated_at" field.
	UpdateDefaultUpdatedAt func() time.Time
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() uuid.UUID
)

// OrderOption defines the ordering options for the SystemAnalysis queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByCreatedAt orders the results by the created_at field.
func ByCreatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCreatedAt, opts...).ToFunc()
}

// ByUpdatedAt orders the results by the updated_at field.
func ByUpdatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldUpdatedAt, opts...).ToFunc()
}

// ByRetrospectiveField orders the results by retrospective field.
func ByRetrospectiveField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newRetrospectiveStep(), sql.OrderByField(field, opts...))
	}
}

// ByComponentsCount orders the results by components count.
func ByComponentsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newComponentsStep(), opts...)
	}
}

// ByComponents orders the results by components terms.
func ByComponents(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newComponentsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByRelationshipsCount orders the results by relationships count.
func ByRelationshipsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newRelationshipsStep(), opts...)
	}
}

// ByRelationships orders the results by relationships terms.
func ByRelationships(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newRelationshipsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByAnalysisComponentsCount orders the results by analysis_components count.
func ByAnalysisComponentsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newAnalysisComponentsStep(), opts...)
	}
}

// ByAnalysisComponents orders the results by analysis_components terms.
func ByAnalysisComponents(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newAnalysisComponentsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}
func newRetrospectiveStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(RetrospectiveInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2O, false, RetrospectiveTable, RetrospectiveColumn),
	)
}
func newComponentsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(ComponentsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, true, ComponentsTable, ComponentsPrimaryKey...),
	)
}
func newRelationshipsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(RelationshipsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, true, RelationshipsTable, RelationshipsColumn),
	)
}
func newAnalysisComponentsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(AnalysisComponentsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, true, AnalysisComponentsTable, AnalysisComponentsColumn),
	)
}
