package slackagent

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"

	"github.com/rezible/rezible/internal/integrations/slack"
	"github.com/rezible/rezible/pkg/execution"
)

const viewCallbackIdUserHome = "user_home"

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
