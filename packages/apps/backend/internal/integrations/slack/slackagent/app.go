package slackagent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	rez "github.com/rezible/rezible"
	in "github.com/rezible/rezible/ent/integration"
	"github.com/rezible/rezible/ent/predicate"
	slackintegration "github.com/rezible/rezible/internal/integrations/slack"
	"github.com/rezible/rezible/pkg/ai"
	"github.com/slack-go/slack/slackevents"
)

type App struct {
	cfg      rez.Config
	jobs     rez.JobService
	messages rez.MessageService
	intgs    rez.IntegrationService
	users    rez.UserService
	agents   rez.AiAgentService
	events   rez.EventsService
}

func MakeApp(cfg rez.Config, jobSvc rez.JobService, msgs rez.MessageService, intgs rez.IntegrationService, users rez.UserService, agents rez.AiAgentService, events rez.EventsService) (*App, error) {
	h := &App{
		cfg:      cfg,
		jobs:     jobSvc,
		messages: msgs,
		intgs:    intgs,
		users:    users,
		agents:   agents,
		events:   events,
	}
	if msgsErr := h.registerMessageHandlers(); msgsErr != nil {
		return nil, fmt.Errorf("message handlers: %w", msgsErr)
	}
	return h, nil
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

func (a *App) registerMessageHandlers() error {
	return errors.Join(
		a.messages.AddEventHandlers(rez.NewEventHandler("slackagent.OnAiAgentRunSnapshot", a.onAiAgentRunSnapshot)))
}

func (a *App) GetIntegrationClientWrapper(ctx context.Context, preds ...predicate.Integration) (*slackintegration.ClientWrapper, error) {
	preds = append(preds, in.IntegrationName(integrationName))
	intg, intgErr := a.intgs.LookupInstallation(ctx, in.And(preds...))
	if intgErr != nil {
		return nil, fmt.Errorf("query slack agent integration: %w", intgErr)
	}
	return slackintegration.NewClientWrapper(intg)
}

func (a *App) onAiAgentRunSnapshot(ctx context.Context, ev *rez.EventOnAiAgentRunSnapshot) error {
	if ev.AgentName != ai.ChatAgent.Name {
		return nil
	}

	integrationRef, hasIntegrationRef := ev.RunMetadata["integration_ref"].(string)
	replyChannelId, hasSlackReplyChannelId := ev.RunMetadata["slack_reply_channel_id"].(string)
	replyThreadId, hasSlackReplyThreadId := ev.RunMetadata["slack_reply_thread_id"].(string)
	if !hasSlackReplyChannelId || !hasIntegrationRef || !hasSlackReplyThreadId {
		fmt.Printf("not a slack message?: %t %t %t\n", hasIntegrationRef, hasSlackReplyChannelId, hasSlackReplyThreadId)
		return nil
	}

	cw, wrapperErr := a.GetIntegrationClientWrapper(ctx, in.ExternalRef(integrationRef))
	if wrapperErr != nil {
		slog.Warn("failed to get slack integration client wrapper", "err", wrapperErr)
		return fmt.Errorf("get integration client wrapper: %w", wrapperErr)
	}

	output, outputErr := a.agents.GetAgentRunResult(ctx, ev.AgentRunId)
	if outputErr != nil {
		return fmt.Errorf("get agent run output: %w", outputErr)
	}

	parsed, parseErr := ai.ChatAgent.ParseOutput(output.Output)
	if parseErr != nil {
		return fmt.Errorf("parse output: %w", parseErr)
	}

	_, replyErr := cw.SendReply(ctx, replyChannelId, replyThreadId, parsed.Reply)
	if replyErr != nil {
		return fmt.Errorf("send reply: %w", replyErr)
	}
	return nil
}
