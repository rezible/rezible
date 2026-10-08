package webhook

import (
	"testing"

	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/test"
)

type IntegrationSuite struct {
	test.Suite
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, &IntegrationSuite{Suite: test.NewSuite()})
}

func (s *IntegrationSuite) TestInstallationConfigAcceptsOnlyAKnownPreset() {
	i := &Integration{}

	config, configErr := i.ValidateInstallationConfig([]byte(`{"preset":"deployment"}`))
	s.Require().NoError(configErr)
	encoded, encodeErr := config.Encode()
	s.Require().NoError(encodeErr)
	s.JSONEq(`{"preset":"deployment"}`, string(encoded))

	other, otherErr := i.ValidateInstallationConfig([]byte(`{"preset":"deployment"}`))
	s.Require().NoError(otherErr)
	s.NotEqual(config.InstallationTargetRef(), other.InstallationTargetRef(), "each installation has its own reference")

	for _, raw := range []string{`{"preset":"incident"}`, `{"preset":""}`, `{}`, `null`, ``, `{"preset":`} {
		_, rejectErr := i.ValidateInstallationConfig([]byte(raw))
		s.ErrorIs(rejectErr, rez.ErrInvalidInput, "config %q", raw)
	}
}

func (s *IntegrationSuite) TestUserSettingsCannotBeSet() {
	i := &Integration{}

	s.NoError(i.ValidateUserSettings(nil))
	s.NoError(i.ValidateUserSettings(map[string]any{}))
	s.ErrorIs(i.ValidateUserSettings(map[string]any{"preset": "incident"}), rez.ErrInvalidInput)
}
