package postgres

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

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
	*pgxpool.Pool
	client *ent.Client
}

func Open(ctx context.Context, uri string) (*DatabaseClient, error) {
	pool, poolErr := pgxpool.New(ctx, uri)
	if poolErr != nil {
		return nil, fmt.Errorf("create: %w", poolErr)
	}

	if pingErr := pool.Ping(ctx); pingErr != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", pingErr)
	}

	return &DatabaseClient{Pool: pool}, nil
}

func (d *DatabaseClient) Client() *ent.Client {
	if d.client == nil {
		driver := ent.Driver(entpgx.NewPgxPoolDriver(d.Pool))
		d.client = ent.NewClient(driver)
		d.client.Use(ensureTenantIdSetHook)
		d.client.Intercept(setTenantContextInterceptor())
	}
	return d.client
}

func (d *DatabaseClient) Close() error {
	d.Pool.Close()
	if d.client != nil {
		if clientErr := d.client.Close(); clientErr != nil {
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
