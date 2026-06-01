package slackincidents

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/rezible/rezible/ent"
	im "github.com/rezible/rezible/ent/incidentmilestone"
	"github.com/rezible/rezible/internal/integrations/slack"
	goslack "github.com/slack-go/slack"
)

const (
	actionCallbackIdIncidentDetailsModalButton   = "open_incident_details_modal"
	actionCallbackIdIncidentMilestoneModalButton = "open_incident_milestone_modal"
)

type incidentDetailsMessageBuilder struct {
	appUrl   string
	blocks   []goslack.Block
	incident *ent.Incident
}

var incidentDetailsTemplate = template.Template{}

func newIncidentDetailsMessageBuilder(appUrl string, inc *ent.Incident) *incidentDetailsMessageBuilder {
	return &incidentDetailsMessageBuilder{
		appUrl:   appUrl,
		blocks:   []goslack.Block{},
		incident: inc,
	}
}

func (b *incidentDetailsMessageBuilder) build() []goslack.Block {
	b.makeDetailsText()
	b.makeActions()
	return b.blocks
}

func (b *incidentDetailsMessageBuilder) isDetailsMessage(msg *goslack.Message) bool {
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
	webLink := fmt.Sprintf("%s/incidents/%s", b.appUrl, b.incident.Slug)
	sev := b.incident.Edges.Severity
	detailsText := fmt.Sprintf(
		"*Incident Details*\n*Title:* %s\n*Severity:* %s\n*Status:* %s\n*Roles:*\n%s\n*Latest Update:* %s\n*Web:* %s",
		b.incident.Title,
		sev.Name,
		b.currentStatus(),
		b.roleSummary(),
		b.latestUpdateSummary(),
		webLink,
	)

	if vc := b.incident.Edges.GetPrimaryVideoConference(); vc != nil {
		detailsText += fmt.Sprintf("\n*Video Conference:* %s", vc.JoinURL)
	}

	b.blocks = append(b.blocks, goslack.NewSectionBlock(slack.MarkdownBlock(detailsText), nil, nil))
}

func (b *incidentDetailsMessageBuilder) makeActions() {
	detailsButton := goslack.NewButtonBlockElement(actionCallbackIdIncidentDetailsModalButton, "details",
		slack.PlainTextBlock("Update Incident Details :gear:"))
	milestoneButton := goslack.NewButtonBlockElement(actionCallbackIdIncidentMilestoneModalButton, "milestone",
		slack.PlainTextBlock("Update Incident Status"))
	b.blocks = append(b.blocks, goslack.NewActionBlock("incident_details_actions", detailsButton, milestoneButton))
}

type incidentAnnouncementMessageBuilder struct {
	blocks   []goslack.Block
	incident *ent.Incident
	builder  *incidentDetailsMessageBuilder
}

func newIncidentAnnouncementMessageBuilder(appUrl string, inc *ent.Incident) *incidentAnnouncementMessageBuilder {
	return &incidentAnnouncementMessageBuilder{
		blocks:   []goslack.Block{},
		incident: inc,
		builder:  newIncidentDetailsMessageBuilder(appUrl, inc),
	}
}

func (b *incidentAnnouncementMessageBuilder) build() []goslack.Block {
	sev := b.incident.Edges.Severity
	headerText := fmt.Sprintf(":rotating_light: Incident declared in <#%s> [%s]", b.incident.ChatChannelID, sev.Name)

	contextText := fmt.Sprintf("*%s*\nStatus: %s", b.incident.Title, b.builder.currentStatus())

	b.blocks = []goslack.Block{
		goslack.NewSectionBlock(slack.MarkdownBlock(headerText), nil, nil),
		goslack.NewSectionBlock(slack.MarkdownBlock(contextText), nil, nil),
	}

	return b.blocks
}
