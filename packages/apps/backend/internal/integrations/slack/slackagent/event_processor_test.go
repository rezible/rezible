package slackagent

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/pkg/projections"
	"github.com/rezible/rezible/test"
)

func TestAppSuite(t *testing.T) {
	suite.Run(t, &AppSuite{Suite: test.NewSuite()})
}

func (s *AppSuite) TestProcessUserEventUsesNativeUserResourceRef() {
	events, err := (&Integration{}).ProcessProviderEvent(context.Background(), rez.ProviderEvent{
		ProviderNamespace:   "workspace",
		ProviderEventSource: sourceUsers,
		ProviderEventRef:    "event-1",
		ReceivedAt:          time.Now().UTC(),
		Attributes:          json.RawMessage(`{"slack_id":"U123","name":"Alice","email":"alice@example.com"}`),
	})
	s.Require().NoError(err)
	s.Require().Len(events, 1)
	s.Require().Equal("U123", events[0].ProviderResourceRef)
	s.Require().Equal(projections.KindUser, events[0].Kind)
}

func (s *AppSuite) TestProcessTeamMembershipEventUsesTypedResourceRef() {
	events, err := (&Integration{}).ProcessProviderEvent(context.Background(), rez.ProviderEvent{
		ProviderNamespace:   "workspace",
		ProviderEventSource: sourceTeamMemberships,
		ProviderEventRef:    "event-1",
		ReceivedAt:          time.Now().UTC(),
		Attributes: json.RawMessage(`{
			"team":{"slack_id":"S123","name":"Platform","slug":"platform"},
			"user":{"slack_id":"U123","name":"Alice","email":"alice@example.com"}
		}`),
	})
	s.Require().NoError(err)
	s.Require().Len(events, 1)
	s.Require().Equal("membership:S123:U123", events[0].ProviderResourceRef)
	s.Require().Equal(projections.KindTeamMembership, events[0].Kind)
}
