package slackagent

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/rezible/rezible/ent"
	slackintegration "github.com/rezible/rezible/internal/integrations/slack"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

func (a *app) handleEventsApiEvent(ctx context.Context, ii *ent.Integration, ev *slackevents.EventsAPIEvent) error {
	cw, cwErr := slackintegration.NewClientWrapper(ii)
	if cwErr != nil {
		return fmt.Errorf("failed to create client wrapper: %w", cwErr)
	}

	switch data := ev.InnerEvent.Data.(type) {
	case *slackevents.AppHomeOpenedEvent:
		return a.onUserHomeOpenedEvent(ctx, cw, data)
	case *slackevents.AppMentionEvent:
		return a.onMentionEvent(ctx, data)
	case *slackevents.AssistantThreadStartedEvent:
		return a.onAssistantThreadStartedEvent(ctx, data)
	case *slackevents.MessageEvent:
		return a.onMessageEvent(ctx, data)
	default:
		slog.Warn("unhandled slack callback event", "innerEventType", ev.InnerEvent.Type)
		return nil
	}
}

func (a *app) onMentionEvent(ctx context.Context, data *slackevents.AppMentionEvent) error {
	replyTs := data.TimeStamp
	if data.ThreadTimeStamp != "" {
		replyTs = data.ThreadTimeStamp
	}

	// data.Channel, replyTs, data.User, data.Text
	slog.Debug("mention event", "replyTs", replyTs)
	return nil
}

func (a *app) onMessageEvent(ctx context.Context, data *slackevents.MessageEvent) error {
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

func (a *app) onAssistantThreadStartedEvent(ctx context.Context, data *slackevents.AssistantThreadStartedEvent) error {
	slog.Debug("assistant thread started")
	return nil
}

func (a *app) onUserHomeOpenedEvent(ctx context.Context, cw *slackintegration.ClientWrapper, data *slackevents.AppHomeOpenedEvent) error {
	homeView, viewErr := makeUserHomeView(ctx)
	if viewErr != nil || homeView == nil {
		return fmt.Errorf("failed to create user home view: %w", viewErr)
	}

	req := slack.PublishViewContextRequest{
		UserID: data.User,
		View:   *homeView,
		Hash:   nil,
	}
	resp, publishErr := cw.Client().PublishViewContext(ctx, req)
	if publishErr != nil {
		slackintegration.LogSlackViewErrorResponse(slog.Default(), publishErr, resp)
		return fmt.Errorf("failed to publish user home view: %w", publishErr)
	}

	return nil
}
