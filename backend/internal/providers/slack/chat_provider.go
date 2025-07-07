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

	msgCtxProvider *rez.ChatMessageContextProvider
	mentionHandler rez.ChatMentionHandler
}

type ChatProviderConfig struct {
	BotApiKey     string `json:"bot_api_key"`
	SigningSecret string `json:"signing_secret"`
}

func NewChatProvider(cfg ChatProviderConfig) (*ChatProvider, error) {
	p := &ChatProvider{}
	p.client = slack.New(cfg.BotApiKey)
	p.webhookHandler = newWebhookHandler(cfg.SigningSecret, p)
	return p, nil
}

func (p *ChatProvider) SetMessageContextProvider(cp rez.ChatMessageContextProvider) {
	p.msgCtxProvider = &cp
}

func (p *ChatProvider) SetMentionHandler(handler rez.ChatMentionHandler) {
	p.mentionHandler = handler
}

func (p *ChatProvider) GetWebhooks() rez.Webhooks {
	return rez.Webhooks{
		"slack/options":     http.HandlerFunc(p.webhookHandler.handleOptions),
		"slack/events":      http.HandlerFunc(p.webhookHandler.handleEvents),
		"slack/interaction": http.HandlerFunc(p.webhookHandler.handleInteractions),
	}
}

func (p *ChatProvider) lookupUser(ctx context.Context, userId string) (*ent.User, error) {
	if p.msgCtxProvider == nil {
		return nil, errors.New("msg ctx provider not initialized")
	}
	return p.msgCtxProvider.LookupChatUserFn(ctx, userId)
}

func (p *ChatProvider) sendMessage(ctx context.Context, channelId string, msgOpts ...slack.MsgOption) error {
	_, _, msgErr := p.client.PostMessageContext(ctx, channelId, msgOpts...)
	return msgErr
}

//func (p *ChatProvider) sendUserMessage(ctx context.Context, id string, msg slack.MsgOption) error {
//	params := &slack.OpenConversationParameters{Users: []string{id}}
//	convo, _, _, convoErr := p.client.OpenConversationContext(ctx, params)
//	if convoErr != nil {
//		return fmt.Errorf("failed to open conversation with user %s: %w", id, convoErr)
//	}
//
//	if sendErr := p.sendMessage(ctx, convo.ID, msg); sendErr != nil {
//		return fmt.Errorf("send user %s message: %w", id, sendErr)
//	}
//
//	return nil
//}

func (p *ChatProvider) SendMessage(ctx context.Context, channelId string, content *rez.ContentNode) error {
	return p.sendMessage(ctx, channelId, slack.MsgOptionBlocks(convertContentToBlocks(content, "")...))
}

func (p *ChatProvider) SendTextMessage(ctx context.Context, channelId string, text string) error {
	return p.sendMessage(ctx, channelId, slack.MsgOptionText(text, false))
}

func (p *ChatProvider) SendReply(ctx context.Context, channelId string, threadId string, text string) error {
	return p.sendMessage(ctx, channelId, slack.MsgOptionText(text, false), slack.MsgOptionTS(threadId))
}

func (p *ChatProvider) SendOncallHandover(ctx context.Context, params rez.SendOncallHandoverParams) error {
	channel, msg, err := buildHandoverMessage(params)
	if err != nil {
		return fmt.Errorf("creating handover message: %w", err)
	}
	return p.sendMessage(ctx, channel, msg)
}

func (p *ChatProvider) onMentionEvent(data *slackevents.AppMentionEvent) {
	if p.mentionHandler == nil {
		log.Warn().Msg("chat mention handler not initialized")
		return
	}

	replyTs := data.TimeStamp
	if data.ThreadTimeStamp != "" {
		replyTs = data.ThreadTimeStamp
	}

	p.mentionHandler(&rez.ChatMentionEvent{
		ChatId:      data.Channel,
		ThreadId:    replyTs,
		UserId:      data.User,
		MessageText: data.Text,
	})
}

func (p *ChatProvider) onMessageEvent(data *slackevents.MessageEvent) {
	threadTs := data.ThreadTimeStamp
	// TODO check if thread is 'monitored'

	log.Debug().
		Str("type", data.ChannelType).
		Str("text", data.Text).
		Str("thread", threadTs).
		Str("user", data.User).
		Msg("message")
}

func (p *ChatProvider) onAssistantThreadStartedEvent(data *slackevents.AssistantThreadStartedEvent) {
	log.Debug().Msg("assistant thread started")
}

func (p *ChatProvider) onUserHomeOpenedEvent(data *slackevents.AppHomeOpenedEvent) {
	ctx := context.Background()
	usr, usrErr := p.lookupUser(ctx, data.User)
	if usrErr != nil {
		log.Warn().Err(usrErr).Msg("failed to lookup user")
		return
	}
	homeView, viewErr := makeUserHomeView(ctx, usr)
	if viewErr != nil || homeView == nil {
		log.Error().Err(viewErr).Msg("failed to create user home view")
		return
	}
	resp, publishErr := p.client.PublishViewContext(ctx, data.User, *homeView, "")
	if publishErr != nil {
		logSlackViewErrorResponse(publishErr, resp)
	}
}

func (p *ChatProvider) handleAnnotationModalInteraction(ctx context.Context, ic *slack.InteractionCallback) error {
	view, viewErr := makeAnnotationModalView(ctx, ic, p.msgCtxProvider)
	if viewErr != nil || view == nil {
		return fmt.Errorf("failed to create annotation view: %w", viewErr)
	}

	var resp *slack.ViewResponse
	var respErr error
	if ic.View.State == nil {
		resp, respErr = p.client.OpenViewContext(ctx, ic.TriggerID, *view)
	} else {
		resp, respErr = p.client.UpdateViewContext(ctx, *view, "", ic.Hash, ic.View.ID)
	}
	if respErr != nil {
		logSlackViewErrorResponse(respErr, resp)
		return fmt.Errorf("annotation modal view: %w", respErr)
	}
	return nil
}

func (p *ChatProvider) handleAnnotationModalSubmission(ctx context.Context, ic *slack.InteractionCallback) error {
	anno, annoErr := getAnnotationModalAnnotation(ctx, ic.View, p.msgCtxProvider)
	if annoErr != nil {
		return fmt.Errorf("failed to get view annotation: %w", annoErr)
	}
	_, createErr := p.msgCtxProvider.AnnotateMessageFn(ctx, anno)
	return createErr
}
