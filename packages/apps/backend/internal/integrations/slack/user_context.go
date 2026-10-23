package slackintegration

import (
	"context"
	"errors"
	"fmt"

	"github.com/slack-go/slack"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/user"
	"github.com/rezible/rezible/pkg/execution"
)

// ErrChatUserNotFound is a Slack user with no Rezible user: the workspace's member sync has not seen them yet.
var ErrChatUserNotFound = errors.New("slack user not found in rezible")

// UnknownUserText tells a Slack user that Rezible does not know them yet.
const UnknownUserText = "Rezible hasn't seen your Slack account yet. It is picked up the next time Rezible syncs this workspace's members."

// LookupUserContext adds the Rezible user with the Slack user's ID to the context. It returns
// ErrChatUserNotFound when there is none.
func LookupUserContext(ctx context.Context, users rez.UserService, chatID string) (context.Context, error) {
	usr, getErr := users.Get(ctx, user.ChatID(chatID))
	if ent.IsNotFound(getErr) {
		return nil, fmt.Errorf("%w: %s", ErrChatUserNotFound, chatID)
	}
	if getErr != nil {
		return nil, fmt.Errorf("lookup user by chat id: %w", getErr)
	}
	sess := &ent.UserAuthSession{
		TenantID: usr.TenantID,
		UserID:   usr.ID,
	}
	return execution.NewUserContext(ctx, sess), nil
}

// HasUser is whether the context has a Rezible user. Interactions from Slack users with none run as the
// installation's tenant, with no user.
func HasUser(ctx context.Context) bool {
	_, found := execution.GetContext(ctx).UserID()
	return found
}

// UnknownUserBlocks tells a Slack user that Rezible does not know them yet.
func UnknownUserBlocks() []slack.Block {
	return []slack.Block{slack.NewSectionBlock(PlainTextBlock(UnknownUserText), nil, nil)}
}
