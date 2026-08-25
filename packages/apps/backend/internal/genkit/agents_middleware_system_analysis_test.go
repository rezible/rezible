package genkit

import (
	"testing"

	"github.com/rezible/rezible/test"
	"github.com/stretchr/testify/suite"
)

type SystemAnalysisMiddlewareSuite struct {
	test.Suite
}

func TestSystemAnalysisMiddlewareSuite(t *testing.T) {
	suite.Run(t, &SystemAnalysisMiddlewareSuite{Suite: test.NewSuite()})
}
