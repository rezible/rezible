package slack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"

	rez "github.com/rezible/rezible"
)

type messageHandler struct {
	appCfg       rez.AppConfig
	messages     rez.MessageService
	provEvents   rez.ProviderEventService
	integrations rez.IntegrationsService
	incidents    rez.IncidentService
}

func (i *Integration) makeMessageHandler(cfg rez.Config, msgs rez.MessageService, provEvents rez.ProviderEventService) (*messageHandler, error) {
	h := &messageHandler{
		appCfg:       cfg.App,
		messages:     msgs,
		provEvents:   provEvents,
		integrations: i.integrations,
		incidents:    i.incidents,
	}
	if msgsErr := h.registerHandlers(); msgsErr != nil {
		return nil, fmt.Errorf("message handlers: %w", msgsErr)
	}
	return h, nil
}

func (h *messageHandler) registerHandlers() error {
	return errors.Join(
		h.messages.AddEventHandlers(
			rez.NewEventHandler("slack.events.callback_event", h.handleCallbackEvent),
			rez.NewEventHandler("slack.incidents.updated", h.onIncidentUpdated),
			rez.NewEventHandler("slack.incidents.milestone_updated", h.onIncidentMilestoneUpdated),
		),
		h.messages.AddCommandHandlers(
			rez.NewCommandHandler("slack.handle_slash_command", h.handleSlashCommand),
			rez.NewCommandHandler("slack.handle_interaction", h.handleInteraction),
			rez.NewCommandHandler("slack.create_incident_channel", h.createIncidentChannel),
			rez.NewCommandHandler("slack.send_incident_milestone_message", h.sendIncidentMilestoneMessage),
		),
	)
}

func (h *messageHandler) OnSlashCommand(ctx context.Context, sc slack.SlashCommand) error {
	return h.messages.SendCommand(ctx, handleSlashCommand{Command: sc})
}

func (h *messageHandler) OnInteractionCallback(ctx context.Context, data []byte) error {
	return h.messages.SendCommand(ctx, handleInteraction{Data: data})
}

func (h *messageHandler) OnCallbackEvent(ctx context.Context, ev *slackevents.EventsAPICallbackEvent, data []byte) error {
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
			Provider:           integrationName,
			ProviderSource:     sourceEventsApiCallback,
			ProviderSubjectRef: h.callbackEventProviderSubjectRef(ev, data),
			ProviderEventRef:   ev.EventID,
			Payload:            data,
		}
		if ingestErr := h.provEvents.Ingest(ctx, pe); ingestErr != nil {
			return fmt.Errorf("ingest event: %w", ingestErr)
		}
	}
	if respondCallbackInnerEvents.Contains(innerType) {
		if publishErr := h.messages.PublishEvent(ctx, callbackEvent{Data: data}); publishErr != nil {
			return fmt.Errorf("publish callback event: %w", publishErr)
		}
	}
	return nil
}

func (h *messageHandler) callbackEventProviderSubjectRef(ev *slackevents.EventsAPICallbackEvent, data []byte) string {
	type callbackEventPayload struct {
		TeamID       string `json:"team_id"`
		EnterpriseID string `json:"enterprise_id"`
		Event        struct {
			Channel string `json:"channel"`
			Ts      string `json:"ts"`
			EventTs string `json:"event_ts"`
		} `json:"event"`
	}
	var payload callbackEventPayload
	if err := json.Unmarshal(data, &payload); err == nil {
		workspaceID := payload.TeamID
		if workspaceID == "" {
			workspaceID = payload.EnterpriseID
		}
		ts := payload.Event.Ts
		if ts == "" {
			ts = payload.Event.EventTs
		}
		if workspaceID != "" && payload.Event.Channel != "" && ts != "" {
			return fmt.Sprintf("slack:%s:%s:%s", workspaceID, payload.Event.Channel, ts)
		}
	}

	workspaceID := ev.TeamID
	if workspaceID == "" {
		workspaceID = ev.EnterpriseID
	}
	if ev.EventID != "" {
		return fmt.Sprintf("slack:event:%s", ev.EventID)
	}
	return "slack:event_callback"
}

