package slack

import (
	"context"
	"fmt"

	"github.com/rezible/rezible/jobs"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack/slackevents"
)

const (
	eventKindCallback = "callback"
)

func (s *ChatService) queueCallbackEvent(ctx context.Context, data slackevents.EventsAPIEvent) error {
	return s.jobs.Insert(ctx, jobs.InsertJobParams{
		Args: jobs.ProcessChatEvent{
			Provider:  "slack",
			EventKind: eventKindCallback,
			Data:      data,
		},
	})
}

func (s *ChatService) ProcessEvent(ctx context.Context, args jobs.ProcessChatEvent) error {
	if args.Provider != "slack" {
		return fmt.Errorf("invalid provider")
	}
	if args.EventKind == eventKindCallback {
		data, ok := args.Data.(slackevents.EventsAPIEvent)
		if !ok {
			return fmt.Errorf("invalid event")
		}
		_, handleErr := s.handleCallbackEvent(ctx, &data)
		if handleErr != nil {
			return fmt.Errorf("failed to handle callback event: %w", handleErr)
		}
		return nil
	}
	return nil
}

func (s *ChatService) handleCallbackEvent(ctx context.Context, ev *slackevents.EventsAPIEvent) (bool, error) {
	switch data := ev.InnerEvent.Data.(type) {
	case *slackevents.AppHomeOpenedEvent:
		return true, s.onUserHomeOpenedEvent(data)
	case *slackevents.AppMentionEvent:
		return true, s.onMentionEvent(data)
	case *slackevents.AssistantThreadStartedEvent:
		return true, s.onAssistantThreadStartedEvent(data)
	case *slackevents.MessageEvent:
		return true, s.onMessageEvent(data)
	default:
		log.Debug().
			Str("innerEventType", ev.InnerEvent.Type).
			Msg("unhandled slack callback event")
		return false, nil
	}
}

func (s *ChatService) onMentionEvent(data *slackevents.AppMentionEvent) error {
	replyTs := data.TimeStamp
	if data.ThreadTimeStamp != "" {
		replyTs = data.ThreadTimeStamp
	}

	// data.Channel, replyTs, data.User, data.Text
	log.Debug().Str("replyTs", replyTs).Msg("mention event")
	return nil
}

func (s *ChatService) onMessageEvent(data *slackevents.MessageEvent) error {
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

func (s *ChatService) onAssistantThreadStartedEvent(data *slackevents.AssistantThreadStartedEvent) error {
	log.Debug().Msg("assistant thread started")
	return nil
}

func (s *ChatService) onUserHomeOpenedEvent(data *slackevents.AppHomeOpenedEvent) error {
	usr, ctx, usrErr := s.lookupChatUser(context.Background(), data.User)
	if usrErr != nil {
		return fmt.Errorf("failed to lookup user: %w", usrErr)
	}

	homeView, viewErr := makeUserHomeView(ctx, usr)
	if viewErr != nil || homeView == nil {
		return fmt.Errorf("failed to create user home view: %w", viewErr)
	}

	resp, publishErr := s.client.PublishViewContext(ctx, data.User, *homeView, "")
	if publishErr != nil {
		logSlackViewErrorResponse(publishErr, resp)
		return fmt.Errorf("failed to publish user home view: %w", publishErr)
	}

	return nil
}
