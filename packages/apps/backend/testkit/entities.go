package testkit

import (
	"fmt"
	"sync/atomic"

	"github.com/google/uuid"
)

func WithSkipSeedOrganization() Option {
	return func(o *options) { o.skipSeedOrganization = true }
}

func WithSkipSeedUser() Option {
	return func(o *options) { o.skipSeedUser = true }
}

var seq atomic.Int64

func next(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, seq.Add(1))
}

func (s *Suite) SeedTestEntities() {
	ctx := s.SystemContext()

	client := s.db.Client(ctx)

	tenant, tenantErr := client.Tenant.Create().Save(ctx)
	s.Require().NoError(tenantErr, "failed to create tenant")
	s.SeedTenant = tenant
	ctx = s.SeedTenantContext()

	if s.opts.skipSeedOrganization {
		return
	}
	org, orgErr := client.Organization.Create().
		SetName("Test Organization").
		SetAuthProviderID(uuid.NewString()).
		Save(ctx)
	s.Require().NoError(orgErr, "failed to create organization")
	s.SeedOrganization = org

	if s.opts.skipSeedUser {
		return
	}
	_, usrErr := client.User.Create().
		SetEmail("owner+" + uuid.NewString() + "@example.com").
		SetName("Owner").
		Save(ctx)
	s.Require().NoError(usrErr, "failed to create user")
}
