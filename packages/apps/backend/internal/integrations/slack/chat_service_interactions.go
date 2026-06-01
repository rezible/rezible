package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/slack-go/slack"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/incidentmilestone"
	"github.com/rezible/rezible/ent/user"
)

const (
	createAnnotationActionCallbackID = "create_annotation"
)

func (s *ChatService) handleInteractionCallback(baseCtx context.Context, ic *slack.InteractionCallback) error {
	ctx, usrErr := s.createUserContext(baseCtx, ic.User.ID)
	if usrErr != nil {
		return fmt.Errorf("failed to lookup user: %w", usrErr)
	}

	switch ic.Type {
	case slack.InteractionTypeMessageAction:
		return s.handleMessageActionInteraction(ctx, ic)
	case slack.InteractionTypeBlockActions:
		return s.handleBlockActionsInteraction(ctx, ic)
	case slack.InteractionTypeViewSubmission:
		return s.handleViewSubmissionInteraction(ctx, ic)
	default:
		s.logger.Warn("unknown interaction type")
	}
	return nil
}

func (s *ChatService) handleMessageActionInteraction(ctx context.Context, ic *slack.InteractionCallback) error {
	switch ic.CallbackID {
	case createAnnotationActionCallbackID:
		return s.handleAnnotationModalInteraction(ctx, ic)
	}
	return fmt.Errorf("unknown message actions: %s", ic.CallbackID)
}

func (s *ChatService) handleBlockActionsInteraction(ctx context.Context, ic *slack.InteractionCallback) error {
	for _, action := range ic.ActionCallback.BlockActions {
		switch action.ActionID {
		case actionCallbackIdIncidentDetailsModalButton:
			return s.handleIncidentDetailsModalInteraction(ctx, ic)
		case actionCallbackIdIncidentMilestoneModalButton:
			return s.handleIncidentMilestoneModalInteraction(ctx, ic)
		default:
			return fmt.Errorf("unknown block action id: %s", action.ActionID)
		}
	}
	s.logger.Debug("interaction callback", "ic", ic)

	return fmt.Errorf("unknown block actions: %s", ic.CallbackID)
}

func (s *ChatService) handleViewSubmissionInteraction(ctx context.Context, ic *slack.InteractionCallback) error {
	switch ic.View.CallbackID {
	case viewCallbackIdAnnotationModal:
		return s.handleAnnotationModalSubmission(ctx, ic)
	case viewCallbackIdIncidentDetailsModal:
		return s.handleIncidentDetailsModalSubmission(ctx, ic)
	case viewCallbackIdIncidentMilestoneModal:
		return s.handleIncidentMilestoneModalSubmission(ctx, ic)
	}
	return fmt.Errorf("unknown view submission: %s", ic.View.CallbackID)
}

func (s *ChatService) handleAnnotationModalInteraction(ctx context.Context, ic *slack.InteractionCallback) error {
	meta := &annotationModalMetadata{
		UserId:  ic.User.ID,
		MsgId:   messageId(fmt.Sprintf("%s_%s", ic.Channel.ID, ic.Message.Timestamp)),
		MsgText: ic.Message.Text,
	}
	if ic.View.PrivateMetadata != "" {
		if jsonErr := json.Unmarshal([]byte(ic.View.PrivateMetadata), &meta); jsonErr != nil {
			return fmt.Errorf("failed to unmarshal annotation metadata: %w", jsonErr)
		}
	}
	view, viewErr := s.makeAnnotationModalView(ctx, meta)
	if viewErr != nil || view == nil {
		return fmt.Errorf("failed to create annotation view: %w", viewErr)
	}
	return s.openOrUpdateModal(ctx, ic, view)
}

func (s *ChatService) handleAnnotationModalSubmission(ctx context.Context, ic *slack.InteractionCallback) error {
	anno, annoErr := s.getAnnotationModalAnnotation(ctx, ic.View)
	if annoErr != nil {
		return fmt.Errorf("failed to get view annotation: %w", annoErr)
	}
	_, createErr := s.ci.eventAnnos.SetAnnotation(ctx, anno)
	if createErr != nil {
		return fmt.Errorf("failed to create annotation: %w", createErr)
	}
	return nil
}

