package slack

import (
	"context"
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/rezible/rezible/execution"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

var respondCallbackInnerEvents = mapset.NewSet(
	slackevents.AppHomeOpened,
	slackevents.AppMention,
	slackevents.AssistantThreadStarted,
	slackevents.Message,
)

func (s *ChatService) handleCallbackEvent(ctx context.Context, ev *slackevents.EventsAPIEvent) error {
	ctx = execution.AnonymousTenantContext(ctx, s.ci.intg.TenantID)

	switch data := ev.InnerEvent.Data.(type) {
	case *slackevents.AppHomeOpenedEvent:
		return s.onUserHomeOpenedEvent(ctx, data)
	case *slackevents.AppMentionEvent:
		return s.onMentionEvent(ctx, data)
	case *slackevents.AssistantThreadStartedEvent:
		return s.onAssistantThreadStartedEvent(ctx, data)
	case *slackevents.MessageEvent:
		return s.onMessageEvent(ctx, data)
	default:
		s.logger.Warn("unhandled slack callback event", "innerEventType", ev.InnerEvent.Type)
		return nil
	}
}

func (s *ChatService) onMentionEvent(ctx context.Context, data *slackevents.AppMentionEvent) error {
	replyTs := data.TimeStamp
	if data.ThreadTimeStamp != "" {
		replyTs = data.ThreadTimeStamp
	}

	// data.Channel, replyTs, data.User, data.Text
	s.logger.Debug("mention event", "replyTs", replyTs)
	return nil
}

func (s *ChatService) onMessageEvent(ctx context.Context, data *slackevents.MessageEvent) error {
	//slog.Debug("message event", "message", data)
	/*
		threadTs := data.ThreadTimeStamp
		// TODO check if thread is 'monitored'

		slog.Debug("message event",
			"type", data.ChannelType,
			"text", data.Text,
			"thread", threadTs,
			"user", data.User,
		)
	*/

	return nil
}

func (s *ChatService) onAssistantThreadStartedEvent(ctx context.Context, data *slackevents.AssistantThreadStartedEvent) error {
	s.logger.Debug("assistant thread started")
	return nil
}

func (s *ChatService) onUserHomeOpenedEvent(baseCtx context.Context, data *slackevents.AppHomeOpenedEvent) error {
	ctx, usrErr := s.createUserContext(baseCtx, data.User)
	if usrErr != nil {
		return fmt.Errorf("failed to lookup user: %w", usrErr)
	}

	homeView, viewErr := makeUserHomeView(ctx)
	if viewErr != nil || homeView == nil {
		return fmt.Errorf("failed to create user home view: %w", viewErr)
	}

	req := slack.PublishViewContextRequest{
		UserID: data.User,
		View:   *homeView,
		Hash:   nil,
	}
	resp, publishErr := s.client.PublishViewContext(ctx, req)
	if publishErr != nil {
		logSlackViewErrorResponse(publishErr, resp)
		return fmt.Errorf("failed to publish user home view: %w", publishErr)
	}

	return nil
}
