package postgres

import (
	"fmt"
	"strings"

	rez "github.com/rezible/rezible"
)

type Config struct {
	User       string      `koanf:"user"`
	Password   string      `koanf:"password"`
	Host       string      `koanf:"host"`
	Port       string      `koanf:"port"`
	Database   string      `koanf:"database"`
	SSLMode    string      `koanf:"sslmode"`
	PoolConfig *PoolConfig `koanf:"pool"`
}

type PoolConfig struct {
	MaxConns *int `koanf:"pool_max_conns"`
}

func (pc *PoolConfig) GetDsn() string {
	var dsn []string
	if pc.MaxConns != nil {
		dsn = append(dsn, fmt.Sprintf("pool_max_conns='%d'", *pc.MaxConns))
	}

	return strings.Join(dsn, " ")
}

func (cfg *Config) GetDsn() string {
	var dsn []string
	dsn = append(dsn, fmt.Sprintf("user='%s'", cfg.User))
	if cfg.Password != "" {
		dsn = append(dsn, fmt.Sprintf("password='%s'", cfg.Password))
	}
	dsn = append(dsn, fmt.Sprintf("host='%s'", cfg.Host))
	dsn = append(dsn, fmt.Sprintf("port='%s'", cfg.Port))
	if cfg.Database != "" {
		dsn = append(dsn, fmt.Sprintf("dbname='%s'", cfg.Database))
	}
	dsn = append(dsn, fmt.Sprintf("sslmode='%s'", cfg.SSLMode))
	if cfg.PoolConfig != nil {
		dsn = append(dsn, cfg.PoolConfig.GetDsn())
	}
	return strings.Join(dsn, " ")
}

func LoadConfig() (Config, error) {
	cfg := Config{
		User:       "postgres",
		Password:   "",
		Host:       "localhost",
		Port:       "5432",
		Database:   "",
		SSLMode:    "require",
		PoolConfig: nil,
	}
	return cfg, rez.Config.Unmarshal("postgres", &cfg)
}
