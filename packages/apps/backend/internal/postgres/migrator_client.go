package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rezible/rezible/internal/postgres/migrations"
	"github.com/rezible/rezible/internal/postgres/river"

	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"

	migratelib "github.com/golang-migrate/migrate/v4"
	postgresmigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	entmigrate "github.com/rezible/rezible/ent/migrate"
)

const (
	SchemaName      = "rezible"
	migrationsTable = "migrations"
)

var postgresMigrateConfig = &postgresmigrate.Config{
	SchemaName:      SchemaName,
	MigrationsTable: migrationsTable,
}

func RequireCurrentMigrations(ctx context.Context, cfg Config) error {
	mg := &MigratorClient{connectionString: cfg.getDsn(cfg.AppRole)}
	status, statusErr := mg.GetCurrentStatus(ctx)
	if statusErr != nil {
		return fmt.Errorf("get current migration status: %w", statusErr)
	}
	return status.requireUpToDate()
}

func CreateSchemaMigration(ctx context.Context, name string) error {
	c, cErr := newAdminMigratorClient()
	if cErr != nil {
		return fmt.Errorf("make client: %w", cErr)
	}
	return c.CreateSchemaMigration(ctx, name)
}

func RunMigrations(ctx context.Context, direction string) error {
	c, cErr := newAdminMigratorClient()
	if cErr != nil {
		return fmt.Errorf("make client: %w", cErr)
	}
	return c.Run(ctx, direction)
}

type MigratorClient struct {
	connectionString string
}

func newAdminMigratorClient() (*MigratorClient, error) {
	cfg, cfgErr := LoadConfig()
	if cfgErr != nil || cfg.AdminRole.Name == "" {
		return nil, fmt.Errorf("config error: %w", cfgErr)
	}

	m := &MigratorClient{
		connectionString: cfg.getDsn(cfg.AdminRole),
	}
	return m, nil
}

func (m *MigratorClient) withDb(ctx context.Context, fn func(db *sql.DB) error) error {
	return m.withDbPool(ctx, func(pool *pgxpool.Pool) error {
		return withDbFromPool(pool, fn)
	})
}

func (m *MigratorClient) withDbPool(ctx context.Context, fn func(*pgxpool.Pool) error) error {
	pool, poolErr := openPgxPool(ctx, m.connectionString)
	if poolErr != nil {
		return fmt.Errorf("open pgxpool: %w", poolErr)
	}
	defer pool.Close()
	return fn(pool)
}

func (m *MigratorClient) Run(ctx context.Context, direction string) error {
	return m.withDbPool(ctx, func(pool *pgxpool.Pool) error {
		if mErr := m.runSchemaMigration(ctx, pool, direction); mErr != nil {
			return fmt.Errorf("schema migration: %w", mErr)
		}
		if riverErr := river.RunMigration(ctx, pool, "up"); riverErr != nil {
			return fmt.Errorf("river migration: %w", riverErr)
		}
		return nil
	})
}

func (m *MigratorClient) CreateSchemaMigration(ctx context.Context, name string) error {
	dir, dirErr := getGolangMigrateDir()
	if dirErr != nil {
		return fmt.Errorf("getting output dir: %w", dirErr)
	}
	formatter, fmtErr := makeMigrationFormatter(name)
	if fmtErr != nil {
		return fmt.Errorf("creating formatter: %w", fmtErr)
	}
	opts := []schema.MigrateOption{
		schema.WithDir(dir),
		schema.WithDialect(dialect.Postgres),
		schema.WithDiffOptions(),
		schema.WithMigrationMode(schema.ModeInspect),
		schema.WithFormatter(formatter),
	}
	return m.withDb(ctx, func(db *sql.DB) error {
		driver := entsql.OpenDB(dialect.Postgres, db)
		mig, mErr := schema.NewMigrate(driver, opts...)
		if mErr != nil {
			return fmt.Errorf("creating migrate: %w", mErr)
		}
		if diffErr := mig.NamedDiff(ctx, name, entmigrate.Tables...); diffErr != nil {
			return fmt.Errorf("diff: %w", diffErr)
		}
		return nil
	})
}

