// Code generated by ent, DO NOT EDIT.

package incidentrole

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
)

const (
	// Label holds the string label denoting the incidentrole type in the database.
	Label = "incident_role"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldArchiveTime holds the string denoting the archive_time field in the database.
	FieldArchiveTime = "archive_time"
	// FieldName holds the string denoting the name field in the database.
	FieldName = "name"
	// FieldProviderID holds the string denoting the provider_id field in the database.
	FieldProviderID = "provider_id"
	// FieldRequired holds the string denoting the required field in the database.
	FieldRequired = "required"
	// EdgeAssignments holds the string denoting the assignments edge name in mutations.
	EdgeAssignments = "assignments"
	// EdgeDebriefQuestions holds the string denoting the debrief_questions edge name in mutations.
	EdgeDebriefQuestions = "debrief_questions"
	// Table holds the table name of the incidentrole in the database.
	Table = "incident_roles"
	// AssignmentsTable is the table that holds the assignments relation/edge.
	AssignmentsTable = "incident_role_assignments"
	// AssignmentsInverseTable is the table name for the IncidentRoleAssignment entity.
	// It exists in this package in order to avoid circular dependency with the "incidentroleassignment" package.
	AssignmentsInverseTable = "incident_role_assignments"
	// AssignmentsColumn is the table column denoting the assignments relation/edge.
	AssignmentsColumn = "role_id"
	// DebriefQuestionsTable is the table that holds the debrief_questions relation/edge. The primary key declared below.
	DebriefQuestionsTable = "incident_debrief_question_incident_roles"
	// DebriefQuestionsInverseTable is the table name for the IncidentDebriefQuestion entity.
	// It exists in this package in order to avoid circular dependency with the "incidentdebriefquestion" package.
	DebriefQuestionsInverseTable = "incident_debrief_questions"
)

// Columns holds all SQL columns for incidentrole fields.
var Columns = []string{
	FieldID,
	FieldArchiveTime,
	FieldName,
	FieldProviderID,
	FieldRequired,
}

var (
	// DebriefQuestionsPrimaryKey and DebriefQuestionsColumn2 are the table columns denoting the
	// primary key for the debrief_questions relation (M2M).
	DebriefQuestionsPrimaryKey = []string{"incident_debrief_question_id", "incident_role_id"}
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
//	import _ "github.com/rezible/rezible/ent/runtime"
var (
	Hooks        [1]ent.Hook
	Interceptors [1]ent.Interceptor
	// DefaultRequired holds the default value on creation for the "required" field.
	DefaultRequired bool
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() uuid.UUID
)

// OrderOption defines the ordering options for the IncidentRole queries.
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

// ByProviderID orders the results by the provider_id field.
func ByProviderID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldProviderID, opts...).ToFunc()
}

// ByRequired orders the results by the required field.
func ByRequired(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldRequired, opts...).ToFunc()
}

// ByAssignmentsCount orders the results by assignments count.
func ByAssignmentsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newAssignmentsStep(), opts...)
	}
}

// ByAssignments orders the results by assignments terms.
func ByAssignments(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newAssignmentsStep(), append([]sql.OrderTerm{term}, terms...)...)
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
func newAssignmentsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(AssignmentsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, true, AssignmentsTable, AssignmentsColumn),
	)
}
func newDebriefQuestionsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(DebriefQuestionsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, true, DebriefQuestionsTable, DebriefQuestionsPrimaryKey...),
	)
}
