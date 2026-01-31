package viper

import (
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

	viper.AutomaticEnv()

	if cfg.DebugMode() {
		log.Logger = log.Level(zerolog.DebugLevel).Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	return cfg
}

func (c *Config) GetString(key string) string {
	return viper.GetString(key)
}

func (c *Config) GetStringOr(key string, orDefault string) string {
	if viper.IsSet(key) {
		return viper.GetString(key)
	}
	return orDefault
}

func (c *Config) GetStrings(key string) []string {
	return viper.GetStringSlice(key)
}

func (c *Config) GetBool(key string) bool {
	return viper.GetBool(key)
}

func (c *Config) GetBoolOr(key string, orDefault bool) bool {
	if viper.IsSet(key) {
		return viper.GetBool(key)
	}
	return orDefault
}

func (c *Config) GetDuration(key string) time.Duration {
	return viper.GetDuration(key)
}

func (c *Config) GetDurationOr(key string, orDefault time.Duration) time.Duration {
	if viper.IsSet(key) {
		return viper.GetDuration(key)
	}
	return orDefault
}

func (c *Config) DebugMode() bool {
	return c.GetBool("debug_mode")
}

func (c *Config) DataSyncMode() bool {
	return c.GetBool("datasync_mode")
}

func (c *Config) DatabaseUrl() string {
	return c.GetString("db_url")
}

func (c *Config) AppUrl() string {
	return c.GetString("app_url")
}

func (c *Config) ApiRouteBase() string {
	return "/api"
}

func (c *Config) AuthRouteBase() string {
	return "/auth"
}

func (c *Config) DocumentsServerAddress() string {
	return c.GetString("documents_server_address")
}

func (c *Config) AuthSessionSecret() string {
	return c.GetString("auth.session_secret")
}

func (c *Config) SingleTenantMode() bool {
	return c.GetBool("single_tenant_mode")
}

func (c *Config) AllowUserCreation() bool {
	return c.DebugMode()
}

func (c *Config) AllowTenantCreation() bool {
	return c.DebugMode()
}
