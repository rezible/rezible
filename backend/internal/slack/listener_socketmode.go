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
	chatSvc   *ChatService
	client    *socketmode.Client
	eventPool *pool.ContextPool
	cancelFn  context.CancelFunc
}

func NewSocketModeEventListener(chatSvc *ChatService) (*SocketModeListener, error) {
	if !rez.Config.SingleTenantMode() {
		return nil, errors.New("can't use socket mode in multi-tenant mode")
	}
	stc, cErr := LoadSingleTenantClient()
	if cErr != nil {
		return nil, fmt.Errorf("single tenant client: %w", cErr)
	}

	smc := socketmode.New(stc)
	sml := &SocketModeListener{
		chatSvc:  chatSvc,
		client:   smc,
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
		eev, ok := evt.Data.(slackevents.EventsAPIEvent)
		if !ok {
			return false, nil, fmt.Errorf("invalid events api event data")
		}
		return sml.handleEventsApiEvent(ctx, &eev)
	case socketmode.EventTypeInteractive:
		ic, ok := evt.Data.(slack.InteractionCallback)
		if !ok {
			return false, nil, fmt.Errorf("invalid interaction callback data")
		}
		return sml.chatSvc.handleInteractionEvent(ctx, &ic)
	case socketmode.EventTypeSlashCommand:
		cmd, ok := evt.Data.(slack.SlashCommand)
		if !ok {
			return false, nil, fmt.Errorf("invalid slash command data")
		}
		userCtx, userErr := sml.chatSvc.getChatUserContext(ctx, cmd.UserID)
		if userErr != nil {
			return false, nil, fmt.Errorf("failed to lookup user: %w", userErr)
		}
		return sml.chatSvc.handleSlashCommand(userCtx, &cmd)
	default:
		log.Warn().Str("type", string(evt.Type)).Msg("skipped socketmode event")
		return false, nil, nil
	}
}

func (sml *SocketModeListener) handleEventsApiEvent(ctx context.Context, evt *slackevents.EventsAPIEvent) (bool, any, error) {
	if evt.Type == slackevents.CallbackEvent {
		handled, handleErr := sml.chatSvc.handleCallbackEvent(ctx, evt)
		if handleErr != nil {
			return true, nil, fmt.Errorf("handling callback event: %w", handleErr)
		}
		return handled, nil, nil
	}
	log.Warn().Str("type", evt.Type).Msg("didnt handle slack callback event")
	return false, nil, nil
}
