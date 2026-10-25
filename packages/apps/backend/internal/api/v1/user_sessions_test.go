package apiv1

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/internal/db"
	"github.com/rezible/rezible/pkg/execution"
	oapi "github.com/rezible/rezible/pkg/openapi/v1"
	"github.com/rezible/rezible/test"
	"github.com/rezible/rezible/test/mocks"
)

type UserSessionsHandlerSuite struct {
	test.Suite
}

func TestUserSessionsHandlerSuite(t *testing.T) {
	suite.Run(t, &UserSessionsHandlerSuite{Suite: test.NewSuite()})
}

func (s *UserSessionsHandlerSuite) TestSessionReportsItsOwnRole() {
	_, tdb := s.SetupTestDatabase()
	orgs, orgsErr := db.NewOrganizationService(tdb, mocks.NewMockJobService(s.T()))
	s.Require().NoError(orgsErr)
	users, usersErr := db.NewUserService(tdb, orgs)
	s.Require().NoError(usersErr)
	sessions, sessionsErr := db.NewAuthSessionService(tdb, orgs, users)
	s.Require().NoError(sessionsErr)

	signIn := func(role user.Role) *ent.UserAuthSession {
		sess, err := sessions.CreateFromUserAuthResponse(s.T().Context(), &rez.UserAuthProviderSession{
			User:      ent.User{Email: "person@example.com", Name: "Person", AuthProviderID: "person"},
			Org:       ent.Organization{Name: "Org", AuthProviderID: "org"},
			Role:      role,
			ExpiresAt: time.Now().Add(time.Hour),
		})
		s.Require().NoError(err)
		return sess
	}
	adminSession := signIn(user.RoleAdmin)
	memberSession := signIn(user.RoleMember)

	sessionsHandler := newUserSessionsHandler(&baseHandler{users: users}, orgs)
	sessionRole := func(sess *ent.UserAuthSession) user.Role {
		resp, err := sessionsHandler.GetUserSession(execution.NewUserContext(s.T().Context(), sess), &oapi.GetUserSessionRequest{})
		s.Require().NoError(err)
		return resp.Body.Data.OrganizationRole
	}
	s.Equal(user.RoleAdmin, sessionRole(adminSession))
	s.Equal(user.RoleMember, sessionRole(memberSession))

	listed, listErr := newUsersHandler(users).ListUsers(execution.NewUserContext(s.T().Context(), adminSession), &oapi.ListUsersRequest{})
	s.Require().NoError(listErr)
	s.Require().Len(listed.Body.Data, 1)
	s.Equal(user.RoleMember, listed.Body.Data[0].Attributes.OrganizationRole)
}
