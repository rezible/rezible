package db

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	ifo "github.com/rezible/rezible/ent/incidentfieldoption"
	ii "github.com/rezible/rezible/ent/incidentimpact"
	incsev "github.com/rezible/rezible/ent/incidentseverity"
	inctype "github.com/rezible/rezible/ent/incidenttype"
	knev "github.com/rezible/rezible/ent/knowledgeevidence"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/projections"
	"github.com/rezible/rezible/testkit"
	"github.com/rezible/rezible/testkit/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type IncidentServiceSuite struct {
	testkit.Suite
}

func TestIncidentServiceSuite(t *testing.T) {
	suite.Run(t, &IncidentServiceSuite{Suite: testkit.NewSuite()})
}

func (s *IncidentServiceSuite) newService() *IncidentService {
	msgs := mocks.NewMockMessageService(s.T())
	msgs.EXPECT().AddEventHandlers(mock.Anything).Return(nil)
	msgs.EXPECT().PublishEvent(mock.Anything, mock.Anything).Return(nil).Maybe()

	svc, err := NewIncidentService(s.Database(), msgs, nil)
	s.Require().NoError(err)
	return svc
}

func (s *IncidentServiceSuite) newServiceCapturingEvents(events *[]rez.EventOnIncidentUpdated) *IncidentService {
	msgs := mocks.NewMockMessageService(s.T())
	msgs.EXPECT().AddEventHandlers(mock.Anything).Return(nil)
	msgs.EXPECT().
		PublishEvent(mock.Anything, mock.Anything).
		Run(func(_ context.Context, event any) {
			if updated, ok := event.(rez.EventOnIncidentUpdated); ok {
				*events = append(*events, updated)
			}
		}).
		Return(nil).
		Maybe()

	svc, err := NewIncidentService(s.Database(), msgs, NewKnowledgeService(s.Database()))
	s.Require().NoError(err)
	return svc
}

func (s *IncidentServiceSuite) createIncidentProjectionEvent(subjectRef string, occurredAt time.Time, attrs projections.IncidentSubjectAttributes) *ent.NormalizedEvent {
	ctx := s.SeedTenantContext()
	encoded, err := projections.EncodeAttributes(attrs)
	s.Require().NoError(err)
	ev, err := s.Client(ctx).NormalizedEvent.Create().
		SetProvider("test").
		SetProviderSource("incidents").
		SetProviderEventRef("incident-event-" + uuid.NewString()).
		SetProviderSubjectRef(subjectRef).
		SetActivityKind(ne.ActivityKindObserved).
		SetSubjectKind(projections.SubjectKindIncident.String()).
		SetOccurredAt(occurredAt).
		SetReceivedAt(occurredAt).
		SetAttributes(encoded).
		Save(ctx)
	s.Require().NoError(err)
	return ev
}

func (s *IncidentServiceSuite) createBasicIncident(ctx context.Context, svc *IncidentService, title string) *ent.Incident {
	client := s.Client(ctx)
	severity, err := client.IncidentSeverity.Create().
		SetName("SEV-" + uuid.NewString()).
		SetRank(1).
		Save(ctx)
	s.Require().NoError(err)
	incidentType, err := client.IncidentType.Create().
		SetName("Type-" + uuid.NewString()).
		Save(ctx)
	s.Require().NoError(err)
	inc, err := svc.Set(ctx, uuid.Nil, func(m *ent.IncidentMutation) {
		m.SetTitle(title)
		m.SetSeverityID(severity.ID)
		m.SetTypeID(incidentType.ID)
	})
	s.Require().NoError(err)
	return inc
}

