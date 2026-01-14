package slack

import (
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/slack-go/slack"
)

type annotationModalBuilder struct {
	blocks         []slack.Block
	currAnnotation *ent.EventAnnotation
	metadata       *annotationModalMetadata
}

func newAnnotationModalBuilder(curr *ent.EventAnnotation, meta *annotationModalMetadata) *annotationModalBuilder {
	return &annotationModalBuilder{
		blocks:         []slack.Block{},
		currAnnotation: curr,
		metadata:       meta,
	}
}

func (b *annotationModalBuilder) build() slack.Blocks {
	b.makeMessageDetailsBLocks()
	b.makeNotesInputBlocks()
	return slack.Blocks{BlockSet: b.blocks}
}

func (b *annotationModalBuilder) makeMessageDetailsBLocks() {
	msgTime := b.metadata.MsgId.getTimestamp().Unix()
	messageUserDetails := slack.NewRichTextSection(
		slack.NewRichTextSectionUserElement(b.metadata.UserId, nil),
		slack.NewRichTextSectionDateElement(msgTime, " - {date_short_pretty} at {time}", nil, nil))
	messageContentsDetails := slack.NewRichTextSection(
		slack.NewRichTextSectionTextElement(b.metadata.MsgText, &slack.RichTextSectionTextStyle{Italic: true}))

	b.blocks = append(b.blocks, slack.NewRichTextBlock("anno_msg", messageUserDetails, messageContentsDetails))
}

func (b *annotationModalBuilder) makeNotesInputBlocks() {
	inputBlock := slack.NewPlainTextInputBlockElement(nil, "notes_input_text")
	//inputBlock.WithMinLength(1)
	inputHint := plainTextBlock("You can edit this later")
	if b.currAnnotation != nil {
		inputBlock.WithInitialValue(b.currAnnotation.Notes)
		inputHint = nil
	}

	b.blocks = append(b.blocks,
		slack.NewDividerBlock(),
		slack.NewInputBlock("notes_input", plainTextBlock("Notes"), inputHint, inputBlock))
}
