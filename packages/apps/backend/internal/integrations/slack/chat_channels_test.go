package slackintegration

import (
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/slack-go/slack"

	"github.com/rezible/rezible/pkg/errs"
	"github.com/rezible/rezible/pkg/integrations"
)

// fakeConversationsAPI records the last conversations.list request and replies with a fixed response.
type fakeConversationsAPI struct {
	request url.Values
	status  int
	body    string
}

func (s *SlackIntegrationSuite) newFakeConversationsClient(api *fakeConversationsAPI) *slack.Client {
	handler := func(w http.ResponseWriter, r *http.Request) {
		s.Require().Equal("/conversations.list", r.URL.Path)
		s.Require().NoError(r.ParseForm())
		api.request = r.PostForm

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(api.status)
		_, writeErr := w.Write([]byte(api.body))
		s.Require().NoError(writeErr)
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	s.T().Cleanup(server.Close)

	return slack.New("xoxb-test", slack.OptionAPIURL(server.URL+"/"))
}

func (s *SlackIntegrationSuite) TestListChatChannelsReadsOnePublicPage() {
	api := &fakeConversationsAPI{
		status: http.StatusOK,
		body: `{
			"ok": true,
			"channels": [
				{"id": "C456", "name": "incident-announcements", "is_member": true},
				{"id": "C789", "name": "general"}
			],
			"response_metadata": {"next_cursor": "page-2"}
		}`,
	}
	client := s.newFakeConversationsClient(api)

	page, listErr := ListChatChannels(s.T().Context(), client, integrations.ListChatChannelsParams{})

	s.Require().NoError(listErr)
	s.Equal("public_channel", api.request.Get("types"))
	s.Equal("true", api.request.Get("exclude_archived"))
	s.Equal("100", api.request.Get("limit"))
	s.Empty(api.request.Get("cursor"))

	expected := []integrations.ChatChannel{
		{ID: "C456", Name: "incident-announcements", IsMember: true},
		{ID: "C789", Name: "general"},
	}
	s.Equal(expected, page.Channels)
	s.Equal("page-2", page.NextCursor)
}

func (s *SlackIntegrationSuite) TestListChatChannelsPassesParamsAndKeepsEmptyPageCursor() {
	api := &fakeConversationsAPI{
		status: http.StatusOK,
		body: `{
			"ok": true,
			"channels": [],
			"response_metadata": {"next_cursor": "page-3"}
		}`,
	}
	client := s.newFakeConversationsClient(api)
	params := integrations.ListChatChannelsParams{
		Cursor:         "page-2",
		Limit:          25,
		IncludePrivate: true,
	}

	page, listErr := ListChatChannels(s.T().Context(), client, params)

	s.Require().NoError(listErr)
	s.Equal("public_channel,private_channel", api.request.Get("types"))
	s.Equal("page-2", api.request.Get("cursor"))
	s.Equal("25", api.request.Get("limit"))

	s.NotNil(page.Channels)
	s.Empty(page.Channels)
	s.Equal("page-3", page.NextCursor)
}

func (s *SlackIntegrationSuite) TestListChatChannelsTranslatesErrors() {
	testCases := []struct {
		slackError string
		expected   error
	}{
		{slackError: "token_revoked", expected: errs.ErrForbidden},
		{slackError: "missing_scope", expected: errs.ErrForbidden},
		{slackError: "invalid_cursor", expected: errs.ErrInvalidInput},
	}
	for _, tc := range testCases {
		s.Run(tc.slackError, func() {
			api := &fakeConversationsAPI{
				status: http.StatusOK,
				body:   `{"ok": false, "error": "` + tc.slackError + `"}`,
			}
			client := s.newFakeConversationsClient(api)

			page, listErr := ListChatChannels(s.T().Context(), client, integrations.ListChatChannelsParams{})

			s.Nil(page)
			s.ErrorIs(listErr, tc.expected)
		})
	}

	s.Run("rate limited", func() {
		api := &fakeConversationsAPI{
			status: http.StatusTooManyRequests,
		}
		client := s.newFakeConversationsClient(api)

		page, listErr := ListChatChannels(s.T().Context(), client, integrations.ListChatChannelsParams{})

		s.Nil(page)
		s.ErrorIs(listErr, errs.ErrRateLimited)
		s.ErrorAs(listErr, new(*slack.RateLimitedError))
	})

	s.Run("other provider failure", func() {
		api := &fakeConversationsAPI{
			status: http.StatusOK,
			body:   `{"ok": false, "error": "internal_error"}`,
		}
		client := s.newFakeConversationsClient(api)

		page, listErr := ListChatChannels(s.T().Context(), client, integrations.ListChatChannelsParams{})

		s.Nil(page)
		s.ErrorContains(listErr, "internal_error")
		s.NotErrorIs(listErr, errs.ErrForbidden)
	})
}
