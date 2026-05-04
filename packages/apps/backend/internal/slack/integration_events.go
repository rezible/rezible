package slack

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"

	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
)

type eventHandler struct {
	services *rez.Services
}

func (i *integration) makeEventHandler() (*eventHandler, error) {
	h := &eventHandler{services: i.services}

	i.services.ProviderEvents.RegisterEventProcessor(&slackEventsAPIProcessor{handler: h})

	if msgsErr := h.registerMessageHandlers(i.services.Messages); msgsErr != nil {
		return nil, fmt.Errorf("message handlers: %w", msgsErr)
	}

	return h, nil
}

func (h *eventHandler) registerMessageHandlers(msgs rez.MessageService) error {
	return errors.Join(
		msgs.AddEventHandlers(
			rez.NewEventHandler("slack.events.callback_event", h.onCallbackEvent),
			rez.NewEventHandler("slack.incidents.on_updated", h.onIncidentUpdated),
			rez.NewEventHandler("slack.incidents.on_milestone_updated", h.onIncidentMilestoneUpdated),
		),
		msgs.AddCommandHandlers(
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

func (h *eventHandler) OnEventsAPIEvent(ctx context.Context, event rez.ProviderEvent) error {
	return h.services.ProviderEvents.IngestEvent(ctx, event)
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

func (h *eventHandler) withIncidentUpdateProcessor(ctx context.Context, incId uuid.UUID, fn func(*incidentUpdateProcessor) error) error {
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
	p, procErr := newIncidentUpdateProcessor(ctx, newChatService(ci), h.services, incId)
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

type slackEventsAPIProcessor struct {
	handler *eventHandler
}

func (p *slackEventsAPIProcessor) Provider() string { return integrationName }

func (p *slackEventsAPIProcessor) Source() string { return slackEventsAPISource }

func (p *slackEventsAPIProcessor) ProcessProviderEvent(ctx context.Context, event rez.ProviderEvent) error {
	ev, parseErr := slackevents.ParseEvent(event.Payload, slackevents.OptionNoVerifyToken())
	if parseErr != nil {
		return fmt.Errorf("parse event: %w", parseErr)
	}
	switch ev.Type {
	case slackevents.AppRateLimited:
		return p.handler.OnAppRateLimitedEvent(ctx)
	case slackevents.CallbackEvent:
		return p.handler.onCallbackEvent(ctx, &callbackEvent{Data: event.Payload})
	default:
		return fmt.Errorf("unhandled events api event type: %s", ev.Type)
	}
}
