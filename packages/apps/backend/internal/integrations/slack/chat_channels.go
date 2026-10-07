package slackintegration

import (
	"cmp"
	"context"
	"errors"
	"fmt"

	"github.com/slack-go/slack"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/pkg/integrations"
)

// ListChatChannels reads one page of channels with a single conversations.list call.
// A short or empty page may still have a next cursor.
func ListChatChannels(ctx context.Context, client *slack.Client, params integrations.ListChatChannelsParams) (*integrations.ChatChannelPage, error) {
	types := []string{"public_channel"}
	if params.IncludePrivate {
		types = append(types, "private_channel")
	}
	listParams := &slack.GetConversationsParameters{
		Types:           types,
		ExcludeArchived: true,
		Limit:           min(cmp.Or(params.Limit, integrations.DefaultChatChannelPageSize), integrations.MaxChatChannelPageSize),
		Cursor:          params.Cursor,
	}
	channels, nextCursor, listErr := client.GetConversationsContext(ctx, listParams)
	if listErr != nil {
		if slackErr, isSlackErr := errors.AsType[slack.SlackErrorResponse](listErr); isSlackErr {
			switch slackErr.Err {
			case "invalid_auth", "token_revoked", "account_inactive", "missing_scope":
				return nil, fmt.Errorf("%w: list slack channels: %w", rez.ErrForbidden, listErr)
			case "invalid_cursor":
				return nil, fmt.Errorf("%w: list slack channels: %w", rez.ErrInvalidInput, listErr)
			}
		}
		if _, isRateLimited := errors.AsType[*slack.RateLimitedError](listErr); isRateLimited {
			return nil, fmt.Errorf("%w: list slack channels: %w", rez.ErrRateLimited, listErr)
		}
		return nil, fmt.Errorf("list slack channels: %w", listErr)
	}

	page := &integrations.ChatChannelPage{
		Channels:   make([]integrations.ChatChannel, len(channels)),
		NextCursor: nextCursor,
	}
	for i, ch := range channels {
		page.Channels[i] = integrations.ChatChannel{
			ID:         ch.ID,
			Name:       ch.Name,
			IsPrivate:  ch.IsPrivate,
			IsArchived: ch.IsArchived,
			IsMember:   ch.IsMember,
		}
	}
	return page, nil
}
