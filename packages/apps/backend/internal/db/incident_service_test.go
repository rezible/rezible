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
	"github.com/rezible/rezible/ent/incidentseverity"
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

func (s *IncidentServiceSuite) newService(tdb rez.Database) *IncidentService {
	msgs := mocks.NewMockMessageService(s.T())
	msgs.EXPECT().AddHandlers(mock.Anything).Return(nil)
	msgs.EXPECT().Publish(mock.Anything, mock.Anything).Return(nil).Maybe()

	svc, err := NewIncidentService(tdb, msgs, nil)
	s.Require().NoError(err)
	return svc
}

func (s *IncidentServiceSuite) createBasicIncident(ctx context.Context, client *ent.Client, svc *IncidentService, title string) *ent.Incident {
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
	tdb := s.CreateTestDatabase()
	svc := s.newService(tdb)

	client := tdb.Client(ctx)

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

func (s *IncidentServiceSuite) TestListIncidentsUsesFilteredTotalsAndStablePages() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	svc := s.newService(tdb)
	client := tdb.Client(ctx)
	openedAt := time.Now().UTC()

	incidents := make([]*ent.Incident, 2)
	for i := range incidents {
		incidents[i] = s.createBasicIncident(ctx, client, svc, "Matching incident")
		updated, updateErr := svc.Set(ctx, incidents[i].ID, func(m *ent.IncidentMutation) {
			m.SetOpenedAt(openedAt)
		})
		s.Require().NoError(updateErr)
		incidents[i] = updated
	}
	s.createBasicIncident(ctx, client, svc, "Unrelated incident")

	params := rez.ListIncidentsParams{Search: "matching", PageSize: 1}
	firstPage, firstErr := svc.ListIncidents(ctx, params)
	s.Require().NoError(firstErr)
	s.Equal(2, firstPage.Total)
	s.Equal(1, firstPage.Page)
	s.Equal(1, firstPage.PageSize)
	s.Require().Len(firstPage.Data, 1)

	params.Page = 2
	secondPage, secondErr := svc.ListIncidents(ctx, params)
	s.Require().NoError(secondErr)
	s.Equal(2, secondPage.Total)
	s.Require().Len(secondPage.Data, 1)
	s.NotEqual(firstPage.Data[0].ID, secondPage.Data[0].ID)

	repeatedFirstPage, repeatedErr := svc.ListIncidents(ctx, rez.ListIncidentsParams{
		Search: "matching", PageSize: 1,
	})
	s.Require().NoError(repeatedErr)
	s.Equal(firstPage.Data[0].ID, repeatedFirstPage.Data[0].ID)

}

func (s *IncidentServiceSuite) TestDoListQueryUsesSameArchiveContextForTotalAndPage() {
	ctx := s.SeedTenantContext()
	tdb := s.CreateTestDatabase()
	client := tdb.Client(ctx)

	active := client.IncidentSeverity.Create().
		SetName("Active").
		SetRank(1).
		SaveX(ctx)
	archived := client.IncidentSeverity.Create().
		SetName("Archived").
		SetRank(2).
		SaveX(ctx)
	client.IncidentSeverity.DeleteOneID(archived.ID).ExecX(ctx)

	activeOnly, activeErr := ent.DoListQuery[ent.IncidentSeverity, *ent.IncidentSeverityQuery](
		ctx,
		client.IncidentSeverity.Query().Order(incidentseverity.ByID()),
		ent.ListParams{},
	)
	s.Require().NoError(activeErr)
	s.Equal(1, activeOnly.Total)
	s.Require().Len(activeOnly.Data, 1)
	s.Equal(active.ID, activeOnly.Data[0].ID)

	withArchived, archivedErr := ent.DoListQuery[ent.IncidentSeverity, *ent.IncidentSeverityQuery](
		ctx,
		client.IncidentSeverity.Query().Order(incidentseverity.ByID()),
		ent.ListParams{IncludeArchived: true},
	)
	s.Require().NoError(archivedErr)
	s.Equal(2, withArchived.Total)
	s.Require().Len(withArchived.Data, 2)
}
