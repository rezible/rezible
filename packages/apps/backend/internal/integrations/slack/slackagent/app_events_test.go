package slackagent

import (
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/google/uuid"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/stretchr/testify/mock"

	"github.com/rezible/rezible/ent"
	slackintegration "github.com/rezible/rezible/internal/integrations/slack"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/test/mocks"
)

func (s *AppSuite) TestUserHomeView() {
	ctx := execution.NewTenantContext(s.T().Context(), 7)

	s.Run("a known Slack user", func() {
		known := &ent.User{ID: uuid.New(), TenantID: 7}
		users := mocks.NewMockUserService(s.T())
		users.EXPECT().
			Get(mock.Anything, mock.Anything).
			Return(known, nil)
		app := &App{users: users}

		view, viewErr := app.userHomeView(ctx, "U1")

		s.Require().NoError(viewErr)
		s.Equal(known.ID.String(), view.ExternalID)
	})

	s.Run("a Slack user the member sync has not seen", func() {
		users := mocks.NewMockUserService(s.T())
		users.EXPECT().
			Get(mock.Anything, mock.Anything).
			Return(nil, &ent.NotFoundError{})
		app := &App{users: users}

		view, viewErr := app.userHomeView(ctx, "U2")

		s.Require().NoError(viewErr)
		s.Equal(slack.VTHomeTab, view.Type)
		s.Equal(slackintegration.UnknownUserBlocks(), view.Blocks.BlockSet)
	})
}

func (s *AppSuite) TestMentionFromUnknownUserIsToldWhy() {
	var ephemeral url.Values
	handler := func(w http.ResponseWriter, r *http.Request) {
		s.Equal("/chat.postEphemeral", r.URL.Path)
		s.NoError(r.ParseForm())
		ephemeral = r.PostForm
		w.Header().Set("Content-Type", "application/json")
		_, writeErr := w.Write([]byte(`{"ok": true, "message_ts": "1700000000.000002"}`))
		s.NoError(writeErr)
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	s.T().Cleanup(server.Close)

	users := mocks.NewMockUserService(s.T())
	users.EXPECT().
		Get(mock.Anything, mock.Anything).
		Return(nil, &ent.NotFoundError{})
	app := &App{users: users}
	installation := &ent.Integration{
		Name:               integrationName,
		InstallationConfig: []byte(`{"AccessToken": "xoxb-test", "BotUserID": "B1", "Team": {"Id": "T1"}}`),
	}
	cw, clientErr := slackintegration.NewClientWrapper(installation, slack.OptionAPIURL(server.URL+"/"))
	s.Require().NoError(clientErr)
	mention := &slackevents.AppMentionEvent{
		User:      "U2",
		Channel:   "C1",
		TimeStamp: "1700000000.000001",
		Text:      "<@B1> what is happening?",
	}

	mentionErr := app.onMentionEvent(execution.NewTenantContext(s.T().Context(), 7), cw, mention)

	s.Require().NoError(mentionErr)
	s.Require().NotNil(ephemeral, "the mentioning user is replied to")
	s.Equal("U2", ephemeral.Get("user"))
	s.Equal("C1", ephemeral.Get("channel"))
	s.Equal("1700000000.000001", ephemeral.Get("thread_ts"))
	s.Equal(slackintegration.UnknownUserText, ephemeral.Get("text"))
}
