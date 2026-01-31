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

func UseSocketMode() bool {
	return rez.Config.GetBool("slack.socketmode.enabled")
}

type SocketModeListener struct {
	client    *socketmode.Client
	chat      *ChatService
	eventPool *pool.ContextPool
	cancelFn  context.CancelFunc
}

func newSocketModeEventListener(svcs *rez.Services) (*SocketModeListener, error) {
	if !rez.Config.SingleTenantMode() {
		return nil, errors.New("can't use socket mode in multi-tenant mode")
	}

	appToken := rez.Config.GetString("slack.app_token")
	if appToken != "" {
		return nil, errors.New("slack.app_token not set")
	}

	botToken := rez.Config.GetString("slack.bot_token")
	if botToken == "" {
		return nil, errors.New("slack.bot_token not set")
	}

	client := slack.New(botToken, slack.OptionAppLevelToken(appToken))

	sml := &SocketModeListener{
		chat:     newChatService(client, svcs),
		client:   socketmode.New(client),
		cancelFn: func() {},
	}
	return sml, nil
}

func (sml *SocketModeListener) Start(ctx context.Context) error {
	cancelCtx, cancel := context.WithCancel(ctx)
	p := pool.New().
		WithErrors().
		WithContext(cancelCtx)

	log.Info().Msg("Listening for slack events in socket mode")

	p.Go(sml.runEventConsumerLoop)
	p.Go(sml.client.RunContext)

	sml.eventPool = p
	sml.cancelFn = cancel

	return nil
}

func (sml *SocketModeListener) Stop(ctx context.Context) error {
	sml.cancelFn()
	if poolErr := sml.eventPool.Wait(); poolErr != nil && !errors.Is(poolErr, context.Canceled) {
		return fmt.Errorf("slack socket mode handler: %w", poolErr)
	}
	log.Info().Msg("Stopped Slack socket mode listener")
	return nil
}

func (sml *SocketModeListener) runEventConsumerLoop(ctx context.Context) error {
	defer func() {
		if err := recover(); err != nil {
			log.Error().Interface("panic", err).Msg("panic while handling socket mode event")
		}
	}()
	for {
		select {
		case evt, ok := <-sml.client.Events:
			if ok && !sml.shouldIgnoreEvent(&evt) {
				sml.onEvent(ctx, &evt)
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (sml *SocketModeListener) shouldIgnoreEvent(e *socketmode.Event) bool {
	return e.Request == nil || e.Type == socketmode.EventTypeHello
}

func (sml *SocketModeListener) onEvent(ctx context.Context, evt *socketmode.Event) {
	var payload any
	var handled bool
	var handlerErr error
	switch evt.Type {
	case socketmode.EventTypeSlashCommand:
		handled, payload, handlerErr = sml.onSlashCommand(ctx, evt)
	case socketmode.EventTypeInteractive:
		handled, payload, handlerErr = sml.onInteraction(ctx, evt)
	case socketmode.EventTypeEventsAPI:
		handled, payload, handlerErr = sml.onEventsApi(ctx, evt)
	default:
		log.Warn().Str("type", string(evt.Type)).Msg("skipped socketmode event")
	}
	if handlerErr != nil {
		log.Error().Err(handlerErr).Msgf("Error handling socket mode event")
	} else if !handled {
		log.Warn().Str("type", string(evt.Type)).Msgf("unhandled socket mode event")
	}
	if ackErr := sml.client.AckCtx(ctx, evt.Request.EnvelopeID, payload); ackErr != nil {
		log.Error().Err(ackErr).Msgf("Error acking socket mode event")
	}
}

func (sml *SocketModeListener) onSlashCommand(ctx context.Context, e *socketmode.Event) (bool, any, error) {
	if cmd, ok := e.Data.(slack.SlashCommand); ok {
		return sml.chat.handleSlashCommand(ctx, &cmd)
	}
	return false, nil, fmt.Errorf("invalid slash command data")
}

func (sml *SocketModeListener) onInteraction(ctx context.Context, e *socketmode.Event) (bool, any, error) {
	if ic, ok := e.Data.(slack.InteractionCallback); ok {
		return sml.chat.handleInteractionEvent(ctx, &ic)
	}
	return false, nil, fmt.Errorf("invalid interaction callback data")
}

func (sml *SocketModeListener) onEventsApi(ctx context.Context, e *socketmode.Event) (bool, any, error) {
	evt, ok := e.Data.(slackevents.EventsAPIEvent)
	if !ok {
		return false, nil, fmt.Errorf("invalid events api event data")
	}
	handled := false
	var handlerErr error
	if evt.Type == slackevents.CallbackEvent {
		handled, handlerErr = sml.chat.handleCallbackEvent(ctx, &evt)
	} else if evt.Type == slackevents.AppRateLimited {
		handled = true
		log.Warn().Msg("slack app rate limited")
	} else {
		log.Warn().Str("type", evt.Type).Msg("unknown slack callback event type")
	}
	return handled, nil, handlerErr
}
