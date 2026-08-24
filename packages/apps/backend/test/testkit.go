package test

import (
	"cmp"
	"context"
	"os"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/stretchr/testify/suite"

	"github.com/rezible/rezible/internal/koanf"
	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/internal/postgres/pgtestdb"
	"github.com/rezible/rezible/pkg/execution"
)

type Suite struct {
	suite.Suite

	cfg  *rez.Config
	opts options
}

func NewSuite(optFns ...SuiteOption) Suite {
	opts := options{}
	for _, opt := range optFns {
		opt(&opts)
	}
	return Suite{opts: opts}
}

func (s *Suite) SetupSuite() {
	s.cfg = new(s.loadConfig())
}

func (s *Suite) TearDownSuite() {

}

func (s *Suite) Config() rez.Config {
	if s.cfg == nil {
		s.cfg = new(s.loadConfig())
	}
	return *s.cfg
}

func getEnvOr(key, fallback string) string {
	return cmp.Or(os.Getenv(key), fallback)
}

func (s *Suite) loadConfig() rez.Config {
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
	return cfg
}

func (s *Suite) SystemContext() context.Context {
	return execution.NewSystemContext(s.T().Context())
}

func (s *Suite) SeedTenantContext() context.Context {
	return execution.NewTenantContext(s.T().Context(), seedTenantId)
}

func (s *Suite) CreateTestDatabase() rez.Database {
	start := time.Now()
	cfg := s.Config().Postgres
	s.Require().NotEmpty(cfg.AdminRole.Name, "postgres migrations admin config empty")

	tdb, tdbErr := pgtestdb.New(cfg)
	s.Require().NoError(tdbErr)
	s.T().Cleanup(func() {
		s.Require().NoError(tdb.Shutdown())
	})

	pool, poolErr := postgres.MakePgxPool(s.T().Context(), tdb.Config(), false)
	s.Require().NoError(poolErr)

	db, dbErr := postgres.NewPgxPoolDatabaseClient(pool)
	s.Require().NoError(dbErr)

	s.T().Logf("created test database in %dms", time.Since(start).Milliseconds())

	s.SeedTestEntities(db)

	s.T().Cleanup(func() {
		if closeErr := db.Shutdown(); closeErr != nil {
			s.T().Logf("failed to close database client: %v", closeErr)
		}
	})

	return db
}
