package postgres

import (
	"context"
	"fmt"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/rezible/rezible/ent/runtime"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/jackc/pgx/v5/pgxpool"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/entpgx"
)

type DatabaseClient struct {
	pool   *pgxpool.Pool
	driver dialect.Driver
	client *ent.Client
}

func NewDatabaseClient(ctx context.Context) (*DatabaseClient, error) {
	cfg, cfgErr := LoadConfig()
	if cfgErr != nil {
		return nil, fmt.Errorf("load config: %w", cfgErr)
	}

	if migrationsErr := RequireCurrentMigrations(ctx, cfg); migrationsErr != nil {
		return nil, fmt.Errorf("migration status: %w", migrationsErr)
	}

	pool, poolErr := openPgxPool(ctx, cfg.getDsn(cfg.AppRole))
	if poolErr != nil {
		return nil, fmt.Errorf("open pgxpool: %w", poolErr)
	}

	dbc := &DatabaseClient{
		pool:   pool,
		driver: entpgx.NewPgxPoolDriver(pool),
	}

	return dbc, nil
}

func (dbc *DatabaseClient) Client() *ent.Client {
	if dbc.client == nil {
		dbc.client = MakeEntClient(dbc.driver)
	}
	return dbc.client
}

func (dbc *DatabaseClient) Pool() *pgxpool.Pool {
	return dbc.pool
}

func (dbc *DatabaseClient) Close() {
	if dbc.client != nil {
		if clientErr := dbc.client.Close(); clientErr != nil {
			slog.Error("failed to close client", "error", clientErr)
		}
	}
	if dbc.driver != nil {
		if driverErr := dbc.driver.Close(); driverErr != nil {
			slog.Error("failed to close pool driver", "error", driverErr)
		}
	}
	if dbc.pool != nil {
		dbc.pool.Close()
	}
}

func setTenantContextInterceptor() ent.Interceptor {
	return ent.InterceptFunc(func(q ent.Querier) ent.Querier {
		return ent.QuerierFunc(func(ctx context.Context, query ent.Query) (ent.Value, error) {
			if tenantId, tenantIdSet := access.GetTenantId(ctx); tenantIdSet {
				ctx = entsql.WithIntVar(ctx, "app.current_tenant", tenantId)
			}
			return q.Query(ctx, query)
		})
	})
}

func ensureTenantIdSetHook(next ent.Mutator) ent.Mutator {
	type tenantedMutation interface {
		SetTenantID(int)
	}
	return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
		if tm, ok := m.(tenantedMutation); ok {
			if _, alreadySet := m.Field("tenant_id"); !alreadySet {
				tenantId, tenantIdSet := access.GetTenantId(ctx)
				if !tenantIdSet {
					return nil, rez.ErrTenantContextMissing
				}
				tm.SetTenantID(tenantId)
			}
		}
		return next.Mutate(ctx, m)
	})
}
