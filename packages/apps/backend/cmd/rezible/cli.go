package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rezible/rezible/execution"
	"github.com/samber/do/v2"
	"github.com/urfave/cli/v3"

	"github.com/rezible/rezible/internal/postgres"
	oapiv1 "github.com/rezible/rezible/openapi/v1"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	ctx = execution.NewRootContext(ctx, execution.KindAnonymous, execution.SourceCLI)

	if runErr := runCli(ctx); runErr != nil {
		fmt.Printf("run error: %v\n", runErr)
		os.Exit(1)
	}
}

func runCli(ctx context.Context) error {
	i := do.New()

	cmd := &cli.Command{
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
	}
	return cmd.Run(ctx, os.Args)
}

func makeServeCommand(i do.Injector) *cli.Command {
	return &cli.Command{
		Name:  "serve",
		Usage: "Run rezible server",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			app, appErr := do.Invoke[*appRunner](i)
			if appErr != nil {
				return fmt.Errorf("making app runner: %w", appErr)
			}
			return app.start(ctx)
		},
		After: func(_ context.Context, command *cli.Command) error {
			shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelShutdown()
			shutdown := i.ShutdownWithContext(shutdownCtx)

			var shutdownErr error
			for sd, sErr := range shutdown.Errors {
				if !errors.Is(sErr, context.Canceled) {
					fmt.Printf("\n\t[%s] ERROR: %s\n", sd.Service, sErr.Error())
					shutdownErr = errors.Join(shutdownErr, sErr)
				}
			}
			return shutdownErr
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
	var commands []*cli.Command
	nameArg := &cli.StringArg{
		Name:      "name",
		UsageText: "name of the migration",
		Config:    cli.StringConfig{TrimSpace: true},
	}
	commands = append(commands, &cli.Command{
		Name:      "create",
		Usage:     "Create a new database migration",
		Arguments: []cli.Argument{nameArg},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			mc := do.MustInvoke[*postgres.MigratorClient](i)
			return mc.CreateSchemaMigration(ctx, nameArg.Value)
		},
	})

	commands = append(commands, &cli.Command{
		Name:  "update-checksum",
		Usage: "Update the database migrations checksum file",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return postgres.UpdateMigrationsChecksum()
		},
	})

	directionArg := &cli.StringArg{
		Name:      "direction",
		Value:     "up",
		UsageText: "direction to migrate",
		Config:    cli.StringConfig{TrimSpace: true},
	}
	commands = append(commands, &cli.Command{
		Name:      "apply",
		Usage:     "Apply pending database migrations",
		Arguments: []cli.Argument{directionArg},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			mc := do.MustInvoke[*postgres.MigratorClient](i)
			return mc.Run(ctx, directionArg.Value)
		},
	})

	return &cli.Command{
		Name:     "migrations",
		Usage:    "database migrations control",
		Commands: commands,
	}
}
