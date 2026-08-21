package db

import (
	"context"
	"errors"
	"fmt"
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
	"github.com/rezible/rezible/pkg/projections"
	"github.com/rezible/rezible/test"
	"github.com/rezible/rezible/test/mocks"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const (
	pipelineTestProvider    = "test"
	pipelineTestSource      = "pipeline-test"
	pipelineTestSubjectKind = projections.SubjectKind("PipelineTestSubject")
)

type ProviderEventPipelineServiceSuite struct {
	test.Suite
}

func TestProviderEventPipelineServiceSuite(t *testing.T) {
	suite.Run(t, &ProviderEventPipelineServiceSuite{Suite: test.NewSuite()})
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
		Provider:           pipelineTestProvider,
		ProviderSource:     pipelineTestSource,
		ProviderEventRef:   "delivery-" + uuid.NewString(),
		ProviderSubjectRef: "subject-1",
		ReceivedAt:         receivedAt,
		Payload:            []byte(`{"summary":"received"}`),
		ContentType:        "application/json",
	}
}

func (s *ProviderEventPipelineServiceSuite) makeTestUser(ctx context.Context, tdb rez.Database) *ent.User {
	create := tdb.Client(ctx).User.Create().
		SetEmail("pipeline-test+" + uuid.NewString() + "@example.com").
		SetName("Pipeline Test User")
	user, err := create.Save(ctx)
	s.Require().NoError(err)
	return user
}

func (s *ProviderEventPipelineServiceSuite) createPipelineNormalizedEvent(ctx context.Context, tdb rez.Database) *ent.NormalizedEvent {
	ev := s.makeTestEvent()
	normalized, err := tdb.Client(ctx).NormalizedEvent.Create().
		SetProvider(ev.Provider).
		SetProviderSource(ev.ProviderSource).
		SetProviderEventRef("normalized-" + uuid.NewString()).
		SetProviderSubjectRef(ev.ProviderSubjectRef).
		SetKind(ne.KindObserved).
		SetSubjectKind(pipelineTestSubjectKind.String()).
		SetOccurredAt(ev.ReceivedAt.Add(-time.Minute)).
		SetReceivedAt(ev.ReceivedAt).
		SetAttributes([]byte(fmt.Sprintf(`{"summary": "processed %s"}`, ev.ProviderSubjectRef))).
		Save(ctx)
	s.Require().NoError(err)
	return normalized
}

func (s *ProviderEventPipelineServiceSuite) TestIngestProcessAndProjectEndToEnd() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()

	jobSvc := mocks.NewMockJobService(s.T())

	projector := &countingEventProjectionService{}
	svc := s.newPipelineService(tdb, jobSvc, projector)

	ev := s.makeTestEvent()

	var capturedProcessArgs processProviderEventArgs
	jobSvc.EXPECT().
		Insert(mock.Anything, mock.Anything, mock.Anything).
		Run(func(_ context.Context, args river.JobArgs, opts *river.InsertOpts) {
			s.Require().NotNil(opts)
			s.True(opts.UniqueOpts.ByArgs)

			var ok bool
			capturedProcessArgs, ok = args.(processProviderEventArgs)
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
	s.Equal(pipelineTestSource, normalized.ProviderSource)
	s.Equal(ev.ProviderSubjectRef, normalized.ProviderSubjectRef)
	s.Equal(ne.KindObserved, normalized.Kind)
	s.Equal(pipelineTestSubjectKind.String(), normalized.SubjectKind)
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

func (s *ProviderEventPipelineServiceSuite) TestProcessProviderEventDoesNotReinsertOrReprojectDuplicate() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()

	jobSvc := mocks.NewMockJobService(s.T())
	jobSvc.EXPECT().
		InsertMany(mock.Anything, mock.Anything).
		Return([]*rivertype.JobInsertResult{{}}, nil).
		Twice()

	svc := s.newPipelineService(tdb, jobSvc, nil)

	args := processProviderEventArgs{Event: s.makeTestEvent()}

	s.Require().NoError(svc.HandleProcessEventJob(ctx, args))
	s.Require().NoError(svc.HandleProcessEventJob(ctx, args))

	queryCount := tdb.Client(ctx).NormalizedEvent.Query().
		Where(ne.ProviderEventRef(args.Event.ProviderEventRef))
	count, countErr := queryCount.Count(ctx)
	s.Require().NoError(countErr)
	s.Equal(1, count)
}

func (s *ProviderEventPipelineServiceSuite) createProjectionResult(ctx context.Context, tdb rez.Database, eventId uuid.UUID) {
	s.Require().NoError(tdb.Client(ctx).NormalizedEventProjection.Create().
		SetEventID(eventId).
		Exec(ctx))
}

func (s *ProviderEventPipelineServiceSuite) getProjection(ctx context.Context, tdb rez.Database, eventId uuid.UUID) (*ent.NormalizedEventProjection, error) {
	return tdb.Client(ctx).NormalizedEventProjection.Query().
		Where(nep.EventID(eventId)).
		Only(ctx)
}

func (s *ProviderEventPipelineServiceSuite) TestProjectionSkipsEventWithReceipt() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	projector := &countingEventProjectionService{}
	svc := s.newPipelineService(tdb, mocks.NewMockJobService(s.T()), projector)
	ev := s.createPipelineNormalizedEvent(ctx, tdb)

	s.createProjectionResult(ctx, tdb, ev.ID)

	s.Require().NoError(svc.HandleEventProjectionJob(ctx, jobs.ProjectNormalizedEvent{EventId: ev.ID}))
	s.Equal(0, projector.calls)
}

