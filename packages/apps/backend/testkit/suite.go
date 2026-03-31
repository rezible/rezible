package testkit

import (
	"context"
	"fmt"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/rezible/rezible/migrations"
	"github.com/stretchr/testify/suite"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/peterldowns/pgtestdb"
	"github.com/peterldowns/pgtestdb/migrators/golangmigrator"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
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
	return access.SystemContext(s.T().Context())
}

func (s *Suite) SeedTenantContext() context.Context {
	return access.TenantContext(s.T().Context(), s.SeedTenant.ID)
}

func (s *Suite) GetAnonymousContext() context.Context {
	return access.AnonymousContext(s.T().Context())
}

func (s *Suite) setupTestDatabase() {
	pgCfg, pgCfgErr := postgres.LoadConfig()
	s.Require().NoError(pgCfgErr, "loading postgres config")
	s.Require().NotNil(pgCfg.Migrations, "migrations config nil")

	pgxConf := pgtestdb.Config{
		DriverName: "pgx",
		Host:       pgCfg.Host,
		Port:       fmt.Sprintf("%d", pgCfg.Port),
		Options:    "sslmode=" + pgCfg.SSLMode,
		User:       pgCfg.Migrations.User,
		Password:   pgCfg.Migrations.Password,
		TestRole: &pgtestdb.Role{
			Username: pgCfg.User,
			Password: pgCfg.Password,
		},
	}
	testDb := pgtestdb.New(s.T(), pgxConf, golangmigrator.New(migrations.Path, golangmigrator.WithFS(migrations.FS)))
	s.dbClient = postgres.MakeEntClient(entsql.OpenDB("postgres", testDb))
}

func (s *Suite) closeTestDatabase() {
	if closeErr := s.dbClient.Close(); closeErr != nil {
		s.T().Logf("failed to close database client: %v", closeErr)
	}
}
