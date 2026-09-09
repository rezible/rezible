package pgtestdb_test

import (
	"testing"

	"github.com/rezible/rezible/internal/postgres"
	postgrespgtestdb "github.com/rezible/rezible/internal/postgres/pgtestdb"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/test"
	"github.com/stretchr/testify/suite"
)

type DatabaseSuite struct {
	test.Suite
}

func TestDatabaseSuite(t *testing.T) {
	suite.Run(t, &DatabaseSuite{Suite: test.NewSuite()})
}

func (s *DatabaseSuite) TestCreatesAndRemovesMigratedDatabase() {
	database, databaseErr := postgrespgtestdb.New(s.Config().Postgres)
	s.Require().NoError(databaseErr)
	s.T().Cleanup(func() {
		s.Require().NoError(database.Shutdown())
	})

	pool, poolErr := postgres.MakePgxPool(s.T().Context(), database.Config(), false)
	s.Require().NoError(poolErr)
	client, clientErr := postgres.NewPgxPoolDatabaseClient(pool)
	s.Require().NoError(clientErr)

	ctx := execution.NewSystemContext(s.T().Context())
	s.Require().NoError(client.Client(ctx).Tenant.Create().Exec(ctx))
	databaseName := database.Config().Database
	s.Require().NoError(client.Shutdown())
	s.Require().NoError(pool.Shutdown())
	s.Require().NoError(database.Shutdown())

	adminPool, adminPoolErr := postgres.MakePgxPool(s.T().Context(), s.Config().Postgres, true)
	s.Require().NoError(adminPoolErr)
	defer adminPool.Shutdown()
	var exists bool
	queryErr := adminPool.QueryRow(s.T().Context(), "SELECT EXISTS (SELECT FROM pg_database WHERE datname = $1)", databaseName).Scan(&exists)
	s.Require().NoError(queryErr)
	s.False(exists)
}
