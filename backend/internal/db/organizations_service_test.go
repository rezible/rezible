package db

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/testkit"
	"github.com/rezible/rezible/testkit/mocks"
)

type OrganizationsServiceSuite struct {
	testkit.Suite
}

func TestOrganizationsServiceSuite(t *testing.T) {
	s := &OrganizationsServiceSuite{
		Suite: testkit.NewSuite(),
	}
	suite.Run(t, s)
}

func (s *OrganizationsServiceSuite) TestFindOrCreateFromProviderCreatesTenantAndOrg() {
	dbc := s.Client()

	orgs, orgsErr := NewOrganizationsService(dbc, mocks.NewMockJobsService(s.T()))
	s.Require().NoError(orgsErr)

	systemCtx := s.GetSystemContext()

	beforeCount, beforeCountErr := dbc.Tenant.Query().Count(systemCtx)
	s.Require().NoError(beforeCountErr)

	domain := "example.com"

	created, createErr := orgs.FindOrCreateFromProviderDomain(systemCtx, domain)
	s.Require().NoError(createErr)
	s.Equal(domain, created.ExternalID)

	found, findErr := orgs.FindOrCreateFromProviderDomain(systemCtx, domain)
	s.Require().NoError(findErr)
	s.Equal(created.ID, found.ID)

	afterCount, afterCountErr := dbc.Tenant.Query().Count(systemCtx)
	s.Require().NoError(afterCountErr)
	s.Equal(beforeCount+1, afterCount)
}

func (s *OrganizationsServiceSuite) TestFindOrCreateFromProviderDisallowsTenantCreationWhenConfigDisabled() {
	orgs, orgsErr := NewOrganizationsService(s.Client(), mocks.NewMockJobsService(s.T()))
	s.Require().NoError(orgsErr)

	s.SetConfigOverrides(map[string]any{"disable_tenant_creation": true})
	_, createErr := orgs.FindOrCreateFromProviderDomain(s.GetSystemContext(), "example.com")
	s.Require().Error(createErr)
	s.ErrorIs(createErr, rez.ErrInvalidTenant)
	s.SetConfigOverrides(nil)
}

func (s *OrganizationsServiceSuite) TestCompleteSetupEnqueuesSyncJobAndSetsTimestamp() {
	jobs := mocks.NewMockJobsService(s.T())
	jobs.On("Insert", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	dbc := s.Client()
	orgs, orgsErr := NewOrganizationsService(dbc, jobs)
	s.Require().NoError(orgsErr)

	tenantCtx := s.GetSeedTenantContext()
	setupErr := orgs.CompleteSetup(tenantCtx, s.SeedOrganization)
	s.Require().NoError(setupErr)

	updated, getErr := dbc.Organization.Get(tenantCtx, s.SeedOrganization.ID)
	s.Require().NoError(getErr)
	s.False(updated.InitialSetupAt.IsZero())
}
