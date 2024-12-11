// Code generated by ent, DO NOT EDIT.

package service

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
)

const (
	// Label holds the string label denoting the service type in the database.
	Label = "service"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldSlug holds the string denoting the slug field in the database.
	FieldSlug = "slug"
	// FieldName holds the string denoting the name field in the database.
	FieldName = "name"
	// EdgeIncidents holds the string denoting the incidents edge name in mutations.
	EdgeIncidents = "incidents"
	// EdgeOwnerTeam holds the string denoting the owner_team edge name in mutations.
	EdgeOwnerTeam = "owner_team"
	// Table holds the table name of the service in the database.
	Table = "services"
	// IncidentsTable is the table that holds the incidents relation/edge.
	IncidentsTable = "incident_resource_impacts"
	// IncidentsInverseTable is the table name for the IncidentResourceImpact entity.
	// It exists in this package in order to avoid circular dependency with the "incidentresourceimpact" package.
	IncidentsInverseTable = "incident_resource_impacts"
	// IncidentsColumn is the table column denoting the incidents relation/edge.
	IncidentsColumn = "service_id"
	// OwnerTeamTable is the table that holds the owner_team relation/edge.
	OwnerTeamTable = "services"
	// OwnerTeamInverseTable is the table name for the Team entity.
	// It exists in this package in order to avoid circular dependency with the "team" package.
	OwnerTeamInverseTable = "teams"
	// OwnerTeamColumn is the table column denoting the owner_team relation/edge.
	OwnerTeamColumn = "service_owner_team"
)

// Columns holds all SQL columns for service fields.
var Columns = []string{
	FieldID,
	FieldSlug,
	FieldName,
}

// ForeignKeys holds the SQL foreign-keys that are owned by the "services"
// table and are not defined as standalone fields in the schema.
var ForeignKeys = []string{
	"incident_event_services",
	"service_owner_team",
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

// OrderOption defines the ordering options for the Service queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// BySlug orders the results by the slug field.
func BySlug(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldSlug, opts...).ToFunc()
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

// ByOwnerTeamField orders the results by owner_team field.
func ByOwnerTeamField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newOwnerTeamStep(), sql.OrderByField(field, opts...))
	}
}
func newIncidentsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(IncidentsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, IncidentsTable, IncidentsColumn),
	)
}
func newOwnerTeamStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(OwnerTeamInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, false, OwnerTeamTable, OwnerTeamColumn),
	)
}
