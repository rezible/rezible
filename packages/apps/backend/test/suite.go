package test

import (
	"testing"

	rez "github.com/rezible/rezible"
	"github.com/stretchr/testify/suite"
)

type Suite struct {
	suite.Suite

	cfg             *rez.Config
	configOverrides map[string]any

	suiteT   *testing.T
	sharedDB rez.Database
}

func NewSuite() Suite {
	return Suite{configOverrides: make(map[string]any)}
}

func (s *Suite) SetupSuite() {
	s.suiteT = s.T()
	s.cfg = new(s.loadConfig())
}

func (s *Suite) Config() rez.Config {
	if s.cfg == nil {
		s.cfg = new(s.loadConfig())
	}
	return *s.cfg
}

func (s *Suite) loadConfig() rez.Config {
	s.T().Helper()
	cfg, configErr := loadConfig(s.T().Context(), s.configOverrides)
	s.Require().NoError(configErr)
	return cfg
}
