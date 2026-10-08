package webhook

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/rezible/rezible/test"
)

type EventProcessorSuite struct {
	test.Suite
}

func TestEventProcessorSuite(t *testing.T) {
	suite.Run(t, &EventProcessorSuite{Suite: test.NewSuite()})
}

func (s *EventProcessorSuite) TestRejectsEventSourcesThatNameNoPreset() {
	// The attributes are a valid deployment report, so only the source can reject them.
	event, mapErr := deploymentPreset{}.MapDelivery(testInstallationID, encodeReport(s.T(), testReport()), testReceivedAt)
	s.Require().NoError(mapErr)

	for _, source := range []string{"", "incident", "deployments"} {
		s.Run(source, func() {
			unknown := *event
			unknown.ProviderEventSource = source

			normalized, processErr := EventProcessor{}.ProcessProviderEvent(s.T().Context(), unknown)

			s.Error(processErr)
			s.Empty(normalized)
		})
	}
}
