package slackincidents

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/rezible/rezible/execution"
	slackintegration "github.com/rezible/rezible/internal/integrations/slack"
	"github.com/slack-go/slack"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/ent/incidentmilestone"
)

func (a *app) handleMessageActionInteraction(ctx context.Context, ii *ent.Integration, ic *slack.InteractionCallback) error {
	return fmt.Errorf("unknown message actions: %s", ic.CallbackID)
}

func (a *app) handleBlockActionsInteraction(ctx context.Context, ii *ent.Integration, ic *slack.InteractionCallback) error {
	cw, cwErr := slackintegration.NewClientWrapper(ii)
	if cwErr != nil {
		return fmt.Errorf("failed to create client wrapper: %w", cwErr)
	}

	for _, action := range ic.ActionCallback.BlockActions {
		switch action.ActionID {
		case actionCallbackIdIncidentDetailsModalButton:
			return a.handleIncidentDetailsModalInteraction(ctx, cw, ic)
		case actionCallbackIdIncidentMilestoneModalButton:
			return a.handleIncidentMilestoneModalInteraction(ctx, cw, ic)
		default:
			return fmt.Errorf("unknown block action id: %s", action.ActionID)
		}
	}
	slog.Debug("interaction callback", "ic", ic)

	return fmt.Errorf("unknown block actions: %s", ic.CallbackID)
}

func (a *app) handleViewSubmissionInteraction(ctx context.Context, ii *ent.Integration, ic *slack.InteractionCallback) error {
	switch ic.View.CallbackID {
	case viewCallbackIdIncidentDetailsModal:
		return a.handleIncidentDetailsModalSubmission(ctx, ic)
	case viewCallbackIdIncidentMilestoneModal:
		return a.handleIncidentMilestoneModalSubmission(ctx, ic)
	}
	return fmt.Errorf("unknown view submission: %s", ic.View.CallbackID)
}

func (a *app) getIncidentModalViewMetadata(ctx context.Context, ic *slack.InteractionCallback) (*incidentDetailsModalViewMetadata, error) {
	var meta incidentDetailsModalViewMetadata
	if ic.Type == slack.InteractionTypeBlockActions {
		inc, incErr := a.incidents.Get(ctx, incident.ChatChannelID(ic.Channel.ID))
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

func (a *app) handleIncidentDetailsModalInteraction(ctx context.Context, cw *slackintegration.ClientWrapper, ic *slack.InteractionCallback) error {
	meta, metaErr := a.getIncidentModalViewMetadata(ctx, ic)
	if metaErr != nil {
		return fmt.Errorf("getting modal view metadata: %w", metaErr)
	}
	// TODO: load these from configured integration
	prefs := UserSettingsIncidents{}
	view, viewErr := a.makeIncidentDetailsModalView(ctx, prefs, meta)
	if viewErr != nil || view == nil {
		return fmt.Errorf("failed to create incident view: %w", viewErr)
	}
	return cw.OpenOrUpdateModal(ctx, ic, view)
}

func (a *app) handleIncidentMilestoneModalInteraction(ctx context.Context, cw *slackintegration.ClientWrapper, ic *slack.InteractionCallback) error {
	meta, metaErr := a.getIncidentMilestoneModalViewMetadata(ctx, ic)
	if metaErr != nil {
		return fmt.Errorf("getting modal view metadata: %w", metaErr)
	}
	view, viewErr := a.makeIncidentMilestoneModalView(ctx, meta)
	if viewErr != nil || view == nil {
		return fmt.Errorf("failed to create incident view: %w", viewErr)
	}
	return cw.OpenOrUpdateModal(ctx, ic, view)
}

func (a *app) handleIncidentDetailsModalSubmission(ctx context.Context, ic *slack.InteractionCallback) error {
	meta, metaErr := a.getIncidentModalViewMetadata(ctx, ic)
	if metaErr != nil {
		return fmt.Errorf("getting modal view metadata: %w", metaErr)
	}

	state := ic.View.State
	if state == nil {
		return fmt.Errorf("missing incident details modal view state")
	}

	userId, ok := execution.GetContext(ctx).UserID()
	if !ok {
		return fmt.Errorf("missing user context")
	}

	return a.db.WithTx(ctx, func(ctx context.Context, client *ent.Client) error {
		inc, incErr := a.incidents.Set(ctx, meta.IncidentId, func(m *ent.IncidentMutation) {
			setIncidentDetailsModalInputMutationFields(m, state)
		})
		if incErr != nil {
			return fmt.Errorf("upsert incident from modal data: %w", incErr)
		}
		if inc.ID == meta.IncidentId { // updated existing
			return nil
		}
		_, msErr := a.incidents.SetIncidentMilestone(ctx, uuid.Nil, func(m *ent.IncidentMilestoneMutation) {
			m.SetKind(incidentmilestone.KindOpened)
			m.SetDescription("Incident declared via slack")
			m.SetTimestamp(time.Now())
			m.SetSource(integrationName)
			m.SetIncidentID(inc.ID)
			m.SetUserID(userId)
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

func (a *app) getIncidentMilestoneModalViewMetadata(ctx context.Context, ic *slack.InteractionCallback) (*incidentMilestoneModalViewMetadata, error) {
	var meta incidentMilestoneModalViewMetadata
	if ic.View.PrivateMetadata != "" {
		if jsonErr := json.Unmarshal([]byte(ic.View.PrivateMetadata), &meta); jsonErr != nil {
			return nil, fmt.Errorf("failed to unmarshal incident modal metadata: %w", jsonErr)
		}
	} else {
		inc, incErr := a.incidents.Get(ctx, incident.ChatChannelID(ic.Channel.ID))
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

func (a *app) handleIncidentMilestoneModalSubmission(ctx context.Context, ic *slack.InteractionCallback) error {
	meta, metaErr := a.getIncidentModalViewMetadata(ctx, ic)
	if metaErr != nil {
		return fmt.Errorf("getting modal view metadata: %w", metaErr)
	}

	userId, ok := execution.GetContext(ctx).UserID()
	if !ok {
		return fmt.Errorf("missing user context")
	}

	state := ic.View.State
	if state == nil {
		return fmt.Errorf("invalid incident milestone modal view state")
	}

	setFn := func(m *ent.IncidentMilestoneMutation) {
		m.SetIncidentID(meta.IncidentId)
		m.SetSource(integrationName)
		m.SetTimestamp(time.Now())
		m.SetUserID(userId)
		setIncidentMilestoneModalInputMutationFields(m, state)
	}
	_, incErr := a.incidents.SetIncidentMilestone(ctx, uuid.Nil, setFn)
	if incErr != nil {
		return fmt.Errorf("set milestone from modal data: %w", incErr)
	}
	return nil
}
