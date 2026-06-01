package slackagent

import (
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/internal/integrations/slack"
	goslack "github.com/slack-go/slack"
)

type annotationModalBuilder struct {
	blocks         []goslack.Block
	currAnnotation *ent.EventAnnotation
	metadata       *annotationModalMetadata
}

func newAnnotationModalBuilder(curr *ent.EventAnnotation, meta *annotationModalMetadata) *annotationModalBuilder {
	return &annotationModalBuilder{
		blocks:         []goslack.Block{},
		currAnnotation: curr,
		metadata:       meta,
	}
}

func (b *annotationModalBuilder) build() goslack.Blocks {
	b.makeMessageDetailsBlocks()
	b.makeNotesInputBlocks()
	return goslack.Blocks{BlockSet: b.blocks}
}

func (b *annotationModalBuilder) makeMessageDetailsBlocks() {
	msgTime := b.metadata.MsgId.GetTimestamp().Unix()
	messageUserDetails := goslack.NewRichTextSection(
		goslack.NewRichTextSectionUserElement(b.metadata.UserId, nil),
		goslack.NewRichTextSectionDateElement(msgTime, " - {date_short_pretty} at {time}", nil, nil))
	messageContentsDetails := goslack.NewRichTextSection(
		goslack.NewRichTextSectionTextElement(b.metadata.MsgText, &goslack.RichTextSectionTextStyle{Italic: true}))

	b.blocks = append(b.blocks, goslack.NewRichTextBlock("anno_msg", messageUserDetails, messageContentsDetails))
}

func (b *annotationModalBuilder) makeNotesInputBlocks() {
	inputBlock := goslack.NewPlainTextInputBlockElement(nil, "notes_input_text")
	//inputBlock.WithMinLength(1)
	inputHint := slack.PlainTextBlock("You can edit this later")
	if b.currAnnotation != nil {
		inputBlock.WithInitialValue(b.currAnnotation.Notes)
		inputHint = nil
	}

	b.blocks = append(b.blocks,
		goslack.NewDividerBlock(),
		goslack.NewInputBlock("notes_input", slack.PlainTextBlock("Notes"), inputHint, inputBlock))
}
