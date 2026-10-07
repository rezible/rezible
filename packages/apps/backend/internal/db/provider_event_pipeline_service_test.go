package db

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/eventannotation"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	nep "github.com/rezible/rezible/ent/normalizedeventprojection"
	"github.com/rezible/rezible/pkg/jobs"
	"github.com/rezible/rezible/test"
	"github.com/rezible/rezible/test/mocks"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const (
	pipelineTestProvider  = "test"
	pipelineTestNamespace = "pipeline-account"
	pipelineTestSource    = "pipeline-test"
)

type ProviderEventPipelineServiceSuite struct {
	test.Suite
}

func TestProviderEventPipelineServiceSuite(t *testing.T) {
	suite.Run(t, &ProviderEventPipelineServiceSuite{
		Suite: test.NewSuite(),
	})
}

func (s *ProviderEventPipelineServiceSuite) newPipelineService(tdb rez.Database, jobSvc rez.JobService, proj rez.EventProjectionService) *ProviderEventPipelineService {
	return &ProviderEventPipelineService{
		logger:     slog.Default(),
		db:         tdb,
		jobs:       jobSvc,
		processors: map[string]rez.ProviderEventProcessor{pipelineTestProvider: pipelineTestProcessor{}},
		projection: proj,
	}
}

func (s *ProviderEventPipelineServiceSuite) makeTestEvent() rez.ProviderEvent {
	receivedAt := time.Date(2026, 6, 4, 9, 30, 0, 0, time.UTC)
	return rez.ProviderEvent{
		Provider:            pipelineTestProvider,
		ProviderNamespace:   pipelineTestNamespace,
		ProviderEventSource: pipelineTestSource,
		ProviderEventRef:    "delivery-" + uuid.NewString(),
		ReceivedAt:          receivedAt,
		Attributes:          []byte(`{"summary":"received"}`),
	}
}

func (s *ProviderEventPipelineServiceSuite) makeTestUser(ctx context.Context, tdb rez.Database) *ent.User {
	create := tdb.Client(ctx).User.Create().
		SetEmail("pipeline-test+" + uuid.NewString() + "@example.com").
		SetName("Pipeline Test User")
	user, userErr := create.Save(ctx)
	s.Require().NoError(userErr)

	return user
}

func (s *ProviderEventPipelineServiceSuite) createPipelineNormalizedEvent(ctx context.Context, tdb rez.Database) *ent.NormalizedEvent {
	ev := s.makeTestEvent()
	createNormalized := tdb.Client(ctx).NormalizedEvent.Create().
		SetProvider(ev.Provider).
		SetProviderNamespace(ev.ProviderNamespace).
		SetProviderEventSource(ev.ProviderEventSource).
		SetProviderEventRef("normalized-" + uuid.NewString()).
		SetProviderResourceRef("subject-1").
		SetKind("pipeline_test").
		SetOccurredAt(ev.ReceivedAt.Add(-time.Minute)).
		SetReceivedAt(ev.ReceivedAt).
		SetAttributes([]byte(`{"summary": "processed subject-1"}`))
	normalized, normalizedErr := createNormalized.Save(ctx)
	s.Require().NoError(normalizedErr)

	return normalized
}

