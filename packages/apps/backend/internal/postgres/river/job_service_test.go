package river_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"
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
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/rezible/rezible/test"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
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

type userJobArgs struct {
	Payload string `json:"payload"`
}

func (userJobArgs) Kind() string {
	return "test-user-job"
}

func (userJobArgs) WorkerExecutionContext() execution.ActorKind {
	return execution.KindUser
}

type systemJobArgs struct {
	Payload string `json:"payload"`
}

func (systemJobArgs) Kind() string {
	return "test-system-job"
}

func (systemJobArgs) WorkerExecutionContext() execution.ActorKind {
	return execution.KindSystem
}

// failingJobArgs is worked by failingWorker, which fails or panics on every attempt.
type failingJobArgs struct {
	Panic bool `json:"panic"`
}

func (failingJobArgs) Kind() string {
	return "test-failing-job"
}

type failingWorker struct {
	river.WorkerDefaults[failingJobArgs]
}

// NextRetry keeps a retried job from running again during the test.
func (failingWorker) NextRetry(*river.Job[failingJobArgs]) time.Time {
	return time.Now().Add(time.Hour)
}

func (failingWorker) Work(_ context.Context, job *river.Job[failingJobArgs]) error {
	if job.Args.Panic {
		panic("worker exploded")
	}
	return errors.New("worker failed")
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
	spans        *tracetest.SpanRecorder
	logs         *capturedLogs
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

	// The job service reads the global logger and tracer provider when it is built.
	s.spans = tracetest.NewSpanRecorder()
	previousProvider := otel.GetTracerProvider()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(s.spans)))
	s.T().Cleanup(func() { otel.SetTracerProvider(previousProvider) })
	s.logs = &capturedLogs{}
	previousLogger := slog.Default()
	slog.SetDefault(slog.New(s.logs))
	s.T().Cleanup(func() { slog.SetDefault(previousLogger) })

	service, serviceErr := jobriver.NewJobService(s.Config(), pool.Pool)
	s.Require().NoError(serviceErr)

	s.observations = make(chan jobObservation, 4)
	definition := jobs.Definition{Workers: []jobs.WorkerDefinition{
		jobs.DefineWorker(river.WorkFunc(func(ctx context.Context, job *river.Job[*tenantJobArgs]) error {
			return s.observe(ctx, job.ID)
		})),
		jobs.DefineWorker(river.WorkFunc(func(ctx context.Context, job *river.Job[userJobArgs]) error {
			return s.observe(ctx, job.ID)
		})),
		jobs.DefineWorker(river.WorkFunc(func(ctx context.Context, job *river.Job[systemJobArgs]) error {
			return s.observe(ctx, job.ID)
		})),
		jobs.DefineWorkerFunc(func(context.Context, jobs.ScanOncallShifts) error { return nil }),
		jobs.DefineWorker[failingJobArgs](failingWorker{}),
	}}
	s.Require().NoError(service.Register(definition))
	s.service = service
}

func (s *JobServiceSuite) observe(ctx context.Context, id int64) error {
	observation := jobObservation{
		ID:      id,
		Context: execution.GetContext(ctx),
	}
	select {
	case s.observations <- observation:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *JobServiceSuite) TestInsertionPreparesArgsWithAndWithoutTransaction() {
	for _, transactional := range []bool{false, true} {
		s.Run(map[bool]string{false: "direct", true: "transaction"}[transactional], func() {
			ctx := execution.NewTenantContext(s.T().Context(), 101)
			payload := s.T().Name()
			var inserted []*rivertype.JobInsertResult
			insert := func(ctx context.Context, _ *ent.Client) error {
				args := &tenantJobArgs{Payload: payload}
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
			s.Require().ErrorIs(missingErr, errs.ErrTenantContextMissing)
			params := []river.InsertManyParams{
				{Args: jobs.ScanOncallShifts{}},
				{Args: &tenantJobArgs{Payload: "invalid"}},
			}
			_, batchErr := s.service.InsertMany(execution.NewSystemContext(ctx), params)
			s.Require().ErrorContains(batchErr, "batch index 1")
			s.Require().ErrorIs(batchErr, errs.ErrTenantContextMissing)
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
	_, insertErr := s.service.Insert(execution.NewTenantContext(s.T().Context(), 101), jobs.ScanOncallShifts{}, nil)
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
				s.Equal(execution.GetContext(execution.NewTenantContext(ctx, identity.Session.TenantID)), observed)
			}
		})
	}
}

func (s *JobServiceSuite) TestInsertRejectsIdentityNotHeldByCaller() {
	userCtx, identity := s.NewIdentity(s.db, "Alice")
	tenantCtx := execution.NewTenantContext(s.T().Context(), identity.Session.TenantID)
	systemCtx := execution.NewSystemContext(s.T().Context())
	otherTenantArgs := &tenantJobArgs{Payload: "other tenant"}
	otherTenantArgs.TenantID = identity.Session.TenantID + 1

	cases := []struct {
		name    string
		ctx     context.Context
		args    river.JobArgs
		wantErr error
	}{
		{name: "default/system", ctx: systemCtx, args: jobs.ScanOncallShifts{}, wantErr: errs.ErrTenantContextMissing},
		{name: "user/tenant", ctx: tenantCtx, args: userJobArgs{}, wantErr: errs.ErrForbidden},
		{name: "system/user", ctx: userCtx, args: systemJobArgs{}, wantErr: errs.ErrForbidden},
		{name: "tenant args/other tenant", ctx: userCtx, args: otherTenantArgs, wantErr: errs.ErrForbidden},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			_, insertErr := s.service.Insert(tc.ctx, tc.args, nil)
			s.Require().ErrorIs(insertErr, tc.wantErr)
		})
	}

	var count int
	queryErr := s.pool.QueryRow(s.T().Context(), "SELECT count(*) FROM river.river_job").Scan(&count)
	s.Require().NoError(queryErr)
	s.Zero(count)
}

