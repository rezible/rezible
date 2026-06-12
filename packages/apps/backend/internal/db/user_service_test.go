package db

import (
	"testing"

	"github.com/rezible/rezible/testkit/mocks"
	"github.com/stretchr/testify/suite"

	"github.com/rezible/rezible/testkit"
)

type UserServiceSuite struct {
	testkit.Suite
}

func TestUserServiceSuite(t *testing.T) {
	suite.Run(t, &UserServiceSuite{Suite: testkit.NewSuite()})
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
