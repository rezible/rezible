package main

import (
	"context"
	"fmt"
	"github.com/rezible/rezible/internal/river"
	"github.com/rezible/rezible/jobs"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"

	"github.com/rezible/rezible/internal/api"
	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/internal/providers"
	"github.com/rezible/rezible/openapi"
)

func printSpecCmd(ctx context.Context, opts *Options) error {
	spec, yamlErr := yaml.Marshal(openapi.MakeDefaultApi(&api.Handler{}).OpenAPI())
	if yamlErr != nil {
		return yamlErr
	}
	fmt.Println(string(spec))
	return nil
}

func withDatabase(ctx context.Context, opts *Options, fn func(db *postgres.Database) error) error {
	db, dbErr := postgres.Open(ctx, opts.DatabaseUrl)
	if dbErr != nil {
		return fmt.Errorf("failed to open database: %w", dbErr)
	}

	defer func(cdb *postgres.Database) {
		if closeErr := cdb.Close(); closeErr != nil {
			log.Error().Err(closeErr).Msg("failed to close database connection")
		}
	}(db)

	return fn(db)
}

func migrateCmd(ctx context.Context, opts *Options) error {
	return withDatabase(ctx, opts, func(db *postgres.Database) error {
		if dbErr := db.RunEntMigrations(ctx); dbErr != nil {
			return fmt.Errorf("failed to run ent migrations: %w", dbErr)
		}

		if riverErr := river.RunMigrations(ctx, db.Pool); riverErr != nil {
			return fmt.Errorf("failed to run river migrations: %w", riverErr)
		}

		return nil
	})
}

func syncCmd(ctx context.Context, opts *Options) error {
	return withDatabase(ctx, opts, func(db *postgres.Database) error {
		// TODO: use cli flags
		args := &jobs.SyncProviderData{
			Hard:             true,
			Users:            true,
			Teams:            true,
			Incidents:        true,
			Oncall:           true,
			Alerts:           true,
			SystemComponents: true,
		}

		dbc := db.Client()

		pl := providers.NewProviderLoader(dbc.ProviderConfig)

		return providers.SyncData(ctx, args, dbc, pl)
	})
}

func loadConfigCmd(ctx context.Context, opts *Options) error {
	return withDatabase(ctx, opts, func(db *postgres.Database) error {
		return providers.LoadConfigFromFile(ctx, db.Client(), ".dev_provider_configs.json")
	})
}

func seedCmd(ctx context.Context, opts *Options) error {
	return withDatabase(ctx, opts, func(db *postgres.Database) error {
		return seedDatabase(ctx, db)
	})
}

func seedDatabase(ctx context.Context, db *postgres.Database) error {

	return nil
}
