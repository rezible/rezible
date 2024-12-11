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

	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"

	"github.com/twohundreds/rezible/ent"
	"github.com/twohundreds/rezible/ent/entpgx"
	_ "github.com/twohundreds/rezible/ent/runtime"
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

// for some reason the entpgx.PgxPoolDriver hangs when migrating
func (d *Database) migrationEntDriver() *entsql.Driver {
	return entsql.OpenDB(dialect.Postgres, stdlib.OpenDBFromPool(d.Pool))
}

func (d *Database) RunMigrations(ctx context.Context) error {
	if entErr := runEntMigrations(ctx, d); entErr != nil {
		return fmt.Errorf("failed to run ent migrations: %w", entErr)
	}

	if riverErr := runRiverMigrations(ctx, d); riverErr != nil {
		return fmt.Errorf("failed to run river migrations: %w", riverErr)
	}

	return nil
}

//go:embed sql/001_ladder_ancestry.sql
var ancestrySql string

func runEntMigrations(ctx context.Context, db *Database) error {
	client := ent.NewClient(ent.Driver(db.migrationEntDriver()))
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

func runRiverMigrations(ctx context.Context, db *Database) error {
	cfg := &rivermigrate.Config{}
	migrator, migErr := rivermigrate.New(riverpgxv5.New(db.Pool), cfg)
	if migErr != nil {
		return fmt.Errorf("failed to create migrator: %w", migErr)
	}

	opts := &rivermigrate.MigrateOpts{}
	res, migrationErr := migrator.Migrate(ctx, rivermigrate.DirectionUp, opts)
	if migrationErr != nil {
		return fmt.Errorf("failed to migrate: %w", migrationErr)
	}

	if len(res.Versions) > 0 {
		log.Info().Int("versions", len(res.Versions)).Msg("ran river migrations")
	}

	return nil
}
