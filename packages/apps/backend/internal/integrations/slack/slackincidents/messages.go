package slackincidents

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
)

type messageHandler struct {
	appCfg       rez.AppConfig
	db           rez.Database
	messages     rez.MessageService
	provEvents   rez.ProviderEventService
	integrations rez.IntegrationService
	incidents    rez.IncidentService
}

func (i *Integration) makeMessageHandler(cfg rez.Config, msgs rez.MessageService, provEvents rez.ProviderEventService) (*messageHandler, error) {
	h := &messageHandler{
		appCfg:       cfg.App,
		messages:     msgs,
		provEvents:   provEvents,
		db:           i.db,
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
			rez.NewEventHandler("slack.incidents.updated", h.onIncidentUpdated),
			rez.NewEventHandler("slack.incidents.milestone_updated", h.onIncidentMilestoneUpdated),
		),
		h.messages.AddCommandHandlers(
			rez.NewCommandHandler("slack.create_incident_channel", h.createIncidentChannel),
			rez.NewCommandHandler("slack.send_incident_milestone_message", h.sendIncidentMilestoneMessage),
		),
	)
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
