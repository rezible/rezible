package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v3"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	if runErr := makeCli().Run(ctx, os.Args); runErr != nil {
		fmt.Printf("run error: %v\n", runErr)
		os.Exit(1)
	}
}

func makeCli() *cli.Command {
	a := newApp()

	return &cli.Command{
		Name:  "rezible",
		Usage: "backend server control",
		Before: func(ctx context.Context, command *cli.Command) (context.Context, error) {
			return a.setup(ctx)
		},
		After: func(_ context.Context, command *cli.Command) error {
			return a.shutdown()
		},
		Commands: []*cli.Command{
			{
				Name:  "serve",
				Usage: "Run rezible server",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return a.serveHttp(ctx)
				},
			},
			{
				Name:  "print-config",
				Usage: "print loaded configuration",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return a.printConfig()
				},
			},
			{
				Name:  "spec",
				Usage: "Print the OpenAPI spec",
				Flags: []cli.Flag{&cli.BoolFlag{Name: "json"}},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return a.printOpenApiSpec(cmd.Bool("json"))
				},
			},
			{
				Name:  "migrations",
				Usage: "database migrations control",
				Commands: []*cli.Command{
					{
						Name:  "create",
						Usage: "Create a new database migration",
						Arguments: []cli.Argument{&cli.StringArg{
							Name:      "name",
							UsageText: "name of the migration",
							Config:    cli.StringConfig{TrimSpace: true},
						}},
						Action: func(ctx context.Context, cmd *cli.Command) error {
							return a.createSchemaMigration(ctx, cmd.StringArg("name"))
						},
					},
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
							return a.applySchemaMigrations(ctx, cmd.StringArg("direction"))
						},
					},
					{
						Name:  "update-checksum",
						Usage: "Update the database migrations checksum file",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							return a.updateMigrationChecksumFile()
						},
					},
				},
			},
		},
	}
}
