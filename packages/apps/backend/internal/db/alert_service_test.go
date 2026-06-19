package db

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	entalert "github.com/rezible/rezible/ent/alert"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/execution"
	"github.com/rezible/rezible/projections"
	"github.com/rezible/rezible/testkit"
	"github.com/stretchr/testify/suite"
)

type AlertServiceSuite struct {
	testkit.Suite
}

func TestAlertServiceSuite(t *testing.T) {
	suite.Run(t, &AlertServiceSuite{Suite: testkit.NewSuite()})
}

func (s *AlertServiceSuite) createKnowledgeEntity(kind, name string) *ent.KnowledgeEntity {
	ctx := s.SeedTenantContext()
	entity, err := s.Client(ctx).KnowledgeEntity.Create().
		SetKind(kind).
		SetDisplayName(name).
		Save(ctx)
	s.Require().NoError(err)
	return entity
}

func (s *AlertServiceSuite) createAlertForEntity(entity *ent.KnowledgeEntity, title string) *ent.Alert {
	ctx := s.SeedTenantContext()
	alert, err := s.Client(ctx).Alert.Create().
		SetKnowledgeEntityID(entity.ID).
		SetTitle(title).
		SetDescription("description").
		Save(ctx)
	s.Require().NoError(err)
	return alert
}

func (s *AlertServiceSuite) createAlertProjectionEvent(subjectRef string, attrs projections.AlertSubjectAttributes) *ent.NormalizedEvent {
	ctx := s.SeedTenantContext()
	encoded, err := projections.EncodeAttributes(attrs)
	s.Require().NoError(err)
	occurredAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	ev, err := s.Client(ctx).NormalizedEvent.Create().
		SetProvider("test").
		SetProviderSource("alerts").
		SetProviderEventRef("alert-event-" + uuid.NewString()).
		SetProviderSubjectRef(subjectRef).
		SetActivityKind(ne.ActivityKindObserved).
		SetSubjectKind(projections.SubjectKindAlert.String()).
		SetOccurredAt(occurredAt).
		SetReceivedAt(occurredAt).
		SetAttributes(encoded).
		Save(ctx)
	s.Require().NoError(err)
	return ev
}

func (s *AlertServiceSuite) TestAlertProjectionCreatesUpdatesAndRecordsEvidence() {
	ctx := s.SeedTenantContext()
	svc, err := NewAlertService(s.Database(), NewKnowledgeService(s.Database()))
	s.Require().NoError(err)
	attrs := projections.AlertSubjectAttributes{
		Title:       "Search latency high",
		Description: "p95 latency above threshold",
		Definition:  "latency > 2000",
	}
	first := s.createAlertProjectionEvent("alert-1", attrs)

	s.Require().NoError(svc.HandleEventProjection(ctx, first))
	s.Require().NoError(svc.HandleEventProjection(ctx, first))

	alerts, err := s.Client(ctx).Alert.Query().All(ctx)
	s.Require().NoError(err)
	s.Require().Len(alerts, 1)
	s.Equal(attrs.Title, alerts[0].Title)
	s.NotNil(alerts[0].KnowledgeEntityID)

	attrs.Title = "Search latency critical"
	second := s.createAlertProjectionEvent("alert-1", attrs)
	s.Require().NoError(svc.HandleEventProjection(ctx, second))

	updated, err := s.Client(ctx).Alert.Query().
		Where(entalert.KnowledgeEntityID(*alerts[0].KnowledgeEntityID)).
		Only(ctx)
	s.Require().NoError(err)
	s.Equal("Search latency critical", updated.Title)

	evidenceCount, err := s.Client(ctx).KnowledgeEvidence.Query().Count(ctx)
	s.Require().NoError(err)
	s.Equal(2, evidenceCount)
}

