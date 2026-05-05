package slack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/rezible/rezible/ent/normalizedevent"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/internal/projections"
)

const (
	callbackEventsSource = "events_api/callback"
)

type eventHandler struct {
	services *rez.Services
}

func (i *integration) makeEventHandler() (*eventHandler, error) {
	svcs := i.services
	h := &eventHandler{services: svcs}

	svcs.ProviderEvents.RegisterEventProcessor(integrationName, callbackEventsSource, &callbackEventProcessor{handler: h})

	if msgsErr := h.registerMessageHandlers(svcs.Messages); msgsErr != nil {
		return nil, fmt.Errorf("message handlers: %w", msgsErr)
	}

	return h, nil
}

func (h *eventHandler) registerMessageHandlers(msgs rez.MessageService) error {
	return errors.Join(
		msgs.AddEventHandlers(
			rez.NewEventHandler("slack.events.callback_event", h.handleCallbackEvent),
			rez.NewEventHandler("slack.incidents.updated", h.onIncidentUpdated),
			rez.NewEventHandler("slack.incidents.milestone_updated", h.onIncidentMilestoneUpdated),
		),
		msgs.AddCommandHandlers(
			rez.NewCommandHandler("slack.handle_slash_command", h.handleSlashCommand),
			rez.NewCommandHandler("slack.handle_interaction", h.handleInteraction),
			rez.NewCommandHandler("slack.create_incident_channel", h.createIncidentChannel),
			rez.NewCommandHandler("slack.send_incident_milestone_message", h.sendIncidentMilestoneMessage),
		),
	)
}

func (h *eventHandler) OnSlashCommand(ctx context.Context, sc slack.SlashCommand) error {
	return h.services.Messages.SendCommand(ctx, handleSlashCommand{Command: sc})
}

func (h *eventHandler) OnInteractionCallback(ctx context.Context, data []byte) error {
	return h.services.Messages.SendCommand(ctx, handleInteraction{Data: data})
}

func (h *eventHandler) OnCallbackEvent(ctx context.Context, ev *slackevents.EventsAPICallbackEvent, data []byte) error {
	if ev.InnerEvent == nil {
		return nil
	}
	var inner slackevents.EventsAPIInnerEvent
	if jsonErr := json.Unmarshal(*ev.InnerEvent, &inner); jsonErr != nil {
		return fmt.Errorf("inner event json: %w", jsonErr)
	}
	innerType := slackevents.EventsAPIType(inner.Type)
	if processCallbackEventTypes.Contains(innerType) {
		pe := rez.ProviderEvent{
			Provider:        integrationName,
			Source:          callbackEventsSource,
			DedupeKey:       ev.EventID,
			Payload:         data,
			RequestMetadata: map[string]string{},
		}
		if ingestErr := h.services.ProviderEvents.Ingest(ctx, pe); ingestErr != nil {
			return fmt.Errorf("ingest event: %w", ingestErr)
		}
	}
	if respondCallbackInnerEvents.Contains(innerType) {
		if publishErr := h.services.Messages.PublishEvent(ctx, callbackEvent{Data: data}); publishErr != nil {
			return fmt.Errorf("publish callback event: %w", publishErr)
		}
	}
	return nil
}

func (h *eventHandler) OnAppRateLimitedEvent(ctx context.Context) error {
	slog.Warn("slack app rate limited")
	return nil
}

func (h *eventHandler) OnOptions(ctx context.Context, data []byte) error {
	slog.Warn("not handling slack options event")
	return nil
}

func (h *eventHandler) withChatService(ctx context.Context, ids installIds, fn func(*ChatService) error) error {
	ci, lookupErr := lookupTenantIntegration(ctx, h.services.Integrations, ids)
	if lookupErr != nil {
		return lookupErr
	}
	if ci == nil {
		slog.Warn("received slack event with no configured integration found!",
			"teamId", ids.TeamId,
			"enterpriseId", ids.EnterpriseId,
		)
		return nil
	}
	return fn(newChatService(ci))
}

func (h *eventHandler) withIncidentUpdateProcessor(ctx context.Context, id uuid.UUID, fn func(*incidentUpdateProcessor) error) error {
	intg, lookupErr := h.services.Integrations.GetConfigured(ctx, integrationName)
	if lookupErr != nil {
		if ent.IsNotFound(lookupErr) {
			return nil
		}
		return fmt.Errorf("getting configured integration: %w", lookupErr)
	}
	ci, ok := intg.(*ConfiguredIntegration)
	if !ok {
		return fmt.Errorf("failed to cast to *ConfiguredIntegration")
	}
	p, procErr := newIncidentUpdateProcessor(ctx, newChatService(ci), h.services, id)
	if procErr != nil {
		return fmt.Errorf("creating incident update processor: %w", procErr)
	}
	return fn(p)
}

func (h *eventHandler) onIncidentUpdated(ctx context.Context, ev *rez.EventOnIncidentUpdated) error {
	return h.withIncidentUpdateProcessor(ctx, ev.IncidentId, func(p *incidentUpdateProcessor) error {
		return p.processIncidentUpdate(ctx)
	})
}

func (h *eventHandler) onIncidentMilestoneUpdated(ctx context.Context, ev *rez.EventOnIncidentMilestoneUpdated) error {
	return h.withIncidentUpdateProcessor(ctx, ev.IncidentId, func(p *incidentUpdateProcessor) error {
		return p.processIncidentMilestoneUpdate(ctx, ev.MilestoneId)
	})
}

type cmdCreateIncidentChannel struct {
	IncidentId uuid.UUID
}

func (h *eventHandler) createIncidentChannel(ctx context.Context, ev *cmdCreateIncidentChannel) error {
	return h.withIncidentUpdateProcessor(ctx, ev.IncidentId, func(p *incidentUpdateProcessor) error {
		return p.createIncidentChannel(ctx)
	})
}

