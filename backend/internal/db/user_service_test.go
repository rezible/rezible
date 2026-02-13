package db

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/internal/testkit"
	"github.com/rezible/rezible/internal/testkit/mocks"
)

type UserServiceSuite struct {
	testkit.Suite
}

func TestUserServiceSuite(t *testing.T) {
	suite.Run(t, &UserServiceSuite{Suite: testkit.NewSuite()})
}

func (s *UserServiceSuite) TestCreateUserContextSetsTenantAndUserContext() {
	base := s.SeedBaseTenant()

	users, err := NewUserService(s.Client(), mocks.NewMockOrganizationService(s.T()))
	s.Require().NoError(err)

	usr := testkit.CreateUser(s.T(), s.Client(), base.Context)
	ctx, err := users.CreateUserContext(access.AnonymousContext(s.Context()), usr.ID)
	s.Require().NoError(err)

	s.Equal(base.TenantID, access.GetContext(ctx).GetTenantId())
	s.Equal(usr.ID, users.GetUserContext(ctx).ID)
}

func (s *UserServiceSuite) TestCreateUserContextReturnsInvalidUserForUnknownID() {
	base := s.SeedBaseTenant()
	
	users, err := NewUserService(s.Client(), mocks.NewMockOrganizationService(s.T()))
	s.Require().NoError(err)

	_, err = users.CreateUserContext(base.Context, uuid.New())
	s.Require().Error(err)
	s.ErrorIs(err, rez.ErrInvalidUser)
}
