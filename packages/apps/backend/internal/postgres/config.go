package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/internal/postgres/river"
)

func LoadConfig(cl rez.ConfigLoader) (Config, error) {
	cfg := Config{
		Host:     "postgres",
		Port:     5432,
		Database: "rezible",
		SSLMode:  "require",
	}
	return cfg, cl.Unmarshal("postgres", &cfg)
}

type Config struct {
	Host      string      `cfg:"host"`
	Port      uint16      `cfg:"port"`
	Database  string      `cfg:"database"`
	AppRole   RoleConfig  `cfg:"role_app"`
	AdminRole RoleConfig  `cfg:"role_admin"`
	SSLMode   string      `cfg:"sslmode"`
	Pool      *PoolConfig `cfg:"pool"`
}

type PoolConfig struct {
	MaxConns int32 `cfg:"pool_max_conns"`
}

type RoleConfig struct {
	Name     string `cfg:"name"`
	Password string `cfg:"password"`
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
		slog.Error("failed to ping postgres", "error", pingErr)
	}

	return pool, nil
}
