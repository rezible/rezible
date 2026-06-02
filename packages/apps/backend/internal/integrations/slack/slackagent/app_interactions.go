package slackagent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/rezible/rezible/ent"
	"github.com/slack-go/slack"

	"github.com/rezible/rezible/internal/integrations/slack"
)

const (
	createAnnotationActionCallbackID = "create_annotation"
)

func (a *app) handleMessageActionInteraction(ctx context.Context, ii *ent.Integration, ic *slack.InteractionCallback) error {
	cw, cwErr := slackintegration.NewClientWrapper(ii)
	if cwErr != nil {
		return fmt.Errorf("failed to create client wrapper: %w", cwErr)
	}

	switch ic.CallbackID {
	case createAnnotationActionCallbackID:
		return a.handleAnnotationModalInteraction(ctx, cw, ic)
	}
	return fmt.Errorf("unknown message actions: %s", ic.CallbackID)
}

func (a *app) handleBlockActionsInteraction(ctx context.Context, ii *ent.Integration, ic *slack.InteractionCallback) error {
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

func (a *app) handleViewSubmissionInteraction(ctx context.Context, ii *ent.Integration, ic *slack.InteractionCallback) error {
	switch ic.View.CallbackID {
	case viewCallbackIdAnnotationModal:
		return a.handleAnnotationModalSubmission(ctx, ic)
	}
	return fmt.Errorf("unknown view submission: %s", ic.View.CallbackID)
}

func (a *app) handleAnnotationModalInteraction(ctx context.Context, cw *slackintegration.ClientWrapper, ic *slack.InteractionCallback) error {
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
	view, viewErr := a.makeAnnotationModalView(ctx, meta)
	if viewErr != nil || view == nil {
		return fmt.Errorf("failed to create annotation view: %w", viewErr)
	}
	return cw.OpenOrUpdateModal(ctx, ic, view)
}

func (a *app) handleAnnotationModalSubmission(ctx context.Context, ic *slack.InteractionCallback) error {
	anno, annoErr := a.getAnnotationModalAnnotation(ctx, ic.View)
	if annoErr != nil {
		return fmt.Errorf("failed to get view annotation: %w", annoErr)
	}
	_, createErr := a.eventAnnos.SetAnnotation(ctx, anno)
	if createErr != nil {
		return fmt.Errorf("failed to create annotation: %w", createErr)
	}
	return nil
}
