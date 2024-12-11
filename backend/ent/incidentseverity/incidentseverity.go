// Code generated by ent, DO NOT EDIT.

package incidentseverity

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
)

const (
	// Label holds the string label denoting the incidentseverity type in the database.
	Label = "incident_severity"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldArchiveTime holds the string denoting the archive_time field in the database.
	FieldArchiveTime = "archive_time"
	// FieldName holds the string denoting the name field in the database.
	FieldName = "name"
	// FieldColor holds the string denoting the color field in the database.
	FieldColor = "color"
	// FieldDescription holds the string denoting the description field in the database.
	FieldDescription = "description"
	// EdgeIncidents holds the string denoting the incidents edge name in mutations.
	EdgeIncidents = "incidents"
	// EdgeDebriefQuestions holds the string denoting the debrief_questions edge name in mutations.
	EdgeDebriefQuestions = "debrief_questions"
	// Table holds the table name of the incidentseverity in the database.
	Table = "incident_severities"
	// IncidentsTable is the table that holds the incidents relation/edge.
	IncidentsTable = "incidents"
	// IncidentsInverseTable is the table name for the Incident entity.
	// It exists in this package in order to avoid circular dependency with the "incident" package.
	IncidentsInverseTable = "incidents"
	// IncidentsColumn is the table column denoting the incidents relation/edge.
	IncidentsColumn = "severity_id"
	// DebriefQuestionsTable is the table that holds the debrief_questions relation/edge. The primary key declared below.
	DebriefQuestionsTable = "incident_debrief_question_incident_severities"
	// DebriefQuestionsInverseTable is the table name for the IncidentDebriefQuestion entity.
	// It exists in this package in order to avoid circular dependency with the "incidentdebriefquestion" package.
	DebriefQuestionsInverseTable = "incident_debrief_questions"
)

// Columns holds all SQL columns for incidentseverity fields.
var Columns = []string{
	FieldID,
	FieldArchiveTime,
	FieldName,
	FieldColor,
	FieldDescription,
}

var (
	// DebriefQuestionsPrimaryKey and DebriefQuestionsColumn2 are the table columns denoting the
	// primary key for the debrief_questions relation (M2M).
	DebriefQuestionsPrimaryKey = []string{"incident_debrief_question_id", "incident_severity_id"}
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

// Note that the variables below are initialized by the runtime
// package on the initialization of the application. Therefore,
// it should be imported in the main as follows:
//
//	import _ "github.com/twohundreds/rezible/ent/runtime"
var (
	Hooks        [1]ent.Hook
	Interceptors [1]ent.Interceptor
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() uuid.UUID
)

// OrderOption defines the ordering options for the IncidentSeverity queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByArchiveTime orders the results by the archive_time field.
func ByArchiveTime(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldArchiveTime, opts...).ToFunc()
}

// ByName orders the results by the name field.
func ByName(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldName, opts...).ToFunc()
}

// ByColor orders the results by the color field.
func ByColor(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldColor, opts...).ToFunc()
}

// ByDescription orders the results by the description field.
func ByDescription(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDescription, opts...).ToFunc()
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

// ByDebriefQuestionsCount orders the results by debrief_questions count.
func ByDebriefQuestionsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newDebriefQuestionsStep(), opts...)
	}
}

// ByDebriefQuestions orders the results by debrief_questions terms.
func ByDebriefQuestions(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newDebriefQuestionsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}
func newIncidentsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(IncidentsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, true, IncidentsTable, IncidentsColumn),
	)
}
func newDebriefQuestionsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(DebriefQuestionsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, true, DebriefQuestionsTable, DebriefQuestionsPrimaryKey...),
	)
}
