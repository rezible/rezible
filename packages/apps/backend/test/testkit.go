package test

import (
	"cmp"
	"context"
	"database/sql"
	"fmt"
	"os"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/internal/postgres/migrations"
	"github.com/stretchr/testify/suite"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/peterldowns/pgtestdb"
	"github.com/peterldowns/pgtestdb/migrators/golangmigrator"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/internal/koanf"
	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/pkg/execution"
)

type Suite struct {
	suite.Suite

	opts options

	cfg rez.Config
	db  rez.Database

	SeedTenant       *ent.Tenant
	SeedOrganization *ent.Organization
	SeedUser         *ent.User
}

func NewSuite(optFns ...SuiteOption) Suite {
	opts := options{}
	for _, opt := range optFns {
		opt(&opts)
	}
	return Suite{opts: opts}
}

func (s *Suite) SetupSuite() {
	s.loadConfig()
	s.setupTestDatabase()
	s.SeedTestEntities()
}

func (s *Suite) TearDownSuite() {
	s.closeTestDatabase()
}

func (s *Suite) BeforeTest(suiteName, testName string) {
	// s.loadConfig()
}

func getEnvOr(key, fallback string) string {
	return cmp.Or(os.Getenv(key), fallback)
}

func (s *Suite) loadConfig() {
	pgAdminUser := getEnvOr("POSTGRES_ADMIN_USER", "postgres")
	pgAppUser := getEnvOr("POSTGRES_APP_USER", "rez_app")
	overrides := map[string]any{
		"postgres.host":                getEnvOr("POSTGRES_HOST", "localhost"),
		"postgres.port":                getEnvOr("POSTGRES_PORT", "7010"),
		"postgres.database":            getEnvOr("POSTGRES_APP_DB", "rezible-main"),
		"postgres.role_admin.name":     pgAdminUser,
		"postgres.role_admin.password": pgAdminUser,
		"postgres.role_app.name":       pgAppUser,
		"postgres.role_app.password":   pgAppUser,
		"postgres.sslmode":             "disable",
	}
	for k, v := range s.opts.configOverrides {
		overrides[k] = v
	}

	opts := koanf.Options{
		LoadEnvironment: true,
		SkipValidation:  true,
		Overrides:       overrides,
	}
	cfg, cfgErr := koanf.LoadConfig(s.T().Context(), opts)
	s.Require().NoError(cfgErr)
	s.cfg = *cfg
}

func (s *Suite) Config() rez.Config { return s.cfg }

func (s *Suite) Database() rez.Database { return s.db }

func (s *Suite) Client(ctx context.Context) *ent.Client { return s.db.Client(ctx) }

func (s *Suite) SystemContext() context.Context {
	return execution.NewSystemContext(s.T().Context())
}

func (s *Suite) SeedTenantContext() context.Context {
	return execution.NewTenantContext(s.T().Context(), s.SeedTenant.ID)
}

func (s *Suite) setupTestDatabase() {
	cfg := s.cfg.Postgres
	s.Require().NotEmpty(cfg.AdminRole.Name, "postgres migrations admin config empty")

	opts := fmt.Sprintf("sslmode=%s&search_path=%s", cfg.SSLMode, postgres.SchemaName)
	pgxConf := pgtestdb.Config{
		DriverName: "pgx",
		Host:       cfg.Host,
		Port:       fmt.Sprintf("%d", cfg.Port),
		Database:   cfg.Database,
		Options:    opts,
		User:       cfg.AdminRole.Name,
		Password:   cfg.AdminRole.Password,
		TestRole: &pgtestdb.Role{
			Username: cfg.AppRole.Name,
			Password: cfg.AppRole.Password,
		},
	}
	s.db = postgres.NewStdDatabaseClient(pgtestdb.New(s.T(), pgxConf, newTestDbMigrator()))
}

type testDbMigrator struct {
	gm *golangmigrator.GolangMigrator
}

func newTestDbMigrator() *testDbMigrator {
	return &testDbMigrator{
		gm: golangmigrator.New(migrations.EmbedFSDir, golangmigrator.WithFS(migrations.FS)),
	}
}

func (m *testDbMigrator) Hash() (string, error) {
	return m.gm.Hash()
}

func (m *testDbMigrator) Migrate(ctx context.Context, db *sql.DB, config pgtestdb.Config) error {
	var setupDbQueryTemplate = `
		CREATE SCHEMA IF NOT EXISTS %[1]s;
		GRANT USAGE ON SCHEMA %[1]s TO %[2]s;
		ALTER DEFAULT PRIVILEGES IN SCHEMA %[1]s GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO %[2]s;
		ALTER DEFAULT PRIVILEGES IN SCHEMA %[1]s GRANT USAGE, SELECT ON SEQUENCES TO %[2]s;
		ALTER ROLE %[2]s SET search_path TO %[1]s;`
	setupQuery := fmt.Sprintf(setupDbQueryTemplate, postgres.SchemaName, config.TestRole.Username)
	if _, setupErr := db.ExecContext(ctx, setupQuery); setupErr != nil {
		return fmt.Errorf("setup schema (query=[%s]): %w", setupQuery, setupErr)
	}
	return m.gm.Migrate(ctx, db, config)
}

func (s *Suite) closeTestDatabase() {
	if closeErr := s.db.Shutdown(); closeErr != nil {
		s.T().Logf("failed to close database client: %v", closeErr)
	}
}
