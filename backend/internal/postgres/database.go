package postgres

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/rezible/rezible/access"
	"github.com/rs/zerolog/log"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/entpgx"
	_ "github.com/rezible/rezible/ent/runtime"
)

type Database struct {
	*pgxpool.Pool
}

func Open(ctx context.Context, uri string) (*Database, error) {
	pool, poolErr := pgxpool.New(ctx, uri)
	if poolErr != nil {
		return nil, fmt.Errorf("create: %w", poolErr)
	}

	if pingErr := pool.Ping(ctx); pingErr != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", pingErr)
	}

	return &Database{Pool: pool}, nil
}

func (d *Database) Close() error {
	d.Pool.Close()
	return nil
}

func ensureTenantIdSetHook(next ent.Mutator) ent.Mutator {
	type tenantedMutation interface {
		SetTenantID(int)
	}
	return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
		if m.Op() == ent.OpCreate {
			if tm, ok := m.(tenantedMutation); ok {
				tid, tenantExists := access.GetAuthContext(ctx).TenantId()
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
			authCtx := access.GetAuthContext(ctx)
			log.Debug().Bool("isSystem", authCtx.HasRole(access.RoleSystem)).Msg("query")
			return q.Query(ctx, query)
		})
	})
}

func (d *Database) Client() *ent.Client {
	c := ent.NewClient(ent.Driver(entpgx.NewPgxPoolDriver(d.Pool)))
	c.Use(ensureTenantIdSetHook)
	// c.Intercept(debugLogQueryAccessAuthContext())
	return c
}

func (d *Database) tryCloseClient(c *ent.Client) {
	if closeErr := c.Close(); closeErr != nil {
		log.Error().Err(closeErr).Msg("failed to close ent client")
	}
}

func (d *Database) RunEntMigrations(ctx context.Context) error {
	driver := ent.Driver(entsql.OpenDB(dialect.Postgres, stdlib.OpenDBFromPool(d.Pool)))
	client := ent.NewClient(driver)
	defer d.tryCloseClient(client)

	schemaErr := client.Schema.Create(ctx)
	if schemaErr != nil {
		return fmt.Errorf("create schema: %w", schemaErr)
	}

	return nil
}
