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

func NewClientWrapper(intg *ent.Integration, opts ...slack.Option) (*ClientWrapper, error) {
	cfg, cfgErr := GetValidatedConfig(intg.InstallationConfig)
	if cfgErr != nil {
		return nil, fmt.Errorf("unable to decode integration config: %w", cfgErr)
	}
	cw := &ClientWrapper{
		intg:   intg,
		client: slack.New(cfg.AccessToken, opts...),
	}
	return cw, nil
}

// ClientOptions are the client options the Slack configuration asks for.
func ClientOptions(cfg rez.IntegrationsConfigSlack) []slack.Option {
	if cfg.ApiUrl == "" {
		return nil
	}
	return []slack.Option{slack.OptionAPIURL(cfg.ApiUrl)}
}

func (w *ClientWrapper) Integration() *ent.Integration {
	return w.intg
}

func (w *ClientWrapper) Client() *slack.Client {
	return w.client
}

func (w *ClientWrapper) PostMessage(ctx context.Context, channelId string, msgOpts ...slack.MsgOption) (string, error) {
	_, msgTs, msgErr := w.client.PostMessageContext(ctx, channelId, msgOpts...)
	return msgTs, msgErr
}

func (w *ClientWrapper) UpdateMessage(ctx context.Context, channelId, messageTs string, msgOpts ...slack.MsgOption) error {
	_, _, _, updateErr := w.client.UpdateMessageContext(ctx, channelId, messageTs, msgOpts...)
	return updateErr
}

func (w *ClientWrapper) GetPermalink(ctx context.Context, channelId, messageTs string) (string, error) {
	params := &slack.PermalinkParameters{Channel: channelId, Ts: messageTs}
	return w.client.GetPermalinkContext(ctx, params)
}

// GetChannelName returns the channel's name, as conversations.info reports it.
func (w *ClientWrapper) GetChannelName(ctx context.Context, channelId string) (string, error) {
	input := &slack.GetConversationInfoInput{ChannelID: channelId}
	channel, infoErr := w.client.GetConversationInfoContext(ctx, input)
	if infoErr != nil {
		return "", infoErr
	}
	return channel.Name, nil
}

func (w *ClientWrapper) PostEphemeralMessage(ctx context.Context, channelId, userId string, msgOpts ...slack.MsgOption) (string, error) {
	return w.client.PostEphemeralContext(ctx, channelId, userId, msgOpts...)
}

func (w *ClientWrapper) SendMessage(ctx context.Context, channelId string, content *rez.ContentNode) (string, error) {
	blocks := ConvertContentToBlocks("", content)
	return w.PostMessage(ctx, channelId, slack.MsgOptionBlocks(blocks...))
}

func (w *ClientWrapper) SendTextMessage(ctx context.Context, channelId string, text string) (string, error) {
	return w.PostMessage(ctx, channelId, slack.MsgOptionText(text, false))
}

func (w *ClientWrapper) SendReply(ctx context.Context, channelId string, threadId string, text string) (string, error) {
	return w.PostMessage(ctx, channelId, slack.MsgOptionText(text, false), slack.MsgOptionTS(threadId))
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
