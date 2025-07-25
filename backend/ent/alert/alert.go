// Code generated by ent, DO NOT EDIT.

package alert

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
)

const (
	// Label holds the string label denoting the alert type in the database.
	Label = "alert"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldTitle holds the string denoting the title field in the database.
	FieldTitle = "title"
	// FieldProviderID holds the string denoting the provider_id field in the database.
	FieldProviderID = "provider_id"
	// EdgeMetrics holds the string denoting the metrics edge name in mutations.
	EdgeMetrics = "metrics"
	// EdgePlaybooks holds the string denoting the playbooks edge name in mutations.
	EdgePlaybooks = "playbooks"
	// EdgeInstances holds the string denoting the instances edge name in mutations.
	EdgeInstances = "instances"
	// Table holds the table name of the alert in the database.
	Table = "alerts"
	// MetricsTable is the table that holds the metrics relation/edge.
	MetricsTable = "alert_metrics"
	// MetricsInverseTable is the table name for the AlertMetrics entity.
	// It exists in this package in order to avoid circular dependency with the "alertmetrics" package.
	MetricsInverseTable = "alert_metrics"
	// MetricsColumn is the table column denoting the metrics relation/edge.
	MetricsColumn = "alert_id"
	// PlaybooksTable is the table that holds the playbooks relation/edge. The primary key declared below.
	PlaybooksTable = "playbook_alerts"
	// PlaybooksInverseTable is the table name for the Playbook entity.
	// It exists in this package in order to avoid circular dependency with the "playbook" package.
	PlaybooksInverseTable = "playbooks"
	// InstancesTable is the table that holds the instances relation/edge.
	InstancesTable = "oncall_events"
	// InstancesInverseTable is the table name for the OncallEvent entity.
	// It exists in this package in order to avoid circular dependency with the "oncallevent" package.
	InstancesInverseTable = "oncall_events"
	// InstancesColumn is the table column denoting the instances relation/edge.
	InstancesColumn = "alert_id"
)

// Columns holds all SQL columns for alert fields.
var Columns = []string{
	FieldID,
	FieldTitle,
	FieldProviderID,
}

var (
	// PlaybooksPrimaryKey and PlaybooksColumn2 are the table columns denoting the
	// primary key for the playbooks relation (M2M).
	PlaybooksPrimaryKey = []string{"playbook_id", "alert_id"}
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

// OrderOption defines the ordering options for the Alert queries.
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

// ByMetricsCount orders the results by metrics count.
func ByMetricsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newMetricsStep(), opts...)
	}
}

// ByMetrics orders the results by metrics terms.
func ByMetrics(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newMetricsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByPlaybooksCount orders the results by playbooks count.
func ByPlaybooksCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newPlaybooksStep(), opts...)
	}
}

// ByPlaybooks orders the results by playbooks terms.
func ByPlaybooks(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newPlaybooksStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByInstancesCount orders the results by instances count.
func ByInstancesCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newInstancesStep(), opts...)
	}
}

// ByInstances orders the results by instances terms.
func ByInstances(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newInstancesStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}
func newMetricsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(MetricsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, true, MetricsTable, MetricsColumn),
	)
}
func newPlaybooksStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(PlaybooksInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, true, PlaybooksTable, PlaybooksPrimaryKey...),
	)
}
func newInstancesStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(InstancesInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, true, InstancesTable, InstancesColumn),
	)
}
