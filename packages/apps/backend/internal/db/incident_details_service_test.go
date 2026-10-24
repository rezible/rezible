package db

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/incidentseverity"
	"github.com/rezible/rezible/test"
	"github.com/rezible/rezible/test/mocks"
)

type IncidentDetailsServiceSuite struct {
	test.Suite
}

func TestIncidentDetailsServiceSuite(t *testing.T) {
	suite.Run(t, &IncidentDetailsServiceSuite{Suite: test.NewSuite()})
}

func (s *IncidentDetailsServiceSuite) newService(tdb rez.Database) *IncidentDetailsService {
	svc, svcErr := NewIncidentDetailsService(tdb, mocks.NewMockMessageQueue(s.T()), mocks.NewMockSituationService(s.T()))
	s.Require().NoError(svcErr)
	return svc
}

func (s *IncidentDetailsServiceSuite) createIncident(ctx context.Context, client *ent.Client, title string, openedAt time.Time) *ent.Incident {
	createSeverity := client.IncidentSeverity.Create().
		SetName("SEV-" + uuid.NewString()).
		SetRank(1)
	severity, severityErr := createSeverity.Save(ctx)
	s.Require().NoError(severityErr)
	createIncident := client.Incident.Create().
		SetSlug("incident-" + uuid.NewString()).
		SetTitle(title).
		SetSeverityID(severity.ID).
		SetOpenedAt(openedAt)
	inc, incidentErr := createIncident.Save(ctx)
	s.Require().NoError(incidentErr)
	return inc
}

func (s *IncidentDetailsServiceSuite) TestListIncidentsUsesFilteredTotalsAndStablePages() {
	ctx, tdb := s.SetupTestDatabase()
	svc := s.newService(tdb)
	client := tdb.Client(ctx)
	openedAt := time.Now().UTC()

	incidents := make([]*ent.Incident, 2)
	for i := range incidents {
		incidents[i] = s.createIncident(ctx, client, "Matching incident", openedAt)
	}
	highest := client.IncidentSeverity.Create().
		SetName("SEV-0").
		SetRank(0).
		SaveX(ctx)
	lower := client.IncidentSeverity.Create().
		SetName("SEV-2").
		SetRank(2).
		SaveX(ctx)
	incidents[0] = client.Incident.UpdateOneID(incidents[0].ID).
		SetSeverityID(highest.ID).
		SetUpdatedAt(openedAt.Add(2 * time.Hour)).
		SaveX(ctx)
	incidents[1] = client.Incident.UpdateOneID(incidents[1].ID).
		SetSeverityID(lower.ID).
		SetUpdatedAt(openedAt.Add(3 * time.Hour)).
		SaveX(ctx)
	unrelated := s.createIncident(ctx, client, "Unrelated incident", openedAt)

	params := rez.ListIncidentsParams{Search: "matching", PageSize: 1}
	firstPage, firstErr := svc.ListIncidents(ctx, params)
	s.Require().NoError(firstErr)
	s.Equal(2, firstPage.Total)
	s.Equal(1, firstPage.Page)
	s.Equal(1, firstPage.PageSize)
	s.Require().Len(firstPage.Data, 1)
	s.Equal(incidents[0].ID, firstPage.Data[0].ID)

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

	// Within the same severity, latest update wins before the page is sliced.
	updateSecond := client.Incident.UpdateOneID(incidents[1].ID).
		SetSeverityID(highest.ID).
		SetUpdatedAt(openedAt.Add(3 * time.Hour))
	s.Require().NoError(updateSecond.Exec(ctx))
	latest, latestErr := svc.ListIncidents(ctx, rez.ListIncidentsParams{Search: "matching", PageSize: 1})
	s.Require().NoError(latestErr)
	s.Require().Len(latest.Data, 1)
	s.Equal(incidents[1].ID, latest.Data[0].ID)

	updateFirst := client.Incident.UpdateOneID(incidents[0].ID).
		SetUpdatedAt(openedAt.Add(3 * time.Hour))
	s.Require().NoError(updateFirst.Exec(ctx))
	tied, tiedErr := svc.ListIncidents(ctx, rez.ListIncidentsParams{Search: "matching", PageSize: 1})
	s.Require().NoError(tiedErr)
	s.Require().Len(tied.Data, 1)
	expected := incidents[0].ID
	if incidents[1].ID.String() < expected.String() {
		expected = incidents[1].ID
	}
	s.Equal(expected, tied.Data[0].ID)

	resolve := client.Incident.UpdateOneID(unrelated.ID).
		SetResponseState(incident.ResponseStateResolved).
		SetResolvedAt(openedAt)
	s.Require().NoError(resolve.Exec(ctx))
	resolved, resolvedErr := svc.ListIncidents(ctx, rez.ListIncidentsParams{Resolved: new(true)})
	s.Require().NoError(resolvedErr)
	s.Require().Len(resolved.Data, 1)
	s.Equal(unrelated.ID, resolved.Data[0].ID)
	unresolved, unresolvedErr := svc.ListIncidents(ctx, rez.ListIncidentsParams{Resolved: new(false)})
	s.Require().NoError(unresolvedErr)
	s.Equal(2, unresolved.Total)
}

func (s *IncidentDetailsServiceSuite) TestDoListQueryUsesSameArchiveContextForTotalAndPage() {
	ctx, tdb := s.SetupTestDatabase()
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
