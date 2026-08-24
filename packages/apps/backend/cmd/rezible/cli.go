package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/rezible/rezible/internal/genkit"
	"github.com/rezible/rezible/pkg/ai/evals"
	"github.com/samber/do/v2"
	"github.com/urfave/cli/v3"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/internal/http"
	"github.com/rezible/rezible/pkg/execution"
	oapiv1 "github.com/rezible/rezible/pkg/openapi/v1"
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
	i := makePackageInjector()

	return &cli.Command{
		Name:  "rezible",
		Usage: "backend server control",
		Before: func(ctx context.Context, command *cli.Command) (context.Context, error) {
			return ctx, initPackages(ctx, i)
		},
		After: func(ctx context.Context, command *cli.Command) error {
			return shutdownInjector(ctx, i)
		},
		Commands: []*cli.Command{
			{
				Name:  "serve",
				Usage: "Run rezible server",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return runServicesFor[*http.Server](ctx, i)
				},
			},
			{
				Name:  "genkit-server",
				Usage: "Run genkit dev server",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return runServicesFor[*genkit.DevServer](ctx, i)
				},
			},
			{
				Name:  "print-config",
				Usage: "print loaded configuration",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return withConfig(i, func(cfg rez.Config) error {
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
							return withMigrationService(i, func(ms rez.MigrationService) error {
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
							return withMigrationService(i, func(ms rez.MigrationService) error {
								return ms.CreateSchemaMigration(ctx, cmd.StringArg("name"))
							})
						},
					},
					{
						Name:  "update-checksum",
						Usage: "Update the database migrations checksum file",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							return withMigrationService(i, func(ms rez.MigrationService) error {
								return ms.UpdateChecksum()
							})
						},
					},
				},
			},
			{
				Name:  "evals",
				Usage: "list and run AI agent evaluation scenarios",
				Before: func(ctx context.Context, command *cli.Command) (context.Context, error) {
					useTestDatabase(i)
					return ctx, nil
				},
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
							scenario, lookupErr := evals.Lookup(cmd.StringArg("name"))
							if lookupErr != nil {
								return lookupErr
							}
							svc, svcErr := do.Invoke[*genkit.EvaluationService](i)
							if svcErr != nil {
								return svcErr
							}
							runner, runnerErr := svc.MakeRunner(scenario)
							if runnerErr != nil {
								return runnerErr
							}
							report := runner.RunEvaluation(ctx)
							return report.Write(cmd.Writer)
						},
					},
				},
			},
		},
	}
}
