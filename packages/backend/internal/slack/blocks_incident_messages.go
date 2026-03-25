package slack

import (
	"fmt"
	"strings"
	"text/template"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	im "github.com/rezible/rezible/ent/incidentmilestone"
	"github.com/slack-go/slack"
)

const (
	actionCallbackIdIncidentDetailsModalButton   = "open_incident_details_modal"
	actionCallbackIdIncidentMilestoneModalButton = "open_incident_milestone_modal"
)

type incidentDetailsMessageBuilder struct {
	blocks   []slack.Block
	incident *ent.Incident
}

var incidentDetailsTemplate = template.Template{}

func newIncidentDetailsMessageBuilder(inc *ent.Incident) *incidentDetailsMessageBuilder {
	return &incidentDetailsMessageBuilder{
		blocks:   []slack.Block{},
		incident: inc,
	}
}

func (b *incidentDetailsMessageBuilder) build() []slack.Block {
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
	webLink := fmt.Sprintf("%s/incidents/%s", rez.Config.AppUrl(), b.incident.Slug)
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

	b.blocks = append(b.blocks, slack.NewSectionBlock(markdownText(detailsText), nil, nil))
}

func (b *incidentDetailsMessageBuilder) makeActions() {
	detailsButton := slack.NewButtonBlockElement(actionCallbackIdIncidentDetailsModalButton, "details",
		plainText("Update Incident Details :gear:"))
	milestoneButton := slack.NewButtonBlockElement(actionCallbackIdIncidentMilestoneModalButton, "milestone",
		plainText("Update Incident Status"))
	b.blocks = append(b.blocks, slack.NewActionBlock("incident_details_actions", detailsButton, milestoneButton))
}

type incidentAnnouncementMessageBuilder struct {
	blocks   []slack.Block
	incident *ent.Incident
}

func newIncidentAnnouncementMessageBuilder(inc *ent.Incident) *incidentAnnouncementMessageBuilder {
	return &incidentAnnouncementMessageBuilder{
		blocks:   []slack.Block{},
		incident: inc,
	}
}

func (b *incidentAnnouncementMessageBuilder) build() []slack.Block {
	sev := b.incident.Edges.Severity
	headerText := fmt.Sprintf(":rotating_light: Incident declared in <#%s> [%s]", b.incident.ChatChannelID, sev.Name)
	contextText := fmt.Sprintf("*%s*\nStatus: %s", b.incident.Title, newIncidentDetailsMessageBuilder(b.incident).currentStatus())

	b.blocks = []slack.Block{
		slack.NewSectionBlock(markdownText(headerText), nil, nil),
		slack.NewSectionBlock(markdownText(contextText), nil, nil),
	}

	return b.blocks
}
