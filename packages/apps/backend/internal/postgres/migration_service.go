package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"path"
	"path/filepath"
	"runtime"
	"text/template"

	migrate "ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/sqltool"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"
	"github.com/rezible/rezible/ent/entpgx"

	rez "github.com/rezible/rezible"
	schemamigrate "github.com/rezible/rezible/ent/migrate"
	"github.com/rezible/rezible/internal/postgres/migrations"
	"github.com/rezible/rezible/internal/postgres/river"

	_ "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
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
	pool   *PgxPool
	driver dialect.Driver
}

func NewMigrationService(pool *PgxPool) (*MigrationService, error) {
	return &MigrationService{pool: pool, driver: entpgx.NewPgxPoolDriver(pool)}, nil
}

func (m *MigrationService) Shutdown() {
	if m.pool != nil {
		m.pool.Close()
	}
}

func (m *MigrationService) UpdateChecksum() error {
	dir, dirErr := m.getGolangMigrateDir()
	if dirErr != nil {
		return dirErr
	}
	sum, sumErr := dir.Checksum()
	if sumErr != nil {
		return fmt.Errorf("calculating checksum: %w", sumErr)
	}
	if writeErr := migrate.WriteSumFile(dir, sum); writeErr != nil {
		return fmt.Errorf("writing sum: %w", writeErr)
	}
	return nil
}

func (m *MigrationService) CreateSchemaMigration(ctx context.Context, name string) error {
	dir, dirErr := m.getGolangMigrateDir()
	if dirErr != nil {
		return fmt.Errorf("get golang migrate dir: %w", dirErr)
	}
	formatter, fmtErr := m.makeFormatter(name)
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
	return m.withDbFromPool(func(db *sql.DB) error {
		driver := entsql.OpenDB(dialect.Postgres, db)
		mig, mErr := schema.NewMigrate(driver, opts...)
		if mErr != nil {
			return fmt.Errorf("creating migrate: %w", mErr)
		}
		if diffErr := mig.NamedDiff(ctx, name, schemamigrate.Tables...); diffErr != nil {
			return fmt.Errorf("diff: %w", diffErr)
		}
		return nil
	})
}

func (m *MigrationService) Run(ctx context.Context, direction string) error {
	schemaErr := m.withDbFromPool(func(db *sql.DB) error {
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

func (m *MigrationService) withDbFromPool(fn func(db *sql.DB) error) error {
	if m.pool == nil {
		return fmt.Errorf("pool is nil")
	}
	db := stdlib.OpenDBFromPool(m.pool)
	defer closeDatabaseResource("db from pool", db)
	return fn(db)
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

func (m *MigrationService) GetCurrentStatus(ctx context.Context) (*rez.MigrationStatus, error) {
	var status rez.MigrationStatus

	latest, latestErr := migrations.GetLatestMigrationVersion()
	if latestErr != nil {
		return nil, fmt.Errorf("getting latest version: %w", latestErr)
	}
	status.LatestVersion = latest

	var rows sql.Rows
	queryErr := m.driver.Query(ctx, `SELECT version, dirty FROM migrations LIMIT 1`, nil, &rows)
	if queryErr != nil && !errors.Is(queryErr, sql.ErrNoRows) {
		return nil, fmt.Errorf("querying: %w", queryErr)
	}
	defer closeDatabaseResource("migration status query rows", &rows)

	if scanErr := rows.Scan(&status.CurrentVersion, &status.Dirty); scanErr != nil {
		return nil, fmt.Errorf("scanning rows: %w", scanErr)
	}

	return &status, nil
}

func (m *MigrationService) makeFormatter(name string) (migrate.Formatter, error) {
	version, versionErr := migrations.GetLatestMigrationVersion()
	if versionErr != nil {
		return nil, fmt.Errorf("failed to get latest migration version: %w", versionErr)
	}
	nextVersion := version + 1
	if name == "init" {
		nextVersion = 1
	}
	df, ok := sqltool.GolangMigrateFormatter.(migrate.TemplateFormatter)
	if !ok || len(df) != 2 {
		return nil, fmt.Errorf("unsupported migration formatter")
	}
	namePrefix := fmt.Sprintf("%04d{{ with .Name }}_{{ . }}{{ end }}", nextVersion)
	upNameTemplate := template.Must(template.New("").Parse(namePrefix + ".up.sql"))
	upContentTemplate := df[0].C
	downNameTemplate := template.Must(template.New("").Parse(namePrefix + ".down.sql"))
	downContentTemplate := df[1].C
	return migrate.NewTemplateFormatter(upNameTemplate, upContentTemplate, downNameTemplate, downContentTemplate)
}

func (m *MigrationService) getGolangMigrateDir() (*sqltool.GolangMigrateDir, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("failed to get runtime caller path")
	}
	migDir, dirErr := sqltool.NewGolangMigrateDir(path.Join(filepath.Dir(filename), "migrations"))
	if dirErr != nil {
		return nil, fmt.Errorf("new golang migrate dir: %w", dirErr)
	}
	return migDir, nil
}