type cmdSendIncidentMilestoneMessage struct {
	IncidentId  uuid.UUID
	MilestoneId uuid.UUID
}

func (h *eventHandler) sendIncidentMilestoneMessage(ctx context.Context, ev *cmdSendIncidentMilestoneMessage) error {
	return h.withIncidentUpdateProcessor(ctx, ev.IncidentId, func(p *incidentUpdateProcessor) error {
		return p.sendIncidentMilestoneMessage(ctx, ev.MilestoneId)
	})
}

type handleSlashCommand struct {
	Command slack.SlashCommand
}

func (h *eventHandler) handleSlashCommand(ctx context.Context, cmd *handleSlashCommand) error {
	ids := installIds{TeamId: cmd.Command.TeamID, EnterpriseId: cmd.Command.EnterpriseID}
	return h.withChatService(ctx, ids, func(chat *ChatService) error {
		return chat.handleSlashCommand(ctx, &cmd.Command)
	})
}

type handleInteraction struct {
	Data []byte
}

func (h *eventHandler) handleInteraction(ctx context.Context, ev *handleInteraction) error {
	var ic slack.InteractionCallback
	if err := ic.UnmarshalJSON(ev.Data); err != nil {
		return fmt.Errorf("invalid interaction payload: %w", err)
	}
	ids := installIds{TeamId: ic.Team.ID, EnterpriseId: ic.Enterprise.ID}
	return h.withChatService(ctx, ids, func(chat *ChatService) error {
		return chat.handleInteractionCallback(ctx, &ic)
	})
}

type callbackEvent struct {
	Data []byte
}

func (h *eventHandler) handleCallbackEvent(ctx context.Context, ev *callbackEvent) error {
	cb, parseErr := slackevents.ParseEvent(ev.Data, slackevents.OptionNoVerifyToken())
	if parseErr != nil {
		return fmt.Errorf("parse event: %w", parseErr)
	}
	ids := installIds{TeamId: cb.TeamID, EnterpriseId: cb.EnterpriseID}
	return h.withChatService(ctx, ids, func(chat *ChatService) error {
		return chat.handleCallbackEvent(ctx, &cb)
	})
}

type callbackEventProcessor struct {
	handler *eventHandler
}

var processCallbackEventTypes = mapset.NewSet(
	slackevents.AppMention,
	slackevents.Message,
)

func (p *callbackEventProcessor) Process(ctx context.Context, prov rez.ProviderEvent) (ent.NormalizedEvents, error) {
	ev, parseErr := slackevents.ParseEvent(prov.Payload, slackevents.OptionNoVerifyToken())
	if parseErr != nil {
		return nil, fmt.Errorf("parse event: %w", parseErr)
	}

	ids := installIds{TeamId: ev.TeamID, EnterpriseId: ev.EnterpriseID}
	ci, lookupErr := lookupTenantIntegration(ctx, p.handler.services.Integrations, ids)
	if lookupErr != nil {
		return nil, lookupErr
	}
	if ci == nil {
		slog.WarnContext(ctx, "received slack event with no configured integration found!",
			"teamId", ids.TeamId,
			"enterpriseId", ids.EnterpriseId,
		)
		return nil, nil
	}

	result := &ent.NormalizedEvent{
		Provider:          integrationName,
		ProviderSource:    callbackEventsSource,
		Kind:              normalizedevent.KindChatMessage,
		SubjectKind:       "message",
		DedupeKey:         prov.DedupeKey,
		ProcessingVersion: "slack.chat-message-posted.v1",
	}

	attrs := projections.ChatMessageAttributes{
		ConversationExternalRef: "",
		Body:                    "",
		SenderExternalRef:       "",
		ThreadExternalRef:       "",
	}

	if cb, ok := ev.Data.(*slackevents.EventsAPICallbackEvent); ok {
		result.ProviderEventRef = cb.EventID
		if result.DedupeKey == "" {
			result.DedupeKey = cb.EventID
		}
	}

	var (
		ts      string
		eventTS string
	)

	switch data := ev.InnerEvent.Data.(type) {
	case *slackevents.MessageEvent:
		attrs.ConversationExternalRef = data.Channel
		attrs.SenderExternalRef = data.User
		attrs.Body = data.Text
		attrs.ThreadExternalRef = data.ThreadTimeStamp

		ts = data.TimeStamp
		eventTS = data.EventTimeStamp
	case *slackevents.AppMentionEvent:
		attrs.ConversationExternalRef = data.Channel
		attrs.SenderExternalRef = data.User
		attrs.Body = data.Text
		attrs.ThreadExternalRef = data.ThreadTimeStamp

		ts = data.TimeStamp
		eventTS = data.EventTimeStamp
	default:
		return nil, nil
	}

	if attrs.ConversationExternalRef == "" || ts == "" {
		return nil, nil
	}

	occurredAt := tryConvertTs(ts, tryConvertTs(eventTS, prov.ReceivedAt))
	receivedAt := prov.ReceivedAt
	if receivedAt.IsZero() {
		receivedAt = occurredAt
	}
	result.OccurredAt = occurredAt
	result.ReceivedAt = receivedAt

	// TODO: this may be ambiguous?
	workspaceID := ev.TeamID
	if workspaceID == "" {
		workspaceID = ci.teamId()
	}

	result.SubjectRef = fmt.Sprintf("slack:%s:%s:%s", workspaceID, attrs.ConversationExternalRef, ts)
	if result.ProviderEventRef == "" {
		result.ProviderEventRef = result.SubjectRef
	}

	result.Attributes = attrs.Encode()

	return ent.NormalizedEvents{result}, nil
}
