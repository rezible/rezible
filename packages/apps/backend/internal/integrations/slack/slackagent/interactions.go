package slackagent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/slack-go/slack"

	"github.com/rezible/rezible/internal/integrations/slack"
)

const (
	createAnnotationActionCallbackID = "create_annotation"
)

func (i *Integration) handleMessageActionInteraction(ctx context.Context, ic *slack.InteractionCallback) error {
	switch ic.CallbackID {
	case createAnnotationActionCallbackID:
		return i.handleAnnotationModalInteraction(ctx, ic)
	}
	return fmt.Errorf("unknown message actions: %s", ic.CallbackID)
}

func (i *Integration) handleBlockActionsInteraction(ctx context.Context, ic *slack.InteractionCallback) error {
	for _, action := range ic.ActionCallback.BlockActions {
		switch action.ActionID {
		//case actionCallbackIdIncidentDetailsModalButton:
		//	return s.handleIncidentDetailsModalInteraction(ctx, ic)
		default:
			return fmt.Errorf("unknown block action id: %s", action.ActionID)
		}
	}
	slog.Debug("interaction callback", "ic", ic)

	return fmt.Errorf("unknown block actions: %s", ic.CallbackID)
}

func (i *Integration) handleViewSubmissionInteraction(ctx context.Context, ic *slack.InteractionCallback) error {
	switch ic.View.CallbackID {
	case viewCallbackIdAnnotationModal:
		return i.handleAnnotationModalSubmission(ctx, ic)
	}
	return fmt.Errorf("unknown view submission: %s", ic.View.CallbackID)
}

func (i *Integration) handleAnnotationModalInteraction(ctx context.Context, ic *slack.InteractionCallback) error {
	meta := &annotationModalMetadata{
		UserId:  ic.User.ID,
		MsgId:   slackintegration.MessageId(fmt.Sprintf("%s_%s", ic.Channel.ID, ic.Message.Timestamp)),
		MsgText: ic.Message.Text,
	}
	if ic.View.PrivateMetadata != "" {
		if jsonErr := json.Unmarshal([]byte(ic.View.PrivateMetadata), &meta); jsonErr != nil {
			return fmt.Errorf("failed to unmarshal annotation metadata: %w", jsonErr)
		}
	}
	view, viewErr := i.makeAnnotationModalView(ctx, meta)
	if viewErr != nil || view == nil {
		return fmt.Errorf("failed to create annotation view: %w", viewErr)
	}
	return i.service.OpenOrUpdateModal(ctx, ic, view)
}

func (i *Integration) handleAnnotationModalSubmission(ctx context.Context, ic *slack.InteractionCallback) error {
	anno, annoErr := i.getAnnotationModalAnnotation(ctx, ic.View)
	if annoErr != nil {
		return fmt.Errorf("failed to get view annotation: %w", annoErr)
	}
	_, createErr := i.eventAnnos.SetAnnotation(ctx, anno)
	if createErr != nil {
		return fmt.Errorf("failed to create annotation: %w", createErr)
	}
	return nil
}
