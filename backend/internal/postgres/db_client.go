package postgres

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent/entpgx"
	"github.com/rs/zerolog/log"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/rezible/rezible/ent"
	_ "github.com/rezible/rezible/ent/runtime"
)

type DatabaseClient struct {
	pool   *pgxpool.Pool
	client *ent.Client
}

func NewDatabaseClient(ctx context.Context) (*DatabaseClient, error) {
	pool, poolErr := pgxpool.New(ctx, rez.Config.DatabaseUrl())
	if poolErr != nil {
		return nil, fmt.Errorf("create: %w", poolErr)
	}

	if pingErr := pool.Ping(ctx); pingErr != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", pingErr)
	}
	log.Debug().Msg("successfully connected to postgres")

	return &DatabaseClient{pool: pool}, nil
}

func (dbc *DatabaseClient) Client() *ent.Client {
	if dbc.client == nil {
		driver := ent.Driver(entpgx.NewPgxPoolDriver(dbc.pool))
		dbc.client = ent.NewClient(driver)
		dbc.client.Use(ensureTenantIdSetHook)
		dbc.client.Intercept(setTenantContextInterceptor())
	}
	return dbc.client
}

func (dbc *DatabaseClient) Pool() *pgxpool.Pool {
	return dbc.pool
}

func (dbc *DatabaseClient) Close() error {
	dbc.pool.Close()
	if dbc.client != nil {
		if clientErr := dbc.client.Close(); clientErr != nil {
			return fmt.Errorf("closing ent client: %w", clientErr)
		}
	}
	return nil
}

func setTenantContextInterceptor() ent.Interceptor {
	return ent.InterceptFunc(func(q ent.Querier) ent.Querier {
		return ent.QuerierFunc(func(ctx context.Context, query ent.Query) (ent.Value, error) {
			authCtx := access.GetContext(ctx)
			if tenantId, idExists := authCtx.TenantId(); idExists {
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
				tid, tenantExists := access.GetContext(ctx).TenantId()
				if !tenantExists {
					return nil, errors.New("tenant not found in auth context")
				}
				tm.SetTenantID(tid)
			}
		}
		return next.Mutate(ctx, m)
	})
}

func debugLogQueryAccessAuthContext() ent.Interceptor {
	return ent.InterceptFunc(func(q ent.Querier) ent.Querier {
		return ent.QuerierFunc(func(ctx context.Context, query ent.Query) (ent.Value, error) {
			authCtx := access.GetContext(ctx)
			log.Debug().Bool("isSystem", authCtx.HasRole(access.RoleSystem)).Msg("query")
			return q.Query(ctx, query)
		})
	})
}
