package slack

import (
	"context"
	"errors"
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
	if msgsErr := h.registerMessageHandlers(); msgsErr != nil {
		return nil, fmt.Errorf("message handlers: %w", msgsErr)
	}
	return h, nil
}

func (h *eventHandler) registerMessageHandlers() error {
	return errors.Join(
		h.services.Messages.AddEventHandlers(
			rez.NewEventHandler("slack.events.callback_event", h.onCallbackEvent),
			rez.NewEventHandler("slack.incidents.on_updated", h.onIncidentUpdated),
			rez.NewEventHandler("slack.incidents.on_milestone_updated", h.onIncidentMilestoneUpdated),
		),
		h.services.Messages.AddCommandHandlers(
			rez.NewCommandHandler("slack.process_slash_command", h.processSlashCommand),
			rez.NewCommandHandler("slack.process_interaction", h.processInteraction),
			rez.NewCommandHandler("slack.create_incident_channel", h.createIncidentChannel),
			rez.NewCommandHandler("slack.send_incident_milestone_message", h.sendIncidentMilestoneMessage),
		),
	)
}

func (h *eventHandler) OnSlashCommand(ctx context.Context, sc slack.SlashCommand) error {
	return h.services.Messages.SendCommand(ctx, processSlashCommand{Command: sc})
}

func (h *eventHandler) OnInteractionCallback(ctx context.Context, data []byte) error {
	return h.services.Messages.SendCommand(ctx, processInteraction{Data: data})
}

func (h *eventHandler) OnCallbackEvent(ctx context.Context, data []byte) error {
	return h.services.Messages.PublishEvent(ctx, callbackEvent{Data: data})
}

func (h *eventHandler) OnAppRateLimitedEvent(ctx context.Context) error {
	log.Warn().Msg("slack app rate limited")
	return nil
}

func (h *eventHandler) OnOptions(ctx context.Context, data []byte) error {
	log.Warn().Msg("not handling slack options event")
	return nil
}

func (h *eventHandler) withChatService(ctx context.Context, ids installIds, fn func(*ChatService) error) error {
	ci, lookupErr := lookupTenantIntegration(ctx, h.services.Integrations, ids)
	if lookupErr != nil {
		return lookupErr
	}
	if ci == nil {
		log.Warn().
			Str("teamId", ids.TeamId).
			Str("enterpriseId", ids.EnterpriseId).
			Msg("received slack event with no configured integration found!")
		return nil
	}
	return fn(newChatService(ci))
}

func (h *eventHandler) withIncidentUpdateProcessor(ctx context.Context, fn func(*incidentUpdateProcessor) error) error {
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
	return fn(newIncidentUpdateProcessor(newChatService(ci), h.services))
}

func (h *eventHandler) onIncidentUpdated(ctx context.Context, ev *rez.EventOnIncidentUpdated) error {
	return h.withIncidentUpdateProcessor(ctx, func(p *incidentUpdateProcessor) error {
		return p.processIncidentUpdate(ctx, ev.IncidentId)
	})
}

func (h *eventHandler) onIncidentMilestoneUpdated(ctx context.Context, ev *rez.EventOnIncidentMilestoneUpdated) error {
	return h.withIncidentUpdateProcessor(ctx, func(p *incidentUpdateProcessor) error {
		return p.processIncidentMilestoneUpdate(ctx, ev.MilestoneId)
	})
}

type cmdCreateIncidentChannel struct {
	IncidentId uuid.UUID
}

func (h *eventHandler) createIncidentChannel(ctx context.Context, ev *cmdCreateIncidentChannel) error {
	return h.withIncidentUpdateProcessor(ctx, func(p *incidentUpdateProcessor) error {
		return p.createIncidentChannel(ctx)
	})
}

type cmdSendIncidentMilestoneMessage struct {
	MilestoneId uuid.UUID `json:"mid"`
}

func (h *eventHandler) sendIncidentMilestoneMessage(ctx context.Context, ev *cmdSendIncidentMilestoneMessage) error {
	return h.withIncidentUpdateProcessor(ctx, func(p *incidentUpdateProcessor) error {
		return p.sendIncidentMilestoneMessage(ctx, ev.MilestoneId)
	})
}

type processSlashCommand struct {
	Command slack.SlashCommand
}

func (h *eventHandler) processSlashCommand(ctx context.Context, cmd *processSlashCommand) error {
	ids := installIds{TeamId: cmd.Command.TeamID, EnterpriseId: cmd.Command.EnterpriseID}
	return h.withChatService(ctx, ids, func(chat *ChatService) error {
		return chat.handleSlashCommand(ctx, &cmd.Command)
	})
}

type processInteraction struct {
	Data []byte
}

func (h *eventHandler) processInteraction(ctx context.Context, ev *processInteraction) error {
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

func (h *eventHandler) onCallbackEvent(ctx context.Context, ev *callbackEvent) error {
	cb, parseErr := slackevents.ParseEvent(ev.Data, slackevents.OptionNoVerifyToken())
	if parseErr != nil {
		return fmt.Errorf("parse event: %w", parseErr)
	}
	ids := installIds{TeamId: cb.TeamID, EnterpriseId: cb.EnterpriseID}
	return h.withChatService(ctx, ids, func(chat *ChatService) error {
		return chat.handleCallbackEvent(ctx, &cb)
	})
}
