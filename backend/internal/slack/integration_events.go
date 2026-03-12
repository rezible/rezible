package slack

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type eventHandler struct {
	services *rez.Services
}

func (i *integration) makeEventHandler() (*eventHandler, error) {
	h := &eventHandler{services: i.services}
	if slackErr := h.registerSlackHandlers(); slackErr != nil {
		return nil, fmt.Errorf("slack events: %w", slackErr)
	}
	if slackErr := h.registerIncidentHandlers(); slackErr != nil {
		return nil, fmt.Errorf("slack events: %w", slackErr)
	}
	return h, nil
}

func (h *eventHandler) registerSlackHandlers() error {
	evsErr := h.services.Messages.AddEventHandlers(
		rez.NewEventHandler("slack.events.callback", h.processCallbackEvent))
	if evsErr != nil {
		return fmt.Errorf("event handlers: %w", evsErr)
	}
	cmdsErr := h.services.Messages.AddCommandHandlers(
		rez.NewCommandHandler("slack.events.command", h.processSlashCommand),
		rez.NewCommandHandler("slack.events.interaction", h.processInteractionCallback))
	if cmdsErr != nil {
		return fmt.Errorf("command handlers: %w", cmdsErr)
	}
	return nil
}

func (h *eventHandler) registerIncidentHandlers() error {
	eventsErr := h.services.Messages.AddEventHandlers(
		rez.NewEventHandler("slack.on_incident_updated", h.onIncidentUpdated),
		rez.NewEventHandler("slack.on_incident_milestone_updated", h.onIncidentMilestoneUpdated))
	if eventsErr != nil {
		return fmt.Errorf("events: %w", eventsErr)
	}
	cmdsErr := h.services.Messages.AddCommandHandlers(
		rez.NewCommandHandler("slack.create_incident_channel", h.createIncidentChannel),
		rez.NewCommandHandler("slack.send_incident_milestone_message", h.sendIncidentMilestoneMessage))
	if cmdsErr != nil {
		return fmt.Errorf("commands: %w", cmdsErr)
	}
	return nil
}

func (h *eventHandler) lookupTenantChatService(ctx context.Context, teamId string, enterpriseId string) (*ChatService, error) {
	intg, lookupErr := lookupTenantIntegration(ctx, h.services.Integrations, teamId, enterpriseId)
	if lookupErr != nil {
		return nil, lookupErr
	}
	return newChatService(newConfiguredIntegration(h.services, intg))
}

func (h *eventHandler) makeIncidentUpdateProcessor(ctx context.Context, incidentId uuid.UUID) (*incidentUpdateProcessor, error) {
	intg, lookupErr := h.services.Integrations.Get(ctx, integrationName)
	if lookupErr != nil {
		if ent.IsNotFound(lookupErr) {
			return nil, nil
		}
		return nil, lookupErr
	}
	chat, chatErr := newChatService(newConfiguredIntegration(h.services, intg))
	if chatErr != nil {
		return nil, fmt.Errorf("make chat service: %w", chatErr)
	}
	inc, incErr := h.services.Incidents.Get(ctx, incidentId)
	if incErr != nil {
		return nil, fmt.Errorf("failed to lookup incident: %w", incErr)
	}
	return newIncidentUpdateProcessor(chat, h.services, inc), nil
}

func (h *eventHandler) onIncidentUpdated(ctx context.Context, ev *rez.EventOnIncidentUpdated) error {
	p, pErr := h.makeIncidentUpdateProcessor(ctx, ev.IncidentId)
	if p == nil {
		return pErr
	}
	return p.processUpdate(ctx)
}

func (h *eventHandler) onIncidentMilestoneUpdated(ctx context.Context, ev *rez.EventOnIncidentMilestoneUpdated) error {
	// TODO: check if we care about this milestone kind
	return h.services.Messages.SendCommand(ctx, &cmdSendIncidentMilestoneMessage{
		IncidentId:  ev.IncidentId,
		MilestoneId: ev.MilestoneId,
	})
}

