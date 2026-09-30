package test

import (
	"context"
	"time"

	rez "github.com/rezible/rezible"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/internal/postgres/pgtestdb"
	"github.com/rezible/rezible/pkg/execution"
)

type Suite struct {
	suite.Suite

	cfg       *rez.Config
	opts      options
	telemetry *Telemetry
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

func (s *Suite) Config() rez.Config {
	if s.cfg == nil {
		s.cfg = new(s.loadConfig())
	}
	return *s.cfg
}

func (s *Suite) loadConfig() rez.Config {
	s.T().Helper()
	cfg, configErr := loadConfig(s.T().Context(), s.opts.configOverrides)
	s.Require().NoError(configErr)
	return cfg
}

func (s *Suite) Telemetry() *Telemetry {
	if s.telemetry == nil {
		s.telemetry = NewTelemetry()
	}
	return s.telemetry
}

func (s *Suite) SystemContext() context.Context {
	return execution.NewSystemContext(s.T().Context())
}

func (s *Suite) SeedTenantContext() context.Context {
	return execution.NewTenantContext(s.T().Context(), seedTenantId)
}

func (s *Suite) CreateTestDatabase() rez.Database {
	t := s.T()
	t.Helper()
	start := time.Now()
	cfg := s.Config().Postgres
	s.Require().NotEmpty(cfg.AdminRole.Name, "postgres migrations admin config empty")

	tdb, tdbErr := pgtestdb.New(cfg)
	s.Require().NoError(tdbErr)
	t.Cleanup(func() {
		assert.NoError(t, tdb.Shutdown())
	})

	pool, poolErr := postgres.MakePgxPool(t.Context(), tdb.Config(), false)
	s.Require().NoError(poolErr)
	t.Cleanup(func() {
		assert.NoError(t, pool.Shutdown())
	})

	db, dbErr := postgres.NewPgxPoolDatabaseClient(pool)
	s.Require().NoError(dbErr)

	t.Cleanup(func() {
		assert.NoError(t, db.Shutdown())
	})
	t.Logf("created test database in %dms", time.Since(start).Milliseconds())
	s.seedTestEntities(db)
	return db
}
