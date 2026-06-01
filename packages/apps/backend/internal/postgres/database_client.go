package postgres

import (
	"context"
	"database/sql"
	"errors"
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

func NewPgxPoolDatabaseClient(pool *pgxpool.Pool) *DatabaseClient {
	return newDatabaseClient(entpgx.NewPgxPoolDriver(pool))
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
	return dbc.client
}

func (dbc *DatabaseClient) WithTx(ctx context.Context, fn func(txCtx context.Context) error, opts ...ent.TxOption) error {

	return nil
}

//func (dbc *DatabaseClient) RequireUpToDateMigrations(ctx context.Context) error {
//	return NewMigrator().requireUpToDate(ctx)
//}

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
