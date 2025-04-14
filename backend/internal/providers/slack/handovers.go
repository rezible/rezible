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

	roster        *ent.OncallRoster
	senderId      string
	receiverId    string
	endingShift   *ent.OncallUserShift
	startingShift *ent.OncallUserShift
	incidents     []*ent.Incident
	annotations   []*ent.OncallAnnotation
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

	switch section.Kind {
	case "annotations":
		b.addAnnotations()
	case "incidents":
		b.addIncidents()
	case "regular":
		{
			conv := &blockConverter{prefix: fmt.Sprintf("section_%d", idx)}
			contentBlocks := conv.convertDocument(section.Content)
			b.addBlocks(contentBlocks...)
		}
	default:
		return fmt.Errorf("unknown section kind: %s", section.Kind)
	}

	return nil
}

func (b *handoverMessageBuilder) addAnnotations() {
	if len(b.annotations) == 0 {
		b.addBlocks(slack.NewSectionBlock(plainText("No Annotations"), nil, nil))
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
		blockId := fmt.Sprintf("handover_annotation_list_%d", numListBlocks)
		b.addBlocks(slack.NewRichTextBlock(blockId, listEl))
		els = make([]slack.RichTextSectionElement, 0)
		numListBlocks++
	}

	addNotes := func(notes string) {
		style := &slack.RichTextSectionTextStyle{
			Italic: true,
		}
		section := slack.NewRichTextSection(slack.NewRichTextSectionTextElement(notes, style))
		blockId := fmt.Sprintf("handover_annotation_notes_%d", numNoteBlocks)
		numNoteBlocks++
		listBlock := slack.NewRichTextBlock(blockId, slack.NewRichTextList(slack.RTEListBullet, 1, section))
		b.addBlocks(listBlock)
	}

	for _, anno := range b.annotations {
		var el slack.RichTextSectionElement
		// TODO: get annotation event
		el = slack.NewRichTextSectionTextElement("Unknown Event", nil)
		/*
			if ev := anno.Event; ev != nil {
				if ev.Kind == "incident" {
					link := fmt.Sprintf("http://localhost:5173/incidents/%s", ev.ID)
					el = slack.NewRichTextSectionLinkElement(link, ev.Title, nil)
				} else {
					el = slack.NewRichTextSectionTextElement(ev.Title, nil)
				}
			} else {
				el = slack.NewRichTextSectionTextElement("Unknown Event", nil)
			}
		*/
		els = append(els, el)
		if anno.Notes != "" {
			flushList()
			addNotes(anno.Notes)
		}
	}
	if len(els) > 0 {
		flushList()
	}
}

func (b *handoverMessageBuilder) addIncidents() {
	if len(b.incidents) == 0 {
		b.addBlocks(slack.NewSectionBlock(plainText("No Incidents"), nil, nil))
		return
	}

	listEl := slack.NewRichTextList(slack.RTEListBullet, 0)
	for _, inc := range b.incidents {
		incLink := fmt.Sprintf("%s/incidents/%s", rez.FrontendUrl, inc.ID)
		el := slack.NewRichTextSectionLinkElement(incLink, inc.Title, nil)
		listEl.Elements = append(listEl.Elements, slack.NewRichTextSection(el))
	}
	b.addBlocks(slack.NewRichTextBlock("handover_incidents", listEl))
}

func (b *handoverMessageBuilder) addFooter() {
	endingShiftLink := fmt.Sprintf("%s/oncall/shifts/%s", rez.FrontendUrl, b.endingShift.ID)
	footerEl := slack.NewRichTextSection(slack.NewRichTextSectionLinkElement(
		endingShiftLink, "View Full Shift Details in Rezible", nil))
	b.addBlocks(slack.NewRichTextBlock("handover_footer", footerEl))
}
