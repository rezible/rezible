// Code generated by ent, DO NOT EDIT.

package incidentdebriefsuggestion

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
)

const (
	// Label holds the string label denoting the incidentdebriefsuggestion type in the database.
	Label = "incident_debrief_suggestion"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldContent holds the string denoting the content field in the database.
	FieldContent = "content"
	// EdgeDebrief holds the string denoting the debrief edge name in mutations.
	EdgeDebrief = "debrief"
	// Table holds the table name of the incidentdebriefsuggestion in the database.
	Table = "incident_debrief_suggestions"
	// DebriefTable is the table that holds the debrief relation/edge.
	DebriefTable = "incident_debrief_suggestions"
	// DebriefInverseTable is the table name for the IncidentDebrief entity.
	// It exists in this package in order to avoid circular dependency with the "incidentdebrief" package.
	DebriefInverseTable = "incident_debriefs"
	// DebriefColumn is the table column denoting the debrief relation/edge.
	DebriefColumn = "incident_debrief_suggestions"
)

// Columns holds all SQL columns for incidentdebriefsuggestion fields.
var Columns = []string{
	FieldID,
	FieldContent,
}

// ForeignKeys holds the SQL foreign-keys that are owned by the "incident_debrief_suggestions"
// table and are not defined as standalone fields in the schema.
var ForeignKeys = []string{
	"incident_debrief_suggestions",
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
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() uuid.UUID
)

// OrderOption defines the ordering options for the IncidentDebriefSuggestion queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByContent orders the results by the content field.
func ByContent(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldContent, opts...).ToFunc()
}

// ByDebriefField orders the results by debrief field.
func ByDebriefField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newDebriefStep(), sql.OrderByField(field, opts...))
	}
}
func newDebriefStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(DebriefInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, DebriefTable, DebriefColumn),
	)
}