func (s *IncidentServiceSuite) TestCreateIncidentWithMetadataRoundTrips() {
	ctx := s.SeedTenantContext()
	svc := s.newService()

	client := s.Database().Client(ctx)

	severity, err := client.IncidentSeverity.Create().
		SetName("SEV-1").
		SetRank(1).
		SetDescription("Critical").
		Save(ctx)
	s.Require().NoError(err)

	incidentType, err := client.IncidentType.Create().
		SetName("Customer Impact").
		Save(ctx)
	s.Require().NoError(err)

	tag, err := client.IncidentTag.Create().
		SetKey("service").
		SetValue("api").
		Save(ctx)
	s.Require().NoError(err)

	field, err := client.IncidentField.Create().
		SetName("Environment").
		Save(ctx)
	s.Require().NoError(err)

	option, err := client.IncidentFieldOption.Create().
		SetIncidentFieldID(field.ID).
		SetType(ifo.TypeCustom).
		SetValue("production").
		Save(ctx)
	s.Require().NoError(err)

	summary := "Customer requests are failing"
	created, err := svc.Set(ctx, uuid.Nil, func(m *ent.IncidentMutation) {
		m.SetTitle("API outage")
		m.SetSummary(summary)
		m.SetSeverityID(severity.ID)
		m.SetTypeID(incidentType.ID)
		m.AddTagAssignmentIDs(tag.ID)
		m.AddFieldSelectionIDs(option.ID)
	})
	s.Require().NoError(err)

	s.Equal("API outage", created.Title)
	s.Equal(summary, created.Summary)
	s.Equal(severity.ID, created.SeverityID)
	s.Equal(incidentType.ID, created.TypeID)

	loaded, err := svc.Get(ctx, incident.ID(created.ID))
	s.Require().NoError(err)
	s.Require().Len(loaded.Edges.TagAssignments, 1)
	s.Equal(tag.ID, loaded.Edges.TagAssignments[0].ID)
	s.Require().Len(loaded.Edges.FieldSelections, 1)
	s.Equal(option.ID, loaded.Edges.FieldSelections[0].ID)
	s.Require().NotNil(loaded.Edges.FieldSelections[0].Edges.IncidentField)
	s.Equal(field.ID, loaded.Edges.FieldSelections[0].Edges.IncidentField.ID)

	metadata, err := svc.GetIncidentMetadata(ctx)
	s.Require().NoError(err)
	s.Require().Len(metadata.Tags, 1)
	s.Equal(tag.ID, metadata.Tags[0].ID)
	s.Require().Len(metadata.Fields, 1)
	s.Require().Len(metadata.Fields[0].Edges.Options, 1)
	s.Equal(option.ID, metadata.Fields[0].Edges.Options[0].ID)
}

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
		OpenedAt:    openedAt,
	}
	first := s.createIncidentProjectionEvent("incident-1", openedAt, attrs)

	s.Require().NoError(svc.HandleEventProjection(ctx, first))
	s.Require().Len(events, 1)
	s.True(events[0].Created)

	created, err := s.Client(ctx).Incident.Query().
		Where(incident.Title(attrs.Title)).
		Only(ctx)
	s.Require().NoError(err)
	s.True(created.OpenedAt.Equal(openedAt))
	s.Contains(created.Slug, "260601-")

	s.Require().NoError(svc.HandleEventProjection(ctx, first))
	s.Len(events, 1)

	attrs.Title = "Search outage updated"
	second := s.createIncidentProjectionEvent("incident-1", openedAt.Add(time.Minute), attrs)
	s.Require().NoError(svc.HandleEventProjection(ctx, second))
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

func (s *IncidentServiceSuite) TestSetIncidentImpactsSelectOrCreateAndReplacesLinks() {
	ctx := s.SeedTenantContext()
	svc := s.newService()
	inc := s.createBasicIncident(ctx, svc, "Checkout outage")

	impacts, err := svc.SetIncidentImpacts(ctx, inc.ID, []rez.IncidentImpactInput{
		{
			Kind:        "functionality",
			DisplayName: "order_checkout",
			Description: "Customers cannot complete checkout.",
			Source:      "responder",
			Note:        "Reported by support.",
		},
	})
	s.Require().NoError(err)
	s.Require().Len(impacts, 1)
	s.Equal("responder", impacts[0].Source)
	s.Equal("order_checkout", impacts[0].Edges.KnowledgeEntity.DisplayName)

	serviceEntity, err := s.Client(ctx).KnowledgeEntity.Create().
		SetKind("service").
		SetDisplayName("billing").
		Save(ctx)
	s.Require().NoError(err)

	impacts, err = svc.SetIncidentImpacts(ctx, inc.ID, []rez.IncidentImpactInput{
		{
			KnowledgeEntityID: serviceEntity.ID,
			Source:            "responder",
		},
	})
	s.Require().NoError(err)
	s.Require().Len(impacts, 1)
	s.Equal(serviceEntity.ID, impacts[0].KnowledgeEntityID)

	count, err := s.Client(ctx).IncidentImpact.Query().
		Where(ii.IncidentID(inc.ID)).
		Count(ctx)
	s.Require().NoError(err)
	s.Equal(1, count)
}

