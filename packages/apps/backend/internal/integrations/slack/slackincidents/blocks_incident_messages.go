package slackincidents

import (
	"bytes"
	"cmp"
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/slack-go/slack"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	im "github.com/rezible/rezible/ent/incidentmilestone"
	"github.com/rezible/rezible/internal/integrations/slack"
)

const (
	actionCallbackIdIncidentDetailsModalButton   = "open_incident_details_modal"
	actionCallbackIdIncidentMilestoneModalButton = "open_incident_milestone_modal"
	actionCallbackIdIncidentOpen                 = "open_incident"

	// maxHeaderRunes is Slack's limit on a header's text.
	maxHeaderRunes = 150
	// Slack limits a section's text to 3000 characters and a field or context element to 2000. Text people
	// wrote is capped below those, leaving room for what Rezible adds around it.
	maxNameRunes        = 200
	maxTitleRunes       = 500
	maxDescriptionRunes = 1500
	maxRolesRunes       = 2900
)

// escaped is text people wrote, escaped and cut to the limit.
func escaped(text string, limit int) string {
	return slackintegration.TruncateRunes(slackintegration.EscapeText(text), limit)
}

// incidentMessages renders an incident's Slack messages. Text people wrote is escaped; only links Rezible
// builds, and mentions it builds from chat IDs, are links.
type incidentMessages struct {
	inc         *ent.Incident
	incidentUrl *url.URL
}

func (m incidentMessages) title() string {
	return escaped(m.inc.Title, maxTitleRunes)
}

func (m incidentMessages) severityText() string {
	if sev := m.inc.Edges.Severity; sev != nil {
		return escaped(sev.Name, maxNameRunes)
	}
	return "None"
}

func (m incidentMessages) typeText() string {
	if incType := m.inc.Edges.Type; incType != nil {
		return escaped(incType.Name, maxNameRunes)
	}
	return "None"
}

func (m incidentMessages) statusText() string {
	return responseStateLabels[m.inc.ResponseState]
}

var responseStateLabels = map[incident.ResponseState]string{
	incident.ResponseStateStarted:   "Started",
	incident.ResponseStateMitigated: "Mitigated",
	incident.ResponseStateResolved:  "Resolved",
}

func (m incidentMessages) declaredByText() string {
	if declarer := m.inc.Edges.DeclaredBy; declarer != nil {
		return mention(declarer)
	}
	return "Rezible"
}

// rolesText lists who holds each role, one line per role, ordered by role name.
func (m incidentMessages) rolesText() string {
	holders := map[string][]*ent.User{}
	for _, assignment := range m.inc.Edges.RoleAssignments {
		if assignment.Edges.Role == nil || assignment.Edges.User == nil {
			continue
		}
		role := assignment.Edges.Role.Name
		holders[role] = append(holders[role], assignment.Edges.User)
	}
	if len(holders) == 0 {
		return "No roles assigned"
	}
	roles := make([]string, 0, len(holders))
	for role := range holders {
		roles = append(roles, role)
	}
	slices.Sort(roles)
	lines := make([]string, len(roles))
	for i, role := range roles {
		users := holders[role]
		slices.SortFunc(users, func(a, b *ent.User) int {
			return cmp.Or(cmp.Compare(a.Name, b.Name), bytes.Compare(a.ID[:], b.ID[:]))
		})
		mentions := make([]string, len(users))
		for j, user := range users {
			mentions[j] = mention(user)
		}
		lines[i] = fmt.Sprintf("%s: %s", escaped(role, maxNameRunes), strings.Join(mentions, ", "))
	}
	return slackintegration.TruncateRunes(strings.Join(lines, "\n"), maxRolesRunes)
}

// latestUpdateText is the newest milestone of any kind; equal timestamps are ordered by ID.
func (m incidentMessages) latestUpdateText() string {
	milestones := ent.IncidentMilestones(m.inc.Edges.Milestones)
	latest := milestones.GetLatest()
	if latest == nil {
		return "No updates yet"
	}
	parts := []string{milestoneLabel(milestones, latest)}
	if description := strings.TrimSpace(latest.Description); description != "" {
		parts = append(parts, escaped(description, maxDescriptionRunes))
	}
	parts = append(parts, slackintegration.SlackDate(latest.Timestamp))
	return strings.Join(parts, " · ")
}

func (m incidentMessages) openButton() *slack.ButtonBlockElement {
	open := slack.NewButtonBlockElement(actionCallbackIdIncidentOpen, m.inc.ID.String(),
		slackintegration.PlainTextBlock("Open in Rezible"))
	open.URL = m.incidentUrl.String()
	return open
}

