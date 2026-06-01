package migrations

import (
	"embed"
	"fmt"
	"io/fs"

	"github.com/golang-migrate/migrate/v4/source"
)

//go:embed *
var FS embed.FS
var EmbedFSDir = "."

func GetLatestMigrationVersion() (uint, error) {
	entries, readErr := fs.ReadDir(FS, EmbedFSDir)
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
