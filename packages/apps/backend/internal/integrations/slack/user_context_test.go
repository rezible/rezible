package slackintegration

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"time"

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

func (s *SlackIntegrationSuite) TestSlashCommandReplyIsTheWebhookResponse() {
	const signingSecret = "test-signing-secret"
	known := &ent.User{ID: uuid.New(), TenantID: 7}

	installations := mocks.NewMockIntegrationService(s.T())
	installations.EXPECT().
		LookupInstallation(mock.Anything, mock.Anything).
		Return(&ent.Integration{TenantID: 7}, nil)
	users := mocks.NewMockUserService(s.T())
	users.EXPECT().
		Get(mock.Anything, mock.Anything).
		Return(known, nil).Once()
	users.EXPECT().
		Get(mock.Anything, mock.Anything).
		Return(nil, &ent.NotFoundError{}).Once()
	handler := func(ctx context.Context, _ *ent.Integration, _ *slack.SlashCommand) (*slack.Blocks, error) {
		userID, hasUser := execution.GetContext(ctx).UserID()
		s.True(hasUser)
		s.Equal(known.ID, userID)
		return CommandReply("Declared."), nil
	}
	service := &AppService[App]{
		integrationName:      "slack_test",
		installations:        installations,
		users:                users,
		slashCommandHandlers: map[string]SlashCommandHandler{"/rezible": handler},
	}
	eventHandler := &appEventHandler{onCommand: service.handleSlashCommand}
	webhooks := makeWebhookHandler(signingSecret, eventHandler)

	runCommand := func(userID string) *slack.WebhookMessage {
		form := url.Values{"command": {"/rezible"}, "team_id": {"T1"}, "user_id": {userID}}
		body := form.Encode()
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		signer := hmac.New(sha256.New, []byte(signingSecret))
		signer.Write([]byte("v0:" + timestamp + ":" + body))
		request := httptest.NewRequest(http.MethodPost, "/commands", strings.NewReader(body))
		request = request.WithContext(execution.NewRootContext(s.T().Context(), execution.KindAnonymous, execution.SourceHTTP))
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		request.Header.Set("X-Slack-Request-Timestamp", timestamp)
		request.Header.Set("X-Slack-Signature", "v0="+hex.EncodeToString(signer.Sum(nil)))
		recorder := httptest.NewRecorder()

		webhooks.ServeHTTP(recorder, request)

		s.Require().Equal(http.StatusOK, recorder.Code)
		var reply slack.WebhookMessage
		s.Require().NoError(json.Unmarshal(recorder.Body.Bytes(), &reply))
		return &reply
	}

	s.Run("a member", func() {
		reply := runCommand("U1")

		s.Equal(slack.ResponseTypeEphemeral, reply.ResponseType)
		s.Require().NotNil(reply.Blocks)
		s.Equal(CommandReply("Declared.").BlockSet, reply.Blocks.BlockSet)
	})

	s.Run("a Slack user the member sync has not seen", func() {
		reply := runCommand("U2")

		s.Equal(slack.ResponseTypeEphemeral, reply.ResponseType)
		s.Require().NotNil(reply.Blocks)
		s.Equal(CommandReply(UnknownUserText).BlockSet, reply.Blocks.BlockSet)
	})
}
