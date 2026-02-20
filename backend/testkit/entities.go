package testkit

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
)

var seq atomic.Int64

func next(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, seq.Add(1))
}

func (s *Suite) seedTestEntities() {
	if !s.opts.seedTenant {
		return
	}
	sysCtx := access.SystemContext(s.T().Context())

	tenant, tenantErr := s.Client().Tenant.Create().Save(sysCtx)
	s.Require().NoError(tenantErr, "failed to create tenant")
	s.SeedTenant = tenant

	tenantCtx := access.TenantContext(sysCtx, tenant.ID)
	if s.opts.seedOrganization {
		id := uuid.NewString()
		org, orgErr := s.Client().Organization.Create().
			SetName("Test Organization" + id[:4]).
			SetExternalID("org-" + id).
			Save(tenantCtx)
		s.Require().NoError(orgErr, "failed to create organization")
		s.SeedOrganization = org
	}

	if s.opts.seedUser {
		_, usrCtx := s.Client().User.Create().
			SetEmail("owner+" + uuid.NewString() + "@example.com").
			SetName("Owner").
			Save(tenantCtx)
		s.Require().NoError(usrCtx, "failed to create user")
	}
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
