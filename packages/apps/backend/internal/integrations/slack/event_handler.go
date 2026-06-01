package slack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"

	rez "github.com/rezible/rezible"
)

type EventHandler struct {
	messages              rez.MessageService
	svc                   *Service
	processCallbackEvents mapset.Set[slackevents.EventsAPIType]
	respondCallbackEvents mapset.Set[slackevents.EventsAPIType]
}

func MakeEventHandler(svc *Service, msgs rez.MessageService) (*EventHandler, error) {
	h := &EventHandler{
		svc:                   svc,
		messages:              msgs,
		processCallbackEvents: mapset.NewSet[slackevents.EventsAPIType](),
		respondCallbackEvents: mapset.NewSet[slackevents.EventsAPIType](),
	}
	if msgsErr := h.registerMessageHandlers(); msgsErr != nil {
		return nil, fmt.Errorf("register message handlers: %w", msgsErr)
	}
	return h, nil
}

func (h *EventHandler) registerMessageHandlers() error {
	prefix := "slack." + h.svc.integrationName
	return errors.Join(
		h.messages.AddCommandHandlers(
			rez.NewCommandHandler(prefix+".handle_events_api_callback_event", h.handleEventsApiCallbackEvent),
			rez.NewCommandHandler(prefix+".handle_slash_command", h.handleSlashCommand),
			rez.NewCommandHandler(prefix+".handle_interaction", h.handleInteractionCallback),
		),
	)
}

func (h *EventHandler) OnSlashCommand(ctx context.Context, sc slack.SlashCommand) error {
	return h.messages.SendCommand(ctx, handleSlashCommandEvent{Command: sc})
}

func (h *EventHandler) OnInteractionCallback(ctx context.Context, data []byte) error {
	return h.messages.SendCommand(ctx, handleInteractionCallbackEvent{Data: data})
}

func (h *EventHandler) OnCallbackEvent(ctx context.Context, ev *slackevents.EventsAPICallbackEvent, data []byte) error {
	if ev.InnerEvent == nil {
		return nil
	}
	var inner slackevents.EventsAPIInnerEvent
	if jsonErr := json.Unmarshal(*ev.InnerEvent, &inner); jsonErr != nil {
		return fmt.Errorf("inner event json: %w", jsonErr)
	}
	innerType := slackevents.EventsAPIType(inner.Type)
	if h.processCallbackEvents.Contains(innerType) {
		/*
			pe := rez.ProviderEvent{
				Provider:           integrationName,
				ProviderSource:     sourceEventsApiCallback,
				ProviderSubjectRef: h.callbackEventProviderSubjectRef(ev, data),
				ProviderEventRef:   ev.EventID,
				Payload:            data,
			}
			if ingestErr := h.provEvents.Ingest(ctx, pe); ingestErr != nil {
				return fmt.Errorf("ingest event: %w", ingestErr)
			}
		*/
	}
	if h.respondCallbackEvents.Contains(innerType) {
		cmd := handleEventsApiCallbackEvent{Data: data}
		if publishErr := h.messages.SendCommand(ctx, cmd); publishErr != nil {
			return fmt.Errorf("publish callback event: %w", publishErr)
		}
	}
	return nil
}

func (h *EventHandler) OnAppRateLimitedEvent(ctx context.Context) error {
	slog.Warn("slack app rate limited")
	return nil
}

func (h *EventHandler) OnOptions(ctx context.Context, data []byte) error {
	slog.Warn("not handling slack options event")
	return nil
}

type handleSlashCommandEvent struct {
	Command slack.SlashCommand
}

func (h *EventHandler) handleSlashCommand(ctx context.Context, cmd *handleSlashCommandEvent) error {
	return h.svc.handleSlashCommand(ctx, &cmd.Command)
}

type handleInteractionCallbackEvent struct {
	Data []byte
}

func (h *EventHandler) handleInteractionCallback(ctx context.Context, ev *handleInteractionCallbackEvent) error {
	var ic slack.InteractionCallback
	if err := ic.UnmarshalJSON(ev.Data); err != nil {
		return fmt.Errorf("invalid interaction payload: %w", err)
	}
	return h.svc.handleInteractionCallback(ctx, &ic)
}

type handleEventsApiCallbackEvent struct {
	Data []byte
}

func (h *EventHandler) handleEventsApiCallbackEvent(ctx context.Context, ev *handleEventsApiCallbackEvent) error {
	cb, parseErr := slackevents.ParseEvent(ev.Data, slackevents.OptionNoVerifyToken())
	if parseErr != nil {
		return fmt.Errorf("parse event: %w", parseErr)
	}
	return h.svc.handleEventsApiEvent(ctx, &cb)
}
