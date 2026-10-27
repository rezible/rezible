package migrations_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/sqltool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/internal/postgres/migrations"
	"github.com/rezible/rezible/internal/postgres/river"
	"github.com/rezible/rezible/test"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
	"github.com/stretchr/testify/suite"
)

type MigrationsSuite struct {
	test.Suite
}

func TestMigrationsSuite(t *testing.T) {
	suite.Run(t, &MigrationsSuite{Suite: test.NewSuite()})
}

func (s *MigrationsSuite) TestChecksum() {
	dirPath, pathErr := migrations.LocalDir()
	s.Require().NoError(pathErr)
	dir, dirErr := sqltool.NewGooseDir(dirPath)
	s.Require().NoError(dirErr)
	validateErr := migrate.Validate(dir)
	if errors.Is(validateErr, migrate.ErrChecksumMismatch) {
		s.FailNow("migration checksum mismatch; run just backend::migrations-checksum", "%v", validateErr)
	}
	s.Require().NoError(validateErr)
}

func (s *MigrationsSuite) TestDrift() {
	db := stdlib.OpenDBFromPool(s.SetupTestPool().Pool)
	defer db.Close()
	outputDir := s.T().TempDir()
	files, planErr := migrations.Plan(s.T().Context(), db, outputDir, "drift")
	s.Require().NoError(planErr)
	for _, file := range files {
		content, readErr := os.ReadFile(filepath.Join(outputDir, file))
		s.Require().NoError(readErr)
		s.Fail("migration directory differs from the schema", "%s:\n%s", file, content)
	}
}

func (s *MigrationsSuite) TestMigrationFiles() {
	destructive := regexp.MustCompile(`(?i)\bDROP\s+(TABLE|COLUMN)\b`)
	acknowledged := regexp.MustCompile(`(?m)^-- destructive: \S`)
	filename := regexp.MustCompile(`^[0-9]{4}_[a-z][a-z0-9_]*\.sql$`)
	down := regexp.MustCompile(`(?mi)^\s*--\s*\+goose\s+Down\b`)
	files, globErr := fs.Glob(migrations.FS, "*.sql")
	s.Require().NoError(globErr)
	s.Require().NotEmpty(files)
	var versions []int
	for _, file := range files {
		s.Regexp(filename, file)
		version, parseErr := goose.NumericComponent(file)
		s.Require().NoError(parseErr)
		versions = append(versions, int(version))
		content, readErr := migrations.FS.ReadFile(file)
		s.Require().NoError(readErr)
		s.False(down.Match(content), "%s: migrations are forward-only", file)
		if destructive.Match(content) {
			s.True(acknowledged.Match(content), "%s: DROP TABLE/COLUMN requires a -- destructive: <reason> line", file)
		}
	}
	sort.Ints(versions)
	for index, version := range versions {
		s.Equal(index+1, version, "migration versions must be contiguous from 1")
	}
}

func (s *MigrationsSuite) TestRiverIsCurrent() {
	pool := s.SetupTestPool()
	migrator, migratorErr := rivermigrate.New(riverpgxv5.New(pool.Pool), &rivermigrate.Config{Schema: river.SchemaName})
	s.Require().NoError(migratorErr)
	var applied int
	s.Require().NoError(pool.QueryRow(s.T().Context(), `SELECT max(version) FROM river.river_migration WHERE line = 'main'`).Scan(&applied))
	var latest int
	var missing strings.Builder
	for _, version := range migrator.AllVersions() {
		if version.Version > latest {
			latest = version.Version
		}
		if version.Version > applied {
			missing.WriteString(version.SQLUp)
			missing.WriteString("\n")
		}
	}
	s.Equal(latest, applied, "River bundled version=%d, applied version=%d; copy River's new .up.sql files into a migration:\n%s", latest, applied, missing.String())
}

func (s *MigrationsSuite) TestStatusWithoutVersionTable() {
	pool := s.SetupTestPool()
	ctx := s.T().Context()
	db := stdlib.OpenDBFromPool(pool.Pool)
	defer db.Close()
	files, pending, statusErr := migrations.Status(ctx, db)
	s.Require().NoError(statusErr)
	s.Empty(pending)
	s.NotEmpty(files)

	cfg := s.Config().Postgres
	cfg.Database = pool.Config().ConnConfig.Database
	admin, adminErr := postgres.MakePgxPool(ctx, cfg, true)
	s.Require().NoError(adminErr)
	defer admin.Shutdown()
	_, dropErr := admin.Exec(ctx, `DROP TABLE rezible.goose_db_version`)
	s.Require().NoError(dropErr)
	applied, pending, statusErr := migrations.Status(ctx, db)
	s.Require().NoError(statusErr)
	s.Empty(applied)
	s.Equal(files, pending)
	var absent bool
	s.Require().NoError(pool.QueryRow(ctx, `SELECT to_regclass('rezible.goose_db_version') IS NULL`).Scan(&absent))
	s.True(absent, "status must not create the version table")
}
