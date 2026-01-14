package slack

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/slack-go/slack"

	"github.com/rezible/rezible/ent"
)

const (
	annotationModalViewCallbackID = "create_annotation_confirm"
)

func (s *ChatService) makeAnnotationModalView(ctx context.Context, ic *slack.InteractionCallback) (*slack.ModalViewRequest, error) {
	var meta annotationModalMetadata
	if ic.View.PrivateMetadata != "" {
		if jsonErr := json.Unmarshal([]byte(ic.View.PrivateMetadata), &meta); jsonErr != nil {
			return nil, jsonErr
		}
	} else {
		meta = annotationModalMetadata{
			UserId:  ic.User.ID,
			MsgId:   getMessageId(ic),
			MsgText: ic.Message.Text,
		}
	}

	usr, usrCtx, userErr := s.lookupChatUser(ctx, meta.UserId)
	if userErr != nil {
		return nil, fmt.Errorf("failed to lookup user: %w", userErr)
	}

	ev := &ent.Event{ExternalID: meta.MsgId.String()}

	curr, currErr := s.annos.LookupByUserEvent(usrCtx, usr.ID, ev)
	if currErr != nil && !ent.IsNotFound(currErr) {
		return nil, fmt.Errorf("failed to lookup existing event annotation: %w", currErr)
	}
	if curr != nil {
		meta.AnnotationId = curr.ID
	}

	builder := newAnnotationModalBuilder(curr, &meta)
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
		CallbackID:      annotationModalViewCallbackID,
		Title:           plainTextBlock(titleText),
		Submit:          plainTextBlock(submitText),
		Close:           plainTextBlock("Cancel"),
		Blocks:          blockSet,
		PrivateMetadata: string(jsonMetadata),
	}, nil
}

func (s *ChatService) getAnnotationModalAnnotation(ctx context.Context, view slack.View) (*ent.EventAnnotation, error) {
	var meta annotationModalMetadata
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
