package river_test

import (
	"context"
	"encoding/json"
	"testing"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/internal/opentelemetry"
	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/internal/postgres/pgtestdb"
	jobriver "github.com/rezible/rezible/internal/postgres/river"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/rezible/rezible/test"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/stretchr/testify/suite"
)

type tenantJobArgs struct {
	jobs.TenantArgs
	Payload string `json:"payload" river:"unique"`
}

func (*tenantJobArgs) Kind() string { return "test-tenant-job" }
func (*tenantJobArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{UniqueOpts: river.UniqueOpts{ByArgs: true}}
}

type JobServiceSuite struct {
	test.Suite
	service *jobriver.JobService
	db      rez.Database
	pool    *postgres.ConnectionPool
}

func TestJobServiceSuite(t *testing.T) {
	suite.Run(t, &JobServiceSuite{Suite: test.NewSuite()})
}

func (s *JobServiceSuite) SetupTest() {
	testDB, dbErr := pgtestdb.New(s.Config().Postgres)
	s.Require().NoError(dbErr)
	s.T().Cleanup(func() { s.Require().NoError(testDB.Shutdown()) })
	pool, poolErr := postgres.MakePgxPool(s.T().Context(), testDB.Config(), false)
	s.Require().NoError(poolErr)
	s.T().Cleanup(func() { s.Require().NoError(pool.Shutdown()) })
	s.pool = pool
	database, databaseErr := postgres.NewPgxPoolDatabaseClient(pool)
	s.Require().NoError(databaseErr)
	s.T().Cleanup(func() { s.Require().NoError(database.Shutdown()) })
	s.db = database

	tel, telemetryErr := opentelemetry.NewOpenTelemetryService(s.T().Context(), rez.Config{})
	s.Require().NoError(telemetryErr)
	s.T().Cleanup(func() { s.Require().NoError(tel.Shutdown(context.WithoutCancel(s.T().Context()))) })
	service, serviceErr := jobriver.NewJobService(s.Config(), pool.Pool, tel)
	s.Require().NoError(serviceErr)
	definition := jobs.Definition{Workers: []jobs.WorkerDefinition{
		jobs.DefineWorkerFunc(func(context.Context, *tenantJobArgs) error { return nil }),
		jobs.DefineWorkerFunc(func(context.Context, jobs.ScanOncallShifts) error { return nil }),
	}}
	s.Require().NoError(service.Register(definition))
	s.service = service
}

func (s *JobServiceSuite) TestInsertionPreparesArgsWithAndWithoutTransaction() {
	for _, transactional := range []bool{false, true} {
		s.Run(map[bool]string{false: "direct", true: "transaction"}[transactional], func() {
			ctx := execution.NewTenantContext(s.T().Context(), 101)
			payload := s.T().Name()
			var inserted []*rivertype.JobInsertResult
			insert := func(ctx context.Context, _ *ent.Client) error {
				args := &tenantJobArgs{Payload: payload}
				args.TenantID = 999
				one, insertErr := s.service.Insert(ctx, args, nil)
				if insertErr != nil {
					return insertErr
				}
				params := []river.InsertManyParams{
					{Args: &tenantJobArgs{Payload: payload}},
					{Args: &tenantJobArgs{Payload: payload + "-second"}},
				}
				many, batchErr := s.service.InsertMany(ctx, params)
				if batchErr != nil {
					return batchErr
				}
				s.False(one.UniqueSkippedAsDuplicate)
				s.True(many[0].UniqueSkippedAsDuplicate)
				s.False(many[1].UniqueSkippedAsDuplicate)
				inserted = append(inserted, one, many[1])
				return nil
			}
			if transactional {
				s.Require().NoError(s.db.WithTx(ctx, insert))
			} else {
				s.Require().NoError(insert(ctx, nil))
			}
			for i, result := range inserted {
				var encoded []byte
				queryErr := s.pool.QueryRow(ctx, "SELECT args FROM river.river_job WHERE id = $1", result.Job.ID).Scan(&encoded)
				s.Require().NoError(queryErr)
				var decoded tenantJobArgs
				s.Require().NoError(json.Unmarshal(encoded, &decoded))
				s.Equal(101, decoded.TenantID)
				s.Equal([]string{payload, payload + "-second"}[i], decoded.Payload)
			}
			otherCtx := execution.NewTenantContext(s.T().Context(), 202)
			other, otherErr := s.service.InsertMany(otherCtx, []river.InsertManyParams{{Args: &tenantJobArgs{Payload: payload}}})
			s.Require().NoError(otherErr)
			s.False(other[0].UniqueSkippedAsDuplicate)
		})
	}
}

func (s *JobServiceSuite) TestPreparationFailureInsertsNoJobs() {
	for _, transactional := range []bool{false, true} {
		ctx := execution.NewTenantContext(s.T().Context(), 101)
		attempt := func(ctx context.Context, _ *ent.Client) error {
			_, missingErr := s.service.Insert(execution.NewSystemContext(ctx), &tenantJobArgs{}, nil)
			s.Require().ErrorIs(missingErr, rez.ErrTenantContextMissing)
			params := []river.InsertManyParams{
				{Args: jobs.ScanOncallShifts{}},
				{Args: &tenantJobArgs{Payload: "invalid"}},
			}
			_, batchErr := s.service.InsertMany(execution.NewSystemContext(ctx), params)
			s.Require().ErrorContains(batchErr, "batch index 1")
			s.Require().ErrorIs(batchErr, rez.ErrTenantContextMissing)
			return nil
		}
		if transactional {
			s.Require().NoError(s.db.WithTx(ctx, attempt))
		} else {
			s.Require().NoError(attempt(ctx, nil))
		}
	}
	var count int
	queryErr := s.pool.QueryRow(s.T().Context(), "SELECT count(*) FROM river.river_job").Scan(&count)
	s.Require().NoError(queryErr)
	s.Zero(count)
	_, insertErr := s.service.Insert(s.SystemContext(), jobs.ScanOncallShifts{}, nil)
	s.Require().NoError(insertErr)
}
