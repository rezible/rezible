package migrations

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/sqltool"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/schema/field"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"
	"github.com/pressly/goose/v3/lock"

	schemamigrate "github.com/rezible/rezible/ent/migrate"
)

//go:embed *.sql
var FS embed.FS

const (
	EmbedFSDir = "."
	SchemaName = "rezible"

	versionTableName = "goose_db_version"
	versionTable     = SchemaName + "." + versionTableName

	OutboxTableName = "message_outbox"
)

// outboxTable matches watermill's PostgreSQL queue schema, so the message outbox can use it without initializing it.
var outboxTable = schema.NewTable(OutboxTableName).
	AddPrimary(&schema.Column{Name: "offset", Type: field.TypeInt64, Increment: true}).
	AddColumn(&schema.Column{Name: "uuid", Type: field.TypeString, Size: 36}).
	AddColumn(&schema.Column{Name: "payload", Type: field.TypeJSON, Nullable: true,
		SchemaType: map[string]string{dialect.Postgres: "json"}}).
	AddColumn(&schema.Column{Name: "metadata", Type: field.TypeJSON, Nullable: true,
		SchemaType: map[string]string{dialect.Postgres: "json"}}).
	AddColumn(&schema.Column{Name: "acked", Type: field.TypeBool, Default: false}).
	AddColumn(&schema.Column{Name: "created_at", Type: field.TypeTime,
		Default: schema.Expr("CURRENT_TIMESTAMP")})

// infraTables are the tables outside the Ent schema that the migrations also own.
var infraTables = []*schema.Table{outboxTable}

func newStore() (database.Store, error) {
	return database.NewStore(goose.DialectPostgres, versionTable)
}

func newProvider(db *sql.DB, store database.Store, opts ...goose.ProviderOption) (*goose.Provider, error) {
	opts = append(opts, goose.WithStore(store), goose.WithDisableGlobalRegistry(true))
	return goose.NewProvider("", db, FS, opts...)
}

// Apply applies every pending migration in version order, each file in its own transaction.
func Apply(ctx context.Context, db *sql.DB) error {
	// Concurrent appliers must hold one session lock for the entire apply so they cannot replay the same file.
	locker, lockerErr := lock.NewPostgresSessionLocker()
	if lockerErr != nil {
		return lockerErr
	}
	store, storeErr := newStore()
	if storeErr != nil {
		return storeErr
	}
	provider, providerErr := newProvider(db, store, goose.WithSessionLocker(locker))
	if providerErr != nil {
		return providerErr
	}
	results, applyErr := provider.Up(ctx)
	if partial, ok := errors.AsType[*goose.PartialError](applyErr); ok {
		results = partial.Applied
	}
	for _, result := range results {
		slog.InfoContext(ctx, "applied migration", "file", result.Source.Path, "duration", result.Duration)
	}
	return applyErr
}

// Status returns the applied and pending file names in version order. It only reads catalogs and applied
// versions: it neither takes the applier's lock nor creates tables.
func Status(ctx context.Context, db *sql.DB) (applied, pending []string, statusErr error) {
	store, storeErr := newStore()
	if storeErr != nil {
		return nil, nil, storeErr
	}
	provider, providerErr := newProvider(db, store)
	if providerErr != nil {
		return nil, nil, providerErr
	}
	var table sql.NullString
	if queryErr := db.QueryRowContext(ctx, `SELECT to_regclass($1)::text`, versionTable).Scan(&table); queryErr != nil {
		return nil, nil, fmt.Errorf("find migration version table: %w", queryErr)
	}
	appliedVersions := mapset.NewSet[int64]()
	if table.Valid {
		versions, listErr := store.ListMigrations(ctx, db)
		if listErr != nil {
			return nil, nil, fmt.Errorf("list applied migrations: %w", listErr)
		}
		for _, version := range versions {
			if version.IsApplied {
				appliedVersions.Add(version.Version)
			}
		}
	}
	for _, source := range provider.ListSources() {
		name := path.Base(source.Path)
		if appliedVersions.Contains(source.Version) {
			applied = append(applied, name)
		} else {
			pending = append(pending, name)
		}
	}
	return applied, pending, nil
}

