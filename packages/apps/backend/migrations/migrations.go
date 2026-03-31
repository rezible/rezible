package migrations

import (
	"embed"
	"fmt"
	"io/fs"

	"github.com/golang-migrate/migrate/v4/source"
)

//go:embed *
var FS embed.FS

//go:embed 0_bootstrap.sql.tpl
var BootstrapQueryTemplate string

const Path = "."

const OutputDir = "migrations"

func GetLatestVersion() (uint, error) {
	entries, readErr := fs.ReadDir(FS, Path)
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
