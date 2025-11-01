package atlas

/*
import (
	"context"
	"fmt"

	entmigrate "github.com/rezible/rezible/ent/migrate"

	"ariga.io/atlas/sql/sqltool"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql/schema"
)

func GenerateMigrationFile(ctx context.Context, dbUrl string, fileName string) error {
	dir, dirErr := sqltool.NewGolangMigrateDir("ent/migrate/migrations")
	if dirErr != nil {
		return fmt.Errorf("creating migration directory: %w", dirErr)
	}
	opts := []schema.MigrateOption{
		schema.WithDir(dir),
		schema.WithMigrationMode(schema.ModeReplay),
		schema.WithDialect(dialect.Postgres),
		schema.WithErrNoPlan(true),
		// withConcurrentIndexes(true, true),
	}

	dbSchema := entmigrate.NewSchema(drv)

	dbSchema.WriteTo(ctx,, )

	diffErr := entmigrate.(cmd.Context(), url, name, opts...)
	if diffErr != nil {
		return fmt.Errorf("generating migration file: %w", diffErr)
	}
	return nil
}

*/
