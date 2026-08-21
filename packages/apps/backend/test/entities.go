package test

import (
	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
)

var (
	seedTenantId = 1
	seedOrgId    = uuid.New()
	seedUserId   = uuid.New()
)

func (s *Suite) SeedOrganizationId() uuid.UUID {
	return seedOrgId
}

func (s *Suite) SeedTestEntities(db rez.Database) {
	ctx := s.SystemContext()
	client := db.Client(ctx)

	tenantErr := client.Tenant.Create().Exec(ctx)
	s.Require().NoError(tenantErr, "failed to create tenant")

	ctx = s.SeedTenantContext()

	if s.opts.skipSeedOrganization {
		s.T().Logf("skipping seeding organization")
		return
	}
	orgErr := client.Organization.Create().
		SetID(seedOrgId).
		SetName("Test Organization").
		SetAuthProviderID(uuid.NewString()).
		Exec(ctx)
	s.Require().NoError(orgErr, "failed to create organization")

	if s.opts.skipSeedUser {
		s.T().Logf("skipping seeding user")
		return
	}
	usrErr := client.User.Create().
		SetID(seedUserId).
		SetEmail("owner+" + uuid.NewString() + "@example.com").
		SetName("Owner").
		Exec(ctx)
	s.Require().NoError(usrErr, "failed to create user")
}
