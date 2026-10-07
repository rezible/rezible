package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertype"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	at "github.com/rezible/rezible/ent/agentturn"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	nep "github.com/rezible/rezible/ent/normalizedeventprojection"
	"github.com/rezible/rezible/internal/genkit"
	"github.com/rezible/rezible/internal/http"
	"github.com/rezible/rezible/internal/integrations/alertmanager"
	demo "github.com/rezible/rezible/internal/integrations/demo"
	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/pkg/jobs"
	oapi "github.com/rezible/rezible/pkg/openapi"
	oapiv1 "github.com/rezible/rezible/pkg/openapi/v1"
	"github.com/rezible/rezible/test"
)

// Waits return as soon as their condition holds; this only bounds failures.
const appWaitTimeout = 30 * time.Second

type BackendSuite struct {
	test.Suite
}

func TestBackendSuite(t *testing.T) {
	suite.Run(t, &BackendSuite{Suite: test.NewSuite()})
}

type appTestOptions struct {
	ModelAction ai.ModelActionFunc[any]
}

type appHarness struct {
	suite *BackendSuite
	app   *Application
	// clock is the application's clock. It starts at real time and only moves forward, so jobs scheduled
	// from it run when RunDueJobs makes them available rather than on their own.
	clock    *test.Clock
	database rez.Database
	jobs     *river.Client[pgx.Tx]
	api      *oapi.TestAPI
	cookie   rez.AppAuthSessionCookie
}