func (s *JobServiceSuite) TestWorkerExecutionContext() {
	firstCtx, first := s.NewIdentity(s.db, "Alice")
	secondCtx, second := s.NewIdentity(s.db, "Bob")
	expected := make(map[int64]execution.Context)
	insert := func(ctx context.Context, args river.JobArgs) int64 {
		inserted, insertErr := s.service.Insert(ctx, args, nil)
		s.Require().NoError(insertErr)
		s.Require().False(inserted.UniqueSkippedAsDuplicate)
		return inserted.Job.ID
	}
	userCtx := func(ctx context.Context, identity test.Identity) context.Context {
		ctx = execution.NewRootContext(ctx, execution.KindAnonymous, execution.SourceHTTP)
		return execution.NewUserContext(ctx, identity.Session)
	}

	// Default jobs run as the inserting user's tenant, without the user.
	for _, identity := range []struct {
		ctx      context.Context
		identity test.Identity
	}{{firstCtx, first}, {secondCtx, second}} {
		ctx := userCtx(identity.ctx, identity.identity)
		id := insert(ctx, &tenantJobArgs{Payload: "same payload"})
		expected[id] = execution.GetContext(execution.NewTenantContext(ctx, identity.identity.Session.TenantID))
	}

	// User jobs run as the inserting user.
	firstUserCtx := userCtx(firstCtx, first)
	userJobID := insert(firstUserCtx, userJobArgs{Payload: "user"})
	expected[userJobID] = execution.GetContext(firstUserCtx)

	// System jobs run under the job client's system root context.
	systemJobID := insert(execution.NewSystemContext(s.T().Context()), systemJobArgs{Payload: "system"})

	// Run the real worker and compare its context with the expected identity.
	ctx := s.runWorkers()

	seen := mapset.NewSet[int64]()
	for range len(expected) + 1 {
		select {
		case result := <-s.observations:
			s.False(seen.Contains(result.ID), "duplicate worker observation")
			seen.Add(result.ID)
			if result.ID == systemJobID {
				s.True(result.Context.IsSystem())
				_, hasTenant := result.Context.TenantID()
				s.False(hasTenant)
				continue
			}
			want, found := expected[result.ID]
			s.Require().True(found)
			s.Equal(want, result.Context)
		case <-ctx.Done():
			s.FailNow("worker observations timed out")
		}
	}

	for _, id := range seen.ToSlice() {
		s.Require().Eventually(func() bool {
			var state string
			queryErr := s.pool.QueryRow(ctx, "SELECT state FROM river.river_job WHERE id = $1", id).Scan(&state)
			return queryErr == nil && state == "completed"
		}, 5*time.Second, 10*time.Millisecond, "job %d did not complete", id)
	}
}

// runWorkers works jobs until the test ends.
func (s *JobServiceSuite) runWorkers() context.Context {
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
	return ctx
}

func (s *JobServiceSuite) readMetadata(ctx context.Context, jobID int64) map[string]any {
	var encoded []byte
	queryErr := s.pool.QueryRow(ctx, "SELECT metadata FROM river.river_job WHERE id = $1", jobID).Scan(&encoded)
	s.Require().NoError(queryErr)
	var metadata map[string]any
	s.Require().NoError(json.Unmarshal(encoded, &metadata))
	return metadata
}