func (s *ProviderEventPipelineServiceSuite) TestIngestProcessAndProjectEndToEnd() {
	ctx, tdb := s.SetupTestDatabase()

	jobSvc := mocks.NewMockJobService(s.T())

	projector := &countingEventProjectionService{}
	svc := s.newPipelineService(tdb, jobSvc, projector)

	ev := s.makeTestEvent()

	var capturedProcessArgs *jobs.ProcessProviderEventArgs
	jobSvc.EXPECT().
		Insert(mock.Anything, mock.Anything, mock.Anything).
		Run(func(_ context.Context, args river.JobArgs, opts *river.InsertOpts) {
			s.Require().NotNil(opts)
			s.True(opts.UniqueOpts.ByArgs)

			var ok bool
			capturedProcessArgs, ok = args.(*jobs.ProcessProviderEventArgs)
			s.Require().True(ok)
		}).
		Return(&rivertype.JobInsertResult{}, nil).
		Once()

	s.Require().NoError(svc.Ingest(ctx, ev))
	s.Equal(ev, capturedProcessArgs.Event)

	var capturedProjectArgs jobs.ProjectNormalizedEvent
	jobSvc.EXPECT().
		InsertMany(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, params []river.InsertManyParams) ([]*rivertype.JobInsertResult, error) {
			s.Require().Len(params, 1)

			var ok bool
			capturedProjectArgs, ok = params[0].Args.(jobs.ProjectNormalizedEvent)
			s.Require().True(ok)

			return []*rivertype.JobInsertResult{{}}, nil
		}).
		Once()

	s.Require().NoError(svc.HandleProcessEventJob(ctx, capturedProcessArgs))

	queryNormalized := tdb.Client(ctx).NormalizedEvent.Query().
		Where(ne.ProviderEventRef(ev.ProviderEventRef))
	normalized, normalizedErr := queryNormalized.Only(ctx)
	s.Require().NoError(normalizedErr)
	s.Equal(pipelineTestProvider, normalized.Provider)
	s.Equal(pipelineTestNamespace, normalized.ProviderNamespace)
	s.Equal(pipelineTestSource, normalized.ProviderEventSource)
	s.Equal("subject-1", normalized.ProviderResourceRef)
	s.Equal("pipeline_test", normalized.Kind)
	s.Equal(`{"summary": "processed subject-1"}`, string(normalized.Attributes))
	s.Equal(normalized.ID, capturedProjectArgs.EventId)

	s.Require().NoError(svc.HandleEventProjectionJob(ctx, capturedProjectArgs))

	s.Require().NotZero(projector.calls)

	queryProj := tdb.Client(ctx).NormalizedEventProjection.Query().
		Where(nep.EventID(normalized.ID))
	proj, projErr := queryProj.Only(ctx)
	s.Require().NoError(projErr)
	s.False(proj.CompletedAt.IsZero())
}

func (s *ProviderEventPipelineServiceSuite) TestSyncEventsEnqueuesEventPointers() {
	ctx, tdb := s.SetupTestDatabase()
	event := s.makeTestEvent()

	secondEvent := event
	secondEvent.ProviderEventRef = "second-event"

	jq := mocks.NewMockJobService(s.T())
	jq.EXPECT().
		InsertMany(mock.Anything, mock.Anything).
		Run(func(_ context.Context, params []river.InsertManyParams) {
			s.Require().Len(params, 2)
			for i, param := range params {
				args, argsOK := param.Args.(*jobs.ProcessProviderEventArgs)
				s.Require().True(argsOK)
				s.Equal([]rez.ProviderEvent{event, secondEvent}[i], args.Event)
				s.True(param.InsertOpts.UniqueOpts.ByArgs)
			}
		}).
		Return([]*rivertype.JobInsertResult{{}, {}}, nil).
		Once()

	querierFn := func(yield func(*rez.ProviderEventQueryResult, error) bool) {
		if yield(&rez.ProviderEventQueryResult{
			Event:                          event,
			ProviderEventSourceCursorAfter: new("next"),
		}, nil) {
			yield(&rez.ProviderEventQueryResult{Event: secondEvent}, nil)
		}
	}

	svc := s.newPipelineService(tdb, jq, nil)
	result := svc.SyncEvents(ctx, providerEventQuerierFunc(querierFn), make(rez.ProviderEventSourceCursors))
	s.Require().Empty(result.SyncErrors)
	s.Equal(2, result.EventsIngested)
}

func (s *ProviderEventPipelineServiceSuite) TestIngestManyQueuesAllEventsInOneInsert() {
	ctx, tdb := s.SetupTestDatabase()
	event := s.makeTestEvent()
	secondEvent := s.makeTestEvent()

	jq := mocks.NewMockJobService(s.T())
	jq.EXPECT().
		InsertMany(mock.Anything, mock.Anything).
		Run(func(_ context.Context, params []river.InsertManyParams) {
			s.Require().Len(params, 2)
			for i, param := range params {
				args, argsOK := param.Args.(*jobs.ProcessProviderEventArgs)
				s.Require().True(argsOK)
				s.Equal([]rez.ProviderEvent{event, secondEvent}[i], args.Event)
				s.True(param.InsertOpts.UniqueOpts.ByArgs)
			}
		}).
		Return([]*rivertype.JobInsertResult{{}, {}}, nil).
		Once()

	svc := s.newPipelineService(tdb, jq, nil)
	s.Require().NoError(svc.IngestMany(ctx, []rez.ProviderEvent{event, secondEvent}))
}

