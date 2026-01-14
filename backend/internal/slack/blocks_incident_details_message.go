package slack

import (
	"fmt"
	"strings"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/slack-go/slack"
)

const actionCallbackIdIncidentModalButton = "open_incident_modal"

type incidentDetailsMessageBuilder struct {
	blocks   []slack.Block
	incident *ent.Incident
}

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

func (b *incidentDetailsMessageBuilder) makeDetailsText() {
	webLink := fmt.Sprintf("%s/incidents/%s", rez.Config.AppUrl(), b.incident.Slug)
	sev := b.incident.Edges.Severity
	detailsText := fmt.Sprintf("*Incident Details*\n*Title:* %s\n*Severity:* %s\n*Status:* %s\n*Web:* %s",
		b.incident.Title, sev.Name, "OPEN", webLink)

	detailsTextBlock := &slack.TextBlockObject{
		Type: slack.MarkdownType,
		Text: detailsText,
	}

	b.blocks = append(b.blocks, slack.NewSectionBlock(detailsTextBlock, nil, nil))
}

func (b *incidentDetailsMessageBuilder) makeActions() {
	buttonText := slack.NewTextBlockObject(slack.PlainTextType, "Edit :gear:", true, false)
	openModalButton := slack.NewButtonBlockElement(actionCallbackIdIncidentModalButton, "open_modal", buttonText)
	openModalButton.Style = slack.StyleDefault
	b.blocks = append(b.blocks, slack.NewActionBlock("incident_details_actions", openModalButton))
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
	headerText := fmt.Sprintf(":rotating_light: Incident Declared: <#%s> [*%s*] :rotating_light:",
		b.incident.ChatChannelID, sev.Name)

	b.blocks = []slack.Block{
		slack.NewSectionBlock(
			&slack.TextBlockObject{
				Type: slack.MarkdownType,
				Text: headerText,
			},
			nil, nil,
		),
		slack.NewSectionBlock(
			&slack.TextBlockObject{
				Type: slack.MarkdownType,
				Text: fmt.Sprintf("_%s_", b.incident.Title),
			},
			nil, nil,
		),
	}

	return b.blocks
}
