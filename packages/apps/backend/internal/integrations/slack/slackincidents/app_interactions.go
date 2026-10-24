package slackincidents

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/slack-go/slack"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	slackintegration "github.com/rezible/rezible/internal/integrations/slack"
	"github.com/rezible/rezible/pkg/execution"
)

func (a *App) handleMessageActionInteraction(ctx context.Context, ii *ent.Integration, ic *slack.InteractionCallback) (slackintegration.ViewErrors, error) {
	return nil, fmt.Errorf("unknown message actions: %s", ic.CallbackID)
}

// handleBlockActionsInteraction opens the modal a button asks for. The incident is the button's value.
func (a *App) handleBlockActionsInteraction(ctx context.Context, ii *ent.Integration, ic *slack.InteractionCallback) (slackintegration.ViewErrors, error) {
	for _, action := range ic.ActionCallback.BlockActions {
		switch action.ActionID {
		case actionCallbackIdIncidentOpen:
			// The button opens its URL in the browser; Slack still reports the click.
			return nil, nil
		case actionCallbackIdIncidentDetailsModalButton:
			return nil, a.openIncidentModalFromButton(ctx, ii, ic, action, a.makeButtonDetailsModalView)
		case actionCallbackIdIncidentMilestoneModalButton:
			return nil, a.openIncidentModalFromButton(ctx, ii, ic, action, a.makeButtonMilestoneModalView)
		default:
			return nil, fmt.Errorf("unknown block action id: %s", action.ActionID)
		}
	}
	return nil, fmt.Errorf("no block actions: %s", ic.CallbackID)
}

func (a *App) openIncidentModalFromButton(ctx context.Context, ii *ent.Integration, ic *slack.InteractionCallback, action *slack.BlockAction, makeView func(context.Context, uuid.UUID) (*slack.ModalViewRequest, error)) error {
	if !slackintegration.HasUser(ctx) {
		return errUnknownUser
	}
	incidentID, parseErr := uuid.Parse(action.Value)
	if parseErr != nil {
		return fmt.Errorf("parse incident id: %w", parseErr)
	}
	view, viewErr := makeView(ctx, incidentID)
	if viewErr != nil {
		return viewErr
	}
	cw, cwErr := a.newClientWrapper(ii)
	if cwErr != nil {
		return fmt.Errorf("create client wrapper: %w", cwErr)
	}
	// A button inside a modal pushes onto it; one on a message opens a new modal.
	if ic.View.ID != "" {
		resp, pushErr := cw.Client().PushViewContext(ctx, ic.TriggerID, *view)
		if pushErr != nil {
			slackintegration.LogSlackViewErrorResponse(slog.Default(), pushErr, resp)
			return fmt.Errorf("push modal: %w", pushErr)
		}
		return nil
	}
	return cw.OpenModalView(ctx, ic.TriggerID, *view)
}

func (a *App) makeButtonDetailsModalView(ctx context.Context, incidentID uuid.UUID) (*slack.ModalViewRequest, error) {
	return a.makeIncidentDetailsModalView(ctx, &incidentDetailsModalViewMetadata{IncidentId: incidentID})
}

func (a *App) makeButtonMilestoneModalView(ctx context.Context, incidentID uuid.UUID) (*slack.ModalViewRequest, error) {
	inc, incErr := a.incidents.GetIncident(ctx, incident.ID(incidentID))
	if incErr != nil {
		return nil, fmt.Errorf("get incident: %w", incErr)
	}
	return makeIncidentMilestoneModalView(inc)
}

// handleViewSubmissionInteraction saves a submitted modal. Errors are shown on the modal, which stays open.
func (a *App) handleViewSubmissionInteraction(ctx context.Context, ii *ent.Integration, ic *slack.InteractionCallback) (slackintegration.ViewErrors, error) {
	userID, hasUser := execution.GetContext(ctx).UserID()
	if !hasUser {
		return nil, errUnknownUser
	}
	if ic.View.State == nil {
		return nil, fmt.Errorf("missing view state")
	}
	switch ic.View.CallbackID {
	case viewCallbackIdIncidentDetailsModal:
		return nil, a.handleIncidentDetailsModalSubmission(ctx, ii, ic, userID)
	case viewCallbackIdIncidentMilestoneModal:
		return nil, a.handleIncidentMilestoneModalSubmission(ctx, ic, userID)
	}
	return nil, fmt.Errorf("unknown view submission: %s", ic.View.CallbackID)
}

