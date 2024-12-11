// Code generated by ent, DO NOT EDIT.

package incidentdebriefmessage

import (
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
)

const (
	// Label holds the string label denoting the incidentdebriefmessage type in the database.
	Label = "incident_debrief_message"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldDebriefID holds the string denoting the debrief_id field in the database.
	FieldDebriefID = "debrief_id"
	// FieldQuestionID holds the string denoting the question_id field in the database.
	FieldQuestionID = "question_id"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// FieldType holds the string denoting the type field in the database.
	FieldType = "type"
	// FieldRequestedTool holds the string denoting the requested_tool field in the database.
	FieldRequestedTool = "requested_tool"
	// FieldBody holds the string denoting the body field in the database.
	FieldBody = "body"
	// EdgeDebrief holds the string denoting the debrief edge name in mutations.
	EdgeDebrief = "debrief"
	// EdgeFromQuestion holds the string denoting the from_question edge name in mutations.
	EdgeFromQuestion = "from_question"
	// Table holds the table name of the incidentdebriefmessage in the database.
	Table = "incident_debrief_messages"
	// DebriefTable is the table that holds the debrief relation/edge.
	DebriefTable = "incident_debrief_messages"
	// DebriefInverseTable is the table name for the IncidentDebrief entity.
	// It exists in this package in order to avoid circular dependency with the "incidentdebrief" package.
	DebriefInverseTable = "incident_debriefs"
	// DebriefColumn is the table column denoting the debrief relation/edge.
	DebriefColumn = "debrief_id"
	// FromQuestionTable is the table that holds the from_question relation/edge.
	FromQuestionTable = "incident_debrief_messages"
	// FromQuestionInverseTable is the table name for the IncidentDebriefQuestion entity.
	// It exists in this package in order to avoid circular dependency with the "incidentdebriefquestion" package.
	FromQuestionInverseTable = "incident_debrief_questions"
	// FromQuestionColumn is the table column denoting the from_question relation/edge.
	FromQuestionColumn = "question_id"
)

// Columns holds all SQL columns for incidentdebriefmessage fields.
var Columns = []string{
	FieldID,
	FieldDebriefID,
	FieldQuestionID,
	FieldCreatedAt,
	FieldType,
	FieldRequestedTool,
	FieldBody,
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
	// DefaultCreatedAt holds the default value on creation for the "created_at" field.
	DefaultCreatedAt func() time.Time
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() uuid.UUID
)

// Type defines the type for the "type" enum field.
type Type string

// Type values.
const (
	TypeUser      Type = "user"
	TypeAssistant Type = "assistant"
	TypeQuestion  Type = "question"
)

func (_type Type) String() string {
	return string(_type)
}

// TypeValidator is a validator for the "type" field enum values. It is called by the builders before save.
func TypeValidator(_type Type) error {
	switch _type {
	case TypeUser, TypeAssistant, TypeQuestion:
		return nil
	default:
		return fmt.Errorf("incidentdebriefmessage: invalid enum value for type field: %q", _type)
	}
}

// RequestedTool defines the type for the "requested_tool" enum field.
type RequestedTool string

// RequestedTool values.
const (
	RequestedToolRating RequestedTool = "rating"
)

func (rt RequestedTool) String() string {
	return string(rt)
}

// RequestedToolValidator is a validator for the "requested_tool" field enum values. It is called by the builders before save.
func RequestedToolValidator(rt RequestedTool) error {
	switch rt {
	case RequestedToolRating:
		return nil
	default:
		return fmt.Errorf("incidentdebriefmessage: invalid enum value for requested_tool field: %q", rt)
	}
}

// OrderOption defines the ordering options for the IncidentDebriefMessage queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByDebriefID orders the results by the debrief_id field.
func ByDebriefID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDebriefID, opts...).ToFunc()
}

// ByQuestionID orders the results by the question_id field.
func ByQuestionID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldQuestionID, opts...).ToFunc()
}

// ByCreatedAt orders the results by the created_at field.
func ByCreatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCreatedAt, opts...).ToFunc()
}

// ByType orders the results by the type field.
func ByType(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldType, opts...).ToFunc()
}

// ByRequestedTool orders the results by the requested_tool field.
func ByRequestedTool(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldRequestedTool, opts...).ToFunc()
}

// ByBody orders the results by the body field.
func ByBody(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldBody, opts...).ToFunc()
}

// ByDebriefField orders the results by debrief field.
func ByDebriefField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newDebriefStep(), sql.OrderByField(field, opts...))
	}
}

// ByFromQuestionField orders the results by from_question field.
func ByFromQuestionField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newFromQuestionStep(), sql.OrderByField(field, opts...))
	}
}
func newDebriefStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(DebriefInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, DebriefTable, DebriefColumn),
	)
}
func newFromQuestionStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(FromQuestionInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, false, FromQuestionTable, FromQuestionColumn),
	)
}