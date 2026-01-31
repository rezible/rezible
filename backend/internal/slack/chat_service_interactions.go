package slack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incidentmilestone"
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
	_, usrCtx, usrErr := s.lookupUser(ctx, ic.User.ID)
	if usrErr != nil {
		return false, nil, fmt.Errorf("failed to lookup user: %w", usrErr)
	}
	ctx = usrCtx

	handled := true
	var payload any
	var handlerErr error
	switch ic.Type {
	case slack.InteractionTypeMessageAction:
		payload, handlerErr = s.handleMessageActionInteraction(ctx, ic)
	case slack.InteractionTypeBlockActions:
		payload, handlerErr = s.handleBlockActionsInteraction(ctx, ic)
	case slack.InteractionTypeViewSubmission:
		payload, handlerErr = s.handleViewSubmissionInteraction(ctx, ic)
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
	for _, action := range ic.ActionCallback.BlockActions {
		switch action.ActionID {
		case actionCallbackIdIncidentDetailsModalButton:
			return s.handleIncidentDetailsModalInteraction(ctx, ic)
		case actionCallbackIdIncidentMilestoneModalButton:
			return s.handleIncidentMilestoneModalInteraction(ctx, ic)
		default:
			return nil, fmt.Errorf("unknown block action id: %s", action.ActionID)
		}
	}
	//switch ic.View.CallbackID {
	//case viewCallbackIdAnnotationModal:
	//	return s.handleAnnotationModalInteraction(ctx, ic)
	//case viewCallbackIdIncidentDetailsModal:
	//	return s.handleIncidentDetailsModalInteraction(ctx, ic)
	//}
	log.Debug().Interface("ic", ic).Msg("interaction callback")

	return nil, fmt.Errorf("unknown block actions: %s", ic.CallbackID)
}

func (s *ChatService) handleViewSubmissionInteraction(ctx context.Context, ic *slack.InteractionCallback) (any, error) {
	switch ic.View.CallbackID {
	case viewCallbackIdAnnotationModal:
		return s.handleAnnotationModalSubmission(ctx, ic)
	case viewCallbackIdIncidentDetailsModal:
		return s.handleIncidentDetailsModalSubmission(ctx, ic)
	case viewCallbackIdIncidentMilestoneModal:
		return s.handleIncidentMilestoneModalSubmission(ctx, ic)
	}
	return nil, fmt.Errorf("unknown view submission: %s", ic.View.CallbackID)
}

func getInteractionAnnoationModalViewMetadata(ic *slack.InteractionCallback) (*annotationModalMetadata, error) {
	if ic.View.PrivateMetadata != "" {
		var meta annotationModalMetadata
		if jsonErr := json.Unmarshal([]byte(ic.View.PrivateMetadata), &meta); jsonErr != nil {
			return nil, jsonErr
		}
		return &meta, nil
	}
	return &annotationModalMetadata{
		UserId:  ic.User.ID,
		MsgId:   messageId(fmt.Sprintf("%s_%s", ic.Channel.ID, ic.Message.Timestamp)),
		MsgText: ic.Message.Text,
	}, nil
}

