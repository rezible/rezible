package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/internal/db"
	"github.com/rezible/rezible/internal/http"
	demoprovider "github.com/rezible/rezible/internal/integrations/demo"
	"github.com/rezible/rezible/internal/postgres/migrations"
	"github.com/rezible/rezible/pkg/execution"
	oapiv1 "github.com/rezible/rezible/pkg/openapi/v1"
	"github.com/urfave/cli/v3"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	if runErr := makeApplicationCli().Run(ctx, os.Args); runErr != nil {
		log.Fatalf("run: %v", runErr)
	}
}

// runApplication builds the application with the profiles, runs fn with it, and shuts it down.
// Each command builds its own, so a command's profiles apply before anything is resolved.
func runApplication(ctx context.Context, fn func(context.Context, *Application) error, profiles ...Package) error {
	app := NewApplication(profiles...)
	appCtx, initErr := app.Init(ctx, true)
	if initErr != nil {
		return fmt.Errorf("init application: %w", initErr)
	}
	runErr := fn(appCtx, app)
	return errors.Join(runErr, app.Shutdown(appCtx))
}

func makeApplicationCli() *cli.Command {
	return &cli.Command{
		Name:  "rezible",
		Usage: "backend server control",
		Commands: []*cli.Command{
			serveCommand,
			printConfigCommand,
			printSpecCommand,
			setupDemoCommand,
			migrationsCommands,
			aiCommands,
		},
	}
}

var (
	serveCommand = &cli.Command{
		Name:  "serve",
		Usage: "Run rezible server",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runApplication(ctx, func(ctx context.Context, app *Application) error {
				if prepareErr := app.prepare(); prepareErr != nil {
					return prepareErr
				}
				if seedErr := seedDevelopmentIdentity(ctx, app); seedErr != nil {
					return fmt.Errorf("seed development identity: %w", seedErr)
				}
				return app.With(func(server *http.Server) error {
					return app.RunLifecycle(ctx, server)
				})
			})
		},
	}

	printConfigCommand = &cli.Command{
		Name:  "print-config",
		Usage: "print loaded configuration",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runApplication(ctx, func(ctx context.Context, app *Application) error {
				return app.With(func(cfg rez.Config) error {
					_, printErr := fmt.Fprintln(cmd.Writer, cfg.Format())
					return printErr
				})
			})
		},
	}

	printSpecCommand = &cli.Command{
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
	}
)

var setupDemoCommand = &cli.Command{
	Name:  "setup-demo",
	Usage: "Create a local incident workspace for the development user",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		return runApplication(ctx, func(ctx context.Context, app *Application) error {
			return app.runLifecycleFunc(ctx, func(ctx context.Context) error {
				if seedErr := seedDevelopmentIdentity(ctx, app); seedErr != nil {
					return fmt.Errorf("seed development identity: %w", seedErr)
				}
				sess, sessErr := createDevelopmentSession(ctx, app)
				if sessErr != nil {
					return fmt.Errorf("seed development identity: %w", sessErr)
				}
				ctx = execution.NewUserContext(ctx, sess)

				var installation rez.InstalledIntegration
				installErr := app.With(func(i rez.IntegrationService) error {
					var installErr error
					installation, installErr = i.InstallNew(ctx, "demo", []byte("{}"))
					return installErr
				})
				if installErr != nil {
					return fmt.Errorf("install demo integration: %w", installErr)
				}

				return demoprovider.SeedDemoData(
					ctx,
					app.mustInvoke[rez.Database](),
					app.mustInvoke[rez.KnowledgeGraphQueryService](),
					app.mustInvoke[rez.OrganizationService](),
					app.mustInvoke[rez.IncidentDetailsService](),
					app.mustInvoke[rez.IncidentManagementService](),
					app.mustInvoke[rez.RetrospectiveService](),
					app.mustInvoke[rez.SystemAnalysisService](),
					app.mustInvoke[rez.EventsService](),
					app.mustInvoke[rez.SituationService](),
					app.mustInvoke[*db.SituationService](),
					app.mustInvoke[rez.ProviderEventIngester](),
					installation.Integration(),
				)
			})
		})
	},
}

var migrationsCommands = &cli.Command{
	Name:  "migrations",
	Usage: "database migrations control",
	Commands: []*cli.Command{
		{
			Name:  "apply",
			Usage: "Apply pending database migrations",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				if cmd.Args().Len() != 0 {
					return fmt.Errorf("migrations apply takes no arguments")
				}
				return withMigrationDatabase(ctx, migrations.Apply)
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
				name := cmd.StringArg("name")
				if !migrationName.MatchString(name) {
					return fmt.Errorf("invalid migration name %q: use lowercase snake_case", name)
				}
				return runApplication(ctx, func(ctx context.Context, app *Application) error {
					return app.With(func(cfg rez.PostgresConfig) error {
						return createMigration(ctx, cfg, name, cmd.Writer)
					})
				})
			},
		},
		{
			Name:  "status",
			Usage: "Show applied and pending migrations",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				return outputMigrationStatus(ctx, cmd.Writer)
			},
		},
		{
			Name:  "update-checksum",
			Usage: "Update the database migrations checksum file",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				return migrations.UpdateChecksum()
			},
		},
	},
}

var (
	aiCommands = &cli.Command{
		Name:  "ai",
		Usage: "commands for working with rezible ai",
		Commands: []*cli.Command{
			{
				Name:  "evals",
				Usage: "evaluate AI workflows and agents",
				Commands: []*cli.Command{
					evalsListCommand,
					evalsRunCommand,
				},
			},
		},
	}
	evalsListCommand = &cli.Command{
		Name:  "list",
		Usage: "List evaluation targets",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runApplication(ctx, func(ctx context.Context, app *Application) error {
				return printEvalTargets(cmd.Writer, app)
			})
		},
	}
	evalsRunCommand = &cli.Command{
		Name:  "run",
		Usage: "Run an evaluation target several times and summarize the results",
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:      "target",
				UsageText: "target name",
				Config:    cli.StringConfig{TrimSpace: true},
			},
		},
		Flags: []cli.Flag{
			&cli.IntFlag{Name: "runs", Value: 3, Usage: "how many times to run the target"},
			&cli.StringFlag{Name: "out", Usage: "directory for the results (default .evals/<target>-<time>)"},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			er := newEvalRunner(cmd.StringArg("target"), cmd.Int("runs"), cmd.String("out"))
			return er.run(ctx, cmd.Writer)
		},
	}
)
