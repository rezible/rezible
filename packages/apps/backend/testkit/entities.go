package testkit

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
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

	tenant, tenantErr := s.Client().Tenant.Create().Save(ctx)
	s.Require().NoError(tenantErr, "failed to create tenant")
	s.SeedTenant = tenant
	ctx = s.SeedTenantContext()

	if s.opts.skipSeedOrganization {
		return
	}
	org, orgErr := s.Client().Organization.Create().
		SetName("Test Organization").
		SetAuthProviderID(uuid.NewString()).
		Save(ctx)
	s.Require().NoError(orgErr, "failed to create organization")
	s.SeedOrganization = org

	if s.opts.skipSeedUser {
		return
	}
	_, usrErr := s.Client().User.Create().
		SetEmail("owner+" + uuid.NewString() + "@example.com").
		SetName("Owner").
		Save(ctx)
	s.Require().NoError(usrErr, "failed to create user")
}

func (s *Suite) CreateTestUser(ctx context.Context, name string) *ent.User {
	s.T().Helper()
	create := s.Client().User.Create().
		SetEmail(next("user") + "@example.com").
		SetName(name)
	u, saveErr := create.Save(ctx)
	s.Require().NoError(saveErr, "failed to save user")
	return u
}
