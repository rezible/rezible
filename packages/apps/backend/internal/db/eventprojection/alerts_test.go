package eventprojection

import (
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"

	"github.com/rezible/rezible/ent"
	entalert "github.com/rezible/rezible/ent/alert"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/pkg/projections"
)

func (s *ProjectionServiceSuite) createAlertProjectionEvent(db rez.Database, subjectRef string, attrs projections.AlertInstanceSubjectAttributes) *ent.NormalizedEvent {
	ctx := s.SeedTenantContext()
	encoded, encodeErr := projections.EncodeAttributes(attrs)
	s.Require().NoError(encodeErr)
	occurredAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	createEvent := db.Client(ctx).NormalizedEvent.Create().
		SetProvider("test").
		SetProviderSource("alerts").
		SetProviderEventRef("alert-event-" + uuid.NewString()).
		SetProviderSubjectRef(subjectRef).
		SetKind(ne.KindObserved).
		SetSubjectKind(projections.SubjectKindAlertInstance.String()).
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
	attrs := projections.AlertInstanceSubjectAttributes{
		Title:       "Search latency high",
		Description: "p95 latency above threshold",
		Definition:  "latency > 2000",
		ExternalRef: "external-ref-" + uuid.NewString(),
	}
	first := s.createAlertProjectionEvent(tdb, "alert-1", attrs)

	_, projectErr := runProjection(ctx, service, first)
	s.Require().NoError(projectErr)

	_, projectErr = runProjection(ctx, service, first)
	s.Require().NoError(projectErr)

	alerts, err := client.Alert.Query().All(ctx)
	s.Require().NoError(err)
	s.Require().Len(alerts, 1)
	s.Equal(attrs.Title, alerts[0].Title)
	s.NotNil(alerts[0].KnowledgeEntityID)

	attrs.Title = "Search latency critical"
	second := s.createAlertProjectionEvent(tdb, "alert-1", attrs)
	_, projectErr = runProjection(ctx, service, second)
	s.Require().NoError(projectErr)

	updated, err := client.Alert.Query().
		Where(entalert.KnowledgeEntityID(*alerts[0].KnowledgeEntityID)).
		Only(ctx)
	s.Require().NoError(err)
	s.Equal("Search latency critical", updated.Title)

	evidenceCount, err := client.KnowledgeEvidence.Query().Count(ctx)
	s.Require().NoError(err)
	s.Equal(2, evidenceCount)
}
