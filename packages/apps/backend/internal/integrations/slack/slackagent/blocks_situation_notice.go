package slackagent

import (
	"fmt"
	"strings"

	"github.com/slack-go/slack"

	sit "github.com/rezible/rezible/ent/situation"
	slackintegration "github.com/rezible/rezible/internal/integrations/slack"
	"github.com/rezible/rezible/pkg/integrations"
	"github.com/rezible/rezible/pkg/projections"
)

const (
	actionIdSituationNoticeUseful    = "situation_notice_useful"
	actionIdSituationNoticeNotUseful = "situation_notice_not_useful"
	actionIdSituationNoticeOpen      = "situation_notice_open"
	// The incident buttons open their URLs: declaring an incident in Rezible, or the incident it is part of.
	actionIdSituationNoticeDeclareIncident = "situation_notice_declare_incident"
	actionIdSituationNoticeOpenIncident    = "situation_notice_open_incident"

	// maxNoticeSummaryRunes bounds the interpretation shown in a notice.
	maxNoticeSummaryRunes = 600
	// maxNoticeTitleRunes is Slack's limit on a header's text.
	maxNoticeTitleRunes = 150
	shortShaLength      = 7
)

// noticeContent is a notice's content, rendered as a Slack message.
type noticeContent integrations.SituationNoticeContent

// stageText says where the situation stands, without claiming that the problem is over.
func (c *noticeContent) stageText() string {
	situation := c.Situation
	if situation.ClosedAt != nil {
		if situation.CloseReason != nil && *situation.CloseReason == sit.CloseReasonMerged {
			if link := (*integrations.SituationNoticeContent)(c).MergedInto(); link != nil && c.MergedIntoURL != "" {
				title := "another situation"
				if link.Edges.LinkedSituation != nil {
					title = link.Edges.LinkedSituation.Title
				}
				return fmt.Sprintf("Merged into <%s|%s>", c.MergedIntoURL, slackintegration.EscapeText(title))
			}
			return "Merged into another situation"
		}
		return "Closed: " + c.closeReasonText()
	}
	if situation.MutedAt != nil {
		if situation.MuteReason != nil && *situation.MuteReason == sit.MuteReasonExpected {
			return "Muted (expected)"
		}
		return "Muted (not noteworthy)"
	}
	if situation.HoldUntil != nil && situation.HoldUntil.After(c.Now) {
		return "Held until " + slackintegration.SlackDate(*situation.HoldUntil)
	}
	return "Raised"
}

func (c *noticeContent) closeReasonText() string {
	if c.Situation.CloseReason == nil {
		return "closed"
	}
	switch *c.Situation.CloseReason {
	case sit.CloseReasonDismissed:
		return "dismissed after it was muted"
	case sit.CloseReasonExpired:
		return "no longer watched"
	default:
		return "signals stopped"
	}
}

func (c *noticeContent) servicesText() string {
	if len(c.Services) == 0 {
		return "None identified"
	}
	names := make([]string, len(c.Services))
	for i, service := range c.Services {
		names[i] = slackintegration.EscapeText(service.State.DisplayName)
	}
	return strings.Join(names, ", ")
}

// deploymentText states the deployment as plain text: its URL comes from its source, and only links Rezible
// builds are links. The label states an order in time, not a cause.
func (c *noticeContent) deploymentText() string {
	deployment := c.Deployment
	parts := make([]string, 0, 5)
	if deployment.Service != nil {
		parts = append(parts, slackintegration.EscapeText(deployment.Service.State.DisplayName))
	}
	if deployment.Environment != "" {
		parts = append(parts, slackintegration.EscapeText(deployment.Environment))
	}
	if deployment.Version != "" {
		parts = append(parts, slackintegration.EscapeText(deployment.Version))
	} else if sha := deployment.Sha; sha != "" {
		parts = append(parts, slackintegration.EscapeText(sha[:min(len(sha), shortShaLength)]))
	}
	if deployment.Status != "" && deployment.Status != projections.DeploymentStatusSucceeded {
		parts = append(parts, slackintegration.EscapeText(deployment.Status))
	}
	parts = append(parts, slackintegration.SlackDate(deployment.DeployedAt))
	return "Last deployment before it started: " + strings.Join(parts, " · ")
}

