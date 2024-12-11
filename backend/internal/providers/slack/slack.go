package slack

import (
	"context"
	"fmt"
	"github.com/slack-go/slack"
	rez "github.com/twohundreds/rezible"
	"github.com/twohundreds/rezible/ent"
	"net/http"
)

type ChatProvider struct {
	client        *slack.Client
	signingSecret string

	lookupUserFn func(context.Context, string) (*ent.User, error)
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
		lookupUserFn: func(ctx context.Context, s string) (*ent.User, error) {
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
	p.lookupUserFn = lookupFn
}

func (p *ChatProvider) lookupUser(ctx context.Context, id string) (*ent.User, error) {
	return p.lookupUserFn(ctx, id)
}

func (p *ChatProvider) SendUserMessage(ctx context.Context, user *ent.User, msg string) error {
	return p.sendUserMessage(ctx, user, slack.MsgOptionText(msg, false))
}

func (p *ChatProvider) sendUserMessage(ctx context.Context, user *ent.User, msg slack.MsgOption) error {
	slackUser, userErr := p.client.GetUserByEmailContext(ctx, user.Email)
	if userErr != nil {
		return fmt.Errorf("failed to find user by email: %w", userErr)
	}

	convo, _, _, convoErr := p.client.OpenConversationContext(ctx, &slack.OpenConversationParameters{
		Users: []string{slackUser.ID},
	})
	if convoErr != nil {
		return fmt.Errorf("failed to open conversation with user %s: %w", slackUser.ID, convoErr)
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

func (p *ChatProvider) SendUserLinkMessage(ctx context.Context, user *ent.User, msgText string, linkUrl string, linkText string) error {
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
	return p.sendUserMessage(ctx, user, msg)
}

func (p *ChatProvider) SendOncallHandover(ctx context.Context, params rez.SendOncallHandoverParams) error {
	roster, rosterErr := params.EndingShift.QueryRoster().Only(ctx)
	if rosterErr != nil {
		return fmt.Errorf("get shift roster: %w", rosterErr)
	}

	if roster.ChatChannelID == "" {
		return fmt.Errorf("no chat channel found for roster: %s", roster.ID)
	}

	senderUser, senderUserErr := params.EndingShift.QueryUser().Only(ctx)
	if senderUserErr != nil {
		return fmt.Errorf("get EndingShift user: %w", senderUserErr)
	}
	sender, senderErr := p.client.GetUserByEmailContext(ctx, senderUser.Email)
	if senderErr != nil {
		return fmt.Errorf("get slack sender: %w", senderErr)
	}

	receiverUser, receiverUserErr := params.StartingShift.QueryUser().Only(ctx)
	if receiverUserErr != nil {
		return fmt.Errorf("get StartingShift user: %w", receiverUserErr)
	}
	receiver, receiverErr := p.client.GetUserByEmailContext(ctx, receiverUser.Email)
	if receiverErr != nil {
		return fmt.Errorf("get slack receiver: %w", receiverErr)
	}

	builder := handoverMessageBuilder{
		roster:        roster,
		sender:        sender,
		receiver:      receiver,
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
