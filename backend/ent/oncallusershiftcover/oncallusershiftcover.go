// Code generated by ent, DO NOT EDIT.

package oncallusershiftcover

import (
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
)

const (
	// Label holds the string label denoting the oncallusershiftcover type in the database.
	Label = "oncall_user_shift_cover"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldUserID holds the string denoting the user_id field in the database.
	FieldUserID = "user_id"
	// FieldShiftID holds the string denoting the shift_id field in the database.
	FieldShiftID = "shift_id"
	// FieldStartAt holds the string denoting the start_at field in the database.
	FieldStartAt = "start_at"
	// FieldEndAt holds the string denoting the end_at field in the database.
	FieldEndAt = "end_at"
	// FieldProviderID holds the string denoting the provider_id field in the database.
	FieldProviderID = "provider_id"
	// EdgeUser holds the string denoting the user edge name in mutations.
	EdgeUser = "user"
	// EdgeShift holds the string denoting the shift edge name in mutations.
	EdgeShift = "shift"
	// Table holds the table name of the oncallusershiftcover in the database.
	Table = "oncall_user_shift_covers"
	// UserTable is the table that holds the user relation/edge.
	UserTable = "oncall_user_shift_covers"
	// UserInverseTable is the table name for the User entity.
	// It exists in this package in order to avoid circular dependency with the "user" package.
	UserInverseTable = "users"
	// UserColumn is the table column denoting the user relation/edge.
	UserColumn = "user_id"
	// ShiftTable is the table that holds the shift relation/edge.
	ShiftTable = "oncall_user_shift_covers"
	// ShiftInverseTable is the table name for the OncallUserShift entity.
	// It exists in this package in order to avoid circular dependency with the "oncallusershift" package.
	ShiftInverseTable = "oncall_user_shifts"
	// ShiftColumn is the table column denoting the shift relation/edge.
	ShiftColumn = "shift_id"
)

// Columns holds all SQL columns for oncallusershiftcover fields.
var Columns = []string{
	FieldID,
	FieldUserID,
	FieldShiftID,
	FieldStartAt,
	FieldEndAt,
	FieldProviderID,
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

// OrderOption defines the ordering options for the OncallUserShiftCover queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByUserID orders the results by the user_id field.
func ByUserID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldUserID, opts...).ToFunc()
}

// ByShiftID orders the results by the shift_id field.
func ByShiftID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldShiftID, opts...).ToFunc()
}

// ByStartAt orders the results by the start_at field.
func ByStartAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStartAt, opts...).ToFunc()
}

// ByEndAt orders the results by the end_at field.
func ByEndAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldEndAt, opts...).ToFunc()
}

// ByProviderID orders the results by the provider_id field.
func ByProviderID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldProviderID, opts...).ToFunc()
}

// ByUserField orders the results by user field.
func ByUserField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newUserStep(), sql.OrderByField(field, opts...))
	}
}

// ByShiftField orders the results by shift field.
func ByShiftField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newShiftStep(), sql.OrderByField(field, opts...))
	}
}
func newUserStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(UserInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, false, UserTable, UserColumn),
	)
}
func newShiftStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(ShiftInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, ShiftTable, ShiftColumn),
	)
}
