package slackagent

import (
	"github.com/slack-go/slack"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/internal/integrations/slack"
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
	b.makeMessageDetailsBlocks()
	b.makeNotesInputBlocks()
	return slack.Blocks{BlockSet: b.blocks}
}

func (b *annotationModalBuilder) makeMessageDetailsBlocks() {
	msgTime := b.metadata.MsgId.GetTimestamp().Unix()
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
	inputHint := slackintegration.PlainTextBlock("You can edit this later")
	if b.currAnnotation != nil {
		inputBlock.WithInitialValue(b.currAnnotation.Notes)
		inputHint = nil
	}

	b.blocks = append(b.blocks,
		slack.NewDividerBlock(),
		slack.NewInputBlock("notes_input", slackintegration.PlainTextBlock("Notes"), inputHint, inputBlock))
}
