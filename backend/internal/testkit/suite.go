package testkit

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	rez "github.com/rezible/rezible"
	"github.com/stretchr/testify/suite"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/internal/postgres"
)

type Option func(*options)

type options struct {
	allowTenantCreation bool
	allowUserCreation   bool
}

func WithAllowTenantCreation(allow bool) Option {
	return func(o *options) { o.allowTenantCreation = allow }
}

func WithAllowUserCreation(allow bool) Option {
	return func(o *options) { o.allowUserCreation = allow }
}

type Suite struct {
	suite.Suite

	opts options

	dbUrl      string
	testDbName string
	testDbUrl  string
	dbc        *postgres.DatabaseClient
	client     *ent.Client
	context    context.Context
}

type BaseTenant struct {
	TenantID     int
	Tenant       *ent.Tenant
	Organization *ent.Organization
	User         *ent.User
	Context      context.Context
}

func NewSuite(opts ...Option) Suite {
	cfg := options{allowTenantCreation: true, allowUserCreation: true}
	for _, opt := range opts {
		opt(&cfg)
	}
	return Suite{opts: cfg}
}

func (s *Suite) SetupSuite() {
	s.dbUrl = os.Getenv("DB_URL")
	if s.dbUrl == "" {
		s.T().Fatal("TEST_DB_URL or DB_URL must be set for backend integration tests")
	}

	ctx := access.SystemContext(s.T().Context())

	s.testDbName = makeTestDatabaseName(s.T().Name())
	s.Require().NoError(createDatabase(ctx, s.dbUrl, s.testDbName))

	testDatabaseURL, err := replaceDatabaseName(s.dbUrl, s.testDbName)
	s.Require().NoError(err)
	s.testDbUrl = testDatabaseURL

	s.Require().NoError(migrateSchema(ctx, testDatabaseURL))

	s.dbc, err = postgres.NewDatabaseClient(ctx)
	s.Require().NoError(err)
	s.client = s.dbc.Client()

	rez.Config = &ConfigLoader{
		DatabaseURL:             testDatabaseURL,
		AllowTenantCreationFlag: s.opts.allowTenantCreation,
		AllowUserCreationFlag:   s.opts.allowUserCreation,
		Values:                  map[string]string{},
		Bools:                   map[string]bool{},
		Durations:               map[string]time.Duration{},
	}

}

func (s *Suite) TearDownSuite() {
	if s.dbc != nil {
		_ = s.dbc.Close()
	}
	if s.dbUrl != "" && s.testDbName != "" {
		s.Require().NoError(dropDatabase(context.Background(), s.dbUrl, s.testDbName))
	}
}

func (s *Suite) Context() context.Context { return access.AnonymousContext(s.T().Context()) }
func (s *Suite) Client() *ent.Client      { return s.client }

func (s *Suite) SeedBaseTenant() BaseTenant {
	sysCtx := access.SystemContext(s.Context())
	tenant, err := s.client.Tenant.Create().Save(sysCtx)
	s.Require().NoError(err)

	tenantCtx := access.TenantContext(sysCtx, tenant.ID)
	org, err := s.client.Organization.Create().
		SetName("Test Organization").
		SetExternalID("org-" + uuid.NewString()).
		Save(tenantCtx)
	s.Require().NoError(err)

	usr, err := s.client.User.Create().
		SetEmail("owner+" + uuid.NewString() + "@example.com").
		SetName("Owner").
		Save(tenantCtx)
	s.Require().NoError(err)

	return BaseTenant{TenantID: tenant.ID, Tenant: tenant, Organization: org, User: usr}
}

func makeTestDatabaseName(testName string) string {
	name := strings.ToLower(testName)
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.Trim(name, "_")
	if name == "" {
		name = "suite"
	}
	if len(name) > 20 {
		name = name[:20]
	}
	return fmt.Sprintf("rez_t_%s_%s", name, strings.ReplaceAll(uuid.NewString()[:8], "-", ""))
}

func replaceDatabaseName(databaseURL string, databaseName string) (string, error) {
	u, err := url.Parse(databaseURL)
	if err != nil {
		return "", fmt.Errorf("parse database URL: %w", err)
	}
	u.Path = "/" + databaseName
	return u.String(), nil
}

func createDatabase(ctx context.Context, adminURL, databaseName string) error {
	conn, err := pgx.Connect(ctx, adminURL)
	if err != nil {
		return fmt.Errorf("connect admin db: %w", err)
	}
	defer conn.Close(ctx)
	_, err = conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", pgx.Identifier{databaseName}.Sanitize()))
	if err != nil {
		return fmt.Errorf("exec create database: %w", err)
	}
	return nil
}

func dropDatabase(ctx context.Context, adminURL, databaseName string) error {
	conn, err := pgx.Connect(ctx, adminURL)
	if err != nil {
		return fmt.Errorf("connect admin db: %w", err)
	}
	defer conn.Close(ctx)
	query := fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", pgx.Identifier{databaseName}.Sanitize())
	if _, err := conn.Exec(ctx, query); err == nil {
		return nil
	}
	fallback := fmt.Sprintf("DROP DATABASE IF EXISTS %s", pgx.Identifier{databaseName}.Sanitize())
	if _, err := conn.Exec(ctx, fallback); err != nil {
		return fmt.Errorf("exec drop database: %w", err)
	}
	return nil
}

func migrateSchema(ctx context.Context, databaseURL string) error {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("open migration database: %w", err)
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping migration database: %w", err)
	}
	driver := entsql.OpenDB(dialect.Postgres, db)
	client := ent.NewClient(ent.Driver(driver))
	defer client.Close()
	if err := client.Schema.Create(ctx); err != nil {
		return fmt.Errorf("create schema: %w", err)
	}
	return nil
}
