package river

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
	"github.com/rs/zerolog/log"
)

func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	cfg := &rivermigrate.Config{}
	migrator, migErr := rivermigrate.New(riverpgxv5.New(pool), cfg)
	if migErr != nil {
		return fmt.Errorf("failed to create migrator: %w", migErr)
	}

	opts := &rivermigrate.MigrateOpts{}
	res, migrationErr := migrator.Migrate(ctx, rivermigrate.DirectionUp, opts)
	if migrationErr != nil {
		return fmt.Errorf("failed to migrate: %w", migrationErr)
	}

	if len(res.Versions) > 0 {
		log.Info().Int("versions", len(res.Versions)).Msg("ran river migrations")
	}

	return nil
}

func GetMigrations(ctx context.Context, pool *pgxpool.Pool) ([]rivermigrate.Migration, error) {
	cfg := &rivermigrate.Config{}
	migrator, migErr := rivermigrate.New(riverpgxv5.New(pool), cfg)
	if migErr != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", migErr)
	}
	return migrator.AllVersions(), nil
}
