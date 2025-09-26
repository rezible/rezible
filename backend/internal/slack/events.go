package slack

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack/slackevents"
)

func (s *ChatService) handleEventsAPIEvent(ev slackevents.EventsAPIEvent) {
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
