package river

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river/riverdriver"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
	"github.com/riverqueue/river/rivershared/sqlctemplate"
)

func RunMigration(ctx context.Context, pool *pgxpool.Pool, direction string) error {
	cfg := &rivermigrate.Config{
		Line:   riverdriver.MigrationLineMain,
		Schema: SchemaName,
		Logger: nil,
	}
	rm, rmErr := rivermigrate.New(riverpgxv5.New(pool), cfg)
	if rmErr != nil {
		return fmt.Errorf("rivermigrate: %w", rmErr)
	}
	rd := rivermigrate.DirectionUp
	if direction == "down" {
		rd = rivermigrate.DirectionDown
	} else if direction != "up" {
		return fmt.Errorf("unknown direction: %s", direction)
	}
	opts := &rivermigrate.MigrateOpts{
		DryRun:        false,
		MaxSteps:      0,
		TargetVersion: 0,
	}
	_, mErr := rm.Migrate(ctx, rd, opts)
	if mErr != nil {
		return fmt.Errorf("migrate: %w", mErr)
	}
	return nil
}

const (
	migrationsDir      = "./migrations"
	initialRiverSchema = 1
)

func GenerateMigration(name string) error {
	migrator, migErr := rivermigrate.New(riverpgxv5.New(nil), nil)
	if migErr != nil {
		return fmt.Errorf("create migrator: %w", migErr)
	}

	currentVersion, currentErr := currentVendoredVersion()
	if currentErr != nil {
		return currentErr
	}

	pending := pendingMigrations(migrator.AllVersions(), currentVersion)
	if len(pending) == 0 {
		return nil
	}

	version, versionErr := nextMigrationVersion()
	if versionErr != nil {
		return versionErr
	}

	upSQL, upErr := renderMigrations(pending, rivermigrate.DirectionUp)
	if upErr != nil {
		return upErr
	}

	reversed := slices.Clone(pending)
	slices.Reverse(reversed)
	downSQL, downErr := renderMigrations(reversed, rivermigrate.DirectionDown)
	if downErr != nil {
		return downErr
	}

	if writeErr := os.WriteFile(filepath.Join(migrationsDir, version+"_"+name+".up.sql"), []byte(upSQL+"\n"), 0o644); writeErr != nil {
		return fmt.Errorf("write river up migration: %w", writeErr)
	}
	if writeErr := os.WriteFile(filepath.Join(migrationsDir, version+"_"+name+".down.sql"), []byte(downSQL+"\n"), 0o644); writeErr != nil {
		return fmt.Errorf("write river down migration: %w", writeErr)
	}

	return nil
}

func pendingMigrations(all []rivermigrate.Migration, currentVersion int) []rivermigrate.Migration {
	var pending []rivermigrate.Migration
	for _, migration := range all {
		if migration.Version <= max(currentVersion, initialRiverSchema) {
			continue
		}
		pending = append(pending, migration)
	}
	return pending
}

func migrationComment(line string, version int, direction rivermigrate.Direction) string {
	return fmt.Sprintf("-- River %s migration %03d [%s]", line, version, direction)
}

var migrationCommentRE = regexp.MustCompile(`(?m)^-- River main migration (\d{3}) \[(up|down)\]$`)

func currentVendoredVersion() (int, error) {
	entries, readErr := os.ReadDir(migrationsDir)
	if readErr != nil {
		return 0, fmt.Errorf("read migrations dir: %w", readErr)
	}

	maxVersion := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		contents, contentsErr := os.ReadFile(filepath.Join(migrationsDir, entry.Name()))
		if contentsErr != nil {
			return 0, fmt.Errorf("read migration file %s: %w", entry.Name(), contentsErr)
		}

		matches := migrationCommentRE.FindAllStringSubmatch(string(contents), -1)
		for _, match := range matches {
			version, parseErr := strconv.Atoi(match[1])
			if parseErr != nil {
				return 0, fmt.Errorf("parse river migration version from %s: %w", entry.Name(), parseErr)
			}
			if version > maxVersion {
				maxVersion = version
			}
		}
	}

	return maxVersion, nil
}

func renderMigrations(migrations []rivermigrate.Migration, direction rivermigrate.Direction) (string, error) {
	var sections []string
	var replacer sqlctemplate.Replacer

	for _, migration := range migrations {
		sql := migration.SQLUp
		if direction == rivermigrate.DirectionDown {
			sql = migration.SQLDown
		}

		if strings.Contains(sql, "/* TEMPLATE: schema */") {
			ctx := sqlctemplate.WithReplacements(context.Background(), map[string]sqlctemplate.Replacement{
				"schema": {Stable: true, Value: ""},
			}, nil)
			rendered, _, renderErr := replacer.RunSafely(ctx, "$", sql, nil)
			if renderErr != nil {
				return "", fmt.Errorf("render river migration %03d [%s]: %w", migration.Version, direction, renderErr)
			}
			sql = rendered
		}

		sections = append(sections, fmt.Sprintf("%s\n%s",
			migrationComment(riverdriver.MigrationLineMain, migration.Version, direction),
			strings.TrimSpace(sql),
		))
	}

	return strings.Join(sections, "\n\n"), nil
}

func nextMigrationVersion() (string, error) {
	ts := time.Now().UTC()
	for i := 0; i < 100; i++ {
		version := ts.Add(time.Duration(i) * time.Second).Format("20060102150405")
		matches, globErr := filepath.Glob(filepath.Join(migrationsDir, version+"_*.sql"))
		if globErr != nil {
			return "", fmt.Errorf("glob migration versions: %w", globErr)
		}
		if len(matches) == 0 {
			return version, nil
		}
	}
	return "", fmt.Errorf("unable to allocate a unique migration timestamp")
}
