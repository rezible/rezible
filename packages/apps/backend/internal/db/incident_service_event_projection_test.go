package db

import (
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent/incident"
	incsev "github.com/rezible/rezible/ent/incidentseverity"
	inctype "github.com/rezible/rezible/ent/incidenttype"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/pkg/projections"
)

func (s *IncidentServiceSuite) TestIncidentProjectionPublishesCreateChangeAndSkipsIdenticalRepeat() {
	ctx := s.SeedTenantContext()
	var events []rez.EventOnIncidentUpdated
	svc := s.newServiceCapturingEvents(&events)
	openedAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	attrs := projections.IncidentSubjectAttributes{
		Title:       "Search outage",
		Summary:     "Search requests are failing.",
		SeverityRef: "SEV-1",
		TypeRef:     "Customer Impact",
		ExternalRef: "foo-bar-2",
		OpenedAt:    openedAt,
	}
	first := s.createIncidentProjectionEvent("incident-1", openedAt, attrs)

	_, projErr := svc.HandleEventProjection(ctx, first)
	s.Require().NoError(projErr)
	s.Require().Len(events, 1)
	s.True(events[0].Created)

	created, err := s.Client(ctx).Incident.Query().
		Where(incident.Title(attrs.Title)).
		Only(ctx)
	s.Require().NoError(err)
	s.True(created.OpenedAt.Equal(openedAt))
	s.Contains(created.Slug, "260601-")

	_, projErr = svc.HandleEventProjection(ctx, first)
	s.Require().NoError(projErr)
	s.Len(events, 1)

	attrs.Title = "Search outage updated"
	second := s.createIncidentProjectionEvent("incident-1", openedAt.Add(time.Minute), attrs)

	_, projSecondErr := svc.HandleEventProjection(ctx, second)
	s.Require().NoError(projSecondErr)
	s.Require().Len(events, 2)
	s.False(events[1].Created)

	severityCount, err := s.Client(ctx).IncidentSeverity.Query().
		Where(incsev.Name("SEV-1")).
		Count(ctx)
	s.Require().NoError(err)
	s.Equal(1, severityCount)
	typeCount, err := s.Client(ctx).IncidentType.Query().
		Where(inctype.Name("Customer Impact")).
		Count(ctx)
	s.Require().NoError(err)
	s.Equal(1, typeCount)
}

func (s *IncidentServiceSuite) TestIncidentProjectionDoesNotPanicForDemoCatalogSearchEntity() {
	ctx := s.SeedTenantContext()
	var events []rez.EventOnIncidentUpdated
	svc := s.newServiceCapturingEvents(&events)

	eventID := uuid.MustParse("d1be3113-c03a-45f0-adcb-1191041c3b02")
	createdAt := time.Date(2026, 6, 19, 10, 4, 46, 429693000, time.UTC)
	occurredAt := time.Date(2026, 4, 18, 2, 30, 0, 0, time.UTC)
	ev, err := s.Client(ctx).NormalizedEvent.Create().
		SetID(eventID).
		SetKind(ne.KindObserved).
		SetProvider("demo").
		SetProviderSource("incidents").
		SetProviderEventRef("demo:incidents:catalog-search-stale-results-observed").
		SetProviderSubjectRef("demo:incident:catalog-search-stale-results").
		SetSubjectKind(projections.SubjectKindIncident.String()).
		SetAttributes(map[string]any{
			"title":        "Catalog search returning stale results",
			"summary":      "The catalog search index failed to refresh after the nightly product import.",
			"type_ref":     "Data Freshness",
			"opened_at":    "2026-04-18T02:30:00Z",
			"severity_ref": "SEV-2",
			"external_ref": "foo-bar",
		}).
		SetCreatedAt(createdAt).
		SetOccurredAt(occurredAt).
		SetReceivedAt(occurredAt).
		Save(ctx)
	s.Require().NoError(err)

	var projectionErr error
	s.Require().NotPanics(func() {
		_, projectionErr = svc.HandleEventProjection(ctx, ev)
	})
	s.Require().NoError(projectionErr)

	created, err := s.Client(ctx).Incident.Query().
		Where(incident.Title("Catalog search returning stale results")).
		Only(ctx)
	s.Require().NoError(err)
	s.True(created.OpenedAt.Equal(occurredAt))
}
