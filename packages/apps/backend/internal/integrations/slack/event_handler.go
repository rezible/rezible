package slackintegration

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"

	rez "github.com/rezible/rezible"
)

type EventHandler struct {
	integrationName string

	provEvents                rez.ProviderEventService
	publishProviderEventTypes mapset.Set[slackevents.EventsAPIType]

	messages          rez.MessageService
	respondEventTypes mapset.Set[slackevents.EventsAPIType]
}

type slashCommandEvent struct {
	integrationName string
	Command         slack.SlashCommand
}

func (h *EventHandler) OnSlashCommand(ctx context.Context, sc slack.SlashCommand) error {
	return h.messages.PublishEvent(ctx, slashCommandEvent{
		integrationName: h.integrationName,
		Command:         sc,
	})
}

type interactionCallbackEvent struct {
	integrationName string
	Data            []byte
}

func (h *EventHandler) OnInteractionCallback(ctx context.Context, data []byte) error {
	return h.messages.PublishEvent(ctx, interactionCallbackEvent{
		integrationName: h.integrationName,
		Data:            data,
	})
}

type handleEventsApiCallbackEvent struct {
	integrationName string
	Data            []byte
}

func (h *EventHandler) OnEventsApiCallback(ctx context.Context, ev *slackevents.EventsAPICallbackEvent, data []byte) error {
	if ev.InnerEvent == nil {
		return nil
	}
	var inner slackevents.EventsAPIInnerEvent
	if jsonErr := json.Unmarshal(*ev.InnerEvent, &inner); jsonErr != nil {
		return fmt.Errorf("inner event json: %w", jsonErr)
	}
	innerType := slackevents.EventsAPIType(inner.Type)
	if h.respondEventTypes.Contains(innerType) {
		publishErr := h.messages.SendCommand(ctx, handleEventsApiCallbackEvent{
			integrationName: h.integrationName,
			Data:            data,
		})
		if publishErr != nil {
			return fmt.Errorf("publish callback event: %w", publishErr)
		}
	}
	/*
		if h.publishProviderEventTypes.Contains(innerType) {
			pe := rez.ProviderEvent{
				Provider:           h.integrationName,
				ProviderSource:     "events_api",
				ProviderSubjectRef: fmt.Sprintf("slack:event:%s", ev.EventID),
				ProviderEventRef:   ev.EventID,
				Payload:            data,
			}
			if ingestErr := h.provEvents.Ingest(ctx, pe); ingestErr != nil {
				return fmt.Errorf("ingest event: %w", ingestErr)
			}
		}
	*/
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