func (s *ChatService) getIncidentModalViewMetadata(ctx context.Context, ic *slack.InteractionCallback) (*incidentDetailsModalViewMetadata, error) {
	var meta incidentDetailsModalViewMetadata
	if ic.Type == slack.InteractionTypeBlockActions {
		inc, incErr := s.ci.incidents.Get(ctx, incident.ChatChannelID(ic.Channel.ID))
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

func (s *ChatService) handleIncidentDetailsModalInteraction(ctx context.Context, ic *slack.InteractionCallback) error {
	meta, metaErr := s.getIncidentModalViewMetadata(ctx, ic)
	if metaErr != nil {
		return fmt.Errorf("getting modal view metadata: %w", metaErr)
	}
	view, viewErr := s.makeIncidentDetailsModalView(ctx, meta)
	if viewErr != nil || view == nil {
		return fmt.Errorf("failed to create incident view: %w", viewErr)
	}
	return s.openOrUpdateModal(ctx, ic, view)
}

func (s *ChatService) handleIncidentMilestoneModalInteraction(ctx context.Context, ic *slack.InteractionCallback) error {
	meta, metaErr := s.getIncidentMilestoneModalViewMetadata(ctx, ic)
	if metaErr != nil {
		return fmt.Errorf("getting modal view metadata: %w", metaErr)
	}
	view, viewErr := s.makeIncidentMilestoneModalView(ctx, meta)
	if viewErr != nil || view == nil {
		return fmt.Errorf("failed to create incident view: %w", viewErr)
	}
	return s.openOrUpdateModal(ctx, ic, view)
}

func (s *ChatService) handleIncidentDetailsModalSubmission(ctx context.Context, ic *slack.InteractionCallback) error {
	meta, metaErr := s.getIncidentModalViewMetadata(ctx, ic)
	if metaErr != nil {
		return fmt.Errorf("getting modal view metadata: %w", metaErr)
	}

	state := ic.View.State
	if state == nil {
		return fmt.Errorf("missing incident details modal view state")
	}

	usr, userErr := s.ci.users.Get(ctx, user.ChatID(meta.UserId))
	if userErr != nil {
		return fmt.Errorf("failed to get user: %w", userErr)
	}

	return s.ci.db.WithTx(ctx, func(ctx context.Context, client *ent.Client) error {
		inc, incErr := s.ci.incidents.Set(ctx, meta.IncidentId, func(m *ent.IncidentMutation) {
			setIncidentDetailsModalInputMutationFields(m, state)
		})
		if incErr != nil {
			return fmt.Errorf("upsert incident from modal data: %w", incErr)
		}
		if inc.ID == meta.IncidentId { // updated existing
			return nil
		}
		_, msErr := s.ci.incidents.SetIncidentMilestone(ctx, uuid.Nil, func(m *ent.IncidentMilestoneMutation) {
			m.SetKind(incidentmilestone.KindOpened)
			m.SetDescription("Incident declared via slack")
			m.SetTimestamp(time.Now())
			m.SetSource(integrationName)
			m.SetIncidentID(inc.ID)
			m.SetUserID(usr.ID)
			m.SetMetadata(map[string]string{
				"channel_id": meta.CommandChannelId,
				"user_id":    meta.UserId,
			})
		})
		if msErr != nil {
			return fmt.Errorf("incident milestone create: %w", msErr)
		}
		return nil
	})
}

func (s *ChatService) getIncidentMilestoneModalViewMetadata(ctx context.Context, ic *slack.InteractionCallback) (*incidentMilestoneModalViewMetadata, error) {
	var meta incidentMilestoneModalViewMetadata
	if ic.View.PrivateMetadata != "" {
		if jsonErr := json.Unmarshal([]byte(ic.View.PrivateMetadata), &meta); jsonErr != nil {
			return nil, fmt.Errorf("failed to unmarshal incident modal metadata: %w", jsonErr)
		}
	} else {
		inc, incErr := s.ci.incidents.Get(ctx, incident.ChatChannelID(ic.Channel.ID))
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

func (s *ChatService) handleIncidentMilestoneModalSubmission(ctx context.Context, ic *slack.InteractionCallback) error {
	meta, metaErr := s.getIncidentModalViewMetadata(ctx, ic)
	if metaErr != nil {
		return fmt.Errorf("getting modal view metadata: %w", metaErr)
	}

	usr, userErr := s.ci.users.Get(ctx, user.ChatID(meta.UserId))
	if userErr != nil {
		return fmt.Errorf("failed to get user: %w", userErr)
	}

	state := ic.View.State
	if state == nil {
		return fmt.Errorf("invalid incident milestone modal view state")
	}

	setFn := func(m *ent.IncidentMilestoneMutation) {
		m.SetIncidentID(meta.IncidentId)
		m.SetSource(integrationName)
		m.SetTimestamp(time.Now())
		m.SetUserID(usr.ID)
		setIncidentMilestoneModalInputMutationFields(m, state)
	}
	_, incErr := s.ci.incidents.SetIncidentMilestone(ctx, uuid.Nil, setFn)
	if incErr != nil {
		return fmt.Errorf("set milestone from modal data: %w", incErr)
	}
	return nil
}