func (s *AlertServiceSuite) TestAlertProjectionLinksRelatedEntities() {
	ctx := s.SeedTenantContext()
	svc, err := NewAlertService(s.Database(), NewKnowledgeService(s.Database()))
	s.Require().NoError(err)

	attrs := projections.AlertSubjectAttributes{
		Title:       "Search latency high",
		Description: "p95 latency above threshold",
		Definition:  "latency > 2000",
		RelatedEntities: []projections.RelatedEntityRef{
			{
				ExternalRef: "demo:component:search_api",
				Kind:        "service",
				DisplayName: "Search API",
			},
		},
	}
	ev := s.createAlertProjectionEvent("demo:alert:search-api-latency", attrs)

	s.Require().NoError(svc.HandleEventProjection(ctx, ev))

	relationships, err := s.Client(ctx).KnowledgeRelationship.Query().
		Where().
		WithSourceEntity().
		WithTargetEntity().
		All(ctx)
	s.Require().NoError(err)
	s.Require().Len(relationships, 1)
	s.Equal(relationshipKindRelatedTo, relationships[0].Kind)
	s.Equal(knowledgeEntityKindAlert, relationships[0].Edges.SourceEntity.Kind)
	s.Equal("Search API", relationships[0].Edges.TargetEntity.DisplayName)
}

func (s *AlertServiceSuite) TestGetActiveAlertsForComponentsReturnsGraphCorrelatedAlerts() {
	ctx := s.SeedTenantContext()
	svc := &AlertService{db: s.Database()}
	component := s.createKnowledgeEntity("service", "Search API")
	alertEntity := s.createKnowledgeEntity(knowledgeEntityKindAlert, "Search latency alert")
	alert := s.createAlertForEntity(alertEntity, "Search latency high")
	unrelatedEntity := s.createKnowledgeEntity(knowledgeEntityKindAlert, "Unrelated alert")
	s.createAlertForEntity(unrelatedEntity, "Unrelated")

	_, err := s.Client(ctx).KnowledgeRelationship.Create().
		SetSourceEntityID(alertEntity.ID).
		SetTargetEntityID(component.ID).
		SetKind("alerts_component").
		SetDisplayName("alert targets component").
		Save(ctx)
	s.Require().NoError(err)

	alerts, err := svc.GetActiveAlertsForComponents(ctx, []uuid.UUID{component.ID})
	s.Require().NoError(err)
	s.Require().Len(alerts, 1)
	s.Equal(alert.ID, alerts[0].ID)
}

func (s *AlertServiceSuite) TestGetActiveAlertsForComponentsExcludesUnrelatedAndOtherTenantAlerts() {
	ctx := s.SeedTenantContext()
	svc := &AlertService{db: s.Database()}
	component := s.createKnowledgeEntity("service", "Checkout API")
	otherTenant := s.SystemContext()
	tenant, err := s.Client(otherTenant).Tenant.Create().Save(otherTenant)
	s.Require().NoError(err)
	otherTenant = execution.NewTenantContext(otherTenant, tenant.ID)
	otherComponent, err := s.Client(otherTenant).KnowledgeEntity.Create().
		SetKind("service").
		SetDisplayName("Checkout API").
		Save(otherTenant)
	s.Require().NoError(err)
	otherAlertEntity, err := s.Client(otherTenant).KnowledgeEntity.Create().
		SetKind(knowledgeEntityKindAlert).
		SetDisplayName("Other tenant alert").
		Save(otherTenant)
	s.Require().NoError(err)
	_, err = s.Client(otherTenant).Alert.Create().
		SetKnowledgeEntityID(otherAlertEntity.ID).
		SetTitle("Other tenant alert").
		Save(otherTenant)
	s.Require().NoError(err)
	_, err = s.Client(otherTenant).KnowledgeRelationship.Create().
		SetSourceEntityID(otherAlertEntity.ID).
		SetTargetEntityID(otherComponent.ID).
		SetKind("alerts_component").
		Save(otherTenant)
	s.Require().NoError(err)

	alerts, err := svc.GetActiveAlertsForComponents(ctx, []uuid.UUID{component.ID})
	s.Require().NoError(err)
	s.Empty(alerts)
}

func (s *AlertServiceSuite) TestGetActiveAlertsForComponentsReturnsEmptyForNoComponents() {
	ctx := s.SeedTenantContext()
	svc := &AlertService{db: s.Database()}

	alerts, err := svc.GetActiveAlertsForComponents(ctx, nil)
	s.Require().NoError(err)
	s.Empty(alerts)
}
