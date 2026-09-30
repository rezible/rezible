package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"slices"
	"sync"
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
	"github.com/stretchr/testify/require"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	at "github.com/rezible/rezible/ent/agentturn"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	nep "github.com/rezible/rezible/ent/normalizedeventprojection"
	"github.com/rezible/rezible/internal/genkit"
	"github.com/rezible/rezible/internal/http"
	demo "github.com/rezible/rezible/internal/integrations/demo"
	"github.com/rezible/rezible/internal/postgres"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/pkg/jobs"
	oapi "github.com/rezible/rezible/pkg/openapi"
	oapiv1 "github.com/rezible/rezible/pkg/openapi/v1"
	"github.com/rezible/rezible/test"
)

const (
	// Waits return as soon as their condition holds; this only bounds failures.
	appWaitTimeout           = 30 * time.Second
	investigationReportText  = "Checkout errors need investigation."
	investigationQuestion    = "What should we check next?"
	investigationAnswerTitle = "Next check"
	investigationAnswerBody  = "Check the checkout error rate."
)

type appHarness struct {
	suite    *BackendSuite
	app      *Application
	database rez.Database
	jobs     *river.Client[pgx.Tx]
	api      *oapi.TestAPI
	cookie   rez.AppAuthSessionCookie
}

func (s *BackendSuite) newAppHarness(modelAction ai.ModelActionFunc[any]) *appHarness {
	t := s.T()
	t.Helper()
	telemetry := s.Telemetry()
	ctx, cancel := context.WithCancel(t.Context())
	app := NewApplication()

	// The injector owns services and the database. The harness owns the
	// lifecycle goroutine.
	var lifecycleDone <-chan error
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
	})

	_, initErr := app.Init(ctx)
	require.NoError(t, initErr, "initialize application")

	cfg := s.Config()
	app.Override[rez.Config](func(do.Injector) (rez.Config, error) {
		return cfg, nil
	})
	app.Override[rez.TelemetryService](func(do.Injector) (rez.TelemetryService, error) {
		return telemetry, nil
	})
	app.Override[rez.PostgresConfig](providePostgresTestDatabaseConfig)
	// Select the real demo processor without installing its unrelated sync jobs.
	app.Override[[]rez.IntegrationDefinition](func(i do.Injector) ([]rez.IntegrationDefinition, error) {
		integration, integrationErr := do.Invoke[*demo.Integration](i)
		if integrationErr != nil {
			return nil, fmt.Errorf("resolve demo integration: %w", integrationErr)
		}
		return []rez.IntegrationDefinition{integration}, nil
	})

	options, optionsErr := app.invoke[[]genkit.AiRuntimeOption]()
	require.NoError(t, optionsErr, "resolve runtime options")
	if modelAction == nil {
		modelAction = func(context.Context, *ai.ModelRequest, any, ai.ModelStreamCallback) (*ai.ModelResponse, error) {
			return nil, fmt.Errorf("unexpected model call in application test")
		}
	}
	modelOptions := &ai.ModelOptions{
		Supports: &ai.ModelSupports{
			Tools:      true,
			Multiturn:  true,
			SystemRole: true,
		},
	}
	model := genkit.NewModelDefinition("test/application", modelOptions, modelAction)
	model.IsDefault = true
	options = slices.Clone(options)
	options = append(options, genkit.WithDefinedModel(model))
	app.Override[[]genkit.AiRuntimeOption](func(do.Injector) ([]genkit.AiRuntimeOption, error) {
		return options, nil
	})

	definition, definitionErr := do.InvokeNamed[jobs.Definition](app.i, "jobs-default")
	require.NoError(t, definitionErr, "resolve application jobs")
	definition.PeriodicJobs = nil
	do.OverrideNamed(app.i, "jobs-default", func(do.Injector) (jobs.Definition, error) {
		return definition, nil
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
	require.NoError(t, databaseErr, "resolve application database")
	pool, poolErr := app.invoke[*postgres.ConnectionPool]()
	require.NoError(t, poolErr, "resolve application pool")
	jobConfig := &river.Config{
		Schema: "river",
		Logger: telemetry.Logger(),
	}
	jobReader, jobReaderErr := river.NewClient(riverpgxv5.New(pool.Pool), jobConfig)
	require.NoError(t, jobReaderErr, "create read-only River client")

	cookie, cookieErr := app.invoke[rez.AppAuthSessionCookie]()
	require.NoError(t, cookieErr, "resolve application authentication")
	api, apiErr := app.invoke[oapiv1.API]()
	require.NoError(t, apiErr, "resolve application v1 API")
	server, serverErr := app.invoke[*http.Server]()
	require.NoError(t, serverErr, "resolve application HTTP server")
	testAPI := oapiv1.NewTestAPI(t, api).
		WithHandler(server.Handler(), cfg.HttpServer.BasePath+oapiv1.VersionPrefix)
	return &appHarness{
		suite:    s,
		app:      app,
		database: database,
		jobs:     jobReader,
		api:      testAPI,
		cookie:   cookie,
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
		require.NoError(t, probeErr, "%s: %s", label, diagnostic)
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
	t := ah.suite.T()
	t.Helper()
	rows, listErr := ah.processJobs(ctx, tenantID, event)
	require.NoError(t, listErr, "read jobs for provider delivery %s", event.ProviderEventRef)
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
	t := ah.suite.T()
	t.Helper()
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
	require.NoError(t, eventErr, "read normalized delivery after process job completion")

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
	require.NoError(t, receiptErr, "read committed projection receipt")
	return providerProjection{
		Event:         normalized,
		Receipt:       receipt,
		ProcessJobIDs: processIDs,
	}
}

func (ah *appHarness) AwaitInvestigationTurn(ctx context.Context, sessionID uuid.UUID, sequence int) *ent.AgentTurn {
	ah.suite.T().Helper()
	var turn *ent.AgentTurn
	label := fmt.Sprintf("session %s turn %d", sessionID, sequence)
	ah.await(ctx, label, func(waitCtx context.Context) (bool, string, error) {
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

// The script exercises real tools across two turns; it never writes domain data.
type investigationModel struct {
	mu   sync.Mutex
	step int
}

func (m *investigationModel) AssertComplete(t *testing.T) {
	t.Helper()
	m.mu.Lock()
	defer m.mu.Unlock()
	assert.Equal(t, 4, m.step, "investigation model must consume all four calls")
}

func (m *investigationModel) generate(
	_ context.Context,
	request *ai.ModelRequest,
	_ any,
	_ ai.ModelStreamCallback,
) (*ai.ModelResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.step++

	switch m.step {
	case 1:
		input := rezai.PublishInvestigationReportToolInput{
			Text:         investigationReportText,
			EvidenceRefs: []string{},
		}
		call := &ai.ToolRequest{
			Name:  rezai.PublishInvestigationReportTool.Name(),
			Ref:   "report-1",
			Input: input,
		}
		return m.requestTool(request, call)

	case 2:
		var result rezai.InvestigationReportToolResult
		responseErr := m.decodeToolResult(request, rezai.PublishInvestigationReportTool.Name(), "report-1", &result)
		if responseErr != nil {
			return nil, responseErr
		}
		if result.Text != investigationReportText {
			return nil, fmt.Errorf("step 2: report text=%q, want %q", result.Text, investigationReportText)
		}
		if len(result.EvidenceRefs) != 0 {
			return nil, fmt.Errorf("step 2: report evidence_refs=%v, want empty", result.EvidenceRefs)
		}
		return &ai.ModelResponse{
			Message:      ai.NewModelTextMessage("Recorded."),
			FinishReason: ai.FinishReasonStop,
		}, nil

	case 3:
		expectedQuestion := "User question:\n" + investigationQuestion
		questionFound := false
		for _, message := range request.Messages {
			if message.Role == ai.RoleUser && message.Text() == expectedQuestion {
				questionFound = true
				break
			}
		}
		if !questionFound {
			return nil, fmt.Errorf("step 3: follow-up user question %q absent from model request", investigationQuestion)
		}
		input := rezai.PublishInvestigationAnswerToolInput{
			Title:        investigationAnswerTitle,
			Body:         investigationAnswerBody,
			EvidenceRefs: []string{},
		}
		call := &ai.ToolRequest{
			Name:  rezai.PublishInvestigationAnswerTool.Name(),
			Ref:   "answer-1",
			Input: input,
		}
		return m.requestTool(request, call)

	case 4:
		var result rezai.InvestigationFindingVersionToolResult
		responseErr := m.decodeToolResult(request, rezai.PublishInvestigationAnswerTool.Name(), "answer-1", &result)
		if responseErr != nil {
			return nil, responseErr
		}
		if result.Title != investigationAnswerTitle {
			return nil, fmt.Errorf("step 4: answer title=%q, want %q", result.Title, investigationAnswerTitle)
		}
		if result.Body != investigationAnswerBody {
			return nil, fmt.Errorf("step 4: answer body=%q, want %q", result.Body, investigationAnswerBody)
		}
		if !result.IsAnswer {
			return nil, fmt.Errorf("step 4: published finding is not an answer")
		}
		if len(result.EvidenceRefs) != 0 {
			return nil, fmt.Errorf("step 4: answer evidence_refs=%v, want empty", result.EvidenceRefs)
		}
		return &ai.ModelResponse{
			Message:      ai.NewModelTextMessage("Recorded."),
			FinishReason: ai.FinishReasonStop,
		}, nil

	default:
		return nil, fmt.Errorf("unexpected model call %d", m.step)
	}
}

func (m *investigationModel) requestTool(request *ai.ModelRequest, call *ai.ToolRequest) (*ai.ModelResponse, error) {
	for _, tool := range request.Tools {
		if tool.Name == call.Name {
			part := ai.NewToolRequestPart(call)
			return &ai.ModelResponse{
				Message: ai.NewModelMessage(part),
			}, nil
		}
	}
	return nil, fmt.Errorf("step %d: required tool %s missing", m.step, call.Name)
}

func (m *investigationModel) decodeToolResult(request *ai.ModelRequest, toolName, callID string, output any) error {
	for _, message := range request.Messages {
		for _, part := range message.Content {
			response := part.ToolResponse
			if response == nil || response.Name != toolName || response.Ref != callID {
				continue
			}
			encoded, encodeErr := json.Marshal(response.Output)
			if encodeErr != nil {
				return fmt.Errorf("step %d: encode %s result: %w", m.step, toolName, encodeErr)
			}
			if decodeErr := json.Unmarshal(encoded, output); decodeErr != nil {
				return fmt.Errorf("step %d: decode %s result: %w", m.step, toolName, decodeErr)
			}
			return nil
		}
	}
	return fmt.Errorf("step %d: response for tool %s call %s absent", m.step, toolName, callID)
}
