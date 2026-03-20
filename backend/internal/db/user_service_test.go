package db

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/rezible/rezible/testkit"
)

type UserServiceSuite struct {
	testkit.Suite
}

func TestUserServiceSuite(t *testing.T) {
	suite.Run(t, &UserServiceSuite{Suite: testkit.NewSuite()})
}
