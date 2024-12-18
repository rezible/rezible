// Code generated by ent, DO NOT EDIT.

package functionality

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
)

const (
	// Label holds the string label denoting the functionality type in the database.
	Label = "functionality"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldName holds the string denoting the name field in the database.
	FieldName = "name"
	// EdgeIncidents holds the string denoting the incidents edge name in mutations.
	EdgeIncidents = "incidents"
	// Table holds the table name of the functionality in the database.
	Table = "functionalities"
	// IncidentsTable is the table that holds the incidents relation/edge.
	IncidentsTable = "incident_resource_impacts"
	// IncidentsInverseTable is the table name for the IncidentResourceImpact entity.
	// It exists in this package in order to avoid circular dependency with the "incidentresourceimpact" package.
	IncidentsInverseTable = "incident_resource_impacts"
	// IncidentsColumn is the table column denoting the incidents relation/edge.
	IncidentsColumn = "functionality_id"
)

// Columns holds all SQL columns for functionality fields.
var Columns = []string{
	FieldID,
	FieldName,
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
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() uuid.UUID
)

// OrderOption defines the ordering options for the Functionality queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByName orders the results by the name field.
func ByName(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldName, opts...).ToFunc()
}

// ByIncidentsCount orders the results by incidents count.
func ByIncidentsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newIncidentsStep(), opts...)
	}
}

// ByIncidents orders the results by incidents terms.
func ByIncidents(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newIncidentsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}
func newIncidentsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(IncidentsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, IncidentsTable, IncidentsColumn),
	)
}