func (s *BackendSuite) newAppHarness(options appTestOptions) *appHarness {
	t := s.T()
	t.Helper()

	modelAction := options.ModelAction
	var unexpectedModelCalls atomic.Int64
	if modelAction == nil {
		// Workers record unexpected calls; cleanup asserts only after they have stopped.
		modelAction = func(context.Context, *ai.ModelRequest, any, ai.ModelStreamCallback) (*ai.ModelResponse, error) {
			unexpectedModelCalls.Add(1)
			return nil, fmt.Errorf("unexpected model call in application test")
		}
	}

	app := NewApplication()

	// The injector owns services and the database. The harness owns the lifecycle goroutine.
	var lifecycleDone <-chan error
	ctx, cancel := context.WithCancel(t.Context())

	t.Cleanup(func() {
		cancel()
		stopCtx, stop := context.WithTimeout(context.WithoutCancel(ctx), 35*time.Second)
		defer stop()
		if lifecycleDone != nil {
			select {
			case lifecycleErr := <-lifecycleDone:
				assert.NoError(t, lifecycleErr, "stop application lifecycle")
			case <-stopCtx.Done():
				t.Errorf("application lifecycle did not stop: %v", stopCtx.Err())
				return
			}
		}
		assert.NoError(t, app.Shutdown(stopCtx), "release application resources")
		if options.ModelAction == nil {
			assert.Zero(t, unexpectedModelCalls.Load(), "application test must not call the model without an explicit ModelAction")
		}
	})

	_, initErr := app.Init(ctx)
	s.Require().NoError(initErr, "initialize application")

	app.Override[rez.Config](func(do.Injector) (rez.Config, error) {
		return s.Config(), nil
	})

	app.Override[rez.TelemetryService](func(do.Injector) (rez.TelemetryService, error) {
		return s.Telemetry(), nil
	})

	app.Override[rez.PostgresConfig](providePostgresTestDatabaseConfig)

	testClock := test.NewClock(time.Now())
	app.Override[rez.Clock](func(do.Injector) (rez.Clock, error) {
		return testClock, nil
	})

	// Select the real demo processor without installing its unrelated sync jobs, and Alertmanager, which
	// has no sync jobs.
	app.Override[[]rez.IntegrationDefinition](func(i do.Injector) ([]rez.IntegrationDefinition, error) {
		integration, integrationErr := do.Invoke[*demo.Integration](i)
		if integrationErr != nil {
			return nil, fmt.Errorf("resolve demo integration: %w", integrationErr)
		}
		alertmanagerIntegration, alertmanagerErr := do.Invoke[*alertmanager.Integration](i)
		if alertmanagerErr != nil {
			return nil, fmt.Errorf("resolve alertmanager integration: %w", alertmanagerErr)
		}
		return []rez.IntegrationDefinition{integration, alertmanagerIntegration}, nil
	})

	baseOpts := do.MustInvoke[[]genkit.AiRuntimeOption](app.i)
	app.Override[[]genkit.AiRuntimeOption](func(i do.Injector) ([]genkit.AiRuntimeOption, error) {
		modelOptions := &ai.ModelOptions{
			Supports: &ai.ModelSupports{
				Tools:      true,
				Multiturn:  true,
				SystemRole: true,
			},
		}
		model := genkit.NewModelDefinition("test/application", modelOptions, modelAction)
		model.IsDefault = true
		return append(baseOpts, genkit.WithDefinedModel(model)), nil
	})

	app.Override[[]*jobs.PeriodicJob](func(i do.Injector) ([]*jobs.PeriodicJob, error) {
		return nil, nil
	})

	ready, done := app.startLifecycle(ctx)
	lifecycleDone = done

	startupCtx, stopStartup := context.WithTimeout(ctx, appWaitTimeout)
	defer stopStartup()

	select {
	case <-ready:
	case lifecycleErr := <-done:
		lifecycleDone = nil // The result has been consumed; cleanup must not wait again.
		t.Fatalf("application stopped before readiness: %v", lifecycleErr)
	case <-startupCtx.Done():
		t.Fatalf("application readiness timed out: %v", startupCtx.Err())
	}

	database, databaseErr := app.invoke[rez.Database]()
	s.Require().NoError(databaseErr, "resolve application database")

	pool, poolErr := app.invoke[*postgres.ConnectionPool]()
	s.Require().NoError(poolErr, "resolve application pool")

	// TODO: just invoke this?
	jobConfig := &river.Config{
		Schema: "river",
		Logger: app.mustInvoke[rez.TelemetryService]().Logger(),
	}
	jobReader, jobReaderErr := river.NewClient(riverpgxv5.New(pool.Pool), jobConfig)
	s.Require().NoError(jobReaderErr, "create read-only River client")

	cookie, cookieErr := app.invoke[rez.AppAuthSessionCookie]()
	s.Require().NoError(cookieErr, "resolve application authentication")

	api, apiErr := app.invoke[oapiv1.API]()
	s.Require().NoError(apiErr, "resolve application v1 API")

	server, serverErr := app.invoke[*http.Server]()
	s.Require().NoError(serverErr, "resolve application HTTP server")

	httpBasePath := app.mustInvoke[rez.Config]().HttpServer.BasePath
	testAPI := oapiv1.NewTestAPI(t, api).WithHandler(server.Handler(), httpBasePath+oapiv1.VersionPrefix)
	return &appHarness{
		suite:    s,
		app:      app,
		clock:    testClock,
		database: database,
		jobs:     jobReader,
		cookie:   cookie,
		api:      testAPI,
	}
}

func (ah *appHarness) Client(ctx context.Context) *ent.Client {
	return ah.database.Client(ctx)
}

func (ah *appHarness) NewIdentity(label string) (context.Context, test.Identity) {
	ah.suite.T().Helper()
	return ah.suite.NewIdentity(ah.database, label)
}

func (ah *appHarness) API(identity test.Identity) *oapi.TestAPI {
	t := ah.suite.T()
	t.Helper()
	recorder := httptest.NewRecorder()
	ah.cookie.Set(recorder, identity.Session)
	response := recorder.Result()
	defer response.Body.Close()
	return ah.api.WithCookies(response.Cookies()...)
}

func (ah *appHarness) await(ctx context.Context, label string, probe func(context.Context) (bool, string, error)) {
	t := ah.suite.T()
	t.Helper()
	waitCtx, cancel := context.WithTimeout(ctx, appWaitTimeout)
	defer cancel()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	diagnostic := "not yet probed"
	for {
		ready, nextDiagnostic, probeErr := probe(waitCtx)
		diagnostic = nextDiagnostic
		if waitCtx.Err() != nil {
			t.Fatalf("%s timed out: %s (%v)", label, diagnostic, waitCtx.Err())
		}
		ah.suite.Require().NoError(probeErr, "%s: %s", label, diagnostic)
		if ready {
			return
		}
		select {
		case <-ticker.C:
		case <-waitCtx.Done():
			t.Fatalf("%s timed out: %s", label, diagnostic)
		}
	}
}

