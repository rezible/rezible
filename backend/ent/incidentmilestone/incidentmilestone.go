// Code generated by ent, DO NOT EDIT.

package incidentmilestone

import (
	"fmt"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
)

const (
	// Label holds the string label denoting the incidentmilestone type in the database.
	Label = "incident_milestone"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldIncidentID holds the string denoting the incident_id field in the database.
	FieldIncidentID = "incident_id"
	// FieldType holds the string denoting the type field in the database.
	FieldType = "type"
	// FieldTime holds the string denoting the time field in the database.
	FieldTime = "time"
	// EdgeIncident holds the string denoting the incident edge name in mutations.
	EdgeIncident = "incident"
	// Table holds the table name of the incidentmilestone in the database.
	Table = "incident_milestones"
	// IncidentTable is the table that holds the incident relation/edge.
	IncidentTable = "incident_milestones"
	// IncidentInverseTable is the table name for the Incident entity.
	// It exists in this package in order to avoid circular dependency with the "incident" package.
	IncidentInverseTable = "incidents"
	// IncidentColumn is the table column denoting the incident relation/edge.
	IncidentColumn = "incident_id"
)

// Columns holds all SQL columns for incidentmilestone fields.
var Columns = []string{
	FieldID,
	FieldIncidentID,
	FieldType,
	FieldTime,
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

// Type defines the type for the "type" enum field.
type Type string

// Type values.
const (
	TypeImpact        Type = "impact"
	TypeDetected      Type = "detected"
	TypeInvestigating Type = "investigating"
	TypeMitigated     Type = "mitigated"
	TypeResolved      Type = "resolved"
)

func (_type Type) String() string {
	return string(_type)
}

// TypeValidator is a validator for the "type" field enum values. It is called by the builders before save.
func TypeValidator(_type Type) error {
	switch _type {
	case TypeImpact, TypeDetected, TypeInvestigating, TypeMitigated, TypeResolved:
		return nil
	default:
		return fmt.Errorf("incidentmilestone: invalid enum value for type field: %q", _type)
	}
}

// OrderOption defines the ordering options for the IncidentMilestone queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByIncidentID orders the results by the incident_id field.
func ByIncidentID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldIncidentID, opts...).ToFunc()
}

// ByType orders the results by the type field.
func ByType(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldType, opts...).ToFunc()
}

// ByTime orders the results by the time field.
func ByTime(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldTime, opts...).ToFunc()
}

// ByIncidentField orders the results by incident field.
func ByIncidentField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newIncidentStep(), sql.OrderByField(field, opts...))
	}
}
func newIncidentStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(IncidentInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, IncidentTable, IncidentColumn),
	)
}
