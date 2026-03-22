package postgres

import (
	"context"
	_ "embed"
	"fmt"

	_ "github.com/rezible/rezible/ent/runtime"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog/log"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/entpgx"
)

type Config struct {
	Url string `koanf:"connection_url"`
}

type DatabaseClient struct {
	pool   *pgxpool.Pool
	driver dialect.Driver
	client *ent.Client
}

func NewDatabasePoolClient(ctx context.Context) (*DatabaseClient, error) {
	pool, poolErr := openPgxPool(ctx)
	if poolErr != nil {
		return nil, poolErr
	}

	driver := entpgx.NewPgxPoolDriver(pool)

	return &DatabaseClient{pool: pool, driver: driver, client: MakeClient(driver)}, nil
}

func MakeClient(driver dialect.Driver) *ent.Client {
	client := ent.NewClient(ent.Driver(driver))
	client.Use(ensureTenantIdSetHook)
	client.Intercept(setTenantContextInterceptor())
	return client
}

func openPgxPool(ctx context.Context) (*pgxpool.Pool, error) {
	connCfg, cfgErr := GetPgxConfig()
	if cfgErr != nil {
		return nil, fmt.Errorf("config: %w", cfgErr)
	}
	parsedCfg, parseErr := pgxpool.ParseConfig(connCfg.ConnString())
	if parseErr != nil {
		return nil, fmt.Errorf("parse: %w", parseErr)
	}
	pool, poolErr := pgxpool.NewWithConfig(ctx, parsedCfg)
	if poolErr != nil {
		return nil, fmt.Errorf("create: %w", poolErr)
	}

	if pingErr := pool.Ping(ctx); pingErr != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", pingErr)
	}

	return pool, nil
}

func GetPgxConfig() (*pgx.ConnConfig, error) {
	if connUrl := rez.Config.GetString("db_url"); connUrl != "" {
		return pgx.ParseConfig(connUrl)
	}
	var cfg Config
	if cfgErr := rez.Config.Unmarshal("postgres", &cfg); cfgErr != nil {
		return nil, fmt.Errorf("config error: %w", cfgErr)
	}
	return pgx.ParseConfig(cfg.Url)
}

func (dbc *DatabaseClient) Client() *ent.Client {
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
