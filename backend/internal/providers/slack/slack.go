package slack

import (
	"context"
	"fmt"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/user"
	"github.com/slack-go/slack"
	"net/http"
)

type ChatProvider struct {
	client        *slack.Client
	signingSecret string

	lookupUser func(context.Context, string) (*ent.User, error)
}

type ChatProviderConfig struct {
	BotApiKey     string `json:"bot_api_key"`
	SigningSecret string `json:"signing_secret"`
}

func NewChatProvider(cfg ChatProviderConfig) (*ChatProvider, error) {
	client := slack.New(cfg.BotApiKey)
	p := &ChatProvider{
		client:        client,
		signingSecret: cfg.SigningSecret,
		lookupUser: func(ctx context.Context, s string) (*ent.User, error) {
			return nil, fmt.Errorf("no user lookup func registered")
		},
	}

	return p, nil
}

func (p *ChatProvider) GetWebhooks() rez.Webhooks {
	return rez.Webhooks{
		"slack/events":      http.HandlerFunc(p.handleEventsWebhook),
		"slack/interaction": http.HandlerFunc(p.handleInteractionWebhook),
		"slack/options":     http.HandlerFunc(p.handleOptionsWebhook),
	}
}

func (p *ChatProvider) SetUserLookupFunc(lookupFn func(ctx context.Context, id string) (*ent.User, error)) {
	p.lookupUser = lookupFn
}

func (p *ChatProvider) SendUserMessage(ctx context.Context, id string, msg string) error {
	return p.sendUserMessage(ctx, id, slack.MsgOptionText(msg, false))
}

func (p *ChatProvider) sendUserMessage(ctx context.Context, id string, msg slack.MsgOption) error {
	convo, _, _, convoErr := p.client.OpenConversationContext(ctx, &slack.OpenConversationParameters{
		Users: []string{id},
	})
	if convoErr != nil {
		return fmt.Errorf("failed to open conversation with user %s: %w", id, convoErr)
	}

	if sendErr := p.sendMessage(ctx, convo.ID, msg); sendErr != nil {
		return fmt.Errorf("send user '%s' message: %w", user.ID, sendErr)
	}

	return nil
}

func (p *ChatProvider) sendMessage(ctx context.Context, channel string, msg slack.MsgOption) error {
	_, _, msgErr := p.client.PostMessageContext(ctx, channel, msg)
	if msgErr != nil {
		return fmt.Errorf("post message: %w", msgErr)
	}
	return nil
}

func (p *ChatProvider) SendUserLinkMessage(ctx context.Context, id string, msgText string, linkUrl string, linkText string) error {
	buttonElement := slack.NewButtonBlockElement(
		"link_button_action1",
		"button_value",
		slack.NewTextBlockObject("plain_text", linkText, false, false),
	)
	buttonElement.URL = linkUrl
	buttonElement.Style = "primary"

	textElement := slack.NewTextBlockObject("mrkdwn", msgText, false, false)

	msg := slack.MsgOptionBlocks(
		slack.NewSectionBlock(textElement, nil, nil),
		slack.NewActionBlock("link_button_action_block1", buttonElement))

	return p.sendUserMessage(ctx, id, msg)
}

func (p *ChatProvider) SendOncallHandover(ctx context.Context, params rez.SendOncallHandoverParams) error {
	roster, rosterErr := params.EndingShift.QueryRoster().Only(ctx)
	if rosterErr != nil {
		return fmt.Errorf("get shift roster: %w", rosterErr)
	}

	if roster.ChatChannelID == "" {
		return fmt.Errorf("no chat channel found for roster: %s", roster.ID)
	}

	sender, senderUserErr := params.EndingShift.QueryUser().Only(ctx)
	if senderUserErr != nil {
		return fmt.Errorf("get EndingShift user: %w", senderUserErr)
	}
	if sender.ChatID == "" {
		return fmt.Errorf("no chat id for handover sender %s", sender.ID)
	}

	receiver, receiverUserErr := params.StartingShift.QueryUser().Only(ctx)
	if receiverUserErr != nil {
		return fmt.Errorf("get StartingShift user: %w", receiverUserErr)
	}
	if receiver.ChatID == "" {
		return fmt.Errorf("no chat id for handover receiver %s", receiver.ID)
	}

	builder := handoverMessageBuilder{
		roster:        roster,
		senderId:      sender.ChatID,
		receiverId:    receiver.ChatID,
		endingShift:   params.EndingShift,
		startingShift: params.StartingShift,
		incidents:     params.Incidents,
		annotations:   params.Annotations,
	}

	if buildErr := builder.build(params.Content); buildErr != nil {
		return fmt.Errorf("building handover message: %w", buildErr)
	}

	return p.sendMessage(ctx, roster.ChatChannelID, builder.getMessage())
}
