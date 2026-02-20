package postgres

import (
	"context"
	"fmt"

	"ariga.io/atlas/sql/sqltool"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"
	_ "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"

	rez "github.com/rezible/rezible"
	entmigrate "github.com/rezible/rezible/ent/migrate"
)

func GenerateEntMigrations(ctx context.Context, name string) error {
	pool, poolErr := openPgxPool(ctx, rez.Config.DatabaseUrl())
	if poolErr != nil {
		return poolErr
	}
	defer pool.Close()

	dir, err := sqltool.NewGolangMigrateDir("./migrations")
	if err != nil {
		return fmt.Errorf("failed creating atlas migration directory: %v", err)
	}
	// Migrate diff options.
	opts := []schema.MigrateOption{
		schema.WithDir(dir),
		schema.WithMigrationMode(schema.ModeReplay),
		schema.WithDialect(dialect.Postgres),
	}

	driver := entsql.OpenDB(dialect.Postgres, stdlib.OpenDBFromPool(pool))
	m, mErr := schema.NewMigrate(driver, opts...)
	if mErr != nil {
		return fmt.Errorf("failed creating migrate: %v", mErr)
	}
	return m.NamedDiff(ctx, name, entmigrate.Tables...)
}
