package slackagent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/slack-go/slack"

	"github.com/rezible/rezible/ent"
	sit "github.com/rezible/rezible/ent/situation"
	sitnr "github.com/rezible/rezible/ent/situationnoticerating"
	"github.com/rezible/rezible/internal/integrations/slack"
	"github.com/rezible/rezible/pkg/execution"
)

func (a *App) SlashCommandHandlers() map[string]slackintegration.SlashCommandHandler {
	return map[string]slackintegration.SlashCommandHandler{}
}

func (a *App) InteractionCallbackHandlers() map[slack.InteractionType]slackintegration.InteractionCallbackHandler {
	return map[slack.InteractionType]slackintegration.InteractionCallbackHandler{
		slack.InteractionTypeMessageAction:  a.handleMessageActionInteraction,
		slack.InteractionTypeBlockActions:   a.handleBlockActionsInteraction,
		slack.InteractionTypeViewSubmission: a.handleViewSubmissionInteraction,
	}
}

const (
	createAnnotationActionCallbackID = "create_annotation"
)

func (a *App) handleMessageActionInteraction(ctx context.Context, ii *ent.Integration, ic *slack.InteractionCallback) (slackintegration.ViewErrors, error) {
	if !slackintegration.HasUser(ctx) {
		return nil, nil
	}
	cw, cwErr := a.newClientWrapper(ii)
	if cwErr != nil {
		return nil, fmt.Errorf("failed to create client wrapper: %w", cwErr)
	}

	switch ic.CallbackID {
	case createAnnotationActionCallbackID:
		return nil, a.handleAnnotationModalInteraction(ctx, cw, ic)
	}
	return nil, fmt.Errorf("unknown message actions: %s", ic.CallbackID)
}

func (a *App) handleBlockActionsInteraction(ctx context.Context, ii *ent.Integration, ic *slack.InteractionCallback) (slackintegration.ViewErrors, error) {
	for _, action := range ic.ActionCallback.BlockActions {
		switch action.ActionID {
		case actionIdSituationNoticeUseful:
			return nil, a.rateSituationNotice(ctx, ic, action, true)
		case actionIdSituationNoticeNotUseful:
			return nil, a.rateSituationNotice(ctx, ic, action, false)
		case actionIdSituationNoticeOpen, actionIdSituationNoticeDeclareIncident, actionIdSituationNoticeOpenIncident:
			// The button opens its URL in the browser; Slack still reports the click.
			return nil, nil
		default:
			return nil, fmt.Errorf("unknown block action id: %s", action.ActionID)
		}
	}
	slog.Debug("interaction callback", "ic", ic)

	return nil, fmt.Errorf("unknown block actions: %s", ic.CallbackID)
}

// rateSituationNotice records the Slack user's rating of a situation's notice, replacing an earlier one,
// and acknowledges it to them alone. The Rezible user is recorded when the Slack user maps to one.
func (a *App) rateSituationNotice(ctx context.Context, ic *slack.InteractionCallback, action *slack.BlockAction, useful bool) error {
	situationID, parseErr := uuid.Parse(action.Value)
	if parseErr != nil {
		return fmt.Errorf("parse situation notice rating situation: %w", parseErr)
	}
	client := a.db.Client(ctx)
	querySituation := client.Situation.Query().
		Where(sit.ID(situationID))
	situationExists, situationErr := querySituation.Exist(ctx)
	if situationErr != nil {
		return fmt.Errorf("check rated situation: %w", situationErr)
	}
	if !situationExists {
		return nil
	}

	createRating := client.SituationNoticeRating.Create().
		SetSituationID(situationID).
		SetChatUserID(ic.User.ID).
		SetUseful(useful)
	if userID, hasUser := execution.GetContext(ctx).UserID(); hasUser {
		createRating.SetUserID(userID)
	}
	upsertRating := createRating.
		OnConflictColumns(sitnr.FieldTenantID, sitnr.FieldSituationID, sitnr.FieldChatUserID).
		UpdateUseful().
		UpdateUserID().
		UpdateUpdatedAt()
	if upsertErr := upsertRating.Exec(ctx); upsertErr != nil {
		return fmt.Errorf("record situation notice rating: %w", upsertErr)
	}

	acknowledgement := "Recorded as not useful. Thanks."
	if useful {
		acknowledgement = "Recorded as useful. Thanks."
	}
	reply := &slack.WebhookMessage{
		Text:            acknowledgement,
		ResponseType:    slack.ResponseTypeEphemeral,
		ReplaceOriginal: false,
	}
	if replyErr := slack.PostWebhookContext(ctx, ic.ResponseURL, reply); replyErr != nil {
		return fmt.Errorf("acknowledge situation notice rating: %w", replyErr)
	}
	return nil
}

func (a *App) handleViewSubmissionInteraction(ctx context.Context, ii *ent.Integration, ic *slack.InteractionCallback) (slackintegration.ViewErrors, error) {
	if !slackintegration.HasUser(ctx) {
		return nil, nil
	}
	switch ic.View.CallbackID {
	case viewCallbackIdAnnotationModal:
		return nil, a.handleAnnotationModalSubmission(ctx, ic)
	}
	return nil, fmt.Errorf("unknown view submission: %s", ic.View.CallbackID)
}

func (a *App) handleAnnotationModalInteraction(ctx context.Context, cw *slackintegration.ClientWrapper, ic *slack.InteractionCallback) error {
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

func (a *App) handleAnnotationModalSubmission(ctx context.Context, ic *slack.InteractionCallback) error {
	anno, annoErr := a.getAnnotationModalAnnotation(ctx, ic.View)
	if annoErr != nil {
		return fmt.Errorf("failed to get view annotation: %w", annoErr)
	}
	_, createErr := a.events.SetAnnotation(ctx, anno)
	if createErr != nil {
		return fmt.Errorf("failed to create annotation: %w", createErr)
	}
	return nil
}
