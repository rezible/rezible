package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/execution"
	"github.com/rezible/rezible/integrations"
	"github.com/rezible/rezible/integrations/projections"
	apiv1 "github.com/rezible/rezible/internal/api/v1"
	"github.com/rezible/rezible/internal/db"
	"github.com/rezible/rezible/internal/http"
	"github.com/rezible/rezible/internal/postgres/river"
	"github.com/rezible/rezible/internal/watermill"
	"github.com/samber/do/v2"
	"github.com/urfave/cli/v3"

	"github.com/rezible/rezible/internal/koanf"
	"github.com/rezible/rezible/internal/postgres"
	oapiv1 "github.com/rezible/rezible/openapi/v1"
	"github.com/rezible/rezible/telemetry"
)

func main() {
	baseCtx := context.Background()
	ctx, stop := signal.NotifyContext(baseCtx, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	i := do.New()
	ctx = execution.NewRootContext(ctx, execution.KindAnonymous, execution.SourceCLI)

	root := &cli.Command{
		Name:  "rezible",
		Usage: "backend server control",
		Commands: []*cli.Command{
			makeServeCommand(i),
			makeMigrationsCommand(i),
			makeSpecCommand(),
		},
		Before: func(ctx context.Context, command *cli.Command) (context.Context, error) {
			if pkgErr := provideDependencies(ctx, i); pkgErr != nil {
				return nil, fmt.Errorf("failed to provide dependencies: %w", pkgErr)
			}
			return ctx, nil
		},
		After: func(ctx context.Context, command *cli.Command) error {
			shutdownCtx, cancelShutdown := context.WithTimeout(baseCtx, 5*time.Second)
			defer cancelShutdown()
			if shutdown := i.ShutdownWithContext(shutdownCtx); !shutdown.Succeed {
				return shutdown
			}
			return nil
		},
	}

	if runErr := root.Run(ctx, os.Args); runErr != nil {
		fmt.Printf("run error: %v\n", runErr)
		os.Exit(1)
	}
}

func provideDependencies(ctx context.Context, i do.Injector) error {
	koanf.Package(i)
	telemetry.PackageContext(ctx, i)
	postgres.PackageContext(ctx, i)
	river.Package(i)
	watermill.Package(i)
	integrations.Package(i)
	projections.Package(i)
	db.Package(i)
	http.Package(i)
	apiv1.Package(i)

	if intgErr := integrations.RegisterIntegrations(i); intgErr != nil {
		return fmt.Errorf("failed to register integrations: %w", intgErr)
	}

	cfg, cfgErr := do.MustInvoke[rez.ConfigLoader](i).LoadConfig(ctx)
	if cfgErr != nil || cfg == nil {
		return fmt.Errorf("failed to load config: %w", cfgErr)
	}
	do.ProvideValue[rez.Config](i, *cfg)

	return nil
}

func makeServeCommand(i do.Injector) *cli.Command {
	return &cli.Command{
		Name:  "serve",
		Usage: "Run rezible server",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runServer(ctx, i)
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

func makeMigrationsCommand(i do.Injector) *cli.Command {
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
					mc := do.MustInvoke[*postgres.MigratorClient](i)
					return mc.CreateSchemaMigration(ctx, nameArg.Value)
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
					mc := do.MustInvoke[*postgres.MigratorClient](i)
					return mc.Run(ctx, directionArg.Value)
				},
			},
		},
	}
}
