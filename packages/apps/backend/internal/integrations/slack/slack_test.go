package slackintegration

import (
	"testing"

	"github.com/rezible/rezible/test"
	"github.com/stretchr/testify/suite"
)

type SlackIntegrationSuite struct {
	test.Suite
}

func TestSlackIntegrationSuite(t *testing.T) {
	suite.Run(t, &SlackIntegrationSuite{Suite: test.NewSuite()})
}
