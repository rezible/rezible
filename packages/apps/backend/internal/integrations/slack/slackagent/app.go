package slackagent

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	"github.com/rezible/rezible/ent"
	"github.com/rezible/rezible/ent/integration"
	slackintegration "github.com/rezible/rezible/internal/integrations/slack"
	"github.com/slack-go/slack/slackevents"
)

type App struct {
	cfg      rez.Config
	db       rez.Database
	jobs     rez.JobService
	messages rez.MessageService
	agents   rez.AgentService
	events   rez.EventsService
}

func MakeApp(cfg rez.Config, db rez.Database, jobSvc rez.JobService, msgs rez.MessageService, agents rez.AgentService, events rez.EventsService) (*App, error) {
	h := &App{
		cfg:      cfg,
		db:       db,
		jobs:     jobSvc,
		messages: msgs,
		agents:   agents,
		events:   events,
	}
	if msgsErr := h.registerMessageHandlers(); msgsErr != nil {
		return nil, fmt.Errorf("message handlers: %w", msgsErr)
	}
	return h, nil
}

func (a *App) registerMessageHandlers() error {
	return errors.Join()
}

func (a *App) IntegrationName() string {
	return integrationName
}

func (a *App) Config() rez.IntegrationsConfigSlackApp {
	return a.cfg.Integrations.Slack.Agent
}

func (a *App) PublishProviderEventPipelineEventTypes() []slackevents.EventsAPIType {
	return []slackevents.EventsAPIType{}
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

func (a *App) getEnabledIntegrationClient(ctx context.Context) (*slackintegration.ClientWrapper, error) {
	intgs, intgErr := a.db.Client(ctx).Integration.Query().
		Where(integration.IntegrationName(integrationName)).
		All(ctx)
	if intgErr != nil && !ent.IsNotFound(intgErr) {
		return nil, fmt.Errorf("query slack agent integration: %w", intgErr)
	}
	if len(intgs) == 0 {
		return nil, nil
	}
	return slackintegration.NewClientWrapper(intgs[0])
}

func (a *App) findMessageIdForAlertEvent(ctx context.Context, alertID uuid.UUID) (slackintegration.MessageId, error) {
	_ = ctx
	_ = alertID
	return "", nil
}
