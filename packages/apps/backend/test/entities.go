package test

import (
	"github.com/google/uuid"
)

func (s *Suite) SeedTestEntities() {
	ctx := s.SystemContext()

	client := s.db.Client(ctx)

	tenant, tenantErr := client.Tenant.Create().Save(ctx)
	s.Require().NoError(tenantErr, "failed to create tenant")
	s.SeedTenant = tenant
	ctx = s.SeedTenantContext()

	if s.opts.skipSeedOrganization {
		s.T().Logf("skipping seeding organization")
		return
	}
	org, orgErr := client.Organization.Create().
		SetName("Test Organization").
		SetAuthProviderID(uuid.NewString()).
		Save(ctx)
	s.Require().NoError(orgErr, "failed to create organization")
	s.SeedOrganization = org

	if s.opts.skipSeedUser {
		s.T().Logf("skipping seeding user")
		return
	}
	usr, usrErr := client.User.Create().
		SetEmail("owner+" + uuid.NewString() + "@example.com").
		SetName("Owner").
		Save(ctx)
	s.Require().NoError(usrErr, "failed to create user")
	s.SeedUser = usr
}
