package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	rez "github.com/rezible/rezible"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cli = &cobra.Command{
	Use:   "rezible",
	Short: "",
	Run:   serveCmd.Run,
}

func init() {
	// TODO: actually parse
	rez.DebugMode = os.Getenv("REZ_DEBUG") == "true"
	if rez.DebugMode {
		log.Logger = log.Level(zerolog.DebugLevel).Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	flags := cli.Flags()
	flags.String("addr", ":8080", "HTTP server address")
	_ = viper.BindPFlag("addr", flags.Lookup("addr"))

	flags.Bool("debug", false, "Enable debug logging")
	_ = viper.BindPFlag("debug", flags.Lookup("debug"))

	flags.String("db_url", "", "Database URL")
	_ = viper.BindPFlag("db_url", flags.Lookup("db_url"))

	viper.SetEnvPrefix("REZ")

	viper.AutomaticEnv()

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