func (s *ProviderEventPipelineServiceSuite) TestIngestManyQueuesNothingWhenAnyEventIsInvalid() {
	ctx, tdb := s.SetupTestDatabase()
	invalid := s.makeTestEvent()
	invalid.ProviderEventRef = ""

	// The job service mock fails the test on any insert.
	svc := s.newPipelineService(tdb, mocks.NewMockJobService(s.T()), nil)
	s.Require().Error(svc.IngestMany(ctx, []rez.ProviderEvent{s.makeTestEvent(), invalid}))
}

type providerEventQuerierFunc func(func(*rez.ProviderEventQueryResult, error) bool)

func (f providerEventQuerierFunc) QueryProviderEvents(context.Context, rez.ProviderEventSourceCursors) iter.Seq2[*rez.ProviderEventQueryResult, error] {
	return func(yield func(*rez.ProviderEventQueryResult, error) bool) {
		f(yield)
	}
}

func (s *ProviderEventPipelineServiceSuite) TestProcessProviderEventDoesNotReinsertOrReprojectDuplicate() {
	ctx, tdb := s.SetupTestDatabase()

	jobSvc := mocks.NewMockJobService(s.T())
	jobSvc.EXPECT().
		InsertMany(mock.Anything, mock.Anything).
		Return([]*rivertype.JobInsertResult{{}}, nil).
		Once()

	svc := s.newPipelineService(tdb, jobSvc, nil)

	args := &jobs.ProcessProviderEventArgs{
		Event: s.makeTestEvent(),
	}

	s.Require().NoError(svc.HandleProcessEventJob(ctx, args))
	s.Require().NoError(svc.HandleProcessEventJob(ctx, args))

	queryCount := tdb.Client(ctx).NormalizedEvent.Query().
		Where(ne.ProviderEventRef(args.Event.ProviderEventRef))
	count, countErr := queryCount.Count(ctx)
	s.Require().NoError(countErr)
	s.Equal(1, count)
}

func (s *ProviderEventPipelineServiceSuite) TestNormalizedEventIdentityIncludesNamespaceAndResource() {
	ctx, tdb := s.SetupTestDatabase()

	jobSvc := mocks.NewMockJobService(s.T())
	jobSvc.EXPECT().
		InsertMany(mock.Anything, mock.Anything).
		Run(func(_ context.Context, params []river.InsertManyParams) {
			s.Len(params, 2)
		}).
		Return([]*rivertype.JobInsertResult{{}, {}}, nil).
		Twice()

	svc := s.newPipelineService(tdb, jobSvc, nil)
	svc.processors = map[string]rez.ProviderEventProcessor{
		pipelineTestProvider: multiResourcePipelineProcessor{},
	}
	event := s.makeTestEvent()
	event.ProviderEventRef = "same-delivery"

	processArgs := &jobs.ProcessProviderEventArgs{
		Event: event,
	}
	s.Require().NoError(svc.HandleProcessEventJob(ctx, processArgs))

	otherNamespaceEvent := event
	otherNamespaceEvent.ProviderNamespace = "another-pipeline-account"
	otherProcessArgs := &jobs.ProcessProviderEventArgs{
		Event: otherNamespaceEvent,
	}
	s.Require().NoError(svc.HandleProcessEventJob(ctx, otherProcessArgs))

	queryEvents := tdb.Client(ctx).NormalizedEvent.Query().
		Where(ne.ProviderEventRef(event.ProviderEventRef))
	normalized, queryErr := queryEvents.All(ctx)
	s.Require().NoError(queryErr)
	s.Require().Len(normalized, 4)
	type eventIdentity struct{ namespace, resource string }
	identities := make([]eventIdentity, 0, len(normalized))
	for index, item := range normalized {
		identities = append(identities, eventIdentity{
			namespace: item.ProviderNamespace,
			resource:  item.ProviderResourceRef,
		})
		s.NotEqual(uuid.Nil, item.ID)
		for _, previous := range normalized[:index] {
			s.NotEqual(previous.ID, item.ID)
		}
	}
	s.ElementsMatch([]eventIdentity{
		{
			namespace: pipelineTestNamespace,
			resource:  "subject-1",
		},
		{
			namespace: pipelineTestNamespace,
			resource:  "subject-2",
		},
		{
			namespace: otherNamespaceEvent.ProviderNamespace,
			resource:  "subject-1",
		},
		{
			namespace: otherNamespaceEvent.ProviderNamespace,
			resource:  "subject-2",
		},
	}, identities)
}

