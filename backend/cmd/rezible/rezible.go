package main

import (
	"context"
	
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func main() {
	onParsed := func(hooks humacli.Hooks, opts *Options) {
		rezSrv := newRezServer(opts)
		hooks.OnStart(rezSrv.Start)
		hooks.OnStop(rezSrv.Stop)
	}
	cli := humacli.New(onParsed)

	addCmd := func(use string, short string, cmdFn func(ctx context.Context, opts *Options) error) {
		cmd := &cobra.Command{Use: use, Short: short,
			Run: humacli.WithOptions(func(cmd *cobra.Command, args []string, o *Options) {
				if cmdErr := cmdFn(cmd.Context(), o); cmdErr != nil {
					log.Fatal().Err(cmdErr).Str("cmd", use).Msg("Failed to execute command")
				}
			}),
		}
		cli.Root().AddCommand(cmd)
	}
	addCmd("openapi", "Print the OpenAPI spec", printSpecCmd)
	addCmd("migrate", "Run database migrations", migrateCmd)
	addCmd("seed", "Seed the database", seedCmd)
	addCmd("load-configs", "Load provider configs file into database", loadConfigCmd)
	addCmd("sync", "Sync the data providers", syncCmd)

	cli.Run()
}
