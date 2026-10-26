package slackincidents

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/predicate"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/rezible/rezible/test"
	"github.com/rezible/rezible/test/mocks"
)

type AppSuite struct {
	test.Suite
}

func TestAppSuite(t *testing.T) {
	suite.Run(t, &AppSuite{Suite: test.NewSuite()})
}

const (
	testInstallChannel = "C0INSTALL"
	testBotUser        = "UBOT"
)

var testInstallationConfig = fmt.Sprintf(`{"AccessToken": "xoxb-test", "BotUserID": %q, "WebhookChannelId": %q, "Team": {"Id": "T1"}}`,
	testBotUser, testInstallChannel)

// fakeSlack records the Web API calls the app makes. It cannot list pins, so a call to list them fails.
type fakeSlack struct {
	mu       sync.Mutex
	channels []string
	posts    []url.Values
	updates  []url.Values
	pinned   []string
	messages int
}

func (f *fakeSlack) serve(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if parseErr := r.ParseForm(); parseErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	response := map[string]any{"ok": true}
	switch strings.TrimPrefix(r.URL.Path, "/") {
	case "conversations.create":
		f.channels = append(f.channels, r.PostForm.Get("name"))
		response["channel"] = map[string]any{"id": "C0RESP", "name": r.PostForm.Get("name")}
	case "chat.postMessage", "chat.postEphemeral":
		f.posts = append(f.posts, r.PostForm)
		f.messages++
		response["channel"] = r.PostForm.Get("channel")
		response["ts"] = fmt.Sprintf("1700000000.%06d", f.messages)
		response["message_ts"] = response["ts"]
	case "chat.update":
		f.updates = append(f.updates, r.PostForm)
		response["channel"] = r.PostForm.Get("channel")
		response["ts"] = r.PostForm.Get("ts")
	case "pins.add":
		f.pinned = append(f.pinned, r.PostForm.Get("timestamp"))
	case "conversations.info":
		response["channel"] = map[string]any{"id": r.PostForm.Get("channel"), "topic": map[string]any{"value": ""}}
	case "conversations.setTopic":
		response["channel"] = map[string]any{"id": r.PostForm.Get("channel")}
	case "bookmarks.list":
		response["bookmarks"] = []any{}
	case "bookmarks.add":
		response["bookmark"] = map[string]any{"id": "B1"}
	case "conversations.members":
		response["members"] = []string{}
	case "conversations.invite":
		response["channel"] = map[string]any{"id": r.PostForm.Get("channel")}
	default:
		response = map[string]any{"ok": false, "error": "unknown_method " + r.URL.Path}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (f *fakeSlack) postsTo(channel string) []url.Values {
	f.mu.Lock()
	defer f.mu.Unlock()
	var posts []url.Values
	for _, post := range f.posts {
		if post.Get("channel") == channel {
			posts = append(posts, post)
		}
	}
	return posts
}

func (s *AppSuite) newApp(db rez.Database, jobs rez.JobService, orgs rez.OrganizationService, incidents rez.IncidentDetailsService, management rez.IncidentManagementService) (*App, *fakeSlack) {
	fake := &fakeSlack{}
	server := httptest.NewServer(http.HandlerFunc(fake.serve))
	s.T().Cleanup(server.Close)
	cfg := s.Config()
	cfg.Integrations.Slack.ApiUrl = server.URL + "/"
	return MakeApp(cfg, db, jobs, orgs, incidents, management), fake
}

func (s *AppSuite) TestMessagesKeepTheReferenceOfALongTitle() {
	inc := &ent.Incident{ID: uuid.New(), Number: 12, Title: strings.Repeat("Checkout errors ", 100)}
	messages := incidentMessages{inc: inc, incidentUrl: &url.URL{Scheme: "https", Host: "app.test"}}
	for prefix, options := range map[string][]slack.MsgOption{
		"Incident declared: INC-12 · Checkout errors": messages.announcementMessageOptions(),
		"Incident details: INC-12 · Checkout errors":  messages.detailsMessageOptions(),
	} {
		_, values, applyErr := slack.UnsafeApplyMsgOptions("", "C1", "", options...)
		s.Require().NoError(applyErr, prefix)
		s.True(strings.HasPrefix(values.Get("text"), prefix), values.Get("text"))
		s.Less(len([]rune(values.Get("text"))), len([]rune(inc.Title)), "the title is cut")
	}
}

func (s *AppSuite) TestDetailsModalDeclaresWithoutSeverityOrType() {
	userID := uuid.New()
	ctx := execution.NewUserContext(s.T().Context(), &ent.UserAuthSession{TenantID: 7, UserID: userID})
	declared := &ent.Incident{ID: uuid.New(), Number: 12, Title: "Checkout errors"}

	management := mocks.NewMockIncidentManagementService(s.T())
	management.EXPECT().
		Declare(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, params rez.DeclareIncidentParams) (*ent.Incident, error) {
			s.Equal("Checkout errors", params.Title)
			s.Nil(params.SeverityID)
			s.Nil(params.TypeID)
			s.Equal(userID, params.DeclaredByID)
			return declared, nil
		})
	app, fake := s.newApp(nil, nil, nil, nil, management)
	installation := &ent.Integration{Name: integrationName, InstallationConfig: []byte(testInstallationConfig)}
	submission := &slack.InteractionCallback{
		Type: slack.InteractionTypeViewSubmission,
		User: slack.User{ID: "U1"},
		View: slack.View{
			CallbackID:      viewCallbackIdIncidentDetailsModal,
			PrivateMetadata: `{"cid": "C0CMD"}`,
			State: &slack.ViewState{Values: map[string]map[string]slack.BlockAction{
				incidentModalTitleIds.Block:   {incidentModalTitleIds.Input: {Value: "Checkout errors"}},
				incidentModalSummaryIds.Block: {incidentModalSummaryIds.Input: {Value: ""}},
			}},
		},
	}

	viewErrors, submitErr := app.handleViewSubmissionInteraction(ctx, installation, submission)

	s.Require().NoError(submitErr)
	s.Empty(viewErrors)
	confirmations := fake.postsTo("C0CMD")
	s.Require().Len(confirmations, 1, "the declarer is told where they ran the command")
	s.Equal("U1", confirmations[0].Get("user"))
}

func (s *AppSuite) TestIncidentEventCreatesChannelAndPinsDetails() {
	ctx, db := s.SetupTestDatabase()
	createInstallation := db.Client(ctx).Integration.Create().
		SetProvider("slack").
		SetName(integrationName).
		SetDisplayName("Rezible").
		SetProviderInstallationRef("team:T1").
		SetInstallationConfig(json.RawMessage(testInstallationConfig))
	s.Require().NoError(createInstallation.Exec(ctx))

	declarerID := uuid.New()
	inc := &ent.Incident{
		ID:           uuid.New(),
		Number:       12,
		Title:        "Checkout errors after deploy",
		DeclaredByID: &declarerID,
		Edges:        ent.IncidentEdges{DeclaredBy: &ent.User{ID: declarerID, Name: "Ava", ChatID: "U0AVA"}},
	}
	incidents := mocks.NewMockIncidentDetailsService(s.T())
	incidents.EXPECT().
		GetIncident(mock.Anything, mock.Anything).
		RunAndReturn(func(context.Context, predicate.Incident) (*ent.Incident, error) {
			current := *inc
			return &current, nil
		})
	incidents.EXPECT().
		SetChatDetailsMessage(mock.Anything, inc.ID, mock.Anything).
		RunAndReturn(func(_ context.Context, _ uuid.UUID, messageID string) error {
			inc.ChatDetailsMessageID = messageID
			return nil
		})
	incidents.EXPECT().
		SetChatChannel(mock.Anything, inc.ID, "C0RESP").
		RunAndReturn(func(context.Context, uuid.UUID, string) error {
			inc.ChatChannelID = "C0RESP"
			return nil
		})
	orgs := mocks.NewMockOrganizationService(s.T())
	orgs.EXPECT().IsIncidentManagementEnabled(mock.Anything).Return(true, nil)
	var inserted []river.JobArgs
	jobs := mocks.NewMockJobService(s.T())
	jobs.EXPECT().
		Insert(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, args river.JobArgs, _ *river.InsertOpts) (*rivertype.JobInsertResult, error) {
			inserted = append(inserted, args)
			return &rivertype.JobInsertResult{}, nil
		})
	app, fake := s.newApp(db, jobs, orgs, incidents, nil)

	s.Require().NoError(app.onIncidentUpdated(ctx, &rez.EventOnIncidentUpdated{Created: true, IncidentId: inc.ID}))
	s.Require().Len(inserted, 1)
	s.IsType(syncIncidentChannelJobArgs{}, inserted[0])

	s.Require().NoError(app.syncIncidentChannel(ctx, inc.ID))
	s.Require().Len(inserted, 2)
	s.Equal(createIncidentChannelJobArgs{IncidentId: inc.ID}, inserted[1], "an incident without a channel gets one")

	s.Require().NoError(app.createIncidentChannel(ctx, inc.ID))
	s.Equal([]string{"inc-12"}, fake.channels, "the default pattern names the channel by number")
	s.Len(fake.postsTo(testInstallChannel), 1, "the announcement goes to the channel chosen at install")
	s.Require().Len(fake.postsTo("C0RESP"), 1, "the channel job posts the details message")
	s.Equal([]string{inc.ChatDetailsMessageID}, fake.pinned, "the details message is pinned and stored")

	for range 2 {
		s.Require().NoError(app.syncIncidentChannel(ctx, inc.ID), "a sync does not list pins")
	}
	s.Len(inserted, 2, "a sync with a stored details message needs no channel job")
	s.Len(fake.postsTo("C0RESP"), 1, "a sync posts nothing new")
	s.Require().Len(fake.updates, 2, "each sync edits the stored message")
	s.Equal(inc.ChatDetailsMessageID, fake.updates[1].Get("ts"))
}
