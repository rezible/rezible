package postgres

import (
	"embed"
	"fmt"
	"io/fs"
	"text/template"

	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/sqltool"
	"github.com/golang-migrate/migrate/v4/source"
)

//go:embed migrations
var MigrationsFS embed.FS

const MigrationsDir = "migrations"

func UpdateMigrationsChecksum() error {
	dir, dirErr := sqltool.NewGolangMigrateDir(MigrationsDir)
	if dirErr != nil {
		return fmt.Errorf("getting output dir: %w", dirErr)
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

func getLatestMigrationVersion() (uint, error) {
	entries, readErr := fs.ReadDir(MigrationsFS, MigrationsDir)
	if readErr != nil {
		return 0, fmt.Errorf("read embedded migrations: %w", readErr)
	}
	var latest uint
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		migration, parseErr := source.DefaultParse(entry.Name())
		if parseErr != nil {
			continue
		}
		if migration.Version > latest {
			latest = migration.Version
		}
	}
	return latest, nil
}

func makeMigrationFormatter(name string) (migrate.Formatter, error) {
	version, versionErr := getLatestMigrationVersion()
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
