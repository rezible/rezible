package slack

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

func (s *ChatService) RunSocketMode(ctx context.Context) error {
	smc := socketmode.New(
		s.client,
		//socketmode.OptionDebug(true),
	)
	smh := socketmode.NewSocketmodeHandler(smc)

	smh.Handle(socketmode.EventTypeConnectionError, s.onSocketModeConnectionFailed)
	smh.Handle(socketmode.EventTypeEventsAPI, s.handleSocketModeEventsAPI)
	smh.Handle(socketmode.EventTypeInteractive, s.handleSocketModeInteraction)
	smh.HandleDefault(s.handleSocketModeDefault)

	return smh.RunEventLoopContext(ctx)
}

func (s *ChatService) onSocketModeConnectionFailed(evt *socketmode.Event, client *socketmode.Client) {
	log.Error().Interface("evt", evt).Msg("socket mode connection failed")
}

func (s *ChatService) handleSocketModeDefault(evt *socketmode.Event, client *socketmode.Client) {
	log.Debug().Str("type", string(evt.Type)).Msg("didnt handle socket mode event")
}

func (s *ChatService) handleSocketModeEventsAPI(evt *socketmode.Event, smc *socketmode.Client) {
	eae, _ := evt.Data.(slackevents.EventsAPIEvent)
	smc.Ack(*evt.Request)
	s.handleEventsAPIEvent(eae)
}

func (s *ChatService) handleSocketModeInteraction(evt *socketmode.Event, smc *socketmode.Client) {
	ic, ok := evt.Data.(slack.InteractionCallback)
	if !ok {
		log.Error().Interface("evt", evt).Msg("failed to cast interaction callback")
		return
	}
	ctx := context.Background()
	if handlerErr := s.handleInteractionEvent(ctx, &ic); handlerErr != nil {
		log.Error().Err(handlerErr).Msg("handling socket mode interaction")
	}
	if ackErr := smc.AckCtx(ctx, evt.Request.EnvelopeID, nil); ackErr != nil {
		log.Error().Err(ackErr).Msg("acking socket mode interaction")
	}
}
