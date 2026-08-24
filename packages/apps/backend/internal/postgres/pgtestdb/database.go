package pgtestdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"runtime/debug"
	"strconv"
	"sync"

	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/peterldowns/pgtestdb"
	"github.com/peterldowns/pgtestdb/migrators/golangmigrator"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/internal/postgres/migrations"
	"github.com/rezible/rezible/internal/postgres/river"
)

const (
	riverModulePath = "github.com/riverqueue/river"
	setupSchemasSQL = `
		CREATE SCHEMA IF NOT EXISTS %[1]s;
		CREATE SCHEMA IF NOT EXISTS %[2]s;
		GRANT USAGE ON SCHEMA %[1]s, %[2]s TO %[3]s;
		ALTER DEFAULT PRIVILEGES IN SCHEMA %[1]s GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO %[3]s;
		ALTER DEFAULT PRIVILEGES IN SCHEMA %[1]s GRANT USAGE, SELECT ON SEQUENCES TO %[3]s;
		ALTER DEFAULT PRIVILEGES IN SCHEMA %[2]s GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO %[3]s;
		ALTER DEFAULT PRIVILEGES IN SCHEMA %[2]s GRANT USAGE, SELECT ON SEQUENCES TO %[3]s;
		ALTER ROLE %[3]s SET search_path TO %[1]s, %[2]s;`
)

type Database struct {
	config    rez.PostgresConfig
	lifecycle *testLifecycle
}

func New(cfg rez.PostgresConfig) (*Database, error) {
	lifecycle := &testLifecycle{}
	var testConfig *pgtestdb.Config
	createErr := catchFatal(func() {
		tdbCfg := pgtestdb.Config{
			DriverName: "pgx",
			Host:       cfg.Host,
			Port:       strconv.FormatUint(uint64(cfg.Port), 10),
			Database:   cfg.Database,
			Options:    fmt.Sprintf("sslmode=%s&search_path=%s,%s", cfg.SSLMode, postgres.SchemaName, river.SchemaName),
			User:       cfg.AdminRole.Name,
			Password:   cfg.AdminRole.Password,
			TestRole: &pgtestdb.Role{
				Username: cfg.AppRole.Name,
				Password: cfg.AppRole.Password,
			},
		}
		testConfig = pgtestdb.Custom(lifecycle, tdbCfg, newMigrator(cfg.SSLMode))
	})
	if createErr != nil {
		return nil, errors.Join(createErr, lifecycle.shutdown())
	}

	testDatabaseConfig, configErr := makePostgresConfig(cfg, *testConfig)
	if configErr != nil {
		return nil, errors.Join(configErr, lifecycle.shutdown())
	}
	return &Database{config: testDatabaseConfig, lifecycle: lifecycle}, nil
}

func (d *Database) Config() rez.PostgresConfig {
	return d.config
}

func (d *Database) Shutdown() error {
	if d == nil {
		return nil
	}
	return d.lifecycle.shutdown()
}

func makePostgresConfig(base rez.PostgresConfig, cfg pgtestdb.Config) (rez.PostgresConfig, error) {
	port, portErr := strconv.ParseUint(cfg.Port, 10, 16)
	if portErr != nil {
		return rez.PostgresConfig{}, fmt.Errorf("parse test database port: %w", portErr)
	}
	base.Host = cfg.Host
	base.Port = uint16(port)
	base.Database = cfg.Database
	base.AppRole = rez.PostgresRoleConfig{Name: cfg.User, Password: cfg.Password}
	return base, nil
}

type migrator struct {
	schema  *golangmigrator.GolangMigrator
	sslMode string
}

func newMigrator(sslMode string) *migrator {
	return &migrator{
		schema:  golangmigrator.New(migrations.EmbedFSDir, golangmigrator.WithFS(migrations.FS)),
		sslMode: sslMode,
	}
}

func (m *migrator) Hash() (string, error) {
	schemaHash, hashErr := m.schema.Hash()
	if hashErr != nil {
		return "", hashErr
	}
	return schemaHash + ":river:" + riverVersion(), nil
}

func riverVersion() string {
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		for _, dependency := range buildInfo.Deps {
			if dependency.Path == riverModulePath {
				if dependency.Replace != nil {
					return dependency.Replace.Path + "@" + dependency.Replace.Version
				}
				return dependency.Version
			}
		}
	}
	return "unknown"
}

func (m *migrator) Migrate(ctx context.Context, db *sql.DB, cfg pgtestdb.Config) error {
	applicationSchema := pgx.Identifier{postgres.SchemaName}.Sanitize()
	jobsSchema := pgx.Identifier{river.SchemaName}.Sanitize()
	applicationRole := pgx.Identifier{cfg.TestRole.Username}.Sanitize()
	setupQuery := fmt.Sprintf(setupSchemasSQL, applicationSchema, jobsSchema, applicationRole)
	if _, setupErr := db.ExecContext(ctx, setupQuery); setupErr != nil {
		return fmt.Errorf("setup test database schema: %w", setupErr)
	}
	if migrationErr := m.schema.Migrate(ctx, db, cfg); migrationErr != nil {
		return migrationErr
	}

	riverConfig, configErr := makePostgresConfig(rez.PostgresConfig{SSLMode: m.sslMode}, cfg)
	if configErr != nil {
		return configErr
	}
	pool, poolErr := postgres.MakePgxPool(ctx, riverConfig, false)
	if poolErr != nil {
		return fmt.Errorf("connect for river migrations: %w", poolErr)
	}
	defer pool.Close()
	return river.RunMigration(ctx, pool, "up")
}

// testLifecycle adapts pgtestdb's testing interface for non-test callers. The
// upstream package reports failures through Fatalf and owns database cleanup;
// this adapter turns those failures back into errors for Database's API.
type testLifecycle struct {
	cleanups  []func()
	closeOnce sync.Once
	closeErr  error
}

func (l *testLifecycle) Cleanup(cleanup func()) {
	l.cleanups = append(l.cleanups, cleanup)
}

func (*testLifecycle) Failed() bool {
	return false
}

func (*testLifecycle) Fatalf(format string, args ...any) {
	panic(fatalError{err: fmt.Errorf(format, args...)})
}

func (*testLifecycle) Helper() {}

func (*testLifecycle) Logf(string, ...any) {}

func (l *testLifecycle) shutdown() error {
	l.closeOnce.Do(func() {
		for index := len(l.cleanups) - 1; index >= 0; index-- {
			l.closeErr = errors.Join(l.closeErr, catchFatal(l.cleanups[index]))
		}
	})
	return l.closeErr
}

type fatalError struct {
	err error
}

func catchFatal(action func()) (err error) {
	defer func() {
		recovered := recover()
		if recovered == nil {
			return
		}
		failure, ok := recovered.(fatalError)
		if !ok {
			panic(recovered)
		}
		err = failure.err
	}()

	action()
	return nil
}
