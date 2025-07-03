package slack

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"net/http"

	rez "github.com/rezible/rezible"
)

type ChatProvider struct {
	client           *slack.Client
	webhookHandler   *webhookHandler
	messageAnnotator rez.ChatMessageAnnotator
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

func (p *ChatProvider) SetMessageAnnotator(ma rez.ChatMessageAnnotator) {
	p.messageAnnotator = ma
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

func (p *ChatProvider) SendMessage(ctx context.Context, id string, content *rez.ContentNode) error {
	return p.sendUserMessage(ctx, id, slack.MsgOptionBlocks(convertContentToBlocks(content, "")...))
}

func (p *ChatProvider) SendTextMessage(ctx context.Context, id string, text string) error {
	return p.sendUserMessage(ctx, id, slack.MsgOptionText(text, false))
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

func (p *ChatProvider) handleCreateAnnotationInteraction(ctx context.Context, ic *slack.InteractionCallback) error {
	if p.messageAnnotator == nil {
		return fmt.Errorf("no chat message annotator")
	}

	msgId := fmt.Sprintf("%s_%s", ic.Channel.ID, ic.Message.Timestamp)
	rosters, event, infoErr := p.messageAnnotator.QueryUserChatMessageEventDetails(ctx, ic.User.ID, msgId)
	if infoErr != nil {
		return fmt.Errorf("failed to get annotation information: %w", infoErr)
	}

	view, viewErr := makeCreateAnnotationModalView(ctx, ic, msgId, rosters, event)
	if viewErr != nil || view == nil {
		return fmt.Errorf("failed to create annotation view: %w", viewErr)
	}

	resp, respErr := p.client.OpenViewContext(ctx, ic.TriggerID, *view)
	if respErr != nil {
		logSlackViewErrorResponse(respErr, resp)
		return fmt.Errorf("open view: %w", respErr)
	}
	return nil
}

func (p *ChatProvider) handleBlockActionInteraction(ctx context.Context, ic *slack.InteractionCallback) error {
	if ic.View.CallbackID == createAnnotationModalViewCallbackID {
		view := makeUpdatedCreateAnnotationModalView(ic)

		resp, respErr := p.client.UpdateViewContext(ctx, view, "", "", ic.View.ID)
		if respErr != nil {
			logSlackViewErrorResponse(respErr, resp)
			return fmt.Errorf("open view: %w", respErr)
		}
	}
	return nil
}

func (p *ChatProvider) handleCreateAnnotationModalSubmission(ctx context.Context, ic *slack.InteractionCallback) error {
	anno, annoErr := getCreateAnnotationModalViewAnnotation(ic.View)
	if annoErr != nil {
		return fmt.Errorf("failed to get view annotation: %w", annoErr)
	}
	if p.messageAnnotator == nil {
		return fmt.Errorf("no chat message annotator")
	}
	_, createErr := p.messageAnnotator.CreateAnnotation(ctx, anno)
	return createErr
}