func (a *App) handleIncidentDetailsModalSubmission(ctx context.Context, ii *ent.Integration, ic *slack.InteractionCallback, userID uuid.UUID) error {
	var meta incidentDetailsModalViewMetadata
	if jsonErr := json.Unmarshal([]byte(ic.View.PrivateMetadata), &meta); jsonErr != nil {
		return fmt.Errorf("unmarshal incident details modal metadata: %w", jsonErr)
	}

	details := incidentDetailsModalInput(ic.View.State)
	if meta.IncidentId != uuid.Nil {
		if _, updateErr := a.incidents.UpdateIncident(ctx, meta.IncidentId, details); updateErr != nil {
			return fmt.Errorf("update incident details: %w", updateErr)
		}
		return nil
	}

	declareParams := rez.DeclareIncidentParams{
		Title:        *details.Title,
		Summary:      *details.Summary,
		SeverityID:   details.SeverityID,
		TypeID:       details.TypeID,
		DeclaredByID: userID,
	}
	if details.FieldOptionIDs != nil {
		declareParams.FieldOptionIDs = *details.FieldOptionIDs
	}
	if details.TagIDs != nil {
		declareParams.TagIDs = *details.TagIDs
	}
	inc, declareErr := a.management.Declare(ctx, declareParams)
	if declareErr != nil {
		return fmt.Errorf("declare incident: %w", declareErr)
	}
	a.confirmDeclared(ctx, ii, meta.CommandChannelId, ic.User.ID, inc)
	return nil
}

// confirmDeclared tells the declarer, privately in the channel where they ran the command, that the incident
// was declared. The command may have run where the app can't post, such as a DM, so a failure is only logged.
func (a *App) confirmDeclared(ctx context.Context, ii *ent.Integration, channelID string, slackUserID string, inc *ent.Incident) {
	incUrl, urlErr := a.cfg.App.GetFrontendUrl("incidents", inc.Slug)
	if urlErr != nil {
		slog.ErrorContext(ctx, "incident url", "error", urlErr, "incident_id", inc.ID.String())
		return
	}
	text := fmt.Sprintf(":white_check_mark: Incident declared: *%s* · <%s|Open in Rezible>. You'll be invited to its channel.",
		escaped(inc.Title, maxTitleRunes), incUrl)
	cw, cwErr := a.newClientWrapper(ii)
	if cwErr != nil {
		slog.ErrorContext(ctx, "create client wrapper", "error", cwErr)
		return
	}
	section := slack.NewSectionBlock(slackintegration.MarkdownBlock(text), nil, nil)
	msgOpts := []slack.MsgOption{
		slack.MsgOptionText("Incident declared: "+escaped(inc.Title, maxTitleRunes), false),
		slack.MsgOptionParse(false),
		slack.MsgOptionBlocks(section),
	}
	if _, postErr := cw.PostEphemeralMessage(ctx, channelID, slackUserID, msgOpts...); postErr != nil {
		slog.WarnContext(ctx, "failed to confirm incident declaration", "error", postErr, "incident_id", inc.ID.String())
	}
}

// handleIncidentMilestoneModalSubmission records the chosen status. A rejected milestone is shown on the
// status block, the modal's first.
func (a *App) handleIncidentMilestoneModalSubmission(ctx context.Context, ic *slack.InteractionCallback, userID uuid.UUID) error {
	meta, metaErr := getIncidentMilestoneModalViewMetadata(ic.View)
	if metaErr != nil {
		return metaErr
	}
	params := incidentMilestoneModalInput(ic.View.State)
	params.IncidentID = meta.IncidentId
	params.Timestamp = time.Now()
	params.UserID = &userID
	params.Source = slackintegration.ProviderName
	if _, addErr := a.management.AddMilestone(ctx, params); addErr != nil {
		return fmt.Errorf("add milestone: %w", addErr)
	}
	return nil
}
