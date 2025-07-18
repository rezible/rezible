// Code generated by ent, DO NOT EDIT.

package playbook

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
)

const (
	// Label holds the string label denoting the playbook type in the database.
	Label = "playbook"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldTitle holds the string denoting the title field in the database.
	FieldTitle = "title"
	// FieldProviderID holds the string denoting the provider_id field in the database.
	FieldProviderID = "provider_id"
	// FieldContent holds the string denoting the content field in the database.
	FieldContent = "content"
	// EdgeAlerts holds the string denoting the alerts edge name in mutations.
	EdgeAlerts = "alerts"
	// Table holds the table name of the playbook in the database.
	Table = "playbooks"
	// AlertsTable is the table that holds the alerts relation/edge. The primary key declared below.
	AlertsTable = "playbook_alerts"
	// AlertsInverseTable is the table name for the Alert entity.
	// It exists in this package in order to avoid circular dependency with the "alert" package.
	AlertsInverseTable = "alerts"
)

// Columns holds all SQL columns for playbook fields.
var Columns = []string{
	FieldID,
	FieldTitle,
	FieldProviderID,
	FieldContent,
}

var (
	// AlertsPrimaryKey and AlertsColumn2 are the table columns denoting the
	// primary key for the alerts relation (M2M).
	AlertsPrimaryKey = []string{"playbook_id", "alert_id"}
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
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() uuid.UUID
)

// OrderOption defines the ordering options for the Playbook queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByTitle orders the results by the title field.
func ByTitle(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldTitle, opts...).ToFunc()
}

// ByProviderID orders the results by the provider_id field.
func ByProviderID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldProviderID, opts...).ToFunc()
}

// ByAlertsCount orders the results by alerts count.
func ByAlertsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newAlertsStep(), opts...)
	}
}

// ByAlerts orders the results by alerts terms.
func ByAlerts(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newAlertsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}
func newAlertsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(AlertsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, false, AlertsTable, AlertsPrimaryKey...),
	)
}