func (c *noticeContent) interpretationText() string {
	if c.Report == nil {
		return "*Rezible's interpretation:* Investigation in progress"
	}
	summary := strings.TrimSpace(c.Report.Summary)
	if summary == "" {
		summary = strings.TrimSpace(c.Report.Text)
	}
	summary = slackintegration.EscapeText(slackintegration.TruncateRunes(summary, maxNoticeSummaryRunes))
	return fmt.Sprintf("*Rezible's interpretation, as of %s:*\n%s", slackintegration.SlackDate(c.Report.CreatedAt), summary)
}

// fallbackText is the message's plain text, shown in notifications.
func (c *noticeContent) fallbackText() string {
	return fmt.Sprintf("%s · %s", slackintegration.EscapeText(c.Situation.Title), c.stageText())
}

func (c *noticeContent) blocks() []slack.Block {
	// Plain text cannot form mentions or links, so the header is not escaped.
	title := slackintegration.TruncateRunes(c.Situation.Title, maxNoticeTitleRunes)
	header := slack.NewHeaderBlock(slack.NewTextBlockObject(slack.PlainTextType, title, false, false))

	fields := []*slack.TextBlockObject{
		markdownText("*Stage*\n" + c.stageText()),
		markdownText("*Started*\n" + slackintegration.SlackDate(c.Situation.OpenedAt)),
		markdownText("*Services*\n" + c.servicesText()),
	}
	blocks := []slack.Block{
		header,
		slack.NewSectionBlock(nil, fields, nil),
	}
	if c.Deployment != nil {
		deploymentText := markdownText(c.deploymentText())
		blocks = append(blocks, slack.NewSectionBlock(deploymentText, nil, nil))
	}
	interpretation := markdownText(c.interpretationText())
	blocks = append(blocks, slack.NewSectionBlock(interpretation, nil, nil))

	situationID := c.Situation.ID.String()
	open := slack.NewButtonBlockElement(actionIdSituationNoticeOpen, situationID,
		slack.NewTextBlockObject(slack.PlainTextType, "Open in Rezible", false, false))
	open.URL = c.SituationURL
	actions := []slack.BlockElement{open}
	if c.IncidentURL != "" {
		openIncident := slack.NewButtonBlockElement(actionIdSituationNoticeOpenIncident, situationID,
			slack.NewTextBlockObject(slack.PlainTextType, "Open incident", false, false))
		openIncident.URL = c.IncidentURL
		actions = append(actions, openIncident)
	} else if c.DeclareIncidentURL != "" {
		declare := slack.NewButtonBlockElement(actionIdSituationNoticeDeclareIncident, situationID,
			slack.NewTextBlockObject(slack.PlainTextType, "Declare incident", false, false))
		declare.URL = c.DeclareIncidentURL
		actions = append(actions, declare)
	}
	useful := slack.NewButtonBlockElement(actionIdSituationNoticeUseful, situationID,
		slack.NewTextBlockObject(slack.PlainTextType, "Useful", false, false))
	notUseful := slack.NewButtonBlockElement(actionIdSituationNoticeNotUseful, situationID,
		slack.NewTextBlockObject(slack.PlainTextType, "Not useful", false, false))
	actions = append(actions, useful, notUseful)
	blocks = append(blocks, slack.NewActionBlock("situation_notice_actions", actions...))
	return blocks
}

// markdownText is mrkdwn that Slack shows verbatim: escaped source text cannot become a link or mention
// through Slack's automatic parsing, while Rezible's own <url|label> links and date tokens still work.
func markdownText(text string) *slack.TextBlockObject {
	return slack.NewTextBlockObject(slack.MarkdownType, text, false, true)
}

// messageOptions turn off Slack's parsing of the notification text, for posts and edits alike.
func (c *noticeContent) messageOptions() []slack.MsgOption {
	return []slack.MsgOption{
		slack.MsgOptionText(c.fallbackText(), false),
		slack.MsgOptionParse(false),
		slack.MsgOptionBlocks(c.blocks()...),
	}
}