func (s *ProviderEventPipelineServiceSuite) createProjectionResult(ctx context.Context, tdb rez.Database, eventId uuid.UUID) {
	createNormalizedEventProjection := tdb.Client(ctx).NormalizedEventProjection.Create().
		SetEventID(eventId)

	s.Require().NoError(createNormalizedEventProjection.Exec(ctx))
}

func (s *ProviderEventPipelineServiceSuite) getProjection(ctx context.Context, tdb rez.Database, eventId uuid.UUID) (*ent.NormalizedEventProjection, error) {
	queryNormalizedEventProjection := tdb.Client(ctx).NormalizedEventProjection.Query().
		Where(nep.EventID(eventId))

	return queryNormalizedEventProjection.Only(ctx)
}

func (s *ProviderEventPipelineServiceSuite) TestProjectionSkipsEventWithReceipt() {
	ctx, tdb := s.SetupTestDatabase()
	projector := &countingEventProjectionService{}
	svc := s.newPipelineService(tdb, mocks.NewMockJobService(s.T()), projector)
	ev := s.createPipelineNormalizedEvent(ctx, tdb)

	s.createProjectionResult(ctx, tdb, ev.ID)

	s.Require().NoError(svc.HandleEventProjectionJob(ctx, jobs.ProjectNormalizedEvent{
		EventId: ev.ID,
	}))
	s.Equal(0, projector.calls)
}

func (s *ProviderEventPipelineServiceSuite) TestProjectionFailureRollsBackProjectorWrites() {
	ctx, tdb := s.SetupTestDatabase()
	creator := s.makeTestUser(ctx, tdb)
	projector := &rollbackPipelineProjector{
		db:        tdb,
		creatorID: creator.ID,
	}
	svc := s.newPipelineService(tdb, mocks.NewMockJobService(s.T()), projector)
	ev := s.createPipelineNormalizedEvent(ctx, tdb)

	projectionErr := svc.HandleEventProjectionJob(ctx, jobs.ProjectNormalizedEvent{
		EventId: ev.ID,
	})
	s.Require().Error(projectionErr)
	var cancelErr *river.JobCancelError
	s.Require().ErrorAs(projectionErr, &cancelErr)

	_, projErr := s.getProjection(ctx, tdb, ev.ID)
	s.True(ent.IsNotFound(projErr))

	queryCount := tdb.Client(ctx).EventAnnotation.Query().
		Where(eventannotation.EventID(ev.ID))
	count, countErr := queryCount.Count(ctx)
	s.Require().NoError(countErr)
	s.Equal(0, count)
}

func (s *ProviderEventPipelineServiceSuite) TestRetryableProjectionFailureReturnsError() {
	ctx, tdb := s.SetupTestDatabase()
	svc := s.newPipelineService(tdb, mocks.NewMockJobService(s.T()), &retryablePipelineProjector{})
	ev := s.createPipelineNormalizedEvent(ctx, tdb)

	projectionErr := svc.HandleEventProjectionJob(ctx, jobs.ProjectNormalizedEvent{
		EventId: ev.ID,
	})
	s.Require().Error(projectionErr)
	s.True(jobs.IsRetryableError(projectionErr))

	_, projErr := s.getProjection(ctx, tdb, ev.ID)
	s.True(ent.IsNotFound(projErr))
}

