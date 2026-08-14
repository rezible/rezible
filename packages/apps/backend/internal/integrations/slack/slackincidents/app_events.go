package slackincidents

import (
	"context"
	"log/slog"

	"github.com/rezible/rezible/ent"
	slackintegration "github.com/rezible/rezible/internal/integrations/slack"
	"github.com/slack-go/slack/slackevents"
)

func (a *App) RespondEventTypes() []slackevents.EventsAPIType {
	return []slackevents.EventsAPIType{
		slackevents.AppMention,
		slackevents.Message,
	}
}

func (a *App) EventsApiHandler() slackintegration.EventsApiHandler {
	return func(ctx context.Context, ii *ent.Integration, ev *slackevents.EventsAPIEvent) error {
		//cw, cwErr := slackintegration.NewClientWrapper(ii)
		//if cwErr != nil {
		//	return fmt.Errorf("failed to create client wrapper: %w", cwErr)
		//}

		switch data := ev.InnerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			return a.onMentionEvent(ctx, data)
		case *slackevents.MessageEvent:
			return a.onMessageEvent(ctx, data)
		default:
			slog.Warn("unhandled slack callback event", "innerEventType", ev.InnerEvent.Type)
			return nil
		}
	}
}

func (a *App) onMentionEvent(ctx context.Context, data *slackevents.AppMentionEvent) error {
	replyTs := data.TimeStamp
	if data.ThreadTimeStamp != "" {
		replyTs = data.ThreadTimeStamp
	}

	// data.Channel, replyTs, data.User, data.Text
	slog.Debug("mention event", "replyTs", replyTs)
	return nil
}

func (a *App) onMessageEvent(ctx context.Context, data *slackevents.MessageEvent) error {
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
