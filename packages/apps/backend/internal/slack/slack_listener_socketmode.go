package slack

import (
	"context"
	"errors"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"github.com/sourcegraph/conc/pool"
)

type SocketModeListener struct {
	client  *socketmode.Client
	handler *eventHandler
	stopFn  func() error
}

func (i *integration) newSocketModeEventListener(handler *eventHandler) (*SocketModeListener, error) {
	if !rez.Config.SingleTenantMode() {
		return nil, errors.New("can't use socket mode in multi-tenant mode")
	}
	return &SocketModeListener{
		client:  socketmode.New(slack.New(i.cfg.BotToken, slack.OptionAppLevelToken(i.cfg.AppToken))),
		handler: handler,
		stopFn:  func() error { return nil },
	}, nil
}

func (l *SocketModeListener) Start(baseCtx context.Context) error {
	cancelCtx, cancel := context.WithCancel(baseCtx)

	p := pool.New().WithErrors().WithContext(cancelCtx)

	l.stopFn = func() error {
		cancel()
		if p == nil {
			return nil
		}
		if poolErr := p.Wait(); poolErr != nil && !errors.Is(poolErr, context.Canceled) {
			return fmt.Errorf("slack socket mode handler: %w", poolErr)
		}
		return nil
	}

	p.Go(l.client.RunContext)
	p.Go(l.runEventConsumerLoop)

	log.Info().Msg("Listening for slack events in socket mode")

	return nil
}

func (l *SocketModeListener) Stop(ctx context.Context) error {
	log.Info().Msg("Stopping Slack socket mode listener")
	return l.stopFn()
}

func (l *SocketModeListener) runEventConsumerLoop(ctx context.Context) error {
	defer func() {
		if panicErr := recover(); panicErr != nil {
			log.Error().
				Interface("panic", panicErr).
				Msg("panic while handling socket mode event")
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
		log.Debug().Str("type", string(evt.Type)).Msg("ignoring socketmode event")
		return
	}

	handleErr := fmt.Errorf("unknown event type")
	if evt.Type == socketmode.EventTypeEventsAPI {
		handleErr = l.onEventsApi(ctx, evt)
	} else if evt.Type == socketmode.EventTypeSlashCommand {
		handleErr = l.onSlashCommand(ctx, evt)
	} else if evt.Type == socketmode.EventTypeInteractive {
		handleErr = l.onInteractionCallback(ctx, evt)
	}
	if handleErr != nil {
		log.Error().Err(handleErr).
			Str("event_type", string(evt.Type)).
			Msg("socketmode handler error")
	}
	if ackErr := l.client.AckCtx(ctx, evt.Request.EnvelopeID, nil); ackErr != nil {
		log.Error().Err(ackErr).Msgf("Error acking socket mode event")
	}
}

func (l *SocketModeListener) onEventsApi(ctx context.Context, e *socketmode.Event) error {
	if evt, ok := e.Data.(slackevents.EventsAPIEvent); ok {
		if evt.Type == slackevents.CallbackEvent {
			return l.handler.OnCallbackEvent(ctx, e.Request.Payload)
		} else if evt.Type == slackevents.AppRateLimited {
			return l.handler.OnAppRateLimitedEvent(ctx)
		}
		return fmt.Errorf("unknown slack callback event type: %s", evt.Type)
	}
	return fmt.Errorf("invalid events api event data")
}

func (l *SocketModeListener) onSlashCommand(ctx context.Context, e *socketmode.Event) error {
	if cmd, ok := e.Data.(slack.SlashCommand); ok {
		return l.handler.OnSlashCommand(ctx, cmd)
	}
	return fmt.Errorf("invalid SlashCommand data")
}

func (l *SocketModeListener) onInteractionCallback(ctx context.Context, e *socketmode.Event) error {
	return l.handler.OnInteractionCallback(ctx, e.Request.Payload)
}
