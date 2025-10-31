package viper

import (
	"net"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	rez "github.com/rezible/rezible"
)

type Config struct {
	cli *cobra.Command
	db  *rez.Database
}

func InitConfig(cli *cobra.Command) *Config {
	cfg := &Config{cli: cli}

	flags := cli.Flags()
	flags.Bool("debug", false, "Enable debug logging")
	_ = viper.BindPFlag("debug", flags.Lookup("debug"))

	flags.String("db_url", "", "Database URL")
	_ = viper.BindPFlag("db_url", flags.Lookup("db_url"))

	flags.Int("stop_timeout", 10, "server stop timeout")
	_ = viper.BindPFlag("stop_timeout", flags.Lookup("stop_timeout"))

	viper.SetEnvPrefix("REZ")
	viper.AutomaticEnv()

	if cfg.DebugMode() {
		log.Logger = log.Level(zerolog.DebugLevel).Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	return cfg
}

func (c *Config) DebugMode() bool {
	return os.Getenv("REZ_DEBUG") == "true"
}

func (c *Config) DatabaseUrl() string {
	return viper.GetString("db_url")
}

func (c *Config) BackendUrl() string {
	return "http://localhost:8888"
}

func (c *Config) FrontendUrl() string {
	return "http://localhost:5173"
}

func (c *Config) AllowTenantCreation() bool {
	return c.DebugMode()
}

func (c *Config) AllowUserCreation() bool {
	return c.DebugMode()
}

func (c *Config) HttpServerAddress() string {
	host := "localhost"
	port := "8888"
	return net.JoinHostPort(host, port)
}

func (c *Config) DocumentServerAddress() string {
	return "localhost:8889"
}

func (c *Config) ServerStopTimeout() time.Duration {
	secs := viper.GetInt("stop_timeout")
	return time.Duration(secs) * time.Second
}
