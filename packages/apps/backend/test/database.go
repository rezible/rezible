package test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/stretchr/testify/assert"

	"github.com/rezible/rezible/internal/postgres"
	"github.com/rezible/rezible/internal/postgres/pgtestdb"
	"github.com/rezible/rezible/pkg/execution"
)

type testDatabaseOptions struct {
	fresh                bool
	skipSeedOrganization bool
	skipSeedUser         bool
}

type TestDatabaseOption func(*testDatabaseOptions)

// WithFreshDatabase creates a new migrated database for this call instead of
// using the suite's shared database. It is dropped when the current test ends.
func WithFreshDatabase() TestDatabaseOption {
	return func(o *testDatabaseOptions) { o.fresh = true }
}

// WithoutSeedOrganization creates only the tenant. No organization or user is seeded.
func WithoutSeedOrganization() TestDatabaseOption {
	return func(o *testDatabaseOptions) { o.skipSeedOrganization = true }
}

// WithoutSeedUser creates the tenant and organization but no user.
func WithoutSeedUser() TestDatabaseOption {
	return func(o *testDatabaseOptions) { o.skipSeedUser = true }
}

// SetupTestDatabase creates a new tenant and returns its context with the
// database it lives in. The database is shared by the suite unless
// WithFreshDatabase is passed.
func (s *Suite) SetupTestDatabase(optFns ...TestDatabaseOption) (context.Context, rez.Database) {
	s.T().Helper()

	opts := testDatabaseOptions{}
	for _, optFn := range optFns {
		optFn(&opts)
	}

	var db rez.Database
	if opts.fresh {
		db = s.createDatabase(s.T())
	} else {
		db = s.sharedDatabase()
	}

	ctx := s.seedTenant(db, opts)
	return ctx, db
}

func (s *Suite) sharedDatabase() rez.Database {
	s.T().Helper()
	if s.sharedDB != nil {
		return s.sharedDB
	}

	s.Require().NotNil(s.suiteT, "test.Suite.SetupSuite did not run; a suite defining SetupSuite must call s.Suite.SetupSuite()")
	s.sharedDB = s.createDatabase(s.suiteT)
	return s.sharedDB
}

// createDatabase creates a migrated database owned by owner: the pool is opened
// with owner's context and cleanup is registered on owner.
func (s *Suite) createDatabase(owner *testing.T) rez.Database {
	s.T().Helper()
	start := time.Now()
	cfg := s.Config().Postgres
	s.Require().NotEmpty(cfg.AdminRole.Name, "postgres migrations admin config empty")

	tdb, tdbErr := pgtestdb.New(cfg)
	s.Require().NoError(tdbErr)
	owner.Cleanup(func() {
		assert.NoError(owner, tdb.Shutdown())
	})

	pool, poolErr := postgres.MakePgxPool(owner.Context(), tdb.Config(), false)
	s.Require().NoError(poolErr)
	owner.Cleanup(func() {
		assert.NoError(owner, pool.Shutdown())
	})

	db, dbErr := postgres.NewPgxPoolDatabaseClient(pool)
	s.Require().NoError(dbErr)
	owner.Cleanup(func() {
		assert.NoError(owner, db.Shutdown())
	})

	s.T().Logf("created test database in %dms", time.Since(start).Milliseconds())
	return db
}

func (s *Suite) seedTenant(db rez.Database, opts testDatabaseOptions) context.Context {
	s.T().Helper()
	systemCtx := execution.NewSystemContext(s.T().Context())
	tenant, tenantErr := db.Client(systemCtx).Tenant.Create().Save(systemCtx)
	s.Require().NoError(tenantErr, "failed to create tenant")

	ctx := execution.NewTenantContext(s.T().Context(), tenant.ID)
	client := db.Client(ctx)

	if opts.skipSeedOrganization {
		return ctx
	}
	createOrganization := client.Organization.Create().
		SetName("Test Organization").
		SetAuthProviderID(uuid.NewString())
	organizationErr := createOrganization.Exec(ctx)
	s.Require().NoError(organizationErr, "failed to create organization")

	if opts.skipSeedUser {
		return ctx
	}
	createUser := client.User.Create().
		SetEmail("owner+" + uuid.NewString() + "@example.com").
		SetName("Owner")
	userErr := createUser.Exec(ctx)
	s.Require().NoError(userErr, "failed to create user")

	return ctx
}

// TODO: remove this and just make `WithUserAuthSession` an option above

type Identity struct {
	Session *ent.UserAuthSession
}

func (s *Suite) NewIdentity(db rez.Database, label string) (context.Context, Identity) {
	s.T().Helper()

	systemCtx := execution.NewSystemContext(s.T().Context())
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

	return execution.NewUserContext(s.T().Context(), session), Identity{Session: session}
}
