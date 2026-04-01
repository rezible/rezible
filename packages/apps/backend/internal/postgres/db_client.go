package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/rezible/rezible/ent/runtime"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/entpgx"
)

func MakeEntClient(driver dialect.Driver) *ent.Client {
	client := ent.NewClient(ent.Driver(driver))
	client.Use(ensureTenantIdSetHook)
	client.Intercept(setTenantContextInterceptor())
	return client
}

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

	pool, poolErr := openPgxPool(ctx, cfg.getDsn())
	if poolErr != nil {
		return nil, poolErr
	}

	driver := entpgx.NewPgxPoolDriver(pool)
	dbc := &DatabaseClient{pool: pool, driver: driver}

	if migrationsErr := dbc.requireUpToDateMigrations(ctx); migrationsErr != nil {
		dbc.Close()
		return nil, fmt.Errorf("migrations: %w", migrationsErr)
	}

	return dbc, nil
}

func (dbc *DatabaseClient) requireUpToDateMigrations(ctx context.Context) error {
	db := stdlib.OpenDBFromPool(dbc.pool)
	defer closeSqlDb(db)
	status, statusErr := getMigrationStatusFromDB(ctx, db)
	if statusErr != nil {
		return fmt.Errorf("read migration status: %w", statusErr)
	}
	if status.Dirty {
		return fmt.Errorf("database migrations are dirty: %s", status)
	}
	if status.Pending() {
		return fmt.Errorf("database migrations are pending: %s", status)
	}
	return nil
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
			log.Error().Err(clientErr).Msg("failed to close client")
		}
	}
	if dbc.driver != nil {
		if driverErr := dbc.driver.Close(); driverErr != nil {
			log.Error().Err(driverErr).Msg("failed to close pool driver")
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