type cmdCreateIncidentChannel struct {
	IncidentId uuid.UUID
}

func (h *eventHandler) createIncidentChannel(ctx context.Context, ev *cmdCreateIncidentChannel) error {
	p, pErr := h.makeIncidentUpdateProcessor(ctx, ev.IncidentId)
	if p == nil {
		return pErr
	}
	return p.createIncidentChannel(ctx)
}

type cmdSendIncidentMilestoneMessage struct {
	IncidentId  uuid.UUID `json:"iid"`
	MilestoneId uuid.UUID `json:"mid"`
}

func (h *eventHandler) sendIncidentMilestoneMessage(ctx context.Context, ev *cmdSendIncidentMilestoneMessage) error {
	p, pErr := h.makeIncidentUpdateProcessor(ctx, ev.IncidentId)
	if p == nil {
		return pErr
	}
	return p.sendIncidentMilestoneMessage(ctx, ev.MilestoneId)
}

func (h *eventHandler) SlashCommand(ctx context.Context, sc slack.SlashCommand) error {
	return h.services.Messages.SendCommand(ctx, processSlashCommand{Command: sc})
}

type processSlashCommand struct {
	Command slack.SlashCommand
}

func (h *eventHandler) processSlashCommand(ctx context.Context, cmd *processSlashCommand) error {
	sc := cmd.Command
	chat, chatErr := h.lookupTenantChatService(ctx, sc.TeamID, sc.EnterpriseID)
	if chatErr != nil {
		return fmt.Errorf("get user client: %w", chatErr)
	}
	handleErr := chat.handleSlashCommand(ctx, &sc)
	if handleErr != nil {
		log.Error().Err(handleErr).Msg("failed to handle slash command")
		return handleErr
	}
	return nil
}

func (h *eventHandler) InteractionCallback(ctx context.Context, ic *slack.InteractionCallback) error {
	return h.services.Messages.PublishEvent(ctx, interactionCallbackEvent{Data: ic})
}

type interactionCallbackEvent struct {
	Data *slack.InteractionCallback
}

func (h *eventHandler) processInteractionCallback(ctx context.Context, e *interactionCallbackEvent) error {
	chat, chatErr := h.lookupTenantChatService(ctx, e.Data.Team.ID, e.Data.Enterprise.ID)
	if chatErr != nil {
		return fmt.Errorf("get user client: %w", chatErr)
	}
	if handlerErr := chat.handleInteractionCallback(ctx, e.Data); handlerErr != nil {
		log.Error().Err(handlerErr).Str("type", string(e.Data.Type)).Msg("failed to handle interaction")
		return handlerErr
	}
	return nil
}

func (h *eventHandler) Options(ctx context.Context, body []byte) error {
	return nil
}

func (h *eventHandler) CallbackEvent(ctx context.Context, evt *slackevents.EventsAPIEvent) error {
	raw, jsonErr := json.Marshal(evt)
	if jsonErr != nil {
		return fmt.Errorf("marshal callback event: %w", jsonErr)
	}
	return h.services.Messages.PublishEvent(ctx, callbackEvent{RawEvent: raw})
}

type callbackEvent struct {
	RawEvent json.RawMessage
}

func (h *eventHandler) processCallbackEvent(ctx context.Context, ev *callbackEvent) error {
	cbe, parseErr := slackevents.ParseEvent(ev.RawEvent)
	if parseErr != nil {
		return fmt.Errorf("parse event: %w", parseErr)
	}
	chat, chatErr := h.lookupTenantChatService(ctx, cbe.TeamID, cbe.EnterpriseID)
	if chatErr != nil {
		return fmt.Errorf("get user client: %w", chatErr)
	}
	if handleErr := chat.handleCallbackEvent(ctx, &cbe); handleErr != nil {
		return fmt.Errorf("processing callback event: %w", handleErr)
	}
	return nil
}
