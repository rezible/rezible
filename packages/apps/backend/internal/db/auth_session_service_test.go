package db

import (
	"testing"
	"time"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/ent/userauthsession"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/test"
	"github.com/rezible/rezible/test/mocks"
	"github.com/stretchr/testify/suite"
)

type AuthSessionServiceSuite struct {
	test.Suite
}

func TestAuthSessionServiceSuite(t *testing.T) {
	suite.Run(t, &AuthSessionServiceSuite{Suite: test.NewSuite()})
}

func (s *AuthSessionServiceSuite) TestCreatingSessionPreservesExistingSessions() {
	sessions, _ := s.newServices()
	providerSession := &rez.UserAuthProviderSession{
		User:      ent.User{Email: "session-test@example.com", Name: "Session Test", AuthProviderID: "session-test-user"},
		Org:       ent.Organization{Name: "Session Test Org", AuthProviderID: "session-test-org"},
		ExpiresAt: time.Now().Add(time.Hour),
	}

	first, firstErr := sessions.CreateFromUserAuthResponse(s.T().Context(), providerSession)
	s.Require().NoError(firstErr)
	second, secondErr := sessions.CreateFromUserAuthResponse(s.T().Context(), providerSession)
	s.Require().NoError(secondErr)
	s.NotEqual(first.ID, second.ID)

	loaded, lookupErr := sessions.LookupSession(s.T().Context(), first.ID)
	s.Require().NoError(lookupErr)
	s.Equal(first.ID, loaded.ID)
}

func (s *AuthSessionServiceSuite) TestLookupSessionMissingSessionIsNotFound() {
	sessions, _ := s.newServices()

	_, lookupErr := sessions.LookupSession(s.T().Context(), uuid.New())
	s.Require().Error(lookupErr)
	s.True(ent.IsNotFound(lookupErr))
}

func (s *AuthSessionServiceSuite) newServices() (*AuthSessionService, *UserService) {
	_, database := s.SetupTestDatabase()
	jobs := mocks.NewMockJobService(s.T())
	organizations, _ := NewOrganizationService(database, jobs)
	users, _ := NewUserService(database, organizations)
	sessions, _ := NewAuthSessionService(database, organizations, users)
	return sessions, users
}

func (s *AuthSessionServiceSuite) signIn(sessions *AuthSessionService, orgID, userID string, role user.Role) *ent.UserAuthSession {
	sess, err := sessions.CreateFromUserAuthResponse(s.T().Context(), &rez.UserAuthProviderSession{
		User:      ent.User{Email: userID + "@example.com", Name: userID, AuthProviderID: userID},
		Org:       ent.Organization{Name: orgID, AuthProviderID: orgID},
		Role:      role,
		ExpiresAt: time.Now().Add(time.Hour),
	})
	s.Require().NoError(err)
	return sess
}

func (s *AuthSessionServiceSuite) userRole(users *UserService, sess *ent.UserAuthSession) user.Role {
	ctx := execution.NewTenantContext(s.T().Context(), sess.TenantID)
	usr, err := users.Get(ctx, user.ID(sess.UserID))
	s.Require().NoError(err)
	return usr.Role
}

func (s *AuthSessionServiceSuite) TestSigningInTakesTheProviderRole() {
	sessions, users := s.newServices()

	// A member creating the organization stays a member.
	created := s.signIn(sessions, "org-a", "creator", user.RoleMember)
	s.Equal(userauthsession.RoleMember, created.Role)
	s.Equal(user.RoleMember, s.userRole(users, created))

	asAdmin := s.signIn(sessions, "org-a", "person", user.RoleAdmin)
	asMember := s.signIn(sessions, "org-a", "person", user.RoleMember)
	s.Equal(asAdmin.UserID, asMember.UserID)
	s.Equal(user.RoleMember, s.userRole(users, asMember))

	loaded, err := sessions.LookupSession(s.T().Context(), asAdmin.ID)
	s.Require().NoError(err)
	s.Equal(userauthsession.RoleAdmin, loaded.Role)

	// The same provider user in another organization has its own role.
	otherOrg := s.signIn(sessions, "org-b", "person", user.RoleAdmin)
	s.NotEqual(asMember.TenantID, otherOrg.TenantID)
	s.Equal(user.RoleAdmin, s.userRole(users, otherOrg))
	s.Equal(user.RoleMember, s.userRole(users, asMember))
}

func (s *AuthSessionServiceSuite) TestSigningInWithoutProviderRole() {
	sessions, users := s.newServices()

	first := s.signIn(sessions, "org", "first", "")
	second := s.signIn(sessions, "org", "second", "")
	s.Equal(userauthsession.RoleAdmin, first.Role)
	s.Equal(userauthsession.RoleMember, second.Role)

	s.Equal(user.RoleAdmin, s.userRole(users, s.signIn(sessions, "org", "first", "")))
	s.Equal(user.RoleMember, s.userRole(users, s.signIn(sessions, "org", "second", "")))
}
