package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/internal/postgres/river"
)

func makeConnectionString(cfg rez.PostgresConfig, admin bool) string {
	role := cfg.AppRole
	if admin {
		role = cfg.AdminRole
	}
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

	if cfg.PoolMaxConns != 0 {
		dsn = append(dsn, fmt.Sprintf("pool_max_conns='%d'", cfg.PoolMaxConns))
	}

	return strings.Join(dsn, " ")
}

func MakePgxPool(ctx context.Context, cfg rez.PostgresConfig, admin bool) (*pgxpool.Pool, error) {
	parsedCfg, parseErr := pgxpool.ParseConfig(makeConnectionString(cfg, admin))
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

func withDbFromPool(pool *pgxpool.Pool, fn func(db *sql.DB) error) error {
	db := stdlib.OpenDBFromPool(pool)
	defer closeDatabaseResource("db from pool", db)
	return fn(db)
}
