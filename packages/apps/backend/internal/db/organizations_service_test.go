package db

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/rezible/rezible/testkit"
	"github.com/rezible/rezible/testkit/mocks"
)

type OrganizationsServiceSuite struct {
	testkit.Suite
}

func TestOrganizationsServiceSuite(t *testing.T) {
	suite.Run(t, &OrganizationsServiceSuite{
		Suite: testkit.NewSuite(),
	})
}

func (s *OrganizationsServiceSuite) TestFindOrCreateFromProviderDisallowsTenantCreationWhenConfigDisabled() {
	// TODO: need a better way to override config
	//orgs, orgsErr := NewOrganizationsService(s.Client(), mocks.NewMockJobsService(s.T()))
	//s.Require().NoError(orgsErr)
	//
	//s.SetConfigOverrides(map[string]any{"disable_tenant_creation": true})
	//_, createErr := orgs.FindOrCreateFromDomain(s.SystemContext(), generateRandomDomain())
	//s.Require().Error(createErr)
	//s.ErrorIs(createErr, rez.ErrInvalidTenant)
}

func (s *OrganizationsServiceSuite) TestCompleteSetupEnqueuesSyncJobAndSetsTimestamp() {
	s.SeedTestEntities()

	jobs := mocks.NewMockJobsService(s.T())
	jobs.On("Insert", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	dbc := s.Client()
	orgs, orgsErr := NewOrganizationsService(dbc, jobs)
	s.Require().NoError(orgsErr)

	tenantCtx := s.SeedTenantContext()
	setupErr := orgs.CompleteSetup(tenantCtx, s.SeedOrganization)
	s.Require().NoError(setupErr)

	updated, getErr := dbc.Organization.Get(tenantCtx, s.SeedOrganization.ID)
	s.Require().NoError(getErr)
	s.False(updated.InitialSetupAt.IsZero())
}
