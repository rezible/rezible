package testkit

import (
	"context"
	"fmt"
	"net/url"

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
	cfg, cfgErr := koanf.NewConfigLoader(koanf.ConfigLoaderOptions{
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
	pgConnCfg, connCfgErr := postgres.GetPgxConfig()
	s.Require().NoError(connCfgErr, "failed to get database config")

	connUrl, urlErr := url.Parse(pgConnCfg.ConnString())
	s.Require().NoError(urlErr, "failed to parse conn string")

	pgxConf := pgtestdb.Config{
		DriverName: "pgx",
		User:       pgConnCfg.User,
		Host:       pgConnCfg.Host,
		Password:   pgConnCfg.Password,
		Port:       fmt.Sprintf("%d", pgConnCfg.Port),
		Options:    connUrl.RawQuery,
	}
	mg := golangmigrator.New(".", golangmigrator.WithFS(migrations.FS))
	testDb := pgtestdb.New(s.T(), pgxConf, mg)
	s.dbClient = postgres.MakeClient(entsql.OpenDB("postgres", testDb))
}

func (s *Suite) closeTestDatabase() {
	if closeErr := s.dbClient.Close(); closeErr != nil {
		s.T().Logf("failed to close database client: %v", closeErr)
	}
}