func (m incidentMessages) detailsMessageOptions() []slack.MsgOption {
	header := slack.NewHeaderBlock(slackintegration.PlainTextBlock(slackintegration.TruncateRunes(m.inc.Title, maxHeaderRunes)))
	fields := []*slack.TextBlockObject{
		slackintegration.MarkdownBlock("*Status*\n" + m.statusText()),
		slackintegration.MarkdownBlock("*Severity*\n" + m.severityText()),
		slackintegration.MarkdownBlock("*Type*\n" + m.typeText()),
		slackintegration.MarkdownBlock(fmt.Sprintf("*Declared*\n%s, %s", m.declaredByText(), slackintegration.SlackDate(m.inc.OpenedAt))),
	}
	details := slack.NewSectionBlock(nil, fields, nil)
	roles := slack.NewSectionBlock(slackintegration.MarkdownBlock("*Roles*\n"+m.rolesText()), nil, nil)
	latest := slack.NewSectionBlock(slackintegration.MarkdownBlock("*Latest update*\n"+m.latestUpdateText()), nil, nil)

	incidentID := m.inc.ID.String()
	detailsButton := slack.NewButtonBlockElement(actionCallbackIdIncidentDetailsModalButton, incidentID,
		slackintegration.PlainTextBlock("Update details"))
	statusButton := slack.NewButtonBlockElement(actionCallbackIdIncidentMilestoneModalButton, incidentID,
		slackintegration.PlainTextBlock("Record status"))
	actions := slack.NewActionBlock("incident_details_actions", detailsButton, statusButton, m.openButton())

	return []slack.MsgOption{
		slack.MsgOptionText("Incident details: "+m.title(), false),
		slack.MsgOptionParse(false),
		slack.MsgOptionBlocks(header, details, roles, latest, actions),
	}
}

func (m incidentMessages) announcementMessageOptions() []slack.MsgOption {
	headline := ":rotating_light: *Incident declared:* " + m.title()
	contextParts := []string{}
	if sev := m.inc.Edges.Severity; sev != nil {
		contextParts = append(contextParts, escaped(sev.Name, maxNameRunes))
	}
	contextParts = append(contextParts, m.statusText(), "declared by "+m.declaredByText())
	if m.inc.ChatChannelID != "" {
		contextParts = append(contextParts, fmt.Sprintf("<#%s>", m.inc.ChatChannelID))
	}
	blocks := []slack.Block{
		slack.NewSectionBlock(slackintegration.MarkdownBlock(headline), nil, nil),
		slack.NewContextBlock("", slackintegration.MarkdownBlock(strings.Join(contextParts, " · "))),
		slack.NewActionBlock("incident_announcement_actions", m.openButton()),
	}
	return []slack.MsgOption{
		slack.MsgOptionText("Incident declared: "+m.title(), false),
		slack.MsgOptionParse(false),
		slack.MsgOptionBlocks(blocks...),
	}
}

func (m incidentMessages) milestoneMessageOptions(ms *ent.IncidentMilestone) []slack.MsgOption {
	label := milestoneLabel(m.inc.Edges.Milestones, ms)
	headline := fmt.Sprintf("Rezible recorded *%s*", label)
	if ms.Edges.User != nil {
		headline = fmt.Sprintf("%s marked the incident *%s*", mention(ms.Edges.User), label)
	}
	blocks := []slack.Block{slack.NewSectionBlock(slackintegration.MarkdownBlock(headline), nil, nil)}
	if notes := strings.TrimSpace(ms.Description); notes != "" {
		blocks = append(blocks, slack.NewContextBlock("", slackintegration.MarkdownBlock(escaped(notes, maxDescriptionRunes))))
	}
	blocks = append(blocks, slack.NewContextBlock("", slackintegration.MarkdownBlock(slackintegration.SlackDate(ms.Timestamp))))
	return []slack.MsgOption{
		slack.MsgOptionText(fmt.Sprintf("%s: %s", m.title(), label), false),
		slack.MsgOptionParse(false),
		slack.MsgOptionBlocks(blocks...),
	}
}

// milestoneLabel names what the milestone recorded. An impact while mitigated records that the mitigation
// failed.
func milestoneLabel(milestones ent.IncidentMilestones, ms *ent.IncidentMilestone) string {
	switch ms.Kind {
	case im.KindImpact:
		previous := milestonesBefore(milestones, ms).GetLatest(im.KindImpact, im.KindMitigation)
		if previous != nil && previous.Kind == im.KindMitigation {
			return "Mitigation failed"
		}
		return "Still impacted"
	case im.KindMitigation:
		return "Mitigated"
	case im.KindResolution:
		return "Resolved"
	case im.KindDetection:
		return "Detected"
	case im.KindInvestigation:
		return "Investigating"
	}
	return string(ms.Kind)
}

// milestonesBefore are the milestones earlier than ms; equal timestamps are ordered by ID.
func milestonesBefore(milestones ent.IncidentMilestones, ms *ent.IncidentMilestone) ent.IncidentMilestones {
	before := make(ent.IncidentMilestones, 0, len(milestones))
	for _, other := range milestones {
		earlier := other.Timestamp.Before(ms.Timestamp) ||
			(other.Timestamp.Equal(ms.Timestamp) && bytes.Compare(other.ID[:], ms.ID[:]) < 0)
		if earlier {
			before = append(before, other)
		}
	}
	return before
}

// mention links a member with a Slack account, and names anyone else.
func mention(user *ent.User) string {
	if user.ChatID != "" {
		return fmt.Sprintf("<@%s>", user.ChatID)
	}
	return escaped(user.Name, maxNameRunes)
}