// Plan diffs the Ent schema and infrastructure tables against replayed, a database that has had every migration
// in FS applied and nothing else, and writes the result as the next goose migration into the directory at
// dirPath. It returns the names of the SQL files it wrote; none means no drift. It only reads from replayed.
func Plan(ctx context.Context, replayed *sql.DB, dirPath, name string) ([]string, error) {
	dir, dirErr := sqltool.NewGooseDir(dirPath)
	if dirErr != nil {
		return nil, dirErr
	}
	tables := append(append([]*schema.Table{}, schemamigrate.Tables...), infraTables...)
	// Ent inspects only the tables it is given, so a table whose entity was deleted is invisible to the diff.
	known := mapset.NewSet(versionTableName)
	for _, table := range tables {
		known.Add(table.Name)
	}
	rows, queryErr := replayed.QueryContext(ctx, `SELECT tablename FROM pg_tables WHERE schemaname = $1 ORDER BY tablename`, SchemaName)
	if queryErr != nil {
		return nil, fmt.Errorf("inspect migration tables: %w", queryErr)
	}
	defer rows.Close()
	for rows.Next() {
		var table string
		if scanErr := rows.Scan(&table); scanErr != nil {
			return nil, scanErr
		}
		if !known.Contains(table) {
			return nil, fmt.Errorf("table %q exists in the migration directory but not in the Ent schema; write a DROP TABLE migration by hand", table)
		}
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, rowsErr
	}
	before, filesErr := dir.Files()
	if filesErr != nil {
		return nil, filesErr
	}
	existing := mapset.NewSet[string]()
	for _, file := range before {
		existing.Add(file.Name())
	}
	latest, versionErr := latestVersion()
	if versionErr != nil {
		return nil, versionErr
	}
	nameTemplate := template.Must(template.New("name").Parse(fmt.Sprintf("%04d", latest+1) + "{{ with .Name }}_{{ . }}{{ end }}.sql"))
	contentTemplate := template.Must(template.New("content").Parse("-- +goose Up\n{{ range .Changes }}{{ with .Comment }}-- {{ println . }}{{ end }}{{ printf \"%s;\\n\" .Cmd }}{{ end }}"))
	formatter, formatErr := migrate.NewTemplateFormatter(nameTemplate, contentTemplate)
	if formatErr != nil {
		return nil, formatErr
	}
	// Ent's defaults filter column and index drops out of the plan, which would hide a removed field or index.
	mig, migrateErr := schema.NewMigrate(entsql.OpenDB(dialect.Postgres, replayed),
		schema.WithDialect(dialect.Postgres),
		schema.WithMigrationMode(schema.ModeInspect),
		schema.WithDropColumn(true),
		schema.WithDropIndex(true),
		schema.WithDir(dir),
		schema.WithFormatter(formatter),
	)
	if migrateErr != nil {
		return nil, fmt.Errorf("create schema planner: %w", migrateErr)
	}
	if diffErr := mig.NamedDiff(ctx, name, tables...); diffErr != nil {
		return nil, fmt.Errorf("plan migration: %w", diffErr)
	}
	after, afterErr := dir.Files()
	if afterErr != nil {
		return nil, afterErr
	}
	var written []string
	for _, file := range after {
		if strings.HasSuffix(file.Name(), ".sql") && !existing.Contains(file.Name()) {
			written = append(written, file.Name())
		}
	}
	return written, nil
}

// LocalDir is the migration directory in the source tree, which Plan and UpdateChecksum write to.
func LocalDir() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("get migration source path")
	}
	return filepath.Dir(filename), nil
}

// UpdateChecksum rewrites atlas.sum in the source tree after a migration file is edited by hand.
func UpdateChecksum() error {
	dirPath, pathErr := LocalDir()
	if pathErr != nil {
		return pathErr
	}
	dir, dirErr := sqltool.NewGooseDir(dirPath)
	if dirErr != nil {
		return dirErr
	}
	sum, sumErr := dir.Checksum()
	if sumErr != nil {
		return fmt.Errorf("calculate checksum: %w", sumErr)
	}
	return migrate.WriteSumFile(dir, sum)
}

func latestVersion() (uint, error) {
	entries, readErr := fs.ReadDir(FS, EmbedFSDir)
	if readErr != nil {
		return 0, fmt.Errorf("read embedded migrations: %w", readErr)
	}
	var latest uint
	for _, entry := range entries {
		version, parseErr := goose.NumericComponent(entry.Name())
		if parseErr != nil {
			return 0, fmt.Errorf("parse migration %q: %w", entry.Name(), parseErr)
		}
		if uint(version) > latest {
			latest = uint(version)
		}
	}
	return latest, nil
}
