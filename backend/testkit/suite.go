package testkit

import (
	"context"
	"net/url"

	"github.com/rezible/rezible/migrations"
	"github.com/stretchr/testify/suite"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/peterldowns/pgtestdb"
	"github.com/peterldowns/pgtestdb/migrators/golangmigrator"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/internal/viper"
)

type Option func(*options)

type options struct {
	seedTenant       bool
	seedOrganization bool
	seedUser         bool
}

func WithSeedTenant(seed bool) Option {
	return func(o *options) { o.seedTenant = seed }
}

func WithSeedTestOrganization(seed bool) Option {
	return func(o *options) { o.seedOrganization = seed }
}

type Suite struct {
	suite.Suite

	opts options

	dbClient *ent.Client

	SeedTenant       *ent.Tenant
	SeedOrganization *ent.Organization
}

func NewSuite(opts ...Option) Suite {
	cfg := options{seedTenant: true, seedOrganization: true, seedUser: true}
	for _, opt := range opts {
		opt(&cfg)
	}
	return Suite{opts: cfg}
}

func (s *Suite) SetupSuite() {
	rez.Config = viper.NewConfigLoader(viper.ConfigLoaderOptions{
		LoadEnvironment: true,
	})
	s.setupTestDatabase()
	s.seedTestEntities()
}

func (s *Suite) SetConfigOverrides(overrides map[string]any) {
	rez.Config = viper.NewConfigLoader(viper.ConfigLoaderOptions{
		LoadEnvironment: true,
		Overrides:       overrides,
	})
}

func (s *Suite) TearDownSuite() {
	if closeErr := s.dbClient.Close(); closeErr != nil {
		s.T().Logf("failed to close database client: %v", closeErr)
	}
}

func (s *Suite) Client() *ent.Client { return s.dbClient }

func (s *Suite) GetSystemContext() context.Context {
	return access.SystemContext(s.T().Context())
}

func (s *Suite) GetSeedTenantContext() context.Context {
	return access.TenantContext(s.T().Context(), s.SeedTenant.ID)
}

func (s *Suite) GetAnonymousContext() context.Context {
	return access.AnonymousContext(s.T().Context())
}

func (s *Suite) setupTestDatabase() {
	dbConnUrl := rez.Config.DatabaseUrl()
	s.Require().NotEmpty(dbConnUrl, "database url is empty")

	dbUrl, dbUrlParseErr := url.Parse(dbConnUrl)
	s.Require().NoError(dbUrlParseErr, "failed to parse database url")

	pgxConf := pgtestdb.Config{
		DriverName: "pgx",
		User:       dbUrl.User.Username(),
		Host:       dbUrl.Hostname(),
		Port:       dbUrl.Port(),
		Options:    dbUrl.RawQuery,
	}
	if pw, exists := dbUrl.User.Password(); exists {
		pgxConf.Password = pw
	}
	mg := golangmigrator.New("migrations", golangmigrator.WithFS(migrations.FS))
	tdb := pgtestdb.New(s.T(), pgxConf, mg)
	s.dbClient = postgres.ClientFromSql(tdb)
}
