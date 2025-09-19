package slack

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/rs/zerolog/log"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

const (
	apiKeyEnvVar        = "SLACK_BOT_API_KEY"
	signingSecretEnvVar = "SLACK_WEBHOOK_SIGNING_SECRET"
)

type ChatService struct {
	client               *slack.Client
	webhookSigningSecret string

	users rez.UserService
	annos rez.EventAnnotationsService
}

func LoadClient() (*slack.Client, error) {
	apiKey := os.Getenv(apiKeyEnvVar)
	if apiKey == "" {
		return nil, fmt.Errorf("%s environment variable not set", apiKeyEnvVar)
	}
	return slack.New(apiKey), nil
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
	if s.webhookSigningSecret == "" {
		return nil, fmt.Errorf("%s environment variable not set", signingSecretEnvVar)
	}
	return s, nil
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

func (s *ChatService) onMentionEvent(data *slackevents.AppMentionEvent) {
	replyTs := data.TimeStamp
	if data.ThreadTimeStamp != "" {
		replyTs = data.ThreadTimeStamp
	}

	// data.Channel, replyTs, data.User, data.Text
	log.Debug().Str("replyTs", replyTs).Msg("mention event")
}

func (s *ChatService) onMessageEvent(data *slackevents.MessageEvent) {
	threadTs := data.ThreadTimeStamp
	// TODO check if thread is 'monitored'

	log.Debug().
		Str("type", data.ChannelType).
		Str("text", data.Text).
		Str("thread", threadTs).
		Str("user", data.User).
		Msg("message")
}

func (s *ChatService) onAssistantThreadStartedEvent(data *slackevents.AssistantThreadStartedEvent) {
	log.Debug().Msg("assistant thread started")
}

func (s *ChatService) onUserHomeOpenedEvent(data *slackevents.AppHomeOpenedEvent) {
	ctx := context.Background()
	usr, usrErr := s.users.GetByChatId(ctx, data.User)
	if usrErr != nil {
		log.Warn().Err(usrErr).Msg("failed to lookup user")
		return
	}
	homeView, viewErr := makeUserHomeView(ctx, usr)
	if viewErr != nil || homeView == nil {
		log.Error().Err(viewErr).Msg("failed to create user home view")
		return
	}
	resp, publishErr := s.client.PublishViewContext(ctx, data.User, *homeView, "")
	if publishErr != nil {
		logSlackViewErrorResponse(publishErr, resp)
	}
}

func (s *ChatService) handleAnnotationModalInteraction(ctx context.Context, ic *slack.InteractionCallback) error {
	view, viewErr := s.makeAnnotationModalView(ctx, ic)
	if viewErr != nil || view == nil {
		return fmt.Errorf("failed to create annotation view: %w", viewErr)
	}

	var resp *slack.ViewResponse
	var respErr error
	if ic.View.State == nil {
		resp, respErr = s.client.OpenViewContext(ctx, ic.TriggerID, *view)
	} else {
		resp, respErr = s.client.UpdateViewContext(ctx, *view, "", ic.Hash, ic.View.ID)
	}
	if respErr != nil {
		logSlackViewErrorResponse(respErr, resp)
		return fmt.Errorf("annotation modal view: %w", respErr)
	}
	return nil
}

func (s *ChatService) handleAnnotationModalSubmission(ctx context.Context, ic *slack.InteractionCallback) error {
	anno, annoErr := s.getAnnotationModalAnnotation(ctx, ic.View)
	if annoErr != nil {
		return fmt.Errorf("failed to get view annotation: %w", annoErr)
	}
	_, createErr := s.annos.SetAnnotation(ctx, anno)
	return createErr
}
