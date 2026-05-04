package slack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
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
		slog.Warn("received slack event with no configured integration found!",
			"teamId", ids.TeamId,
			"enterpriseId", ids.EnterpriseId,
		)
		return nil, nil
	}

	var (
		channelID       string
		channelType     string
		userID          string
		text            string
		ts              string
		threadTS        string
		eventTS         string
		subtype         string
		botMentioned    bool
		callbackEventID string
		sourceEventKey  string
	)

	if cb, ok := ev.Data.(*slackevents.EventsAPICallbackEvent); ok {
		callbackEventID = cb.EventID
		sourceEventKey = cb.EventID
	}

	switch data := ev.InnerEvent.Data.(type) {
	case *slackevents.MessageEvent:
		channelID = data.Channel
		channelType = data.ChannelType
		userID = data.User
		text = data.Text
		ts = data.TimeStamp
		threadTS = data.ThreadTimeStamp
		eventTS = data.EventTimeStamp
		subtype = data.SubType
		botMentioned = strings.Contains(data.Text, "<@"+ci.botUserID()+">")
	case *slackevents.AppMentionEvent:
		channelID = data.Channel
		userID = data.User
		text = data.Text
		ts = data.TimeStamp
		threadTS = data.ThreadTimeStamp
		eventTS = data.EventTimeStamp
		botMentioned = true
	default:
		return nil, nil
	}
	if channelID == "" || ts == "" {
		return nil, nil
	}
	occurredAt := convertSlackTs(eventTS)
	if occurredAt.IsZero() {
		occurredAt = convertSlackTs(ts)
	}
	if occurredAt.IsZero() {
		occurredAt = prov.ReceivedAt
	}
	receivedAt := prov.ReceivedAt
	if receivedAt.IsZero() {
		receivedAt = occurredAt
	}

	workspaceID := ev.TeamID
	if workspaceID == "" {
		workspaceID = ci.teamId()
	}
	subjectExternalRef := fmt.Sprintf("slack:%s:%s:%s", workspaceID, channelID, ts)
	if sourceEventKey == "" {
		sourceEventKey = subjectExternalRef
	}

	dedupeKey := prov.DedupeKey
	if dedupeKey == "" {
		dedupeKey = callbackEventID
	}

	normalized := ent.NormalizedEvents{
		{
			Provider:           integrationName,
			Source:             callbackEventsSource,
			Kind:               "ChatMessagePosted",
			SubjectKind:        "chat_message",
			SubjectExternalRef: subjectExternalRef,
			SourceEventKey:     sourceEventKey,
			DedupeKey:          dedupeKey,
			OccurredAt:         occurredAt,
			ReceivedAt:         receivedAt,
			ProcessingVersion:  "chat-message-posted.v1",
			Attributes: map[string]any{
				"team_id":          ev.TeamID,
				"enterprise_id":    ev.EnterpriseID,
				"channel_id":       channelID,
				"channel_type":     channelType,
				"user_id":          userID,
				"text":             text,
				"timestamp":        ts,
				"thread_timestamp": threadTS,
				"subtype":          subtype,
				"bot_mentioned":    botMentioned,
			},
		},
	}

	return normalized, nil
}
