package db

import (
	"context"
	"errors"
	"log/slog"
	"reflect"
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

func (s *ProviderEventPipelineServiceSuite) newPipelineService(jobSvc rez.JobService, projector rez.NormalizedEventProjector) *ProviderEventPipelineService {
	reg := projections.NewPipelineRegistry()
	reg.RegisterProviderEventProcessors(pipelineTestProcessor{}, pipelineTestProvider)
	reg.RegisterEventProjector(projector, pipelineTestSubjectKind)

	return &ProviderEventPipelineService{
		logger:     slog.Default(),
		db:         s.Database(),
		jobService: jobSvc,
		reg:        reg,
	}
}

func (s *ProviderEventPipelineServiceSuite) makeTestEvent() rez.ProviderEvent {
	receivedAt := time.Date(2026, 6, 4, 9, 30, 0, 0, time.UTC)
	return rez.ProviderEvent{
		Provider:           pipelineTestProvider,
		ProviderSource:     pipelineTestSource,
		ProviderEventRef:   "delivery-1",
		ProviderSubjectRef: "subject-1",
		ReceivedAt:         receivedAt,
		Payload:            []byte(`{"summary":"received"}`),
		ContentType:        "application/json",
	}
}

func (s *ProviderEventPipelineServiceSuite) makeTestUser(ctx context.Context) *ent.User {
	create := s.Client(ctx).User.Create().
		SetEmail("pipeline-test+" + uuid.NewString() + "@example.com").
		SetName("Pipeline Test User")
	user, err := create.Save(ctx)
	s.Require().NoError(err)
	return user
}

func (s *ProviderEventPipelineServiceSuite) createPipelineNormalizedEvent(ctx context.Context) *ent.NormalizedEvent {
	ev := s.makeTestEvent()
	normalized, err := s.Client(ctx).NormalizedEvent.Create().
		SetProvider(ev.Provider).
		SetProviderSource(ev.ProviderSource).
		SetProviderEventRef("normalized-" + uuid.NewString()).
		SetProviderSubjectRef(ev.ProviderSubjectRef).
		SetKind(ne.KindObserved).
		SetSubjectKind(pipelineTestSubjectKind.String()).
		SetOccurredAt(ev.ReceivedAt.Add(-time.Minute)).
		SetReceivedAt(ev.ReceivedAt).
		SetAttributes(map[string]any{"summary": "processed " + ev.ProviderSubjectRef}).
		Save(ctx)
	s.Require().NoError(err)
	return normalized
}

func (s *ProviderEventPipelineServiceSuite) TestIngestProcessAndProjectEndToEnd() {
	ctx := s.SeedTenantContext()
	client := s.Client(ctx)
	jobSvc := mocks.NewMockJobService(s.T())
	creator := s.makeTestUser(ctx)
	projector := &pipelineTestProjector{db: s.Database(), creatorID: creator.ID}
	svc := s.newPipelineService(jobSvc, projector)

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
			s.Require().NotNil(params[0].InsertOpts)
			s.True(params[0].InsertOpts.UniqueOpts.ByArgs)
			s.Equal(jobs.UniqueStateNonCompleted, params[0].InsertOpts.UniqueOpts.ByState)

			var ok bool
			capturedProjectArgs, ok = params[0].Args.(jobs.ProjectNormalizedEvent)
			s.Require().True(ok)

			return []*rivertype.JobInsertResult{{}}, nil
		}).
		Once()

	s.Require().NoError(svc.HandleProcessEventJob(ctx, capturedProcessArgs))

	queryNormalized := client.NormalizedEvent.Query().
		Where(ne.ProviderEventRef(ev.ProviderEventRef))

	normalized, normalizedErr := queryNormalized.Only(ctx)
	s.Require().NoError(normalizedErr)
	s.Equal(pipelineTestProvider, normalized.Provider)
	s.Equal(pipelineTestSource, normalized.ProviderSource)
	s.Equal(ev.ProviderSubjectRef, normalized.ProviderSubjectRef)
	s.Equal(ne.KindObserved, normalized.Kind)
	s.Equal(pipelineTestSubjectKind.String(), normalized.SubjectKind)
	s.Equal("processed subject-1", normalized.Attributes["summary"])
	s.Equal(normalized.ID, capturedProjectArgs.EventId)

	s.Require().NoError(svc.HandleEventProjectionJob(ctx, capturedProjectArgs))

	queryAnno := client.EventAnnotation.Query().
		Where(eventannotation.EventID(normalized.ID))
	anno, annoErr := queryAnno.Only(ctx)
	s.Require().NoError(annoErr)
	s.Equal(creator.ID, anno.CreatorID)
	s.Equal(5, anno.MinutesOccupied)
	s.Equal("projected subject-1", anno.Notes)

	queryProj := s.Client(ctx).NormalizedEventProjection.Query().
		Where(nep.EventID(normalized.ID))
	proj, projErr := queryProj.Only(ctx)
	s.Require().NoError(projErr)
	s.Equal(nep.StatusSucceeded, proj.Status)
	s.Equal(reflect.TypeOf(projector).String(), proj.Projector)
	s.NotNil(proj.FinishedAt)
	s.Empty(proj.Error)
}

