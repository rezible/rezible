package slack

import (
	"context"
	"fmt"
	"net/http"

	"github.com/slack-go/slack"

	rez "github.com/rezible/rezible"
)

type ChatProvider struct {
	client         *slack.Client
	webhookHandler *webhookHandler
}

type ChatProviderConfig struct {
	BotApiKey     string `json:"bot_api_key"`
	SigningSecret string `json:"signing_secret"`
}

func NewChatProvider(cfg ChatProviderConfig) (*ChatProvider, error) {
	client := slack.New(cfg.BotApiKey)
	p := &ChatProvider{
		client:         client,
		webhookHandler: newWebhookHandler(cfg.SigningSecret, client),
	}

	return p, nil
}

func (p *ChatProvider) SetMessageAnnotator(ma rez.ChatMessageAnnotator) {
	p.webhookHandler.messageAnnotator = ma
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
	convo, _, _, convoErr := p.client.OpenConversationContext(ctx, &slack.OpenConversationParameters{
		Users: []string{id},
	})
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
