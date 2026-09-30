package db

import (
	"testing"
	"time"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/organization"
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
	tdb := s.CreateTestDatabase()
	jobs := mocks.NewMockJobService(s.T())

	orgs, serviceErr := NewOrganizationService(tdb, jobs)
	s.Require().NoError(serviceErr)
	setupAt := time.Date(2026, 6, 4, 9, 30, 0, 0, time.UTC)

	tenantCtx := s.SeedTenantContext()
	prefs, setErr := orgs.SetPreferences(tenantCtx, s.SeedOrganizationId(), func(m *ent.OrganizationPreferencesMutation) {
		m.SetInitialSetupAt(setupAt)
	})
	s.Require().NoError(setErr)

	s.Equal(setupAt, prefs.InitialSetupAt.UTC())
	s.Equal(s.SeedOrganizationId(), prefs.OrganizationID)

	loaded, queryErr := orgs.Get(tenantCtx, organization.ID(s.SeedOrganizationId()))
	s.Require().NoError(queryErr)
	s.Require().NotNil(loaded.Edges.Preferences)
	s.Equal(prefs.ID, loaded.Edges.Preferences.ID)
	s.Equal(setupAt, loaded.Edges.Preferences.InitialSetupAt.UTC())
	s.Equal(loaded.ID, loaded.Edges.Preferences.OrganizationID)
}
