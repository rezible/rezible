package slack

import (
	"fmt"
	"time"

	"github.com/slack-go/slack"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

/*
	TODO: eventually this should be handled by the chat service,
		and all the provider should do is convert rez.ContentNode
		to []slack.Block and send it
*/

func buildHandoverMessage(params rez.SendOncallHandoverParams) (string, slack.MsgOption, error) {
	mb, builderErr := newHandoverMessageBuilder(params.EndingShift, params.StartingShift, params.PinnedAnnotations)
	if builderErr != nil {
		return "", nil, fmt.Errorf("new builder: %w", builderErr)
	}
	if buildErr := mb.build(params.Content); buildErr != nil {
		return "", nil, fmt.Errorf("building message: %w", buildErr)
	}
	return mb.getChannel(), mb.getMessage(), nil
}

type handoverMessageBuilder struct {
	blocks []slack.Block

	roster            *ent.OncallRoster
	senderId          string
	receiverId        string
	endingShift       *ent.OncallShift
	startingShift     *ent.OncallShift
	pinnedAnnotations []*ent.OncallAnnotation
}

func newHandoverMessageBuilder(ending, starting *ent.OncallShift, pinnedAnnotations []*ent.OncallAnnotation) (*handoverMessageBuilder, error) {
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

func (b *handoverMessageBuilder) getMessage() slack.MsgOption {
	return slack.MsgOptionBlocks(b.blocks...)
}

func (b *handoverMessageBuilder) addBlocks(blocks ...slack.Block) {
	b.blocks = append(b.blocks, blocks...)
}

func (b *handoverMessageBuilder) build(content []rez.OncallShiftHandoverSection) error {
	b.blocks = make([]slack.Block, 0)

	// Header
	headerText := fmt.Sprintf(":pager: %s - Oncall Handover :pager:", b.roster.Name)
	headerObject := slack.NewTextBlockObject(slack.PlainTextType, headerText, true, false)

	usersBlock := slack.NewRichTextSection(
		slack.NewRichTextSectionUserElement(b.senderId, nil),
		slack.NewRichTextSectionTextElement(" to ", nil),
		slack.NewRichTextSectionUserElement(b.receiverId, nil))

	contextText := fmt.Sprintf("Shift Ending %s", b.endingShift.EndAt.Format(time.DateOnly))
	contextObject := slack.NewTextBlockObject(slack.MarkdownType, contextText, false, false)

	b.addBlocks(
		slack.NewHeaderBlock(headerObject, slack.HeaderBlockOptionBlockID("header")),
		slack.NewRichTextBlock("header_users", usersBlock),
		slack.NewContextBlock("header_time", contextObject))

	// Dynamic Sections
	b.addBlocks(slack.NewDividerBlock())
	for idx, s := range content {
		b.addBlocks(slack.NewHeaderBlock(plainTextBlock(s.Header)))

		if s.Kind == "annotations" {
			annoBlocks, annosErr := createPinnedAnnotationsBlocks(b.pinnedAnnotations)
			if annosErr != nil {
				return fmt.Errorf("failed to create annotations block: %w", annosErr)
			}
			b.addBlocks(annoBlocks...)
		} else if s.Kind == "regular" {
			b.addBlocks(convertContentToBlocks(s.Content, fmt.Sprintf("section_%d", idx))...)
		} else {
			return fmt.Errorf("unknown section kind '%s' for idx %d", s.Kind, idx)
		}
	}
	b.addBlocks(slack.NewDividerBlock())

	// Footer
	endingShiftLink := fmt.Sprintf("%s/oncall/shifts/%s", rez.FrontendUrl, b.endingShift.ID)
	footerEl := slack.NewRichTextSection(slack.NewRichTextSectionLinkElement(
		endingShiftLink, "View Full Shift Details in Rezible", nil))
	b.addBlocks(slack.NewRichTextBlock("handover_footer", footerEl))

	return nil
}

func createPinnedAnnotationsBlocks(annos []*ent.OncallAnnotation) ([]slack.Block, error) {
	if len(annos) == 0 {
		sectionBlock := slack.NewSectionBlock(plainTextBlock("No Pinned Annotations"), nil, nil)
		return []slack.Block{sectionBlock}, nil
	}

	var blocks []slack.Block

	for idx, a := range annos {
		blockId := fmt.Sprintf("pinned_annotation_%d", idx)

		ev, evErr := a.Edges.EventOrErr()
		if evErr != nil {
			return nil, fmt.Errorf("annotation event not loaded: %w", evErr)
		}

		var eventEls []slack.RichTextSectionElement
		if ev.Kind == "incident" {
			link := fmt.Sprintf("%s/incidents/%s", rez.FrontendUrl, ev.ID)
			eventEls = append(eventEls, slack.NewRichTextSectionLinkElement(link, ev.Title, nil))
		} else {
			eventEls = append(eventEls, slack.NewRichTextSectionTextElement(ev.Title, nil))
		}

		eventList := slack.NewRichTextList(slack.RTEListBullet, 0)
		for _, el := range eventEls {
			eventList.Elements = append(eventList.Elements, slack.NewRichTextSection(el))
		}

		style := &slack.RichTextSectionTextStyle{Italic: true}
		notesSection := slack.NewRichTextList(slack.RTEListBullet, 1,
			slack.NewRichTextSection(slack.NewRichTextSectionTextElement(a.Notes, style)))

		blocks = append(blocks,
			slack.NewRichTextBlock(blockId+"_events", eventList),
			slack.NewRichTextBlock(blockId, notesSection))
	}

	return blocks, nil
}
