package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rezible/rezible/internal/postgres"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/internal"
	"github.com/rezible/rezible/internal/koanf"
	"github.com/rezible/rezible/jobs"
	oapiv1 "github.com/rezible/rezible/openapi/v1"
)

var rezcli = &cli.Command{
	Name:  "rezible",
	Usage: "backend server control",
	Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
		cfgOpts := koanf.ConfigLoaderOptions{
			LoadEnvironment: true,
		}
		var cfgErr error
		rez.Config, cfgErr = koanf.NewConfigLoader(ctx, cfgOpts)
		if cfgErr != nil {
			return nil, fmt.Errorf("failed to load configuration: %w", cfgErr)
		}
		if rez.Config.DebugMode() {
			log.Logger = log.Level(zerolog.DebugLevel).Output(zerolog.ConsoleWriter{Out: os.Stderr})
		}
		return access.AnonymousContext(ctx), nil
	},
	Commands: []*cli.Command{
		{
			Name:  "serve",
			Usage: "Run rezible server",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				srv, srvErr := internal.NewServer(ctx)
				if srvErr != nil {
					return srvErr
				}
				return srv.RunServe(ctx)
			},
		},
		{
			Name:  "spec",
			Usage: "Print the OpenAPI spec",
			Flags: []cli.Flag{&cli.BoolFlag{Name: "json"}},
			Action: func(ctx context.Context, cmd *cli.Command) error {
				spec, specErr := oapiv1.GetSpec(cmd.Bool("json"))
				if specErr != nil {
					return fmt.Errorf("failed to marshal OpenAPI spec: %w", specErr)
				}
				fmt.Printf("%s", spec)
				return nil
			},
		},
		{
			Name:  "sync-integrations",
			Usage: "Run integration data sync",
			Flags: []cli.Flag{&cli.BoolFlag{Name: "hard"}},
			Action: func(ctx context.Context, cmd *cli.Command) error {
				srv, srvErr := internal.NewServer(ctx)
				if srvErr != nil {
					return srvErr
				}
				return srv.RunDataSync(ctx, jobs.SyncIntegrationsData{IgnoreHistory: cmd.Bool("hard")})
			},
		},
		{
			Name:  "generate-migration",
			Usage: "Create a new database migration",
			Arguments: []cli.Argument{
				&cli.StringArg{
					Name:      "name",
					UsageText: "name of the migration",
					Config:    cli.StringConfig{TrimSpace: true},
				},
			},
			Action: func(ctx context.Context, cmd *cli.Command) error {
				return postgres.GenerateEntMigrations(ctx, cmd.StringArg("name"))
			},
		},
	},
}

func main() {
	ctx, stopFn := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopFn()
	if runErr := rezcli.Run(ctx, os.Args); runErr != nil {
		log.Fatal().Err(runErr).Msg("failed to run")
	}
}
