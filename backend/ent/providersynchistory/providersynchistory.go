// Code generated by ent, DO NOT EDIT.

package providersynchistory

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
)

const (
	// Label holds the string label denoting the providersynchistory type in the database.
	Label = "provider_sync_history"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldDataType holds the string denoting the data_type field in the database.
	FieldDataType = "data_type"
	// FieldStartedAt holds the string denoting the started_at field in the database.
	FieldStartedAt = "started_at"
	// FieldFinishedAt holds the string denoting the finished_at field in the database.
	FieldFinishedAt = "finished_at"
	// FieldNumMutations holds the string denoting the num_mutations field in the database.
	FieldNumMutations = "num_mutations"
	// Table holds the table name of the providersynchistory in the database.
	Table = "provider_sync_histories"
)

// Columns holds all SQL columns for providersynchistory fields.
var Columns = []string{
	FieldID,
	FieldDataType,
	FieldStartedAt,
	FieldFinishedAt,
	FieldNumMutations,
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
	// DefaultStartedAt holds the default value on creation for the "started_at" field.
	DefaultStartedAt func() time.Time
	// DefaultFinishedAt holds the default value on creation for the "finished_at" field.
	DefaultFinishedAt func() time.Time
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() uuid.UUID
)

// OrderOption defines the ordering options for the ProviderSyncHistory queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByDataType orders the results by the data_type field.
func ByDataType(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDataType, opts...).ToFunc()
}

// ByStartedAt orders the results by the started_at field.
func ByStartedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStartedAt, opts...).ToFunc()
}

// ByFinishedAt orders the results by the finished_at field.
func ByFinishedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldFinishedAt, opts...).ToFunc()
}

// ByNumMutations orders the results by the num_mutations field.
func ByNumMutations(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldNumMutations, opts...).ToFunc()
}