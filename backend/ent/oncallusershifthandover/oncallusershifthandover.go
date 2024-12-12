// Code generated by ent, DO NOT EDIT.

package oncallusershifthandover

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
)

const (
	// Label holds the string label denoting the oncallusershifthandover type in the database.
	Label = "oncall_user_shift_handover"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldShiftID holds the string denoting the shift_id field in the database.
	FieldShiftID = "shift_id"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// FieldReminderSent holds the string denoting the reminder_sent field in the database.
	FieldReminderSent = "reminder_sent"
	// FieldUpdatedAt holds the string denoting the updated_at field in the database.
	FieldUpdatedAt = "updated_at"
	// FieldSentAt holds the string denoting the sent_at field in the database.
	FieldSentAt = "sent_at"
	// FieldContents holds the string denoting the contents field in the database.
	FieldContents = "contents"
	// EdgeShift holds the string denoting the shift edge name in mutations.
	EdgeShift = "shift"
	// Table holds the table name of the oncallusershifthandover in the database.
	Table = "oncall_user_shift_handovers"
	// ShiftTable is the table that holds the shift relation/edge.
	ShiftTable = "oncall_user_shift_handovers"
	// ShiftInverseTable is the table name for the OncallUserShift entity.
	// It exists in this package in order to avoid circular dependency with the "oncallusershift" package.
	ShiftInverseTable = "oncall_user_shifts"
	// ShiftColumn is the table column denoting the shift relation/edge.
	ShiftColumn = "shift_id"
)

// Columns holds all SQL columns for oncallusershifthandover fields.
var Columns = []string{
	FieldID,
	FieldShiftID,
	FieldCreatedAt,
	FieldReminderSent,
	FieldUpdatedAt,
	FieldSentAt,
	FieldContents,
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
	// DefaultReminderSent holds the default value on creation for the "reminder_sent" field.
	DefaultReminderSent bool
	// DefaultUpdatedAt holds the default value on creation for the "updated_at" field.
	DefaultUpdatedAt func() time.Time
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() uuid.UUID
)

// OrderOption defines the ordering options for the OncallUserShiftHandover queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByShiftID orders the results by the shift_id field.
func ByShiftID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldShiftID, opts...).ToFunc()
}

// ByCreatedAt orders the results by the created_at field.
func ByCreatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCreatedAt, opts...).ToFunc()
}

// ByReminderSent orders the results by the reminder_sent field.
func ByReminderSent(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldReminderSent, opts...).ToFunc()
}

// ByUpdatedAt orders the results by the updated_at field.
func ByUpdatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldUpdatedAt, opts...).ToFunc()
}

// BySentAt orders the results by the sent_at field.
func BySentAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldSentAt, opts...).ToFunc()
}

// ByShiftField orders the results by shift field.
func ByShiftField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newShiftStep(), sql.OrderByField(field, opts...))
	}
}
func newShiftStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(ShiftInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2O, true, ShiftTable, ShiftColumn),
	)
}
