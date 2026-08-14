package slackagent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	ea "github.com/rezible/rezible/ent/eventannotation"
	ne "github.com/rezible/rezible/ent/normalizedevent"
	"github.com/rezible/rezible/pkg/execution"
	"github.com/slack-go/slack"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/internal/integrations/slack"
)

const viewCallbackIdAnnotationModal = "annotation_modal"

type annotationModalMetadata struct {
	UserId       string                     `json:"uid"`
	MsgId        slackintegration.MessageId `json:"mid"`
	MsgText      string                     `json:"mtx"`
	AnnotationId uuid.UUID                  `json:"aid,omitempty"`
}

func (a *App) makeAnnotationModalView(ctx context.Context, meta *annotationModalMetadata) (*slack.ModalViewRequest, error) {
	userId, userOk := execution.GetContext(ctx).UserID()
	if !userOk {
		return nil, fmt.Errorf("no user context")
	}

	lookupAnno := ea.And(
		ea.HasEventWith(ne.ProviderSubjectRef(meta.MsgId.String())),
		ea.CreatorID(userId))

	curr, currErr := a.events.QueryAnnotation(ctx, lookupAnno)
	if currErr != nil && !ent.IsNotFound(currErr) {
		return nil, fmt.Errorf("failed to lookup existing event annotation: %w", currErr)
	}
	if curr != nil {
		meta.AnnotationId = curr.ID
	}

	builder := newAnnotationModalBuilder(curr, meta)
	blockSet := builder.build()

	titleText := "Create Annotation"
	submitText := "Create"

	if curr != nil {
		titleText = "Update Annotation"
		submitText = "Update"
	}

	jsonMetadata, jsonErr := json.Marshal(meta)
	if jsonErr != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", jsonErr)
	}

	return &slack.ModalViewRequest{
		Type:            "modal",
		CallbackID:      viewCallbackIdAnnotationModal,
		Title:           slackintegration.PlainTextBlock(titleText),
		Submit:          slackintegration.PlainTextBlock(submitText),
		Close:           slackintegration.PlainTextBlock("Cancel"),
		Blocks:          blockSet,
		PrivateMetadata: string(jsonMetadata),
	}, nil
}

func (a *App) getAnnotationModalAnnotation(ctx context.Context, view slack.View) (*ent.EventAnnotation, error) {
	var meta annotationModalMetadata
	if jsonErr := json.Unmarshal([]byte(view.PrivateMetadata), &meta); jsonErr != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", jsonErr)
	}

	userId, userOk := execution.GetContext(ctx).UserID()
	if !userOk {
		return nil, fmt.Errorf("no user context")
	}

	var notes string
	if view.State != nil {
		if notesInput, inputOk := view.State.Values["notes_input"]; inputOk {
			if noteBlock, noteOk := notesInput["notes_input_text"]; noteOk {
				notes = noteBlock.Value
			}
		}
	}

	// TODO: convert this from event processor?
	ev := &ent.NormalizedEvent{
		ProviderSubjectRef: meta.MsgId.String(),
		//Kind:        "message",
		//Timestamp:   meta.MsgId.getTimestamp(),
		//Source:      "slack",
		//Title:       "Slack Message",
		//Description: meta.MsgText,
	}

	anno := &ent.EventAnnotation{
		ID:        meta.AnnotationId,
		CreatorID: userId,
		Notes:     notes,
		Edges: ent.EventAnnotationEdges{
			Event: ev,
		},
	}

	return anno, nil
}

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
