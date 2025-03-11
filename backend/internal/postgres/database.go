package postgres

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/rs/zerolog/log"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"

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
		return nil, fmt.Errorf("create pgxpool: %w", poolErr)
	}

	if pingErr := pool.Ping(ctx); pingErr != nil {
		pool.Close()
		return nil, fmt.Errorf("ping pgxpool: %w", pingErr)
	}

	return &Database{Pool: pool}, nil
}

func (d *Database) Close() error {
	d.Pool.Close()
	return nil
}

func (d *Database) Client() *ent.Client {
	drv := entpgx.NewPgxPoolDriver(d.Pool)
	return ent.NewClient(ent.Driver(drv))
}

//go:embed sql/001_ladder_ancestry.sql
var ancestrySql string

func (d *Database) RunEntMigrations(ctx context.Context) error {
	driver := entsql.OpenDB(dialect.Postgres, stdlib.OpenDBFromPool(d.Pool))
	client := ent.NewClient(ent.Driver(driver))
	defer func(client *ent.Client) {
		err := client.Close()
		if err != nil {
			log.Error().Err(err).Msg("failed to close ent client")
		}
	}(client)

	schemaErr := client.Schema.Create(ctx, schema.WithHooks(removeAncestryForeignKeysHook))
	if schemaErr != nil {
		return fmt.Errorf("create schema: %w", schemaErr)
	}

	/*
		// TODO: use a versioned migration schema
		// https://entgo.io/docs/versioned/intro
		const hackyTriggerExistsQuery = `SELECT COUNT(*)>0 AS exists FROM information_schema.triggers WHERE trigger_name = 'materialize_ancestry_closure';`
		var exists bool
		row := db.QueryRow(ctx, hackyTriggerExistsQuery)
		if scanErr := row.Scan(&exists); scanErr != nil {
			return fmt.Errorf("failed to check existance of trigger: %w", scanErr)
		}
		if !exists {
			// first time setup
			if _, resErr := db.Exec(ctx, ancestrySql); resErr != nil {
				return fmt.Errorf("failed to execute sql: %w", resErr)
			}
		}
	*/
	return nil
}

// ent does not support fine-grained foreign keys, so we remove the existing ones and make our own

func removeAncestryForeignKeysHook(next schema.Creator) schema.Creator {
	return schema.CreateFunc(func(ctx context.Context, tables ...*schema.Table) error {
		/*
			for _, table := range tables {
				if table.Name == ladderancestry.Table {
					table.ForeignKeys = []*schema.ForeignKey{}
					break
				}
			}
		*/
		return next.Create(ctx, tables...)
	})
}
