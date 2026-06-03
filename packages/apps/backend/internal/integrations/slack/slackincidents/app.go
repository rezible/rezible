package slackincidents

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	slackintegration "github.com/rezible/rezible/internal/integrations/slack"
	"github.com/slack-go/slack"
)

type app struct {
	appCfg    rez.AppConfig
	db        rez.Database
	messages  rez.MessageService
	incidents rez.IncidentService
}

func makeApp(cfg rez.Config, db rez.Database, msgs rez.MessageService, incidents rez.IncidentService) (*app, error) {
	h := &app{
		appCfg:    cfg.App,
		db:        db,
		messages:  msgs,
		incidents: incidents,
	}
	if msgsErr := h.registerHandlers(); msgsErr != nil {
		return nil, fmt.Errorf("message handlers: %w", msgsErr)
	}
	return h, nil
}

func (a *app) registerHandlers() error {
	return errors.Join(
		a.messages.AddEventHandlers(
			rez.NewEventHandler("slack.incidents.updated", a.onIncidentUpdated),
			rez.NewEventHandler("slack.incidents.milestone_updated", a.onIncidentMilestoneUpdated),
		),
		a.messages.AddCommandHandlers(
			rez.NewCommandHandler("slack.create_incident_channel", a.createIncidentChannel),
			rez.NewCommandHandler("slack.send_incident_milestone_message", a.sendIncidentMilestoneMessage),
		),
	)
}

func (a *app) EventsApiHandler() slackintegration.EventsApiHandler {
	return a.handleEventsApiEvent
}

func (a *app) SlashCommandHandlers() map[string]slackintegration.SlashCommandHandler {
	return map[string]slackintegration.SlashCommandHandler{
		"/incident": a.handleIncidentCommand,
	}
}

func (a *app) InteractionCallbackHandlers() map[slack.InteractionType]slackintegration.InteractionCallbackHandler {
	return map[slack.InteractionType]slackintegration.InteractionCallbackHandler{
		slack.InteractionTypeMessageAction:  a.handleMessageActionInteraction,
		slack.InteractionTypeBlockActions:   a.handleBlockActionsInteraction,
		slack.InteractionTypeViewSubmission: a.handleViewSubmissionInteraction,
	}
}

func (a *app) withIncidentUpdateProcessor(ctx context.Context, id uuid.UUID, fn func(*incidentUpdateProcessor) error) error {
	p, procErr := a.newUpdateProcessor(ctx, id)
	if procErr != nil {
		return fmt.Errorf("creating incident update processor: %w", procErr)
	}
	return fn(p)
}

func (a *app) onIncidentUpdated(ctx context.Context, ev *rez.EventOnIncidentUpdated) error {
	return a.withIncidentUpdateProcessor(ctx, ev.IncidentId, func(p *incidentUpdateProcessor) error {
		return p.processIncidentUpdate(ctx)
	})
}

func (a *app) onIncidentMilestoneUpdated(ctx context.Context, ev *rez.EventOnIncidentMilestoneUpdated) error {
	return a.withIncidentUpdateProcessor(ctx, ev.IncidentId, func(p *incidentUpdateProcessor) error {
		return p.processIncidentMilestoneUpdate(ctx, ev.MilestoneId)
	})
}

type cmdCreateIncidentChannel struct {
	IncidentId uuid.UUID
}

func (a *app) createIncidentChannel(ctx context.Context, ev *cmdCreateIncidentChannel) error {
	return a.withIncidentUpdateProcessor(ctx, ev.IncidentId, func(p *incidentUpdateProcessor) error {
		return p.createIncidentChannel(ctx)
	})
}

type cmdSendIncidentMilestoneMessage struct {
	IncidentId  uuid.UUID
	MilestoneId uuid.UUID
}

func (a *app) sendIncidentMilestoneMessage(ctx context.Context, ev *cmdSendIncidentMilestoneMessage) error {
	return a.withIncidentUpdateProcessor(ctx, ev.IncidentId, func(p *incidentUpdateProcessor) error {
		return p.sendIncidentMilestoneMessage(ctx, ev.MilestoneId)
	})
}
