package db

import (
	"context"
	"log/slog"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/eventannotation"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	neps "github.com/rezible/rezible/ent/normalizedeventprojectionstatus"
	"github.com/rezible/rezible/jobs"
	"github.com/rezible/rezible/projections"
	"github.com/rezible/rezible/testkit"
	"github.com/rezible/rezible/testkit/mocks"
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
	testkit.Suite
}

func TestProviderEventPipelineServiceSuite(t *testing.T) {
	suite.Run(t, &ProviderEventPipelineServiceSuite{Suite: testkit.NewSuite()})
}

func (s *ProviderEventPipelineServiceSuite) newPipelineService(jobSvc rez.JobService, projector *pipelineTestProjector) *ProviderEventPipelineService {
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
	s.Equal(ne.ActivityKindObserved, normalized.ActivityKind)
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

	queryStatus := s.Client(ctx).NormalizedEventProjectionStatus.Query().
		Where(neps.NormalizedEventID(normalized.ID))
	status, statusErr := queryStatus.Only(ctx)
	s.Require().NoError(statusErr)
	s.Equal(neps.StatusSucceeded, status.Status)
	s.Equal(reflect.TypeOf(projector).String(), status.HandlerName)
	s.NotNil(status.SucceededAt)
	s.Nil(status.FailedAt)
	s.Empty(status.LastError)
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

type pipelineTestProcessor struct{}

func (pipelineTestProcessor) ProcessProviderEvent(_ context.Context, ev rez.ProviderEvent) (ent.NormalizedEvents, error) {
	return ent.NormalizedEvents{
		{
			Provider:           ev.Provider,
			ProviderSource:     ev.ProviderSource,
			ProviderEventRef:   ev.ProviderEventRef,
			ProviderSubjectRef: ev.ProviderSubjectRef,
			ActivityKind:       ne.ActivityKindObserved,
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

func (p *pipelineTestProjector) HandleEventProjection(ctx context.Context, ev *ent.NormalizedEvent) error {
	_, err := p.db.Client(ctx).EventAnnotation.Create().
		SetEventID(ev.ID).
		SetCreatorID(p.creatorID).
		SetMinutesOccupied(5).
		SetNotes("projected " + ev.ProviderSubjectRef).
		SetTags([]string{"pipeline-test"}).
		Save(ctx)
	return err
}