func (s *IncidentServiceSuite) TestIncidentImpactProjectionLinksProjectedIncidentToEntity() {
	ctx := s.SeedTenantContext()
	var events []rez.EventOnIncidentUpdated
	svc := s.newServiceCapturingEvents(&events)
	openedAt := time.Date(2026, 5, 12, 9, 35, 0, 0, time.UTC)
	incidentEvent := s.createIncidentProjectionEvent("demo:incident:checkout-search-timeouts", openedAt, projections.IncidentSubjectAttributes{
		Title:       "Checkout search lookups timing out",
		Summary:     "Checkout requests are timing out.",
		SeverityRef: "SEV-1",
		TypeRef:     "Customer Impact",
		OpenedAt:    openedAt,
	})
	s.Require().NoError(svc.HandleEventProjection(ctx, incidentEvent))

	encoded, err := projections.EncodeAttributes(projections.IncidentImpactSubjectAttributes{
		IncidentExternalRef: "demo:incident:checkout-search-timeouts",
		EntityExternalRef:   "demo:component:search_api",
		EntityKind:          "service",
		EntityDisplayName:   "Search API",
		Source:              "demo",
		Note:                "Search API blocks checkout enrichment.",
	})
	s.Require().NoError(err)
	impactEvent, err := s.Client(ctx).NormalizedEvent.Create().
		SetProvider("test").
		SetProviderSource("incident_impacts").
		SetProviderEventRef("impact-event-" + uuid.NewString()).
		SetProviderSubjectRef("demo:incident_impact:checkout-search").
		SetActivityKind(ne.ActivityKindObserved).
		SetSubjectKind(projections.SubjectKindIncidentImpact.String()).
		SetOccurredAt(openedAt.Add(time.Minute)).
		SetReceivedAt(openedAt.Add(time.Minute)).
		SetAttributes(encoded).
		Save(ctx)
	s.Require().NoError(err)

	s.Require().NoError(svc.HandleEventProjection(ctx, impactEvent))
	s.Require().NoError(svc.HandleEventProjection(ctx, impactEvent))

	impacts, err := s.Client(ctx).IncidentImpact.Query().
		WithKnowledgeEntity().
		All(ctx)
	s.Require().NoError(err)
	s.Require().Len(impacts, 1)
	s.Equal("demo", impacts[0].Source)
	s.Equal("Search API", impacts[0].Edges.KnowledgeEntity.DisplayName)
}

func (s *IncidentServiceSuite) TestIncidentImpactProjectionMissingIncidentAliasIsRetryable() {
	ctx := s.SeedTenantContext()
	var events []rez.EventOnIncidentUpdated
	svc := s.newServiceCapturingEvents(&events)
	openedAt := time.Date(2026, 5, 12, 9, 35, 0, 0, time.UTC)

	encoded, err := projections.EncodeAttributes(projections.IncidentImpactSubjectAttributes{
		IncidentExternalRef: "demo:incident:missing",
		EntityExternalRef:   "demo:component:search_api",
		EntityKind:          "service",
		EntityDisplayName:   "Search API",
		Source:              "demo",
		Note:                "Search API blocks checkout enrichment.",
	})
	s.Require().NoError(err)
	impactEvent, err := s.Client(ctx).NormalizedEvent.Create().
		SetProvider("demo").
		SetProviderSource("incident_impacts").
		SetProviderEventRef("impact-event-" + uuid.NewString()).
		SetProviderSubjectRef("demo:incident_impact:checkout-search").
		SetActivityKind(ne.ActivityKindObserved).
		SetSubjectKind(projections.SubjectKindIncidentImpact.String()).
		SetOccurredAt(openedAt.Add(time.Minute)).
		SetReceivedAt(openedAt.Add(time.Minute)).
		SetAttributes(encoded).
		Save(ctx)
	s.Require().NoError(err)

	err = svc.HandleEventProjection(ctx, impactEvent)
	s.Require().Error(err)
	s.True(projections.IsRetryable(err))
	s.ErrorContains(err, "incident entity alias not found: demo:incident:missing")
}

