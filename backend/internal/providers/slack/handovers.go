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

type handoverMessageBuilder struct {
	blocks []slack.Block

	roster            *ent.OncallRoster
	senderId          string
	receiverId        string
	endingShift       *ent.OncallUserShift
	startingShift     *ent.OncallUserShift
	pinnedAnnotations []*ent.OncallAnnotation
}

func (b *handoverMessageBuilder) addBlocks(blocks ...slack.Block) {
	b.blocks = append(b.blocks, blocks...)
}

func (b *handoverMessageBuilder) getMessage() slack.MsgOption {
	return slack.MsgOptionBlocks(b.blocks...)
}

func (b *handoverMessageBuilder) build(content []rez.OncallShiftHandoverSection) error {
	b.blocks = make([]slack.Block, 0)

	b.addHeader()
	b.addBlocks(slack.NewDividerBlock())
	for i, section := range content {
		sectionErr := b.addSection(i, section)
		if sectionErr != nil {
			return fmt.Errorf("building section %d: %w", i, sectionErr)
		}
	}
	b.addBlocks(slack.NewDividerBlock())
	b.addFooter()

	return nil
}

func (b *handoverMessageBuilder) addHeader() {
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
}

func (b *handoverMessageBuilder) addSection(idx int, section rez.OncallShiftHandoverSection) error {
	sectionHeader := slack.NewTextBlockObject("plain_text", section.Header, false, false)
	b.addBlocks(slack.NewHeaderBlock(sectionHeader))

	if section.Kind == "annotations" {
		b.addPinnedAnnotations()
		return nil
	}

	sectionPfx := fmt.Sprintf("section_%d", idx)
	if section.Kind == "regular" {
		sectionBlocks := convertContentToBlocks(section.Content, &sectionPfx)
		b.addBlocks(sectionBlocks...)
		return nil
	}

	return fmt.Errorf("unknown section kind: %s", section.Kind)
}

func (b *handoverMessageBuilder) addPinnedAnnotations() {
	if len(b.pinnedAnnotations) == 0 {
		b.addBlocks(slack.NewSectionBlock(plainText("No Pinned Annotations"), nil, nil))
		return
	}

	numListBlocks := 0
	numNoteBlocks := 0

	var els []slack.RichTextSectionElement
	flushList := func() {
		listEl := slack.NewRichTextList(slack.RTEListBullet, 0)
		for _, el := range els {
			listEl.Elements = append(listEl.Elements, slack.NewRichTextSection(el))
		}
		blockId := fmt.Sprintf("handover_event_list_%d", numListBlocks)
		b.addBlocks(slack.NewRichTextBlock(blockId, listEl))
		els = make([]slack.RichTextSectionElement, 0)
		numListBlocks++
	}

	for _, a := range b.pinnedAnnotations {
		var el slack.RichTextSectionElement
		if ev := a.Edges.Event; ev != nil {
			if ev.Kind == "incident" {
				link := fmt.Sprintf("http://localhost:5173/incidents/%s", ev.ID)
				el = slack.NewRichTextSectionLinkElement(link, ev.Title, nil)
			} else {
				el = slack.NewRichTextSectionTextElement(ev.Title, nil)
			}
		} else {
			el = slack.NewRichTextSectionTextElement("Unknown Event", nil)
		}

		els = append(els, el)
		flushList()

		style := &slack.RichTextSectionTextStyle{
			Italic: true,
		}
		section := slack.NewRichTextSection(slack.NewRichTextSectionTextElement(a.Notes, style))
		blockId := fmt.Sprintf("handover_pinned_annotation_%d", numNoteBlocks)
		numNoteBlocks++
		annoBlock := slack.NewRichTextBlock(blockId, slack.NewRichTextList(slack.RTEListBullet, 1, section))
		b.addBlocks(annoBlock)
	}
	if len(els) > 0 {
		flushList()
	}
}

func (b *handoverMessageBuilder) addFooter() {
	endingShiftLink := fmt.Sprintf("%s/oncall/shifts/%s", rez.FrontendUrl, b.endingShift.ID)
	footerEl := slack.NewRichTextSection(slack.NewRichTextSectionLinkElement(
		endingShiftLink, "View Full Shift Details in Rezible", nil))
	b.addBlocks(slack.NewRichTextBlock("handover_footer", footerEl))
}
