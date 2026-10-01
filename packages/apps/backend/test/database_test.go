package test

import (
	"context"
	"testing"

	"github.com/rezible/rezible/ent/tenant"
	"github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/stretchr/testify/suite"
)

type TestDatabaseSuite struct {
	Suite
}

func TestSuiteDatabase(t *testing.T) {
	suite.Run(t, &TestDatabaseSuite{Suite: NewSuite()})
}

func (s *TestDatabaseSuite) tenantID(ctx context.Context) int {
	tenantID, tenantIDSet := execution.GetContext(ctx).TenantID()
	s.Require().True(tenantIDSet, "context has no tenant")
	return tenantID
}

func (s *TestDatabaseSuite) TestSharedDatabaseOutlivesTheTestThatFirstUsedIt() {
	var subtestTenantID int
	s.Run("first use", func() {
		ctx, _ := s.SetupTestDatabase()
		subtestTenantID = s.tenantID(ctx)
	})

	ctx, db := s.SetupTestDatabase()
	tenantIDs := []int{subtestTenantID, s.tenantID(ctx)}

	systemCtx := execution.NewSystemContext(s.T().Context())
	queryTenants := db.Client(systemCtx).Tenant.Query().
		Where(tenant.IDIn(tenantIDs...))
	count, countErr := queryTenants.Count(systemCtx)
	s.Require().NoError(countErr)
	s.Equal(2, count)
}

func (s *TestDatabaseSuite) TestEachCallCreatesAnIsolatedSeededTenant() {
	firstCtx, firstDB := s.SetupTestDatabase()
	secondCtx, secondDB := s.SetupTestDatabase()

	s.Same(firstDB, secondDB)
	s.NotEqual(s.tenantID(firstCtx), s.tenantID(secondCtx))

	firstUserID, firstUserErr := firstDB.Client(firstCtx).User.Query().OnlyID(firstCtx)
	s.Require().NoError(firstUserErr)
	firstOrganizations, firstOrganizationsErr := firstDB.Client(firstCtx).Organization.Query().Count(firstCtx)
	s.Require().NoError(firstOrganizationsErr)
	s.Equal(1, firstOrganizations)

	queryOtherTenantUser := secondDB.Client(secondCtx).User.Query().
		Where(user.ID(firstUserID))
	visible, visibleErr := queryOtherTenantUser.Exist(secondCtx)
	s.Require().NoError(visibleErr)
	s.False(visible, "a tenant must not see another tenant's user")
}

func (s *TestDatabaseSuite) TestFreshDatabaseIsSeparateFromSharedDatabase() {
	sharedCtx, sharedDB := s.SetupTestDatabase()
	_, freshDB := s.SetupTestDatabase(WithFreshDatabase())

	s.NotSame(sharedDB, freshDB)

	sharedUserID, sharedUserErr := sharedDB.Client(sharedCtx).User.Query().OnlyID(sharedCtx)
	s.Require().NoError(sharedUserErr)

	systemCtx := execution.NewSystemContext(s.T().Context())
	freshTenants, freshTenantsErr := freshDB.Client(systemCtx).Tenant.Query().Count(systemCtx)
	s.Require().NoError(freshTenantsErr)
	s.Equal(1, freshTenants)

	queryMarker := freshDB.Client(systemCtx).User.Query().
		Where(user.ID(sharedUserID))
	markerExists, markerErr := queryMarker.Exist(systemCtx)
	s.Require().NoError(markerErr)
	s.False(markerExists)
}

func (s *TestDatabaseSuite) TestSeedingOptions() {
	withoutUserCtx, db := s.SetupTestDatabase(WithoutSeedUser())
	organizations, organizationsErr := db.Client(withoutUserCtx).Organization.Query().Count(withoutUserCtx)
	s.Require().NoError(organizationsErr)
	s.Equal(1, organizations)
	users, usersErr := db.Client(withoutUserCtx).User.Query().Count(withoutUserCtx)
	s.Require().NoError(usersErr)
	s.Zero(users)

	tenantOnlyCtx, _ := s.SetupTestDatabase(WithoutSeedOrganization())
	organizations, organizationsErr = db.Client(tenantOnlyCtx).Organization.Query().Count(tenantOnlyCtx)
	s.Require().NoError(organizationsErr)
	s.Zero(organizations)
	users, usersErr = db.Client(tenantOnlyCtx).User.Query().Count(tenantOnlyCtx)
	s.Require().NoError(usersErr)
	s.Zero(users)
}

func (s *TestDatabaseSuite) TestIdentitiesAreIsolated() {
	_, db := s.SetupTestDatabase()
	aliceCtx, alice := s.NewIdentity(db, "Alice")
	_, bob := s.NewIdentity(db, "Bob")

	s.NotEqual(alice.Session.TenantID, bob.Session.TenantID)
	s.NotEqual(alice.Session.OrganizationID, bob.Session.OrganizationID)
	s.NotEqual(alice.Session.UserID, bob.Session.UserID)

	ownUser := db.Client(aliceCtx).User.Query().
		Where(user.ID(alice.Session.UserID))
	ownUserExists, ownUserErr := ownUser.Exist(aliceCtx)
	s.Require().NoError(ownUserErr)
	s.True(ownUserExists)

	otherUser := db.Client(aliceCtx).User.Query().
		Where(user.ID(bob.Session.UserID))
	otherUserExists, otherUserErr := otherUser.Exist(aliceCtx)
	s.Require().NoError(otherUserErr)
	s.False(otherUserExists)
}
