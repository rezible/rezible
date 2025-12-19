package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	rezinternal "github.com/rezible/rezible/internal"
	"github.com/rezible/rezible/internal/dataproviders"
	"github.com/rezible/rezible/internal/db"
	"github.com/rezible/rezible/internal/db/datasync"
	"github.com/rezible/rezible/internal/viper"
	"github.com/rezible/rezible/jobs"
	oapiv1 "github.com/rezible/rezible/openapi/v1"
)

var rootCmd = &cobra.Command{
	Use:   "rezible",
	Short: "",
	Run:   serveCmd.Run,
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "runs the rezible server",
	Run: func(cmd *cobra.Command, args []string) {
		if runErr := rezinternal.RunServer(cmd.Context()); runErr != nil {
			log.Fatal().Err(runErr).Msg("failed to run server")
		}
	},
}

var printSpecCmd = &cobra.Command{
	Use:   "openapi",
	Short: "Print the OpenAPI spec",
	Run: func(cmd *cobra.Command, args []string) {
		spec, specErr := oapiv1.GetYamlSpec()
		if specErr != nil {
			log.Fatal().Err(specErr).Msg("failed to get OpenAPI spec")
		}
		fmt.Println(spec)
	},
}

func withDatabase(ctx context.Context, fn func(dbc rez.Database)) {
	dbc, dbcErr := rezinternal.OpenPostgresDatabase(ctx)
	if dbcErr != nil {
		log.Fatal().Err(dbcErr).Msg("failed to get database")
	}

	defer func() {
		if closeErr := dbc.Close(); closeErr != nil {
			log.Error().Err(closeErr).Msg("failed to close database connection")
		}
	}()

	fn(dbc)
}

var integrationsCmd = &cobra.Command{
	Use: "integrations",
}

var integrationsLoadCmd = &cobra.Command{
	Use:   "load [source]",
	Short: "Load tenant integrations from source",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := access.SystemContext(cmd.Context())
		src := args[0]
		withDatabase(cmd.Context(), func(dbc rez.Database) {
			var loadErr error
			if src == "fake" {
				loadErr = dataproviders.LoadFakeConfig(ctx, dbc.Client())
			} else {
				loadErr = dataproviders.LoadTenantConfig(ctx, dbc.Client(), src)
			}
			if loadErr != nil {
				log.Fatal().Err(loadErr).Msg("failed to load tenant config")
			}
		})
	},
}

var integrationsSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Run integration data sync",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := access.SystemContext(cmd.Context())
		syncArgs := jobs.SyncProviderData{
			Hard: true,
		}
		withDatabase(ctx, func(dbc rez.Database) {
			client := dbc.Client()
			intgs, intgsErr := db.NewIntegrationsService(client)
			if intgsErr != nil {
				log.Fatal().Err(intgsErr).Msg("db.NewIntegrationsService")
			}
			svc := datasync.NewProviderSyncService(client, dataproviders.NewProviderLoader(intgs))
			syncErr := svc.SyncProviderData(ctx, syncArgs)
			if syncErr != nil {
				log.Fatal().Err(syncErr).Msg("failed to sync provider data")
			}
		})
	},
}

var dbCmd = &cobra.Command{
	Use: "db",
}

var dbMigrateCmd = &cobra.Command{
	Use: "migrate",
}

var dbMigrateGenerateCmd = &cobra.Command{
	Use:   "generate [name]",
	Short: "create a new migration",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var dbMigrateApplyCmd = &cobra.Command{
	Use:   "apply [direction]",
	Short: "apply database migrations",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		direction := args[0]
		if direction == "auto" {
			if migErr := rezinternal.RunAutoMigrations(cmd.Context()); migErr != nil {
				log.Fatal().Err(migErr).Msg("failed to apply database migrations")
			}
		} else {
			log.Warn().Str("direction", direction).Msg("version migrations not implemented")
		}
	},
}

func init() {
	rez.Config = viper.InitConfig()

	rootCmd.AddCommand(serveCmd, printSpecCmd, integrationsCmd, dbCmd)

	integrationsCmd.AddCommand(integrationsLoadCmd, integrationsSyncCmd)

	dbCmd.AddCommand(dbMigrateCmd)
	dbMigrateCmd.AddCommand(dbMigrateGenerateCmd, dbMigrateApplyCmd)
}

func main() {
	ctx, stopFn := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopFn()

	if runErr := rootCmd.ExecuteContext(ctx); runErr != nil {
		os.Exit(1)
	}
}
