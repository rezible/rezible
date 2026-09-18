package db

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/rezible/rezible/test"
)

type KnowledgeGraphQueryServiceSuite struct {
	test.Suite
}

func TestKnowledgeGraphQueryServiceSuite(t *testing.T) {
	suite.Run(t, &KnowledgeGraphQueryServiceSuite{Suite: test.NewSuite()})
}
