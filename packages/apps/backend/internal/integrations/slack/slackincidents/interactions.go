package slackincidents

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/slack-go/slack"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/incidentmilestone"
	"github.com/rezible/rezible/ent/user"
)

func (i *Integration) handleMessageActionInteraction(ctx context.Context, ic *slack.InteractionCallback) error {

	return fmt.Errorf("unknown message actions: %s", ic.CallbackID)
}

func (i *Integration) handleBlockActionsInteraction(ctx context.Context, ic *slack.InteractionCallback) error {
	for _, action := range ic.ActionCallback.BlockActions {
		switch action.ActionID {
		case actionCallbackIdIncidentDetailsModalButton:
			return i.handleIncidentDetailsModalInteraction(ctx, ic)
		case actionCallbackIdIncidentMilestoneModalButton:
			return i.handleIncidentMilestoneModalInteraction(ctx, ic)
		default:
			return fmt.Errorf("unknown block action id: %s", action.ActionID)
		}
	}
	slog.Debug("interaction callback", "ic", ic)

	return fmt.Errorf("unknown block actions: %s", ic.CallbackID)
}

func (i *Integration) handleViewSubmissionInteraction(ctx context.Context, ic *slack.InteractionCallback) error {
	switch ic.View.CallbackID {
	case viewCallbackIdIncidentDetailsModal:
		return i.handleIncidentDetailsModalSubmission(ctx, ic)
	case viewCallbackIdIncidentMilestoneModal:
		return i.handleIncidentMilestoneModalSubmission(ctx, ic)
	}
	return fmt.Errorf("unknown view submission: %s", ic.View.CallbackID)
}

func (i *Integration) getIncidentModalViewMetadata(ctx context.Context, ic *slack.InteractionCallback) (*incidentDetailsModalViewMetadata, error) {
	var meta incidentDetailsModalViewMetadata
	if ic.Type == slack.InteractionTypeBlockActions {
		inc, incErr := i.incidents.Get(ctx, incident.ChatChannelID(ic.Channel.ID))
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

func (i *Integration) handleIncidentDetailsModalInteraction(ctx context.Context, ic *slack.InteractionCallback) error {
	meta, metaErr := i.getIncidentModalViewMetadata(ctx, ic)
	if metaErr != nil {
		return fmt.Errorf("getting modal view metadata: %w", metaErr)
	}
	// TODO: load these from configured integration
	prefs := incidentPreferences{}
	view, viewErr := i.makeIncidentDetailsModalView(ctx, prefs, meta)
	if viewErr != nil || view == nil {
		return fmt.Errorf("failed to create incident view: %w", viewErr)
	}
	return i.service.OpenOrUpdateModal(ctx, ic, view)
}

func (i *Integration) handleIncidentMilestoneModalInteraction(ctx context.Context, ic *slack.InteractionCallback) error {
	meta, metaErr := i.getIncidentMilestoneModalViewMetadata(ctx, ic)
	if metaErr != nil {
		return fmt.Errorf("getting modal view metadata: %w", metaErr)
	}
	view, viewErr := i.makeIncidentMilestoneModalView(ctx, meta)
	if viewErr != nil || view == nil {
		return fmt.Errorf("failed to create incident view: %w", viewErr)
	}
	return i.service.OpenOrUpdateModal(ctx, ic, view)
}

func (i *Integration) handleIncidentDetailsModalSubmission(ctx context.Context, ic *slack.InteractionCallback) error {
	meta, metaErr := i.getIncidentModalViewMetadata(ctx, ic)
	if metaErr != nil {
		return fmt.Errorf("getting modal view metadata: %w", metaErr)
	}

	state := ic.View.State
	if state == nil {
		return fmt.Errorf("missing incident details modal view state")
	}

	usr, userErr := i.users.Get(ctx, user.ChatID(meta.UserId))
	if userErr != nil {
		return fmt.Errorf("failed to get user: %w", userErr)
	}

	return i.db.WithTx(ctx, func(ctx context.Context, client *ent.Client) error {
		inc, incErr := i.incidents.Set(ctx, meta.IncidentId, func(m *ent.IncidentMutation) {
			setIncidentDetailsModalInputMutationFields(m, state)
		})
		if incErr != nil {
			return fmt.Errorf("upsert incident from modal data: %w", incErr)
		}
		if inc.ID == meta.IncidentId { // updated existing
			return nil
		}
		_, msErr := i.incidents.SetIncidentMilestone(ctx, uuid.Nil, func(m *ent.IncidentMilestoneMutation) {
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

func (i *Integration) getIncidentMilestoneModalViewMetadata(ctx context.Context, ic *slack.InteractionCallback) (*incidentMilestoneModalViewMetadata, error) {
	var meta incidentMilestoneModalViewMetadata
	if ic.View.PrivateMetadata != "" {
		if jsonErr := json.Unmarshal([]byte(ic.View.PrivateMetadata), &meta); jsonErr != nil {
			return nil, fmt.Errorf("failed to unmarshal incident modal metadata: %w", jsonErr)
		}
	} else {
		inc, incErr := i.incidents.Get(ctx, incident.ChatChannelID(ic.Channel.ID))
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

func (i *Integration) handleIncidentMilestoneModalSubmission(ctx context.Context, ic *slack.InteractionCallback) error {
	meta, metaErr := i.getIncidentModalViewMetadata(ctx, ic)
	if metaErr != nil {
		return fmt.Errorf("getting modal view metadata: %w", metaErr)
	}

	usr, userErr := i.users.Get(ctx, user.ChatID(meta.UserId))
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
	_, incErr := i.incidents.SetIncidentMilestone(ctx, uuid.Nil, setFn)
	if incErr != nil {
		return fmt.Errorf("set milestone from modal data: %w", incErr)
	}
	return nil
}
