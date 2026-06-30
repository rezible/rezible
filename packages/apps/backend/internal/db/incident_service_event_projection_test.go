package db

import (
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent/incident"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/pkg/projections"
)

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
