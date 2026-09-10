package river

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	rez "github.com/rezible/rezible"
	"github.com/riverqueue/river/riverdriver"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
)

var directions = map[rez.MigrationDirection]rivermigrate.Direction{
	"down": rivermigrate.DirectionDown,
	"up":   rivermigrate.DirectionUp,
}

func RunMigration(ctx context.Context, pool *pgxpool.Pool, direction rez.MigrationDirection) error {
	rdir, directionOk := directions[direction]
	if !directionOk {
		return fmt.Errorf("unknown direction: %s", direction)
	}

	cfg := &rivermigrate.Config{
		Line:   riverdriver.MigrationLineMain,
		Schema: SchemaName,
		Logger: nil,
	}
	migrator, migratorErr := rivermigrate.New(riverpgxv5.New(pool), cfg)
	if migratorErr != nil {
		return fmt.Errorf("rivermigrate: %w", migratorErr)
	}
	opts := &rivermigrate.MigrateOpts{
		DryRun:        false,
		MaxSteps:      0,
		TargetVersion: 0,
	}
	_, migrationErr := migrator.Migrate(ctx, rdir, opts)
	if migrationErr != nil {
		return fmt.Errorf("migrate: %w", migrationErr)
	}
	return nil
}
