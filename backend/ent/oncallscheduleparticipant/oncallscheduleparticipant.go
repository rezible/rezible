// Code generated by ent, DO NOT EDIT.

package oncallscheduleparticipant

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
)

const (
	// Label holds the string label denoting the oncallscheduleparticipant type in the database.
	Label = "oncall_schedule_participant"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldScheduleID holds the string denoting the schedule_id field in the database.
	FieldScheduleID = "schedule_id"
	// FieldUserID holds the string denoting the user_id field in the database.
	FieldUserID = "user_id"
	// FieldIndex holds the string denoting the index field in the database.
	FieldIndex = "index"
	// EdgeSchedule holds the string denoting the schedule edge name in mutations.
	EdgeSchedule = "schedule"
	// EdgeUser holds the string denoting the user edge name in mutations.
	EdgeUser = "user"
	// Table holds the table name of the oncallscheduleparticipant in the database.
	Table = "oncall_schedule_participants"
	// ScheduleTable is the table that holds the schedule relation/edge.
	ScheduleTable = "oncall_schedule_participants"
	// ScheduleInverseTable is the table name for the OncallSchedule entity.
	// It exists in this package in order to avoid circular dependency with the "oncallschedule" package.
	ScheduleInverseTable = "oncall_schedules"
	// ScheduleColumn is the table column denoting the schedule relation/edge.
	ScheduleColumn = "schedule_id"
	// UserTable is the table that holds the user relation/edge.
	UserTable = "oncall_schedule_participants"
	// UserInverseTable is the table name for the User entity.
	// It exists in this package in order to avoid circular dependency with the "user" package.
	UserInverseTable = "users"
	// UserColumn is the table column denoting the user relation/edge.
	UserColumn = "user_id"
)

// Columns holds all SQL columns for oncallscheduleparticipant fields.
var Columns = []string{
	FieldID,
	FieldScheduleID,
	FieldUserID,
	FieldIndex,
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

// OrderOption defines the ordering options for the OncallScheduleParticipant queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByScheduleID orders the results by the schedule_id field.
func ByScheduleID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldScheduleID, opts...).ToFunc()
}

// ByUserID orders the results by the user_id field.
func ByUserID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldUserID, opts...).ToFunc()
}

// ByIndex orders the results by the index field.
func ByIndex(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldIndex, opts...).ToFunc()
}

// ByScheduleField orders the results by schedule field.
func ByScheduleField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newScheduleStep(), sql.OrderByField(field, opts...))
	}
}

// ByUserField orders the results by user field.
func ByUserField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newUserStep(), sql.OrderByField(field, opts...))
	}
}
func newScheduleStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(ScheduleInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, ScheduleTable, ScheduleColumn),
	)
}
func newUserStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(UserInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, false, UserTable, UserColumn),
	)
}
