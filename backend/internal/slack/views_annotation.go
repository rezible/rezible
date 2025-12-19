package slack

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/slack-go/slack"

	"github.com/rezible/rezible/ent"
)

const (
	annotationModalViewCallbackID = "create_annotation_confirm"
)

type (
	annotationViewContext struct {
		meta             annotationViewMetadata
		rosters          []*ent.OncallRoster
		selectedRosterId uuid.UUID
		currAnnotation   *ent.EventAnnotation
	}
	annotationViewMetadata struct {
		UserId       string    `json:"uid"`
		MsgId        messageId `json:"mid"`
		MsgText      string    `json:"mtx"`
		AnnotationId uuid.UUID `json:"aid,omitempty"`
	}
)

func (s *ChatService) makeAnnotationViewContext(ctx context.Context, ic *slack.InteractionCallback) (*annotationViewContext, error) {
	d := &annotationViewContext{}
	if ic.View.PrivateMetadata != "" {
		if jsonErr := json.Unmarshal([]byte(ic.View.PrivateMetadata), &d.meta); jsonErr != nil {
			return nil, jsonErr
		}
	} else {
		d.meta = annotationViewMetadata{
			MsgId:   getMessageId(ic),
			UserId:  ic.User.ID,
			MsgText: ic.Message.Text,
		}
	}

	usr, usrCtx, userErr := s.lookupChatUser(ctx, d.meta.UserId)
	if userErr != nil {
		return nil, fmt.Errorf("failed to lookup user: %w", userErr)
	}

	ev := &ent.Event{
		ExternalID: d.meta.MsgId.String(),
	}

	anno, annoErr := s.annos.LookupByUserEvent(usrCtx, usr.ID, ev)
	if annoErr != nil && !ent.IsNotFound(annoErr) {
		return nil, fmt.Errorf("failed to lookup existing event annotation: %w", annoErr)
	}
	d.currAnnotation = anno

	return d, nil
}

func makeAnnotationModalViewBlocks(c *annotationViewContext) []slack.Block {
	var blockSet []slack.Block

	messageUserDetails := slack.NewRichTextSection(
		slack.NewRichTextSectionUserElement(c.meta.UserId, nil),
		slack.NewRichTextSectionDateElement(c.meta.MsgId.getTimestamp().Unix(), " - {date_short_pretty} at {time}", nil, nil))
	messageContentsDetails := slack.NewRichTextSection(
		slack.NewRichTextSectionTextElement(c.meta.MsgText, &slack.RichTextSectionTextStyle{Italic: true}))

	blockSet = append(blockSet, slack.NewRichTextBlock("anno_msg", messageUserDetails, messageContentsDetails))

	inputBlock := slack.NewPlainTextInputBlockElement(nil, "notes_input_text")
	//inputBlock.WithMinLength(1)
	inputHint := plainTextBlock("You can edit this later")
	if c.currAnnotation != nil {
		inputBlock.WithInitialValue(c.currAnnotation.Notes)
		inputHint = nil
	}

	blockSet = append(blockSet,
		slack.NewDividerBlock(),
		slack.NewInputBlock("notes_input", plainTextBlock("Notes"), inputHint, inputBlock))

	return blockSet
}

func (s *ChatService) makeAnnotationModalView(ctx context.Context, ic *slack.InteractionCallback) (*slack.ModalViewRequest, error) {
	c, ctxErr := s.makeAnnotationViewContext(ctx, ic)
	if ctxErr != nil {
		return nil, fmt.Errorf("failed to get message annotation context: %w", ctxErr)
	}

	blockSet := makeAnnotationModalViewBlocks(c)

	titleText := "Create Annotation"
	submitText := "Create"

	if c.currAnnotation != nil {
		c.meta.AnnotationId = c.currAnnotation.ID
		titleText = "Update Annotation"
		submitText = "Update"
	}

	jsonMetadata, jsonErr := json.Marshal(c.meta)
	if jsonErr != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", jsonErr)
	}

	return &slack.ModalViewRequest{
		Type:            "modal",
		CallbackID:      annotationModalViewCallbackID,
		PrivateMetadata: string(jsonMetadata),
		Title:           plainTextBlock(titleText),
		Close:           plainTextBlock("Cancel"),
		Submit:          plainTextBlock(submitText),
		Blocks:          slack.Blocks{BlockSet: blockSet},
	}, nil
}

func (s *ChatService) getAnnotationModalAnnotation(ctx context.Context, view slack.View) (*ent.EventAnnotation, error) {
	var meta annotationViewMetadata
	if jsonErr := json.Unmarshal([]byte(view.PrivateMetadata), &meta); jsonErr != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", jsonErr)
	}

	usr, _, userErr := s.lookupChatUser(ctx, meta.UserId)
	if userErr != nil {
		return nil, fmt.Errorf("failed to lookup user: %w", userErr)
	}

	var notes string
	if view.State != nil {
		if notesInput, inputOk := view.State.Values["notes_input"]; inputOk {
			if noteBlock, noteOk := notesInput["notes_input_text"]; noteOk {
				notes = noteBlock.Value
			}
		}
	}

	ev := &ent.Event{
		ExternalID:  meta.MsgId.String(),
		Kind:        "message",
		Timestamp:   meta.MsgId.getTimestamp(),
		Source:      "slack",
		Title:       "Slack Message",
		Description: meta.MsgText,
	}

	anno := &ent.EventAnnotation{
		ID:        meta.AnnotationId,
		CreatorID: usr.ID,
		Notes:     notes,
		Edges: ent.EventAnnotationEdges{
			Event: ev,
		},
	}

	return anno, nil
}
