package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"text/template"

	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/sqltool"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"
	"github.com/rezible/rezible/internal/postgres/migrations"

	entmigrate "github.com/rezible/rezible/ent/migrate"
)

func UpdateMigrationsChecksum() error {
	dir, dirErr := getGolangMigrateDir()
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

func createSchemaMigration(ctx context.Context, pool *PgxPool, name string) error {
	dir, dirErr := getGolangMigrateDir()
	if dirErr != nil {
		return fmt.Errorf("get golang migrate dir: %w", dirErr)
	}
	formatter, fmtErr := makeFormatter(name)
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
	return withDbFromPool(pool, func(db *sql.DB) error {
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

func GetCurrentMigrationStatus(ctx context.Context, q dialect.ExecQuerier) (*MigrationStatus, error) {
	var status MigrationStatus

	latest, latestErr := migrations.GetLatestMigrationVersion()
	if latestErr != nil {
		return nil, fmt.Errorf("getting latest version: %w", latestErr)
	}
	status.LatestVersion = latest

	var rows sql.Rows
	queryErr := q.Query(ctx, `SELECT version, dirty FROM rezible.migrations LIMIT 1`, nil, &rows)
	if queryErr != nil && !errors.Is(queryErr, sql.ErrNoRows) {
		return nil, fmt.Errorf("querying: %w", queryErr)
	}
	if scanErr := rows.Scan(&status.CurrentVersion, &status.Dirty); scanErr != nil {
		return nil, fmt.Errorf("scanning rows: %w", scanErr)
	}
	closeDatabaseResource("migration status rows", &rows)

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
	return fmt.Sprintf("current=%d latest=%d dirty=%t pending=%t", ms.CurrentVersion, ms.LatestVersion, ms.Dirty, ms.pending())
}

func makeFormatter(name string) (migrate.Formatter, error) {
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

func getGolangMigrateDir() (*sqltool.GolangMigrateDir, error) {
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
