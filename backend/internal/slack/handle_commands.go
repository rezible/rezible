package slack

import (
	"context"
	"fmt"

	"github.com/rezible/rezible/ent"
	"github.com/slack-go/slack"
)

func (s *ChatService) handleSlashCommand(ctx context.Context, ev *slack.SlashCommand) (bool, any, error) {
	var userErr error
	ctx, userErr = s.getChatUserContext(ctx, ev.UserID)
	if userErr != nil {
		return false, nil, fmt.Errorf("failed to lookup user: %w", userErr)
	}

	switch ev.Command {
	case "/incident":
		payload, handlerErr := s.handleIncidentCommand(ctx, ev)
		return true, payload, handlerErr
	default:
		return false, nil, nil
	}
}

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

func (s *ChatService) handleIncidentCommand(ctx context.Context, ev *slack.SlashCommand) (any, error) {
	meta := incidentModalViewMetadata{
		CommandChannelId: ev.ChannelID,
		UserId:           ev.UserID,
	}

	// are we currently in an incident channel?
	inc, incErr := s.incidents.GetByChatChannelID(ctx, ev.ChannelID)
	if incErr != nil && !ent.IsNotFound(incErr) {
		return commandErrorResponse(incErr.Error()), nil
	}
	if inc != nil {
		meta.IncidentId = inc.ID
	}

	view, viewErr := s.makeIncidentModalView(ctx, &meta)
	if viewErr != nil {
		return commandErrorResponse("Failed to create incident view"), viewErr
	}

	openViewErr := s.withClient(ctx, func(client *slack.Client) error {
		resp, respErr := client.OpenViewContext(ctx, ev.TriggerID, *view)
		if respErr != nil {
			logSlackViewErrorResponse(respErr, resp)
			return respErr
		}
		return nil
	})
	if openViewErr != nil {
		return commandErrorResponse("Failed to open view"), openViewErr
	}

	return nil, nil
}
