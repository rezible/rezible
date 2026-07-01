package genkit

import (
	"testing"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/test"
	"github.com/stretchr/testify/suite"
)

type AgentRegistrySuite struct {
	test.Suite
}

func TestAgentsRegistrySuite(t *testing.T) {
	suite.Run(t, &AgentRegistrySuite{Suite: test.NewSuite()})
}

func (s *AgentRegistrySuite) createRegistry() *AgentRegistry {
	r := &AgentRegistry{
		agents:    make(map[string]rez.Agent),
		snapshots: nil,
	}

	return r
}

func (s *AgentRegistrySuite) TestSimpleWorkflowAgent() {
	//ctx := s.SeedTenantContext()

}
