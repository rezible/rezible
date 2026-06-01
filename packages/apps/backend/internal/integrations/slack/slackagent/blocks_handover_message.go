package slackagent

import (
	"fmt"
	"time"

	"github.com/rezible/rezible/integrations/projections"

	"github.com/rezible/rezible/internal/integrations/slack"
	goslack "github.com/slack-go/slack"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

/*
	TODO: eventually this should be handled by the document service - converted to *rez.ContentNode
*/

//func buildHandoverMessage(params rez.SendOncallHandoverParams) (string, goslack.MsgOption, error) {
//	mb, builderErr := newHandoverMessageBuilder(params.EndingShift, params.StartingShift, params.PinnedAnnotations)
//	if builderErr != nil {
//		return "", nil, fmt.Errorf("new builder: %w", builderErr)
//	}
//	if buildErr := mb.build(params.Content); buildErr != nil {
//		return "", nil, fmt.Errorf("building message: %w", buildErr)
//	}
//	return mb.getChannel(), mb.getMessage(), nil
//}

type handoverMessageBuilder struct {
	blocks []goslack.Block

	appUrl            string
	roster            *ent.OncallRoster
	senderId          string
	receiverId        string
	endingShift       *ent.OncallShift
	startingShift     *ent.OncallShift
	pinnedAnnotations []*ent.EventAnnotation
}

func newHandoverMessageBuilder(ending, starting *ent.OncallShift, pinnedAnnotations []*ent.EventAnnotation) (*handoverMessageBuilder, error) {
	roster, rosterErr := ending.Edges.RosterOrErr()
	if rosterErr != nil {
		return nil, fmt.Errorf("get shift roster: %w", rosterErr)
	}
	if roster.ChatChannelID == "" {
		return nil, fmt.Errorf("no chat channel found for roster: %s", roster.ID)
	}

	sender, senderUserErr := ending.Edges.UserOrErr()
	if senderUserErr != nil {
		return nil, fmt.Errorf("get EndingShift user: %w", senderUserErr)
	}
	if sender.ChatID == "" {
		return nil, fmt.Errorf("no chat id for handover sender %s", sender.ID)
	}

	receiver, receiverUserErr := starting.Edges.UserOrErr()
	if receiverUserErr != nil {
		return nil, fmt.Errorf("get StartingShift user: %w", receiverUserErr)
	}
	if receiver.ChatID == "" {
		return nil, fmt.Errorf("no chat id for handover receiver %s", receiver.ID)
	}

	builder := &handoverMessageBuilder{
		roster:            roster,
		senderId:          sender.ChatID,
		receiverId:        receiver.ChatID,
		endingShift:       ending,
		startingShift:     starting,
		pinnedAnnotations: pinnedAnnotations,
	}

	return builder, nil
}

func (b *handoverMessageBuilder) getChannel() string {
	return b.roster.ChatChannelID
}

func (b *handoverMessageBuilder) addBlocks(blocks ...goslack.Block) {
	b.blocks = append(b.blocks, blocks...)
}

func (b *handoverMessageBuilder) build(content []rez.OncallShiftHandoverSection) error {
	b.blocks = make([]goslack.Block, 0)

	// Header
	headerText := fmt.Sprintf(":pager: %s - Oncall Handover :pager:", b.roster.Name)
	headerObject := goslack.NewTextBlockObject(goslack.PlainTextType, headerText, true, false)

	usersBlock := goslack.NewRichTextSection(
		goslack.NewRichTextSectionUserElement(b.senderId, nil),
		goslack.NewRichTextSectionTextElement(" to ", nil),
		goslack.NewRichTextSectionUserElement(b.receiverId, nil))

	contextText := fmt.Sprintf("Shift Ending %s", b.endingShift.EndAt.Format(time.DateOnly))
	contextObject := goslack.NewTextBlockObject(goslack.MarkdownType, contextText, false, false)

	b.addBlocks(
		goslack.NewHeaderBlock(headerObject, goslack.HeaderBlockOptionBlockID("header")),
		goslack.NewRichTextBlock("header_users", usersBlock),
		goslack.NewContextBlock("header_time", contextObject))

	// Dynamic Sections
	b.addBlocks(goslack.NewDividerBlock())
	for idx, s := range content {
		id := fmt.Sprintf("section_%d", idx)
		if sectionErr := b.addSection(id, s.Header, s.Kind, s.Content); sectionErr != nil {
			return fmt.Errorf("section %d: %w", idx, sectionErr)
		}
	}
	b.addBlocks(goslack.NewDividerBlock())

	// Footer
	endingShiftLink := fmt.Sprintf("%s/oncall/shifts/%s", b.appUrl, b.endingShift.ID)
	footerEl := goslack.NewRichTextSection(goslack.NewRichTextSectionLinkElement(
		endingShiftLink, "View Full Shift Details in Rezible", nil))
	b.addBlocks(goslack.NewRichTextBlock("handover_footer", footerEl))

	return nil
}

func (b *handoverMessageBuilder) addSection(id string, header, kind string, content *rez.ContentNode) error {
	b.addBlocks(goslack.NewHeaderBlock(slack.PlainTextBlock(header)))

	if kind == "annotations" {
		annoBlocks, annosErr := b.createPinnedAnnotationsBlocks()
		if annosErr != nil {
			return fmt.Errorf("failed to create annotations block: %w", annosErr)
		}
		b.addBlocks(annoBlocks...)
	} else if kind == "regular" {
		b.addBlocks(slack.ConvertContentToBlocks(id, content)...)
	} else {
		return fmt.Errorf("unknown section kind '%s'", kind)
	}

	return nil
}

func (b *handoverMessageBuilder) createPinnedAnnotationsBlocks() ([]goslack.Block, error) {
	if len(b.pinnedAnnotations) == 0 {
		sectionBlock := goslack.NewSectionBlock(slack.PlainTextBlock("No Pinned Annotations"), nil, nil)
		return []goslack.Block{sectionBlock}, nil
	}

	var blocks []goslack.Block

	for idx, a := range b.pinnedAnnotations {
		blockId := fmt.Sprintf("pinned_annotation_%d", idx)

		ev, evErr := a.Edges.EventOrErr()
		if evErr != nil {
			return nil, fmt.Errorf("annotation event not loaded: %w", evErr)
		}
		disp, dispErr := projections.GetEventDisplay(ev)
		if dispErr != nil {
			return nil, fmt.Errorf("annotation event display err: %w", dispErr)
		}

		var eventEls []goslack.RichTextSectionElement
		if projections.SubjectKindIncident.Matches(ev) {
			link := fmt.Sprintf("%s/incidents/%s", b.appUrl, ev.ID)
			eventEls = append(eventEls, goslack.NewRichTextSectionLinkElement(link, disp.Title, nil))
		} else {
			eventEls = append(eventEls, goslack.NewRichTextSectionTextElement(disp.Title, nil))
		}

		eventList := goslack.NewRichTextList(goslack.RTEListBullet, 0)
		for _, el := range eventEls {
			eventList.Elements = append(eventList.Elements, goslack.NewRichTextSection(el))
		}

		style := &goslack.RichTextSectionTextStyle{Italic: true}
		notesSection := goslack.NewRichTextList(goslack.RTEListBullet, 1,
			goslack.NewRichTextSection(goslack.NewRichTextSectionTextElement(a.Notes, style)))

		blocks = append(blocks,
			goslack.NewRichTextBlock(blockId+"_events", eventList),
			goslack.NewRichTextBlock(blockId, notesSection))
	}

	return blocks, nil
}
