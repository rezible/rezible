package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"sort"

	"ariga.io/atlas/sql/sqltool"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"
	"github.com/rezible/rezible/ent/document"
	"github.com/rs/zerolog/log"

	_ "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/jackc/pgx/v5/stdlib"

	migratelib "github.com/golang-migrate/migrate/v4"
	postgresmigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	entmigrate "github.com/rezible/rezible/ent/migrate"
	"github.com/rezible/rezible/migrations"
)

const (
	SchemaName         = "rezible"
	migrationsTable    = "migrations"
	documentServerRole = "rez_documents"
)

func GenerateEntMigrations(ctx context.Context, name string) error {
	cfg, cfgErr := LoadConfig()
	if cfgErr != nil {
		return fmt.Errorf("config: %w", cfgErr)
	}

	pool, poolErr := openPgxPool(ctx, cfg.getDsn())
	if poolErr != nil {
		return poolErr
	}
	conn := stdlib.OpenDBFromPool(pool)
	defer func() {
		pool.Close()
		if err := conn.Close(); err != nil {
			log.Error().Err(err).Str("name", name).Msg("failed to close database connection")
		}
	}()

	dir, err := sqltool.NewGolangMigrateDir("./migrations")
	if err != nil {
		return fmt.Errorf("failed creating atlas migration directory: %w", err)
	}
	opts := []schema.MigrateOption{
		schema.WithDir(dir),
		schema.WithDialect(dialect.Postgres),
		// schema.WithSchemaName(SchemaName),
		schema.WithDiffOptions(),
		schema.WithMigrationMode(schema.ModeReplay),
	}
	m, mErr := schema.NewMigrate(entsql.OpenDB(dialect.Postgres, conn), opts...)
	if mErr != nil {
		return fmt.Errorf("failed creating migrate: %w", mErr)
	}
	if diffErr := m.NamedDiff(ctx, name, entmigrate.Tables...); diffErr != nil {
		return fmt.Errorf("diff failed: %w", diffErr)
	}
	return nil
}

func getMigrationConnString(override string) (string, error) {
	if override != "" {
		return override, nil
	}
	cfg, cfgErr := LoadConfig()
	if cfgErr != nil {
		return "", fmt.Errorf("postgres config: %w", cfgErr)
	}
	if cfg.Migrations == nil {
		return "", fmt.Errorf("migrations config is nil")
	}
	cfg.User = cfg.Migrations.User
	cfg.Password = cfg.Migrations.Password
	return cfg.getDsn(), nil
}

func withMigratorDb(ctx context.Context, dbUrlOverride string, fn func(db *sql.DB) error) error {
	connString, connStringErr := getMigrationConnString(dbUrlOverride)
	if connStringErr != nil {
		return fmt.Errorf("migrator connection: %w", connStringErr)
	}
	pool, poolErr := openPgxPool(ctx, connString)
	if poolErr != nil {
		return fmt.Errorf("open pgxpool: %w", poolErr)
	}
	defer pool.Close()
	db := stdlib.OpenDBFromPool(pool)
	defer closeSqlDb(db)
	return fn(db)
}

func closeSqlDb(db *sql.DB) {
	if err := db.Close(); err != nil {
		log.Error().Err(err).Msg("failed to close database connection")
	}
}

func RunMigrations(ctx context.Context, direction string, dbUrlOverride string) error {
	if direction != "up" {
		panic("only migrating up is supported")
	}

	return withMigratorDb(ctx, dbUrlOverride, func(db *sql.DB) error {
		m, mErr := newMigrator(db)
		if mErr != nil {
			return mErr
		}
		defer func(m *migratelib.Migrate) {
			sourceErr, dbErr := m.Close()
			if sourceErr != nil || dbErr != nil {
				log.Error().
					AnErr("sourceErr", sourceErr).
					AnErr("dbErr", dbErr).
					Msg("failed to close migrator")
			}
		}(m)

		if upErr := m.Up(); upErr != nil && !errors.Is(upErr, migratelib.ErrNoChange) {
			return fmt.Errorf("apply migrations: %w", upErr)
		}

		if grantErr := syncDocumentsTableGrants(ctx, db); grantErr != nil {
			return grantErr
		}

		return nil
	})
}

