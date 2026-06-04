package slackincidents

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	"github.com/rezible/rezible/internal/integrations/slack"
	"github.com/slack-go/slack"
)

var (
	supportedIncidentSubcommands = []string{"new", "update", "status", "help"}
	incidentCommandsFormatted    = strings.Join(supportedIncidentSubcommands, ", ")
)

func (a *App) handleIncidentCommand(ctx context.Context, ii *ent.Integration, cmd *slack.SlashCommand) (*slack.Blocks, error) {
	// are we currently in an incident channel?
	var channelIncidentId uuid.UUID
	inc, incErr := a.incidents.Get(ctx, incident.ChatChannelID(cmd.ChannelID))
	if incErr != nil && !ent.IsNotFound(incErr) {
		slog.Error("unable to get incident by channel", "error", incErr)
		return slackintegration.CommandErrorResponse(incErr.Error()), nil
	} else if inc != nil {
		channelIncidentId = inc.ID
	}

	cw, cwErr := slackintegration.NewClientWrapper(ii)
	if cwErr != nil {
		return nil, fmt.Errorf("failed to create client wrapper: %w", cwErr)
	}

	subcmd := ""
	if args := strings.Split(cmd.Text, " "); len(args) > 0 {
		subcmd = args[0]
		slog.Debug("incident command", "text", cmd.Text, "subcmd", subcmd)
	}

	if subcmd == "" || subcmd == "update" || subcmd == "new" {
		meta := incidentDetailsModalViewMetadata{
			CommandChannelId: cmd.ChannelID,
			UserId:           cmd.UserID,
			IncidentId:       channelIncidentId,
		}
		if subcmd == "new" {
			meta.IncidentId = uuid.Nil
		}
		if subcmd == "update" && inc == nil {
			return slackintegration.CommandErrorResponse("Not in an incident channel"), nil
		}
		// TODO: load these properly
		prefs := UserSettingsIncidents{}
		view, viewErr := a.makeIncidentDetailsModalView(ctx, prefs, &meta)
		if viewErr != nil {
			slog.Error("failed creating incident details view", "error", viewErr)
			return slackintegration.CommandErrorResponse("Failed to create incident details modal"), viewErr
		}
		if openModalErr := cw.OpenModalView(ctx, cmd.TriggerID, *view); openModalErr != nil {
			return slackintegration.CommandErrorResponse("Failed to open incident details modal"), openModalErr
		}
	} else if subcmd == "status" {
		if inc == nil {
			return slackintegration.CommandErrorResponse("Not in an incident channel. Supported subcommands: " + incidentCommandsFormatted), nil
		}
		meta := incidentMilestoneModalViewMetadata{
			UserId:     cmd.UserID,
			IncidentId: channelIncidentId,
		}
		view, viewErr := a.makeIncidentMilestoneModalView(ctx, &meta)
		if viewErr != nil {
			slog.Error("failed creating incident milestone view", "error", viewErr)
			return slackintegration.CommandErrorResponse("Failed to create incident milestone view"), viewErr
		}
		if openModalErr := cw.OpenModalView(ctx, cmd.TriggerID, *view); openModalErr != nil {
			return slackintegration.CommandErrorResponse("Failed to open incident milestone view"), openModalErr
		}
	} else if subcmd == "help" {
		// TODO
	} else {
		return slackintegration.CommandErrorResponse(
			fmt.Sprintf("Invalid incident command '%s'. Supported subcommands: %s", subcmd, incidentCommandsFormatted),
		), nil
	}

	return nil, nil
}
