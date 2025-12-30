package slack

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

const (
	eventKindCallback = "callback"
)

func (s *ChatService) onCallbackEventReceived(data slackevents.EventsAPIEvent) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	_, handleErr := s.handleCallbackEvent(ctx, &data)
	if handleErr != nil {
		log.Error().
			Err(handleErr).
			Msg("failed to handle callback event")
	}
}

func (s *ChatService) queueCallbackEvent(ctx context.Context, data slackevents.EventsAPIEvent) error {
	return nil
}

func (s *ChatService) ProcessEvent(ctx context.Context, data slackevents.EventsAPIEvent) error {
	_, handleErr := s.handleCallbackEvent(ctx, &data)
	if handleErr != nil {
		return fmt.Errorf("failed to handle callback event: %w", handleErr)
	}
	return nil
}

func (s *ChatService) handleCallbackEvent(ctx context.Context, ev *slackevents.EventsAPIEvent) (bool, error) {
	switch data := ev.InnerEvent.Data.(type) {
	case *slackevents.AppHomeOpenedEvent:
		return true, s.onUserHomeOpenedEvent(ctx, data)
	case *slackevents.AppMentionEvent:
		return true, s.onMentionEvent(ctx, data)
	case *slackevents.AssistantThreadStartedEvent:
		return true, s.onAssistantThreadStartedEvent(ctx, data)
	case *slackevents.MessageEvent:
		return true, s.onMessageEvent(ctx, data)
	default:
		log.Debug().
			Str("innerEventType", ev.InnerEvent.Type).
			Msg("unhandled slack callback event")
		return false, nil
	}
}

func (s *ChatService) onMentionEvent(ctx context.Context, data *slackevents.AppMentionEvent) error {
	replyTs := data.TimeStamp
	if data.ThreadTimeStamp != "" {
		replyTs = data.ThreadTimeStamp
	}

	// data.Channel, replyTs, data.User, data.Text
	log.Debug().Str("replyTs", replyTs).Msg("mention event")
	return nil
}

func (s *ChatService) onMessageEvent(ctx context.Context, data *slackevents.MessageEvent) error {
	threadTs := data.ThreadTimeStamp
	// TODO check if thread is 'monitored'

	log.Debug().
		Str("type", data.ChannelType).
		Str("text", data.Text).
		Str("thread", threadTs).
		Str("user", data.User).
		Msg("message")
	return nil
}

func (s *ChatService) onAssistantThreadStartedEvent(ctx context.Context, data *slackevents.AssistantThreadStartedEvent) error {
	log.Debug().Msg("assistant thread started")
	return nil
}

func (s *ChatService) onUserHomeOpenedEvent(ctx context.Context, data *slackevents.AppHomeOpenedEvent) error {
	usr, ctx, usrErr := s.lookupChatUser(ctx, data.User)
	if usrErr != nil {
		return fmt.Errorf("failed to lookup user: %w", usrErr)
	}

	homeView, viewErr := makeUserHomeView(ctx, usr)
	if viewErr != nil || homeView == nil {
		return fmt.Errorf("failed to create user home view: %w", viewErr)
	}

	return s.withClient(ctx, func(client *slack.Client) error {
		resp, publishErr := client.PublishViewContext(ctx, data.User, *homeView, "")
		if publishErr != nil {
			logSlackViewErrorResponse(publishErr, resp)
			return fmt.Errorf("failed to publish user home view: %w", publishErr)
		}

		return nil
	})
}
