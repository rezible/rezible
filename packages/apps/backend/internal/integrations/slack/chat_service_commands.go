package slack

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/incident"
	"github.com/slack-go/slack"
)

var (
	supportedIncidentSubcommands = []string{"new", "update", "status", "help"}
	incidentCommandsFormatted    = strings.Join(supportedIncidentSubcommands, ", ")
)

func (s *ChatService) handleSlashCommand(baseCtx context.Context, cmd *slack.SlashCommand) error {
	ctx, usrErr := s.createUserContext(baseCtx, cmd.UserID)
	if usrErr != nil {
		return fmt.Errorf("failed to lookup user: %w", usrErr)
	}

	var handled bool
	var response *slack.Blocks
	var handlerErr error
	switch cmd.Command {
	case "/incident":
		handled = true
		response, handlerErr = s.handleIncidentCommand(ctx, cmd)
	}

	if handlerErr != nil {
		return fmt.Errorf("handling command: %w", handlerErr)
	}
	if !handled {
		s.logger.Debug("unknown slack command, ignoring", "command", cmd.Command)
	}
	if response != nil {
		_, msgErr := s.postEphemeralMessage(ctx, cmd.ChannelID, cmd.UserID, slack.MsgOptionBlocks(response.BlockSet...))
		if msgErr != nil {
			return fmt.Errorf("failed to post ephemeral message: %w", msgErr)
		}
	}
	return nil
}

func commandErrorResponse(message string) *slack.Blocks {
	return &slack.Blocks{
		BlockSet: []slack.Block{
			&slack.SectionBlock{
				Type: slack.MBTSection,
				Text: plainText(fmt.Sprintf("❌ %s", message)),
			},
		},
	}
}

func (s *ChatService) handleIncidentCommand(ctx context.Context, cmd *slack.SlashCommand) (*slack.Blocks, error) {
	// are we currently in an incident channel?
	var channelIncidentId uuid.UUID
	inc, incErr := s.incidents.Get(ctx, incident.ChatChannelID(cmd.ChannelID))
	if incErr != nil && !ent.IsNotFound(incErr) {
		slog.Error("unable to get incident by channel", "error", incErr)
		return commandErrorResponse(incErr.Error()), nil
	} else if inc != nil {
		channelIncidentId = inc.ID
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
			return commandErrorResponse("Not in an incident channel"), nil
		}
		view, viewErr := s.makeIncidentDetailsModalView(ctx, &meta)
		if viewErr != nil {
			slog.Error("failed creating incident details view", "error", viewErr)
			return commandErrorResponse("Failed to create incident details modal"), viewErr
		}
		if openModalErr := s.openModalView(ctx, cmd.TriggerID, *view); openModalErr != nil {
			return commandErrorResponse("Failed to open incident details modal"), openModalErr
		}
	} else if subcmd == "status" {
		if inc == nil {
			return commandErrorResponse("Not in an incident channel. Supported subcommands: " + incidentCommandsFormatted), nil
		}
		meta := incidentMilestoneModalViewMetadata{
			UserId:     cmd.UserID,
			IncidentId: channelIncidentId,
		}
		view, viewErr := s.makeIncidentMilestoneModalView(ctx, &meta)
		if viewErr != nil {
			slog.Error("failed creating incident milestone view", "error", viewErr)
			return commandErrorResponse("Failed to create incident milestone view"), viewErr
		}
		if openModalErr := s.openModalView(ctx, cmd.TriggerID, *view); openModalErr != nil {
			return commandErrorResponse("Failed to open incident milestone view"), openModalErr
		}
	} else if subcmd == "help" {
		// TODO
	} else {
		return commandErrorResponse(
			fmt.Sprintf("Invalid incident command '%s'. Supported subcommands: %s", subcmd, incidentCommandsFormatted),
		), nil
	}

	return nil, nil
}
