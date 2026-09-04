package slackintegration

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"

	rez "github.com/rezible/rezible"
)

type appEventHandler struct {
	integrationName string

	eventPipeline                          rez.ProviderEventPipelineService
	providerEventPipelinePublishEventTypes mapset.Set[slackevents.EventsAPIType]

	messages          rez.MessageService
	respondEventTypes mapset.Set[slackevents.EventsAPIType]
}

func makeAppEventHandler(app App, msgs rez.MessageService, eventPipeline rez.ProviderEventPipelineService) *appEventHandler {
	return &appEventHandler{
		integrationName:                        app.IntegrationName(),
		eventPipeline:                          eventPipeline,
		providerEventPipelinePublishEventTypes: mapset.NewSet(app.PublishProviderEventPipelineEventTypes()...),
		messages:                               msgs,
		respondEventTypes:                      mapset.NewSet(app.RespondEventTypes()...),
	}
}

type slashCommandEvent struct {
	IntegrationName string
	Command         slack.SlashCommand
}

func (h *appEventHandler) OnSlashCommand(ctx context.Context, sc slack.SlashCommand) error {
	return h.messages.Publish(ctx, &slashCommandEvent{
		IntegrationName: h.integrationName,
		Command:         sc,
	})
}

type interactionCallbackEvent struct {
	IntegrationName string
	Data            []byte
}

func (h *appEventHandler) OnInteractionCallback(ctx context.Context, data []byte) error {
	return h.messages.Publish(ctx, &interactionCallbackEvent{
		IntegrationName: h.integrationName,
		Data:            data,
	})
}

type handleEventsApiCallbackEvent struct {
	IntegrationName string
	Data            []byte
}

func (h *appEventHandler) OnEventsApiCallback(ctx context.Context, ev *slackevents.EventsAPICallbackEvent, data []byte) error {
	if ev.InnerEvent == nil {
		return nil
	}
	var inner slackevents.EventsAPIInnerEvent
	if jsonErr := json.Unmarshal(*ev.InnerEvent, &inner); jsonErr != nil {
		return fmt.Errorf("inner event json: %w", jsonErr)
	}
	innerType := slackevents.EventsAPIType(inner.Type)
	if h.respondEventTypes.Contains(innerType) {
		fmt.Printf("handle callback event: %s\n", inner.Type)
		cbEv := &handleEventsApiCallbackEvent{
			IntegrationName: h.integrationName,
			Data:            data,
		}
		if publishErr := h.messages.Publish(ctx, cbEv); publishErr != nil {
			return fmt.Errorf("publish callback event: %w", publishErr)
		}
	}
	if h.providerEventPipelinePublishEventTypes.Contains(innerType) {
		namespace := ev.TeamID
		if namespace == "" {
			namespace = ev.EnterpriseID
		}
		pe := rez.ProviderEvent{
			Provider:            ProviderName,
			ProviderNamespace:   namespace,
			ProviderEventSource: "events_api",
			ProviderEventRef:    ev.EventID,
			Attributes:          data,
			ReceivedAt:          time.Now().UTC(),
		}
		if ingestErr := h.eventPipeline.Ingest(ctx, pe); ingestErr != nil {
			return fmt.Errorf("ingest event: %w", ingestErr)
		}
	}
	return nil
}

func (h *appEventHandler) OnAppRateLimitedEvent(ctx context.Context) error {
	slog.Warn("slack app rate limited")
	return nil
}

func (h *appEventHandler) OnOptions(ctx context.Context, data []byte) error {
	slog.Warn("not handling slack options event")
	return nil
}
