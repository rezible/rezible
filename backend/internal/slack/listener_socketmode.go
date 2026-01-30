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
			if ok {
				sml.onEventReceived(ctx, &evt)
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (sml *SocketModeListener) onEventReceived(ctx context.Context, evt *socketmode.Event) {
	if evt.Request == nil || evt.Type == socketmode.EventTypeHello {
		return
	}

	handled, payload, handlerErr := sml.handleEvent(ctx, evt)
	if handlerErr != nil {
		log.Error().Err(handlerErr).Msgf("Error handling socket mode event")
	}

	if handled {
		ackErr := sml.client.AckCtx(ctx, evt.Request.EnvelopeID, payload)
		if ackErr != nil {
			log.Error().Err(ackErr).Msgf("Error acking socket mode event")
		}
	} else {
		log.Warn().Str("type", string(evt.Type)).Msgf("unhandled socket mode event")
	}
}

func (sml *SocketModeListener) handleEvent(ctx context.Context, evt *socketmode.Event) (bool, any, error) {
	switch evt.Type {
	case socketmode.EventTypeEventsAPI:
		return sml.handleEventsApiEvent(ctx, evt)
	case socketmode.EventTypeInteractive:
		return sml.handleInteractiveEvent(ctx, evt)
	case socketmode.EventTypeSlashCommand:
		return sml.handleSlashCommand(ctx, evt)
	default:
		log.Warn().Str("type", string(evt.Type)).Msg("skipped socketmode event")
		return false, nil, nil
	}
}

func (sml *SocketModeListener) handleEventsApiEvent(ctx context.Context, e *socketmode.Event) (bool, any, error) {
	evt, ok := e.Data.(slackevents.EventsAPIEvent)
	if !ok {
		return false, nil, fmt.Errorf("invalid events api event data")
	}
	if evt.Type == slackevents.CallbackEvent {
		handled, handleErr := sml.chat.handleCallbackEvent(ctx, &evt)
		if handleErr != nil {
			return true, nil, fmt.Errorf("handling callback event: %w", handleErr)
		}
		return handled, nil, nil
	}
	log.Warn().Str("type", evt.Type).Msg("didnt handle slack callback event")
	return false, nil, nil
}

func (sml *SocketModeListener) handleInteractiveEvent(ctx context.Context, e *socketmode.Event) (bool, any, error) {
	ic, ok := e.Data.(slack.InteractionCallback)
	if !ok {
		return false, nil, fmt.Errorf("invalid interaction callback data")
	}
	return sml.chat.handleInteractionEvent(ctx, &ic)
}

func (sml *SocketModeListener) handleSlashCommand(ctx context.Context, e *socketmode.Event) (bool, any, error) {
	cmd, ok := e.Data.(slack.SlashCommand)
	if !ok {
		return false, nil, fmt.Errorf("invalid slash command data")
	}
	return sml.chat.handleSlashCommand(ctx, &cmd)
}
