package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"syscall"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/telemetry"
	"github.com/samber/do/v2"
	"github.com/urfave/cli/v3"

	"github.com/rezible/rezible/execution"
	"github.com/rezible/rezible/internal/koanf"
	"github.com/rezible/rezible/internal/postgres"
	oapiv1 "github.com/rezible/rezible/openapi/v1"
)

func main() {
	i := do.New()

	makeCli(i)

	_, shutdown := i.ShutdownOnSignals(syscall.SIGTERM, os.Interrupt)
	if !shutdown.Succeed {
		slog.Error("failed", "error", shutdown.Error())
		os.Exit(1)
	}
}

func makeCli(i do.Injector) *cli.Command {
	beforeFn := func(ctx context.Context, c *cli.Command) (context.Context, error) {
		ctx = execution.NewRootContext(ctx, execution.KindAnonymous, execution.SourceCLI)
		if telemetryErr := telemetry.Init(ctx); telemetryErr != nil {
			return nil, fmt.Errorf("failed to initialize telemetry: %w", telemetryErr)
		}

		cfgOpts := koanf.ConfigLoaderOptions{
			LoadEnvironment: true,
		}
		cfgLoader, cfgErr := koanf.NewConfigLoader(ctx, cfgOpts)
		if cfgErr != nil {
			return nil, fmt.Errorf("failed to load configuration: %w", cfgErr)
		}
		rez.Config = cfgLoader
		do.ProvideValue[rez.ConfigLoader](i, cfgLoader)

		return ctx, nil
	}

	serveCmd := &cli.Command{
		Name:  "serve",
		Usage: "Run rezible server",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runServer(ctx, i)
		},
	}

	return &cli.Command{
		Name:   "rezible",
		Usage:  "backend server control",
		Before: beforeFn,
		Commands: []*cli.Command{
			serveCmd,
			makeSpecCommand(),
			makeMigrationsCommand(),
		},
	}
}

func makeSpecCommand() *cli.Command {
	jsonFlag := &cli.BoolFlag{Name: "json"}
	return &cli.Command{
		Name:  "spec",
		Usage: "Print the OpenAPI spec",
		Flags: []cli.Flag{jsonFlag},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			spec, specErr := oapiv1.GetSpec(jsonFlag.Value)
			if specErr != nil {
				return fmt.Errorf("failed to marshal OpenAPI spec: %w", specErr)
			}
			fmt.Printf("%s", spec)
			return nil
		},
	}
}

func makeMigrationsCommand() *cli.Command {
	nameArg := &cli.StringArg{
		Name:      "name",
		UsageText: "name of the migration",
		Config:    cli.StringConfig{TrimSpace: true},
	}
	directionArg := &cli.StringArg{
		Name:      "direction",
		Value:     "up",
		UsageText: "direction to migrate",
		Config:    cli.StringConfig{TrimSpace: true},
	}

	return &cli.Command{
		Name:  "migrations",
		Usage: "database migrations control",
		Commands: []*cli.Command{
			{
				Name:      "create",
				Usage:     "Create a new database migration",
				Arguments: []cli.Argument{nameArg},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return postgres.CreateSchemaMigration(ctx, nameArg.Value)
				},
			},
			{
				Name:  "update-checksum",
				Usage: "Update the database migrations checksum file",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return postgres.UpdateMigrationsChecksum()
				},
			},
			{
				Name:      "apply",
				Usage:     "Apply pending database migrations",
				Arguments: []cli.Argument{directionArg},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return postgres.RunMigrations(ctx, directionArg.Value)
				},
			},
		},
	}
}
