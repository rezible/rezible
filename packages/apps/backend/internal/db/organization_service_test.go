package db

import (
	"testing"
	"time"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/test"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/rezible/rezible/test/mocks"
)

type OrganizationsServiceSuite struct {
	test.Suite
}

func TestOrganizationsServiceSuite(t *testing.T) {
	suite.Run(t, &OrganizationsServiceSuite{Suite: test.NewSuite()})
}

func (s *OrganizationsServiceSuite) TestCompleteSetupEnqueuesSyncJobAndSetsTimestamp() {
	tdb := s.CreateTestDatabase()
	jobs := mocks.NewMockJobService(s.T())
	jobs.On("Insert", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil).Once()

	orgs, _ := NewOrganizationService(tdb, jobs)

	tenantCtx := s.SeedTenantContext()
	prefs, setErr := orgs.SetPreferences(tenantCtx, s.SeedOrganizationId(), func(m *ent.OrganizationPreferencesMutation) {
		m.SetInitialSetupAt(time.Now().UTC())
	})
	s.Require().NoError(setErr)

	s.False(prefs.InitialSetupAt.IsZero())
	s.True(jobs.AssertCalled(s.T(), "Insert", mock.Anything, orgInitialSetupIntegrationSyncJob, mock.Anything))
}
