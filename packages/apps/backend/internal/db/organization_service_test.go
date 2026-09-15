package db

import (
	"testing"
	"time"

	"github.com/rezible/rezible/ent"
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

	orgs, _ := NewOrganizationService(tdb, jobs)

	tenantCtx := s.SeedTenantContext()
	prefs, setErr := orgs.SetPreferences(tenantCtx, s.SeedOrganizationId(), func(m *ent.OrganizationPreferencesMutation) {
		m.SetInitialSetupAt(time.Now().UTC())
	})
	s.Require().NoError(setErr)

	s.False(prefs.InitialSetupAt.IsZero())
}
