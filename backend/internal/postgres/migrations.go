package postgres

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/internal/postgres/river"
)

func RunMigrations(ctx context.Context, dbUrl string) error {
	dbc, dbcErr := Open(ctx, dbUrl)
	if dbcErr != nil {
		return fmt.Errorf("open postgres db error: %w", dbcErr)
	}
	pool := dbc.Pool
	driver := ent.Driver(entsql.OpenDB(dialect.Postgres, stdlib.OpenDBFromPool(pool)))
	client := ent.NewClient(driver)
	if schemaErr := client.Schema.Create(ctx); schemaErr != nil {
		return fmt.Errorf("create schema: %w", schemaErr)
	}

	if riverErr := river.RunMigrations(ctx, pool); riverErr != nil {
		return fmt.Errorf("failed to run river migrations: %w", riverErr)
	}

	// TODO: enable RLS?
	// https://entgo.io/docs/migration/row-level-security

	return nil
}
