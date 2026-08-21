package eventprojection

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	incsev "github.com/rezible/rezible/ent/incidentseverity"
	inctype "github.com/rezible/rezible/ent/incidenttype"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/internal/db"
	"github.com/rezible/rezible/pkg/projections"
	"github.com/rezible/rezible/test/mocks"
)

func (s *ProjectionServiceSuite) incidentService(tdb rez.Database, events *[]rez.EventOnIncidentUpdated) rez.IncidentService {
	messageService := mocks.NewMockMessageService(s.T())
	messageService.EXPECT().AddHandlers(mock.Anything).Return(nil).Once()
	messageService.EXPECT().
		Publish(mock.Anything, mock.Anything).
		Run(func(_ context.Context, event any) {
			if updated, ok := event.(rez.EventOnIncidentUpdated); ok {
				*events = append(*events, updated)
			}
		}).
		Return(nil).
		Maybe()
	service, err := db.NewIncidentService(tdb, messageService)
	s.Require().NoError(err)
	return service
}

func (s *ProjectionServiceSuite) createIncidentProjectionEvent(tdb rez.Database, subjectRef string, occurredAt time.Time, attrs projections.IncidentSubjectAttributes) *ent.NormalizedEvent {
	ctx := s.SeedTenantContext()
	encoded, err := projections.EncodeAttributes(attrs)
	s.Require().NoError(err)
	event, err := tdb.Client(ctx).NormalizedEvent.Create().
		SetProvider("test").
		SetProviderSource("incidents").
		SetProviderEventRef("incident-event-" + uuid.NewString()).
		SetProviderSubjectRef(subjectRef).
		SetKind(ne.KindObserved).
		SetSubjectKind(projections.SubjectKindIncident.String()).
		SetOccurredAt(occurredAt).
		SetReceivedAt(occurredAt).
		SetAttributes(encoded).
		Save(ctx)
	s.Require().NoError(err)
	return event
}

func (s *ProjectionServiceSuite) TestIncidentProjectionPublishesCreateChangeAndSkipsIdenticalRepeat() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()

	projector := s.projectionService(tdb)
	var events []rez.EventOnIncidentUpdated
	projector.incidents = s.incidentService(tdb, &events)

	openedAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	attrs := projections.IncidentSubjectAttributes{
		Title:       "Search outage",
		Summary:     "Search requests are failing.",
		SeverityRef: "SEV-1",
		TypeRef:     "Customer Impact",
		ExternalRef: "foo-bar-2",
		OpenedAt:    openedAt,
	}
	first := s.createIncidentProjectionEvent(tdb, "incident-1", openedAt, attrs)

	_, projErr := runProjection(ctx, projector, first)
	s.Require().NoError(projErr)
	s.Require().Len(events, 1)
	s.True(events[0].Created)

	created, err := tdb.Client(ctx).Incident.Query().
		Where(incident.Title(attrs.Title)).
		Only(ctx)
	s.Require().NoError(err)
	s.True(created.OpenedAt.Equal(openedAt))
	s.Contains(created.Slug, "260601-")

	_, projErr = runProjection(ctx, projector, first)
	s.Require().NoError(projErr)
	s.Len(events, 1)

	attrs.Title = "Search outage updated"
	second := s.createIncidentProjectionEvent(tdb, "incident-1", openedAt.Add(time.Minute), attrs)

	_, projSecondErr := runProjection(ctx, projector, second)
	s.Require().NoError(projSecondErr)
	s.Require().Len(events, 2)
	s.False(events[1].Created)

	severityCount, err := tdb.Client(ctx).IncidentSeverity.Query().
		Where(incsev.Name("SEV-1")).
		Count(ctx)
	s.Require().NoError(err)
	s.Equal(1, severityCount)

	typeCount, err := tdb.Client(ctx).IncidentType.Query().
		Where(inctype.Name("Customer Impact")).
		Count(ctx)
	s.Require().NoError(err)
	s.Equal(1, typeCount)
}

func (s *ProjectionServiceSuite) TestIncidentProjectionDoesNotPanicForDemoCatalogSearchEntity() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	client := tdb.Client(ctx)

	projector := s.projectionService(tdb)
	var events []rez.EventOnIncidentUpdated
	projector.incidents = s.incidentService(tdb, &events)

	eventID := uuid.MustParse("d1be3113-c03a-45f0-adcb-1191041c3b02")
	createdAt := time.Date(2026, 6, 19, 10, 4, 46, 429693000, time.UTC)
	occurredAt := time.Date(2026, 4, 18, 2, 30, 0, 0, time.UTC)
	attrs, attrsErr := json.Marshal(projections.IncidentSubjectAttributes{
		ExternalRef: "foo-bar",
		Title:       "Catalog search returning stale results",
		Summary:     "The catalog search index failed to refresh after the nightly product import.",
		SeverityRef: "SEV-2",
		TypeRef:     "Data Freshness",
		OpenedAt:    occurredAt,
	})
	s.Require().NoError(attrsErr)
	createEvent := client.NormalizedEvent.Create().
		SetID(eventID).
		SetKind(ne.KindObserved).
		SetProvider("demo").
		SetProviderSource("incidents").
		SetProviderEventRef("demo:incidents:catalog-search-stale-results-observed").
		SetProviderSubjectRef("demo:incident:catalog-search-stale-results").
		SetSubjectKind(projections.SubjectKindIncident.String()).
		SetAttributes(attrs).
		SetCreatedAt(createdAt).
		SetOccurredAt(occurredAt).
		SetReceivedAt(occurredAt)

	ev, err := createEvent.Save(ctx)
	s.Require().NoError(err)

	var projectionErr error
	s.Require().NotPanics(func() {
		_, projectionErr = runProjection(ctx, projector, ev)
	})
	s.Require().NoError(projectionErr)

	created, err := client.Incident.Query().
		Where(incident.Title("Catalog search returning stale results")).
		Only(ctx)
	s.Require().NoError(err)
	s.True(created.OpenedAt.Equal(occurredAt))
}
