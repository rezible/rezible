package river_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/user"
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

func (*tenantJobArgs) Kind() string {
	return "test-tenant-job"
}

func (*tenantJobArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{UniqueOpts: river.UniqueOpts{ByArgs: true}}
}

type jobObservation struct {
	ID      int64
	Context execution.Context
}

type JobServiceSuite struct {
	test.Suite
	service      *jobriver.JobService
	db           rez.Database
	pool         *postgres.ConnectionPool
	observations chan jobObservation
}

func TestJobServiceSuite(t *testing.T) {
	suite.Run(t, &JobServiceSuite{Suite: test.NewSuite()})
}

func (s *JobServiceSuite) SetupTest() {
	testDB, dbErr := pgtestdb.New(s.Config().Postgres)
	s.Require().NoError(dbErr)
	s.T().Cleanup(func() {
		s.NoError(testDB.Shutdown())
	})
	pool, poolErr := postgres.MakePgxPool(s.T().Context(), testDB.Config(), false)
	s.Require().NoError(poolErr)
	s.T().Cleanup(func() {
		s.NoError(pool.Shutdown())
	})
	s.pool = pool
	database, databaseErr := postgres.NewPgxPoolDatabaseClient(pool)
	s.Require().NoError(databaseErr)
	s.T().Cleanup(func() {
		s.NoError(database.Shutdown())
	})
	s.db = database

	service, serviceErr := jobriver.NewJobService(s.Config(), pool.Pool, s.Telemetry())
	s.Require().NoError(serviceErr)

	s.observations = make(chan jobObservation, 2)
	definition := jobs.Definition{Workers: []jobs.WorkerDefinition{
		jobs.DefineWorker(river.WorkFunc(func(ctx context.Context, job *river.Job[*tenantJobArgs]) error {
			observation := jobObservation{
				ID:      job.ID,
				Context: execution.GetContext(ctx),
			}
			select {
			case s.observations <- observation:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})),
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

func (s *JobServiceSuite) TestDomainAndJobsTransaction() {
	ctx, identity := s.NewIdentity(s.db, "Transaction owner")
	cases := []struct {
		name     string
		batch    bool
		rollback bool
	}{
		{name: "single/commit"},
		{name: "single/rollback", rollback: true},
		{name: "batch/commit", batch: true},
		{name: "batch/rollback", batch: true, rollback: true},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			rowID := uuid.New()

			var inserted []*rivertype.JobInsertResult
			sentinel := errors.New("rollback after domain and job inserts")
			txErr := s.db.WithTx(ctx, func(txCtx context.Context, client *ent.Client) error {
				createUser := client.User.Create().
					SetID(rowID).
					SetName("transaction marker").
					SetEmail(rowID.String() + "@example.com")
				if createErr := createUser.Exec(txCtx); createErr != nil {
					return createErr
				}
				if tc.batch {
					var insertErr error
					params := []river.InsertManyParams{
						{Args: &tenantJobArgs{Payload: rowID.String()}},
						{Args: &tenantJobArgs{Payload: rowID.String() + "-second"}},
					}
					inserted, insertErr = s.service.InsertMany(txCtx, params)
					if insertErr != nil {
						return insertErr
					}
				} else {
					result, insertErr := s.service.Insert(txCtx, &tenantJobArgs{Payload: rowID.String()}, nil)
					if insertErr != nil {
						return insertErr
					}
					inserted = []*rivertype.JobInsertResult{result}
				}
				if tc.rollback {
					return sentinel
				}
				return nil
			})
			if tc.rollback {
				s.Require().ErrorIs(txErr, sentinel)
			} else {
				s.Require().NoError(txErr)
			}

			// Read outside the transaction to verify what actually committed.
			queryMarker := s.db.Client(ctx).User.Query().
				Where(user.ID(rowID))
			exists, queryErr := queryMarker.Exist(ctx)
			s.Require().NoError(queryErr)
			s.Equal(!tc.rollback, exists)
			s.Require().NotEmpty(inserted)

			for _, result := range inserted {
				var count int
				countErr := s.pool.QueryRow(ctx, "SELECT count(*) FROM river.river_job WHERE id = $1", result.Job.ID).Scan(&count)
				s.Require().NoError(countErr)
				if tc.rollback {
					s.Zero(count)
					continue
				}

				s.Equal(1, count)
				var args, metadata []byte
				queryJob := "SELECT args, metadata FROM river.river_job WHERE id = $1"
				readErr := s.pool.QueryRow(ctx, queryJob, result.Job.ID).Scan(&args, &metadata)
				s.Require().NoError(readErr)

				var decoded tenantJobArgs
				s.Require().NoError(json.Unmarshal(args, &decoded))
				s.Equal(identity.Session.TenantID, decoded.TenantID)

				var meta struct {
					Execution []byte `json:"ec"`
				}
				s.Require().NoError(json.Unmarshal(metadata, &meta))
				observed, decodeErr := execution.DecodeContext(meta.Execution)
				s.Require().NoError(decodeErr)
				s.Equal(execution.GetContext(ctx), observed)
			}
		})
	}
}

func (s *JobServiceSuite) TestWorkerPreservesExecutionContext() {
	firstCtx, first := s.NewIdentity(s.db, "Alice")
	secondCtx, second := s.NewIdentity(s.db, "Bob")
	expected := make(map[int64]execution.Context)
	enqueue := func(ctx context.Context, identity test.Identity) {
		ctx = execution.NewRootContext(ctx, execution.KindAnonymous, execution.SourceHTTP)
		ctx = execution.NewUserContext(ctx, identity.Session)
		inserted, insertErr := s.service.Insert(ctx, &tenantJobArgs{Payload: "same payload"}, nil)
		s.Require().NoError(insertErr)
		s.Require().False(inserted.UniqueSkippedAsDuplicate)
		expected[inserted.Job.ID] = execution.GetContext(ctx)
	}

	enqueue(firstCtx, first)
	enqueue(secondCtx, second)

	// Run the real worker and compare its context with the submitted context.
	ctx, cancel := context.WithTimeout(s.T().Context(), 10*time.Second)
	ready := make(chan struct{})
	finished := make(chan error, 1)
	go func() {
		finished <- s.service.Run(ctx, ready)
	}()
	s.T().Cleanup(func() {
		defer cancel()
		stopCtx, stop := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer stop()
		s.NoError(s.service.Shutdown(stopCtx))
		select {
		case runErr := <-finished:
			s.NoError(runErr)
		case <-stopCtx.Done():
			s.Error(stopCtx.Err(), "worker did not stop")
		}
	})
	select {
	case <-ready:
	case runErr := <-finished:
		finished <- runErr
		s.Require().NoError(runErr)
		s.FailNow("worker stopped before ready")
	case <-ctx.Done():
		s.FailNow("worker startup timed out")
	}

	seen := mapset.NewSet[int64]()
	for range 2 {
		select {
		case result := <-s.observations:
			want, found := expected[result.ID]
			s.Require().True(found)
			s.False(seen.Contains(result.ID), "duplicate worker observation")
			seen.Add(result.ID)
			s.Equal(want, result.Context)
		case <-ctx.Done():
			s.FailNow("worker observations timed out")
		}
	}

	for id := range expected {
		s.Require().Eventually(func() bool {
			var state string
			queryErr := s.pool.QueryRow(ctx, "SELECT state FROM river.river_job WHERE id = $1", id).Scan(&state)
			return queryErr == nil && state == "completed"
		}, 5*time.Second, 10*time.Millisecond, "job %d did not complete", id)
	}
}
