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

	createInstallationCtx createInstallationCtxFunc
	onCommand             commandFunc
	onInteraction         interactionFunc

	pipeline          rez.ProviderEventPipelineService
	publishEventTypes mapset.Set[slackevents.EventsAPIType]

	mq                rez.MessageQueue
	respondEventTypes mapset.Set[slackevents.EventsAPIType]
}

func makeAppEventHandler(app App, mq rez.MessageQueue, pipeline rez.ProviderEventPipelineService, ctxFn createInstallationCtxFunc, onCommand commandFunc, onInteraction interactionFunc) *appEventHandler {
	return &appEventHandler{
		integrationName:       app.IntegrationName(),
		createInstallationCtx: ctxFn,
		onCommand:             onCommand,
		onInteraction:         onInteraction,
		mq:                    mq,
		pipeline:              pipeline,
		publishEventTypes:     mapset.NewSet(app.PublishEventTypes()...),
		respondEventTypes:     mapset.NewSet(app.RespondEventTypes()...),
	}
}

// commandFunc answers a slash command with the reply to send in the response, or nil for none.
type commandFunc func(context.Context, slack.SlashCommand) *slack.WebhookMessage

// interactionFunc answers an interaction with the response to send, or nil for an empty one.
type interactionFunc func(context.Context, []byte) (*slack.ViewSubmissionResponse, error)

// OnSlashCommand answers the command while Slack waits, so a modal opens before its trigger ID expires.
func (h *appEventHandler) OnSlashCommand(ctx context.Context, sc slack.SlashCommand) *slack.WebhookMessage {
	return h.onCommand(ctx, sc)
}

// OnInteractionCallback answers the interaction while Slack waits, so a submitted modal can show errors.
func (h *appEventHandler) OnInteractionCallback(ctx context.Context, data []byte) (*slack.ViewSubmissionResponse, error) {
	return h.onInteraction(ctx, data)
}

type eventsApiCallbackEvent struct {
	IntegrationName string
	Data            []byte
}

func (eventsApiCallbackEvent) MessageName() string {
	return "slack.events-api-received.v1"
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
		cbEv := &eventsApiCallbackEvent{
			IntegrationName: h.integrationName,
			Data:            data,
		}
		if publishErr := h.mq.Publish(ctx, cbEv); publishErr != nil {
			return fmt.Errorf("publish callback event: %w", publishErr)
		}
	}
	if h.publishEventTypes.Contains(innerType) {
		installIds := InstallationIds{TeamId: ev.TeamID, EnterpriseId: ev.EnterpriseID}
		installationCtx, installationErr := h.createInstallationCtx(ctx, installIds)
		if installationErr != nil {
			return fmt.Errorf("lookup integration: %w", installationErr)
		}

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
		if ingestErr := h.pipeline.Ingest(installationCtx, pe); ingestErr != nil {
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
