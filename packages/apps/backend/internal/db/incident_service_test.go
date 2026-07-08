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

	svc, err := NewIncidentService(s.Database(), msgs, NewKnowledgeFactService(s.Database()))
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
