package db

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/organization"
	"github.com/rezible/rezible/ent/organizationpreferences"
	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/test"
	"github.com/stretchr/testify/suite"

	"github.com/rezible/rezible/test/mocks"
)

type OrganizationsServiceSuite struct {
	test.Suite
}

func TestOrganizationsServiceSuite(t *testing.T) {
	suite.Run(t, &OrganizationsServiceSuite{Suite: test.NewSuite()})
}

func (s *OrganizationsServiceSuite) TestSetPreferencesSetsTimestamp() {
	tenantCtx, tdb := s.SetupTestDatabase()
	organizationID, organizationErr := tdb.Client(tenantCtx).Organization.Query().OnlyID(tenantCtx)
	s.Require().NoError(organizationErr)
	jobs := mocks.NewMockJobService(s.T())

	orgs, serviceErr := NewOrganizationService(tdb, jobs)
	s.Require().NoError(serviceErr)
	setupAt := time.Date(2026, 6, 4, 9, 30, 0, 0, time.UTC)

	prefs, setErr := orgs.SetPreferences(tenantCtx, organizationID, func(m *ent.OrganizationPreferencesMutation) {
		m.SetInitialSetupAt(setupAt)
	})
	s.Require().NoError(setErr)

	s.Equal(setupAt, prefs.InitialSetupAt.UTC())
	s.Equal(organizationID, prefs.OrganizationID)

	loaded, queryErr := orgs.Get(tenantCtx, organization.ID(organizationID))
	s.Require().NoError(queryErr)
	s.Require().NotNil(loaded.Edges.Preferences)
	s.Equal(prefs.ID, loaded.Edges.Preferences.ID)
	s.Equal(setupAt, loaded.Edges.Preferences.InitialSetupAt.UTC())
	s.Equal(loaded.ID, loaded.Edges.Preferences.OrganizationID)
}

func (s *OrganizationsServiceSuite) TestTurningOffIncidentManagementNeedsNoUnresolvedIncidents() {
	tenantCtx, tdb := s.SetupTestDatabase()
	client := tdb.Client(tenantCtx)
	organizationID, organizationErr := client.Organization.Query().OnlyID(tenantCtx)
	s.Require().NoError(organizationErr)
	orgs, serviceErr := NewOrganizationService(tdb, mocks.NewMockJobService(s.T()))
	s.Require().NoError(serviceErr)

	_, enableErr := orgs.SetPreferences(tenantCtx, organizationID, func(m *ent.OrganizationPreferencesMutation) {
		m.SetEnableIncidentManagement(true)
	})
	s.Require().NoError(enableErr)

	createIncident := client.Incident.Create().
		SetSlug("unresolved-" + uuid.NewString()).
		SetTitle("Checkout errors")
	unresolved, createErr := createIncident.Save(tenantCtx)
	s.Require().NoError(createErr)

	turnOff := func(m *ent.OrganizationPreferencesMutation) {
		m.SetEnableIncidentManagement(false)
		m.SetTimezone("Australia/Perth")
	}
	_, rejectedErr := orgs.SetPreferences(tenantCtx, organizationID, turnOff)
	s.Require().Error(rejectedErr)
	s.Equal(errs.CodeConflict, errs.CodeOf(rejectedErr))
	s.Equal("Incident management can't be turned off while 1 incident is unresolved. Resolve it first.", errs.PublicMessage(rejectedErr))

	queryPrefs := client.OrganizationPreferences.Query().
		Where(organizationpreferences.OrganizationID(organizationID))
	unchanged, queryErr := queryPrefs.Only(tenantCtx)
	s.Require().NoError(queryErr)
	s.True(unchanged.EnableIncidentManagement)
	s.Empty(unchanged.Timezone)

	resolveIncident := client.Incident.UpdateOneID(unresolved.ID).
		SetResponseState(incident.ResponseStateResolved).
		SetResolvedAt(time.Now())
	s.Require().NoError(resolveIncident.Exec(tenantCtx))

	turnedOff, turnOffErr := orgs.SetPreferences(tenantCtx, organizationID, turnOff)
	s.Require().NoError(turnOffErr)
	s.False(turnedOff.EnableIncidentManagement)
	s.Equal("Australia/Perth", turnedOff.Timezone)
}
