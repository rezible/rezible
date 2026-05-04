package testkit

import (
	"context"
	"database/sql"
	"fmt"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/stretchr/testify/suite"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/peterldowns/pgtestdb"
	"github.com/peterldowns/pgtestdb/migrators/golangmigrator"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/execution"
	"github.com/rezible/rezible/internal/koanf"
	"github.com/rezible/rezible/internal/postgres"
)

type Option func(*options)

type options struct {
	skipSeedOrganization bool
	skipSeedUser         bool
}

type Suite struct {
	suite.Suite

	opts options

	dbClient *ent.Client

	SeedTenant       *ent.Tenant
	SeedOrganization *ent.Organization
	SeedUser         *ent.User
}

func NewSuite(opts ...Option) Suite {
	cfg := options{}
	for _, opt := range opts {
		opt(&cfg)
	}
	return Suite{opts: cfg}
}

func (s *Suite) SetupSuite() {
	s.SetConfigOverrides(nil)
	s.setupTestDatabase()
	s.SeedTestEntities()
}

func (s *Suite) TearDownSuite() {
	s.closeTestDatabase()
}

func (s *Suite) BeforeTest(suiteName, testName string) {
	s.SetConfigOverrides(nil)
}

func (s *Suite) SetConfigOverrides(overrides map[string]any) {
	cfg, cfgErr := koanf.NewConfigLoader(s.T().Context(), koanf.ConfigLoaderOptions{
		LoadEnvironment: true,
		Overrides:       overrides,
	})
	s.Require().NoError(cfgErr)
	rez.Config = cfg
}

func (s *Suite) Client() *ent.Client { return s.dbClient }

func (s *Suite) SystemContext() context.Context {
	return execution.SystemContext(s.T().Context())
}

func (s *Suite) SeedTenantContext() context.Context {
	return execution.SystemTenantContext(s.T().Context(), s.SeedTenant.ID)
}

func (s *Suite) setupTestDatabase() {
	pgCfg, pgCfgErr := postgres.LoadConfig()
	s.Require().NoError(pgCfgErr, "loading postgres config")
	s.Require().NotEmpty(pgCfg.AdminRole.Name, "migrations config nil")

	opts := fmt.Sprintf("sslmode=%s&search_path=%s", pgCfg.SSLMode, postgres.SchemaName)
	pgxConf := pgtestdb.Config{
		DriverName: "pgx",
		Host:       pgCfg.Host,
		Port:       fmt.Sprintf("%d", pgCfg.Port),
		Options:    opts,
		User:       pgCfg.AdminRole.Name,
		Password:   pgCfg.AdminRole.Password,
		TestRole: &pgtestdb.Role{
			Username: pgCfg.AppRole.Name,
			Password: pgCfg.AppRole.Password,
		},
	}
	testDb := pgtestdb.New(s.T(), pgxConf, newTestDbMigrator(pgCfg))
	s.dbClient = postgres.MakeEntClient(entsql.OpenDB("postgres", testDb))
}

type testDbMigrator struct {
	cfg postgres.Config
	gm  *golangmigrator.GolangMigrator
}

func newTestDbMigrator(cfg postgres.Config) *testDbMigrator {
	return &testDbMigrator{
		cfg: cfg,
		gm:  golangmigrator.New(postgres.MigrationsDir, golangmigrator.WithFS(postgres.MigrationsFS)),
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
		return fmt.Errorf("setup schema setupQuery: %s", setupErr)
	}
	return m.gm.Migrate(ctx, db, config)
}

func (s *Suite) closeTestDatabase() {
	if closeErr := s.dbClient.Close(); closeErr != nil {
		s.T().Logf("failed to close database client: %v", closeErr)
	}
}
