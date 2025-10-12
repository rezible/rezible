package slack

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

const (
	enableSocketModeEnvVar = "SLACK_USE_SOCKETMODE"
)

func UseSocketMode() bool {
	return os.Getenv(enableSocketModeEnvVar) == "true"
}

type SocketModeListener struct {
	chatSvc  *ChatService
	client   *socketmode.Client
	cancelFn context.CancelFunc
}

func NewSocketModeEventListener(chatSvc *ChatService) (*SocketModeListener, error) {
	smc := socketmode.New(
		chatSvc.client,
		//socketmode.OptionDebug(true),
	)
	sml := &SocketModeListener{
		chatSvc:  chatSvc,
		client:   smc,
		cancelFn: func() {},
	}
	return sml, nil
}

func (sml *SocketModeListener) Start(ctx context.Context) error {
	cancelCtx, cancel := context.WithCancel(ctx)
	sml.cancelFn = cancel

	log.Info().Msg("Listening for slack events in socket mode")
	go sml.runEventConsumerLoop(cancelCtx)

	runErr := sml.client.RunContext(cancelCtx)
	if runErr != nil && !errors.Is(runErr, context.Canceled) {
		return fmt.Errorf("error running socket mode listener: %w", runErr)
	}
	log.Info().Msg("Closed slack socket mode connection")
	return nil
}

func (sml *SocketModeListener) Stop(ctx context.Context) error {
	sml.cancelFn()
	return nil
}

func (sml *SocketModeListener) runEventConsumerLoop(ctx context.Context) {
	for {
		select {
		case evt, ok := <-sml.client.Events:
			if ok {
				sml.onEvent(ctx, &evt)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (sml *SocketModeListener) onEvent(ctx context.Context, evt *socketmode.Event) {
	if evt.Request == nil || evt.Type == socketmode.EventTypeHello {
		return
	}

	handled, payload, handlerErr := sml.handleEvent(ctx, evt)
	if handlerErr != nil {
		log.Error().Err(handlerErr).Msgf("Error handling socket mode event")
	} else if !handled {
		log.Warn().Str("type", string(evt.Type)).Msgf("skipping socket mode event")
	} else {
		ackErr := sml.client.AckCtx(ctx, evt.Request.EnvelopeID, payload)
		if ackErr != nil {
			log.Error().Err(ackErr).Msgf("Error acking socket mode event")
		}
	}
}

func (sml *SocketModeListener) handleEvent(ctx context.Context, evt *socketmode.Event) (bool, any, error) {
	switch evt.Type {
	case socketmode.EventTypeEventsAPI:
		eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
		if !ok {
			return false, nil, fmt.Errorf("invalid events api event data")
		}
		handled, handleErr := sml.chatSvc.handleEventsApiEvent(ctx, eventsAPIEvent)
		return handled, nil, handleErr
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
		return sml.chatSvc.handleSlashCommand(ctx, cmd)
	default:
		log.Warn().Str("type", string(evt.Type)).Msg("skipping socketmode event")
		return false, nil, nil
	}
}
