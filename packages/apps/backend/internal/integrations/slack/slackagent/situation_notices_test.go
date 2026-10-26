package slackagent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/slack-go/slack"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/schema/schematypes"
	sit "github.com/rezible/rezible/ent/situation"
	sitlink "github.com/rezible/rezible/ent/situationlink"
	slackintegration "github.com/rezible/rezible/internal/integrations/slack"
	"github.com/rezible/rezible/pkg/integrations"
	"github.com/rezible/rezible/pkg/projections"
)

var noticeTestOpenedAt = time.Date(2026, 10, 16, 9, 30, 0, 0, time.UTC)

const noticeTestSituationURL = "https://rezible.test/situations/SIT-12"

func (s *AppSuite) noticeService(name string) *ent.KnowledgeEntity {
	return &ent.KnowledgeEntity{
		ID:    uuid.New(),
		State: schematypes.KnowledgeGraphSubjectState{DisplayName: name},
	}
}

// noticeContent is a raised situation's notice content, with a report and no deployment.
func (s *AppSuite) noticeContent() *noticeContent {
	situation := &ent.Situation{
		ID:       uuid.New(),
		Number:   12,
		Title:    "Checkout errors",
		OpenedAt: noticeTestOpenedAt,
		RaisedAt: new(noticeTestOpenedAt.Add(10 * time.Minute)),
	}
	report := &ent.InvestigationReport{
		Summary:   "Errors began when the payments dependency started timing out.",
		CreatedAt: noticeTestOpenedAt.Add(20 * time.Minute),
	}
	return &noticeContent{
		Situation:    situation,
		Services:     []*ent.KnowledgeEntity{s.noticeService("Checkout"), s.noticeService("Payments")},
		Report:       report,
		SituationURL: noticeTestSituationURL,
		Now:          noticeTestOpenedAt.Add(30 * time.Minute),
	}
}

// renderedText is every text the notice's message shows, including its notification text.
func (s *AppSuite) renderedText(content *noticeContent) string {
	texts := []string{content.fallbackText()}
	for _, object := range s.textObjects(content) {
		texts = append(texts, object.text)
	}
	return strings.Join(texts, "\n")
}

// markdownText is the notice's text that Slack reads as markup: its mrkdwn text and notification text.
func (s *AppSuite) markdownText(content *noticeContent) string {
	texts := []string{content.fallbackText()}
	for _, object := range s.textObjects(content) {
		if object.kind == slack.MarkdownType {
			texts = append(texts, object.text)
		}
	}
	return strings.Join(texts, "\n")
}

type noticeTextObject struct {
	kind     string
	text     string
	verbatim bool
}

// textObjects are the texts in the notice's blocks, with their text type; a button's URL has none.
func (s *AppSuite) textObjects(content *noticeContent) []noticeTextObject {
	encoded, encodeErr := json.Marshal(content.blocks())
	s.Require().NoError(encodeErr)
	var decoded any
	s.Require().NoError(json.Unmarshal(encoded, &decoded))

	var objects []noticeTextObject
	var collect func(value any)
	collect = func(value any) {
		switch typed := value.(type) {
		case map[string]any:
			kind, _ := typed["type"].(string)
			verbatim, _ := typed["verbatim"].(bool)
			for key, field := range typed {
				if text, isText := field.(string); isText && (key == "text" || key == "url") {
					objects = append(objects, noticeTextObject{kind: kind, text: text, verbatim: verbatim})
					continue
				}
				collect(field)
			}
		case []any:
			for _, item := range typed {
				collect(item)
			}
		}
	}
	collect(decoded)
	return objects
}

// headerText is the text of the notice's header block.
func (s *AppSuite) headerText(content *noticeContent) *slack.TextBlockObject {
	header, isHeader := content.blocks()[0].(*slack.HeaderBlock)
	s.Require().True(isHeader, "the notice starts with a header")
	return header.Text
}

