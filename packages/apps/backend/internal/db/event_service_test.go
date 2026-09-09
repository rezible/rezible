package db

import (
	"testing"

	"github.com/rezible/rezible/test"
	"github.com/stretchr/testify/suite"
)

type EventsServiceSuite struct {
	test.Suite
}

func TestEventsServiceSuite(t *testing.T) {
	suite.Run(t, &EventsServiceSuite{Suite: test.NewSuite()})
}
