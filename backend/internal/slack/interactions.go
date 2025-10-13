package slack

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
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
	var usrErr error
	ctx, usrErr = s.getChatUserContext(ctx, ic.User.ID)
	if usrErr != nil {
		return false, nil, fmt.Errorf("failed to lookup user: %w", usrErr)
	}

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
	switch ic.CallbackID {
	case createAnnotationActionCallbackID:
		return s.handleAnnotationModalInteraction(ctx, ic)
	}
	return nil, fmt.Errorf("unknown message actions: %s", ic.CallbackID)
}

func (s *ChatService) handleBlockActionsInteraction(ctx context.Context, ic *slack.InteractionCallback) (any, error) {
	switch ic.View.CallbackID {
	case annotationModalViewCallbackID:
		return s.handleAnnotationModalInteraction(ctx, ic)
	case createIncidentModalViewCallbackID:
		return s.handleIncidentModalInteraction(ctx, ic)
	}
	return nil, fmt.Errorf("unknown block actions: %s", ic.View.CallbackID)
}

func (s *ChatService) handleViewSubmissionInteraction(ctx context.Context, ic *slack.InteractionCallback) (any, error) {
	switch ic.View.CallbackID {
	case annotationModalViewCallbackID:
		return s.handleAnnotationModalSubmission(ctx, ic)
	case createIncidentModalViewCallbackID:
		return s.handleIncidentModalSubmission(ctx, ic)
	}
	return nil, fmt.Errorf("unknown view submission: %s", ic.View.CallbackID)
}

func (s *ChatService) handleAnnotationModalInteraction(ctx context.Context, ic *slack.InteractionCallback) (any, error) {
	view, viewErr := s.makeAnnotationModalView(ctx, ic)
	if viewErr != nil || view == nil {
		return nil, fmt.Errorf("failed to create annotation view: %w", viewErr)
	}
	return nil, s.openOrUpdateModal(ctx, ic, view)
}

func (s *ChatService) handleAnnotationModalSubmission(ctx context.Context, ic *slack.InteractionCallback) (any, error) {
	anno, annoErr := s.getAnnotationModalAnnotation(ctx, ic.View)
	if annoErr != nil {
		return nil, fmt.Errorf("failed to get view annotation: %w", annoErr)
	}
	_, createErr := s.annos.SetAnnotation(ctx, anno)
	if createErr != nil {
		return nil, fmt.Errorf("failed to create annotation: %w", createErr)
	}
	return nil, nil
}

func (s *ChatService) handleIncidentModalInteraction(ctx context.Context, ic *slack.InteractionCallback) (any, error) {
	meta, mdErr := s.fetchIncidentViewMetadata(ctx, ic)
	if mdErr != nil {
		return nil, fmt.Errorf("failed to fetch incident modal metadata: %w", mdErr)
	}
	view, viewErr := s.makeIncidentModalView(ctx, meta)
	if viewErr != nil || view == nil {
		return nil, fmt.Errorf("failed to create incident view: %w", viewErr)
	}
	openErr := s.openOrUpdateModal(ctx, ic, view)
	if openErr != nil {
		return nil, fmt.Errorf("failed to open view: %w", openErr)
	}

	return nil, nil
}

func (s *ChatService) handleIncidentModalSubmission(ctx context.Context, ic *slack.InteractionCallback) (any, error) {
	md, mdErr := getIncidentModalMetadata(ic.View)
	if mdErr != nil {
		return nil, fmt.Errorf("failed to get incident view metadata: %w", mdErr)
	}

	// TODO: create/update incident

	if msgErr := s.sendIncidentCreatedMessage(ctx, md); msgErr != nil {
		log.Error().Err(msgErr).Msg("failed to send incident created message")
		return nil, msgErr
	}

	return nil, nil
}
