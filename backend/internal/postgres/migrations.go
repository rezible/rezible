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
	"github.com/rs/zerolog/log"

	entmigrate "github.com/rezible/rezible/ent/migrate"
)

func GenerateEntMigrations(ctx context.Context, name string) error {
	pool, poolErr := openPgxPool(ctx)
	if poolErr != nil {
		return poolErr
	}
	conn := stdlib.OpenDBFromPool(pool)
	defer func() {
		pool.Close()
		if err := conn.Close(); err != nil {
			log.Error().Err(err).Str("name", name).Msg("failed to close database connection")
		}
	}()

	dir, err := sqltool.NewGolangMigrateDir("./migrations")
	if err != nil {
		return fmt.Errorf("failed creating atlas migration directory: %v", err)
	}
	opts := []schema.MigrateOption{
		schema.WithDir(dir),
		schema.WithMigrationMode(schema.ModeReplay),
		schema.WithDialect(dialect.Postgres),
	}
	m, mErr := schema.NewMigrate(entsql.OpenDB(dialect.Postgres, conn), opts...)
	if mErr != nil {
		return fmt.Errorf("failed creating migrate: %v", mErr)
	}
	return m.NamedDiff(ctx, name, entmigrate.Tables...)
}