func (s *JobServiceSuite) TestInsertKeepsOtherMetadata() {
	ctx := execution.NewTenantContext(s.T().Context(), 101)
	ctx, span := otel.Tracer("test").Start(ctx, "request")
	defer span.End()

	opts := &river.InsertOpts{Metadata: []byte(`{"source":"test"}`)}
	inserted, insertErr := s.service.Insert(ctx, &tenantJobArgs{Payload: s.T().Name()}, opts)
	s.Require().NoError(insertErr)

	metadata := s.readMetadata(ctx, inserted.Job.ID)
	s.Equal("test", metadata["source"])
	s.Contains(metadata["traceparent"], span.SpanContext().TraceID().String())
	encodedExec, isText := metadata["ec"].(string)
	s.Require().True(isText)
	decodedExec, decodeErr := base64.StdEncoding.DecodeString(encodedExec)
	s.Require().NoError(decodeErr)
	exec, execErr := execution.DecodeContext(decodedExec)
	s.Require().NoError(execErr)
	s.Equal(execution.GetContext(ctx), exec)
}

func (s *JobServiceSuite) TestWorkSpanLinksToInsertingTrace() {
	ctx := execution.NewTenantContext(s.T().Context(), 101)
	ctx, span := otel.Tracer("test").Start(ctx, "request")
	inserted, insertErr := s.service.Insert(ctx, &tenantJobArgs{Payload: s.T().Name()}, nil)
	s.Require().NoError(insertErr)
	span.End()

	s.runWorkers()
	var observed jobObservation
	s.Require().Eventually(func() bool {
		select {
		case observed = <-s.observations:
			return true
		default:
			return false
		}
	}, 5*time.Second, 10*time.Millisecond)
	s.Require().Equal(inserted.Job.ID, observed.ID)

	endedSpan := func(name string) sdktrace.ReadOnlySpan {
		var found sdktrace.ReadOnlySpan
		s.Require().Eventually(func() bool {
			for _, ended := range s.spans.Ended() {
				if ended.Name() == name {
					found = ended
					return true
				}
			}
			return false
		}, 5*time.Second, 10*time.Millisecond, "%s span not ended", name)
		return found
	}
	insert := endedSpan("river.insert_many")
	work := endedSpan("river.work/test-tenant-job")
	s.Equal(span.SpanContext().TraceID(), insert.SpanContext().TraceID())
	s.NotEqual(span.SpanContext().TraceID(), work.SpanContext().TraceID(), "the job starts its own trace")
	s.False(work.Parent().IsValid())
	s.Require().Len(work.Links(), 1)
	s.Equal(insert.SpanContext().TraceID(), work.Links()[0].SpanContext.TraceID())
	s.Equal(insert.SpanContext().SpanID(), work.Links()[0].SpanContext.SpanID())
	s.Contains(work.Attributes(), attribute.Int("tenant_id", 101))
	s.Contains(work.Attributes(), attribute.String("actor", string(execution.KindTenant)))
}

func (s *JobServiceSuite) TestFinalFailuresAndPanicsAreLoggedOnceAtError() {
	ctx := execution.NewTenantContext(s.T().Context(), 101)
	insert := func(args failingJobArgs, maxAttempts int) int64 {
		inserted, insertErr := s.service.Insert(ctx, args, &river.InsertOpts{MaxAttempts: maxAttempts})
		s.Require().NoError(insertErr)
		return inserted.Job.ID
	}
	final := insert(failingJobArgs{}, 1)
	retried := insert(failingJobArgs{}, 2)
	panicked := insert(failingJobArgs{Panic: true}, 1)

	s.runWorkers()
	for _, id := range []int64{final, retried, panicked} {
		s.Require().Eventually(func() bool {
			var attempts int
			queryErr := s.pool.QueryRow(ctx, "SELECT coalesce(array_length(errors, 1), 0) FROM river.river_job WHERE id = $1", id).Scan(&attempts)
			return queryErr == nil && attempts == 1
		}, 5*time.Second, 10*time.Millisecond, "job %d did not record its attempt", id)
	}

	errorsByJob := s.logs.errorsByJobID()
	s.Equal(map[int64]string{final: "job failed", panicked: "job panicked"}, errorsByJob)
}

// capturedLogs records log records written through the default logger.
type capturedLogs struct {
	mu      sync.Mutex
	records []slog.Record
}

func (c *capturedLogs) Enabled(context.Context, slog.Level) bool { return true }

func (c *capturedLogs) Handle(_ context.Context, record slog.Record) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.records = append(c.records, record)
	return nil
}

func (c *capturedLogs) WithAttrs([]slog.Attr) slog.Handler { return c }
func (c *capturedLogs) WithGroup(string) slog.Handler      { return c }

// errorsByJobID returns the message of each Error record naming a job, failing on a second record for one job.
func (c *capturedLogs) errorsByJobID() map[int64]string {
	c.mu.Lock()
	defer c.mu.Unlock()
	messages := make(map[int64]string)
	for _, record := range c.records {
		if record.Level != slog.LevelError {
			continue
		}
		record.Attrs(func(attr slog.Attr) bool {
			if attr.Key == "job_id" {
				id := attr.Value.Int64()
				if _, logged := messages[id]; logged {
					messages[id] = "logged more than once"
				} else {
					messages[id] = record.Message
				}
			}
			return true
		})
	}
	return messages
}
