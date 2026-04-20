package postgres

import (
	"database/sql"
	"errors"
	"io"
	"log/slog"

	"entgo.io/ent/dialect"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/rezible/rezible/ent"
)

func MakeEntClient(driver dialect.Driver) *ent.Client {
	client := ent.NewClient(ent.Driver(driver))
	client.Use(ensureTenantIdSetHook)
	client.Intercept(setTenantContextInterceptor())
	return client
}

func withDbFromPool(pool *pgxpool.Pool, fn func(db *sql.DB) error) error {
	db := stdlib.OpenDBFromPool(pool)
	defer closeResource(db, "*sql.DB (from pool)")
	return fn(db)
}

func closeResource(r io.Closer, name string) {
	if closeErr := r.Close(); closeErr != nil && !errors.Is(closeErr, sql.ErrConnDone) {
		slog.Error("failed to close "+name, "error", closeErr)
	}
}
