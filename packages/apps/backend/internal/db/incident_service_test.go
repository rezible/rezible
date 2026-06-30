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
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/pkg/projections"
	"github.com/rezible/rezible/test"
	"github.com/rezible/rezible/test/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type IncidentServiceSuite struct {
	test.Suite
}

func TestIncidentServiceSuite(t *testing.T) {
	suite.Run(t, &IncidentServiceSuite{Suite: test.NewSuite()})
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
		SetKind(ne.KindObserved).
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

	_, projErr := svc.HandleEventProjection(ctx, incidentEvent)
	s.Require().NoError(projErr)

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
		SetKind(ne.KindObserved).
		SetSubjectKind(projections.SubjectKindIncidentImpact.String()).
		SetOccurredAt(openedAt.Add(time.Minute)).
		SetReceivedAt(openedAt.Add(time.Minute)).
		SetAttributes(encoded).
		Save(ctx)
	s.Require().NoError(err)

	_, proj1Err := svc.HandleEventProjection(ctx, impactEvent)
	s.Require().NoError(proj1Err)

	_, proj2Err := svc.HandleEventProjection(ctx, impactEvent)
	s.Require().NoError(proj2Err)

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
		SetKind(ne.KindObserved).
		SetSubjectKind(projections.SubjectKindIncidentImpact.String()).
		SetOccurredAt(openedAt.Add(time.Minute)).
		SetReceivedAt(openedAt.Add(time.Minute)).
		SetAttributes(encoded).
		Save(ctx)
	s.Require().NoError(err)

	_, err = svc.HandleEventProjection(ctx, impactEvent)
	s.Require().Error(err)
	s.True(projections.IsRetryable(err))
	s.ErrorContains(err, "incident entity alias not found: demo:incident:missing")
}