func (s *ProviderEventPipelineServiceSuite) TestTransientDatabaseProjectionFailureReturnsRetryableError() {
	ctx, tdb := s.SetupTestDatabase()
	transientErr := errors.New("database deadlock")
	projector := &transientDatabasePipelineProjector{
		projectionErr: transientErr,
	}
	svc := s.newPipelineService(tdb, mocks.NewMockJobService(s.T()), projector)
	svc.db = transientDatabase{
		Database:     tdb,
		transientErr: transientErr,
	}
	ev := s.createPipelineNormalizedEvent(ctx, tdb)

	projectionErr := svc.HandleEventProjectionJob(ctx, jobs.ProjectNormalizedEvent{
		EventId: ev.ID,
	})
	s.Require().Error(projectionErr)
	s.True(jobs.IsRetryableError(projectionErr))

	_, projErr := s.getProjection(ctx, tdb, ev.ID)
	s.True(ent.IsNotFound(projErr))
}

func (s *ProviderEventPipelineServiceSuite) TestProjectionPanicCancelsJobWithoutReceipt() {
	ctx, tdb := s.SetupTestDatabase()
	svc := s.newPipelineService(tdb, mocks.NewMockJobService(s.T()), &panickingEventProjectionService{
		panicText: "boom",
	})
	ev := s.createPipelineNormalizedEvent(ctx, tdb)

	projectionErr := svc.HandleEventProjectionJob(ctx, jobs.ProjectNormalizedEvent{
		EventId: ev.ID,
	})
	s.Require().Error(projectionErr)
	var cancelErr *river.JobCancelError
	s.Require().ErrorAs(projectionErr, &cancelErr)
	s.Contains(projectionErr.Error(), "projector panic: boom")

	_, projErr := s.getProjection(ctx, tdb, ev.ID)
	s.True(ent.IsNotFound(projErr))
}

