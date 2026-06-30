package slackincidents

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/integration"
	slackintegration "github.com/rezible/rezible/internal/integrations/slack"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type App struct {
	cfg       rez.Config
	db        rez.Database
	messages  rez.MessageService
	incidents rez.IncidentService
}

func MakeApp(cfg rez.Config, db rez.Database, msgs rez.MessageService, incidents rez.IncidentService) (*App, error) {
	h := &App{
		cfg:       cfg,
		db:        db,
		messages:  msgs,
		incidents: incidents,
	}
	if msgsErr := h.registerHandlers(); msgsErr != nil {
		return nil, fmt.Errorf("message handlers: %w", msgsErr)
	}
	return h, nil
}

func (a *App) registerHandlers() error {
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

func (a *App) IntegrationName() string {
	return integrationName
}

func (a *App) Config() rez.IntegrationsConfigSlackApp {
	return a.cfg.Integrations.Slack.Incidents
}

func (a *App) OAuthScopes() []string {
	return []string{
		"app_mentions:read",
		"assistant:write",
		"channels:history",
		"channels:join",
		"channels:read",
		"chat:write",
		"chat:write.customize",
		"chat:write.public",
		"commands",
		"groups:history",
		"groups:read",
		"im:history",
		"im:read",
		"im:write",
		"im:write.topic",
		"incoming-webhook",
		"metadata.message:read",
		"mpim:history",
		"pins:read",
		"reactions:read",
		"usergroups:read",
		"users.profile:read",
		"users:read",
		"users:read.email",
		"channels:write.topic",
		"channels:manage",
		"channels:write.invites",
	}
}

func (a *App) PublishProviderEventPipelineEventTypes() []slackevents.EventsAPIType {
	return []slackevents.EventsAPIType{}
}

func (a *App) SlashCommandHandlers() map[string]slackintegration.SlashCommandHandler {
	return map[string]slackintegration.SlashCommandHandler{
		"/incident": a.handleIncidentCommand,
	}
}

func (a *App) InteractionCallbackHandlers() map[slack.InteractionType]slackintegration.InteractionCallbackHandler {
	return map[slack.InteractionType]slackintegration.InteractionCallbackHandler{
		slack.InteractionTypeMessageAction:  a.handleMessageActionInteraction,
		slack.InteractionTypeBlockActions:   a.handleBlockActionsInteraction,
		slack.InteractionTypeViewSubmission: a.handleViewSubmissionInteraction,
	}
}

func (a *App) getEnabledIntegrationClient(ctx context.Context) (*slackintegration.ClientWrapper, error) {
	// TODO: cache?
	queryInstall := a.db.Client(ctx).Integration.Query().
		Where(integration.IntegrationName(integrationName))
	intg, intgErr := queryInstall.All(ctx)
	if intgErr != nil && !ent.IsNotFound(intgErr) {
		return nil, fmt.Errorf("failed to query: %w", intgErr)
	}
	// TODO: check preferences for workspace
	if len(intg) == 0 {
		return nil, nil
	}
	return slackintegration.NewClientWrapper(intg[0])
}

func (a *App) withIncidentUpdateProcessor(ctx context.Context, id uuid.UUID, fn func(*incidentUpdateProcessor) error) error {
	client, clientErr := a.getEnabledIntegrationClient(ctx)
	if clientErr != nil {
		return fmt.Errorf("get incident management client: %w", clientErr)
	}
	if client == nil {
		return nil
	}
	p, procErr := a.newUpdateProcessor(ctx, client, id)
	if procErr != nil {
		return fmt.Errorf("creating incident update processor: %w", procErr)
	}
	return fn(p)
}

func (a *App) onIncidentUpdated(ctx context.Context, ev *rez.EventOnIncidentUpdated) error {
	return a.withIncidentUpdateProcessor(ctx, ev.IncidentId, func(p *incidentUpdateProcessor) error {
		return p.processIncidentUpdate(ctx)
	})
}

func (a *App) onIncidentMilestoneUpdated(ctx context.Context, ev *rez.EventOnIncidentMilestoneUpdated) error {
	return a.withIncidentUpdateProcessor(ctx, ev.IncidentId, func(p *incidentUpdateProcessor) error {
		return p.processIncidentMilestoneUpdate(ctx, ev.MilestoneId)
	})
}

type cmdCreateIncidentChannel struct {
	IncidentId uuid.UUID
}

func (a *App) createIncidentChannel(ctx context.Context, ev *cmdCreateIncidentChannel) error {
	return a.withIncidentUpdateProcessor(ctx, ev.IncidentId, func(p *incidentUpdateProcessor) error {
		return p.createIncidentChannel(ctx)
	})
}

type cmdSendIncidentMilestoneMessage struct {
	IncidentId  uuid.UUID
	MilestoneId uuid.UUID
}

func (a *App) sendIncidentMilestoneMessage(ctx context.Context, ev *cmdSendIncidentMilestoneMessage) error {
	return a.withIncidentUpdateProcessor(ctx, ev.IncidentId, func(p *incidentUpdateProcessor) error {
		return p.sendIncidentMilestoneMessage(ctx, ev.MilestoneId)
	})
}
