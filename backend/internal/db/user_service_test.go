package db

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/testkit"
	"github.com/rezible/rezible/testkit/mocks"
)

type UserServiceSuite struct {
	testkit.Suite
}

func TestUserServiceSuite(t *testing.T) {
	suite.Run(t, &UserServiceSuite{Suite: testkit.NewSuite()})
}

func (s *UserServiceSuite) TestCreateUserContextSetsTenantAndUserContext() {
	orgs := mocks.NewMockOrganizationService(s.T())
	users, usersErr := NewUserService(s.Client(), orgs)
	s.Require().NoError(usersErr)

	tenantCtx := s.GetSeedTenantContext()
	usrName := "test-user"
	usr := s.CreateTestUser(tenantCtx, usrName)
	s.Require().NotNil(usr)
	s.Require().Equal(usr.Name, usrName)

	userCtx, userCtxErr := users.CreateUserContext(s.GetAnonymousContext(), usr.ID)
	s.Require().NoError(userCtxErr)

	s.Equal(access.GetContext(tenantCtx).GetTenantId(), access.GetContext(userCtx).GetTenantId())
	s.Equal(usr.ID, users.GetUserContext(userCtx).ID)
}

func (s *UserServiceSuite) TestCreateUserContextReturnsInvalidUserForUnknownID() {
	users, err := NewUserService(s.Client(), mocks.NewMockOrganizationService(s.T()))
	s.Require().NoError(err)

	_, err = users.CreateUserContext(s.GetSeedTenantContext(), uuid.New())
	s.Require().Error(err)
	s.ErrorIs(err, rez.ErrInvalidUser)
}
