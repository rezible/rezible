package eventprojection

import (
	"time"

	"github.com/google/uuid"

	"github.com/rezible/rezible/ent"
	entalert "github.com/rezible/rezible/ent/alert"
	knr "github.com/rezible/rezible/ent/knowledgerelationship"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/pkg/projections"
)

func (s *ProjectionServiceSuite) createAlertProjectionEvent(subjectRef string, attrs projections.AlertInstanceSubjectAttributes) *ent.NormalizedEvent {
	ctx := s.SeedTenantContext()
	encoded, encodeErr := projections.EncodeAttributes(attrs)
	s.Require().NoError(encodeErr)
	occurredAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	createEvent := s.Client(ctx).NormalizedEvent.Create().
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
	service := s.projectionService()
	attrs := projections.AlertInstanceSubjectAttributes{
		Title:       "Search latency high",
		Description: "p95 latency above threshold",
		Definition:  "latency > 2000",
		ExternalRef: "external-ref-" + uuid.NewString(),
	}
	first := s.createAlertProjectionEvent("alert-1", attrs)

	_, projectErr := runProjection(ctx, service, first)
	s.Require().NoError(projectErr)

	_, projectErr = runProjection(ctx, service, first)
	s.Require().NoError(projectErr)

	alerts, err := s.Client(ctx).Alert.Query().All(ctx)
	s.Require().NoError(err)
	s.Require().Len(alerts, 1)
	s.Equal(attrs.Title, alerts[0].Title)
	s.NotNil(alerts[0].KnowledgeEntityID)

	attrs.Title = "Search latency critical"
	second := s.createAlertProjectionEvent("alert-1", attrs)
	_, projectErr = runProjection(ctx, service, second)
	s.Require().NoError(projectErr)

	updated, err := s.Client(ctx).Alert.Query().
		Where(entalert.KnowledgeEntityID(*alerts[0].KnowledgeEntityID)).
		Only(ctx)
	s.Require().NoError(err)
	s.Equal("Search latency critical", updated.Title)

	evidenceCount, err := s.Client(ctx).KnowledgeEvidence.Query().Count(ctx)
	s.Require().NoError(err)
	s.Equal(2, evidenceCount)
}

func (s *ProjectionServiceSuite) TestAlertProjectionLinksRelatedEntities() {
	ctx := s.SeedTenantContext()
	service := s.projectionService()
	attrs := projections.AlertInstanceSubjectAttributes{
		Title:       "Search latency high",
		Description: "p95 latency above threshold",
		Definition:  "latency > 2000",
		ExternalRef: "external-ref-" + uuid.NewString(),
		RelatedEntities: []projections.RelatedEntityRef{
			{
				ExternalRef: "demo:component:search_api",
				Kind:        "service",
				DisplayName: "Search API",
			},
		},
	}
	event := s.createAlertProjectionEvent("demo:alert:search-api-latency", attrs)

	_, projectErr := runProjection(ctx, service, event)
	s.Require().NoError(projectErr)

	relationships, err := s.Client(ctx).KnowledgeRelationship.Query().
		Where(knr.Kind("alert_related_to")).
		WithSourceEntity().
		WithTargetEntity().
		All(ctx)
	s.Require().NoError(err)
	s.Require().Len(relationships, 1)
	s.Equal("alert", relationships[0].Edges.SourceEntity.Kind)
	s.Equal("Search API", relationships[0].Edges.TargetEntity.DisplayName)
}
