package eventprojection

import (
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"

	"github.com/rezible/rezible/ent"
	ad "github.com/rezible/rezible/ent/alertdefinition"
	kne "github.com/rezible/rezible/ent/knowledgeentity"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	"github.com/rezible/rezible/pkg/projections"
)

func (s *ProjectionServiceSuite) createAlertProjectionEvent(db rez.Database, subjectRef string, attrs projections.AlertInstanceEventAttributes) *ent.NormalizedEvent {
	ctx := s.SeedTenantContext()
	encoded, encodeErr := projections.EncodeAttributes(attrs)
	s.Require().NoError(encodeErr)
	occurredAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	createEvent := db.Client(ctx).NormalizedEvent.Create().
		SetProvider("test").
		SetProviderNamespace("projection-tests").
		SetProviderResourceRef(subjectRef).
		SetProviderEventSource("alerts").
		SetProviderEventRef("alert-event-" + uuid.NewString()).
		SetKind(projections.KindAlertInstance).
		SetOccurredAt(occurredAt).
		SetReceivedAt(occurredAt).
		SetAttributes(encoded)
	event, saveErr := createEvent.Save(ctx)
	s.Require().NoError(saveErr)
	return event
}

func (s *ProjectionServiceSuite) TestAlertProjectionCreatesUpdatesAndRecordsEvidence() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	client := tdb.Client(ctx)
	service := s.projectionService(tdb)
	attrs := projections.AlertInstanceEventAttributes{
		Title:       "Search latency high",
		Description: "p95 latency above threshold",
		Definition:  "latency > 2000",
		ObservedEntities: []projections.EntityObservation{{
			Ref: rez.ProviderResourceRef{
				Provider:          "test",
				ProviderNamespace: "projection-tests",
				ResourceRef:       "search-api",
			},
			Category:    kne.CategoryContainer,
			Kind:        "service",
			DisplayName: "Search API",
		}},
	}
	first := s.createAlertProjectionEvent(tdb, "alert-1", attrs)

	_, projectErr := runProjection(ctx, service, first)
	s.Require().NoError(projectErr)

	_, projectErr = runProjection(ctx, service, first)
	s.Require().NoError(projectErr)

	alerts, err := client.AlertDefinition.Query().All(ctx)
	s.Require().NoError(err)
	s.Require().Len(alerts, 1)
	s.Equal(attrs.Title, alerts[0].Title)
	s.NotNil(alerts[0].KnowledgeEntityID)

	attrs.Title = "Search latency critical"
	second := s.createAlertProjectionEvent(tdb, "alert-1", attrs)
	_, projectErr = runProjection(ctx, service, second)
	s.Require().NoError(projectErr)

	updated, err := client.AlertDefinition.Query().
		Where(ad.KnowledgeEntityID(*alerts[0].KnowledgeEntityID)).
		Only(ctx)
	s.Require().NoError(err)
	s.Equal("Search latency critical", updated.Title)

	evidenceCount, err := client.KnowledgeEvidence.Query().Count(ctx)
	s.Require().NoError(err)
	s.Equal(6, evidenceCount)
	s.Equal(1, client.KnowledgeRelationship.Query().Where(knr.PredicateEQ(knr.PredicateObserves)).CountX(ctx))
}
