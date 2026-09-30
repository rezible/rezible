package test

import (
	"context"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/pkg/execution"
)

// seedTenantId is fixed so SeedTenantContext works before CreateTestDatabase.
// Tenant IDs come from a sequence, so seeding verifies the first tenant gets it.
const seedTenantId = 1

var (
	seedOrgId  = uuid.New()
	seedUserId = uuid.New()
)

func (s *Suite) SeedOrganizationId() uuid.UUID {
	return seedOrgId
}

func (s *Suite) seedTestEntities(db rez.Database) {
	ctx := s.SystemContext()
	client := db.Client(ctx)

	tenant, tenantErr := client.Tenant.Create().Save(ctx)
	s.Require().NoError(tenantErr, "failed to create tenant")
	s.Require().Equal(seedTenantId, tenant.ID, "seed tenant must be the first tenant in a fresh test database")

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

// Identity contains the persisted data used to authenticate a test user.
type Identity struct {
	Session *ent.UserAuthSession
}

// NewIdentity creates a separate tenant and identity in the supplied database.
// It returns the authenticated context separately from the persisted identity.
func (s *Suite) NewIdentity(db rez.Database, label string) (context.Context, Identity) {
	s.T().Helper()
	systemCtx := s.SystemContext()
	tenant, tenantErr := db.Client(systemCtx).Tenant.Create().Save(systemCtx)
	s.Require().NoError(tenantErr, "create tenant for %s identity", label)

	ctx := execution.NewTenantContext(s.T().Context(), tenant.ID)
	client := db.Client(ctx)

	createOrg := client.Organization.Create().
		SetName(label).
		SetAuthProviderID(uuid.NewString())
	org, orgErr := createOrg.Save(ctx)
	s.Require().NoError(orgErr, "create organization for %s identity", label)

	createUser := client.User.Create().
		SetName(label).
		SetEmail(uuid.NewString() + "@example.com")
	user, userErr := createUser.Save(ctx)
	s.Require().NoError(userErr, "create user for %s identity", label)

	createSession := client.UserAuthSession.Create().
		SetOrganizationID(org.ID).
		SetUserID(user.ID).
		SetExpiresAt(time.Now().UTC().Add(time.Hour))
	session, sessionErr := createSession.Save(ctx)
	s.Require().NoError(sessionErr, "create session for %s identity", label)

	userCtx := execution.NewUserContext(s.T().Context(), session)
	return userCtx, Identity{Session: session}
}
