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
	"github.com/rezible/rezible/ent"
	rezinternal "github.com/rezible/rezible/internal"
	"github.com/rezible/rezible/internal/dataproviders"
	"github.com/rezible/rezible/internal/db"
	"github.com/rezible/rezible/internal/db/datasync"
	"github.com/rezible/rezible/internal/viper"
	"github.com/rezible/rezible/jobs"
	"github.com/rezible/rezible/openapi"
)

var cli = &cobra.Command{
	Use:   "rezible",
	Short: "",
	Run:   serveCmd.Run,
}

func init() {
	rez.Config = viper.InitConfig()

	cli.AddCommand(serveCmd)
	cli.AddCommand(printSpecCmd)
	cli.AddCommand(migrateCmd)
	cli.AddCommand(syncCmd)
	cli.AddCommand(loadFakeConfigCmd)
	cli.AddCommand(loadDevConfigCmd)
}

func main() {
	ctx, stopFn := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopFn()

	if runErr := cli.ExecuteContext(ctx); runErr != nil {
		os.Exit(1)
	}
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
		spec, specErr := openapi.GetYamlSpec("")
		if specErr != nil {
			log.Fatal().Err(specErr).Msg("failed to get OpenAPI spec")
		}
		fmt.Println(spec)
	},
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := access.SystemContext(cmd.Context())
		dbc, dbcErr := rezinternal.OpenDatabase(ctx)
		if dbcErr != nil {
			log.Fatal().Err(dbcErr).Msg("failed to load database")
		}
		if migErr := dbc.RunMigrations(ctx); migErr != nil {
			log.Fatal().Err(migErr).Msg("failed to run migrations")
		}
	},
}

func withDatabaseClient(fn func(ctx context.Context, client *ent.Client)) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		dbc, dbcErr := rezinternal.OpenDatabase(ctx)
		if dbcErr != nil {
			log.Fatal().Err(dbcErr).Msg("failed to get database")
		}

		defer func() {
			if closeErr := dbc.Close(); closeErr != nil {
				log.Error().Err(closeErr).Msg("failed to close database connection")
			}
		}()

		fn(ctx, dbc.Client())
	}
}

var loadDevConfigCmd = &cobra.Command{
	Use:   "load-dev-config",
	Short: "Load dev config",
	Run:   withDatabaseClient(dataproviders.LoadDevConfig),
}

var loadFakeConfigCmd = &cobra.Command{
	Use:   "load-fake-config",
	Short: "Load fake config",
	Run:   withDatabaseClient(dataproviders.LoadFakeConfig),
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Run sync",
	Run: withDatabaseClient(func(ctx context.Context, client *ent.Client) {
		cfgs, cfgsErr := db.NewProviderConfigService(client)
		if cfgsErr != nil {
			log.Fatal().Err(cfgsErr).Msg("failed to load provider configs")
		}
		syncSvc := datasync.NewProviderSyncService(client, dataproviders.NewProviderLoader(cfgs))
		syncErr := syncSvc.SyncProviderData(ctx, jobs.SyncProviderData{Hard: true})
		if syncErr != nil {
			log.Fatal().Err(syncErr).Msg("failed to sync provider data")
		}
	}),
}
