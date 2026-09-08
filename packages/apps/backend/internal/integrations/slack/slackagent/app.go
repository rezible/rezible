package slackagent

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rezible/rezible/ent"
	asb "github.com/rezible/rezible/ent/agentsessionbinding"
	"github.com/rezible/rezible/pkg/messages"
	"github.com/slack-go/slack/slackevents"

	rez "github.com/rezible/rezible"
	in "github.com/rezible/rezible/ent/integration"
	"github.com/rezible/rezible/ent/predicate"
	slackintegration "github.com/rezible/rezible/internal/integrations/slack"
	rezai "github.com/rezible/rezible/pkg/ai"
	"github.com/rezible/rezible/test"
)

type App struct {
	db     rez.Database
	cfg    rez.Config
	jobs   rez.JobService
	intgs  rez.IntegrationService
	users  rez.UserService
	agents rez.AgentSessionService
	events rez.EventsService

	responseClassifier rezai.ClassifyAgentThreadResponseWorkflowRunner
}

type AppSuite struct {
	test.Suite
}

func MakeApp(cfg rez.Config, db rez.Database, jobSvc rez.JobService, intgs rez.IntegrationService, users rez.UserService, agents rez.AgentSessionService, events rez.EventsService, responseClassifier rezai.ClassifyAgentThreadResponseWorkflowRunner) *App {
	return &App{
		db:                 db,
		cfg:                cfg,
		jobs:               jobSvc,
		intgs:              intgs,
		users:              users,
		agents:             agents,
		events:             events,
		responseClassifier: responseClassifier,
	}
}

func (a *App) MakeIntegration(deps *slackintegration.AppServiceDependencies) (*Integration, error) {
	svc, svcErr := slackintegration.NewAppService(a, deps)
	if svcErr != nil {
		return nil, fmt.Errorf("make app service: %w", svcErr)
	}
	return MakeIntegration(svc), nil
}

func (a *App) GetMessageHandlers() []rez.MessageEventHandler {
	return []rez.MessageEventHandler{
		messages.NewEventHandler("slackagent.OnAiAgentTurnFinished", a.onAiAgentTurnFinished),
	}
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

func (a *App) findMessageIdForAlertEvent(ctx context.Context, alertID uuid.UUID) (slackintegration.MessageId, error) {
	_ = ctx
	_ = alertID
	return "", nil
}

func (a *App) GetIntegrationClientWrapper(ctx context.Context, preds ...predicate.Integration) (*slackintegration.ClientWrapper, error) {
	preds = append(preds, in.Name(integrationName))
	intg, intgErr := a.intgs.LookupInstallation(ctx, in.And(preds...))
	if intgErr != nil {
		return nil, fmt.Errorf("query slack agent integration: %w", intgErr)
	}
	return slackintegration.NewClientWrapper(intg)
}

func (a *App) lookupAiAgentSessionBinding(ctx context.Context, sessionId uuid.UUID) (*ent.AgentSessionBinding, error) {
	preds := []predicate.AgentSessionBinding{
		asb.AgentSessionID(sessionId),
		asb.Provider(slackintegration.ProviderName),
	}
	binding, bindingErr := a.agents.LookupAgentSessionBinding(ctx, preds...)
	if bindingErr != nil {
		if ent.IsNotFound(bindingErr) {
			return nil, nil
		}
		return nil, fmt.Errorf("lookup: %w", bindingErr)
	}
	if binding.IntegrationID == nil {
		return nil, fmt.Errorf("slack session binding missing integration id")
	}
	return binding, nil
}

func (a *App) onAiAgentTurnFinished(ctx context.Context, ev *rezai.EventOnAgentTurnFinished) error {
	if len(ev.Response.Text()) == 0 {
		return nil
	}
	mdIntg, ok := ev.AgentSessionMetadata[agentSessionMetadataIntegrationKey]
	if !ok {
		return nil
	}
	if intgName, intgNameOk := mdIntg.(string); !intgNameOk || intgName != integrationName {
		return nil
	}
	return a.onSlackAgentResponseEvent(ctx, ev)
}
