package river

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river/riverdriver"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
)

func RunMigration(ctx context.Context, pool *pgxpool.Pool, direction string) error {
	cfg := &rivermigrate.Config{
		Line:   riverdriver.MigrationLineMain,
		Schema: SchemaName,
		Logger: nil,
	}
	rm, rmErr := rivermigrate.New(riverpgxv5.New(pool), cfg)
	if rmErr != nil {
		return fmt.Errorf("rivermigrate: %w", rmErr)
	}
	rd := rivermigrate.DirectionUp
	if direction == "down" {
		rd = rivermigrate.DirectionDown
	} else if direction != "up" {
		return fmt.Errorf("unknown direction: %s", direction)
	}
	opts := &rivermigrate.MigrateOpts{
		DryRun:        false,
		MaxSteps:      0,
		TargetVersion: 0,
	}
	_, mErr := rm.Migrate(ctx, rd, opts)
	if mErr != nil {
		return fmt.Errorf("migrate: %w", mErr)
	}
	return nil
}
