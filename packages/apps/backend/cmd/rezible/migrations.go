package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"regexp"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/internal/postgres/migrations"
	"github.com/rezible/rezible/internal/postgres/pgtestdb"
)

var migrationName = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// withMigrationDatabase runs fn against this workspace's database as the admin role, which owns the schema.
func withMigrationDatabase(ctx context.Context, fn func(context.Context, *sql.DB) error) error {
	return runApplication(ctx, func(ctx context.Context, app *Application) error {
		return app.With(func(cfg rez.PostgresConfig) error {
			return withSQLDatabase(ctx, cfg, true, func(db *sql.DB) error {
				return fn(ctx, db)
			})
		})
	})
}

func withSQLDatabase(ctx context.Context, cfg rez.PostgresConfig, admin bool, fn func(*sql.DB) error) error {
	pool, poolErr := postgres.MakePgxPool(ctx, cfg, admin)
	if poolErr != nil {
		return fmt.Errorf("pgx pool: %w", poolErr)
	}
	db := stdlib.OpenDBFromPool(pool.Pool)
	runErr := fn(db)
	return errors.Join(runErr, db.Close(), pool.Shutdown())
}

// createMigration plans the next migration against a scratch database replayed from the migration directory,
// so the workspace's own database, with whatever drift it has, never shapes the generated SQL.
func createMigration(ctx context.Context, cfg rez.PostgresConfig, name string, w io.Writer) error {
	dirPath, dirErr := migrations.LocalDir()
	if dirErr != nil {
		return dirErr
	}
	scratch, scratchErr := pgtestdb.New(cfg)
	if scratchErr != nil {
		return fmt.Errorf("create scratch database: %w", scratchErr)
	}
	planErr := withSQLDatabase(ctx, scratch.Config(), false, func(db *sql.DB) error {
		files, planErr := migrations.Plan(ctx, db, dirPath, name)
		if planErr != nil {
			return planErr
		}
		if len(files) == 0 {
			_, writeErr := fmt.Fprintln(w, "no schema changes")
			return writeErr
		}
		for _, file := range files {
			if _, writeErr := fmt.Fprintln(w, file); writeErr != nil {
				return writeErr
			}
		}
		return nil
	})
	return errors.Join(planErr, scratch.Shutdown())
}

// regenerateInitMigration plans 0001_init.sql against a scratch database whose application schema has been
// emptied, so the plan creates every table.
func regenerateInitMigration(ctx context.Context, cfg rez.PostgresConfig) error {
	scratch, scratchErr := pgtestdb.New(cfg)
	if scratchErr != nil {
		return fmt.Errorf("create scratch database: %w", scratchErr)
	}
	regenerateErr := withSQLDatabase(ctx, scratch.Config(), true, func(db *sql.DB) error {
		schemaName := pgx.Identifier{migrations.SchemaName}.Sanitize()
		if _, emptyErr := db.ExecContext(ctx, fmt.Sprintf("DROP SCHEMA %[1]s CASCADE; CREATE SCHEMA %[1]s", schemaName)); emptyErr != nil {
			return fmt.Errorf("empty scratch schema: %w", emptyErr)
		}
		return migrations.RegenerateInit(ctx, db)
	})
	return errors.Join(regenerateErr, scratch.Shutdown())
}

func outputMigrationStatus(ctx context.Context, w io.Writer) error {
	return withMigrationDatabase(ctx, func(ctx context.Context, db *sql.DB) error {
		applied, pending, statusErr := migrations.Status(ctx, db)
		if statusErr != nil {
			return statusErr
		}
		for _, group := range []struct {
			name  string
			files []string
		}{{"Applied", applied}, {"Pending", pending}} {
			if _, writeErr := fmt.Fprintln(w, group.name+":"); writeErr != nil {
				return writeErr
			}
			for _, file := range group.files {
				if _, writeErr := fmt.Fprintln(w, "  "+file); writeErr != nil {
					return writeErr
				}
			}
		}
		return nil
	})
}
