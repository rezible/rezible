package slack

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	rez "github.com/rezible/rezible"
	"github.com/sourcegraph/conc/pool"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

type SocketModeListener struct {
	client  *socketmode.Client
	handler *messageHandler
	stopFn  func() error
}

func makeSocketModeEventListener(cl rez.ConfigLoader, mh *messageHandler) (*SocketModeListener, error) {
	var cfg Config
	if cfgErr := cl.Unmarshal("slack", &cfg); cfgErr != nil {
		return nil, fmt.Errorf("config error: %w", cfgErr)
	}
	l := &SocketModeListener{
		handler: mh,
		stopFn:  func() error { return nil },
	}
	if cl.SingleTenantMode() {

		l.client = socketmode.New(slack.New(cfg.BotToken, slack.OptionAppLevelToken(cfg.AppToken)))
	}
	return l, nil
}

func (l *SocketModeListener) Start(baseCtx context.Context) error {
	if l.client == nil {
		fmt.Printf("no client for socketmode\n")
		return nil
	}

	cancelCtx, cancel := context.WithCancel(baseCtx)

	p := pool.New().
		WithErrors().
		WithContext(cancelCtx)

	l.stopFn = func() error {
		cancel()
		if p != nil {
			if poolErr := p.Wait(); poolErr != nil && !errors.Is(poolErr, context.Canceled) {
				return fmt.Errorf("slack socket mode handler: %w", poolErr)
			}
		}
		return nil
	}

	p.Go(l.client.RunContext)
	p.Go(l.runEventConsumerLoop)

	slog.Info("Listening for slack events in socket mode")

	return nil
}

func (l *SocketModeListener) Shutdown(ctx context.Context) error {
	slog.Info("Stopping Slack socket mode listener")
	return l.stopFn()
}

func (l *SocketModeListener) runEventConsumerLoop(ctx context.Context) error {
	defer func() {
		if panicErr := recover(); panicErr != nil {
			slog.Error("panic while handling socket mode event", "panic", panicErr)
		}
	}()
	for {
		select {
		case evt, ok := <-l.client.Events:
			if ok {
				l.onEvent(ctx, &evt)
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (l *SocketModeListener) onEvent(ctx context.Context, evt *socketmode.Event) {
	if evt.Request == nil || evt.Type == socketmode.EventTypeHello {
		slog.Debug("ignoring socketmode event", "type", string(evt.Type))
		return
	}

	var handleErr error
	if evt.Type == socketmode.EventTypeInteractive {
		handleErr = l.handler.OnInteractionCallback(ctx, evt.Request.Payload)
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
	if ackErr := l.client.AckCtx(ctx, evt.Request.EnvelopeID, nil); ackErr != nil {
		slog.Error("Error acking socket mode event", "error", ackErr)
	}
}

func (l *SocketModeListener) onSlashCommand(ctx context.Context, e *socketmode.Event) error {
	if cmd, ok := e.Data.(slack.SlashCommand); ok {
		return l.handler.OnSlashCommand(ctx, cmd)
	}
	return fmt.Errorf("invalid SlashCommand data")
}

func (l *SocketModeListener) onEventsApi(ctx context.Context, e *socketmode.Event) error {
	if evt, ok := e.Data.(slackevents.EventsAPIEvent); ok {
		if evt.Type == slackevents.CallbackEvent {
			cb, cbOk := evt.Data.(*slackevents.EventsAPICallbackEvent)
			if !cbOk {
				return fmt.Errorf("failed to cast callback event")
			}
			return l.handler.OnCallbackEvent(ctx, cb, e.Request.Payload)
		} else if evt.Type == slackevents.AppRateLimited {
			return l.handler.OnAppRateLimitedEvent(ctx)
		}
		return fmt.Errorf("unknown slack callback event type: %s", evt.Type)
	}
	return fmt.Errorf("invalid events api event data")
}
