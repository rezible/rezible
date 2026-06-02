package slackintegration

import (
	"context"
	"fmt"
	"log/slog"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/slack-go/slack"
)

type ClientWrapper struct {
	client *slack.Client
	intg   *ent.Integration
}

func NewClientWrapper(intg *ent.Integration) (*ClientWrapper, error) {
	cfg, cfgErr := DecodeInstallationConfig(intg)
	if cfgErr != nil {
		return nil, fmt.Errorf("unable to decode integration config: %w", cfgErr)
	}
	cw := &ClientWrapper{
		intg:   intg,
		client: slack.New(cfg.AccessToken),
	}
	return cw, nil
}

func (w *ClientWrapper) Client() *slack.Client {
	return w.client
}

func (w *ClientWrapper) postMessage(ctx context.Context, channelId string, msgOpts ...slack.MsgOption) (string, error) {
	_, msgTs, msgErr := w.client.PostMessageContext(ctx, channelId, msgOpts...)
	return msgTs, msgErr
}

func (w *ClientWrapper) postEphemeralMessage(ctx context.Context, channelId, userId string, msgOpts ...slack.MsgOption) (string, error) {
	return w.client.PostEphemeralContext(ctx, channelId, userId, msgOpts...)
}

func (w *ClientWrapper) SendMessage(ctx context.Context, channelId string, content *rez.ContentNode) (string, error) {
	blocks := ConvertContentToBlocks("", content)
	return w.postMessage(ctx, channelId, slack.MsgOptionBlocks(blocks...))
}

func (w *ClientWrapper) SendTextMessage(ctx context.Context, channelId string, text string) (string, error) {
	return w.postMessage(ctx, channelId, slack.MsgOptionText(text, false))
}

func (w *ClientWrapper) SendReply(ctx context.Context, channelId string, threadId string, text string) (string, error) {
	return w.postMessage(ctx, channelId, slack.MsgOptionText(text, false), slack.MsgOptionTS(threadId))
}

func (w *ClientWrapper) OpenModalView(ctx context.Context, triggerId string, viewReq slack.ModalViewRequest) error {
	resp, respErr := w.client.OpenViewContext(ctx, triggerId, viewReq)
	if respErr != nil {
		LogSlackViewErrorResponse(slog.Default(), respErr, resp)
		return respErr
	}
	return nil
}

func (w *ClientWrapper) OpenOrUpdateModal(ctx context.Context, ic *slack.InteractionCallback, view *slack.ModalViewRequest) error {
	var viewResp *slack.ViewResponse
	var respErr error
	if ic.View.State == nil {
		viewResp, respErr = w.client.OpenViewContext(ctx, ic.TriggerID, *view)
	} else {
		viewResp, respErr = w.client.UpdateViewContext(ctx, *view, "", ic.Hash, ic.View.ID)
	}
	if respErr != nil {
		LogSlackViewErrorResponse(slog.Default(), respErr, viewResp)
		return respErr
	}

	return nil
}
