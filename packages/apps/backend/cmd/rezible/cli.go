package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/execution"
	"github.com/rezible/rezible/internal/http"
	"github.com/rezible/rezible/internal/koanf"
	oapiv1 "github.com/rezible/rezible/openapi/v1"
	"github.com/samber/do/v2"
	"github.com/urfave/cli/v3"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	if runErr := makeCli().Run(ctx, os.Args); runErr != nil {
		log.Fatalf("error: %v", runErr)
	}
}

func makeCli() *cli.Command {
	i := do.New()

	return &cli.Command{
		Name:  "rezible",
		Usage: "backend server control",
		Before: func(ctx context.Context, command *cli.Command) (context.Context, error) {
			ctx = execution.NewRootContext(ctx, execution.KindAnonymous, execution.SourceCLI)
			cfg, cfgErr := koanf.LoadConfig(ctx, koanf.Options{LoadEnvironment: true})
			if cfgErr != nil {
				return nil, fmt.Errorf("load config: %w", cfgErr)
			}
			do.ProvideValue(i, *cfg)
			declareServices(ctx, i)
			return ctx, nil
		},
		After: func(ctx context.Context, command *cli.Command) error {
			return shutdownServices(ctx, i)
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
				Name:  "print-config",
				Usage: "print loaded configuration",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Printf("%+v\n", do.MustInvoke[rez.Config](i))
					return nil
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
							direction := cmd.StringArg("direction")
							return do.MustInvoke[rez.MigrationService](i).Run(ctx, direction)
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
							return do.MustInvoke[rez.MigrationService](i).CreateSchemaMigration(ctx, name)
						},
					},
					{
						Name:  "update-checksum",
						Usage: "Update the database migrations checksum file",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							return do.MustInvoke[rez.MigrationService](i).UpdateChecksum()
						},
					},
				},
			},
		},
	}
}
