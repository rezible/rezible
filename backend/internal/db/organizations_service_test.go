package db

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/internal/testkit"
	"github.com/rezible/rezible/internal/testkit/mocks"
)

type OrganizationsServiceSuite struct {
	testkit.Suite
}

func TestOrganizationsServiceSuite(t *testing.T) {
	s := &OrganizationsServiceSuite{Suite: testkit.NewSuite()}
	suite.Run(t, s)
}

func (s *OrganizationsServiceSuite) TestFindOrCreateFromProviderCreatesTenantAndOrg() {
	ctx := s.Context()
	jobs := mocks.NewMockJobsService(s.T())

	svc, err := NewOrganizationsService(s.Client(), jobs)
	s.Require().NoError(err)

	beforeCount, err := s.Client().Tenant.Query().Count(ctx)
	s.Require().NoError(err)

	created, err := svc.FindOrCreateFromProvider(ctx, ent.Organization{ExternalID: "provider-org-1", Name: "Acme"})
	s.Require().NoError(err)
	s.Equal("provider-org-1", created.ExternalID)

	tenantCount, err := s.Client().Tenant.Query().Count(ctx)
	s.Require().NoError(err)
	s.Equal(beforeCount+1, tenantCount)

	again, err := svc.FindOrCreateFromProvider(ctx, ent.Organization{ExternalID: "provider-org-1", Name: "Acme"})
	s.Require().NoError(err)
	s.Equal(created.ID, again.ID)
}

func (s *OrganizationsServiceSuite) TestFindOrCreateFromProviderDisallowsTenantCreationWhenConfigDisabled() {
	svc, err := NewOrganizationsService(s.Client(), mocks.NewMockJobsService(s.T()))
	s.Require().NoError(err)

	_, err = svc.FindOrCreateFromProvider(s.Context(), ent.Organization{ExternalID: "provider-org-2", Name: "Nope"})
	s.Require().Error(err)
	s.ErrorIs(err, rez.ErrInvalidTenant)
}

func (s *OrganizationsServiceSuite) TestCompleteSetupEnqueuesSyncJobAndSetsTimestamp() {
	base := s.SeedBaseTenant()
	jobs := mocks.NewMockJobsService(s.T())
	jobs.On("Insert", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	svc, err := NewOrganizationsService(s.Client(), jobs)
	s.Require().NoError(err)

	err = svc.CompleteSetup(base.Context, base.Organization)
	s.Require().NoError(err)

	updated, err := s.Client().Organization.Get(base.Context, base.Organization.ID)
	s.Require().NoError(err)
	s.False(updated.InitialSetupAt.IsZero())
}
