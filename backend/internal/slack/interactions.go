package slack

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
)

func (s *ChatService) handleInteractionEvent(ctx context.Context, ic *slack.InteractionCallback) error {
	switch ic.Type {
	case slack.InteractionTypeMessageAction:
		return s.handleMessageActionInteraction(ctx, ic)
	case slack.InteractionTypeViewSubmission:
		return s.handleViewSubmissionInteraction(ctx, ic)
	case slack.InteractionTypeBlockActions:
		return s.handleBlockActionsInteraction(ctx, ic)
	default:
		return fmt.Errorf("unknown interaction type: %s", string(ic.Type))
	}
}

func (s *ChatService) handleMessageActionInteraction(ctx context.Context, ic *slack.InteractionCallback) error {
	if ic.CallbackID == createAnnotationActionCallbackID {
		return s.handleAnnotationModalInteraction(ctx, ic)
	}
	return fmt.Errorf("unknown message actions: %s", ic.CallbackID)
}

func (s *ChatService) handleViewSubmissionInteraction(ctx context.Context, ic *slack.InteractionCallback) error {
	if ic.View.CallbackID == createAnnotationModalViewCallbackID {
		return s.handleAnnotationModalSubmission(ctx, ic)
	}
	return fmt.Errorf("unknown view submission: %s", ic.View.CallbackID)
}

func (s *ChatService) handleBlockActionsInteraction(ctx context.Context, ic *slack.InteractionCallback) error {
	if ic.View.CallbackID == createAnnotationModalViewCallbackID {
		return s.handleAnnotationModalInteraction(ctx, ic)
	}
	return fmt.Errorf("unknown block actions: %s", ic.View.CallbackID)
}

func (s *ChatService) handleAnnotationModalInteraction(ctx context.Context, ic *slack.InteractionCallback) error {
	view, viewErr := s.makeAnnotationModalView(ctx, ic)
	if viewErr != nil || view == nil {
		return fmt.Errorf("failed to create annotation view: %w", viewErr)
	}
	var viewResp *slack.ViewResponse
	var respErr error
	if ic.View.State == nil {
		viewResp, respErr = s.client.OpenViewContext(ctx, ic.TriggerID, *view)
	} else {
		viewResp, respErr = s.client.UpdateViewContext(ctx, *view, "", ic.Hash, ic.View.ID)
	}
	if respErr != nil {
		logSlackViewErrorResponse(respErr, viewResp)
		return fmt.Errorf("annotation modal view: %w", respErr)
	}
	return nil
}

func (s *ChatService) handleAnnotationModalSubmission(ctx context.Context, ic *slack.InteractionCallback) error {
	anno, annoErr := s.getAnnotationModalAnnotation(ctx, ic.View)
	if annoErr != nil {
		return fmt.Errorf("failed to get view annotation: %w", annoErr)
	}
	_, createErr := s.annos.SetAnnotation(ctx, anno)
	return createErr
}
