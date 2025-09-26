package slack

import (
	"context"
	"fmt"
	"net/http"
	"os"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/access"
	"github.com/rezible/rezible/ent"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
)

const (
	botTokenEnvVar         = "SLACK_BOT_TOKEN"
	appTokenEnvVar         = "SLACK_APP_TOKEN"
	enableSocketModeEnvVar = "SLACK_USE_SOCKETMODE"
	signingSecretEnvVar    = "SLACK_WEBHOOK_SIGNING_SECRET"
)

type ChatService struct {
	client               *slack.Client
	webhookSigningSecret string

	users rez.UserService
	annos rez.EventAnnotationsService
}

func LoadClient() (*slack.Client, error) {
	botToken := os.Getenv(botTokenEnvVar)
	if botToken == "" {
		return nil, fmt.Errorf("%s environment variable not set", botTokenEnvVar)
	}

	appToken := os.Getenv(appTokenEnvVar)
	if appToken != "" {
		// TODO: check if socketmode enabled
	}

	client := slack.New(botToken,
		slack.OptionAppLevelToken(appToken))

	return client, nil
}

func UseSocketMode() bool {
	return os.Getenv(enableSocketModeEnvVar) == "true"
}

func NewChatService(users rez.UserService, annos rez.EventAnnotationsService) (*ChatService, error) {
	s := &ChatService{
		users: users,
		annos: annos,
	}
	client, clientErr := LoadClient()
	if clientErr != nil {
		return nil, clientErr
	}
	s.client = client

	s.webhookSigningSecret = os.Getenv(signingSecretEnvVar)
	if s.webhookSigningSecret == "" && !UseSocketMode() {
		return nil, fmt.Errorf("%s environment variable not set", signingSecretEnvVar)
	}
	return s, nil
}

func (s *ChatService) lookupChatUser(baseCtx context.Context, chatId string) (*ent.User, context.Context, error) {
	usr, usrErr := s.users.GetByChatId(access.SystemContext(baseCtx), chatId)
	if usrErr != nil {
		log.Error().Err(usrErr).Str("chat_id", chatId).Msg("failed to lookup chat user")
		return nil, nil, usrErr
	}
	return usr, access.TenantUserContext(baseCtx, usr.TenantID), nil
}

func (s *ChatService) GetWebhooksHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/slack/options", s.handleOptionsWebhook)
	mux.HandleFunc("/slack/events", s.handleEventsWebhook)
	mux.HandleFunc("/slack/interaction", s.handleInteractionsWebhook)
	return mux
}

func (s *ChatService) sendMessage(ctx context.Context, channelId string, msgOpts ...slack.MsgOption) error {
	_, _, msgErr := s.client.PostMessageContext(ctx, channelId, msgOpts...)
	return msgErr
}

func (s *ChatService) SendMessage(ctx context.Context, channelId string, content *rez.ContentNode) error {
	return s.sendMessage(ctx, channelId, slack.MsgOptionBlocks(convertContentToBlocks(content, "")...))
}

func (s *ChatService) SendTextMessage(ctx context.Context, channelId string, text string) error {
	return s.sendMessage(ctx, channelId, slack.MsgOptionText(text, false))
}

func (s *ChatService) SendReply(ctx context.Context, channelId string, threadId string, text string) error {
	return s.sendMessage(ctx, channelId, slack.MsgOptionText(text, false), slack.MsgOptionTS(threadId))
}

func (s *ChatService) SendOncallHandover(ctx context.Context, params rez.SendOncallHandoverParams) error {
	channel, msg, err := buildHandoverMessage(params)
	if err != nil {
		return fmt.Errorf("creating handover message: %w", err)
	}
	return s.sendMessage(ctx, channel, msg)
}

func (s *ChatService) SendOncallHandoverReminder(ctx context.Context, shift *ent.OncallShift) error {
	return nil
}