func newMigrator(db *sql.DB) (*migratelib.Migrate, error) {
	sourceDriver, sourceErr := iofs.New(migrations.FS, migrations.Path)
	if sourceErr != nil {
		return nil, fmt.Errorf("load embedded migration source: %w", sourceErr)
	}

	pgmCfg := &postgresmigrate.Config{
		MigrationsTable: migrationsTable,
		DatabaseName:    "",
		SchemaName:      SchemaName,
	}
	dbDriver, driverErr := postgresmigrate.WithInstance(db, pgmCfg)
	if driverErr != nil {
		return nil, fmt.Errorf("create postgres migration driver: %w", driverErr)
	}

	m, migratorErr := migratelib.NewWithInstance("iofs", sourceDriver, "postgres", dbDriver)
	if migratorErr != nil {
		return nil, fmt.Errorf("create migrator: %w", migratorErr)
	}
	return m, nil
}

// TODO: this should probably be handled elsewhere?
func syncDocumentsTableGrants(ctx context.Context, db *sql.DB) error {
	documentsTable := fmt.Sprintf("%s.%s", SchemaName, document.Table)
	grantQuery := fmt.Sprintf(`
DO $$
BEGIN
	IF to_regclass('%[1]s') IS NOT NULL THEN
		GRANT SELECT, INSERT, UPDATE ON TABLE %[1]s TO %[2]s;
		REVOKE DELETE, TRUNCATE, REFERENCES, TRIGGER ON TABLE %[1]s FROM %[2]s;
	END IF;
END $$;`, documentsTable, documentServerRole)
	if _, err := db.ExecContext(ctx, grantQuery); err != nil {
		return fmt.Errorf("grant documents table access to %s: %w", documentServerRole, err)
	}
	return nil
}

type MigrationStatus struct {
	CurrentVersion uint
	LatestVersion  uint
	Dirty          bool
	Pending        bool
}

func (ms MigrationStatus) String() string {
	return fmt.Sprintf("current=%d latest=%d dirty=%t pending=%t",
		ms.CurrentVersion, ms.LatestVersion, ms.Dirty, ms.Pending)
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

func requireUpToDateMigrations(ctx context.Context, db *sql.DB) error {
	status, statusErr := getMigrationStatusFromDB(ctx, db)
	if statusErr != nil {
		return fmt.Errorf("read migration status: %w", statusErr)
	}
	if status.Dirty {
		return fmt.Errorf("database migrations are dirty: %s", status)
	}
	if status.Pending {
		return fmt.Errorf("database migrations are pending: %s", status)
	}
	return nil
}

func getMigrationStatusFromDB(ctx context.Context, db *sql.DB) (MigrationStatus, error) {
	current, dirty, currentErr := getCurrentMigrationVersion(ctx, db)
	if currentErr != nil {
		return MigrationStatus{}, currentErr
	}

	latest, latestErr := latestEmbeddedMigrationVersion()
	if latestErr != nil {
		return MigrationStatus{}, latestErr
	}

	return MigrationStatus{
		CurrentVersion: current,
		LatestVersion:  latest,
		Dirty:          dirty,
		Pending:        dirty || current < latest,
	}, nil
}

func latestEmbeddedMigrationVersion() (uint, error) {
	entries, readErr := fs.ReadDir(migrations.FS, migrations.Path)
	if readErr != nil {
		return 0, fmt.Errorf("read embedded migrations: %w", readErr)
	}

	var versions []uint
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		migration, parseErr := source.DefaultParse(entry.Name())
		if parseErr != nil {
			continue
		}
		versions = append(versions, migration.Version)
	}
	if len(versions) == 0 {
		return 0, nil
	}
	sort.Slice(versions, func(i, j int) bool { return versions[i] < versions[j] })
	return versions[len(versions)-1], nil
}

func getCurrentMigrationVersion(ctx context.Context, db *sql.DB) (uint, bool, error) {
	var exists bool
	existsQuery := `SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = $1 AND table_name = $2)`
	if err := db.QueryRowContext(ctx, existsQuery, SchemaName, migrationsTable).Scan(&exists); err != nil {
		return 0, false, fmt.Errorf("check migrations table: %w", err)
	}
	if !exists {
		return 0, false, nil
	}

	var (
		version uint
		dirty   bool
	)
	row := db.QueryRowContext(ctx, `SELECT version, dirty FROM $1 LIMIT 1`, migrationsTable)
	switch err := row.Scan(&version, &dirty); {
	case errors.Is(err, sql.ErrNoRows):
		return 0, false, nil
	case err != nil:
		return 0, false, fmt.Errorf("read migration version: %w", err)
	default:
		return version, dirty, nil
	}
}