func (s *IncidentServiceSuite) TestContextPackInfersImpactFromRecentAlertEvidenceAndFunctionalityDependencies() {
	ctx := s.SeedTenantContext()
	svc := s.newService()
	inc := s.createBasicIncident(ctx, svc, "Checkout failures")
	now := time.Now().UTC()

	serviceEntity, err := s.Client(ctx).KnowledgeEntity.Create().
		SetKind("service").
		SetDisplayName("billing").
		Save(ctx)
	s.Require().NoError(err)
	functionalityEntity, err := s.Client(ctx).KnowledgeEntity.Create().
		SetKind("functionality").
		SetDisplayName("order_checkout").
		Save(ctx)
	s.Require().NoError(err)
	alertEntity, err := s.Client(ctx).KnowledgeEntity.Create().
		SetKind(knowledgeEntityKindAlert).
		SetDisplayName("Billing errors high").
		Save(ctx)
	s.Require().NoError(err)
	_, err = s.Client(ctx).Alert.Create().
		SetKnowledgeEntityID(alertEntity.ID).
		SetTitle("Billing errors high").
		SetDescription("5xx rate above threshold").
		Save(ctx)
	s.Require().NoError(err)
	_, err = s.Client(ctx).KnowledgeRelationship.Create().
		SetSourceEntityID(alertEntity.ID).
		SetTargetEntityID(serviceEntity.ID).
		SetKind("alerts_component").
		SetDisplayName("alert targets service").
		SetLastObservedAt(now).
		Save(ctx)
	s.Require().NoError(err)
	_, err = s.Client(ctx).KnowledgeRelationship.Create().
		SetSourceEntityID(functionalityEntity.ID).
		SetTargetEntityID(serviceEntity.ID).
		SetKind(relationshipFunctionalityDependsOnComponent).
		SetDisplayName("checkout depends on billing").
		SetLastObservedAt(now).
		Save(ctx)
	s.Require().NoError(err)

	event, err := s.Client(ctx).NormalizedEvent.Create().
		SetProvider("test").
		SetProviderSource("alerts").
		SetProviderEventRef("alert-" + uuid.NewString()).
		SetProviderSubjectRef("billing-errors").
		SetActivityKind(ne.ActivityKindObserved).
		SetSubjectKind(projections.SubjectKindAlert.String()).
		SetOccurredAt(now).
		SetReceivedAt(now).
		SetAttributes(map[string]any{}).
		Save(ctx)
	s.Require().NoError(err)
	_, err = s.Client(ctx).KnowledgeEvidence.Create().
		SetSubjectType(knev.SubjectTypeEntity).
		SetEntityID(alertEntity.ID).
		SetEventID(event.ID).
		SetAssertion(assertionAlertDefinitionObserved).
		SetEvidenceKind(knev.EvidenceKindObserved).
		SetObservedAt(now).
		Save(ctx)
	s.Require().NoError(err)

	artifacts, err := svc.GetIncidentContextArtifacts(ctx, inc.ID)
	s.Require().NoError(err)
	activeAlerts := make([]rez.AgentCaseArtifactInput, 0)
	byID := map[string]rez.AgentCaseArtifactInput{}
	for _, artifact := range artifacts {
		if artifact.Role == "active_alert" {
			activeAlerts = append(activeAlerts, artifact)
		}
		if artifact.Role == "inferred_impact" {
			if entityID, ok := artifact.Payload["entityId"].(string); ok {
				byID[entityID] = artifact
			}
		}
	}
	s.Require().Len(activeAlerts, 1)
	s.Equal("Billing errors high", activeAlerts[0].Name)
	s.Contains(byID, serviceEntity.ID.String())
	s.Contains(byID, functionalityEntity.ID.String())
	s.Contains(byID[serviceEntity.ID.String()].Payload["reason"], "recent_alert_relationship")
	s.Contains(byID[functionalityEntity.ID.String()].Payload["reason"], "functionality_dependency")
}
