package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/internal/postgres/migrations"
	"github.com/rezible/rezible/internal/postgres/river"

	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"

	migratelib "github.com/golang-migrate/migrate/v4"
	postgresmigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

const (
	SchemaName      = "rezible"
	migrationsTable = "migrations"
)

var postgresMigrateConfig = &postgresmigrate.Config{
	SchemaName:      SchemaName,
	MigrationsTable: migrationsTable,
}

type MigrationService struct {
	cfg  rez.PostgresConfig
	pool *pgxpool.Pool
}

func NewMigrationService(cfg rez.PostgresConfig) *MigrationService {
	return &MigrationService{cfg: cfg}
}

func (m *MigrationService) getPool(ctx context.Context) (*pgxpool.Pool, error) {
	if m.pool == nil {
		pool, poolErr := MakePgxPool(ctx, m.cfg, true)
		if poolErr != nil {
			return nil, poolErr
		}
		m.pool = pool
	}
	return m.pool, nil
}

func (m *MigrationService) requireUpToDate(ctx context.Context) error {
	pool, poolErr := m.getPool(ctx)
	if poolErr != nil {
		return fmt.Errorf("get pool: %w", poolErr)
	}
	return withDbFromPool(pool, func(db *sql.DB) error {
		status, statusErr := GetCurrentMigrationStatus(ctx, db)
		if statusErr != nil {
			return fmt.Errorf("get current migration status: %w", statusErr)
		}
		if status.Dirty {
			return fmt.Errorf("database migrations are dirty: %s", status)
		}
		if status.pending() {
			return fmt.Errorf("database migrations are pending: %s", status)
		}
		return nil
	})
}

func (m *MigrationService) Run(ctx context.Context, direction string) error {
	pool, poolErr := m.getPool(ctx)
	if poolErr != nil {
		return fmt.Errorf("get pool: %w", poolErr)
	}

	schemaErr := withDbFromPool(pool, func(db *sql.DB) error {
		return m.runSchemaMigration(ctx, db, direction)
	})
	if schemaErr != nil && !errors.Is(schemaErr, migratelib.ErrNoChange) {
		return fmt.Errorf("schema migration: %w", schemaErr)
	}

	riverErr := river.RunMigration(ctx, pool, direction)
	if riverErr != nil {
		return fmt.Errorf("river migration: %w", riverErr)
	}

	return nil
}

func (m *MigrationService) runSchemaMigration(ctx context.Context, db *sql.DB, direction string) error {
	migrateUp := direction == "up"
	if !migrateUp && direction != "down" {
		return fmt.Errorf("unsupported migration direction: %s", direction)
	}
	conn, connErr := db.Conn(ctx)
	if connErr != nil {
		return fmt.Errorf("connect to postgres: %w", connErr)
	}
	defer closeDatabaseResource("db conn", conn)
	return m.withMigrator(ctx, conn, func(m *migratelib.Migrate) error {
		if migrateUp {
			return m.Up()
		}
		return m.Down()
	})
}

func (m *MigrationService) withMigrator(ctx context.Context, conn *sql.Conn, fn func(*migratelib.Migrate) error) error {
	srcDriver, srcDriverErr := iofs.New(migrations.FS, migrations.EmbedFSDir)
	if srcDriverErr != nil {
		return fmt.Errorf("load embedded migration source: %w", srcDriverErr)
	}

	dbDriver, dbDriverErr := postgresmigrate.WithConnection(ctx, conn, postgresMigrateConfig)
	if dbDriverErr != nil {
		closeDatabaseResource("embedded fs migration source driver", srcDriver)
		return fmt.Errorf("create postgres migration driver: %w", dbDriverErr)
	}

	migrator, migratorErr := migratelib.NewWithInstance("iofs", srcDriver, "postgres", dbDriver)

	defer func(m *migratelib.Migrate) {
		if migrator == nil {
			closeDatabaseResource("migration embedded fs source driver", srcDriver)
			closeDatabaseResource("migration db driver", dbDriver)
		} else {
			if srcErr, dbErr := migrator.Close(); srcErr != nil || dbErr != nil {
				slog.ErrorContext(ctx, "failed to close migrator",
					"srcErr", srcErr,
					"dbErr", dbErr,
				)
			}
		}
	}(migrator)

	if migratorErr != nil {
		return fmt.Errorf("create migrator: %w", migratorErr)
	}

	return fn(migrator)
}
