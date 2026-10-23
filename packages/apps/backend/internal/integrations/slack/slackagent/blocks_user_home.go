package slackagent

import (
	"context"
	"errors"
	"fmt"

	"github.com/slack-go/slack"

	"github.com/rezible/rezible/internal/integrations/slack"
	"github.com/rezible/rezible/pkg/execution"
)

const viewCallbackIdUserHome = "user_home"

// userHomeView is the Slack user's Home tab, or one saying Rezible does not know them yet.
func (a *App) userHomeView(ctx context.Context, chatID string) (*slack.HomeTabViewRequest, error) {
	userCtx, lookupErr := slackintegration.LookupUserContext(ctx, a.users, chatID)
	if errors.Is(lookupErr, slackintegration.ErrChatUserNotFound) {
		unknown := slack.HomeTabViewRequest{
			Type:   slack.VTHomeTab,
			Blocks: slack.Blocks{BlockSet: slackintegration.UnknownUserBlocks()},
		}
		return &unknown, nil
	}
	if lookupErr != nil {
		return nil, fmt.Errorf("lookup user: %w", lookupErr)
	}
	return makeUserHomeViewRequest(userCtx)
}

func makeUserHomeViewRequest(ctx context.Context) (*slack.HomeTabViewRequest, error) {
	var blocks []slack.Block
	blocks = append(blocks, slack.NewSectionBlock(slackintegration.PlainTextBlock("Home Tab"), nil, nil))
	userId, userOk := execution.GetContext(ctx).UserID()
	if !userOk {
		return nil, fmt.Errorf("no user context")
	}
	homeView := slack.HomeTabViewRequest{
		Type:            slack.VTHomeTab,
		CallbackID:      viewCallbackIdUserHome,
		PrivateMetadata: "foo",
		Blocks:          slack.Blocks{BlockSet: blocks},
		ExternalID:      userId.String(),
	}
	return &homeView, nil
}