func (s *AppSuite) TestNoticeWording() {
	survivorID := uuid.New()
	survivorURL := "https://rezible.test/situations/SIT-7"
	stages := []struct {
		name  string
		setup func(content *noticeContent)
		stage string
	}{
		{
			name:  "raised",
			setup: func(*noticeContent) {},
			stage: "Raised",
		},
		{
			name: "muted",
			setup: func(content *noticeContent) {
				content.Situation.MutedAt = new(noticeTestOpenedAt.Add(15 * time.Minute))
				content.Situation.MuteReason = new(sit.MuteReasonExpected)
			},
			stage: "Muted (expected)",
		},
		{
			name: "held",
			setup: func(content *noticeContent) {
				content.Situation.HoldUntil = new(noticeTestOpenedAt.Add(2 * time.Hour))
			},
			stage: "Held until <!date^",
		},
		{
			name: "closed",
			setup: func(content *noticeContent) {
				content.Situation.ClosedAt = new(noticeTestOpenedAt.Add(time.Hour))
				content.Situation.CloseReason = new(sit.CloseReasonStabilized)
			},
			stage: "Closed: signals stopped",
		},
		{
			name: "dismissed",
			setup: func(content *noticeContent) {
				content.Situation.ClosedAt = new(noticeTestOpenedAt.Add(time.Hour))
				content.Situation.CloseReason = new(sit.CloseReasonDismissed)
			},
			stage: "Closed: dismissed after it was muted",
		},
		{
			name: "merged",
			setup: func(content *noticeContent) {
				content.Situation.ClosedAt = new(noticeTestOpenedAt.Add(time.Hour))
				content.Situation.CloseReason = new(sit.CloseReasonMerged)
				link := &ent.SituationLink{
					Kind:              sitlink.KindMergedInto,
					LinkedSituationID: survivorID,
					Edges: ent.SituationLinkEdges{
						LinkedSituation: &ent.Situation{ID: survivorID, Number: 7, Title: "Payments timeouts"},
					},
				}
				content.Situation.Edges.Links = []*ent.SituationLink{link}
				content.MergedIntoURL = survivorURL
			},
			stage: "Merged into <" + survivorURL + "|SIT-7 · Payments timeouts>",
		},
	}
	for _, stage := range stages {
		s.Run(stage.name, func() {
			content := s.noticeContent()
			stage.setup(content)

			text := s.renderedText(content)

			s.Contains(text, "SIT-12 · Checkout errors · "+stage.stage, "the notification text leads with the reference")
			s.Contains(text, stage.stage)
			s.Contains(text, "Checkout, Payments")
			s.Contains(text, slackintegration.SlackDate(noticeTestOpenedAt))
			s.Contains(text, noticeTestSituationURL)
			s.Contains(text, "Rezible's interpretation, as of "+slackintegration.SlackDate(content.Report.CreatedAt)+":")
			s.Contains(text, "Errors began when the payments dependency started timing out.")
			for _, claim := range []string{"healthy", "resolved", "fixed", "safe"} {
				s.NotContains(strings.ToLower(text), claim)
			}
		})
	}

	s.Run("without a report", func() {
		content := s.noticeContent()
		content.Report = nil

		s.Contains(s.renderedText(content), "*Rezible's interpretation:* Investigation in progress")
	})
}

func (s *AppSuite) TestNoticeWordingOfDeployments() {
	deployedAt := noticeTestOpenedAt.Add(-40 * time.Minute)
	deployment := func(status string) *rez.SituationLatestDeployment {
		return &rez.SituationLatestDeployment{
			Service:     s.noticeService("Checkout"),
			DeployedAt:  deployedAt,
			Environment: "production",
			Status:      status,
			Version:     "v2.4.1",
			Sha:         "4f9c2a1b7d3e",
		}
	}

	s.Run("succeeded", func() {
		content := s.noticeContent()
		content.Deployment = deployment(projections.DeploymentStatusSucceeded)

		text := s.renderedText(content)

		expected := "Last deployment before it started: Checkout · production · v2.4.1 · " + slackintegration.SlackDate(deployedAt)
		s.Contains(text, expected)
		s.NotContains(text, "<http", "the deployment is not a link")
	})

	s.Run("failed", func() {
		content := s.noticeContent()
		content.Deployment = deployment(projections.DeploymentStatusFailed)

		s.Contains(s.renderedText(content), "Checkout · production · v2.4.1 · failed · ")
	})

	s.Run("without a version", func() {
		content := s.noticeContent()
		content.Deployment = deployment(projections.DeploymentStatusSucceeded)
		content.Deployment.Version = ""

		s.Contains(s.renderedText(content), "Checkout · production · 4f9c2a1 · ")
	})

	s.Run("none", func() {
		content := s.noticeContent()

		s.NotContains(s.renderedText(content), "Last deployment")
	})
}

