package slack

import (
	"fmt"
	"time"

	"github.com/slack-go/slack"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/oncallusershiftannotation"
)

type handoverMessageBuilder struct {
	blocks []slack.Block

	roster        *ent.OncallRoster
	sender        *slack.User
	receiver      *slack.User
	endingShift   *ent.OncallUserShift
	startingShift *ent.OncallUserShift
	incidents     []*ent.Incident
	annotations   []*ent.OncallUserShiftAnnotation
}

func (b *handoverMessageBuilder) addBlocks(blocks ...slack.Block) {
	b.blocks = append(b.blocks, blocks...)
}

func (b *handoverMessageBuilder) getMessage() slack.MsgOption {
	return slack.MsgOptionBlocks(b.blocks...)
}

func (b *handoverMessageBuilder) build(content []rez.OncallShiftHandoverSection) error {
	b.blocks = make([]slack.Block, 0)

	if headerErr := b.buildHeader(); headerErr != nil {
		return fmt.Errorf("building header: %w", headerErr)
	}

	for i, section := range content {
		if sectionErr := b.buildSection(i, section); sectionErr != nil {
			return fmt.Errorf("building section %d: %w", i, sectionErr)
		}
	}

	if footerErr := b.buildFooter(); footerErr != nil {
		return fmt.Errorf("building footer: %w", footerErr)
	}

	return nil
}

func (b *handoverMessageBuilder) buildHeader() error {
	headerText := fmt.Sprintf(":pager: %s - Oncall Handover :pager:", b.roster.Name)
	headerObject := slack.NewTextBlockObject(slack.PlainTextType, headerText, true, false)

	usersBlock := slack.NewRichTextSection(
		slack.NewRichTextSectionUserElement(b.sender.ID, nil),
		slack.NewRichTextSectionTextElement(" to ", nil),
		slack.NewRichTextSectionUserElement(b.receiver.ID, nil))

	contextText := fmt.Sprintf("Shift Ending %s", b.endingShift.EndAt.Format(time.DateOnly))
	contextObject := slack.NewTextBlockObject(slack.MarkdownType, contextText, false, false)

	b.addBlocks(
		slack.NewHeaderBlock(headerObject, slack.HeaderBlockOptionBlockID("header")),
		slack.NewRichTextBlock("header_users", usersBlock),
		slack.NewContextBlock("header_time", contextObject),
		slack.NewDividerBlock())

	return nil
}

func (b *handoverMessageBuilder) buildSection(idx int, section rez.OncallShiftHandoverSection) error {
	sectionHeader := slack.NewTextBlockObject("plain_text", section.Header, false, false)
	b.addBlocks(slack.NewHeaderBlock(sectionHeader))

	switch section.Kind {
	case "annotations":
		return b.buildAnnotations()
	case "incidents":
		return b.buildIncidents()
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

func (b *handoverMessageBuilder) buildAnnotations() error {
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
		if anno.EventKind == oncallusershiftannotation.EventKindIncident {
			link := fmt.Sprintf("http://localhost:5173/incidents/%s", anno.EventID)
			el = slack.NewRichTextSectionLinkElement(link, anno.Title, nil)
		} else {
			el = slack.NewRichTextSectionTextElement(anno.Title, nil)
		}
		els = append(els, el)
		if anno.Notes != "" {
			flushList()
			addNotes(anno.Notes)
		}
	}
	if len(els) > 0 {
		flushList()
	}

	return nil
}

func (b *handoverMessageBuilder) buildIncidents() error {
	return nil
}

func (b *handoverMessageBuilder) buildFooter() error {
	return nil
}
