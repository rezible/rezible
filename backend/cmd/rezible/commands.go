package main

import (
	"context"
	"fmt"
	"github.com/rezible/rezible/internal/documents"
	"github.com/rezible/rezible/internal/river"

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
		const (
			syncUsers     = true
			syncOncall    = true
			syncIncidents = true
		)

		c := db.Client()
		pl := providers.NewProviderLoader(c.ProviderConfig)

		users, usersErr := postgres.NewUserService(c, pl)
		if usersErr != nil {
			return fmt.Errorf("user service: %w", usersErr)
		}

		if syncUsers {
			_, chatErr := documents.NewChatService(ctx, pl, users)
			if chatErr != nil {
				return fmt.Errorf("to create chat: %w", chatErr)
			}
			if syncErr := users.SyncData(ctx); syncErr != nil {
				return fmt.Errorf("users sync failed: %w", syncErr)
			}
		}

		if syncOncall {
			oncall, oncallErr := postgres.NewOncallService(ctx, c, nil, pl, nil, nil, users, nil)
			if oncallErr != nil {
				return fmt.Errorf("postgres.NewOncallService: %w", oncallErr)
			}
			if syncErr := oncall.SyncData(ctx); syncErr != nil {
				return fmt.Errorf("oncall sync failed: %w", syncErr)
			}
		}

		if syncIncidents {
			inc, incErr := postgres.NewIncidentService(ctx, c, nil, pl, nil, nil, users)
			if incErr != nil {
				return fmt.Errorf("postgres.NewIncidentService: %w", incErr)
			}
			if syncErr := inc.SyncData(ctx); syncErr != nil {
				return fmt.Errorf("incidents sync failed: %w", syncErr)
			}
		}

		return nil
	})
}

func loadConfigCmd(ctx context.Context, opts *Options) error {
	return withDatabase(ctx, opts, func(db *postgres.Database) error {
		return providers.LoadFromFile(ctx, db.Client(), ".dev_provider_configs.json")
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