func (h *messageHandler) OnAppRateLimitedEvent(ctx context.Context) error {
	slog.Warn("slack app rate limited")
	return nil
}

func (h *messageHandler) OnOptions(ctx context.Context, data []byte) error {
	slog.Warn("not handling slack options event")
	return nil
}

func (h *messageHandler) withChatService(ctx context.Context, ids installIds, fn func(*ChatService) error) error {
	ci, lookupErr := lookupTenantIntegration(ctx, h.integrations, ids)
	if lookupErr != nil {
		return lookupErr
	}
	if ci == nil {
		slog.Warn("received slack event with no configured Integration found!",
			"teamId", ids.TeamId,
			"enterpriseId", ids.EnterpriseId,
		)
		return nil
	}
	return fn(newChatService(ci))
}

func (h *messageHandler) withIncidentUpdateProcessor(ctx context.Context, id uuid.UUID, fn func(*incidentUpdateProcessor) error) error {
	p, procErr := h.newIncidentUpdateProcessor(ctx, id)
	if procErr != nil {
		return fmt.Errorf("creating incident update processor: %w", procErr)
	}
	return fn(p)
}

func (h *messageHandler) onIncidentUpdated(ctx context.Context, ev *rez.EventOnIncidentUpdated) error {
	return h.withIncidentUpdateProcessor(ctx, ev.IncidentId, func(p *incidentUpdateProcessor) error {
		return p.processIncidentUpdate(ctx)
	})
}

func (h *messageHandler) onIncidentMilestoneUpdated(ctx context.Context, ev *rez.EventOnIncidentMilestoneUpdated) error {
	return h.withIncidentUpdateProcessor(ctx, ev.IncidentId, func(p *incidentUpdateProcessor) error {
		return p.processIncidentMilestoneUpdate(ctx, ev.MilestoneId)
	})
}

type cmdCreateIncidentChannel struct {
	IncidentId uuid.UUID
}

func (h *messageHandler) createIncidentChannel(ctx context.Context, ev *cmdCreateIncidentChannel) error {
	return h.withIncidentUpdateProcessor(ctx, ev.IncidentId, func(p *incidentUpdateProcessor) error {
		return p.createIncidentChannel(ctx)
	})
}

type cmdSendIncidentMilestoneMessage struct {
	IncidentId  uuid.UUID
	MilestoneId uuid.UUID
}

func (h *messageHandler) sendIncidentMilestoneMessage(ctx context.Context, ev *cmdSendIncidentMilestoneMessage) error {
	return h.withIncidentUpdateProcessor(ctx, ev.IncidentId, func(p *incidentUpdateProcessor) error {
		return p.sendIncidentMilestoneMessage(ctx, ev.MilestoneId)
	})
}

type handleSlashCommand struct {
	Command slack.SlashCommand
}

func (h *messageHandler) handleSlashCommand(ctx context.Context, cmd *handleSlashCommand) error {
	ids := installIds{TeamId: cmd.Command.TeamID, EnterpriseId: cmd.Command.EnterpriseID}
	return h.withChatService(ctx, ids, func(chat *ChatService) error {
		return chat.handleSlashCommand(ctx, &cmd.Command)
	})
}

type handleInteraction struct {
	Data []byte
}

func (h *messageHandler) handleInteraction(ctx context.Context, ev *handleInteraction) error {
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

func (h *messageHandler) handleCallbackEvent(ctx context.Context, ev *callbackEvent) error {
	cb, parseErr := slackevents.ParseEvent(ev.Data, slackevents.OptionNoVerifyToken())
	if parseErr != nil {
		return fmt.Errorf("parse event: %w", parseErr)
	}
	ids := installIds{TeamId: cb.TeamID, EnterpriseId: cb.EnterpriseID}
	return h.withChatService(ctx, ids, func(chat *ChatService) error {
		return chat.handleCallbackEvent(ctx, &cb)
	})
}