func (m *MigratorClient) runSchemaMigration(ctx context.Context, pool *pgxpool.Pool, direction string) error {
	return withDbFromPool(pool, func(db *sql.DB) error {
		return m.withMigrator(ctx, db, func(m *migratelib.Migrate) error {
			var migrateErr error
			if direction == "up" {
				migrateErr = m.Up()
			} else {
				migrateErr = fmt.Errorf("direction %q is not supported", direction)
			}
			if migrateErr != nil && !errors.Is(migrateErr, migratelib.ErrNoChange) {
				return fmt.Errorf("migrate: %w", migrateErr)
			}
			return nil
		})
	})
}

func (m *MigratorClient) withMigrator(ctx context.Context, db *sql.DB, fn func(*migratelib.Migrate) error) error {
	conn, connErr := db.Conn(ctx)
	if connErr != nil {
		return fmt.Errorf("connect to postgres: %w", connErr)
	}
	defer closeResource(conn, "migrations db.Conn")

	sourceDriver, sourceDriverErr := iofs.New(migrations.FS, migrations.EmbedFSDir)
	if sourceDriverErr != nil {
		return fmt.Errorf("load embedded migration source: %w", sourceDriverErr)
	}

	dbDriver, dbDriverErr := postgresmigrate.WithConnection(ctx, conn, postgresMigrateConfig)
	if dbDriverErr != nil {
		closeResource(sourceDriver, "migrations iofs driver")
		return fmt.Errorf("create postgres migration driver: %w", dbDriverErr)
	}

	migrator, migratorErr := migratelib.NewWithInstance("iofs", sourceDriver, "postgres", dbDriver)
	if migratorErr != nil {
		return fmt.Errorf("create migrator: %w", migratorErr)
	}
	defer func(mig *migratelib.Migrate) {
		if mig != nil {
			if srcErr, dbErr := mig.Close(); srcErr != nil || dbErr != nil {
				slog.ErrorContext(ctx, "failed to close migrator", "srcErr", srcErr, "dbErr", dbErr)
			}
		}
	}(migrator)

	return fn(migrator)
}

func (m *MigratorClient) GetCurrentStatus(ctx context.Context) (*MigrationStatus, error) {
	latest, latestErr := getLatestMigrationVersion()
	if latestErr != nil {
		return nil, fmt.Errorf("getting latest version: %w", latestErr)
	}

	status := MigrationStatus{
		LatestVersion: latest,
	}
	scanDbStatus := func(db *sql.DB) error {
		row := db.QueryRowContext(ctx, `SELECT version, dirty FROM migrations LIMIT 1`)
		scanErr := row.Scan(&status.CurrentVersion, &status.Dirty)
		if scanErr != nil && !errors.Is(scanErr, sql.ErrNoRows) {
			return scanErr
		}
		return nil
	}
	if scanErr := m.withDb(ctx, scanDbStatus); scanErr != nil {
		return nil, fmt.Errorf("getting db migration status: %w", scanErr)
	}

	return &status, nil
}

type MigrationStatus struct {
	CurrentVersion uint
	LatestVersion  uint
	Dirty          bool
}

func (ms MigrationStatus) pending() bool {
	return ms.Dirty || ms.CurrentVersion < ms.LatestVersion
}

func (ms MigrationStatus) String() string {
	return fmt.Sprintf("current=%d latest=%d dirty=%t pending=%t",
		ms.CurrentVersion, ms.LatestVersion, ms.Dirty, ms.pending())
}

func (ms MigrationStatus) requireUpToDate() error {
	if ms.Dirty {
		return fmt.Errorf("database migrations are dirty: %s", ms)
	}
	if ms.pending() {
		return fmt.Errorf("database migrations are pending: %s", ms)
	}
	return nil
}
