package slackagent

import (
	"errors"
	"fmt"

	rez "github.com/rezible/rezible"
	"github.com/slack-go/slack/slackevents"
)

type App struct {
	cfg      rez.Config
	db       rez.Database
	messages rez.MessageService
	events   rez.EventsService
}

func MakeApp(cfg rez.Config, db rez.Database, msgs rez.MessageService, events rez.EventsService) (*App, error) {
	h := &App{
		cfg:      cfg,
		db:       db,
		messages: msgs,
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

func (a *App) AppConfig() rez.IntegrationsConfigSlackApp {
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
