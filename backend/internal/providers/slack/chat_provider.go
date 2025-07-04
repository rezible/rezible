package slack

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type ChatProvider struct {
	client         *slack.Client
	webhookHandler *webhookHandler

	annotateMessageFn    rez.AnnotateMessageFn
	lookupUserFn         rez.LookupChatUserFn
	lookupMessageEventFn rez.LookupChatMessageEventFn
}

type ChatProviderConfig struct {
	BotApiKey     string `json:"bot_api_key"`
	SigningSecret string `json:"signing_secret"`
}

func NewChatProvider(cfg ChatProviderConfig) (*ChatProvider, error) {
	p := &ChatProvider{
		annotateMessageFn: func(ctx context.Context, anno *ent.OncallAnnotation) (*ent.OncallAnnotation, error) {
			return nil, errors.New("no annotateMessageFn supplied")
		},
		lookupUserFn: func(ctx context.Context, chatId string) (*ent.User, error) {
			return nil, errors.New("no lookupUserFn supplied")
		},
		lookupMessageEventFn: func(ctx context.Context, msgId string) (*ent.OncallEvent, error) {
			return nil, errors.New("no lookupMessageEventFn supplied")
		},
	}
	p.client = slack.New(cfg.BotApiKey)
	p.webhookHandler = newWebhookHandler(cfg.SigningSecret, p)
	return p, nil
}

func (p *ChatProvider) SetAnnotateMessageFn(fn func(ctx context.Context, anno *ent.OncallAnnotation) (*ent.OncallAnnotation, error)) {
	p.annotateMessageFn = fn
}
func (p *ChatProvider) SetUserLookupFn(fn func(context.Context, string) (*ent.User, error)) {
	p.lookupUserFn = fn
}
func (p *ChatProvider) SetMessageEventLookupFn(fn func(context.Context, string) (*ent.OncallEvent, error)) {
	p.lookupMessageEventFn = fn
}

func (p *ChatProvider) GetWebhooks() rez.Webhooks {
	return rez.Webhooks{
		"slack/options":     http.HandlerFunc(p.webhookHandler.handleOptions),
		"slack/events":      http.HandlerFunc(p.webhookHandler.handleEvents),
		"slack/interaction": http.HandlerFunc(p.webhookHandler.handleInteractions),
	}
}

func (p *ChatProvider) sendMessage(ctx context.Context, channel string, msg slack.MsgOption) error {
	_, _, msgErr := p.client.PostMessageContext(ctx, channel, msg)
	return msgErr
}

func (p *ChatProvider) sendUserMessage(ctx context.Context, id string, msg slack.MsgOption) error {
	params := &slack.OpenConversationParameters{Users: []string{id}}
	convo, _, _, convoErr := p.client.OpenConversationContext(ctx, params)
	if convoErr != nil {
		return fmt.Errorf("failed to open conversation with user %s: %w", id, convoErr)
	}

	if sendErr := p.sendMessage(ctx, convo.ID, msg); sendErr != nil {
		return fmt.Errorf("send user %s message: %w", id, sendErr)
	}

	return nil
}

func (p *ChatProvider) SendMessage(ctx context.Context, userId string, content *rez.ContentNode) error {
	return p.sendMessage(ctx, userId, slack.MsgOptionBlocks(convertContentToBlocks(content, "")...))
}

func (p *ChatProvider) SendTextMessage(ctx context.Context, userId string, text string) error {
	return p.sendMessage(ctx, userId, slack.MsgOptionText(text, false))
}

func (p *ChatProvider) SendOncallHandover(ctx context.Context, params rez.SendOncallHandoverParams) error {
	mb, builderErr := newHandoverMessageBuilder(params.EndingShift, params.StartingShift, params.PinnedAnnotations)
	if builderErr != nil {
		return fmt.Errorf("failed to create handover message builder: %w", builderErr)
	}

	if buildErr := mb.build(params.Content); buildErr != nil {
		return fmt.Errorf("building handover message: %w", buildErr)
	}

	return p.sendMessage(ctx, mb.getChannel(), mb.getMessage())
}

func (p *ChatProvider) onMentionEvent(data *slackevents.AppMentionEvent) {
	_, _, msgErr := p.client.PostMessage(data.Channel, slack.MsgOptionText("hello", false))
	if msgErr != nil {
		log.Warn().Err(msgErr).Msg("failed to message")
	}
}

func logSlackViewErrorResponse(err error, resp *slack.ViewResponse) {
	if resp != nil {
		log.Debug().
			Strs("messages", resp.ResponseMetadata.Messages).
			Msg("publish response")
	}
	log.Error().Err(err).Msg("slack view response error")
}

func (p *ChatProvider) onUserHomeOpenedEvent(data *slackevents.AppHomeOpenedEvent) {
	ctx := context.Background()
	homeView, hash, viewErr := makeUserHomeView(ctx)
	if viewErr != nil || homeView == nil {
		log.Error().Err(viewErr).Msg("failed to create user home view")
		return
	}
	resp, publishErr := p.client.PublishViewContext(ctx, data.User, *homeView, hash)
	if publishErr != nil {
		logSlackViewErrorResponse(publishErr, resp)
	}
}

func (p *ChatProvider) onMessageEvent(data *slackevents.MessageEvent) {
	log.Debug().Interface("data", data).Msg("slack message event")
}

func (p *ChatProvider) handleAnnotationModalInteraction(ctx context.Context, ic *slack.InteractionCallback) error {
	c, ctxErr := makeAnnotationViewContext(ctx, ic, p.lookupUserFn, p.lookupMessageEventFn)
	if ctxErr != nil {
		return fmt.Errorf("failed to get message annotation context: %w", ctxErr)
	}

	view, viewErr := makeAnnotationModalView(c)
	if viewErr != nil || view == nil {
		return fmt.Errorf("failed to create annotation view: %w", viewErr)
	}

	var resp *slack.ViewResponse
	var respErr error
	if ic.View.State == nil {
		resp, respErr = p.client.OpenViewContext(ctx, ic.TriggerID, *view)
	} else {
		resp, respErr = p.client.UpdateViewContext(ctx, *view, "", "", ic.View.ID)
	}
	if respErr != nil {
		logSlackViewErrorResponse(respErr, resp)
		return fmt.Errorf("annotation modal view: %w", respErr)
	}
	return nil
}

func (p *ChatProvider) handleAnnotationModalSubmission(ctx context.Context, ic *slack.InteractionCallback) error {
	anno, annoErr := getAnnotationModalAnnotation(ctx, ic.View, p.lookupUserFn)
	if annoErr != nil {
		return fmt.Errorf("failed to get view annotation: %w", annoErr)
	}
	_, createErr := p.annotateMessageFn(ctx, anno)
	return createErr
}
