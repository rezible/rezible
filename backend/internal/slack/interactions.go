package slack

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
)

const (
	createAnnotationActionCallbackID = "create_annotation"
)

func (s *ChatService) onInteractionEventReceived(ctx context.Context, ic *slack.InteractionCallback) (bool, error) {
	handled, _, err := s.handleInteractionEvent(ctx, ic)
	return handled, err
}

func (s *ChatService) handleInteractionEvent(ctx context.Context, ic *slack.InteractionCallback) (bool, any, error) {
	handled := true
	var payload any
	var handlerErr error
	switch ic.Type {
	case slack.InteractionTypeMessageAction:
		payload, handlerErr = s.handleMessageActionInteraction(ctx, ic)
	case slack.InteractionTypeViewSubmission:
		payload, handlerErr = s.handleViewSubmissionInteraction(ctx, ic)
	case slack.InteractionTypeBlockActions:
		payload, handlerErr = s.handleBlockActionsInteraction(ctx, ic)
	default:
		handled = false
	}
	return handled, payload, handlerErr
}

func (s *ChatService) handleMessageActionInteraction(ctx context.Context, ic *slack.InteractionCallback) (any, error) {
	if ic.CallbackID == createAnnotationActionCallbackID {
		return s.handleAnnotationModalInteraction(ctx, ic)
	}
	return nil, fmt.Errorf("unknown message actions: %s", ic.CallbackID)
}

func (s *ChatService) handleViewSubmissionInteraction(ctx context.Context, ic *slack.InteractionCallback) (any, error) {
	if ic.View.CallbackID == createAnnotationModalViewCallbackID {
		return s.handleAnnotationModalSubmission(ctx, ic)
	}
	return nil, fmt.Errorf("unknown view submission: %s", ic.View.CallbackID)
}

func (s *ChatService) handleBlockActionsInteraction(ctx context.Context, ic *slack.InteractionCallback) (any, error) {
	if ic.View.CallbackID == createAnnotationModalViewCallbackID {
		return s.handleAnnotationModalInteraction(ctx, ic)
	}
	return nil, fmt.Errorf("unknown block actions: %s", ic.View.CallbackID)
}

func (s *ChatService) handleAnnotationModalInteraction(ctx context.Context, ic *slack.InteractionCallback) (any, error) {
	view, viewErr := s.makeAnnotationModalView(ctx, ic)
	if viewErr != nil || view == nil {
		return nil, fmt.Errorf("failed to create annotation view: %w", viewErr)
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
		return nil, fmt.Errorf("annotation modal view: %w", respErr)
	}
	return nil, nil
}

func (s *ChatService) handleAnnotationModalSubmission(ctx context.Context, ic *slack.InteractionCallback) (any, error) {
	anno, annoErr := s.getAnnotationModalAnnotation(ctx, ic.View)
	if annoErr != nil {
		return nil, fmt.Errorf("failed to get view annotation: %w", annoErr)
	}
	_, createErr := s.annos.SetAnnotation(ctx, anno)
	return nil, createErr
}
