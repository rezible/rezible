package slack

import (
	"context"

	"github.com/slack-go/slack"
)

func (s *ChatService) handleSlashCommand(ctx context.Context, ev *slack.SlashCommand) (bool, any, error) {
	// TODO: queue?
	return true, nil, nil
}
