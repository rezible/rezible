package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/internal/http"
	"github.com/rezible/rezible/pkg/ai/evals"
	oapiv1 "github.com/rezible/rezible/pkg/openapi/v1"
	"github.com/urfave/cli/v3"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	serverCli := makeServerCli(NewApplication())

	if runErr := serverCli.Run(ctx, os.Args); runErr != nil {
		log.Fatalf("run: %v", runErr)
	}
}

func makeServerCli(app *Application) *cli.Command {
	serverInitProviders := []InitProvider{
		withEnvironmentConfig,
		withOpenTelemetry,
		withPostgresDatabase,
		withGenkitAiService,
	}
	return &cli.Command{
		Name:  "rezible",
		Usage: "backend server control",
		Before: func(ctx context.Context, command *cli.Command) (context.Context, error) {
			return app.Init(ctx, serverInitProviders...)
		},
		Commands: makeCliCommands(app),
		After: func(ctx context.Context, command *cli.Command) error {
			return app.Shutdown(ctx)
		},
	}
}

func makeCliCommands(app *Application) []*cli.Command {
	return []*cli.Command{
		{
			Name:  "serve",
			Usage: "Run rezible server",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				return app.RunLifecycle[*http.Server](ctx)
			},
		},
		{
			Name:  "print-config",
			Usage: "print loaded configuration",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				return app.With(func(cfg rez.Config) error {
					_, printErr := fmt.Fprintln(cmd.Writer, cfg.Format())
					return printErr
				})
			},
		},
		{
			Name:  "spec",
			Usage: "Print the OpenAPI spec",
			Flags: []cli.Flag{&cli.BoolFlag{Name: "json"}},
			Action: func(ctx context.Context, cmd *cli.Command) error {
				specFmt := oapiv1.SpecFormatYAML
				if cmd.Bool("json") {
					specFmt = oapiv1.SpecFormatJSON
				}
				spec, marshalErr := oapiv1.GetEncodedSpec(specFmt)
				if marshalErr != nil {
					return fmt.Errorf("encoded spec: %w", marshalErr)
				}
				if _, writeErr := cmd.Writer.Write(spec); writeErr != nil {
					return fmt.Errorf("writing spec: %w", writeErr)
				}
				return nil
			},
		},
		{
			Name:  "migrations",
			Usage: "database migrations control",
			Commands: []*cli.Command{
				{
					Name:  "apply",
					Usage: "Apply pending database migrations",
					Arguments: []cli.Argument{&cli.StringArg{
						Name:      "direction",
						Value:     "up",
						UsageText: "direction to migrate",
						Config:    cli.StringConfig{TrimSpace: true},
					}},
					Action: func(ctx context.Context, cmd *cli.Command) error {
						return app.With(func(ms rez.MigrationService) error {
							return ms.Run(ctx, rez.MigrationDirection(cmd.StringArg("direction")))
						})
					},
				},
				{
					Name:  "create",
					Usage: "Create a new database migration",
					Arguments: []cli.Argument{&cli.StringArg{
						Name:      "name",
						UsageText: "name of the migration",
						Config:    cli.StringConfig{TrimSpace: true},
					}},
					Action: func(ctx context.Context, cmd *cli.Command) error {
						return app.With(func(ms rez.MigrationService) error {
							return ms.CreateSchemaMigration(ctx, cmd.StringArg("name"))
						})
					},
				},
				{
					Name:  "update-checksum",
					Usage: "Update the database migrations checksum file",
					Action: func(ctx context.Context, cmd *cli.Command) error {
						return app.With(func(ms rez.MigrationService) error {
							return ms.UpdateChecksum()
						})
					},
				},
			},
		},
		{
			Name:  "ai",
			Usage: "commands for working with rezible ai",
			Before: func(ctx context.Context, command *cli.Command) (context.Context, error) {
				app.Override[rez.PostgresConfig](providePostgresTestDatabaseConfig)
				return ctx, nil
			},
			Commands: []*cli.Command{
				{
					Name:  "evals",
					Usage: "ai evals",
					Commands: []*cli.Command{
						{
							Name:  "list",
							Usage: "List available evaluation scenarios",
							Action: func(ctx context.Context, cmd *cli.Command) error {
								for _, definition := range evals.List() {
									fmt.Printf("%s\t%s\t%s\n", definition.Name, definition.AgentName, definition.Description)
								}
								return nil
							},
						},
						{
							Name:  "run",
							Usage: "Run a named evaluation scenario",
							Arguments: []cli.Argument{&cli.StringArg{
								Name:      "name",
								UsageText: "scenario name",
								Config:    cli.StringConfig{TrimSpace: true},
							}},
							Action: func(ctx context.Context, cmd *cli.Command) error {
								return app.RunAiEvalScenario(ctx, cmd.StringArg("name"), cmd.Writer)
							},
						},
					},
				},
			},
		},
	}
}
