package slackincidents

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/rezible/rezible/ent"
	im "github.com/rezible/rezible/ent/incidentmilestone"
	"github.com/rezible/rezible/internal/integrations/slack"
	"github.com/slack-go/slack"
)

const (
	actionCallbackIdIncidentDetailsModalButton   = "open_incident_details_modal"
	actionCallbackIdIncidentMilestoneModalButton = "open_incident_milestone_modal"
)

type incidentDetailsMessageBuilder struct {
	incident    *ent.Incident
	incidentUrl *url.URL
	blocks      []slack.Block
}

func newIncidentDetailsMessageBuilder(inc *ent.Incident, incUrl *url.URL) *incidentDetailsMessageBuilder {
	return &incidentDetailsMessageBuilder{
		incident:    inc,
		incidentUrl: incUrl,
		blocks:      make([]slack.Block, 0),
	}
}

func (b *incidentDetailsMessageBuilder) makeMessageBlocks() slack.MsgOption {
	return slack.MsgOptionBlocks(b.build()...)
}

func (b *incidentDetailsMessageBuilder) build() []slack.Block {
	b.blocks = make([]slack.Block, 0)
	b.makeDetailsText()
	b.makeActions()
	return b.blocks
}

func (b *incidentDetailsMessageBuilder) isDetailsMessage(msg *slack.Message) bool {
	return msg.Text != "" && strings.HasPrefix(msg.Text, "*Incident Details*")
}

func (b *incidentDetailsMessageBuilder) currentStatus() string {
	for _, milestone := range b.incident.Edges.Milestones {
		switch milestone.Kind {
		case im.KindResolution:
			return "RESOLVED"
		case im.KindMitigation:
			return "MITIGATED"
		case im.KindImpact:
			return "IMPACT"
		case im.KindOpened:
			return "OPEN"
		}
	}
	return "OPEN"
}

func (b *incidentDetailsMessageBuilder) roleSummary() string {
	if len(b.incident.Edges.RoleAssignments) == 0 {
		return "Unassigned"
	}
	parts := make([]string, 0, len(b.incident.Edges.RoleAssignments))
	for _, assignment := range b.incident.Edges.RoleAssignments {
		if assignment == nil || assignment.Edges.Role == nil || assignment.Edges.User == nil {
			continue
		}
		userRef := assignment.Edges.User.Name
		if assignment.Edges.User.ChatID != "" {
			userRef = fmt.Sprintf("<@%s>", assignment.Edges.User.ChatID)
		}
		parts = append(parts, fmt.Sprintf("%s: %s", assignment.Edges.Role.Name, userRef))
	}
	if len(parts) == 0 {
		return "Unassigned"
	}
	return strings.Join(parts, "\n")
}

func (b *incidentDetailsMessageBuilder) latestUpdateSummary() string {
	for _, milestone := range b.incident.Edges.Milestones {
		if milestone.Kind == im.KindOpened {
			continue
		}
		summary := strings.ToUpper(milestone.Kind.String())
		if milestone.Description != "" {
			summary += fmt.Sprintf(" - %s", milestone.Description)
		}
		return summary
	}
	return "No updates yet"
}

func (b *incidentDetailsMessageBuilder) makeDetailsText() {
	sev := b.incident.Edges.Severity
	detailsText := fmt.Sprintf(
		"*Incident Details*\n*Title:* %s\n*Severity:* %s\n*Status:* %s\n*Roles:*\n%s\n*Latest Update:* %s\n*Web:* %s",
		b.incident.Title,
		sev.Name,
		b.currentStatus(),
		b.roleSummary(),
		b.latestUpdateSummary(),
		b.incidentUrl.String(),
	)

	if vc := b.incident.Edges.GetPrimaryVideoConference(); vc != nil {
		detailsText += fmt.Sprintf("\n*Video Conference:* %s", vc.JoinURL)
	}

	b.blocks = append(b.blocks, slack.NewSectionBlock(slackintegration.MarkdownBlock(detailsText), nil, nil))
}

func (b *incidentDetailsMessageBuilder) makeActions() {
	detailsButton := slack.NewButtonBlockElement(actionCallbackIdIncidentDetailsModalButton, "details",
		slackintegration.PlainTextBlock("Update Incident Details :gear:"))
	milestoneButton := slack.NewButtonBlockElement(actionCallbackIdIncidentMilestoneModalButton, "milestone",
		slackintegration.PlainTextBlock("Update Incident Status"))
	b.blocks = append(b.blocks, slack.NewActionBlock("incident_details_actions", detailsButton, milestoneButton))
}

type incidentAnnouncementMessageBuilder struct {
	incident *ent.Incident
	builder  *incidentDetailsMessageBuilder
	blocks   []slack.Block
}

func newIncidentAnnouncementMessageBuilder(inc *ent.Incident, incUrl *url.URL) *incidentAnnouncementMessageBuilder {
	return &incidentAnnouncementMessageBuilder{
		incident: inc,
		builder:  newIncidentDetailsMessageBuilder(inc, incUrl),
		blocks:   []slack.Block{},
	}
}

func (b *incidentAnnouncementMessageBuilder) build() []slack.Block {
	sev := b.incident.Edges.Severity
	headerText := fmt.Sprintf(":rotating_light: Incident declared in <#%s> [%s]", b.incident.ChatChannelID, sev.Name)

	contextText := fmt.Sprintf("*%s*\nStatus: %s", b.incident.Title, b.builder.currentStatus())

	b.blocks = []slack.Block{
		slack.NewSectionBlock(slackintegration.MarkdownBlock(headerText), nil, nil),
		slack.NewSectionBlock(slackintegration.MarkdownBlock(contextText), nil, nil),
	}

	return b.blocks
}