func (s *ProviderEventPipelineServiceSuite) TestProjectionFailureRollsBackProjectorWrites() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	creator := s.makeTestUser(ctx, tdb)
	projector := &rollbackPipelineProjector{db: tdb, creatorID: creator.ID}
	svc := s.newPipelineService(tdb, mocks.NewMockJobService(s.T()), projector)
	ev := s.createPipelineNormalizedEvent(ctx, tdb)

	err := svc.HandleEventProjectionJob(ctx, jobs.ProjectNormalizedEvent{EventId: ev.ID})
	s.Require().Error(err)
	var cancelErr *river.JobCancelError
	s.Require().ErrorAs(err, &cancelErr)

	_, projErr := s.getProjection(ctx, tdb, ev.ID)
	s.True(ent.IsNotFound(projErr))

	count, countErr := tdb.Client(ctx).EventAnnotation.Query().
		Where(eventannotation.EventID(ev.ID)).
		Count(ctx)
	s.Require().NoError(countErr)
	s.Equal(0, count)
}

func (s *ProviderEventPipelineServiceSuite) TestRetryableProjectionFailureReturnsError() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	svc := s.newPipelineService(tdb, mocks.NewMockJobService(s.T()), &retryablePipelineProjector{})
	ev := s.createPipelineNormalizedEvent(ctx, tdb)

	err := svc.HandleEventProjectionJob(ctx, jobs.ProjectNormalizedEvent{EventId: ev.ID})
	s.Require().Error(err)
	s.True(projections.IsRetryable(err))

	_, projErr := s.getProjection(ctx, tdb, ev.ID)
	s.True(ent.IsNotFound(projErr))
}

func (s *ProviderEventPipelineServiceSuite) TestTransientDatabaseProjectionFailureReturnsRetryableError() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	transientErr := errors.New("database deadlock")
	projector := &transientDatabasePipelineProjector{err: transientErr}
	svc := s.newPipelineService(tdb, mocks.NewMockJobService(s.T()), projector)
	svc.db = transientDatabase{Database: tdb, transientErr: transientErr}
	ev := s.createPipelineNormalizedEvent(ctx, tdb)

	err := svc.HandleEventProjectionJob(ctx, jobs.ProjectNormalizedEvent{EventId: ev.ID})
	s.Require().Error(err)
	s.True(projections.IsRetryable(err))

	_, projErr := s.getProjection(ctx, tdb, ev.ID)
	s.True(ent.IsNotFound(projErr))
}

func (s *ProviderEventPipelineServiceSuite) TestProjectionPanicCancelsJobWithoutReceipt() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	svc := s.newPipelineService(tdb, mocks.NewMockJobService(s.T()), &panickingEventProjectionService{panicText: "boom"})
	ev := s.createPipelineNormalizedEvent(ctx, tdb)

	err := svc.HandleEventProjectionJob(ctx, jobs.ProjectNormalizedEvent{EventId: ev.ID})
	s.Require().Error(err)
	var cancelErr *river.JobCancelError
	s.Require().ErrorAs(err, &cancelErr)
	s.Contains(err.Error(), "projector panic: boom")

	_, projErr := s.getProjection(ctx, tdb, ev.ID)
	s.True(ent.IsNotFound(projErr))
}

func (s *ProviderEventPipelineServiceSuite) TestConcurrentProjectionRunsHandlerOnce() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	projector := &blockingPipelineProjector{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	svc := s.newPipelineService(tdb, mocks.NewMockJobService(s.T()), projector)
	ev := s.createPipelineNormalizedEvent(ctx, tdb)
	args := jobs.ProjectNormalizedEvent{EventId: ev.ID}
	results := make(chan error, 2)

	go func() {
		results <- svc.HandleEventProjectionJob(ctx, args)
	}()
	<-projector.started
	go func() {
		results <- svc.HandleEventProjectionJob(ctx, args)
	}()
	close(projector.release)

	s.Require().NoError(<-results)
	s.Require().NoError(<-results)
	s.Equal(int32(1), projector.calls.Load())
	_, receiptErr := s.getProjection(ctx, tdb, ev.ID)
	s.Require().NoError(receiptErr)
}

type pipelineTestProcessor struct{}

func (pipelineTestProcessor) ProcessProviderEvent(_ context.Context, ev rez.ProviderEvent) (ent.NormalizedEvents, error) {
	return ent.NormalizedEvents{
		{
			Provider:           ev.Provider,
			ProviderSource:     ev.ProviderSource,
			ProviderEventRef:   ev.ProviderEventRef,
			ProviderSubjectRef: ev.ProviderSubjectRef,
			Kind:               ne.KindObserved,
			SubjectKind:        pipelineTestSubjectKind.String(),
			OccurredAt:         ev.ReceivedAt.Add(-time.Minute),
			ReceivedAt:         ev.ReceivedAt,
			Attributes:         []byte(fmt.Sprintf(`{"summary": "processed %s"}`, ev.ProviderSubjectRef)),
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
	return func(context.Context, *ent.NormalizedEvent) ([]rez.ProjectedEntityRef, error) {
		p.calls.Add(1)
		close(p.started)
		<-p.release
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
		return nil, projections.Retryable(errors.New("dependency not ready"))
	}, true
}

type transientDatabase struct {
	rez.Database
	transientErr error
}

func (d transientDatabase) IsTransientError(err error) bool {
	return errors.Is(err, d.transientErr)
}

type transientDatabasePipelineProjector struct {
	err error
}

func (p *transientDatabasePipelineProjector) GetEventProjectorFunc(*ent.NormalizedEvent) (rez.EventProjectorFunc, bool) {
	return func(ctx context.Context, ev *ent.NormalizedEvent) ([]rez.ProjectedEntityRef, error) {
		return nil, p.err
	}, true
}
