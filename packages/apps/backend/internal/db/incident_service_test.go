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
	incsev "github.com/rezible/rezible/ent/incidentseverity"
	inctype "github.com/rezible/rezible/ent/incidenttype"
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
