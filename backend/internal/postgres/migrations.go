package postgres

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/rezible/rezible/internal/postgres/river"

	"github.com/jackc/pgx/v5/stdlib"
)

func RunAutoMigrations(ctx context.Context) error {
	dbc, dbcErr := NewDatabaseClient(ctx)
	if dbcErr != nil {
		return fmt.Errorf("creating database client: %w", dbcErr)
	}
	client := dbc.newClient(entsql.OpenDB(dialect.Postgres, stdlib.OpenDBFromPool(dbc.pool)))
	if schemaErr := client.Schema.Create(ctx); schemaErr != nil {
		return fmt.Errorf("create schema: %w", schemaErr)
	}

	// TODO: enable RLS?
	// https://entgo.io/docs/migration/row-level-security

	if riverErr := river.RunMigrations(ctx, dbc.pool); riverErr != nil {
		return fmt.Errorf("failed to run river migrations: %w", riverErr)
	}

	return nil
}
