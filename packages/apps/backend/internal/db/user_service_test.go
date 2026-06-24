package db

import (
	"testing"

	"github.com/rezible/rezible/test"
	"github.com/rezible/rezible/test/mocks"
	"github.com/stretchr/testify/suite"
)

type UserServiceSuite struct {
	test.Suite
}

func TestUserServiceSuite(t *testing.T) {
	suite.Run(t, &UserServiceSuite{Suite: test.NewSuite()})
}

func (s *UserServiceSuite) newUserService(jobs *mocks.MockJobService) *UserService {
	sdb := s.Database()
	svc, err := NewUserService(sdb, NewOrganizationService(sdb, jobs), NewKnowledgeService(sdb))
	s.Require().NoError(err)
	return svc
}

func (s *UserServiceSuite) TestSyncFromAuthProvider() {
	// TODO
}
