package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/internal/postgres/river"
	"github.com/rs/zerolog/log"
)

func LoadConfig() (Config, error) {
	cfg := Config{
		Host:     "postgres",
		Port:     5432,
		Database: "rezible",
		SSLMode:  "require",
	}
	return cfg, rez.Config.Unmarshal("postgres", &cfg)
}

type Config struct {
	Host      string      `koanf:"host"`
	Port      uint16      `koanf:"port"`
	Database  string      `koanf:"database"`
	AppRole   RoleConfig  `koanf:"role_app"`
	AdminRole RoleConfig  `koanf:"role_admin"`
	SSLMode   string      `koanf:"sslmode"`
	Pool      *PoolConfig `koanf:"pool"`
}

type PoolConfig struct {
	MaxConns int32 `koanf:"pool_max_conns"`
}

type RoleConfig struct {
	Name     string `koanf:"name"`
	Password string `koanf:"password"`
}

func (cfg *Config) getDsn(role RoleConfig) string {
	var dsn []string
	dsn = append(dsn, fmt.Sprintf("user='%s'", role.Name))
	if role.Password != "" {
		dsn = append(dsn, fmt.Sprintf("password='%s'", role.Password))
	}
	dsn = append(dsn, fmt.Sprintf("host='%s'", cfg.Host))
	dsn = append(dsn, fmt.Sprintf("port=%d", cfg.Port))
	if cfg.Database != "" {
		dsn = append(dsn, fmt.Sprintf("dbname='%s'", cfg.Database))
	}
	dsn = append(dsn, fmt.Sprintf("sslmode='%s'", cfg.SSLMode))
	if cfg.Pool != nil {
		var poolDsn []string
		if cfg.Pool.MaxConns != 0 {
			dsn = append(dsn, fmt.Sprintf("pool_max_conns='%d'", cfg.Pool.MaxConns))
		}
		if len(poolDsn) > 0 {
			dsn = append(dsn, strings.Join(poolDsn, " "))
		}
	}
	return strings.Join(dsn, " ")
}

func openPgxPool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	parsedCfg, parseErr := pgxpool.ParseConfig(connString)
	if parseErr != nil {
		return nil, fmt.Errorf("parse: %w", parseErr)
	}
	parsedCfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, fmt.Sprintf("SET search_path TO %s, %s", SchemaName, river.SchemaName))
		return err
	}
	pool, poolErr := pgxpool.NewWithConfig(ctx, parsedCfg)
	if poolErr != nil {
		return nil, fmt.Errorf("create: %w", poolErr)
	}
	if pingErr := pool.Ping(ctx); pingErr != nil {
		log.Error().Err(pingErr).Msg("failed to ping postgres")
	}

	return pool, nil
}
