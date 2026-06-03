package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

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
	pool *PgxPool
}

func NewMigrationService(pool *PgxPool) (*MigrationService, error) {
	return &MigrationService{pool: pool}, nil
}

func (m *MigrationService) Shutdown() {
	fmt.Println("shutdown migration service")
	if m.pool != nil {
		m.pool.Close()
	}
}

func (m *MigrationService) CreateSchemaMigration(ctx context.Context, name string) error {
	return createSchemaMigration(ctx, m.pool, name)
}

func (m *MigrationService) Run(ctx context.Context, direction string) error {
	schemaErr := withDbFromPool(m.pool, func(db *sql.DB) error {
		return m.runSchemaMigration(ctx, db, direction)
	})
	if schemaErr != nil && !errors.Is(schemaErr, migratelib.ErrNoChange) {
		return fmt.Errorf("schema migration: %w", schemaErr)
	}

	riverErr := river.RunMigration(ctx, m.pool, direction)
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
