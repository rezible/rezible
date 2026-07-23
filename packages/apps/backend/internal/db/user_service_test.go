package db

import (
	"testing"

	"github.com/rezible/rezible/test"
	"github.com/stretchr/testify/suite"
)

type UserServiceSuite struct {
	test.Suite
}

func TestUserServiceSuite(t *testing.T) {
	suite.Run(t, &UserServiceSuite{Suite: test.NewSuite()})
}

func (s *UserServiceSuite) TestSyncFromAuthProvider() {
	// TODO
}
