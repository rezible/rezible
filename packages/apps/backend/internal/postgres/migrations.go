package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"ariga.io/atlas/sql/sqltool"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rezible/rezible/internal/postgres/river"
	"github.com/rs/zerolog/log"

	_ "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/jackc/pgx/v5/stdlib"

	migratelib "github.com/golang-migrate/migrate/v4"
	postgresmigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	entmigrate "github.com/rezible/rezible/ent/migrate"
	"github.com/rezible/rezible/migrations"
)

const (
	SchemaName      = "rezible"
	migrationsTable = "migrations"
)

func GenerateEntMigrations(ctx context.Context, name string, dbUrlOverride string) error {
	dir, dirErr := sqltool.NewGolangMigrateDir(migrations.OutputDir)
	if dirErr != nil {
		return fmt.Errorf("failed creating atlas migration directory: %w", dirErr)
	}
	opts := []schema.MigrateOption{
		schema.WithDir(dir),
		schema.WithDialect(dialect.Postgres),
		schema.WithSchemaName(SchemaName),
		schema.WithDiffOptions(),
		schema.WithMigrationMode(schema.ModeInspect),
	}
	return withMigratorDb(ctx, dbUrlOverride, func(db *sql.DB) error {
		m, mErr := schema.NewMigrate(entsql.OpenDB(dialect.Postgres, db), opts...)
		if mErr != nil {
			return fmt.Errorf("failed creating migrate: %w", mErr)
		}
		return m.NamedDiff(ctx, name, entmigrate.Tables...)
	})
}

func RunMigrations(ctx context.Context, direction string, dbUrlOverride string) error {
	if direction == "down" {
		panic("only migrating up is supported")
	}

	return withMigrationPool(ctx, dbUrlOverride, func(pool *pgxpool.Pool) error {
		if mErr := runMigration(ctx, pool, direction); mErr != nil {
			return fmt.Errorf("migration: %w", mErr)
		}
		if riverErr := river.RunMigration(ctx, pool, direction); riverErr != nil {
			return fmt.Errorf("river migration: %w", riverErr)
		}

		return nil
	})
}

func withMigratorDb(ctx context.Context, dbUrlOverride string, fn func(db *sql.DB) error) error {
	return withMigrationPool(ctx, dbUrlOverride, func(pool *pgxpool.Pool) error {
		db := stdlib.OpenDBFromPool(pool)
		defer closeSqlDb(db)
		return fn(db)
	})
}

func withMigrationPool(ctx context.Context, connStringOverride string, fn func(*pgxpool.Pool) error) error {
	connString := connStringOverride
	if connString == "" {
		cfg, cfgErr := LoadConfig()
		if cfgErr != nil || cfg.Migrations == nil {
			return fmt.Errorf("postgres migrations config error: %w", cfgErr)
		}
		cfg.User = cfg.Migrations.User
		cfg.Password = cfg.Migrations.Password
		connString = cfg.getDsn()
	}
	pool, poolErr := openPgxPool(ctx, connString)
	if poolErr != nil {
		return fmt.Errorf("open pgxpool: %w", poolErr)
	}
	defer pool.Close()
	return fn(pool)
}

func runMigration(ctx context.Context, pool *pgxpool.Pool, direction string) error {
	db := stdlib.OpenDBFromPool(pool)
	defer closeSqlDb(db)

	conn, connErr := db.Conn(ctx)
	if connErr != nil {
		return fmt.Errorf("connect to postgres: %w", connErr)
	}
	defer conn.Close()

	sourceDriver, sourceErr := iofs.New(migrations.FS, migrations.Path)
	if sourceErr != nil {
		return fmt.Errorf("load embedded migration source: %w", sourceErr)
	}

	pgmCfg := &postgresmigrate.Config{
		SchemaName:      SchemaName,
		MigrationsTable: migrationsTable,
	}
	dbDriver, driverErr := postgresmigrate.WithConnection(ctx, conn, pgmCfg)
	if driverErr != nil {
		return fmt.Errorf("create postgres migration driver: %w", driverErr)
	}

	m, migratorErr := migratelib.NewWithInstance("iofs", sourceDriver, "postgres", dbDriver)
	if migratorErr != nil {
		return fmt.Errorf("create migrator: %w", migratorErr)
	}
	defer func(m *migratelib.Migrate) {
		closeSrcErr, closeDbErr := m.Close()
		if closeSrcErr != nil || closeDbErr != nil {
			log.Error().
				AnErr("closeSrcErr", closeSrcErr).
				AnErr("closeDbErr", closeDbErr).
				Msg("failed to close migrator")
		}
	}(m)

	if direction == "up" {
		if upErr := m.Up(); upErr != nil && !errors.Is(upErr, migratelib.ErrNoChange) {
			return fmt.Errorf("apply migrations: %w", upErr)
		}
	} else {
		return fmt.Errorf("direction %q is not supported", direction)
	}
	return nil
}

type MigrationStatus struct {
	CurrentVersion uint
	LatestVersion  uint
	Dirty          bool
}

func (ms MigrationStatus) Pending() bool {
	return ms.Dirty || ms.CurrentVersion < ms.LatestVersion
}

func (ms MigrationStatus) String() string {
	return fmt.Sprintf("current=%d latest=%d dirty=%t pending=%t",
		ms.CurrentVersion, ms.LatestVersion, ms.Dirty, ms.Pending())
}

func GetMigrationStatus(ctx context.Context, dbUrlOverride string) (MigrationStatus, error) {
	var status MigrationStatus
	dbErr := withMigratorDb(ctx, dbUrlOverride, func(db *sql.DB) error {
		ms, msErr := getMigrationStatusFromDB(ctx, db)
		if msErr != nil {
			return fmt.Errorf("get status from db: %w", msErr)
		}
		status = ms
		return nil
	})
	return status, dbErr
}

func getMigrationStatusFromDB(ctx context.Context, db *sql.DB) (MigrationStatus, error) {
	var s MigrationStatus
	latest, latestErr := migrations.GetLatestVersion()
	if latestErr != nil {
		return s, latestErr
	}
	s.LatestVersion = latest

	row := db.QueryRowContext(ctx, "SELECT version, dirty FROM rezible.migrations LIMIT 1")
	if scanErr := row.Scan(&s.CurrentVersion, &s.Dirty); scanErr != nil && !errors.Is(scanErr, sql.ErrNoRows) {
		return s, scanErr
	}

	return s, nil
}

func closeSqlDb(db *sql.DB) {
	if err := db.Close(); err != nil {
		log.Error().Err(err).Msg("failed to close database connection")
	}
}
