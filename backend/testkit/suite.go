package testkit

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/internal/viper"
	"github.com/stretchr/testify/suite"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/internal/postgres"
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

	dbConnUrl  string
	testDbName string
	testDbUrl  string
	testDb     *postgres.DatabaseClient

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
	s.dbConnUrl = rez.Config.DatabaseUrl()
	s.Require().NotEmpty(s.dbConnUrl, "database url is empty")

	s.setupTestDatabase()
	s.seedData()
}

func (s *Suite) SetConfigOverrides(overrides map[string]any) {
	rez.Config = viper.NewConfigLoader(viper.ConfigLoaderOptions{
		LoadEnvironment: true,
		Overrides:       overrides,
	})
}

func (s *Suite) TearDownSuite() {
	if closeErr := s.testDb.Close(); closeErr != nil {
		s.T().Logf("failed to close database client: %v", closeErr)
	}
	if dropErr := s.dropTestDatabase(); dropErr != nil {
		s.T().Logf("failed to drop test database: %v", dropErr)
	}
}

func (s *Suite) Client() *ent.Client { return s.testDb.Client() }

func (s *Suite) SystemContext() context.Context {
	return access.SystemContext(s.T().Context())
}

func (s *Suite) SeedTenantContext() context.Context {
	return access.TenantContext(s.T().Context(), s.SeedTenant.ID)
}

func (s *Suite) AnonymousContext() context.Context {
	return access.AnonymousContext(s.T().Context())
}

var testDbNameRe = regexp.MustCompile(`[/\- ]`)

func (s *Suite) sanitizedTestDbName() string {
	return pgx.Identifier{s.testDbName}.Sanitize()
}

func makeTestDatabaseName(testName string) string {
	name := strings.Trim(testDbNameRe.ReplaceAllString(strings.ToLower(testName), "_"), "_")
	if name == "" {
		name = "suite"
	} else if len(name) > 6 {
		name = name[:4]
	}
	// TODO: generate a hash from the name??
	suffix := testDbNameRe.ReplaceAllString(uuid.NewString(), "")
	return fmt.Sprintf("reztest_%s_%s", name, suffix)
}

func (s *Suite) withPostgresClient(ctx context.Context, fn func(pc *postgres.DatabaseClient) error) error {
	pc, pcErr := postgres.NewDatabasePoolClient(ctx, s.dbConnUrl)
	if pcErr != nil {
		return fmt.Errorf("could not create postgres client: %w", pcErr)
	}
	defer func(c *postgres.DatabaseClient) {
		if closeErr := c.Close(); closeErr != nil {
			s.T().Logf("failed to close database client: %v", closeErr)
		}
	}(pc)

	return fn(pc)
}

func (s *Suite) setupTestDatabase() {
	ctx := s.T().Context()

	dbUrl, dbUrlParseErr := url.Parse(s.dbConnUrl)
	s.Require().NoError(dbUrlParseErr, "failed to parse database url")

	s.testDbName = makeTestDatabaseName(s.T().Name())
	dbUrl.Path = "/" + s.testDbName
	s.testDbUrl = dbUrl.String()

	createTestDbFn := func(pc *postgres.DatabaseClient) error {
		_, createErr := pc.Pool().Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", s.sanitizedTestDbName()))
		if createErr != nil {
			return fmt.Errorf("failed to create test database: %w", createErr)
		}
		return nil
	}
	s.Require().NoError(s.withPostgresClient(ctx, createTestDbFn))
	pc, pcErr := postgres.NewDatabasePoolClient(ctx, s.testDbUrl)
	s.Require().NoError(pcErr, "failed to connect to postgres database")
	s.Require().NoError(pc.RunAutoMigrations(ctx), "failed to apply test database migrations")
	s.testDb = pc
}

func (s *Suite) dropTestDatabase() error {
	ctx := s.T().Context()
	dropFn := func(pc *postgres.DatabaseClient) error {
		pool := pc.Pool()
		dropQuery := fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", s.sanitizedTestDbName())
		if _, dropErr := pool.Exec(ctx, dropQuery); dropErr == nil {
			return nil
		}
		s.T().Logf("error dropping test database: %s", s.testDbName)

		fallbackDropQuery := fmt.Sprintf("DROP DATABASE IF EXISTS %s", s.sanitizedTestDbName())
		if _, fallbackErr := pool.Exec(ctx, fallbackDropQuery); fallbackErr != nil {
			return fmt.Errorf("fallback exec drop database: %w", fallbackErr)
		}
		return nil
	}

	return s.withPostgresClient(ctx, dropFn)
}
