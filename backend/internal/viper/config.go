package viper

import (
	"net"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	rez "github.com/rezible/rezible"
)

type Config struct {
	db *rez.Database
}

func InitConfig() *Config {
	cfg := &Config{}

	// viper.SetEnvPrefix("REZ")
	viper.SetDefault("stop_timeout", 10)

	viper.AutomaticEnv()

	if cfg.DebugMode() {
		log.Logger = log.Level(zerolog.DebugLevel).Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	return cfg
}

func (c *Config) GetString(key string) string {
	return viper.GetString(key)
}

func (c *Config) GetBool(key string) bool {
	return viper.GetBool(key)
}

func (c *Config) SingleTenantMode() bool {
	return c.GetBool("single_tenant_mode")
}

func (c *Config) DebugMode() bool {
	return c.GetBool("debug_mode")
}

func (c *Config) DatabaseUrl() string {
	return c.GetString("db_url")
}

func (c *Config) ApiRouteBase() string {
	return "/api"
}

func (c *Config) AuthRouteBase() string {
	return "/auth"
}

func (c *Config) BackendUrl() string {
	return "http://localhost:8888"
}

func (c *Config) AppUrl() string {
	return "https://app.rezible.test"
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

func (c *Config) ServerStopTimeout() time.Duration {
	secs := viper.GetInt("stop_timeout")
	return time.Duration(secs) * time.Second
}
