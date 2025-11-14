package slack

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rezible/rezible/ent"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
)

func (s *ChatService) openOrUpdateModal(ctx context.Context, ic *slack.InteractionCallback, view *slack.ModalViewRequest) error {
	var viewResp *slack.ViewResponse

	client, clientErr := s.getClient(ctx)
	if clientErr != nil {
		return fmt.Errorf("get client: %w", clientErr)
	}

	var respErr error
	if ic.View.State == nil {
		viewResp, respErr = client.OpenViewContext(ctx, ic.TriggerID, *view)
	} else {
		viewResp, respErr = client.UpdateViewContext(ctx, *view, "", ic.Hash, ic.View.ID)
	}
	if respErr != nil {
		logSlackViewErrorResponse(respErr, viewResp)
		return respErr
	}
	return nil
}

func logSlackViewErrorResponse(err error, resp *slack.ViewResponse) {
	if resp != nil {
		log.Debug().
			Strs("messages", resp.ResponseMetadata.Messages).
			Msg("publish response")
	}
	log.Error().Err(err).Msg("slack view response error")
}

func convertSlackTs(ts string) time.Time {
	parts := strings.Split(ts, ".")
	if len(parts) < 2 {
		return time.Time{}
	}
	secs, parseErr := strconv.ParseInt(parts[0], 10, 32)
	if parseErr != nil {
		return time.Time{}
	}
	return time.Unix(secs, 0)
}

type messageId string

func (m messageId) getTimestamp() time.Time {
	_, ts, _ := strings.Cut(m.String(), "_")
	return convertSlackTs(ts)
}

func (m messageId) String() string {
	return string(m)
}

func getMessageId(ic *slack.InteractionCallback) messageId {
	return messageId(fmt.Sprintf("%s_%s", ic.Channel.ID, ic.Message.Timestamp))
}

func getViewStateBlockAction(state *slack.ViewState, blockId string, inputId string) *slack.BlockAction {
	if block, blockOk := state.Values[blockId]; blockOk {
		if action, inputOk := block[inputId]; inputOk {
			return &action
		}
	}
	return nil
}

func makeUserHomeView(ctx context.Context, user *ent.User) (*slack.HomeTabViewRequest, error) {
	var blocks []slack.Block
	blocks = append(blocks, slack.NewSectionBlock(plainTextBlock("Home Tab"), nil, nil))
	homeView := slack.HomeTabViewRequest{
		Type:            slack.VTHomeTab,
		CallbackID:      "user_home",
		PrivateMetadata: "foo",
		Blocks:          slack.Blocks{BlockSet: blocks},
		ExternalID:      user.ID.String(),
	}
	return &homeView, nil
}
