package slack

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
)

func commandErrorResponse(message string) *slack.Msg {
	return &slack.Msg{
		Text: message,
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				&slack.SectionBlock{
					Type: slack.MBTSection,
					Text: &slack.TextBlockObject{
						Type: slack.PlainTextType,
						Text: fmt.Sprintf("❌ %s", message),
					},
				},
			},
		},
	}
}

func (s *ChatService) handleSlashCommand(ctx context.Context, ev *slack.SlashCommand) (bool, *slack.Msg, error) {
	switch ev.Command {
	case "/incident":
		payload, handlerErr := s.handleIncidentCommand(ctx, ev)
		return true, payload, handlerErr
	default:
		return false, nil, nil
	}
}

func (s *ChatService) handleIncidentCommand(ctx context.Context, ev *slack.SlashCommand) (*slack.Msg, error) {
	// are we currently in an incident channel?
	var channelIncidentId uuid.UUID
	inc, incErr := s.incidents.GetByChatChannelID(ctx, ev.ChannelID)
	if incErr != nil && !ent.IsNotFound(incErr) {
		log.Error().Err(incErr).Msg("unable to get incident by channel")
		return commandErrorResponse(incErr.Error()), nil
	} else if inc != nil {
		channelIncidentId = inc.ID
	}

	subcmd := ""
	if args := strings.Split(ev.Text, " "); len(args) > 0 {
		subcmd = args[0]
		log.Debug().Str("text", ev.Text).Str("subcmd", subcmd).Msg("incident command")
	}

	if subcmd == "" || subcmd == "update" || subcmd == "new" {
		meta := incidentDetailsModalViewMetadata{
			CommandChannelId: ev.ChannelID,
			UserId:           ev.UserID,
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
			log.Error().Err(viewErr).Msg("failed creating incident details view")
			return commandErrorResponse("Failed to create incident details modal"), viewErr
		}
		if openModalErr := s.openModalView(ctx, ev.TriggerID, *view); openModalErr != nil {
			return commandErrorResponse("Failed to open incident details modal"), openModalErr
		}
	} else if subcmd == "status" {
		meta := incidentMilestoneModalViewMetadata{
			UserId:     ev.UserID,
			IncidentId: channelIncidentId,
		}
		view, viewErr := s.makeIncidentMilestoneModalView(ctx, &meta)
		if viewErr != nil {
			log.Error().Err(viewErr).Msg("failed creating incident milestone view")
			return commandErrorResponse("Failed to create incident milestone view"), viewErr
		}
		if openModalErr := s.openModalView(ctx, ev.TriggerID, *view); openModalErr != nil {
			return commandErrorResponse("Failed to open incident milestone view"), openModalErr
		}
	} else {
		return commandErrorResponse(fmt.Sprintf("Invalid incident command '%s'", subcmd)), nil
	}

	return nil, nil
}
