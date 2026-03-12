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

type socketModeListenerConfig struct {
	AppToken string
	BotToken string
}

func getSocketModeListenerConfig() (*socketModeListenerConfig, error) {
	if !rez.Config.SingleTenantMode() {
		return nil, errors.New("can't use socket mode in multi-tenant mode")
	}

	appToken := rez.Config.GetString("slack.app_token")
	if appToken == "" {
		return nil, errors.New("slack.app_token not set")
	}

	botToken := rez.Config.GetString("slack.bot_token")
	if botToken == "" {
		return nil, errors.New("slack.bot_token not set")
	}

	return &socketModeListenerConfig{
		AppToken: appToken,
		BotToken: botToken,
	}, nil
}

func newSocketModeEventListener(handler *eventHandler) (*SocketModeListener, error) {
	cfg, cfgErr := getSocketModeListenerConfig()
	if cfgErr != nil {
		return nil, cfgErr
	}
	return &SocketModeListener{
		client:  socketmode.New(slack.New(cfg.BotToken, slack.OptionAppLevelToken(cfg.AppToken))),
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
		if err := recover(); err != nil {
			log.Error().Interface("panic", err).Msg("panic while handling socket mode event")
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

func (l *SocketModeListener) shouldIgnoreEvent(e *socketmode.Event) bool {
	return e.Request == nil || e.Type == socketmode.EventTypeHello
}

func (l *SocketModeListener) onEvent(ctx context.Context, evt *socketmode.Event) {
	if !l.shouldIgnoreEvent(evt) {
		return
	}

	var handleErr error
	switch evt.Type {
	case socketmode.EventTypeSlashCommand:
		handleErr = l.onSlashCommand(ctx, evt)
	case socketmode.EventTypeInteractive:
		handleErr = l.onInteraction(ctx, evt)
	case socketmode.EventTypeEventsAPI:
		handleErr = l.onEventsApi(ctx, evt)
	default:
		log.Warn().Str("type", string(evt.Type)).Msg("skipped socketmode event")
	}
	if handleErr != nil {
		log.Error().Str("event_type", string(evt.Type)).Err(handleErr).Msg("socketmode handler error")
		// return
	}
	if ackErr := l.client.AckCtx(ctx, evt.Request.EnvelopeID, nil); ackErr != nil {
		log.Error().Err(ackErr).Msgf("Error acking socket mode event")
	}
}

func (l *SocketModeListener) onSlashCommand(ctx context.Context, e *socketmode.Event) error {
	cmd, ok := e.Data.(slack.SlashCommand)
	if !ok {
		return fmt.Errorf("parsing SlashCommand data")
	}
	if handlerErr := l.handler.SlashCommand(ctx, cmd); handlerErr != nil {
		return fmt.Errorf("handling SlashCommand: %w", handlerErr)
	}
	return nil
}

func (l *SocketModeListener) onInteraction(ctx context.Context, e *socketmode.Event) error {
	ic, ok := e.Data.(slack.InteractionCallback)
	if !ok {
		return fmt.Errorf("parsing InteractionCallback data")
	}
	if handlerErr := l.handler.InteractionCallback(ctx, &ic); handlerErr != nil {
		return fmt.Errorf("handling InteractionCallback: %w", handlerErr)
	}
	return nil
}

func (l *SocketModeListener) onEventsApi(ctx context.Context, e *socketmode.Event) error {
	evt, ok := e.Data.(slackevents.EventsAPIEvent)
	if !ok {
		return fmt.Errorf("invalid events api event data")
	}

	if evt.Type == slackevents.CallbackEvent {
		if handlerErr := l.handler.CallbackEvent(ctx, &evt); handlerErr != nil {
			return fmt.Errorf("handling EventsAPIEvent: %w", handlerErr)
		}
	} else if evt.Type == slackevents.AppRateLimited {
		log.Warn().Msg("slack app rate limited")
	} else {
		log.Warn().Str("type", evt.Type).Msg("unknown slack callback event type")
	}
	return nil
}
