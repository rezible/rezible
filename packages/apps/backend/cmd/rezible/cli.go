package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

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
	r := makeCommandRunner()

	return &cli.Command{
		Name:  "rezible",
		Usage: "backend server control",
		Before: func(ctx context.Context, command *cli.Command) (context.Context, error) {
			return r.setupContext(ctx)
		},
		After: func(ctx context.Context, command *cli.Command) error {
			return r.shutdown(ctx)
		},
		Commands: []*cli.Command{
			{
				Name:  "serve",
				Usage: "Run rezible server",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return r.runServer(ctx)
				},
			},
			{
				Name:  "print-config",
				Usage: "print loaded configuration",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return r.printConfig()
				},
			},
			{
				Name:  "spec",
				Usage: "Print the OpenAPI spec",
				Flags: []cli.Flag{&cli.BoolFlag{Name: "json"}},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return r.printOpenApiSpec(cmd.Bool("json"))
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
							return r.runSchemaMigration(ctx, cmd.StringArg("direction"))
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
							return r.createSchemaMigration(ctx, cmd.StringArg("name"))
						},
					},
					{
						Name:  "update-checksum",
						Usage: "Update the database migrations checksum file",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							return r.updateMigrationChecksumFile()
						},
					},
				},
			},
		},
	}
}
