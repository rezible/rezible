package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/internal/http"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/ai/evals"
	oapiv1 "github.com/rezible/rezible/pkg/openapi/v1"
	"github.com/urfave/cli/v3"

	"github.com/rezible/rezible/pkg/execution"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	ctx = execution.NewRootContext(ctx, execution.KindAnonymous, execution.SourceCLI)

	if runErr := makeServerCli().Run(ctx, os.Args); runErr != nil {
		log.Fatalf("error: %v", runErr)
	}
}

func makeServerCli() *cli.Command {
	app := newApplication()

	return &cli.Command{
		Name:  "rezible",
		Usage: "backend server control",
		Before: func(ctx context.Context, command *cli.Command) (context.Context, error) {
			app.init(ctx,
				withEnvironmentConfig,
				withOpenTelemetry,
				withPostgresDatabase,
				withGenkitAiService)
			return ctx, nil
		},
		Commands: makeCliCommands(app),
		After: func(ctx context.Context, command *cli.Command) error {
			return app.shutdown(ctx)
		},
	}
}

func makeCliCommands(app *application) []*cli.Command {
	return []*cli.Command{
		{
			Name:  "serve",
			Usage: "Run rezible server",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				return app.Run[*http.Server](ctx)
			},
		},
		{
			Name:  "print-config",
			Usage: "print loaded configuration",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				return app.with(func(cfg rez.Config) error {
					fmt.Println(cfg.Format())
					return nil
				})
			},
		},
		{
			Name:  "spec",
			Usage: "Print the OpenAPI spec",
			Flags: []cli.Flag{&cli.BoolFlag{Name: "json"}},
			Action: func(ctx context.Context, cmd *cli.Command) error {
				api := oapiv1.MakeOpenApiSpec()
				marshalFn := api.YAML
				if cmd.Bool("json") {
					marshalFn = api.MarshalJSON
				}
				spec, marshalErr := marshalFn()
				if spec != nil {
					fmt.Printf("%s", spec)
				}
				return marshalErr
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
						return app.with(func(ms rez.MigrationService) error {
							return ms.Run(ctx, cmd.StringArg("direction"))
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
						return app.with(func(ms rez.MigrationService) error {
							return ms.CreateSchemaMigration(ctx, cmd.StringArg("name"))
						})
					},
				},
				{
					Name:  "update-checksum",
					Usage: "Update the database migrations checksum file",
					Action: func(ctx context.Context, cmd *cli.Command) error {
						return app.with(func(ms rez.MigrationService) error {
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
				app.override(postgresTestDatabaseConfig)
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
								return app.with(func(svc rezai.EvalScenarioRunner) error {
									result, runErr := svc.RunNamedScenario(ctx, cmd.StringArg("name"))
									if runErr != nil || result == nil {
										return fmt.Errorf("failed to run scenario: %w", runErr)
									}
									encoder := json.NewEncoder(cmd.Writer)
									encoder.SetIndent("", "  ")
									if jsonErr := encoder.Encode(result); jsonErr != nil {
										return fmt.Errorf("encode evaluation result: %w", jsonErr)
									}
									if result.Status != rezai.EvalRunStatusPassed {
										return fmt.Errorf("evaluation %s", result.Status)
									}
									return nil
								})
							},
						},
					},
				},
			},
		},
	}
}
