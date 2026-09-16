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

	pipeline          rez.ProviderEventPipelineService
	publishEventTypes mapset.Set[slackevents.EventsAPIType]

	mq                rez.MessageQueue
	respondEventTypes mapset.Set[slackevents.EventsAPIType]
}

func makeAppEventHandler(app App, mq rez.MessageQueue, pipeline rez.ProviderEventPipelineService) *appEventHandler {
	return &appEventHandler{
		integrationName:   app.IntegrationName(),
		mq:                mq,
		pipeline:          pipeline,
		publishEventTypes: mapset.NewSet(app.PublishEventTypes()...),
		respondEventTypes: mapset.NewSet(app.RespondEventTypes()...),
	}
}

type slashCommandEvent struct {
	IntegrationName string
	Command         slack.SlashCommand
}

func (slashCommandEvent) MessageName() string { return "slack.slash-command-received.v1" }

func (h *appEventHandler) OnSlashCommand(ctx context.Context, sc slack.SlashCommand) error {
	return h.mq.Publish(ctx, &slashCommandEvent{
		IntegrationName: h.integrationName,
		Command:         sc,
	})
}

type interactionCallbackEvent struct {
	IntegrationName string
	Data            []byte
}

func (interactionCallbackEvent) MessageName() string { return "slack.interaction-received.v1" }

func (h *appEventHandler) OnInteractionCallback(ctx context.Context, data []byte) error {
	return h.mq.Publish(ctx, &interactionCallbackEvent{
		IntegrationName: h.integrationName,
		Data:            data,
	})
}

type handleEventsApiCallbackEvent struct {
	IntegrationName string
	Data            []byte
}

func (handleEventsApiCallbackEvent) MessageName() string { return "slack.events-api-received.v1" }

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
		if publishErr := h.mq.Publish(ctx, cbEv); publishErr != nil {
			return fmt.Errorf("publish callback event: %w", publishErr)
		}
	}
	if h.publishEventTypes.Contains(innerType) {
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
		if ingestErr := h.pipeline.Ingest(ctx, pe); ingestErr != nil {
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
