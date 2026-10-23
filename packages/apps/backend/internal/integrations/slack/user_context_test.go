package slackintegration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/google/uuid"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/mock"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/test/mocks"
)

func (s *SlackIntegrationSuite) TestLookupUserContext() {
	ctx := execution.NewTenantContext(s.T().Context(), 7)

	s.Run("a known Slack user", func() {
		known := &ent.User{ID: uuid.New(), TenantID: 7}
		users := mocks.NewMockUserService(s.T())
		users.EXPECT().
			Get(mock.Anything, mock.Anything).
			Return(known, nil)

		userCtx, lookupErr := LookupUserContext(ctx, users, "U1")

		s.Require().NoError(lookupErr)
		userID, hasUser := execution.GetContext(userCtx).UserID()
		s.True(hasUser)
		s.Equal(known.ID, userID)
	})

	s.Run("a Slack user the member sync has not seen", func() {
		users := mocks.NewMockUserService(s.T())
		users.EXPECT().
			Get(mock.Anything, mock.Anything).
			Return(nil, &ent.NotFoundError{})

		_, lookupErr := LookupUserContext(ctx, users, "U2")

		s.ErrorIs(lookupErr, ErrChatUserNotFound)
	})
}

func (s *SlackIntegrationSuite) TestSlashCommandFromUnknownUserIsToldWhy() {
	var replies []slack.WebhookMessage
	responseURL := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reply slack.WebhookMessage
		s.NoError(json.NewDecoder(r.Body).Decode(&reply))
		replies = append(replies, reply)
	}))
	s.T().Cleanup(responseURL.Close)

	installations := mocks.NewMockIntegrationService(s.T())
	installations.EXPECT().
		LookupInstallation(mock.Anything, mock.Anything).
		Return(&ent.Integration{TenantID: 7}, nil)
	users := mocks.NewMockUserService(s.T())
	users.EXPECT().
		Get(mock.Anything, mock.Anything).
		Return(nil, &ent.NotFoundError{})
	handler := func(context.Context, *ent.Integration, *slack.SlashCommand) (*slack.Blocks, error) {
		s.Fail("the command runs only for known users")
		return nil, nil
	}
	service := &AppService[App]{
		integrationName:      "slack_test",
		installations:        installations,
		users:                users,
		slashCommandHandlers: map[string]SlashCommandHandler{"/rezible": handler},
	}
	command := &slashCommandEvent{
		IntegrationName: "slack_test",
		Command: slack.SlashCommand{
			Command:     "/rezible",
			TeamID:      "T1",
			UserID:      "U2",
			ResponseURL: responseURL.URL,
		},
	}

	handleErr := service.handleSlashCommand(s.T().Context(), command)

	s.Require().NoError(handleErr)
	s.Require().Len(replies, 1)
	s.Equal(UnknownUserText, replies[0].Text)
	s.Equal(slack.ResponseTypeEphemeral, replies[0].ResponseType)
}
