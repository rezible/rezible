package db

import (
	"testing"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	ifo "github.com/rezible/rezible/ent/incidentfieldoption"
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

	jobs := mocks.NewMockJobsService(s.T())
	users := mocks.NewMockUserService(s.T())

	svc, err := NewIncidentService(s.Client(), jobs, msgs, users)
	s.Require().NoError(err)
	return svc
}

func (s *IncidentServiceSuite) TestCreateIncidentWithMetadataRoundTrips() {
	ctx := s.SeedTenantContext()
	svc := s.newService()

	severity, err := s.Client().IncidentSeverity.Create().
		SetName("SEV-1").
		SetRank(1).
		SetDescription("Critical").
		Save(ctx)
	s.Require().NoError(err)

	incidentType, err := s.Client().IncidentType.Create().
		SetName("Customer Impact").
		Save(ctx)
	s.Require().NoError(err)

	tag, err := s.Client().IncidentTag.Create().
		SetKey("service").
		SetValue("api").
		Save(ctx)
	s.Require().NoError(err)

	field, err := s.Client().IncidentField.Create().
		SetName("Environment").
		Save(ctx)
	s.Require().NoError(err)

	option, err := s.Client().IncidentFieldOption.Create().
		SetIncidentFieldID(field.ID).
		SetType(ifo.TypeCustom).
		SetValue("production").
		Save(ctx)
	s.Require().NoError(err)

	summary := "Customer requests are failing"
	created, err := svc.Set(ctx, uuid.Nil, func(m *ent.IncidentMutation) []ent.Mutation {
		m.SetTitle("API outage")
		m.SetSummary(summary)
		m.SetSeverityID(severity.ID)
		m.SetTypeID(incidentType.ID)
		m.AddTagAssignmentIDs(tag.ID)
		m.AddFieldSelectionIDs(option.ID)
		return nil
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

func (s *IncidentServiceSuite) TestCreateIncidentWithInvalidMetadataFails() {
	ctx := s.SeedTenantContext()
	svc := s.newService()

	_, err := svc.Set(ctx, uuid.Nil, func(m *ent.IncidentMutation) []ent.Mutation {
		m.SetTitle("Broken create")
		m.SetSeverityID(uuid.New())
		m.SetTypeID(uuid.New())
		return nil
	})
	s.Require().Error(err)
}