func (s *ProviderEventPipelineServiceSuite) TestProcessProviderEventUpsertsNormalizedEvent() {
	ctx := s.SeedTenantContext()
	jobSvc := mocks.NewMockJobService(s.T())
	creator := s.makeTestUser(ctx)
	svc := s.newPipelineService(jobSvc, &pipelineTestProjector{db: s.Database(), creatorID: creator.ID})

	args := processProviderEventArgs{Event: s.makeTestEvent()}

	jobSvc.EXPECT().
		InsertMany(mock.Anything, mock.Anything).
		Return([]*rivertype.JobInsertResult{{}}, nil).
		Twice()

	s.Require().NoError(svc.HandleProcessEventJob(ctx, args))
	s.Require().NoError(svc.HandleProcessEventJob(ctx, args))

	queryCount := s.Client(ctx).NormalizedEvent.Query().
		Where(ne.ProviderEventRef(args.Event.ProviderEventRef))
	count, countErr := queryCount.Count(ctx)
	s.Require().NoError(countErr)
	s.Equal(1, count)
}

func (s *ProviderEventPipelineServiceSuite) TestProjectionSkipsSucceededStatus() {
	ctx := s.SeedTenantContext()
	projector := &countingPipelineProjector{}
	svc := s.newPipelineService(mocks.NewMockJobService(s.T()), projector)
	ev := s.createPipelineNormalizedEvent(ctx)

	_, err := s.Client(ctx).NormalizedEventProjection.Create().
		SetEventID(ev.ID).
		SetProjector(reflect.TypeOf(projector).String()).
		SetStartedAt(time.Now()).
		SetStatus(nep.StatusSucceeded).
		Save(ctx)
	s.Require().NoError(err)

	s.Require().NoError(svc.HandleEventProjectionJob(ctx, jobs.ProjectNormalizedEvent{EventId: ev.ID}))
	s.Equal(0, projector.calls)
}

func (s *ProviderEventPipelineServiceSuite) TestProjectionRetriesPendingStatus() {
	ctx := s.SeedTenantContext()
	projector := &countingPipelineProjector{}
	svc := s.newPipelineService(mocks.NewMockJobService(s.T()), projector)
	ev := s.createPipelineNormalizedEvent(ctx)

	_, err := s.Client(ctx).NormalizedEventProjection.Create().
		SetEventID(ev.ID).
		SetProjector(reflect.TypeOf(projector).String()).
		SetStartedAt(time.Now()).
		SetStatus(nep.StatusPending).
		Save(ctx)
	s.Require().NoError(err)

	s.Require().NoError(svc.HandleEventProjectionJob(ctx, jobs.ProjectNormalizedEvent{EventId: ev.ID}))
	s.Equal(1, projector.calls)

	status, err := s.Client(ctx).NormalizedEventProjection.Query().
		Where(nep.EventID(ev.ID), nep.Projector(reflect.TypeOf(projector).String())).
		Only(ctx)
	s.Require().NoError(err)
	s.Equal(nep.StatusSucceeded, status.Status)
	s.NotNil(status.StartedAt)
}

func (s *ProviderEventPipelineServiceSuite) TestProjectionFailureRollsBackProjectorWrites() {
	ctx := s.SeedTenantContext()
	creator := s.makeTestUser(ctx)
	projector := &rollbackPipelineProjector{db: s.Database(), creatorID: creator.ID}
	svc := s.newPipelineService(mocks.NewMockJobService(s.T()), projector)
	ev := s.createPipelineNormalizedEvent(ctx)

	s.Require().NoError(svc.HandleEventProjectionJob(ctx, jobs.ProjectNormalizedEvent{EventId: ev.ID}))

	proj, projErr := s.Client(ctx).NormalizedEventProjection.Query().
		Where(nep.EventID(ev.ID), nep.Projector(reflect.TypeOf(projector).String())).
		Only(ctx)
	s.Require().NoError(projErr)
	s.Require().Equal(nep.StatusFailed, proj.Status)
	s.NotEmpty(proj.Error)

	count, countErr := s.Client(ctx).EventAnnotation.Query().
		Where(eventannotation.EventID(ev.ID)).
		Count(ctx)
	s.Require().NoError(countErr)
	s.Equal(0, count)
}

