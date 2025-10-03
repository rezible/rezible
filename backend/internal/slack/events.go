package slack

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack/slackevents"
)

func (s *ChatService) handleEventsApiEvent(ctx context.Context, ev slackevents.EventsAPIEvent) (bool, error) {
	log.Debug().Str("type", string(ev.Type)).Msg("handleEventsApiEvent")
	return s.handleCallbackEvent(ctx, ev)
}

func (s *ChatService) onCallbackEventReceived(ctx context.Context, ev slackevents.EventsAPIEvent) (bool, error) {
	// TODO: queue?
	return s.handleCallbackEvent(ctx, ev)
}

func (s *ChatService) handleCallbackEvent(ctx context.Context, ev slackevents.EventsAPIEvent) (bool, error) {
	switch data := ev.InnerEvent.Data.(type) {
	case *slackevents.AppHomeOpenedEvent:
		s.onUserHomeOpenedEvent(data)
	case *slackevents.AppMentionEvent:
		s.onMentionEvent(data)
	case *slackevents.AssistantThreadStartedEvent:
		s.onAssistantThreadStartedEvent(data)
	case *slackevents.MessageEvent:
		s.onMessageEvent(data)
	default:
		log.Debug().
			Str("innerEventType", ev.InnerEvent.Type).
			Msg("unhandled slack callback event")
	}
	return true, nil
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
	usr, ctx, usrErr := s.lookupChatUser(context.Background(), data.User)
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
