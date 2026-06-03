package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/entpgx"
	_ "github.com/rezible/rezible/ent/runtime"
	"github.com/rezible/rezible/execution"
)

type DatabaseClient struct {
	driver dialect.Driver
	client *ent.Client
}

type PgxPool = pgxpool.Pool

func NewPgxPoolDatabaseClient(pool *pgxpool.Pool) (*DatabaseClient, error) {
	return newDatabaseClient(entpgx.NewPgxPoolDriver(pool)), nil
}

func NewStdDatabaseClient(db *sql.DB) *DatabaseClient {
	return newDatabaseClient(entsql.OpenDB("postgres", db))
}

func newDatabaseClient(driver dialect.Driver) *DatabaseClient {
	dbc := &DatabaseClient{driver: driver}
	dbc.client = ent.NewClient(ent.Driver(driver))
	dbc.client.Use(dbc.ensureTenantIdSetHook())
	dbc.client.Intercept(dbc.setTenantContextInterceptor())
	return dbc
}

func (dbc *DatabaseClient) Client(ctx context.Context) *ent.Client {
	if tx := ent.TxFromContext(ctx); tx != nil {
		return tx.Client()
	}
	return dbc.client
}

func (dbc *DatabaseClient) WithTx(ctx context.Context, fn func(txCtx context.Context, tx *ent.Client) error, opts ...ent.TxOption) error {
	if tx := ent.TxFromContext(ctx); tx != nil {
		applyTxOptions(tx, opts...)
		return fn(ctx, tx.Client())
	}

	tx, txErr := dbc.client.Tx(ctx)
	if txErr != nil {
		return fmt.Errorf("begin transaction: %w", txErr)
	}
	applyTxOptions(tx, opts...)

	defer func() {
		if v := recover(); v != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				panic(fmt.Errorf("%v: rollback transaction: %w", v, rbErr))
			}
			panic(v)
		}
	}()

	txCtx := ent.NewTxContext(ctx, tx)
	if fnErr := fn(txCtx, tx.Client()); fnErr != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("%w: rollback transaction: %w", fnErr, rbErr)
		}
		return fnErr
	}
	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("commit transaction: %w", commitErr)
	}
	return nil
}

func applyTxOptions(tx *ent.Tx, opts ...ent.TxOption) {
	txOpts := &ent.TxOptions{}
	for _, opt := range opts {
		opt(txOpts)
	}
	for _, hook := range txOpts.OnCommit {
		tx.OnCommit(hook)
	}
	for _, hook := range txOpts.OnRollback {
		tx.OnRollback(hook)
	}
}

func (dbc *DatabaseClient) RequireUpToDateMigrations(ctx context.Context) error {
	ms := &MigrationService{driver: dbc.driver}
	status, statusErr := ms.GetCurrentStatus(ctx)
	if statusErr != nil {
		return fmt.Errorf("get current migration status: %w", statusErr)
	}
	fmtStatus := fmt.Sprintf("[current=%d latest=%d]", status.CurrentVersion, status.LatestVersion)
	if status.Dirty {
		return fmt.Errorf("database migrations status is dirty: %s", fmtStatus)
	}
	if status.CurrentVersion < status.LatestVersion {
		return fmt.Errorf("database migrations status is pending: %s", fmtStatus)
	}
	return nil
}

func closeDatabaseResource(name string, c io.Closer) {
	if closeErr := c.Close(); closeErr != nil && !errors.Is(closeErr, sql.ErrConnDone) {
		slog.Error("failed to close "+name, "error", closeErr)
	}
}

func (dbc *DatabaseClient) Shutdown() error {
	if dbc.client != nil {
		closeDatabaseResource("db client", dbc.client)
	}
	return nil
}

func (dbc *DatabaseClient) setTenantContextInterceptor() ent.Interceptor {
	return ent.InterceptFunc(func(q ent.Querier) ent.Querier {
		return ent.QuerierFunc(func(ctx context.Context, query ent.Query) (ent.Value, error) {
			if tenantId, tenantIdSet := execution.GetContext(ctx).TenantID(); tenantIdSet {
				ctx = entsql.WithIntVar(ctx, "app.current_tenant", tenantId)
			}
			return q.Query(ctx, query)
		})
	})
}

func (dbc *DatabaseClient) ensureTenantIdSetHook() ent.Hook {
	type tenantedMutation interface {
		SetTenantID(int)
	}
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if tm, ok := m.(tenantedMutation); ok {
				if _, alreadySet := m.Field("tenant_id"); !alreadySet {
					tenantId, tenantIdSet := execution.GetContext(ctx).TenantID()
					if !tenantIdSet {
						return nil, rez.ErrTenantContextMissing
					}
					tm.SetTenantID(tenantId)
				}
			}
			return next.Mutate(ctx, m)
		})
	}
}