func (s *AppSuite) TestNoticeEscapesForeignText() {
	const foreign = "<!channel> <@U123> <https://evil|click> @here #general https://evil.example"
	content := s.noticeContent()
	content.Situation.Title = "Checkout " + foreign
	content.Services = []*ent.KnowledgeEntity{s.noticeService("Checkout " + foreign)}
	content.Deployment = &rez.SituationLatestDeployment{
		Service:     s.noticeService("Checkout"),
		DeployedAt:  noticeTestOpenedAt.Add(-time.Hour),
		Environment: "production " + foreign,
		Version:     "v1 " + foreign,
	}
	content.Report.Summary = "Errors " + foreign

	markdown := s.markdownText(content)

	for _, markup := range []string{"<!channel>", "<@U123>", "<https://evil|click>"} {
		s.NotContains(markdown, markup)
	}
	escaped := "&lt;!channel&gt; &lt;@U123&gt; &lt;https://evil|click&gt; @here #general https://evil.example"
	s.Equal(5, strings.Count(markdown, escaped),
		"the notification text, service, environment, version and summary are escaped")

	// Slack would otherwise link a bare channel name or URL and parse a bare mention by itself.
	for _, object := range s.textObjects(content) {
		if object.kind == slack.MarkdownType {
			s.True(object.verbatim, "mrkdwn is shown verbatim: %q", object.text)
		}
	}
	_, sent, applyErr := slack.UnsafeApplyMsgOptions("xoxb-test", "C0AAA1", "https://slack.test/api/", content.messageOptions()...)
	s.Require().NoError(applyErr)
	s.Equal("none", sent.Get("parse"), "the notification text is not parsed")
	s.Contains(sent.Get("text"), escaped)

	header := s.headerText(content)
	s.Equal(slack.PlainTextType, header.Type, "plain text cannot form mentions or links")
	s.Equal("SIT-12 · Checkout "+foreign, header.Text)
}

func (s *AppSuite) TestNoticeHeaderFitsSlackLimit() {
	content := s.noticeContent()
	content.Situation.Title = strings.Repeat("Checkout errors & timeouts ", 10)

	header := s.headerText(content)

	s.Len([]rune(header.Text), 150)
	s.True(strings.HasPrefix(header.Text, "SIT-12 · "), "cutting the title keeps the reference")
	s.True(strings.HasSuffix(header.Text, "…"))
}

func (s *AppSuite) TestValidateUserSettings() {
	intg := &Integration{}

	s.NoError(intg.ValidateUserSettings([]byte(`{}`)))
	s.NoError(intg.ValidateUserSettings([]byte(`{"situation_notice_channel_id": ""}`)))
	s.NoError(intg.ValidateUserSettings([]byte(`{"situation_notice_channel_id": "C0123ABC"}`)))
	s.NoError(intg.ValidateUserSettings([]byte(`{"situation_notice_channel_id": "G0123ABC"}`)))

	for _, invalid := range []string{`"#incidents"`, `"c0123abc"`, `"U0123ABC"`, `"C"`, `42`} {
		settings := []byte(`{"situation_notice_channel_id": ` + invalid + `}`)
		s.Error(intg.ValidateUserSettings(settings), "%s is not a channel ID", invalid)
	}
}

func (s *AppSuite) TestInstalledIntegrationListsChatChannels() {
	handler := func(w http.ResponseWriter, r *http.Request) {
		s.Equal("/conversations.list", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, writeErr := w.Write([]byte(`{"ok": true, "channels": [{"id": "C456", "name": "incidents"}]}`))
		s.NoError(writeErr)
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	s.T().Cleanup(server.Close)

	cfg := s.Config()
	cfg.Integrations.Slack.ApiUrl = server.URL + "/"
	app := MakeApp(cfg, nil, nil, nil, nil, nil, nil, nil)
	intg, intgErr := app.MakeIntegration(&slackintegration.AppServiceDependencies{})
	s.Require().NoError(intgErr)
	installation := &ent.Integration{
		Name:               integrationName,
		InstallationConfig: []byte(`{"AccessToken": "xoxb-test", "BotUserID": "B1", "Team": {"Id": "T1"}}`),
	}
	installed, installedErr := intg.GetInstalledIntegration(installation)
	s.Require().NoError(installedErr)

	querier, supportedErr := integrations.As[integrations.ChatChannelQuerier](installed)
	s.Require().NoError(supportedErr)
	page, listErr := querier.ListChatChannels(s.T().Context(), integrations.ListChatChannelsParams{})

	s.Require().NoError(listErr)
	s.Equal([]integrations.ChatChannel{{ID: "C456", Name: "incidents"}}, page.Channels)
}
