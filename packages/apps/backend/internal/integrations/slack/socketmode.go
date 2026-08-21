package slackintegration

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"

	rez "github.com/rezible/rezible"
)

type socketModeListener struct {
	client       *socketmode.Client
	eventHandler *appEventHandler
}

func makeSocketModeListener(client *slack.Client, evth *appEventHandler) *socketModeListener {
	return &socketModeListener{
		eventHandler: evth,
		client:       socketmode.New(client),
	}
}

func (l *socketModeListener) Lifecycle() *rez.ServiceLifecycle {
	slog.Info("Listening for slack events in socket mode")
	return &rez.ServiceLifecycle{
		StartFns: []rez.LifecycleFunc{l.client.RunContext, l.runEventConsumerLoop},
	}
}

func (l *socketModeListener) runEventConsumerLoop(ctx context.Context) (runErr error) {
	defer func() {
		if panicErr := recover(); panicErr != nil {
			runErr = fmt.Errorf("panic handling Slack socket-mode event: %v\n%s", panicErr, debug.Stack())
		}
	}()
	for {
		select {
		case evt, ok := <-l.client.Events:
			if !ok {
				return errors.New("socket-mode events channel closed")
			}
			l.onEvent(ctx, &evt)
		case <-ctx.Done():
			return nil
		}
	}
}

func (l *socketModeListener) onEvent(ctx context.Context, evt *socketmode.Event) {
	if evt.Request == nil || evt.Type == socketmode.EventTypeHello {
		slog.Debug("ignoring socketmode event", "type", string(evt.Type))
		return
	}

	var handleErr error
	if evt.Type == socketmode.EventTypeInteractive {
		handleErr = l.eventHandler.OnInteractionCallback(ctx, evt.Request.Payload)
	} else if evt.Type == socketmode.EventTypeSlashCommand {
		handleErr = l.onSlashCommand(ctx, evt)
	} else if evt.Type == socketmode.EventTypeEventsAPI {
		handleErr = l.onEventsApi(ctx, evt)
	} else {
		handleErr = fmt.Errorf("unknown event type")
	}
	if handleErr != nil {
		slog.Error("socketmode handler error",
			"error", handleErr,
			"event_type", string(evt.Type),
		)
	}
	if ackErr := l.client.AckCtx(ctx, evt.Request.EnvelopeID, nil); ackErr != nil && ctx.Err() == nil {
		slog.Error("Error acking socket mode event", "error", ackErr)
	}
}

func (l *socketModeListener) onSlashCommand(ctx context.Context, e *socketmode.Event) error {
	if cmd, ok := e.Data.(slack.SlashCommand); ok {
		return l.eventHandler.OnSlashCommand(ctx, cmd)
	}
	return fmt.Errorf("invalid SlashCommand data")
}

func (l *socketModeListener) onEventsApi(ctx context.Context, e *socketmode.Event) error {
	if evt, ok := e.Data.(slackevents.EventsAPIEvent); ok {
		if evt.Type == slackevents.CallbackEvent {
			cb, cbOk := evt.Data.(*slackevents.EventsAPICallbackEvent)
			if !cbOk {
				return fmt.Errorf("failed to cast callback event")
			}
			return l.eventHandler.OnEventsApiCallback(ctx, cb, e.Request.Payload)
		} else if evt.Type == slackevents.AppRateLimited {
			return l.eventHandler.OnAppRateLimitedEvent(ctx)
		}
		return fmt.Errorf("unknown slack callback event type: %s", evt.Type)
	}
	return fmt.Errorf("invalid events api event data")
}