func (s *ChatService) handleAnnotationModalInteraction(ctx context.Context, ic *slack.InteractionCallback) (any, error) {
	meta, metaErr := getInteractionAnnoationModalViewMetadata(ic)
	if metaErr != nil {
		return nil, fmt.Errorf("failed to get interaction metadata: %w", metaErr)
	}
	view, viewErr := s.makeAnnotationModalView(ctx, meta)
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

func (s *ChatService) getIncidentModalViewMetadata(ctx context.Context, ic *slack.InteractionCallback) (*incidentDetailsModalViewMetadata, error) {
	var meta incidentDetailsModalViewMetadata
	if ic.Type == slack.InteractionTypeBlockActions {
		inc, incErr := s.incidents.GetByChatChannelID(ctx, ic.Channel.ID)
		if incErr != nil {
			return nil, fmt.Errorf("unable to get incident by channel: %w", incErr)
		}
		meta = incidentDetailsModalViewMetadata{
			UserId:     ic.User.ID,
			IncidentId: inc.ID,
		}
	} else {
		if ic.View.PrivateMetadata == "" {
			return nil, fmt.Errorf("no view metadata provided")
		}
		if jsonErr := json.Unmarshal([]byte(ic.View.PrivateMetadata), &meta); jsonErr != nil {
			return nil, fmt.Errorf("failed to unmarshal incident modal metadata: %w", jsonErr)
		}
	}
	return &meta, nil
}

func (s *ChatService) handleIncidentDetailsModalInteraction(ctx context.Context, ic *slack.InteractionCallback) (any, error) {
	meta, metaErr := s.getIncidentModalViewMetadata(ctx, ic)
	if metaErr != nil {
		return nil, metaErr
	}
	view, viewErr := s.makeIncidentDetailsModalView(ctx, meta)
	if viewErr != nil || view == nil {
		return nil, fmt.Errorf("failed to create incident view: %w", viewErr)
	}
	openErr := s.openOrUpdateModal(ctx, ic, view)
	if openErr != nil {
		return nil, fmt.Errorf("failed to open view: %w", openErr)
	}

	return nil, nil
}

func (s *ChatService) handleIncidentDetailsModalSubmission(ctx context.Context, ic *slack.InteractionCallback) (any, error) {
	meta, metaErr := s.getIncidentModalViewMetadata(ctx, ic)
	if metaErr != nil {
		return nil, metaErr
	}

	creating := meta.IncidentId == uuid.Nil
	state := ic.View.State
	if state == nil {
		return nil, errors.New("invalid view state")
	}

	usr, userErr := s.users.GetByChatId(ctx, meta.UserId)
	if userErr != nil {
		return nil, fmt.Errorf("failed to get user: %w", userErr)
	}

	setFn := func(m *ent.IncidentMutation) []ent.Mutation {
		setIncidentDetailsModalInputMutationFields(m, state)

		incidentId, exists := m.ID()
		if !exists || !creating {
			return nil
		}
		milestoneMeta := map[string]string{
			"channel_id": meta.CommandChannelId,
			"user_id":    meta.UserId,
		}
		milestoneCreate := m.Client().IncidentMilestone.Create().
			SetKind(incidentmilestone.KindOpened).
			SetDescription("Incident declared via slack").
			SetTimestamp(time.Now()).
			SetSource(integrationName).
			SetIncidentID(incidentId).
			SetMetadata(milestoneMeta).
			SetUserID(usr.ID).
			Mutation()

		return []ent.Mutation{milestoneCreate}
	}
	_, incErr := s.incidents.Set(ctx, meta.IncidentId, setFn)
	if incErr != nil {
		return nil, fmt.Errorf("upsert incident from modal data: %w", incErr)
	}
	return nil, nil
}

func (s *ChatService) getIncidentMilestoneModalViewMetadata(ctx context.Context, ic *slack.InteractionCallback) (*incidentMilestoneModalViewMetadata, error) {
	var meta incidentMilestoneModalViewMetadata
	if ic.View.PrivateMetadata != "" {
		if jsonErr := json.Unmarshal([]byte(ic.View.PrivateMetadata), &meta); jsonErr != nil {
			return nil, fmt.Errorf("failed to unmarshal incident modal metadata: %w", jsonErr)
		}
	} else {
		inc, incErr := s.incidents.GetByChatChannelID(ctx, ic.Channel.ID)
		if incErr != nil {
			return nil, fmt.Errorf("unable to get incident by channel: %w", incErr)
		}
		meta = incidentMilestoneModalViewMetadata{
			UserId:     ic.User.ID,
			IncidentId: inc.ID,
		}
	}
	return &meta, nil
}

func (s *ChatService) handleIncidentMilestoneModalInteraction(ctx context.Context, ic *slack.InteractionCallback) (any, error) {
	meta, metaErr := s.getIncidentMilestoneModalViewMetadata(ctx, ic)
	if metaErr != nil {
		return nil, fmt.Errorf("failed to get view metadata: %w", metaErr)
	}
	view, viewErr := s.makeIncidentMilestoneModalView(ctx, meta)
	if viewErr != nil || view == nil {
		return nil, fmt.Errorf("failed to create incident view: %w", viewErr)
	}
	openErr := s.openOrUpdateModal(ctx, ic, view)
	if openErr != nil {
		return nil, fmt.Errorf("failed to open view: %w", openErr)
	}

	return nil, nil
}

func (s *ChatService) handleIncidentMilestoneModalSubmission(ctx context.Context, ic *slack.InteractionCallback) (any, error) {
	meta, metaErr := s.getIncidentModalViewMetadata(ctx, ic)
	if metaErr != nil {
		return nil, metaErr
	}

	usr, userErr := s.users.GetByChatId(ctx, meta.UserId)
	if userErr != nil {
		return nil, fmt.Errorf("failed to get user: %w", userErr)
	}

	state := ic.View.State
	if state == nil {
		return nil, errors.New("invalid view state")
	}

	setFn := func(m *ent.IncidentMilestoneMutation) {
		m.SetIncidentID(meta.IncidentId)
		m.SetSource(integrationName)
		m.SetTimestamp(time.Now())
		m.SetUserID(usr.ID)
		setIncidentMilestoneModalInputMutationFields(m, state)
	}
	_, incErr := s.incidents.SetIncidentMilestone(ctx, uuid.Nil, setFn)
	if incErr != nil {
		return nil, fmt.Errorf("set milestone from modal data: %w", incErr)
	}
	return nil, nil
}