func (ah *appHarness) AwaitJob(ctx context.Context, id int64) {
	ah.suite.T().Helper()
	ah.await(ctx, fmt.Sprintf("job %d", id), func(waitCtx context.Context) (bool, string, error) {
		job, queryErr := ah.jobs.JobGet(waitCtx, id)
		if queryErr != nil {
			return false, "read job", queryErr
		}
		diagnostic := fmt.Sprintf("state=%s attempt=%d errors=%v", job.State, job.Attempt, job.Errors)
		if job.State == rivertype.JobStateCancelled || job.State == rivertype.JobStateDiscarded {
			return false, diagnostic, fmt.Errorf("job reached terminal failure state")
		}
		return job.State == rivertype.JobStateCompleted, diagnostic, nil
	})
}

// RunDueJobs processes the jobs returned by one paginated listing: available or running jobs, plus
// scheduled jobs due by at. Jobs enqueued after that listing (including same-kind follow-ups) are not
// guaranteed to complete here; journeys must wait for their next stage explicitly.
func (ah *appHarness) RunDueJobs(ctx context.Context, kind string, at time.Time) {
	ah.suite.T().Helper()
	rows, listErr := ah.jobsOfKind(ctx, kind)
	ah.suite.Require().NoError(listErr, "list %s jobs", kind)
	for _, job := range rows {
		switch job.State {
		case rivertype.JobStateAvailable, rivertype.JobStateRunning:
		case rivertype.JobStateScheduled:
			if job.ScheduledAt.After(at) {
				continue
			}
			_, retryErr := ah.jobs.JobRetry(ctx, job.ID)
			ah.suite.Require().NoError(retryErr, "make %s job %d available", kind, job.ID)
		default:
			continue
		}
		ah.AwaitJob(ctx, job.ID)
	}
}

func (ah *appHarness) jobsOfKind(ctx context.Context, kind string) ([]*rivertype.JobRow, error) {
	params := river.NewJobListParams().
		Kinds(kind).
		OrderBy(river.JobListOrderByID, river.SortOrderAsc).
		First(100)
	var rows []*rivertype.JobRow
	for {
		page, listErr := ah.jobs.JobList(ctx, params)
		if listErr != nil {
			return nil, fmt.Errorf("list %s jobs: %w", kind, listErr)
		}
		rows = append(rows, page.Jobs...)
		if len(page.Jobs) < 100 {
			return rows, nil
		}
		params = params.After(page.LastCursor)
	}
}

func (ah *appHarness) processJobs(ctx context.Context, tenantID int, event rez.ProviderEvent) ([]*rivertype.JobRow, error) {
	rows, listErr := ah.jobsOfKind(ctx, (&jobs.ProcessProviderEventArgs{}).Kind())
	if listErr != nil {
		return nil, listErr
	}
	var matches []*rivertype.JobRow
	for _, row := range rows {
		var args jobs.ProcessProviderEventArgs
		if decodeErr := json.Unmarshal(row.EncodedArgs, &args); decodeErr != nil {
			return nil, fmt.Errorf("decode process job %d: %w", row.ID, decodeErr)
		}
		if args.TenantID != tenantID || args.Event.Provider != event.Provider ||
			args.Event.ProviderNamespace != event.ProviderNamespace ||
			args.Event.ProviderEventSource != event.ProviderEventSource ||
			args.Event.ProviderEventRef != event.ProviderEventRef {
			continue
		}
		matches = append(matches, row)
	}
	return matches, nil
}

