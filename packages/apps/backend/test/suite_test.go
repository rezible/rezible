package test

import (
	"testing"

	"github.com/rezible/rezible/ent/user"
	"github.com/stretchr/testify/suite"
)

type resourcesSuite struct {
	Suite
}

func TestSuiteResources(t *testing.T) {
	suite.Run(t, &resourcesSuite{Suite: NewSuite()})
}

func (s *resourcesSuite) TestDatabasesAndTenantsAreIsolated() {
	first := s.CreateTestDatabase()
	second := s.CreateTestDatabase()
	aliceCtx, alice := s.NewIdentity(first, "Alice")
	_, bob := s.NewIdentity(first, "Bob")

	s.NotEqual(alice.Session.TenantID, bob.Session.TenantID)
	s.NotEqual(alice.Session.OrganizationID, bob.Session.OrganizationID)
	s.NotEqual(alice.Session.UserID, bob.Session.UserID)

	ownUser := first.Client(aliceCtx).User.Query().
		Where(user.ID(alice.Session.UserID))
	ownUserExists, ownUserErr := ownUser.Exist(aliceCtx)
	s.Require().NoError(ownUserErr)
	s.True(ownUserExists)

	otherUser := first.Client(aliceCtx).User.Query().
		Where(user.ID(bob.Session.UserID))
	otherUserExists, otherUserErr := otherUser.Exist(aliceCtx)
	s.Require().NoError(otherUserErr)
	s.False(otherUserExists)

	systemCtx := s.SystemContext()
	marker := second.Client(systemCtx).User.Query().
		Where(user.ID(alice.Session.UserID))
	markerExists, markerErr := marker.Exist(systemCtx)
	s.Require().NoError(markerErr)
	s.False(markerExists)
}