func (s *ProviderEventPipelineServiceSuite) TestRetryableProjectionFailureReturnsError() {
	ctx := s.SeedTenantContext()
	projector := &retryablePipelineProjector{}
	svc := s.newPipelineService(mocks.NewMockJobService(s.T()), projector)
	ev := s.createPipelineNormalizedEvent(ctx)

	err := svc.HandleEventProjectionJob(ctx, jobs.ProjectNormalizedEvent{EventId: ev.ID})
	s.Require().Error(err)
	s.True(projections.IsRetryable(err))

	status, err := s.Client(ctx).NormalizedEventProjection.Query().
		Where(nep.EventID(ev.ID), nep.Projector(reflect.TypeOf(projector).String())).
		Only(ctx)
	s.Require().NoError(err)
	s.Equal(nep.StatusFailed, status.Status)
	s.Contains(status.Error, "dependency not ready")
}

func (s *ProviderEventPipelineServiceSuite) TestTransientDatabaseProjectionFailureReturnsRetryableError() {
	ctx := s.SeedTenantContext()
	transientErr := errors.New("database deadlock")
	projector := &transientDatabasePipelineProjector{err: transientErr}
	svc := s.newPipelineService(mocks.NewMockJobService(s.T()), projector)
	svc.db = transientDatabase{Database: s.Database(), transientErr: transientErr}
	ev := s.createPipelineNormalizedEvent(ctx)

	err := svc.HandleEventProjectionJob(ctx, jobs.ProjectNormalizedEvent{EventId: ev.ID})
	s.Require().Error(err)
	s.True(projections.IsRetryable(err))

	status, err := s.Client(ctx).NormalizedEventProjection.Query().
		Where(nep.EventID(ev.ID), nep.Projector(reflect.TypeOf(projector).String())).
		Only(ctx)
	s.Require().NoError(err)
	s.Equal(nep.StatusFailed, status.Status)
	s.Contains(status.Error, "database deadlock")
}

func (s *ProviderEventPipelineServiceSuite) TestProjectionPanicMarksStatusFailed() {
	ctx := s.SeedTenantContext()
	projector := &panicPipelineProjector{}
	svc := s.newPipelineService(mocks.NewMockJobService(s.T()), projector)
	ev := s.createPipelineNormalizedEvent(ctx)

	s.Require().NoError(svc.HandleEventProjectionJob(ctx, jobs.ProjectNormalizedEvent{EventId: ev.ID}))

	status, err := s.Client(ctx).NormalizedEventProjection.Query().
		Where(nep.EventID(ev.ID), nep.Projector(reflect.TypeOf(projector).String())).
		Only(ctx)
	s.Require().NoError(err)
	s.Equal(nep.StatusFailed, status.Status)
	s.Contains(status.Error, "projector panic: boom")
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
			Attributes: map[string]any{
				"summary": "processed " + ev.ProviderSubjectRef,
			},
		},
	}, nil
}

type pipelineTestProjector struct {
	db        rez.Database
	creatorID uuid.UUID
}

func (p *pipelineTestProjector) HandleEventProjection(ctx context.Context, ev *ent.NormalizedEvent) (map[string][]uuid.UUID, error) {
	_, err := p.db.Client(ctx).EventAnnotation.Create().
		SetEventID(ev.ID).
		SetCreatorID(p.creatorID).
		SetMinutesOccupied(5).
		SetNotes("projected " + ev.ProviderSubjectRef).
		SetTags([]string{"pipeline-test"}).
		Save(ctx)
	return nil, err
}

type countingPipelineProjector struct {
	calls int
}

func (p *countingPipelineProjector) HandleEventProjection(context.Context, *ent.NormalizedEvent) (map[string][]uuid.UUID, error) {
	p.calls++
	return nil, nil
}

type rollbackPipelineProjector struct {
	db        rez.Database
	creatorID uuid.UUID
}

func (p *rollbackPipelineProjector) HandleEventProjection(ctx context.Context, ev *ent.NormalizedEvent) (map[string][]uuid.UUID, error) {
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
}

type retryablePipelineProjector struct{}

func (p *retryablePipelineProjector) HandleEventProjection(context.Context, *ent.NormalizedEvent) (map[string][]uuid.UUID, error) {
	return nil, projections.Retryable(errors.New("dependency not ready"))
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

func (p *transientDatabasePipelineProjector) HandleEventProjection(context.Context, *ent.NormalizedEvent) (map[string][]uuid.UUID, error) {
	return nil, p.err
}

type panicPipelineProjector struct{}

func (p *panicPipelineProjector) HandleEventProjection(context.Context, *ent.NormalizedEvent) (map[string][]uuid.UUID, error) {
	panic("boom")
}