func (ah *appHarness) projectionJobs(ctx context.Context, eventID uuid.UUID) ([]*rivertype.JobRow, error) {
	rows, listErr := ah.jobsOfKind(ctx, (jobs.ProjectNormalizedEvent{}).Kind())
	if listErr != nil {
		return nil, listErr
	}
	var matches []*rivertype.JobRow
	for _, row := range rows {
		var args jobs.ProjectNormalizedEvent
		if decodeErr := json.Unmarshal(row.EncodedArgs, &args); decodeErr != nil {
			return nil, fmt.Errorf("decode projection job %d: %w", row.ID, decodeErr)
		}
		if args.EventId == eventID {
			matches = append(matches, row)
		}
	}
	return matches, nil
}

func (ah *appHarness) ProcessJobIDs(ctx context.Context, tenantID int, event rez.ProviderEvent) []int64 {
	ah.suite.T().Helper()
	rows, listErr := ah.processJobs(ctx, tenantID, event)
	ah.suite.Require().NoError(listErr, "read jobs for provider delivery %s", event.ProviderEventRef)
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return ids
}

type providerProjection struct {
	Event         *ent.NormalizedEvent
	Receipt       *ent.NormalizedEventProjection
	ProcessJobIDs []int64
}

func (ah *appHarness) AwaitProjection(ctx context.Context, tenantID int, event rez.ProviderEvent) providerProjection {
	ah.suite.T().Helper()
	var processRows []*rivertype.JobRow
	ah.await(ctx, "provider process job", func(waitCtx context.Context) (bool, string, error) {
		rows, listErr := ah.processJobs(waitCtx, tenantID, event)
		processRows = rows
		return len(rows) > 0, "delivery has no process job yet", listErr
	})
	var processIDs []int64
	for _, row := range processRows {
		ah.AwaitJob(ctx, row.ID)
		processIDs = append(processIDs, row.ID)
	}

	queryEvent := ah.Client(ctx).NormalizedEvent.Query().
		Where(
			ne.Provider(event.Provider),
			ne.ProviderNamespace(event.ProviderNamespace),
			ne.ProviderEventSource(event.ProviderEventSource),
			ne.ProviderEventRef(event.ProviderEventRef),
		)
	normalized, eventErr := queryEvent.Only(ctx)
	ah.suite.Require().NoError(eventErr, "read normalized delivery after process job completion")

	var projectRows []*rivertype.JobRow
	ah.await(ctx, "normalized event projection job", func(waitCtx context.Context) (bool, string, error) {
		rows, listErr := ah.projectionJobs(waitCtx, normalized.ID)
		projectRows = rows
		return len(rows) > 0, "normalized event has no projection job yet", listErr
	})
	for _, row := range projectRows {
		ah.AwaitJob(ctx, row.ID)
	}

	queryReceipt := ah.Client(ctx).NormalizedEventProjection.Query().
		Where(nep.EventID(normalized.ID))
	receipt, receiptErr := queryReceipt.Only(ctx)
	ah.suite.Require().NoError(receiptErr, "read committed projection receipt")

	return providerProjection{
		Event:         normalized,
		Receipt:       receipt,
		ProcessJobIDs: processIDs,
	}
}

func (ah *appHarness) AwaitInvestigationTurn(ctx context.Context, sessionID uuid.UUID, sequence int) *ent.AgentTurn {
	ah.suite.T().Helper()
	var turn *ent.AgentTurn
	ah.await(ctx, fmt.Sprintf("session %s turn %d", sessionID, sequence), func(waitCtx context.Context) (bool, string, error) {
		query := ah.Client(waitCtx).AgentTurn.Query().
			Where(
				at.AgentSessionID(sessionID),
				at.Sequence(sequence),
			)
		row, queryErr := query.Only(waitCtx)
		if ent.IsNotFound(queryErr) {
			return false, "turn not yet created", nil
		}
		if queryErr != nil {
			return false, "query turn", queryErr
		}
		turn = row
		turnError := ""
		if row.Error != nil {
			turnError = *row.Error
		}
		diagnostic := fmt.Sprintf("id=%s status=%s error=%s job=%d", row.ID, row.Status, turnError, row.RiverJobID)
		if row.Status == at.StatusFailed {
			return false, diagnostic, fmt.Errorf("investigation turn failed")
		}
		return row.Status == at.StatusCompleted, diagnostic, nil
	})

	ah.AwaitJob(ctx, turn.RiverJobID)

	return turn
}