func (s *ProviderEventPipelineServiceSuite) TestConcurrentProjectionRunsHandlerOnce() {
	tenantCtx, tdb := s.SetupTestDatabase()
	ctx, cancel := context.WithTimeout(tenantCtx, 10*time.Second)
	defer cancel()
	projector := &blockingPipelineProjector{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	svc := s.newPipelineService(tdb, mocks.NewMockJobService(s.T()), projector)
	ev := s.createPipelineNormalizedEvent(ctx, tdb)
	args := jobs.ProjectNormalizedEvent{
		EventId: ev.ID,
	}
	results := make(chan error, 2)

	go func() {
		results <- svc.HandleEventProjectionJob(ctx, args)
	}()
	select {
	case <-projector.started:
	case <-ctx.Done():
		s.FailNow("projector did not start", ctx.Err().Error())
	}
	go func() {
		results <- svc.HandleEventProjectionJob(ctx, args)
	}()
	close(projector.release)

	for range 2 {
		select {
		case projectionErr := <-results:
			s.Require().NoError(projectionErr)
		case <-ctx.Done():
			s.FailNow("concurrent projections did not finish", ctx.Err().Error())
		}
	}
	s.Equal(int32(1), projector.calls.Load())
	_, receiptErr := s.getProjection(ctx, tdb, ev.ID)
	s.Require().NoError(receiptErr)
}

type pipelineTestProcessor struct{}

func (pipelineTestProcessor) ProcessProviderEvent(_ context.Context, ev rez.ProviderEvent) (ent.NormalizedEvents, error) {
	return ent.NormalizedEvents{
		{
			Provider:            ev.Provider,
			ProviderNamespace:   ev.ProviderNamespace,
			ProviderEventSource: ev.ProviderEventSource,
			ProviderEventRef:    ev.ProviderEventRef,
			ProviderResourceRef: "subject-1",
			Kind:                "pipeline_test",
			OccurredAt:          ev.ReceivedAt.Add(-time.Minute),
			ReceivedAt:          ev.ReceivedAt,
			Attributes:          []byte(`{"summary": "processed subject-1"}`),
		},
	}, nil
}

type multiResourcePipelineProcessor struct{}

func (multiResourcePipelineProcessor) ProcessProviderEvent(_ context.Context, ev rez.ProviderEvent) (ent.NormalizedEvents, error) {
	return ent.NormalizedEvents{
		{
			Provider:            ev.Provider,
			ProviderNamespace:   ev.ProviderNamespace,
			ProviderEventSource: ev.ProviderEventSource,
			ProviderEventRef:    ev.ProviderEventRef,
			ProviderResourceRef: "subject-1",
			Kind:                "pipeline_test",
			OccurredAt:          ev.ReceivedAt.Add(-time.Minute),
			ReceivedAt:          ev.ReceivedAt,
			Attributes:          []byte(`{"summary": "processed subject-1"}`),
		},
		{
			Provider:            ev.Provider,
			ProviderNamespace:   ev.ProviderNamespace,
			ProviderEventSource: ev.ProviderEventSource,
			ProviderEventRef:    ev.ProviderEventRef,
			ProviderResourceRef: "subject-2",
			Kind:                "pipeline_test",
			OccurredAt:          ev.ReceivedAt.Add(-time.Minute),
			ReceivedAt:          ev.ReceivedAt,
			Attributes:          []byte(`{"summary": "processed subject-2"}`),
		},
	}, nil
}

type panickingEventProjectionService struct {
	panicText string
}

func (p *panickingEventProjectionService) GetEventProjectorFunc(*ent.NormalizedEvent) (rez.EventProjectorFunc, bool) {
	return func(context.Context, *ent.NormalizedEvent) ([]rez.ProjectedEntityRef, error) {
		msg := p.panicText
		if msg == "" {
			msg = "boom"
		}
		panic(msg)
		return nil, nil
	}, true
}

type countingEventProjectionService struct {
	calls int
}

func (p *countingEventProjectionService) GetEventProjectorFunc(*ent.NormalizedEvent) (rez.EventProjectorFunc, bool) {
	return func(context.Context, *ent.NormalizedEvent) ([]rez.ProjectedEntityRef, error) {
		p.calls++
		return nil, nil
	}, true
}

type blockingPipelineProjector struct {
	calls   atomic.Int32
	started chan struct{}
	release chan struct{}
}

func (p *blockingPipelineProjector) GetEventProjectorFunc(*ent.NormalizedEvent) (rez.EventProjectorFunc, bool) {
	return func(ctx context.Context, _ *ent.NormalizedEvent) ([]rez.ProjectedEntityRef, error) {
		p.calls.Add(1)
		close(p.started)
		select {
		case <-p.release:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		return nil, nil
	}, true
}

type rollbackPipelineProjector struct {
	db        rez.Database
	creatorID uuid.UUID
}

func (p *rollbackPipelineProjector) GetEventProjectorFunc(*ent.NormalizedEvent) (rez.EventProjectorFunc, bool) {
	return func(ctx context.Context, ev *ent.NormalizedEvent) ([]rez.ProjectedEntityRef, error) {
		create := p.db.Client(ctx).EventAnnotation.Create().
			SetEventID(ev.ID).
			SetCreatorID(p.creatorID).
			SetMinutesOccupied(1).
			SetNotes("should roll back").
			SetTags([]string{"rollback"})
		if createErr := create.Exec(ctx); createErr != nil {
			return nil, createErr
		}
		return nil, errors.New("projection failed after write")
	}, true
}

type retryablePipelineProjector struct{}

func (p *retryablePipelineProjector) GetEventProjectorFunc(*ent.NormalizedEvent) (rez.EventProjectorFunc, bool) {
	return func(ctx context.Context, ev *ent.NormalizedEvent) ([]rez.ProjectedEntityRef, error) {
		return nil, jobs.MarkRetryableError(fmt.Errorf("dependency not ready"))
	}, true
}

type transientDatabase struct {
	rez.Database
	transientErr error
}

func (d transientDatabase) IsTransientError(projectionErr error) bool {
	return errors.Is(projectionErr, d.transientErr)
}

type transientDatabasePipelineProjector struct {
	projectionErr error
}

func (p *transientDatabasePipelineProjector) GetEventProjectorFunc(*ent.NormalizedEvent) (rez.EventProjectorFunc, bool) {
	return func(ctx context.Context, ev *ent.NormalizedEvent) ([]rez.ProjectedEntityRef, error) {
		return nil, p.projectionErr
	}, true
}
