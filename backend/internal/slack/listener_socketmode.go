package slack

import (
	"context"
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
		chatSvc: chatSvc,
		client:  smc,
	}
	return sml, nil
}

func (sml *SocketModeListener) Start(baseCtx context.Context) error {
	ctx, cancel := context.WithCancel(baseCtx)
	sml.cancelFn = cancel
	
	go sml.runEventLoop(ctx)
	return sml.client.RunContext(ctx)
}

func (sml *SocketModeListener) Stop(ctx context.Context) error {
	if sml.cancelFn != nil {
		sml.cancelFn()
	}
	return nil
}

func (sml *SocketModeListener) runEventLoop(ctx context.Context) {
	for {
		select {
		case evt, ok := <-sml.client.Events:
			if ok {
				sml.onEvent(ctx, &evt)
			}
		case <-ctx.Done():
			log.Debug().Msg("closed socketmode event loop")
			return
		}
	}
}

func (sml *SocketModeListener) onEvent(ctx context.Context, evt *socketmode.Event) {
	log.Debug().Str("type", string(evt.Type)).Msg("socket event")
	handled, payload, handlerErr := sml.handleEvent(ctx, evt)
	if handlerErr != nil {
		log.Error().Err(handlerErr).Msgf("Error handling socket mode event")
	} else if !handled {
		log.Warn().Str("type", string(evt.Type)).Msgf("skipping socket mode event")
	} else {
		if evt.Request != nil {
			ackErr := sml.client.AckCtx(ctx, evt.Request.EnvelopeID, payload)
			if ackErr != nil {
				log.Error().Err(ackErr).Msgf("Error acking socket mode event")
			}
		}
	}
}

func (sml *SocketModeListener) handleEvent(ctx context.Context, evt *socketmode.Event) (bool, any, error) {
	switch evt.Type {
	case socketmode.EventTypeConnecting:
		return true, nil, nil
	case socketmode.EventTypeConnectionError:
		return true, nil, nil
	case socketmode.EventTypeConnected:
		return true, nil, nil
	case socketmode.EventTypeHello:
		return true, nil, nil
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
