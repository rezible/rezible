package eventprojection

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	incsev "github.com/rezible/rezible/ent/incidentseverity"
	inctype "github.com/rezible/rezible/ent/incidenttype"
	"github.com/rezible/rezible/internal/db"
	"github.com/rezible/rezible/pkg/projections"
	"github.com/rezible/rezible/test/mocks"
)

func (s *ProjectionServiceSuite) incidentService(tdb rez.Database, events *[]rez.EventOnIncidentUpdated) rez.IncidentService {
	messageService := mocks.NewMockMessageQueue(s.T())
	messageService.EXPECT().
		Publish(mock.Anything, mock.Anything).
		Run(func(_ context.Context, event any) {
			if updated, ok := event.(rez.EventOnIncidentUpdated); ok {
				*events = append(*events, updated)
			}
		}).
		Return(nil).
		Maybe()
	retrospectives, retrospectiveServiceErr := db.NewRetrospectiveService(tdb)
	s.Require().NoError(retrospectiveServiceErr)

	service, serviceErr := db.NewIncidentService(tdb, messageService, nil, retrospectives)
	s.Require().NoError(serviceErr)

	return service
}

func (s *ProjectionServiceSuite) createIncidentProjectionEvent(ctx context.Context, tdb rez.Database, subjectRef string, occurredAt time.Time, attrs projections.IncidentEventAttributes) *ent.NormalizedEvent {
	encoded, encodedErr := projections.EncodeAttributes(attrs)
	s.Require().NoError(encodedErr)

	createEvent := tdb.Client(ctx).NormalizedEvent.Create().
		SetProvider("test").
		SetProviderNamespace("projection-tests").
		SetProviderResourceRef(subjectRef).
		SetProviderEventSource("incidents").
		SetProviderEventRef("incident-event-" + uuid.NewString()).
		SetKind(projections.KindIncident).
		SetOccurredAt(occurredAt).
		SetReceivedAt(occurredAt).
		SetAttributes(encoded)
	event, eventErr := createEvent.Save(ctx)
	s.Require().NoError(eventErr)

	return event
}

func (s *ProjectionServiceSuite) TestIncidentProjectionPublishesCreateChangeAndSkipsIdenticalRepeat() {
	ctx, tdb := s.SetupTestDatabase()

	projector := s.projectionService(tdb)
	var events []rez.EventOnIncidentUpdated
	projector.incidents = s.incidentService(tdb, &events)

	openedAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	sourceUpdatedAt := openedAt
	attrs := projections.IncidentEventAttributes{
		Title:           "Search outage",
		Summary:         "Search requests are failing.",
		SeverityRef:     "SEV-1",
		TypeRef:         "Customer Impact",
		OpenedAt:        openedAt,
		SourceUpdatedAt: &sourceUpdatedAt,
	}
	first := s.createIncidentProjectionEvent(ctx, tdb, "incident-1", openedAt, attrs)

	_, projErr := runProjection(ctx, projector, first)
	s.Require().NoError(projErr)
	s.Require().Len(events, 1)
	s.True(events[0].Created)

	queryCreated := tdb.Client(ctx).Incident.Query().
		Where(incident.Title(attrs.Title))
	created, createdErr := queryCreated.Only(ctx)
	s.Require().NoError(createdErr)
	s.True(created.OpenedAt.Equal(openedAt))
	s.Contains(created.Slug, "260601-")

	_, projErr = runProjection(ctx, projector, first)
	s.Require().NoError(projErr)
	s.Len(events, 1)

	attrs.Title = "Search outage updated"
	updatedSourceAt := sourceUpdatedAt.Add(time.Minute)
	attrs.SourceUpdatedAt = &updatedSourceAt
	second := s.createIncidentProjectionEvent(ctx, tdb, "incident-1", openedAt.Add(time.Minute), attrs)

	_, projSecondErr := runProjection(ctx, projector, second)
	s.Require().NoError(projSecondErr)
	s.Require().Len(events, 2)
	s.False(events[1].Created)

	querySeverityCount := tdb.Client(ctx).IncidentSeverity.Query().
		Where(incsev.Name("SEV-1"))
	severityCount, severityCountErr := querySeverityCount.Count(ctx)
	s.Require().NoError(severityCountErr)
	s.Equal(1, severityCount)

	queryTypeCount := tdb.Client(ctx).IncidentType.Query().
		Where(inctype.Name("Customer Impact"))
	typeCount, typeCountErr := queryTypeCount.Count(ctx)
	s.Require().NoError(typeCountErr)
	s.Equal(1, typeCount)
}
