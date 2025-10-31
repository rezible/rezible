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

func (dbc *DatabaseClient) RunMigrations(ctx context.Context) error {
	driver := ent.Driver(entsql.OpenDB(dialect.Postgres, stdlib.OpenDBFromPool(dbc.pool)))
	client := ent.NewClient(driver)
	if schemaErr := client.Schema.Create(ctx); schemaErr != nil {
		return fmt.Errorf("create schema: %w", schemaErr)
	}

	if riverErr := river.RunMigrations(ctx, dbc.pool); riverErr != nil {
		return fmt.Errorf("failed to run river migrations: %w", riverErr)
	}

	// TODO: enable RLS?
	// https://entgo.io/docs/migration/row-level-security

	return nil
}
