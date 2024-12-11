package entpgx

import (
	"context"
	stdsql "database/sql"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// https://github.com/ent/ent/discussions/1797#discussioncomment-5111111

func NewPgxPoolDriver(pool *pgxpool.Pool) dialect.Driver {
	return &EntPgxpoolDriver{
		pool: pool,
	}
}

type EntPgxpoolDriver struct {
	pool *pgxpool.Pool
}

func (e *EntPgxpoolDriver) Exec(ctx context.Context, query string, args, result any) error {
	var _ stdsql.Result
	argv, ok := args.([]any)
	if !ok {
		return fmt.Errorf("dialect/sql: invalid type %T. expect []any for args", result)
	}
	switch result := result.(type) {
	case nil:
		if _, err := e.pool.Exec(ctx, query, argv...); err != nil {
			return err
		}
	case *sql.Result:
		commandTag, err := e.pool.Exec(ctx, query, argv...)
		if err != nil {
			return err
		}
		*result = execResult{rowsAffected: commandTag.RowsAffected()}
	default:
		return fmt.Errorf("dialect/sql: invalid type %T. expect *sql.Result", result)
	}
	return nil
}

func (e *EntPgxpoolDriver) Query(ctx context.Context, query string, args, v any) error {
	vr, ok := v.(*sql.Rows)
	if !ok {
		return fmt.Errorf("dialect/sql: invalid type %T. expect *sql.Rows", v)
	}
	argv, ok := args.([]any)
	if !ok {
		return fmt.Errorf("dialect/sql: invalid type %T. expect []any for args", args)
	}
	pgxRows, err := e.pool.Query(ctx, query, argv...)
	if err != nil {
		return err
	}
	columnScanner := &entPgxRows{pgxRows: pgxRows}
	*vr = sql.Rows{
		columnScanner,
	}
	return nil
}

func (e *EntPgxpoolDriver) ExecContext(ctx context.Context, query string, args ...any) (stdsql.Result, error) {
	commandTag, err := e.pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &execResult{rowsAffected: commandTag.RowsAffected()}, nil
}

func (e *EntPgxpoolDriver) Tx(ctx context.Context) (dialect.Tx, error) {
	return e.BeginTx(ctx, nil)
}

func (e *EntPgxpoolDriver) BeginTx(ctx context.Context, opts *sql.TxOptions) (dialect.Tx, error) {
	pgxOpts, err := getPgxTxOptions(opts)
	if err != nil {
		return nil, err
	}
	tx, err := e.pool.BeginTx(ctx, *pgxOpts)
	if err != nil {
		return nil, err
	}
	return &EntPgxPoolTx{
		tx: tx,
	}, nil
}

func getPgxTxOptions(opts *sql.TxOptions) (*pgx.TxOptions, error) {
	var pgxOpts pgx.TxOptions
	if opts == nil {
		return &pgxOpts, nil
	}
	switch opts.Isolation {
	case stdsql.LevelDefault:
	case stdsql.LevelReadUncommitted:
		pgxOpts.IsoLevel = pgx.ReadUncommitted
	case stdsql.LevelReadCommitted:
		pgxOpts.IsoLevel = pgx.ReadCommitted
	case stdsql.LevelRepeatableRead, stdsql.LevelSnapshot:
		pgxOpts.IsoLevel = pgx.RepeatableRead
	case stdsql.LevelSerializable:
		pgxOpts.IsoLevel = pgx.Serializable
	default:
		return nil, fmt.Errorf("unsupported isolation: %v", opts.Isolation)
	}

	if opts.ReadOnly {
		pgxOpts.AccessMode = pgx.ReadOnly
	}
	return &pgxOpts, nil
}

func (e *EntPgxpoolDriver) Close() error {
	e.pool.Close()
	return nil
}

func (e *EntPgxpoolDriver) Dialect() string {
	return dialect.Postgres
}

type EntPgxPoolTx struct {
	tx pgx.Tx
}

func (e *EntPgxPoolTx) Exec(ctx context.Context, query string, args, result any) error {
	var _ stdsql.Result

	argv, ok := args.([]any)
	if !ok {
		return fmt.Errorf("dialect/sql: invalid type %T. expect []any for args", result)
	}
	switch result := result.(type) {
	case nil:
		if _, err := e.tx.Exec(ctx, query, argv...); err != nil {
			return err
		}
	case *sql.Result:
		commandTag, err := e.tx.Exec(ctx, query, argv...)
		if err != nil {
			return err
		}
		*result = execResult{rowsAffected: commandTag.RowsAffected()}
	default:
		return fmt.Errorf("dialect/sql: invalid type %T. expect *sql.Result", result)
	}
	return nil
}

func (e *EntPgxPoolTx) ExecContext(ctx context.Context, query string, args ...any) (stdsql.Result, error) {
	commandTag, err := e.tx.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &execResult{rowsAffected: commandTag.RowsAffected()}, nil
}

func (e *EntPgxPoolTx) Query(ctx context.Context, query string, args, v any) error {
	vr, ok := v.(*sql.Rows)
	if !ok {
		return fmt.Errorf("dialect/sql: invalid type %T. expect *sql.Rows", v)
	}
	argv, ok := args.([]any)
	if !ok {
		return fmt.Errorf("dialect/sql: invalid type %T. expect []any for args", args)
	}
	pgxRows, err := e.tx.Query(ctx, query, argv...)
	if err != nil {
		return err
	}
	columnScanner := &entPgxRows{pgxRows: pgxRows}
	*vr = sql.Rows{
		columnScanner,
	}
	return nil
}

func (e *EntPgxPoolTx) Commit() error {
	return e.tx.Commit(context.TODO())
}

func (e *EntPgxPoolTx) Rollback() error {
	return e.tx.Rollback(context.TODO())
}

func (e *EntPgxPoolTx) PGXTransaction() pgx.Tx {
	return e.tx
}

type entPgxRows struct {
	pgxRows pgx.Rows
}

func (e entPgxRows) Close() error {
	e.pgxRows.Close()
	return nil
}

// ColumnTypes returns column information such as column type, length,
// and nullable. Some information may not be available from some drivers.
func (e entPgxRows) ColumnTypes() ([]*stdsql.ColumnType, error) {
	return []*stdsql.ColumnType{}, nil
}

// Columns returns the column names.
// Columns returns an error if the rows are closed.
func (e entPgxRows) Columns() ([]string, error) {
	fieldDescs := e.pgxRows.FieldDescriptions()
	columnNames := make([]string, len(fieldDescs))
	for i, fd := range fieldDescs {
		columnNames[i] = fd.Name

	}
	return columnNames, nil
}

func (e entPgxRows) Err() error {
	return e.pgxRows.Err()
}

func (e entPgxRows) Next() bool {
	return e.pgxRows.Next()
}

// NextResultSet prepares the next result set for reading. It reports whether
// there is further result sets, or false if there is no further result set
// or if there is an error advancing to it. The Err method should be consulted
// to distinguish between the two cases.
//
// After calling NextResultSet, the Next method should always be called before
// scanning. If there are further result sets they may not have rows in the result
// set.
func (e entPgxRows) NextResultSet() bool {
	// For now this does not seem like a must have for normal database functionality.
	// This seems to be useful if we want to send 2 sql statements in a single query
	// and when the results of the first query are exhausted, then check if the NextResultSet
	// has values.
	return e.pgxRows.Next()
}

func (e entPgxRows) Scan(dest ...any) error {
	return e.pgxRows.Scan(dest...)
}

type execResult struct {
	lastInsertId int64
	rowsAffected int64
}

func (e execResult) LastInsertId() (int64, error) {
	return e.lastInsertId, nil
}

func (e execResult) RowsAffected() (int64, error) {
	return e.rowsAffected, nil
}
