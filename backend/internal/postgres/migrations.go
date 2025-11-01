package postgres

import (
	"context"
	"fmt"

	"ariga.io/atlas/sql/sqltool"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"
	"github.com/jackc/pgx/v5/stdlib"
	rez "github.com/rezible/rezible"
	entmigrate "github.com/rezible/rezible/ent/migrate"
	"github.com/rezible/rezible/internal/postgres/river"
)

func GenerateMigrationFile(ctx context.Context, name string) error {
	dir, dirErr := sqltool.NewGolangMigrateDir("ent/migrate/migrations")
	if dirErr != nil {
		return fmt.Errorf("creating migration directory: %w", dirErr)
	}

	opts := []schema.MigrateOption{
		schema.WithDir(dir),
		schema.WithMigrationMode(schema.ModeReplay),
		schema.WithDialect(dialect.Postgres),
		schema.WithErrNoPlan(true),
	}

	diffErr := entmigrate.NamedDiff(ctx, rez.Config.DatabaseUrl(), name, opts...)
	if diffErr != nil {
		return fmt.Errorf("generating migration file: %w", diffErr)
	}

	return nil
}

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
