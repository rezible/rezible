package eventprojection

import (
	"context"
	"time"

	"github.com/google/uuid"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	incsev "github.com/rezible/rezible/ent/incidentseverity"
	inctype "github.com/rezible/rezible/ent/incidenttype"
	"github.com/rezible/rezible/pkg/projections"
)

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
		SetLastOccurredAt(occurredAt).
		SetReceivedAt(occurredAt).
		SetAttributes(encoded)
	event, eventErr := createEvent.Save(ctx)
	s.Require().NoError(eventErr)

	return event
}

func (s *ProjectionServiceSuite) TestIncidentProjectionObservesCreateChangeAndRepeat() {
	ctx, tdb := s.SetupTestDatabase()

	projector := s.projectionService(tdb)

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

	queryCreated := tdb.Client(ctx).Incident.Query().
		Where(incident.Title(attrs.Title))
	created, createdErr := queryCreated.Only(ctx)
	s.Require().NoError(createdErr)
	s.True(created.OpenedAt.Equal(openedAt))
	s.Equal("INC-1", created.Reference())

	_, projErr = runProjection(ctx, projector, first)
	s.Require().NoError(projErr)

	attrs.Title = "Search outage updated"
	updatedSourceAt := sourceUpdatedAt.Add(time.Minute)
	attrs.SourceUpdatedAt = &updatedSourceAt
	second := s.createIncidentProjectionEvent(ctx, tdb, "incident-1", openedAt.Add(time.Minute), attrs)

	_, projSecondErr := runProjection(ctx, projector, second)
	s.Require().NoError(projSecondErr)

	observed, observedErr := tdb.Client(ctx).Incident.Query().Only(ctx)
	s.Require().NoError(observedErr)
	s.Equal(created.ID, observed.ID)
	s.Equal(attrs.Title, observed.Title)

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

func (s *ProjectionServiceSuite) TestIncidentProjectionKeepsProviderSeverityRank() {
	ctx, tdb := s.SetupTestDatabase()

	projector := s.projectionService(tdb)

	openedAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	attrs := projections.IncidentEventAttributes{
		Title:       "Search outage",
		SeverityRef: "SEV-2",
		TypeRef:     "Customer Impact",
		OpenedAt:    openedAt,
	}
	unranked := s.createIncidentProjectionEvent(ctx, tdb, "incident-1", openedAt, attrs)
	_, projErr := runProjection(ctx, projector, unranked)
	s.Require().NoError(projErr)

	created, createdErr := tdb.Client(ctx).IncidentSeverity.Query().Only(ctx)
	s.Require().NoError(createdErr)
	s.Equal(0, created.Rank)

	rank := 2
	attrs.SeverityRank = &rank
	ranked := s.createIncidentProjectionEvent(ctx, tdb, "incident-1", openedAt.Add(time.Minute), attrs)
	_, projRankedErr := runProjection(ctx, projector, ranked)
	s.Require().NoError(projRankedErr)

	updated, updatedErr := tdb.Client(ctx).IncidentSeverity.Query().Only(ctx)
	s.Require().NoError(updatedErr)
	s.Equal(created.ID, updated.ID)
	s.Equal(2, updated.Rank)
}
